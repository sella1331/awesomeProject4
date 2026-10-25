package db

import (
	"awesomeProject4/internal/domain/errors"
	"awesomeProject4/internal/domain/models"
	"context"
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type PsqlStorage struct {
	db *sql.DB
}

func NewPsqlStorage(db *sql.DB) *PsqlStorage {
	return &PsqlStorage{db: db}
}

func (s *PsqlStorage) InfoUser(userID string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, Username, Email, Password
	FROM users
	WHERE id = $1;`
	fmt.Println(query, userID)
	var user models.User
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UUID,
		&user.Username,
		&user.Email,
		&user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *PsqlStorage) GetUsers() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, Username, Email, Password FROM users`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var user models.User
		rows.Scan(&user.UUID, &user.Username, &user.Email, &user.Password)

		users = append(users, user)
	}
	return users, nil
}

func (s *PsqlStorage) RegisterUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `
		INSERT INTO users (username, email, password) 
		VALUES ($1, $2, $3)
		RETURNING id, username, email, password;
	`

	var created models.User
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("error hashing password: %s\n", err)
		return err
	}
	user.Password = string(hash)
	err = s.db.QueryRowContext(ctx, query, user.Username, user.Email, user.Password).Scan(
		&created.UUID,
		&created.Username,
		&created.Email,
		&created.Password,
	)
	if err != nil {
		return fmt.Errorf("postgres: upsert error: %w", err)
	}
	return nil
}

func (s *PsqlStorage) LogInNow(user models.UserLogin) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
        SELECT id, username, email, password
        FROM users
        WHERE email = $1;
    `
	fmt.Println(query)
	var u models.User
	err := s.db.QueryRowContext(ctx, query, user.Email).Scan(
		&u.UUID,
		&u.Username,
		&u.Email,
		&u.Password,
	)
	if err != nil {
		return models.User{}, err
	}
	fmt.Println(u)
	if err := bcrypt.CompareHashAndPassword(
		[]byte(u.Password), []byte(user.Password),
	); err != nil {
		return models.User{}, errors.ErrInvalidCreds
	}
	return u, nil
}

func (s *PsqlStorage) DeleteUser(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `DELETE FROM users
	WHERE id = $1;`
	fmt.Println(query, userID)
	result, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("postgres: delete user: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: rows affected: %w", err)
	}
	if n == 0 {
		return errors.ErrNotFind // нечего удалять → 404
	}
	return nil
}

func (s *PsqlStorage) RepeatInfoUsers(userID string, user models.User) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if user.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("error hashing password: %s\n", err)
			return &user, err
		}
		user.Password = string(hash)
	}

	query := `
		UPDATE users 
		SET username = COALESCE(NULLIF($1, ''), username),
		    email    = COALESCE(NULLIF($2, ''), email),
		    password = COALESCE(NULLIF($3, ''), password)
		WHERE id = $4
		RETURNING id, username, email, password
	`

	err := s.db.QueryRowContext(ctx, query,
		user.Username,
		user.Email,
		user.Password,
		userID,
	).Scan(
		&user.UUID,
		&user.Username,
		&user.Email,
		&user.Password,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
