package leaderboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/utils/values"
)

func getUserLeaderboard(ctx *gin.Context) {
	if !values.GetConfig().App.Leaderboard.User {
		ctx.JSON(http.StatusTeapot, errorResponse{Error: "Enable User Leaderboard in the config"})
		return
	}
	ctx.JSON(http.StatusOK, GetCachedUserLeaderboard())
}

func getTeamLeaderboard(ctx *gin.Context) {
	if !values.GetConfig().App.Leaderboard.Team {
		ctx.JSON(http.StatusTeapot, errorResponse{Error: "Enable Team Leaderboard in the config"})
		return
	}
	ctx.JSON(http.StatusOK, GetCachedTeamLeaderboard())
}
