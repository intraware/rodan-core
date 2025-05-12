package api

import "github.com/gin-gonic/gin"

func LoadTeam(r *gin.RouterGroup) {
	teamRouter := r.Group("/team")
	teamRouter.GET("/me", func(ctx *gin.Context) {})
	teamRouter.GET("/:id", func(ctx *gin.Context) {})
	//teamRouter.GET("")
}
