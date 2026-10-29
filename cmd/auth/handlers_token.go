package auth

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RefreshToken(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	fmt.Printf("refresh token test: %s\n", refreshToken)
	if refreshToken == "" || err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Refresh token is required",
		})
		return
	}
	access, refresh, err := RefreshAccessToken(refreshToken)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.SetCookie("auth_token",
		access,
		3600,
		"/",
		"localhost",
		false,
		true)
	ctx.SetCookie("refresh_token",
		refresh,
		86400*7,
		"/",
		"localhost",
		false,
		true)
	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
	})
}
