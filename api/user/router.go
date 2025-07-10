package user

import (
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/utils/middleware"
)

func LoadUser(r *gin.RouterGroup) {
	userRouter := r.Group("/user")
	// Protected routes - middleware applied directly to endpoints
	userRouter.GET("/me", middleware.AuthRequired, getMyProfile)
	userRouter.PATCH("/edit", middleware.AuthRequired, updateProfile)
	userRouter.DELETE("/delete", middleware.AuthRequired, deleteProfile)
	userRouter.GET("/totp-qr", middleware.AuthRequired, profileTOTP)
	userRouter.GET("/backup-code", middleware.AuthRequired, profileBackupCode)

	// Public routes
	userRouter.GET("/:id", getUserProfile)
}
