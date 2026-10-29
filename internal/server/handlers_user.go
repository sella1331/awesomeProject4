package server

import (
	"awesomeProject4/cmd/auth"
	"awesomeProject4/internal/domain/errors"
	"awesomeProject4/internal/domain/models"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

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

	if err := LAPI.users.RegisterUser(user); err != nil {
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
	repoUser, err := LAPI.users.LogInNow(user)
	if err != nil {
		if err == errors.ErrInvalidCreds {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if repoUser.UUID == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "user has no id"})
		return
	}
	token, err := auth.GenerateAccessToken(repoUser.UUID)
	fmt.Println(token)
	if err != nil {
		if err == errors.ErrInvalidCreds {
			log.Printf("Error while logging in: %s", err.Error())
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	refreshToken, err := auth.GenerateRefreshToken(repoUser.UUID)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("check password: %s\n and input %s\n", repoUser.Password, user.Password)
	ctx.SetCookie("auth_token",
		token,
		3600,
		"/",
		"localhost",
		false,
		true)
	ctx.SetCookie("refresh_token",
		refreshToken,
		60*60*24,
		"/",
		"localhost",
		false,
		true)
	fmt.Printf("refresh token: %s\n", refreshToken)

	ctx.JSON(http.StatusOK, gin.H{
		"user":  repoUser,
		"token": token,
	})
}

func (LAPI *LibraryApI) DeletedUser(ctx *gin.Context) {
	userID := ctx.Param("userID")

	err := LAPI.users.DeleteUser(userID)
	if err != nil {
		if err == errors.ErrNotFind {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Юзер успешно удален", "userID": userID})
}

func (LAPI *LibraryApI) RepeaterLogin(ctx *gin.Context) {
	userID, exist := ctx.Get("userID")
	if !exist {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "forbidden"})
		return
	}
	userIdStd, ok := userID.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id type"})
		return
	}
	var newUser models.User
	err := ctx.ShouldBindBodyWithJSON(&newUser)
	log.Println(newUser)
	if err != nil {
		log.Println("Error while binding body", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := LAPI.users.RepeatInfoUsers(userIdStd, newUser)
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
	user, err := LAPI.users.InfoUser(userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (LAPI *LibraryApI) allUsers(ctx *gin.Context) {

	users, err := LAPI.users.GetUsers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, users)
}

func (LAPI *LibraryApI) HowIamToken(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	fmt.Printf(userID)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "not found userID"})
		return
	}
	user, err := LAPI.users.InfoUser(userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, user)
}
