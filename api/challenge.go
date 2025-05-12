package api

import "github.com/gin-gonic/gin"

func LoadChallenges(r *gin.RouterGroup) {
	challengeRouter := r.Group("/challenge")
	challengeRouter.GET("/list", func(ctx *gin.Context) {})
	challengeRouter.GET("/:id", func(ctx *gin.Context) {})
	challengeRouter.POST("/:id/submit", func(ctx *gin.Context) {})
}
