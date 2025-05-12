package api

import "github.com/gin-gonic/gin"

func LoadNotification(r *gin.RouterGroup) {
	r.GET("/notification", func(ctx *gin.Context) {})
}
