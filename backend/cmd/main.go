package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"business_process_efficiency/internal/config"
	"business_process_efficiency/internal/database"
	"business_process_efficiency/internal/routes"
)

// @title Business Process Efficiency API
// @version 1.0
// @description API для системы прогнозирования ресурсов бизнес-процессов
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	cfg := config.LoadConfig()

	database.InitDatabase(cfg)
	database.Migrate()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	routes.SetupRoutes(r, cfg.JWTSecret)

	log.Println("Сервер запущен на :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Ошибка при запуске сервера:", err)
	}
}
