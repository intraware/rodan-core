package api

import (
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/challenges"
	"github.com/intraware/rodan/api/leaderboard"
	"github.com/intraware/rodan/api/shared"
	"github.com/intraware/rodan/internal/utils/values"
)

func LoadRoutes(r *gin.Engine) {
	apiRouter := r.Group("/api")

	challenges.LoadChallenges(apiRouter)
	leaderboard.LoadLeaderboard(apiRouter)

	shared.Init(values.GetConfig())
	apiRouter.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"msg": "pong"})
	})
}
