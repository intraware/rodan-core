package challenges

import (
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/utils/middleware"
)

func LoadChallenges(r *gin.RouterGroup) {
	challengeRouter := r.Group("/challenge")
	challengeRouter.GET("/list", getChallengeList)
	challengeRouter.GET("/:id", getChallengeDetail)

	// Protected routes
	challengeRouter.POST("/:id/submit", middleware.AuthRequired, submitFlag)
	challengeRouter.POST("/:id/start", middleware.AuthRequired, startDynamicChallenge)
	challengeRouter.POST("/:id/stop", middleware.AuthRequired, stopDynamicChallenge)
}
