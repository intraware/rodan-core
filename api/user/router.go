package user

import (
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/internal/utils/middleware"
)

func LoadUser(r *gin.RouterGroup) {
	userRouter := r.Group("/user")

	protectedRouter := userRouter.Group("/")
	protectedRouter.Use(middleware.AuthRequired)
	protectedRouter.GET("/me", middleware.CacheMiddleware, getMyProfile)
	protectedRouter.PATCH("/edit", updateProfile)
	protectedRouter.DELETE("/delete", deleteProfile)
	protectedRouter.GET("/totp-qr", middleware.CacheMiddleware, profileTOTP)
	protectedRouter.GET("/backup-code", profileBackupCode)

	userRouter.GET("/:id", middleware.CacheMiddleware, getUserProfile)
}
