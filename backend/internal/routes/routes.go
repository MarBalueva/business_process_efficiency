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

	profileRepo := repository.NewProfileRepository(database.DB)
	profileService := service.NewProfileService(profileRepo)
	profileController := controller.NewProfileController(profileService)

	processRepo := repository.NewProcessRepository(database.DB)
	processService := service.NewProcessService(processRepo)
	processController := controller.NewProcessController(processService)

	stepController := controller.NewStepController(processService)

	measurementRepo := repository.NewMeasurementRepository(database.DB)
	measurementService := service.NewMeasurementService(measurementRepo)
	measurementController := controller.NewMeasurementController(measurementService)

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

		authRoutes := api.Group("/")

		authRoutes.Use(middleware.JWTMiddleware(authService))
		authRoutes.Use(middleware.RequireAccess(authService, "header", "employee", "analyst", "admin"))
		{
			authRoutes.GET("/employees", employeeController.GetAllEmployees)
			authRoutes.GET("/employees/:id", employeeController.GetEmployee)
			authRoutes.GET("/dict", dictController.GetAll)

			authRoutes.GET("/profile/me", profileController.GetProfile)
			authRoutes.PUT("/profile/me", profileController.UpdateProfile)
		}

		processRoutes := api.Group("/processes")
		processRoutes.Use(middleware.JWTMiddleware(authService))
		processRoutes.Use(middleware.RequireAccess(authService, "analyst", "admin"))
		{
			// процессы
			processRoutes.GET("/registry", processController.GetRegistry)
			processRoutes.GET("/:id", processController.GetProcess)
			processRoutes.POST("", processController.CreateProcess)
			processRoutes.PUT("/:id", processController.UpdateProcess)
			processRoutes.PATCH("/:id/move", processController.MoveProcess)
			processRoutes.DELETE("/:id", processController.DeleteProcess)

			// версии процессов
			processRoutes.POST("/versions", processController.CreateVersion)
			processRoutes.DELETE("/versions/:id", processController.DeleteVersion)

			// этапы процессов
			processRoutes.POST("/steps", stepController.CreateStep)
			processRoutes.PUT("/steps/:id", stepController.UpdateStep)
			processRoutes.POST("/steps/reorder", stepController.ReorderSteps)
			processRoutes.DELETE("/steps/:id", stepController.DeleteStep)

			// замеры времени этапов
			processRoutes.POST("/measurements/start", measurementController.Start)
			processRoutes.GET("/measurements", measurementController.List)
			processRoutes.POST("/measurements/pause", measurementController.Pause)
			processRoutes.POST("/measurements/resume", measurementController.Resume)
			processRoutes.POST("/measurements/finish", measurementController.Finish)
			processRoutes.POST("/measurements/reset", measurementController.Reset)
			processRoutes.DELETE("/measurements/:id", measurementController.Delete)
		}

		folderRoutes := api.Group("/process-folders")
		folderRoutes.Use(middleware.JWTMiddleware(authService))
		folderRoutes.Use(middleware.RequireAccess(authService, "analyst", "admin"))
		{
			folderRoutes.POST("", processController.CreateFolder)
			folderRoutes.PUT("/:id", processController.UpdateFolder)
			folderRoutes.PATCH("/:id/move", processController.MoveFolder)
			folderRoutes.DELETE("/:id", processController.DeleteFolder)
		}

		adminRoutes := api.Group("/")
		adminRoutes.Use(middleware.JWTMiddleware(authService))
		adminRoutes.Use(middleware.RequireAccess(authService, "admin"))
		{
			adminRoutes.POST("/employees", employeeController.CreateEmployee)
			adminRoutes.PUT("/employees/:id", employeeController.UpdateEmployee)
			adminRoutes.DELETE("/employees/:id", employeeController.DeleteEmployee)

			adminRoutes.POST("/users", userController.Create)
			adminRoutes.GET("/users", userController.GetAll)

			adminRoutes.GET("/users/:id", userController.GetByID)
			adminRoutes.PUT("/users/:id", userController.Update)
			adminRoutes.DELETE("/users/:id", userController.Delete)

			adminRoutes.GET("/users/:id/access-groups", userController.GetUserAccessGroups)
			adminRoutes.POST("/users/:id/access-groups", userController.AddAccessGroup)
			adminRoutes.DELETE("/users/:id/access-groups/:group_id", userController.RemoveAccessGroup)
			adminRoutes.GET("/users/by-employee/:employeeId", userController.GetByEmployee)
		}
	}

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
