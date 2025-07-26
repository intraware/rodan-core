package challenges

import (
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/challenges/handlers"
	"github.com/intraware/rodan/utils/middleware"
)

func LoadChallenges(r *gin.RouterGroup) {
	challengeRouter := r.Group("/challenge")
	challengeRouter.GET("/list", handlers.GetChallengeList) // TODO: gotta support chained challenges

	// Protected routes
	challengeRouter.GET("/:id", middleware.AuthRequired, handlers.GetChallengeDetail) // gotta change it ..
	challengeRouter.GET("/:id/config", middleware.AuthRequired, handlers.GetChallengeConfig)
	challengeRouter.POST("/:id/submit", middleware.AuthRequired, handlers.SubmitFlag)

	challengeRouter.POST("/:id/start", middleware.AuthRequired, handlers.StartDynamicChallenge)
	challengeRouter.POST("/:id/stop", middleware.AuthRequired, handlers.StopDynamicChallenge)
	challengeRouter.POST("/:id/extend", middleware.AuthRequired, handlers.ExtendDynamicChallenge)
	challengeRouter.POST("/:id/regenerate", middleware.AuthRequired, handlers.RegenerateDynamicChallenge)

	// Hint routes (all protected)
	hintRouter := challengeRouter.Group("/:challenge_id/hint", middleware.AuthRequired)
	hintRouter.GET("/list", handlers.ListHints)
	hintRouter.GET("/:hint_id", handlers.GetHint)
	hintRouter.POST("/:hint_id/buy", handlers.BuyHint)
}
