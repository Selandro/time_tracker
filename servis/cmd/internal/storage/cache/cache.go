package cache

import (
	"database/sql"
	"log"
	"sort"
	"sync"

	model "main.go/tracker_model"
)

var (
	UserCache      map[int]model.Users
	UserCacheMutex sync.RWMutex

	TasksCache      map[int]model.Task
	TasksCacheMutex sync.RWMutex
)

func InitCache() {
	UserCache = make(map[int]model.Users)
	TasksCache = make(map[int]model.Task)
}

// CacheUser добавляет пользователей в кэш.
func CacheUser(user model.Users) {
	UserCacheMutex.Lock()
	defer UserCacheMutex.Unlock()
	UserCache[user.UserID] = user
}

// CacheTask добавляет задачи в кэш.
func CacheTask(task model.Task) {
	TasksCacheMutex.Lock()
	defer TasksCacheMutex.Unlock()
	TasksCache[task.IDTask] = task
}

// CacheAllUsersFromDB загружает всех пользователей и их задачи из базы данных и кэширует их.
func CacheAllUsersFromDB(db *sql.DB) {
	// Выполнение SQL-запроса для получения всех пользователей.
	userRows, err := db.Query("SELECT id, passport_serie, passport_number, surname, name, patronymic, address FROM users")
	if err != nil {
		log.Fatalf("Ошибка выполнения запроса для получения пользователей: %v", err)
	}
	defer userRows.Close()

	// Очистка кэша перед заполнением новыми данными.
	InitCache()

	// Обработка результатов запроса.
	for userRows.Next() {
		var user model.Users
		err := userRows.Scan(&user.UserID, &user.PassportSerie, &user.PassportNumber, &user.Surname, &user.Name, &user.Patronymic, &user.Address)
		if err != nil {
			log.Fatalf("Ошибка сканирования строки результата: %v", err)
		}

		// Получение задач для текущего пользователя.
		taskRows, err := db.Query("SELECT user_id, id_task, task_name, start_time, end_time, total_minutes FROM users_tasks WHERE user_id = $1", user.UserID)
		if err != nil {
			log.Fatalf("Ошибка выполнения запроса для получения задач пользователя: %v", err)
		}
		defer taskRows.Close()

		for taskRows.Next() {
			var task model.UserTask
			err := taskRows.Scan(&task.UserID, &task.IDTask, &task.TaskName, &task.StartTime, &task.EndTime, &task.TotalMinutes)
			if err != nil {
				log.Fatalf("Ошибка сканирования строки задачи: %v", err)
			}
			user.UserTask = append(user.UserTask, task)
		}

		// Проверка ошибок, возникших при итерации по строкам результата.
		if err := taskRows.Err(); err != nil {
			log.Fatalf("Ошибка итерации по строкам результата задач: %v", err)
		}

		CacheUser(user)
	}

	// Проверка ошибок, возникших при итерации по строкам результата пользователей.
	if err := userRows.Err(); err != nil {
		log.Fatalf("Ошибка итерации по строкам результата пользователей: %v", err)
	}
}

// GetUserTaskFromCache получает задачу из кэша по его идентификатору.
func GetUserTaskFromCache(userID int) ([]model.UserTask, bool) {
	UserCacheMutex.RLock()
	defer UserCacheMutex.RUnlock()
	user, exists := UserCache[userID]
	if !exists {
		return nil, false
	}

	// Копируем задачи пользователя для сортировки
	userTasks := append([]model.UserTask{}, user.UserTask...)

	// Сортируем задачи по трудозатратам (TotalMinutes) от большей к меньшей
	sort.Slice(userTasks, func(i, j int) bool {
		return userTasks[i].TotalMinutes > userTasks[j].TotalMinutes
	})

	return userTasks, true
}
