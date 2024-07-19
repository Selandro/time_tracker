package tracker_model

import (
	"time"
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

type Task struct {
	IDTask   int    `json:"id_task"`
	TaskName string `json:"task_name"`
}
