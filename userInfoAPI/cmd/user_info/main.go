package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"main.go/cmd/internal/config"
	"main.go/cmd/internal/handlers"
	userstorage "main.go/cmd/userStorage"

	"log/slog"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// Загрузка конфигурации
	cfg := config.MustLoad()
	fmt.Println(cfg)

	// Настройка логгера
	log := setupLogger(cfg.Env)
	log.Info("starting time_tracker service", slog.String("env", cfg.Env))
	log.Debug("debug message")

	// Инициализация хранилища пользователей
	UserInfo := userstorage.InitUserStorage()
	userstorage.UserData(UserInfo)
	fmt.Println(UserInfo[1234567890])

	// Настройка маршрутизации
	http.HandleFunc("/userinfo", handlers.GetUserInfo)

	// Настройка и запуск HTTP сервера
	server := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		ReadTimeout:  cfg.HTTPServer.Timeout * time.Second,
		WriteTimeout: cfg.HTTPServer.Timeout * time.Second,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout * time.Second,
	}

	log.Info("HTTP сервер запущен на", slog.String("адрес", cfg.HTTPServer.Address))

	err := server.ListenAndServe()
	if err != nil {
		log.Error("Ошибка запуска сервера", slog.String("ошибка", err.Error()))
		os.Exit(1)
	}
}

// Функция настройки логгера
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
