package server

import (
	"context"
	err "errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/sella1331/awesomeProject4.git/internal/domain/errors"
	"github.com/sella1331/awesomeProject4.git/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gin-gonic/gin"
)

func TestLibraryApI_DeletedUser(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		userIDParam    string
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "deleted user",
			userIDParam:    "user-123",
			method:         "DELETE",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Юзер успешно удален","userID":"user-123"}`,
		},
		{
			name:           "not Found",
			userIDParam:    "user-999",
			method:         "DELETE",
			mockError:      errors.ErrNotFind,
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"пользователь не найден"}`,
		},
		{
			name:           "server eror",
			userIDParam:    "user-456",
			method:         "DELETE",
			mockError:      err.New("database connection failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database connection failed"}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUsers := NewMockUserRepository(t)
			mockUsers.
				On("DeleteUser", tc.userIDParam).
				Return(tc.mockError).
				Once()

			LibrarysApI := LibraryApI{
				domain: "test.org",
				users:  mockUsers,
			}

			r := gin.New()
			gin.SetMode(gin.TestMode)
			r.DELETE("/:userID", LibrarysApI.DeletedUser)

			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			req, err := http.NewRequest(tc.method, httpSrv.URL+"/"+tc.userIDParam, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			assert.JSONEq(t, tc.expectedBody, string(body))
		})
	}
}

func TestLibraryApI_HowIam(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		userIDParam    string
		mockUser       *models.User
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "success",
			userIDParam: "123",
			method:      "GET",
			mockUser: &models.User{
				UUID:     "123",
				Username: "masha",
				Email:    "test@mail.ru",
				Password: "Password123!",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"uuid":"123","username":"masha","email":"test@mail.ru","password":"Password123!"}`,
		},
		{
			name:           "not found",
			userIDParam:    "999",
			method:         "GET",
			mockUser:       nil,
			mockError:      err.New("пользователь не найден"),
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"пользователь не найден"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUsers := NewMockUserRepository(t)
			mockUsers.
				On("InfoUser", tc.userIDParam).
				Return(tc.mockUser, tc.mockError).
				Once()

			LibrarysApI := LibraryApI{
				domain: "test.org",
				users:  mockUsers,
			}

			r := gin.New()
			r.GET("/users/:userID", LibrarysApI.HowIam)
			gin.SetMode(gin.TestMode)
			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			req, err := http.NewRequest(tc.method, httpSrv.URL+"/users/"+tc.userIDParam, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			assert.JSONEq(t, tc.expectedBody, string(body))
		})
	}
}

