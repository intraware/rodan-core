package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/config"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils/middleware"
	"gorm.io/gorm"
)

func LoadTeam(r *gin.RouterGroup) {
	teamRouter := r.Group("/team")

	// Public routes
	teamRouter.GET("/:id", getTeam)

	// Protected routes
	protected := teamRouter.Group("")
	protected.Use(middleware.AuthRequired())
	protected.POST("/create", createTeam)
	protected.POST("/join", joinTeam)
	protected.GET("/me", getMyTeam)
}

func createTeam(ctx *gin.Context) {
	var req CreateTeamRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID := ctx.GetInt("user_id")

	// Check if user is already in a team
	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	if user.TeamID != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "User is already in a team"})
		return
	}

	// Create team
	team := models.Team{
		Name:     req.Name,
		LeaderID: userID,
	}

	if err := models.DB.Create(&team).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create team"})
		return
	}

	// Add user to team
	user.TeamID = &team.ID
	if err := models.DB.Save(&user).Error; err != nil {
		// Rollback team creation
		models.DB.Delete(&team)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to add user to team"})
		return
	}

	// Load team with members for response
	var teamResponse models.Team
	if err := models.DB.Preload("Members").First(&teamResponse, team.ID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to load team"})
		return
	}

	ctx.JSON(http.StatusCreated, buildTeamResponse(teamResponse))
}

func joinTeam(ctx *gin.Context) {
	var req JoinTeamRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID := ctx.GetInt("user_id")

	// Check if user is already in a team
	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	if user.TeamID != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "User is already in a team"})
		return
	}

	// Find team by code
	var team models.Team
	if err := models.DB.Where("code = ?", req.Code).First(&team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Invalid team code"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Check if team is banned
	if team.Ban {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Error: "Team is banned"})
		return
	}

	// Refresh team code if needed
	cfg, _ := config.LoadConfig("./config.toml")
	if err := team.RefreshCodeIfNeeded(models.DB, cfg.TeamCodeRefreshMinutes); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to refresh team code"})
		return
	}

	// If code was refreshed, check again
	if err := models.DB.Where("code = ?", req.Code).First(&team).Error; err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Team code has expired"})
		return
	}

	// Add user to team
	user.TeamID = &team.ID
	if err := models.DB.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to join team"})
		return
	}

	// Load team with members for response
	var teamResponse models.Team
	if err := models.DB.Preload("Members").First(&teamResponse, team.ID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to load team"})
		return
	}

	ctx.JSON(http.StatusOK, buildTeamResponse(teamResponse))
}

func getMyTeam(ctx *gin.Context) {
	userID := ctx.GetInt("user_id")

	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	if user.TeamID == nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User is not in a team"})
		return
	}

	var team models.Team
	if err := models.DB.Preload("Members").First(&team, *user.TeamID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Team not found"})
		return
	}

	// Refresh team code if needed
	cfg, _ := config.LoadConfig("./config.toml")
	if err := team.RefreshCodeIfNeeded(models.DB, cfg.TeamCodeRefreshMinutes); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to refresh team code"})
		return
	}

	ctx.JSON(http.StatusOK, buildTeamResponse(team))
}

func getTeam(ctx *gin.Context) {
	teamIDStr := ctx.Param("id")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid team ID"})
		return
	}

	var team models.Team
	if err := models.DB.Preload("Members").First(&team, teamID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Team not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Don't show the team code to non-members
	userID, exists := ctx.Get("user_id")
	isMember := false
	if exists {
		userIDInt := userID.(int)
		for _, member := range team.Members {
			if member.ID == userIDInt {
				isMember = true
				break
			}
		}
	}

	response := buildTeamResponse(team)
	if !isMember {
		response.Code = "" // Hide code from non-members
	}

	ctx.JSON(http.StatusOK, response)
}

func buildTeamResponse(team models.Team) TeamResponse {
	members := make([]UserInfo, len(team.Members))
	for i, member := range team.Members {
		members[i] = UserInfo{
			ID:             member.ID,
			Username:       member.Username,
			Email:          member.Email,
			GitHubUsername: member.GitHubUsername,
			TeamID:         member.TeamID,
		}
	}

	return TeamResponse{
		ID:       team.ID,
		Name:     team.Name,
		Code:     team.Code,
		LeaderID: team.LeaderID,
		Members:  members,
	}
}
