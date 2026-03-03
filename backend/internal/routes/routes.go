package routes

import (
	"business_process_efficiency/internal/controller"
	"business_process_efficiency/internal/middleware"
	"business_process_efficiency/internal/service"

	_ "business_process_efficiency/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine, jwtSecret string) {

	userController := controller.NewUserController()
	authService := service.NewAuthService(jwtSecret)
	authController := controller.NewAuthController(authService)

	api := r.Group("/api")
	{
		api.POST("/login", authController.Login)

		adminRoutes := api.Group("/")
		adminRoutes.Use(middleware.JWTMiddleware(authService))
		adminRoutes.Use(middleware.RequireAccess(authService, "admin"))
		{
			adminRoutes.POST("/users", userController.Create)
			adminRoutes.GET("/users", userController.GetAll)
		}
	}

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
