// TODO: admin is vibe coded and I can't fix it now
package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/admin/handlers"
	"github.com/intraware/rodan/internal/utils/middleware"
)

func LoadUser(r *gin.RouterGroup) {
	adminRouter := r.Group("/admin", middleware.AuthRequired)

	// Admin management routes
	adminRouter.GET("/me", handlers.GetAdmin)
	adminRouter.POST("/add", handlers.AddAdmin)
	adminRouter.PATCH("/edit", handlers.UpdateAdmin)
	adminRouter.DELETE("/delete", handlers.DeleteAdmin)

	adminRouter.POST("/flush_cache", handlers.FlushCache)
	adminRouter.POST("/close_submission", handlers.CloseChallengeSubmission)
	adminRouter.POST("/open_submission", handlers.OpenChallengeSubmission)
	adminRouter.POST("/close_login", handlers.CloseLogin)
	adminRouter.POST("/open_login", handlers.OpenLogin)

	// Challenge related routes
	challengeRouter := adminRouter.Group("/challenge")
	challengeRouter.GET("/all", handlers.GetAllChallenges)
	challengeRouter.POST("/add", handlers.AddChallenge)
	challengeRouter.PATCH("/edit", handlers.UpdateChallenge)
	challengeRouter.DELETE("/delete", handlers.DeleteChallenge)

	// User management routes
	userRouter := adminRouter.Group("/user")
	userRouter.GET("/all", handlers.GetAllUsers)
	userRouter.PATCH("/edit", handlers.UpdateUser)
	userRouter.DELETE("/delete", handlers.DeleteUser)
	userRouter.POST("/ban", handlers.BanUser)
	userRouter.POST("/unban", handlers.UnbanUser)
	userRouter.POST("/blacklist", handlers.BlacklistUser)
	userRouter.POST("/unblacklist", handlers.UnblacklistUser)
	// userRouter.POST("/reset-password", resetUserPassword) --> temp
	userRouter.POST("/remove_from_team", handlers.RemoveUserFromTeam)
	userRouter.POST("/add_to_team", handlers.AddUserToTeam)

	// Team management routes
	teamRouter := adminRouter.Group("/team")
	teamRouter.GET("/all", handlers.GetAllTeams)
	teamRouter.PATCH("/edit", handlers.UpdateTeam)
	teamRouter.DELETE("/delete", handlers.DeleteTeam)
	teamRouter.POST("/ban", handlers.BanTeam)
	teamRouter.POST("/unban", handlers.UnbanTeam)
	teamRouter.POST("/blacklist", handlers.BlacklistTeam)
	teamRouter.POST("/unblacklist", handlers.UnblacklistTeam)

	// Container management routes
	containerRouter := adminRouter.Group("/container")
	containerRouter.GET("/all", handlers.GetAllContainers)
	containerRouter.DELETE("/stop", handlers.StopContainer)
	containerRouter.DELETE("/stop_team", handlers.StopTeamContainer)
	containerRouter.DELETE("/stop_challenge", handlers.StopChallengeContainer)
	containerRouter.DELETE("/stop_all", handlers.StopAllContainers)
	containerRouter.POST("/kill_all", handlers.KillAllContainers)
}
