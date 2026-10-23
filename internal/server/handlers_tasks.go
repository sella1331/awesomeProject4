package server

import (
	"awesomeProject4/internal/domain/errors"
	"awesomeProject4/internal/domain/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (LAPI *LibraryApI) allTasks(ctx *gin.Context) {

	tasks, err := LAPI.tasks.GetTasks()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	err := LAPI.tasks.DeleteTask(taskID)
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
	task, err := LAPI.tasks.InfoTask(taskID)
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
	task, err := LAPI.tasks.RepeatInfoTask(taskID, newTask)
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
