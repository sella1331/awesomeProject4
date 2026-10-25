package server

import (
	"awesomeProject4/cmd/auth"
	"awesomeProject4/internal/domain/models"
	"context"
	//"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TaskRepository interface {
	GetTasks() ([]models.Task, error)
	SaveTask(task models.Task) error
	DeleteTask(taskID string) error
	InfoTask(taskID string) (*models.Task, error)
	RepeatInfoTask(taskID string, task models.Task) (*models.Task, error)
}

type UserRepository interface {
	GetUsers() ([]models.User, error)
	RegisterUser(user models.User) error
	LogInNow(user models.UserLogin) (models.User, error)
	DeleteUser(userID string) error
	RepeatInfoUsers(userID string, user models.User) (*models.User, error)
	InfoUser(userID string) (*models.User, error)
}

type LibraryApI struct {
	httpSrv *http.Server
	tasks   TaskRepository
	users   UserRepository
}

func NewLiraryAPI(tasks TaskRepository, users UserRepository) *LibraryApI {
	httpSrv := &http.Server{Addr: ":8080"}
	LAPI := &LibraryApI{
		httpSrv: httpSrv,
		tasks:   tasks,
		users:   users,
	}
	LAPI.configRoutes()
	return LAPI
}

func (LAPI *LibraryApI) Run() error {
	log.Printf("Listening on port %s ...", LAPI.httpSrv.Addr)
	err := LAPI.httpSrv.ListenAndServe()
	if err != nil {
		LAPI.httpSrv.Shutdown(context.Background())
	}
	return err
}

func (LAPI LibraryApI) configRoutes() {
	router := gin.Default()

	user := router.Group("/users")
	{
		user.GET("", LAPI.allUsers)
		user.POST("/register", LAPI.UserRegister)
		user.PUT("/update/:userID", LAPI.RepeaterLogin)
		user.DELETE("/delete/:userID", LAPI.DeletedUser)
		user.GET("/:userID", LAPI.HowIam)
		user.POST("/login", LAPI.UserLogin)
		user.POST("/refresh", auth.RefreshToken)

		protected := user.Group("", auth.Middleware)
		protected.GET("/profile", LAPI.HowIamToken)
	}

	tasks := router.Group("/tasks")
	{
		tasks.POST("/create", LAPI.createTask)
		tasks.GET("/allTasks", LAPI.allTasks)
		tasks.GET("/:taskID", LAPI.InfoTask)
		tasks.PUT("/:taskID", LAPI.RepeatInfoTask)
		tasks.DELETE("/:taskID", LAPI.deleteTask)
	}
	LAPI.httpSrv.Handler = router
}
