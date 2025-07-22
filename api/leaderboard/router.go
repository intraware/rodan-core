package leaderboard

import (
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/utils/values"
)

func LoadLeaderboard(r *gin.RouterGroup) {
	lbRouter := r.Group("/leaderboard")
	lbConfig := values.GetConfig().App.EnableLeaderboard
	if lbConfig.User {
		lbRouter.GET("/user", getUserLeaderboard)
	}
	if lbConfig.Team {
		lbRouter.GET("/team", getTeamLeaderboard)
	}
}
