package leaderboard

import (
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/internal/utils/values"
)

func LoadLeaderboard(r *gin.RouterGroup) {
	lbRouter := r.Group("/leaderboard")
	lbRouter.Use(LastModifiedMiddleware)
	lbConfig := values.GetConfig().App.Leaderboard
	if lbConfig.User {
		lbRouter.GET("/user", getUserLeaderboard)
	}
	if lbConfig.Team {
		lbRouter.GET("/team", getTeamLeaderboard)
	}
}
