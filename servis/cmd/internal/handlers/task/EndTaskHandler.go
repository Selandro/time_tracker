package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"log/slog"

	"main.go/cmd/internal/storage/cache"
	model "main.go/tracker_model"
)

// EndTaskHandler обрабатывает HTTP запросы для завершения отсчета времени по задаче для пользователя.

// Обновляет время окончания задачи, вычисляет общее время выполнения и обновляет информацию в базе данных и кэше.
// @Summary Завершение задачи
// @Description Обновляет время окончания задачи, вычисляет общее время выполнения и обновляет информацию в базе данных и кэше.
// @Tags Task
// @Accept json
// @Produce json
// @Param request body TaskRequest true "Данные для завершения задачи"
// @Success 200 {object} UserTask "Информация о задаче"
// @Failure 400 {string} string "Неверный формат ввода"
// @Failure 404 {string} string "Пользователь не найден в кэше"
// @Failure 500 {string} string "Ошибка при обновлении задачи"
// @Router /api/v1/tasks/end [post]
func EndTaskHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req TaskRequest

		log.Info("Получен запрос на завершение задачи", slog.Int("user_id", req.UserID), slog.Int("task_id", req.IDTask))
		log.Debug("Запрос на завершение задачи", slog.Any("request", req))

		// Декодирование JSON данных из тела запроса в структуру TaskRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error("Неверный формат ввода", slog.String("error", err.Error()))
			http.Error(w, "Неверный формат ввода", http.StatusBadRequest)
			return
		}

		log.Info("Начато обновление времени окончания задачи", slog.Int("user_id", req.UserID), slog.Int("task_id", req.IDTask))

		// Обновление времени окончания задачи в базе данных
		err = updateTaskEndTime(db, req.UserID, req.IDTask)
		if err != nil {
			log.Error("Ошибка при обновлении времени окончания задачи в базе данных", slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Ошибка при обновлении времени окончания задачи: %v", err), http.StatusInternalServerError)
			return
		}

		log.Info("Время окончания задачи обновлено", slog.Int("user_id", req.UserID), slog.Int("task_id", req.IDTask))

		// Получение информации о задаче из базы данных для вычисления времени выполнения
		task, err := getTaskFromDB(db, req.UserID, req.IDTask)
		if err != nil {
			log.Error("Ошибка при получении задачи из базы данных", slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Ошибка при получении задачи из базы данных: %v", err), http.StatusInternalServerError)
			return
		}

		log.Info("Задача получена из базы данных", slog.Int("user_id", req.UserID), slog.Int("task_id", req.IDTask))
		log.Debug("Информация о задаче", slog.Any("task", task))

		if !task.StartTime.IsZero() && !task.EndTime.IsZero() {
			// Вычисление общего времени выполнения задачи в минутах
			task.TotalMinutes = int(task.EndTime.Sub(task.StartTime).Minutes())

			log.Info("Общее время выполнения задачи вычислено", slog.Int("total_minutes", task.TotalMinutes))

			// Обновление total_minutes в базе данных
			err = updateTaskTotalMinutes(db, req.UserID, req.IDTask, task.TotalMinutes)
			if err != nil {
				log.Error("Ошибка при обновлении total_minutes задачи в базе данных", slog.String("error", err.Error()))
				http.Error(w, fmt.Sprintf("Ошибка при обновлении total_minutes в базе данных: %v", err), http.StatusInternalServerError)
				return
			}

			log.Info("total_minutes задачи обновлено в базе данных", slog.Int("user_id", req.UserID), slog.Int("task_id", req.IDTask))
		}

		// Обновление кэша
		cache.UserCacheMutex.Lock()
		defer cache.UserCacheMutex.Unlock()

		user, exists := cache.UserCache[req.UserID]
		if !exists {
			log.Warn("Пользователь не найден в кэше", slog.Int("user_id", req.UserID))
			http.Error(w, "Пользователь не найден в кэше", http.StatusNotFound)
			return
		}

		// Найти задачу в кэше и обновить время окончания и общее время выполнения
		for i := range user.UserTask {
			if user.UserTask[i].IDTask == req.IDTask {
				user.UserTask[i].EndTime = time.Now()
				user.UserTask[i].TotalMinutes = task.TotalMinutes
				break
			}
		}

		cache.UserCache[req.UserID] = user

		log.Info("Кэш успешно обновлен", slog.Int("user_id", req.UserID), slog.Int("task_id", req.IDTask))
		log.Debug("Обновленный кэш пользователя", slog.Any("user_cache", user))

		// Установка заголовка и кодирование ответа в JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(task)

		log.Info("Ответ успешно отправлен", slog.Int("user_id", req.UserID), slog.Int("task_id", req.IDTask))
		log.Debug("Отправленный ответ", slog.Any("response", task))
	}
}

// updateTaskEndTime обновляет время окончания задачи в базе данных.
// Устанавливает текущее время как время окончания задачи.
func updateTaskEndTime(db *sql.DB, userID int, taskID int) error {
	_, err := db.Exec(`
		UPDATE users_tasks
		SET end_time = $1
		WHERE user_id = $2 AND id_task = $3
	`, time.Now(), userID, taskID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении времени окончания задачи в базе данных: %v", err)
	}
	return nil
}

// getTaskFromDB получает информацию о задаче из базы данных по идентификатору пользователя и задачи.
func getTaskFromDB(db *sql.DB, userID int, taskID int) (model.UserTask, error) {
	var task model.UserTask
	err := db.QueryRow(`
		SELECT user_id, id_task, task_name, start_time, end_time, total_minutes
		FROM users_tasks
		WHERE user_id = $1 AND id_task = $2
	`, userID, taskID).Scan(&task.UserID, &task.IDTask, &task.TaskName, &task.StartTime, &task.EndTime, &task.TotalMinutes)
	if err != nil {
		return task, fmt.Errorf("ошибка при получении задачи из базы данных: %v", err)
	}
	return task, nil
}

// updateTaskTotalMinutes обновляет общее время выполнения задачи в базе данных.
// Устанавливает значение total_minutes для задачи.
func updateTaskTotalMinutes(db *sql.DB, userID int, taskID int, totalMinutes int) error {
	_, err := db.Exec(`
		UPDATE users_tasks
		SET total_minutes = $1
		WHERE user_id = $2 AND id_task = $3
	`, totalMinutes, userID, taskID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении total_minutes задачи в базе данных: %v", err)
	}
	return nil
}
