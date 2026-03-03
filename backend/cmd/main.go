package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"business_process_efficiency/internal/config"
)

func main() {

	cfg := config.LoadConfig()

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка открытия подключения:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("База не отвечает:", err)
	}

	fmt.Println("Успешное подключение к БД через .env!")
}
