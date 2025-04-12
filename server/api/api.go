package api

import "github.com/gin-gonic/gin"

func LoadRoutes(r *gin.Engine) {
	apiRouter := r.Group("/api")

	apiRouter.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"msg": "pong"})
	})
}
