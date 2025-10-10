package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/admin/handlers"
	"github.com/intraware/rodan/internal/utils/middleware"
)

func LoadUser(r *gin.RouterGroup) {
	adminRouter := r.Group("/admin", middleware.AuthRequired)

	// Admin management
	adminRouter.GET("/me", handlers.GetAdmin)
	adminRouter.POST("/", handlers.AddAdmin)
	adminRouter.PATCH("/:id", handlers.UpdateAdmin)
	adminRouter.DELETE("/:id", handlers.DeleteAdmin)

	// System controls
	adminRouter.POST("/cache/flush", handlers.FlushCache)
	adminRouter.POST("/submissions/close", handlers.CloseChallengeSubmission)
	adminRouter.POST("/submissions/open", handlers.OpenChallengeSubmission)
	adminRouter.POST("/auth/login/close", handlers.CloseLogin)
	adminRouter.POST("/auth/login/open", handlers.OpenLogin)

	// Challenge management
	challengeRouter := adminRouter.Group("/challenges")
	challengeRouter.GET("/", handlers.GetAllChallenges)
	challengeRouter.POST("/", handlers.AddChallenge)
	challengeRouter.PATCH("/:id", handlers.UpdateChallenge)
	challengeRouter.DELETE("/:id", handlers.DeleteChallenge)

	// User management
	userRouter := adminRouter.Group("/users")
	userRouter.GET("/", handlers.GetAllUsers)
	userRouter.PATCH("/:id", handlers.UpdateUser)
	userRouter.DELETE("/:id", handlers.DeleteUser)
	userRouter.POST("/:id/ban", handlers.BanUser)
	userRouter.POST("/:id/unban", handlers.UnbanUser)
	userRouter.POST("/:id/blacklist", handlers.BlacklistUser)
	userRouter.POST("/:id/unblacklist", handlers.UnblacklistUser)
	userRouter.POST("/:id/remove-from-team", handlers.RemoveUserFromTeam)
	userRouter.POST("/:id/add-to-team", handlers.AddUserToTeam)

	// Team management
	teamRouter := adminRouter.Group("/teams")
	teamRouter.GET("/", handlers.GetAllTeams)
	teamRouter.PATCH("/:id", handlers.UpdateTeam)
	teamRouter.DELETE("/:id", handlers.DeleteTeam)
	teamRouter.POST("/:id/ban", handlers.BanTeam)
	teamRouter.POST("/:id/unban", handlers.UnbanTeam)
	teamRouter.POST("/:id/blacklist", handlers.BlacklistTeam)
	teamRouter.POST("/:id/unblacklist", handlers.UnblacklistTeam)

	// Container management
	containerRouter := adminRouter.Group("/containers")
	containerRouter.GET("/", handlers.GetAllSandboxes)
	containerRouter.DELETE("/:id/stop", handlers.StopContainer)
	containerRouter.DELETE("/teams/:id/stop", handlers.StopTeamContainer)
	containerRouter.DELETE("/challenges/:id/stop", handlers.StopChallengeContainer)
	containerRouter.DELETE("/stop-all", handlers.StopAllContainers)
	containerRouter.POST("/kill-all", handlers.KillAllContainers)
}
