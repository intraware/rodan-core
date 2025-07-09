package api

import (
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/challenges"
	"github.com/intraware/rodan/api/team"
	"github.com/intraware/rodan/api/user"
)

func LoadRoutes(r *gin.Engine) {
	apiRouter := r.Group("/api")
	// LoadAuth(apiRouter)
	LoadLeaderBoard(apiRouter)
	LoadNotification(apiRouter)
	challenges.LoadChallenges(apiRouter)
	team.LoadTeam(apiRouter)
	user.LoadUser(apiRouter)

	apiRouter.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"msg": "pong"})
	})
}
