package team

import (
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/utils/middleware"
)

func LoadTeam(r *gin.RouterGroup) {
	teamRouter := r.Group("/team")

	teamRouter.GET("/:id", middleware.CacheMiddleware, getTeam)

	protectedRouter := teamRouter.Group("/")
	protectedRouter.Use(middleware.AuthRequired)
	protectedRouter.POST("/create", createTeam)
	protectedRouter.POST("/join/:id", joinTeam)
	protectedRouter.GET("/me", middleware.CacheMiddleware, getMyTeam)
	protectedRouter.PATCH("/edit", editTeam)
	protectedRouter.DELETE("/delete", deleteTeam)
	protectedRouter.POST("/leave", leaveTeam)
}
