package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	model "main.go/tracker_model"
)

type Users struct {
	UserID         int        `json:"id"`
	PassportSerie  int        `json:"passport_serie"`
	PassportNumber int        `json:"passport_number"`
	Surname        string     `json:"surname"`
	Name           string     `json:"name"`
	Patronymic     string     `json:"patronymic"`
	Address        string     `json:"address"`
	UserTask       []UserTask `json:"userTask"`
}
type UserTask struct {
	UserID       int       `json:"id_user"`
	IDTask       int       `json:"id_task"`
	TaskName     string    `json:"task_name"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	TotalMinutes int       `json:"total_minutes"`
}

// @Summary Get users
// @Description Get a list of users with optional filters and pagination
// @Tags User
// @Accept json
// @Produce json
// @Param passport_serie query string false "Passport Serie"
// @Param passport_number query string false "Passport Number"
// @Param surname query string false "Surname"
// @Param name query string false "Name"
// @Param patronymic query string false "Patronymic"
// @Param address query string false "Address"
// @Param page query int false "Page number"
// @Param limit query int false "Limit per page"
// @Success 200 {array} Users "List of users"
// @Failure 400 {string} string "Invalid query parameters"
// @Failure 500 {string} string "Failed to retrieve users"
// @Router /api/v1/users [get]
func GetUsersHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получение параметров фильтрации и пагинации из URL
		passportSerieStr := r.URL.Query().Get("passport_serie")
		passportNumberStr := r.URL.Query().Get("passport_number")
		surname := r.URL.Query().Get("surname")
		name := r.URL.Query().Get("name")
		patronymic := r.URL.Query().Get("patronymic")
		address := r.URL.Query().Get("address")
		pageStr := r.URL.Query().Get("page")
		limitStr := r.URL.Query().Get("limit")

		// Установка значений по умолчанию для пагинации
		page := 1
		limit := 10
		var err error

		if pageStr != "" {
			page, err = strconv.Atoi(pageStr)
			if err != nil || page < 1 {
				log.Error("Invalid page number", slog.String("pageStr", pageStr), slog.String("error", err.Error()))
				http.Error(w, "Invalid page number", http.StatusBadRequest)
				return
			}
		}

		if limitStr != "" {
			limit, err = strconv.Atoi(limitStr)
			if err != nil || limit < 1 {
				log.Error("Invalid limit number", slog.String("limitStr", limitStr), slog.String("error", err.Error()))
				http.Error(w, "Invalid limit number", http.StatusBadRequest)
				return
			}
		}

		offset := (page - 1) * limit

		// Построение запроса с учетом фильтров
		query := "SELECT passport_serie, passport_number, surname, name, patronymic, address FROM users WHERE 1=1"
		args := []interface{}{}
		argID := 1

		if passportSerieStr != "" {
			query += fmt.Sprintf(" AND passport_serie = $%d", argID)
			args = append(args, passportSerieStr)
			argID++
		}

		if passportNumberStr != "" {
			query += fmt.Sprintf(" AND passport_number = $%d", argID)
			args = append(args, passportNumberStr)
			argID++
		}

		if surname != "" {
			query += fmt.Sprintf(" AND surname ILIKE $%d", argID)
			args = append(args, "%"+surname+"%")
			argID++
		}

		if name != "" {
			query += fmt.Sprintf(" AND name ILIKE $%d", argID)
			args = append(args, "%"+name+"%")
			argID++
		}

		if patronymic != "" {
			query += fmt.Sprintf(" AND patronymic ILIKE $%d", argID)
			args = append(args, "%"+patronymic+"%")
			argID++
		}

		if address != "" {
			query += fmt.Sprintf(" AND address ILIKE $%d", argID)
			args = append(args, "%"+address+"%")
			argID++
		}

		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
		args = append(args, limit, offset)

		log.Debug("Executing database query", slog.String("query", query), slog.Any("args", args))
		// Выполнение запроса к базе данных
		rows, err := db.Query(query, args...)
		if err != nil {
			log.Error("Database query failed", slog.String("query", query), slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Database query failed: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []model.Users
		for rows.Next() {
			var user model.Users
			err := rows.Scan(&user.PassportSerie, &user.PassportNumber, &user.Surname, &user.Name, &user.Patronymic, &user.Address)
			if err != nil {
				log.Error("Failed to scan row", slog.String("error", err.Error()))
				http.Error(w, fmt.Sprintf("Failed to scan row: %v", err), http.StatusInternalServerError)
				return
			}
			users = append(users, user)
		}

		// Проверка на ошибки, возникшие при итерации по строкам
		if err = rows.Err(); err != nil {
			log.Error("Row iteration error", slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Row iteration error: %v", err), http.StatusInternalServerError)
			return
		}

		log.Info("Users retrieved successfully", slog.Int("count", len(users)))
		// Установка заголовка и кодирование ответа в JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}
}
