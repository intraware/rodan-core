package leaderboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/internal/types"
	"github.com/intraware/rodan/internal/utils/values"
)

// getUserLeaderboard godoc
// @Summary      Get user leaderboard
// @Description  Retrieves the current user leaderboard rankings
// @Security     BearerAuth
// @Tags         leaderboard
// @Accept       json
// @Produce      json
// @Success      200  {object}  []userLeaderboardEntry
// @Failure      418  {object}  types.ErrorResponse
// @Router       /leaderboard/user [get]
func getUserLeaderboard(ctx *gin.Context) {
	if !values.GetConfig().App.Leaderboard.User {
		ctx.JSON(http.StatusTeapot, types.ErrorResponse{Error: "Enable User Leaderboard in the config"})
		return
	}
	ctx.JSON(http.StatusOK, GetCachedUserLeaderboard())
}

// getTeamLeaderboard godoc
// @Summary      Get team leaderboard
// @Description  Retrieves the current team leaderboard rankings
// @Security     BearerAuth
// @Tags         leaderboard
// @Accept       json
// @Produce      json
// @Success      200  {object}  []teamLeaderboardEntry
// @Failure      418  {object}  types.ErrorResponse
// @Router       /leaderboard/team [get]
func getTeamLeaderboard(ctx *gin.Context) {
	if !values.GetConfig().App.Leaderboard.Team {
		ctx.JSON(http.StatusTeapot, types.ErrorResponse{Error: "Enable Team Leaderboard in the config"})
		return
	}
	ctx.JSON(http.StatusOK, GetCachedTeamLeaderboard())
}
