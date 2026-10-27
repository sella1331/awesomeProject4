package inMemory

import (
	"awesomeProject4/internal/domain/errors"
	"awesomeProject4/internal/domain/models"
	"fmt"
	"sync"

	"dario.cat/mergo"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UsStorage struct {
	mu    sync.RWMutex
	users map[string]models.User
}

func UserStorage() *UsStorage {
	return &UsStorage{
		users: make(map[string]models.User),
	}
}

func (s *UsStorage) GetUsers() ([]models.User, error) {
	var userList []models.User
	if len(s.users) == 0 {
		return nil, errors.ErrTaskIsEmpty
	}
	for _, user := range s.users {
		userList = append(userList, user)
	}
	return userList, nil
}

func (s *UsStorage) RegisterUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, us := range s.users {
		if us.Email == user.Email {
			fmt.Printf("User с такой почтой уже существует %s. Пропускаем.\n", us.Email)
			return errors.ErrTaskAlreadyExists
		}
	}
	userID := uuid.New().String()
	user.UUID = userID
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("error hashing password: %s\n", err)
		return err
	}
	user.Password = string(hash)
	fmt.Printf("User %s, Password hash: %s\n", user, string(hash))
	s.users[userID] = user
	return nil
}

func (s *UsStorage) DeleteUser(userID string) error {
	if _, exists := s.users[userID]; !exists {
		return errors.ErrNotFind
	}

	delete(s.users, userID)
	return nil
}
func (s *UsStorage) RepeatInfoUsers(userID string, newUser models.User) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, errors.ErrNotFind
	}
	err := mergo.Merge(&user, newUser, mergo.WithOverride)
	if err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("error hashing password: %s\n", err)
		return nil, err
	}
	user.Password = string(hash)
	s.users[userID] = user
	return &user, nil
}

func (s *UsStorage) InfoUser(userID string) (*models.User, error) {
	for _, t := range s.users {
		if t.UUID == userID {
			copyUser := t
			return &copyUser, nil
		}
	}
	return nil, errors.ErrNotFind
}

func (s *UsStorage) LogInNow(user models.UserLogin) (models.User, error) {
	for _, t := range s.users {
		if t.Email == user.Email {
			if err := bcrypt.CompareHashAndPassword(
				[]byte(t.Password), []byte(user.Password),
			); err != nil {
				return models.User{}, errors.ErrInvalidCreds
			}
			return t, nil
		}
	}
	return models.User{}, errors.ErrInvalidCreds
}
