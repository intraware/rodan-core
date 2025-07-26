package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/utils/middleware"
)

func LoadUser(r *gin.RouterGroup) {
	adminRouter := r.Group("/admin")

	// Admin management routes
	adminRouter.GET("/me", middleware.AuthRequired, getAdmin)
	adminRouter.POST("/add", middleware.AuthRequired, addAdmin)
	adminRouter.PATCH("/edit", middleware.AuthRequired, updateAdmin)
	adminRouter.DELETE("/delete", middleware.AuthRequired, deleteAdmin)

	adminRouter.POST("/flush_cache", middleware.AuthRequired, flushCache)
	adminRouter.POST("/close_submission", middleware.AuthRequired, closeChallengeSubmission)

	// Challenge related routes
	challengeRouter := adminRouter.Group("/challenge")
	challengeRouter.GET("/all", middleware.AuthRequired, getAllChallenges)
	challengeRouter.POST("/add", middleware.AuthRequired, addChallenge)
	challengeRouter.PATCH("/edit", middleware.AuthRequired, updateChallenge)
	challengeRouter.DELETE("/delete", middleware.AuthRequired, deleteChallenge)

	// User management routes
	userRouter := adminRouter.Group("/user")
	userRouter.GET("/all", middleware.AuthRequired, getAllUsers)
	userRouter.PATCH("/edit", middleware.AuthRequired, updateUser)
	userRouter.DELETE("/delete", middleware.AuthRequired, deleteUser)
	userRouter.POST("/ban", middleware.AuthRequired, banUser)
	userRouter.POST("/unban", middleware.AuthRequired, unbanUser)
	// userRouter.POST("/reset-password", middleware.AuthRequired, resetUserPassword) --> temp
	userRouter.POST("/remove_from_team", middleware.AuthRequired, removeUserFromTeam)
	userRouter.POST("/add_to_team", middleware.AuthRequired, addUserToTeam)

	// Team management routescacheHit := false
	teamRouter := adminRouter.Group("/team")
	teamRouter.GET("/all", middleware.AuthRequired, getAllTeams)
	teamRouter.PATCH("/edit", middleware.AuthRequired, updateTeam)
	teamRouter.DELETE("/delete", middleware.AuthRequired, deleteTeam)
	teamRouter.POST("/ban", middleware.AuthRequired, banTeam)
	teamRouter.POST("/unban", middleware.AuthRequired, unbanTeam)

	// Container management routes
	containerRouter := adminRouter.Group("/container")
	containerRouter.GET("/all", middleware.AuthRequired, getAllContainers)
	containerRouter.DELETE("/stop", middleware.AuthRequired, stopContainer)
	containerRouter.DELETE("/stop_team", middleware.AuthRequired, stopTeamContainer)
	containerRouter.DELETE("/stop_challenge", middleware.AuthRequired, stopChallengeContainer)
	containerRouter.DELETE("/stop_all", middleware.AuthRequired, stopAllContainers)
}