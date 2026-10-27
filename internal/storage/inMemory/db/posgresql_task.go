package db

import (
	"awesomeProject4/internal/domain/errors"
	"awesomeProject4/internal/domain/models"
	"context"
	"database/sql"
	errGLob "errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PsqlStorageTask struct {
	db *sql.DB
}

func NewPsqlStorageTask(db *sql.DB) *PsqlStorage {
	return &PsqlStorage{db: db}
}

func (s *PsqlStorage) GetTasks(userID string) ([]models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT task_id, user_id, title, description, date_created, urgency, status
		FROM tasks
		WHERE user_id = $1
		ORDER BY date_created DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("postgres: get tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]models.Task, 0)
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(
			&t.TaskID,
			&t.UserID,
			&t.Title,
			&t.Description,
			&t.DateCreated,
			&t.Urgency,
			&t.Status,
		); err != nil {
			return nil, fmt.Errorf("postgres: scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: rows iteration: %w", err)
	}

	return tasks, nil
}

func (s *PsqlStorage) SaveTask(task models.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	task.TaskID = uuid.NewString()

	var status any
	if task.Status != "" {
		status = task.Status
	}

	query := `
		INSERT INTO tasks (task_id, user_id, title, description, urgency, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING task_id, user_id, title, description, date_created, urgency, status`

	var created models.Task
	err := s.db.QueryRowContext(ctx, query,
		task.TaskID,
		task.UserID,
		task.Title,
		task.Description,
		task.Urgency,
		status,
	).Scan(
		&created.TaskID,
		&created.UserID,
		&created.Title,
		&created.Description,
		&created.DateCreated,
		&created.Urgency,
		&created.Status,
	)

	if err != nil {
		return fmt.Errorf("postgres: insert task error: %w", err)
	}

	return nil
}

func (s *PsqlStorage) DeleteTask(taskID string, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := s.db.ExecContext(ctx,
		`DELETE FROM tasks WHERE task_id = $1 AND user_id = $2`,
		taskID, userID)
	if err != nil {
		return fmt.Errorf("postgres: delete task error: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		return fmt.Errorf("postgres: rows affected: %w or not found task", err)
	}

	return nil
}

func (s *PsqlStorage) InfoTask(taskID string, userID string) (*models.Task, error) {
	fmt.Printf("InfoTask ищем: %q\n", taskID)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `
		SELECT task_id, user_id, title, description, date_created, urgency, status
		FROM tasks
		WHERE user_id = $1 AND task_id = $2`

	var t models.Task
	err := s.db.QueryRowContext(ctx, query, userID, taskID).Scan(
		&t.TaskID,
		&t.UserID,
		&t.Title,
		&t.Description,
		&t.DateCreated,
		&t.Urgency,
		&t.Status,
	)
	if errGLob.Is(err, sql.ErrNoRows) {
		return nil, errors.ErrNotFind
	}
	return &t, nil
}

func (s *PsqlStorage) RepeatInfoTask(taskID, userID string, task models.Task) (*models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE tasks
		SET title       = COALESCE(NULLIF($1, ''), title),
		    description = COALESCE(NULLIF($2, ''), description),
		    urgency     = COALESCE(NULLIF($3, ''), urgency),
		    status      = COALESCE(NULLIF($4, ''), status)
		WHERE task_id = $5 AND user_id = $6
		RETURNING task_id, user_id, title, description, date_created, urgency, status`

	var updated models.Task
	err := s.db.QueryRowContext(ctx, query,
		task.Title,
		task.Description,
		task.Urgency,
		task.Status,
		taskID,
		userID,
	).Scan(
		&updated.TaskID,
		&updated.UserID,
		&updated.Title,
		&updated.Description,
		&updated.DateCreated,
		&updated.Urgency,
		&updated.Status,
	)
	if errGLob.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("postgres: task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: update task error: %w", err)
	}
	return &updated, nil
}
