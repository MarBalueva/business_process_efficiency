package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"business_process_efficiency/internal/config"
	"business_process_efficiency/internal/models"
)

var DB *gorm.DB

func InitDatabase(cfg *config.Config) {

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	DB = db

	log.Println("GORM подключен к базе данных")
}

func Migrate() {
	err := DB.AutoMigrate(
		&models.AccessGroup{},
		&models.Department{},
		&models.Employee{},
		&models.EmployeeHR{},
		&models.Position{},
		&models.User{},
		&models.UserAccessGroup{})
	if err != nil {
		log.Fatal("Ошибка миграции:", err)
	}

	log.Println("Миграции выполнены")
}
