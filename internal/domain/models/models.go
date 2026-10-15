package models

type Task struct {
	TaskID      string `json:"task_id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DateCreated string `json:"date_created,omitempty"`
	Urgency     string `json:"urgency"`
	Status      string `json:"status,omitempty"`
}

type User struct {
	UUID     string `json:"uuid"`
	Username string `json:"username" validate:"required,alphanum"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=30,containsany=!@#$%^&*"`
}

type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=30,containsany=!@#$%^&*"`
}
