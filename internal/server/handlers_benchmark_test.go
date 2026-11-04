package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sella1331/awesomeProject4.git/internal/domain/models"
)

func BenchmarkLibraryApI_DeletedUser(b *testing.B) {
	gin.SetMode(gin.TestMode)
	mockUsers := NewMockUserRepository(b)
	mockUsers.On("DeleteUser", "user-123").Return(nil)

	LibrarysApI := LibraryApI{users: mockUsers}
	r := gin.New()
	r.DELETE("/users/delete/:userID", LibrarysApI.DeletedUser)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("DELETE", "/users/delete/user-123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkLibraryApI_HowIam(b *testing.B) {
	gin.SetMode(gin.TestMode)
	mockUsers := NewMockUserRepository(b)
	mockUsers.On("InfoUser", "user-123").Return(&models.User{
		UUID:     "user-123",
		Username: "masha",
		Email:    "test@mail.ru",
	}, nil)

	LibrarysApI := LibraryApI{users: mockUsers}
	r := gin.New()
	r.GET("/users/:userID", LibrarysApI.HowIam)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/users/user-123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkLibraryApI_InfoTask(b *testing.B) {
	gin.SetMode(gin.TestMode)
	mockTasks := NewMockTaskRepository(b)
	mockTasks.On("InfoTask", "task-123", "user-123").Return(&models.Task{
		TaskID: "task-123",
		UserID: "user-123",
		Title:  "Лаба",
	}, nil)

	LibrarysApI := LibraryApI{tasks: mockTasks}
	r := gin.New()
	r.Use(func(ctx *gin.Context) {
		ctx.Set("userID", "user-123")
		ctx.Next()
	})
	r.GET("/tasks/:taskID", LibrarysApI.InfoTask)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/tasks/task-123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkLibraryApI_UserLogin(b *testing.B) {
	gin.SetMode(gin.TestMode)
	mockUsers := NewMockUserRepository(b)
	mockUsers.On("LogInNow", models.UserLogin{
		Email:    "test@mail.ru",
		Password: "Password123!",
	}).Return(models.User{
		UUID:     "user-123",
		Username: "masha",
		Email:    "test@mail.ru",
	}, nil)

	LibrarysApI := LibraryApI{users: mockUsers}
	r := gin.New()
	r.POST("/users/login", LibrarysApI.UserLogin)

	test := `{"email":"test@mail.ru","password":"Password123!"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/users/login", strings.NewReader(test))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}
