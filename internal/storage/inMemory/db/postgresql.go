package db

import (
	"awesomeProject4/internal/domain/models"
	"context"
	"database/sql"
	"fmt"
)

type PsqlStorage struct {
	db *sql.DB
}

func NewPsqlStorage(db *sql.DB) *PsqlStorage {
	return &PsqlStorage{db: db}
}

func (s *PsqlStorage) InfoUser(userID string) (*models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (s *PsqlStorage) GetUsers() ([]models.User, error) {
	ctx := context.Background()
	query := `SELECT * FROM users`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var user models.User
	}

	panic("implement me")
}

func (s *PsqlStorage) RegisterUser(user models.User) error {
	ctx := context.Background()
	query := `
		INSERT INTO users (id, email, password) 
		VALUES ($1, $2, $3)
		RETURNING id, email, password;
	`

	var created models.User
	err := s.db.QueryRowContext(ctx, query, user.UUID, user.Email, user.Password).Scan(
		&created.UUID,
		&created.Email,
		&created.Password,
	)
	if err != nil {
		return fmt.Errorf("postgres: upsert error: %w", err)
	}
	return nil

}
func (s *PsqlStorage) LogInNow(user models.UserLogin) (models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (s *PsqlStorage) DeleteUser(userID string) error {
	//TODO implement me
	panic("implement me")
}

func (s *PsqlStorage) RepeatInfoUsers(userID string, user models.User) (*models.User, error) {
	//TODO implement me
	panic("implement me")
}
