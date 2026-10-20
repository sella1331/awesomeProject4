package auth

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Middleware(ctx *gin.Context) {

	authHeader := ctx.GetHeader("Authorization")
	log.Printf("AUTH HEADER: %q (len=%d)", authHeader, len(authHeader))
	log.Printf("ALL HEADERS: %+v", ctx.Request.Header)

	if authHeader == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Authorization header is empty",
		})
		return
	}
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || authHeaderParts[0] != "Bearer" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{})
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Authorization header is invalid",
		})
		return
	}

	claims, err := ParseAccessToken(authHeaderParts[1])
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authorization header is invalid"})
		return
	}
	ctx.Set("userID", claims.UserID)
	ctx.Next()
}