func TestLibraryApI_InfoTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ttime := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name           string
		method         string
		taskIDParam    string
		userIDContext  string
		mockTask       *models.Task
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:          "Success",
			method:        "GET",
			taskIDParam:   "task-123",
			userIDContext: "user-123",
			mockTask: &models.Task{
				TaskID:      "task-123",
				UserID:      "user-123",
				Title:       "HGGjg",
				Description: "egweg",
				DateCreated: ttime,
				Urgency:     "high",
				Status:      "in_progress",
				Deleted:     false,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"your task","task":{"task_id":"task-123","user_id":"user-123","title":"HGGjg","description":"egweg","date_created":"2026-11-03T12:00:00Z","urgency":"high","status":"in_progress","deleted":false},"user_id":"user-123"}`,
		},
		{
			name:           "Unauthorized",
			method:         "GET",
			taskIDParam:    "task-123",
			userIDContext:  "",
			mockTask:       nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized"}`,
		},
		{
			name:           "Not Found",
			method:         "GET",
			taskIDParam:    "task-999",
			userIDContext:  "user-123",
			mockTask:       nil,
			mockError:      errors.ErrNotFind,
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"` + errors.ErrNotFind.Error() + `"}`,
		},
		{
			name:           "server error",
			method:         "GET",
			taskIDParam:    "task-456",
			userIDContext:  "user-123",
			mockTask:       nil,
			mockError:      err.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTasks := NewMockTaskRepository(t)
			if tc.userIDContext != "" {
				mockTasks.
					On("InfoTask", tc.taskIDParam, tc.userIDContext).
					Return(tc.mockTask, tc.mockError).
					Once()
			}

			LibrarysApI := LibraryApI{
				domain: "test.org",
				tasks:  mockTasks,
			}

			r := gin.New()
			r.Use(func(ctx *gin.Context) {
				if tc.userIDContext != "" {
					ctx.Set("userID", tc.userIDContext)
				}
				ctx.Next()
			})
			r.GET("/tasks/:taskID", LibrarysApI.InfoTask)

			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			req, err := http.NewRequest(tc.method, httpSrv.URL+"/tasks/"+tc.taskIDParam, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			assert.JSONEq(t, tc.expectedBody, string(body))
		})
	}
}

func TestLibraryApI_RepeatInfoTask(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fixedTime := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)

	inputTask := models.Task{
		Title:       "Обновленная задача",
		Description: "Новое описание",
		Urgency:     "medium",
		Status:      "in_progress",
	}

	tests := []struct {
		name           string
		method         string
		taskIDParam    string
		userIDContext  string
		requestBody    string
		shouldCallMock bool
		mockInputTask  models.Task
		mockReturnTask *models.Task
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success Update Task",
			method:         "PUT",
			taskIDParam:    "task-123",
			userIDContext:  "user-123",
			requestBody:    `{"title":"Обновленная задача","description":"Новое описание","urgency":"medium","status":"in_progress"}`,
			shouldCallMock: true,
			mockInputTask:  inputTask,
			mockReturnTask: &models.Task{
				TaskID:      "task-123",
				UserID:      "user-123",
				Title:       "Обновленная задача",
				Description: "Новое описание",
				DateCreated: fixedTime,
				Urgency:     "medium",
				Status:      "in_progress",
				Deleted:     false,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"task_id":"task-123","user_id":"user-123","title":"Обновленная задача","description":"Новое описание","date_created":"2026-11-03T12:00:00Z","urgency":"medium","status":"in_progress","deleted":false}`,
		},
		{
			name:           "Bad Request Invalid JSON",
			method:         "PUT",
			taskIDParam:    "task-123",
			userIDContext:  "user-123",
			requestBody:    `{"title": "незакрытый json"`,
			shouldCallMock: false,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"unexpected EOF"}`,
		},
		{
			name:           "Task Not Found",
			method:         "PUT",
			taskIDParam:    "task-999",
			userIDContext:  "user-123",
			requestBody:    `{"title":"Обновленная задача","description":"Новое описание","urgency":"medium","status":"in_progress"}`,
			shouldCallMock: true,
			mockInputTask:  inputTask,
			mockReturnTask: nil,
			mockError:      errors.ErrNotFind,
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"` + errors.ErrNotFind.Error() + `"}`,
		},
		{
			name:           "Internal Server Error",
			method:         "PUT",
			taskIDParam:    "task-456",
			userIDContext:  "user-123",
			requestBody:    `{"title":"Обновленная задача","description":"Новое описание","urgency":"medium","status":"in_progress"}`,
			shouldCallMock: true,
			mockInputTask:  inputTask,
			mockReturnTask: nil,
			mockError:      err.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTasks := NewMockTaskRepository(t)
			if tc.shouldCallMock {
				mockTasks.
					On("RepeatInfoTask", tc.taskIDParam, tc.userIDContext, tc.mockInputTask).
					Return(tc.mockReturnTask, tc.mockError).
					Once()
			}

			LibrarysApI := LibraryApI{
				domain: "test.org",
				tasks:  mockTasks,
			}

			r := gin.New()
			r.Use(func(ctx *gin.Context) {
				if tc.userIDContext != "" {
					ctx.Set("userID", tc.userIDContext)
				}
				ctx.Next()
			})
			r.PUT("/tasks/:taskID", LibrarysApI.RepeatInfoTask)

			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			req, err := http.NewRequest(tc.method, httpSrv.URL+"/tasks/"+tc.taskIDParam, strings.NewReader(tc.requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			assert.JSONEq(t, tc.expectedBody, string(body))
		})
	}
}

