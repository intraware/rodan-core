package challenges

import (
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/challenges/handlers"
	"github.com/intraware/rodan/utils/middleware"
)

func LoadChallenges(r *gin.RouterGroup) {
	challengeRouter := r.Group("/challenge", middleware.BanMiddleware)
	challengeRouter.GET("/list", middleware.CacheMiddleware, handlers.GetChallengeList) // TODO: gotta support chained challenges

	// Protected routes
	protectedRouter := challengeRouter.Group("/", middleware.AuthRequired
	protectedRouter.GET("/:id", middleware.CacheMiddleware, handlers.GetChallengeDetail)
	protectedRouter.GET("/:id/config", handlers.GetChallengeConfig)
	protectedRouter.POST("/:id/submit", handlers.SubmitFlag)

	protectedRouter.POST("/:id/start", handlers.StartDynamicChallenge)
	protectedRouter.POST("/:id/stop", handlers.StopDynamicChallenge)
	protectedRouter.POST("/:id/extend", handlers.ExtendDynamicChallenge)
	protectedRouter.POST("/:id/regenerate", handlers.RegenerateDynamicChallenge)

	// Hint routes (all protected)
	hintRouter := challengeRouter.Group("/:challenge_id/hint", middleware.AuthRequired)
	hintRouter.GET("/list", middleware.CacheMiddleware, handlers.ListHints)
	hintRouter.GET("/:hint_id", handlers.GetHint)
	hintRouter.POST("/:hint_id/buy", handlers.BuyHint)
}
