package storage

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	model "main.go/tracker_model"
)

func RunMigrations(db *sql.DB) error {
	// Путь к файлам миграции
	files := []string{"C:/dev/projects/time_tracker/servis/cmd/internal/storage/migrations/000001_create_people_and_tasks.up.sql"}

	for _, file := range files {
		// Проверяем существование файла
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("файл миграции %s не найден", file)
		}

		// Читаем содержимое SQL-файла миграции
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return fmt.Errorf("ошибка чтения файла миграции %s: %v", file, err)
		}

		// Выполняем SQL-запрос из файла
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("ошибка выполнения миграции из файла %s: %v", file, err)
		}

		log.Printf("Выполнена миграция из файла: %s", file)
	}
	return nil
}

// insertUser вставляет информацию о пользователе в базу данных.
func InsertUser(user model.Users, db *sql.DB) error {
	query := `
		INSERT INTO users (id, passportSerie, passportNumber, surname, name, patronymic, address)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := db.Exec(query, user.UserID, user.PassportSerie, user.PassportNumber, user.Surname, user.Name, user.Patronymic, user.Address)
	return err
}

// insertUserTask вставляет информацию о задаче пользователя в базу данных.
func InsertUserTask(userTask model.UserTask, db *sql.DB) error {
	query := `
		INSERT INTO users_task (id, id_task, task_name, start_time, end_time, total_minutes)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(query, userTask.UserID, userTask.IDTask, userTask.TaskName, userTask.StartTime, userTask.EndTime, userTask.TotalMinutes)
	return err
}

func InsertTask(task model.Task, db *sql.DB) error {
	query := `
		INSERT INTO users_task (id_task, task_name)
		VALUES ($1, $2)`
	_, err := db.Exec(query, task.IDTask, task.TaskName)
	return err
}

// AddUserToDB добавляет нового пользователя в базу данных и возвращает его ID
func AddUserToDB(db *sql.DB, passportSerie, passportNumber int) (int, error) {
	var userID int
	err := db.QueryRow(`
		INSERT INTO users (passport_serie, passport_number, surname, name, patronymic, address)
		VALUES ($1, $2, '', '', '', '')
		RETURNING id
	`, passportSerie, passportNumber).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
