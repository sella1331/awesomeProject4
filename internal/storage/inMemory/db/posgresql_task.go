package db

import (
	"awesomeProject4/internal/domain/models"
	"database/sql"
)

type PsqlStorageTask struct {
	db *sql.DB
}

func NewPsqlStorageTask(db *sql.DB) *PsqlStorage {
	return &PsqlStorage{db: db}
}
func (s *PsqlStorage) GetTasks() ([]models.Task, error) {
	//TODO implement me
	panic("implement me")
}

func (s *PsqlStorage) SaveTask(task models.Task) error {
	//TODO implement me
	panic("implement me")
}

func (s *PsqlStorage) DeleteTask(taskID string) error {
	//TODO implement me
	panic("implement me")
}

func (s *PsqlStorage) InfoTask(taskID string) (*models.Task, error) {
	//TODO implement me
	panic("implement me")
}

func (s *PsqlStorage) RepeatInfoTask(taskID string, task models.Task) (*models.Task, error) {
	//TODO implement me
	panic("implement me")
}
