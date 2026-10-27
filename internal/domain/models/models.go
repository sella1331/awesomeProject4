package models

import "time"

type Task struct {
	TaskID      string    `json:"task_id,omitempty"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"        validate:"required"`
	Description string    `json:"description"`
	DateCreated time.Time `json:"date_created,omitempty"`
	Urgency     string    `json:"urgency"      validate:"required,oneof=low medium high"`
	Status      string    `json:"status,omitempty" validate:"omitempty,oneof=new in_progress finish"`
}

type User struct {
	UUID     string `json:"uuid"`
	Username string `json:"username" validate:"required,alphanum"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=30,containsany=!@#$%^&*"`
	taskArr  []Task
}

type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=30,containsany=!@#$%^&*"`
}