func TestLibraryApI_RepeaterLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	inputUser := models.User{
		Username: "masha",
		Email:    "test@mail.ru",
		Password: "Password123!",
	}

	tests := []struct {
		name           string
		method         string
		userIDParam    string
		setContext     bool
		userIDContext  any
		requestBody    string
		shouldCallMock bool
		mockInputUser  models.User
		mockReturnUser *models.User
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "success",
			method:         "PUT",
			userIDParam:    "user-123",
			setContext:     true,
			userIDContext:  "user-123",
			requestBody:    `{"username":"masha","email":"test@mail.ru","password":"Password123!"}`,
			shouldCallMock: true,
			mockInputUser:  inputUser,
			mockReturnUser: &models.User{
				UUID:     "user-123",
				Username: "masha",
				Email:    "test@mail.ru",
				Password: "Password123!",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"uuid":"user-123","username":"masha","email":"test@mail.ru","password":"Password123!"}`,
		},
		{
			name:           "Unauthorized",
			method:         "PUT",
			userIDParam:    "user-123",
			setContext:     false,
			userIDContext:  nil,
			requestBody:    `{"username":"masha","email":"test@mail.ru","password":"Password123!"}`,
			shouldCallMock: false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"forbidden"}`,
		},
		{
			name:           "invalid UserID type",
			method:         "PUT",
			userIDParam:    "user-123",
			setContext:     true,
			userIDContext:  12345,
			requestBody:    `{"username":"masha","email":"test@mail.ru","password":"Password123!"}`,
			shouldCallMock: false,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"invalid user id type"}`,
		},
		{
			name:           "Bad JSON",
			method:         "PUT",
			userIDParam:    "user-123",
			setContext:     true,
			userIDContext:  "user-123",
			requestBody:    `{"username": "незакрытый json"`,
			shouldCallMock: false,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"unexpected EOF"}`,
		},
		{
			name:           "Not Found",
			method:         "PUT",
			userIDParam:    "user-999",
			setContext:     true,
			userIDContext:  "user-999",
			requestBody:    `{"username":"masha","email":"test@mail.ru","password":"Password123!"}`,
			shouldCallMock: true,
			mockInputUser:  inputUser,
			mockReturnUser: nil,
			mockError:      errors.ErrNotFind,
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"` + errors.ErrNotFind.Error() + `"}`,
		},
		{
			name:           "server eror",
			method:         "PUT",
			userIDParam:    "user-456",
			setContext:     true,
			userIDContext:  "user-456",
			requestBody:    `{"username":"masha","email":"test@mail.ru","password":"Password123!"}`,
			shouldCallMock: true,
			mockInputUser:  inputUser,
			mockReturnUser: nil,
			mockError:      err.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUsers := NewMockUserRepository(t)
			if tc.shouldCallMock {
				userID, ok := tc.userIDContext.(string)
				if !ok {
					t.Fatalf("failed userIDContext %T", tc.userIDContext)
				}

				mockUsers.
					On("RepeatInfoUsers", userID, tc.mockInputUser).
					Return(tc.mockReturnUser, tc.mockError).
					Once()
			}
			LibrarysApI := LibraryApI{
				domain: "test.org",
				users:  mockUsers,
			}

			r := gin.New()
			r.Use(func(ctx *gin.Context) {
				if tc.setContext {
					ctx.Set("userID", tc.userIDContext)
				}
				ctx.Next()
			})
			r.PUT("/users/update/:userID", LibrarysApI.RepeaterLogin)

			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			req, err := http.NewRequest(tc.method, httpSrv.URL+"/users/update/"+tc.userIDParam, strings.NewReader(tc.requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			assert.JSONEq(t, tc.expectedBody, string(body))
		})
	}
}

func TestLibraryApI_Run(t *testing.T) {
	tests := []struct {
		name        string
		addr        string
		expectedErr error
	}{
		{
			name:        "success",
			addr:        "127.0.0.1:0",
			expectedErr: http.ErrServerClosed,
		},
		{
			name:        "invalid Adress Error",
			addr:        "invalid addr",
			expectedErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			LibrarysApI := LibraryApI{
				httpSrv: &http.Server{
					Addr: tc.addr,
				},
			}

			errCh := make(chan error, 1)
			go func() {
				errCh <- LibrarysApI.Run()
			}()

			if tc.expectedErr == http.ErrServerClosed {
				time.Sleep(50 * time.Millisecond)
				err := LibrarysApI.Shutdown(context.Background())
				require.NoError(t, err)

				runErr := <-errCh
				assert.ErrorIs(t, runErr, tc.expectedErr)
			} else {
				runErr := <-errCh
				assert.Error(t, runErr)
			}
		})
	}
}

func TestLibraryApI_Shutdown(t *testing.T) {
	tests := []struct {
		name      string
		cancelCtx bool
		wantErr   bool
	}{
		{
			name:      "Success Shutdown",
			cancelCtx: false,
			wantErr:   false,
		},
		{
			name:      "Shutdown and Ctx",
			cancelCtx: true,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(200 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			})

			LibrarysApI := LibraryApI{
				httpSrv: &http.Server{
					Addr:    "127.0.0.1:8899",
					Handler: mux,
				},
			}

			go func() {
				_ = LibrarysApI.Run()
			}()

			time.Sleep(50 * time.Millisecond)

			ctx := context.Background()
			if tc.cancelCtx {
				go func() {
					resp, err := http.Get("http://127.0.0.1:8899/slow")
					if err != nil {
						return
					}
					defer func() {
						_ = resp.Body.Close()
					}()
				}()
				time.Sleep(20 * time.Millisecond)

				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			err := LibrarysApI.Shutdown(ctx)
			if tc.wantErr {
				assert.Error(t, err)
				_ = LibrarysApI.httpSrv.Close()
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLibraryApI_UserLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	inputLogin := models.UserLogin{
		Email:    "test@mail.ru",
		Password: "Password123!",
	}

	tests := []struct {
		name           string
		method         string
		requestBody    string
		shouldCallMock bool
		mockInput      models.UserLogin
		mockReturnUser models.User
		mockError      error
		expectedStatus int
		expectedBody   string
		checkToken     bool
	}{
		{
			name:           "success",
			method:         "POST",
			requestBody:    `{"email":"test@mail.ru","password":"Password123!"}`,
			shouldCallMock: true,
			mockInput:      inputLogin,
			mockReturnUser: models.User{
				UUID:     "user-123",
				Username: "student1",
				Email:    "test@mail.ru",
				Password: "Password123!",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			checkToken:     true,
		},
		{
			name:           "Bad JSON",
			method:         "POST",
			requestBody:    `{"email": "незакрытый json"`,
			shouldCallMock: false,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"unexpected EOF"}`,
		},
		{
			name:           "validation Email",
			method:         "POST",
			requestBody:    `{"email":"not-an-email","password":"Password123!"}`,
			shouldCallMock: false,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:           "invalid data",
			method:         "POST",
			requestBody:    `{"email":"test@mail.ru","password":"Password123!"}`,
			shouldCallMock: true,
			mockInput:      inputLogin,
			mockReturnUser: models.User{},
			mockError:      errors.ErrInvalidCreds,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"invalid credentials"}`,
		},
		{
			name:           "Error Repo",
			method:         "POST",
			requestBody:    `{"email":"test@mail.ru","password":"Password123!"}`,
			shouldCallMock: true,
			mockInput:      inputLogin,
			mockReturnUser: models.User{},
			mockError:      err.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
		{
			name:           "no ID",
			method:         "POST",
			requestBody:    `{"email":"test@mail.ru","password":"Password123!"}`,
			shouldCallMock: true,
			mockInput:      inputLogin,
			mockReturnUser: models.User{
				UUID:  "",
				Email: "test@mail.ru",
			},
			mockError:      nil,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"user has no id"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUsers := NewMockUserRepository(t)
			if tc.shouldCallMock {
				mockUsers.
					On("LogInNow", tc.mockInput).
					Return(tc.mockReturnUser, tc.mockError).
					Once()
			}

			LibrarysApI := LibraryApI{
				domain: "test.org",
				users:  mockUsers,
			}

			r := gin.New()
			r.POST("/users/login", LibrarysApI.UserLogin)

			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			req, err := http.NewRequest(tc.method, httpSrv.URL+"/users/login", strings.NewReader(tc.requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.checkToken {
				var respData map[string]any
				err := json.Unmarshal(body, &respData)
				require.NoError(t, err)
				assert.NotEmpty(t, respData["token"])
				assert.NotEmpty(t, resp.Cookies())
			} else if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, string(body))
			} else {
				assert.Contains(t, string(body), `"error"`)
			}
		})
	}
}

func TestLibraryApI_UserRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validUser := models.User{
		Username: "masha",
		Email:    "test@mail.ru",
		Password: "Password123!",
	}

	tests := []struct {
		name           string
		method         string
		requestBody    string
		shouldCallMock bool
		mockUser       models.User
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "success",
			method:         "POST",
			requestBody:    `{"username":"masha","email":"test@mail.ru","password":"Password123!"}`,
			shouldCallMock: true,
			mockUser:       validUser,
			mockError:      nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   ``,
		},
		{
			name:           "Bad JSON",
			method:         "POST",
			requestBody:    `{"username": "незакрытый json"`,
			shouldCallMock: false,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"unexpected EOF"}`,
		},
		{
			name:           "validation",
			method:         "POST",
			requestBody:    `{"username":"","email":"invalid","password":"123"}`,
			shouldCallMock: false,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "",
		},
		{
			name:           "User Already Exists",
			method:         "POST",
			requestBody:    `{"username":"masha","email":"test@mail.ru","password":"Password123!"}`,
			shouldCallMock: true,
			mockUser:       validUser,
			mockError:      errors.ErrUserAlreadyExists,
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"пользователь с таким email уже существует"}`,
		},
		{
			name:           "Internal Server Error",
			method:         "POST",
			requestBody:    `{"username":"masha","email":"test@mail.ru","password":"Password123!"}`,
			shouldCallMock: true,
			mockUser:       validUser,
			mockError:      err.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUsers := NewMockUserRepository(t)
			if tc.shouldCallMock {
				mockUsers.
					On("RegisterUser", tc.mockUser).
					Return(tc.mockError).
					Once()
			}

			LibrarysApI := LibraryApI{
				domain: "test.org",
				users:  mockUsers,
			}

			r := gin.New()
			r.POST("/users/register", LibrarysApI.UserRegister)

			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			req, err := http.NewRequest(tc.method, httpSrv.URL+"/users/register", strings.NewReader(tc.requestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, string(body))
			}
		})
	}
}
func TestLibraryApI_allTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fixedTime := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		method         string
		userIDContext  string
		shouldCallMock bool
		mockTasks      []models.Task
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "success",
			method:         "GET",
			userIDContext:  "user-123",
			shouldCallMock: true,
			mockTasks: []models.Task{
				{
					TaskID:      "task-1",
					UserID:      "user-123",
					Title:       "Лаба",
					Description: "gggggg",
					DateCreated: fixedTime,
					Urgency:     "high",
					Status:      "in_progress",
					Deleted:     false,
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"task_id":"task-1","user_id":"user-123","title":"Лаба","description":"gggggg","date_created":"2026-11-03T12:00:00Z","urgency":"high","status":"in_progress","deleted":false}]`,
		},
		{
			name:           "Unauthorized",
			method:         "GET",
			userIDContext:  "",
			shouldCallMock: false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"not found userID"}`,
		},
		{
			name:           "server error",
			method:         "GET",
			userIDContext:  "user-123",
			shouldCallMock: true,
			mockTasks:      nil,
			mockError:      err.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTasks := NewMockTaskRepository(t)
			if tc.shouldCallMock {
				mockTasks.
					On("GetTasks", tc.userIDContext).
					Return(tc.mockTasks, tc.mockError).
					Once()
			}

			LibrarysApI := LibraryApI{
				domain: "test.org",
				tasks:  mockTasks,
			}

			r := gin.New()
			r.Use(func(ctx *gin.Context) {
				if tc.userIDContext != "" {
					ctx.Set("userID", tc.userIDContext)
				}
				ctx.Next()
			})
			r.GET("/tasks/allTasks", LibrarysApI.allTasks)

			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			req, err := http.NewRequest(tc.method, httpSrv.URL+"/tasks/allTasks", nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			assert.JSONEq(t, tc.expectedBody, string(body))
		})
	}
}

func TestLibraryApI_allUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		shouldCallMock bool
		mockUsers      []models.User
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "success",
			method:         "GET",
			shouldCallMock: true,
			mockUsers: []models.User{
				{
					UUID:     "user-1",
					Username: "masha",
					Email:    "test1@mail.ru",
					Password: "Password123!",
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"uuid":"user-1","username":"masha","email":"test1@mail.ru","password":"Password123!"}]`,
		},
		{
			name:           "server Error",
			method:         "GET",
			shouldCallMock: true,
			mockUsers:      nil,
			mockError:      err.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockUsersRepo := NewMockUserRepository(t)
			if tc.shouldCallMock {
				mockUsersRepo.
					On("GetUsers").
					Return(tc.mockUsers, tc.mockError).
					Once()
			}

			LibrarysApI := LibraryApI{
				domain: "test.org",
				users:  mockUsersRepo,
			}

			r := gin.New()
			r.GET("/users", LibrarysApI.allUsers)

			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			req, err := http.NewRequest(tc.method, httpSrv.URL+"/users", nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			assert.JSONEq(t, tc.expectedBody, string(body))
		})
	}
}

