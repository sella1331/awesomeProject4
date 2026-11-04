package db

import (
	"context"
	"database/sql"
	errGLob "errors"
	"fmt"
	"log"
	"time"

	"github.com/sella1331/awesomeProject4.git/internal/domain/errors"
	"github.com/sella1331/awesomeProject4.git/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *PsqlStorage) GetTasks(userID string) ([]models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT task_id, user_id, title, description, date_created, urgency, status
		FROM tasks
		WHERE user_id = $1 AND deleted = false
		ORDER BY date_created DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("postgres: get tasks: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

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
		`UPDATE tasks SET deleted = true WHERE task_id = $1 AND user_id = $2 AND deleted = false`,
		taskID, userID)

	if err != nil {
		return fmt.Errorf("postgres: delete task error: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: rows affected: %w", err)
	}
	if n == 0 {
		return errors.ErrNotFind
	}

	s.Chan <- struct{}{}
	fmt.Printf("send  channel: %s\n", taskID)

	return nil
}

func (s *PsqlStorage) InfoTask(taskID string, userID string) (*models.Task, error) {
	fmt.Printf("InfoTask ищем: %q\n", taskID)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `
		SELECT task_id, user_id, title, description, date_created, urgency, status
		FROM tasks
		WHERE user_id = $1 AND task_id = $2 AND deleted = false`

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
	if err != nil {
		return nil, err
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
		WHERE task_id = $5 AND user_id = $6 AND deleted = false
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

func (s *PsqlStorage) CountTasks() {
	timer := time.NewTicker(100 * time.Millisecond)
	defer timer.Stop()

	for range timer.C {
		if len(s.Chan) == cap(s.Chan) {
			for i := 0; i < cap(s.Chan); i++ {
				fmt.Println("task full deleted")
				<-s.Chan
			}
			if err := s.FullDeletedTasks(); err != nil {
				log.Printf("error deleted tasks: %v\n", err)
			}
		}
	}
}

func (s *PsqlStorage) FullDeletedTasks() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("transaction error: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.ExecContext(ctx, `DELETE FROM tasks WHERE deleted = true`)
	if err != nil {
		return fmt.Errorf("delete tasks error: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction error: %w", err)
	}

	log.Println("10 tasks deleted")
	return nil
}
