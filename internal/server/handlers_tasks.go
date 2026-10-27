package server

import (
	"awesomeProject4/internal/domain/errors"
	"awesomeProject4/internal/domain/models"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (LAPI *LibraryApI) allTasks(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		fmt.Printf(userID)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not found userID"})
		return
	}
	tasks, err := LAPI.tasks.GetTasks(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, tasks)
}

func (LAPI *LibraryApI) createTask(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	fmt.Printf("USERIDDD %s", userID)

	if userID == "" {
		fmt.Printf(userID)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not found userID"})
		return
	}
	var task models.Task
	err := ctx.ShouldBindBodyWithJSON(&task)
	log.Println(task)
	if err != nil {
		log.Println("Error while binding body", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task.UserID = userID
	if err := LAPI.tasks.SaveTask(task); err != nil {
		if err == errors.ErrTaskAlreadyExists {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "Задача с таким содержимым уже существует",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, task)
}

func (LAPI *LibraryApI) deleteTask(ctx *gin.Context) {
	taskID := ctx.Param("taskID")
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	err := LAPI.tasks.DeleteTask(taskID, userID)
	if err != nil {
		if err == errors.ErrNotFind {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "задача не найдена"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Задача успешно удалена", "taskID": taskID})
}

func (LAPI *LibraryApI) InfoTask(ctx *gin.Context) {
	taskID := ctx.Param("taskID")
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	task, err := LAPI.tasks.InfoTask(taskID, userID)
	if err != nil {
		if err == errors.ErrNotFind {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "your task",
		"task":    task,
		"user_id": userID,
	})
}

func (LAPI *LibraryApI) RepeatInfoTask(ctx *gin.Context) {
	taskID := ctx.Param("taskID")
	userID := ctx.GetString("userID")
	var newTask models.Task
	err := ctx.ShouldBindBodyWithJSON(&newTask)
	log.Println(newTask)
	if err != nil {
		log.Println("Error while binding body", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := LAPI.tasks.RepeatInfoTask(taskID, userID, newTask)
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
