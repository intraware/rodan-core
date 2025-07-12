package team

import (
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/utils/middleware"
)

func LoadTeam(r *gin.RouterGroup) {
	teamRouter := r.Group("/team")

	// Public routes
	teamRouter.GET("/:id", getTeam)

	// Protected routes - middleware applied directly to endpoints
	teamRouter.POST("/create", middleware.AuthRequired, createTeam)
	teamRouter.POST("/join/:id", middleware.AuthRequired, joinTeam)
	teamRouter.GET("/me", middleware.AuthRequired, getMyTeam)
	teamRouter.PATCH("/edit", middleware.AuthRequired, editTeam)
	teamRouter.DELETE("/delete", middleware.AuthRequired, deleteTeam)
	teamRouter.POST("/leave", middleware.AuthRequired, leaveTeam)
}
