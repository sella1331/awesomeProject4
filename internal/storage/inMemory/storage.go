package inMemory

import (
	"awesomeProject4/internal/domain/errors"
	"awesomeProject4/internal/domain/models"
	"fmt"
	"sync"
	"time"

	"dario.cat/mergo"
	"github.com/google/uuid"
)

type Storage struct {
	mu    sync.RWMutex
	tasks map[string]models.Task
}

func NewStorage() *Storage {
	return &Storage{
		tasks: make(map[string]models.Task),
	}
}

func (s *Storage) GetTasks() ([]models.Task, error) {
	var tasksList []models.Task
	if len(s.tasks) == 0 {
		return nil, errors.ErrTaskIsEmpty
	}
	for _, task := range s.tasks {
		tasksList = append(tasksList, task)
	}
	return tasksList, nil
}

func (s *Storage) SaveTask(task models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.tasks {
		if t.Description == task.Description && t.Title == task.Title {
			fmt.Printf("Задача с таким содержимым уже существует: %s. Пропускаем.\n", task.TaskID)
			return errors.ErrTaskAlreadyExists
		}
	}

	taskID := uuid.New().String()
	dateCreated := time.Now().Format("2006-01-02 15:04:05")

	s.tasks[taskID] = models.Task{
		TaskID:      taskID,
		Description: task.Description,
		Title:       task.Title,
		DateCreated: dateCreated,
		Urgency:     task.Urgency,
		Status:      "New",
	}
	return nil
}

func (s *Storage) DeleteTask(taskID string) error {
	if _, exists := s.tasks[taskID]; !exists {
		return errors.ErrNotFind
	}

	delete(s.tasks, taskID)
	return nil
}
func (s *Storage) InfoTask(taskID string) (*models.Task, error) {
	fmt.Printf(">>> InfoTask ищем: %q\n", taskID)
	fmt.Printf(">>> Всего задач: %d\n", len(s.tasks))
	if task, exists := s.tasks[taskID]; exists {
		copyTask := task
		return &copyTask, nil
	}
	return nil, errors.ErrNotFind
}
func (s *Storage) RepeatInfoTask(taskID string, newTaks models.Task) (*models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, errors.ErrNotFind
	}

	err := mergo.Merge(&task, newTaks, mergo.WithOverride)
	if err != nil {
		return nil, err
	}

	s.tasks[taskID] = task
	return &task, nil
}
