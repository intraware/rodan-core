package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/intraware/rodan/internal/utils/values"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	TeamID   uint   `json:"team_id"`
	jwt.RegisteredClaims
}

func GenerateJWT(teamID, userID uint, username string, secret string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		TeamID:   teamID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(values.GetConfig().App.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "rodan",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateJWT(tokenString string, secret string) (*Claims, error) {
	if token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	}); err != nil {
		return nil, err
	} else {
		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			return claims, nil
		}
	}
	return nil, errors.New("invalid token")
}
