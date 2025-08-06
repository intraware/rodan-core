package auth

import (
	"github.com/gin-gonic/gin"
)

func LoadAuth(r *gin.RouterGroup) {
	authRouter := r.Group("/auth")
	authRouter.POST("/signup", signUp)
	authRouter.POST("/login", login)
	authRouter.POST("/forgot-password", forgotPassword)
	authRouter.POST("/reset-password/:token", resetPassword)
}
