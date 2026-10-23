package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RefreshToken(ctx *gin.Context) {
	refreshToken := ctx.GetHeader("Refresh_token")
	if refreshToken == "" {
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
	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
	})
}
