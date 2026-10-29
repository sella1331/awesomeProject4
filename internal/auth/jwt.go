package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const superSecretKey = "hdvbiehfewkjvffwegerhjgnrtgefsvdfbyhegyrgerg45ujyng"

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(15 * time.Minute),
			),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)
	return token.SignedString([]byte(superSecretKey))
}

func GenerateRefreshToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(30 * 24 * time.Hour),
			),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)
	return token.SignedString([]byte(superSecretKey))

}
func RefreshAccessToken(RefrasherToken string) (string, string, error) {
	claims, err := ParseAccessToken(RefrasherToken)
	if err != nil {
		return "", "", err
	}
	newAccessToken, err := GenerateAccessToken(claims.UserID)
	if err != nil {
		return "", "", err
	}
	newRefreshToken, err := GenerateRefreshToken(claims.UserID)
	if err != nil {
		return "", "", err
	}
	return newAccessToken, newRefreshToken, nil
}

func ParseAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(superSecretKey), nil
		})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
