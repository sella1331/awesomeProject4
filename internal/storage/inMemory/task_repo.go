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

func (s *Storage) GetTasks(userID string) ([]models.Task, error) {
	var tasksList []models.Task
	if len(s.tasks) == 0 {
		return nil, errors.ErrTaskIsEmpty
	}
	for _, task := range s.tasks {
		if task.UserID == userID {
			tasksList = append(tasksList, task)
		}
		return nil, errors.ErrNotFind
	}
	return tasksList, nil
}

func (s *Storage) SaveTask(task models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.tasks {
		if t.UserID == task.UserID &&
			t.Description == task.Description &&
			t.Title == task.Title {
			fmt.Printf("Задача с таким содержимым уже существует: %s. Пропускаем.\n", task.TaskID)
			return errors.ErrTaskAlreadyExists
		}
	}

	taskID := uuid.New().String()

	s.tasks[taskID] = models.Task{
		TaskID:      taskID,
		UserID:      task.UserID,
		Description: task.Description,
		Title:       task.Title,
		DateCreated: time.Now(),
		Urgency:     task.Urgency,
		Status:      "new",
	}
	return nil
}

func (s *Storage) DeleteTask(taskID string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if task, exists := s.tasks[taskID]; exists {
		if task.UserID == userID {
			delete(s.tasks, taskID)
			return nil
		}
	}
	return errors.ErrNotFind
}
func (s *Storage) InfoTask(taskID string, UserID string) (*models.Task, error) {
	fmt.Printf("InfoTask ищем: %q\n", taskID)
	fmt.Printf(" Всего задач: %d\n", len(s.tasks))
	if task, exists := s.tasks[taskID]; exists {
		if task.UserID == UserID {
			copyTask := task
			return &copyTask, nil
		}
		return nil, errors.ErrNotFind
	}
	return nil, errors.ErrNotFind
}
func (s *Storage) RepeatInfoTask(taskID string, userID string, newTask models.Task) (*models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists || task.UserID != userID {
		return nil, errors.ErrNotFind
	}

	oldTaskID := task.TaskID
	oldUserID := task.UserID
	oldDateCreated := task.DateCreated

	if err := mergo.Merge(&task, newTask, mergo.WithOverride); err != nil {
		return nil, err
	}

	task.TaskID = oldTaskID
	task.UserID = oldUserID
	task.DateCreated = oldDateCreated

	s.tasks[taskID] = task
	return &task, nil
}