func TestLibraryApI_deleteTask(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		taskIDParam    string
		userIDContext  string
		shouldCallMock bool
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "success",
			method:         "DELETE",
			taskIDParam:    "task-123",
			userIDContext:  "user-123",
			shouldCallMock: true,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Задача успешно удалена","taskID":"task-123"}`,
		},
		{
			name:           "Unauthorized",
			method:         "DELETE",
			taskIDParam:    "task-123",
			userIDContext:  "",
			shouldCallMock: false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized"}`,
		},
		{
			name:           "Not Found",
			method:         "DELETE",
			taskIDParam:    "task-999",
			userIDContext:  "user-123",
			shouldCallMock: true,
			mockError:      errors.ErrNotFind,
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"задача не найдена"}`,
		},
		{
			name:           "server eror",
			method:         "DELETE",
			taskIDParam:    "task-456",
			userIDContext:  "user-123",
			shouldCallMock: true,
			mockError:      err.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTasks := NewMockTaskRepository(t)
			if tc.shouldCallMock {
				mockTasks.
					On("DeleteTask", tc.taskIDParam, tc.userIDContext).
					Return(tc.mockError).
					Once()
			}

			LibrarysApI := LibraryApI{
				domain: "test.org",
				tasks:  mockTasks,
			}

			r := gin.New()
			r.Use(func(ctx *gin.Context) {
				if tc.userIDContext != "" {
					ctx.Set("userID", tc.userIDContext)
				}
				ctx.Next()
			})
			r.DELETE("/tasks/:taskID", LibrarysApI.deleteTask)

			httpSrv := httptest.NewServer(r)
			defer httpSrv.Close()

			req, err := http.NewRequest(tc.method, httpSrv.URL+"/tasks/"+tc.taskIDParam, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			assert.JSONEq(t, tc.expectedBody, string(body))
		})
	}
}
func TestNewLiraryAPI(t *testing.T) {
	mockTasks := NewMockTaskRepository(t)
	mockUsers := NewMockUserRepository(t)

	api := NewLiraryAPI(mockTasks, mockUsers)

	assert.NotNil(t, api)
	assert.Equal(t, mockTasks, api.tasks)
	assert.Equal(t, mockUsers, api.users)
	assert.NotNil(t, api.httpSrv)
	assert.NotNil(t, api.httpSrv.Handler)
}
