package server

import (
	"awesomeProject4/internal/domain/errors"
	"awesomeProject4/internal/domain/models"
	"awesomeProject4/internal/storage/inMemory"
	"context"
	"fmt"

	//"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	GetTasks() ([]models.Task, error)
	SaveTask(task models.Task) error
	DeleteTask(taskID string) error
	InfoTask(taskID string) (*models.Task, error)
	RepeatInfoTask(taskID string, task models.Task) (*models.Task, error)

	RegisterUser(user models.User) error
	LogInNow(user models.UserLogin) (models.User, error)
	DeleteUser(userID string) error
	RepeatInfoUsers(userID string, user models.User) (*models.User, error)
	InfoUser(userID string) (*models.User, error)
}

type LibraryApI struct {
	httpSrv *http.Server
	repo    Repository
}

func NewLiraryAPI(repo *inMemory.Storage) *LibraryApI {
	httpSrv := http.Server{Addr: ":8080"} // TODO edit to config dinamic port
	LAPI := LibraryApI{
		httpSrv: &httpSrv,
		repo:    repo,
	}
	LAPI.configRoutes()
	return &LAPI
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

		user.POST("/register", LAPI.UserRegister)
		user.POST("/login", LAPI.UserLogin)
		user.PUT("/update/:userID", LAPI.RepeaterLogin)
		user.DELETE("/delete/:userID", LAPI.DeletedUser)
		user.GET("/:userID", LAPI.HowIam)
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

func (LAPI *LibraryApI) allTasks(ctx *gin.Context) {

	tasks, err := LAPI.repo.GetTasks()
	if err != nil {
		ctx.JSON(http.StatusNoContent, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, tasks)
}

func (LAPI *LibraryApI) createTask(ctx *gin.Context) {
	task := models.Task{}
	err := ctx.ShouldBindBodyWithJSON(&task)
	log.Println(task)
	if err != nil {
		log.Println("Error while binding body", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := LAPI.repo.SaveTask(task); err != nil {
		if err == errors.ErrTaskAlreadyExists {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "Задача с таким содержимым уже существует",
			})
			return
		}
	}
	LAPI.repo.SaveTask(task)
	ctx.JSON(http.StatusCreated, task)
}

func (LAPI *LibraryApI) deleteTask(ctx *gin.Context) {
	taskID := ctx.Param("taskID")
	err := LAPI.repo.DeleteTask(taskID)
	if err != nil {
		if err == errors.ErrNotFind {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Задача успешно удалена", "taskID": taskID})
}

func (LAPI *LibraryApI) InfoTask(ctx *gin.Context) {
	taskID := ctx.Param("taskID")
	task, err := LAPI.repo.InfoTask(taskID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "your task:", "task": task})
}

func (LAPI *LibraryApI) RepeatInfoTask(ctx *gin.Context) {
	taskID := ctx.Param("taskID")
	var newTask models.Task
	err := ctx.ShouldBindBodyWithJSON(&newTask)
	log.Println(newTask)
	if err != nil {
		log.Println("Error while binding body", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := LAPI.repo.RepeatInfoTask(taskID, newTask)
	if err != nil {
		if err == errors.ErrNotFind {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, task)
}

func (LAPI *LibraryApI) UserRegister(ctx *gin.Context) {
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := LAPI.repo.RegisterUser(user); err != nil {
		if err == errors.ErrUserAlreadyExists {
			ctx.JSON(http.StatusConflict, gin.H{"error": "пользователь с таким email уже существует"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

func (LAPI *LibraryApI) UserLogin(ctx *gin.Context) {
	var user models.UserLogin
	err := ctx.ShouldBindBodyWithJSON(&user)
	if err != nil {
		log.Println("Error while binding body", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	valid := validator.New()
	err = valid.Struct(user)
	if err != nil {
		log.Println("Failed validation", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	repoUser, err := LAPI.repo.LogInNow(user)
	if err != nil {
		if err == errors.ErrInvalidCreds {
			log.Printf("Error while logging in: %s", err.Error())
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("check password: %s\n and input %s\n", repoUser.Password, user.Password)
	err = bcrypt.CompareHashAndPassword(
		[]byte(repoUser.Password),
		[]byte(user.Password))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	ctx.JSON(http.StatusOK, repoUser)
}

func (LAPI *LibraryApI) DeletedUser(ctx *gin.Context) {
	userID := ctx.Param("userID")

	err := LAPI.repo.DeleteUser(userID)
	if err != nil {
		if err == errors.ErrNotFind {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Задача успешно удалена", "userID": userID})
}

func (LAPI *LibraryApI) RepeaterLogin(ctx *gin.Context) {
	userID := ctx.Param("userID")
	var newUser models.User
	err := ctx.ShouldBindBodyWithJSON(&newUser)
	log.Println(newUser)
	if err != nil {
		log.Println("Error while binding body", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	user, err := LAPI.repo.RepeatInfoUsers(userID, newUser)
	if err != nil {
		if err == errors.ErrNotFind {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (LAPI *LibraryApI) HowIam(ctx *gin.Context) {
	userID := ctx.Param("userID")
	LAPI.repo.InfoUser(userID)
	err := errors.ErrNotFind
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, userID)
}
