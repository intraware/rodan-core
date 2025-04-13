package api

import "github.com/gin-gonic/gin"

func LoadAuth(r *gin.RouterGroup) {
	authRouter := r.Group("/auth")
	authRouter.POST("/signup", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"msg": "pong"})
	})
	authRouter.POST("/login", func(ctx *gin.Context) {})
	authRouter.POST("/forgot-password", func(ctx *gin.Context) {})
}
