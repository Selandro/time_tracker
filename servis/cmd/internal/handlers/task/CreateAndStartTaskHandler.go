package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"main.go/cmd/internal/storage/cache"
	"main.go/tracker_model"
)

// TaskRequest представляет данные запроса для начала отсчета времени по задаче.
// Включает идентификатор пользователя и идентификатор задачи.
type TaskRequest struct {
	UserID int `json:"user_id"` // Идентификатор пользователя
	IDTask int `json:"id_task"` // Идентификатор задачи
}

type UserTask struct {
	UserID       int       `json:"id_user"`
	IDTask       int       `json:"id_task"`
	TaskName     string    `json:"task_name"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	TotalMinutes int       `json:"total_minutes"`
}

// StartTaskHandler обрабатывает HTTP запросы для начала отсчета времени по задаче для пользователя.

// Декодирует запрос, добавляет задачу в базу данных, обновляет кэш и возвращает данные о задаче в формате JSON.
// @Summary Start a task
// @Description Start timing for a task for a user
// @Tags Task
// @Accept json
// @Produce json
// @Param task body TaskRequest true "Task Request"
// @Success 200 {object} UserTask "Task details"
// @Failure 400 {string} string "Invalid input"
// @Failure 500 {string} string "Failed to start task"
// @Router /api/v1/tasks/start [post]
func StartTaskHandler(db *sql.DB, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req TaskRequest

		// Декодирование JSON данных из тела запроса в структуру TaskRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error("Неверный формат ввода", slog.String("error", err.Error()))
			http.Error(w, "Неверный формат ввода", http.StatusBadRequest)
			return
		}

		// Добавление новой задачи в базу данных с получением имени задачи
		task, err := AddTaskToDBWithTaskName(log, db, req.UserID, req.IDTask)
		if err != nil {
			log.Error("Ошибка при добавлении задачи в базу данных", slog.Int("userID", req.UserID), slog.Int("taskID", req.IDTask), slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Ошибка при добавлении задачи в базу данных: %v", err), http.StatusInternalServerError)
			return
		}

		// Обновление кэша
		cache.UserCacheMutex.Lock()
		defer cache.UserCacheMutex.Unlock()

		// Проверка наличия пользователя в кэше
		user, exists := cache.UserCache[req.UserID]
		if !exists {
			log.Error("Пользователь не найден в кэше", slog.Int("userID", req.UserID))
			http.Error(w, "Пользователь не найден в кэше", http.StatusNotFound)
			return
		}

		// Добавление новой задачи в список задач пользователя
		user.UserTask = append(user.UserTask, task)
		cache.UserCache[req.UserID] = user

		// Установка заголовка Content-Type и кодирование ответа в JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)

		log.Info("Задача успешно начата", slog.Any("task", task))
	}
}

// AddTaskToDBWithTaskName добавляет новую задачу в базу данных и возвращает созданную задачу.
// Получает имя задачи из таблицы tasks по идентификатору задачи и вставляет запись в таблицу users_tasks.
func AddTaskToDBWithTaskName(log *slog.Logger, db *sql.DB, userID int, taskID int) (tracker_model.UserTask, error) {
	var taskName string
	// Получение имени задачи из таблицы tasks по заданному ID
	err := db.QueryRow(`SELECT task_name FROM tasks WHERE id_task = $1`, taskID).Scan(&taskName)
	if err != nil {
		log.Error("Ошибка при получении имени задачи из базы данных", slog.Int("taskID", taskID), slog.String("error", err.Error()))
		return tracker_model.UserTask{}, fmt.Errorf("ошибка при получении имени задачи из базы данных: %v", err)
	}

	startTime := time.Now() // Время начала отсчета
	endTime := time.Time{}  // Время окончания пока не установлено
	totalMinutes := 0       // Общее количество минут пока не установлено

	var task tracker_model.UserTask
	// Вставка новой задачи в таблицу users_tasks и возврат вставленной задачи
	err = db.QueryRow(`
		INSERT INTO users_tasks (user_id, id_task, task_name, start_time, end_time, total_minutes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING user_id, id_task, task_name, start_time, end_time, total_minutes
	`, userID, taskID, taskName, startTime, endTime, totalMinutes).Scan(&task.UserID, &task.IDTask, &task.TaskName, &task.StartTime, &task.EndTime, &task.TotalMinutes)
	if err != nil {
		log.Error("Ошибка при вставке задачи в базу данных", slog.Any("task", task), slog.String("error", err.Error()))
		return task, fmt.Errorf("ошибка при вставке задачи в базу данных: %v", err)
	}

	log.Info("Задача успешно добавлена в базу данных", slog.Any("task", task))
	return task, nil
}
