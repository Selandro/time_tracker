package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"log/slog"

	model "main.go/tracker_model"
)

// GetUserTaskSummaryHandler обрабатывает запросы на получение трудозатрат по пользователю за период

// @Summary Получение трудозатрат по пользователю за период
// @Description Возвращает список задач пользователя с их трудозатратами за указанный период времени.
// @Tags Task
// @Accept json
// @Produce json
// @Param user_id query int true "Идентификатор пользователя"
// @Param start_date query string true "Дата начала периода в формате YYYY-MM-DD"
// @Param end_date query string true "Дата окончания периода в формате YYYY-MM-DD"
// @Success 200 {array} UserTask "Список трудозатрат пользователя"
// @Failure 400 {string} string "Неверные параметры запроса"
// @Failure 500 {string} string "Ошибка при выполнении запроса к базе данных"
// @Router /api/v1/tasks/summary [get]
func GetUserTaskSummaryHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получение параметров запроса
		userIDStr := r.URL.Query().Get("user_id")
		startDateStr := r.URL.Query().Get("start_date")
		endDateStr := r.URL.Query().Get("end_date")

		log.Info("Получен запрос на получение трудозатрат по пользователю", slog.String("user_id", userIDStr), slog.String("start_date", startDateStr), slog.String("end_date", endDateStr))

		// Проверка наличия параметров
		if userIDStr == "" || startDateStr == "" || endDateStr == "" {
			log.Warn("Отсутствуют параметры запроса", slog.String("user_id", userIDStr), slog.String("start_date", startDateStr), slog.String("end_date", endDateStr))
			http.Error(w, "Missing parameters", http.StatusBadRequest)
			return
		}

		// Преобразование параметров
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			log.Error("Неверный формат user_id", slog.String("error", err.Error()))
			http.Error(w, "Invalid user_id", http.StatusBadRequest)
			return
		}

		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			log.Error("Неверный формат start_date", slog.String("error", err.Error()))
			http.Error(w, "Invalid start_date format", http.StatusBadRequest)
			return
		}

		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			log.Error("Неверный формат end_date", slog.String("error", err.Error()))
			http.Error(w, "Invalid end_date format", http.StatusBadRequest)
			return
		}

		log.Debug("Параметры запроса успешно преобразованы", slog.Int("user_id", userID), slog.Time("start_date", startDate), slog.Time("end_date", endDate))

		// Выполнение запроса к базе данных
		query := `
		SELECT 
			user_id,
			id_task,
			task_name,
			start_time,
			end_time,
			total_minutes
		FROM 
			users_tasks
		WHERE 
			user_id = $1 AND
			start_time >= $2 AND 
			(end_time <= $3 OR end_time IS NULL)
		ORDER BY 
			total_minutes DESC;
		`
		rows, err := db.Query(query, userID, startDate, endDate)
		if err != nil {
			log.Error("Ошибка выполнения запроса к базе данных", slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Database query failed: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		log.Info("Запрос к базе данных выполнен успешно", slog.Int("user_id", userID), slog.Time("start_date", startDate), slog.Time("end_date", endDate))

		var summaries []model.UserTask
		for rows.Next() {
			var summary model.UserTask
			err := rows.Scan(&summary.UserID, &summary.IDTask, &summary.TaskName, &summary.StartTime, &summary.EndTime, &summary.TotalMinutes)
			if err != nil {
				log.Error("Ошибка сканирования строки результата", slog.String("error", err.Error()))
				http.Error(w, fmt.Sprintf("Failed to scan row: %v", err), http.StatusInternalServerError)
				return
			}
			summaries = append(summaries, summary)
		}

		// Проверка на ошибки, возникшие при итерации по строкам
		if err = rows.Err(); err != nil {
			log.Error("Ошибка итерации по строкам результата", slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Row iteration error: %v", err), http.StatusInternalServerError)
			return
		}

		log.Debug("Сформирован список трудозатрат", slog.Any("summaries", summaries))

		// Установка заголовка и кодирование ответа в JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summaries)

		log.Info("Ответ успешно отправлен", slog.Int("user_id", userID))
	}
}
