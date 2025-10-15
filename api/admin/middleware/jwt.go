package middleware

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/intraware/rodan/internal/utils/values"
)

type AdminClaims struct {
	AdminID   uint   `json:"admin_id"`
	Username  string `json:"username"`
	Moderator bool   `json:"moderator"`
	jwt.RegisteredClaims
}

func GenerateJWT(adminID uint, username string, moderator bool, secret string) (string, error) {
	claims := &AdminClaims{
		AdminID:  adminID,
		Username: username,
		Moderator: moderator,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(values.GetConfig().App.TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "rodan",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateJWT(tokenString string, secret string) (*AdminClaims, error) {
	if token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	}); err != nil {
		return nil, err
	} else {
		if claims, ok := token.Claims.(*AdminClaims); ok && token.Valid {
			return claims, nil
		}
	}
	return nil, errors.New("invalid admin token")
}
