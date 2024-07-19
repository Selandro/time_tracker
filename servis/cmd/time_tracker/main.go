package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"main.go/cmd/internal/config"
	"main.go/cmd/internal/handlers/task"
	"main.go/cmd/internal/handlers/user"
	"main.go/cmd/internal/storage/cache"
	"main.go/cmd/internal/storage/postgresql"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	fmt.Println(cfg)

	log := setupLogger(cfg.Env)
	log.Info("starting time_tracker servis", slog.String("env", cfg.Env))
	log.Debug("debug message")

	db := postgresql.Connect(cfg.Database)
	defer db.Close()

	// Инициализация кэша
	cache.InitCache()

	cache.CacheAllUsersFromDB(db)

	//http.HandleFunc()
	// Настройка маршрутов и обработчиков
	http.HandleFunc("/adduser", user.AddUserHandler(db, log))
	http.HandleFunc("/start_task", task.StartTaskHandler(db, log))
	http.HandleFunc("/end_task", task.EndTaskHandler(db, log))
	http.HandleFunc("/user_task", task.GetUserTaskSummaryHandler(db, log))
	http.HandleFunc("/delete_user", user.DeleteUserHandler(db, log))
	http.HandleFunc("/update_user/", user.UpdateUserHandler(db, log))
	http.HandleFunc("/users", user.GetUsersHandler(db, log))

	server := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	log.Info("HTTP сервер запущен на", slog.String("адрес", cfg.HTTPServer.Address))

	err := server.ListenAndServe()
	if err != nil {
		log.Error("Ошибка запуска сервера", slog.String("ошибка", err.Error()))
		os.Exit(1)
	}

}
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
