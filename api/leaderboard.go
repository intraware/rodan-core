package api

import "github.com/gin-gonic/gin"

func LoadLeaderBoard(r *gin.RouterGroup) {
	r.GET("/leaderboard", func(ctx *gin.Context) {})
}
