package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/shared"
	"github.com/intraware/rodan/internal/models"
)

var BanMiddleware gin.HandlerFunc = func(ctx *gin.Context) {
	user_id := ctx.GetInt("user_id")
	var user models.User
	if val, ok := shared.UserCache.Get(user_id); ok {
		user = val
	} else {
		if err := models.DB.First(&user, user_id).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user from DB"})
			return
		} else {
			shared.UserCache.Set(user_id, user)
		}
	}
	if user.Ban {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Account is banned"})
		return
	}
	var team models.Team
	if val, ok := shared.TeamCache.Get(*user.TeamID); ok {
		team = val
	} else {
		if err := models.DB.First(&team, *user.TeamID).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch team from DB"})
			return
		} else {
			shared.TeamCache.Set(team.ID, team)
		}
	}
	if team.Ban {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Team is banned"})
		return
	}
	ctx.Next()
}
