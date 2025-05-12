package api

import "github.com/gin-gonic/gin"

func LoadUser(r *gin.RouterGroup) {
	userRouter := r.Group("/user")
	userRouter.GET("/me", func(ctx *gin.Context) {})
	userRouter.GET("/:id", func(ctx *gin.Context) {})
}
