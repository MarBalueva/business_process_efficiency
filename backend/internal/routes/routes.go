package routes

import (
	"business_process_efficiency/internal/controller"
	"business_process_efficiency/internal/database"
	"business_process_efficiency/internal/middleware"
	"business_process_efficiency/internal/models"
	"business_process_efficiency/internal/repository"
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

	employeeRepo := repository.NewEmployeeRepository(database.DB)
	employeeService := service.NewEmployeeService(employeeRepo)
	employeeController := controller.NewEmployeeController(employeeService)

	dictRepo := repository.NewDictionaryRepository(database.DB)
	dictService := service.NewDictionaryService(dictRepo)
	dictController := controller.NewDictionaryController(dictService)

	api := r.Group("/api")
	{
		api.POST("/login", authController.Login)

		RegisterDictionaryRoutes(
			api,
			authService,
			dictController,
			"departments",
			func() interface{} { return &models.Department{} },
			func() interface{} { return &[]models.Department{} },
		)

		RegisterDictionaryRoutes(
			api,
			authService,
			dictController,
			"positions",
			func() interface{} { return &models.Position{} },
			func() interface{} { return &[]models.Position{} },
		)

		RegisterDictionaryRoutes(
			api,
			authService,
			dictController,
			"access_groups",
			func() interface{} { return &models.AccessGroup{} },
			func() interface{} { return &[]models.AccessGroup{} },
		)

		// Все авторизованные пользователи
		authRoutes := api.Group("/")
		authRoutes.Use(middleware.JWTMiddleware(authService))
		authRoutes.Use(middleware.RequireAccess(authService, "header", "employee", "analyst", "admin"))
		{
			authRoutes.GET("/employees", employeeController.GetAllEmployees)
			authRoutes.GET("/employees/:id", employeeController.GetEmployee)
			authRoutes.GET("/dict", dictController.GetAll)
		}

		// Только администратор
		adminRoutes := api.Group("/")
		adminRoutes.Use(middleware.JWTMiddleware(authService))
		adminRoutes.Use(middleware.RequireAccess(authService, "admin"))
		{
			adminRoutes.POST("/employees", employeeController.CreateEmployee)
			adminRoutes.PUT("/employees/:id", employeeController.UpdateEmployee)
			adminRoutes.DELETE("/employees/:id", employeeController.DeleteEmployee)

			adminRoutes.POST("/users", userController.Create)
			adminRoutes.GET("/users", userController.GetAll)
		}
	}

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
