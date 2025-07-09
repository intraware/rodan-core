package user

import (
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/utils/middleware"
)

func LoadUser(r *gin.RouterGroup) {
	userRouter := r.Group("/user")
	userRouter.POST("/signup", signUp)
	userRouter.POST("/login", login)
	userRouter.POST("/forgot-password", forgotPassword)
	// Protected routes - middleware applied directly to endpoints
	userRouter.GET("/me", middleware.AuthRequired, getMyProfile)

	// Public routes
	userRouter.GET("/:id", getUserProfile)
}
