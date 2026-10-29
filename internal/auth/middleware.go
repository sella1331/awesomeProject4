package auth

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Middleware(ctx *gin.Context) {

	authHeader, err := ctx.Cookie("auth_token")

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Authorization cookie is empty or error",
		})
		return
	}
	log.Printf("COOKIE ERROR: %v", err)
	claims, err := ParseAccessToken(authHeader)
	if err != nil {
		log.Printf("PARSE ERROR: %v", err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authorization header is invalid"})
		return
	}
	ctx.Set("userID", claims.UserID)
	fmt.Printf(claims.UserID)
	ctx.Next()
}
