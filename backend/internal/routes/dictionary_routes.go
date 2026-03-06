package routes

import (
	"business_process_efficiency/internal/controller"
	"business_process_efficiency/internal/middleware"
	"business_process_efficiency/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterDictionaryRoutes(
	api *gin.RouterGroup,
	authService *service.AuthService,
	controller *controller.DictionaryController,
	path string,
	modelFactory func() interface{},
	sliceFactory func() interface{},
) {

	read := api.Group("/")
	read.Use(middleware.JWTMiddleware(authService))
	read.Use(middleware.RequireAccess(authService, "header", "employee", "analyst", "admin"))
	{
		read.GET("/dict/"+path, func(c *gin.Context) {
			items := sliceFactory()
			controller.List(c, items)
		})
	}

	admin := api.Group("/")
	admin.Use(middleware.JWTMiddleware(authService))
	admin.Use(middleware.RequireAccess(authService, "admin"))
	{
		admin.POST("/dict/"+path, func(c *gin.Context) {
			controller.Create(c, modelFactory())
		})

		admin.PUT("/dict/"+path+"/:id", func(c *gin.Context) {
			controller.Update(c, modelFactory())
		})

		admin.DELETE("/dict/"+path+"/:id", func(c *gin.Context) {
			controller.Delete(c, modelFactory())
		})
	}
}
