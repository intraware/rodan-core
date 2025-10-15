package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/internal/utils/values"
)

func AuthAdmin(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization admin header required"})
		ctx.Abort()
		return
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Admin Bearer token required"})
		ctx.Abort()
		return
	}
	adminClaims, err := ValidateJWT(tokenString, values.GetConfig().Server.Security.AdminJWTSecret)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid admin token"})
		ctx.Abort()
		return
	}
	ctx.Set("admin_id", adminClaims.AdminID)
	ctx.Set("username", adminClaims.Username)
	ctx.Set("moderator", adminClaims.Moderator)
	ctx.Next()
}
