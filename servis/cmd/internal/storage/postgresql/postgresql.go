package postgresql

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"main.go/cmd/internal/config"
	"main.go/cmd/internal/storage"
)

// Connect устанавливает соединение с базой данных и возвращает объект DB.
func Connect(cfg config.DatabaseConfig) *sql.DB {
	// Формируем строку подключения к базе данных
	dbInfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)
	// Открываем соединение с базой данных
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// Создаем необходимые таблицы, если они еще не существуют
	if err := storage.RunMigrations(db); err != nil {
		log.Fatalf("Ошибка при выполнении миграций: %v", err)
	}

	log.Println("Подключение к базе данных успешно установлено")
	return db
}
