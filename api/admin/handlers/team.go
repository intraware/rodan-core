package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils"
	"github.com/sirupsen/logrus"
)

// GetAllTeams godoc
// @Summary      Get all teams
// @Description  Retrieves a list of all teams in the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200  {object}  []models.Team
// @Failure      500  {object}  errorResponse
// @Router       /admin/team/all [get]
func GetAllTeams(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var teams []models.Team

	if err := models.DB.Find(&teams).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "get_all_teams",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in getAllTeams")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "get_all_teams",
		"status":  "success",
		"count":   len(teams),
		"ip":      ctx.ClientIP(),
	}).Info("Retrieved all teams successfully")
	ctx.JSON(http.StatusOK, teams)
}

// UpdateTeam godoc
// @Summary      Update team information
// @Description  Updates an existing team's information in the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        team  body      models.Team  true  "Team object"
// @Success      200   {object}  models.Team
// @Failure      400   {object}  errorResponse
// @Failure      500   {object}  errorResponse
// @Router       /admin/team/edit [patch]
func UpdateTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var team models.Team

	if err := ctx.ShouldBindJSON(&team); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "update_team",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in updateTeam")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	if err := models.DB.Save(&team).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "update_team",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in updateTeam")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "update_team",
		"status":  "success",
		"team_id": team.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Team updated successfully")
	ctx.JSON(http.StatusOK, team)
}

// DeleteTeam godoc
// @Summary      Delete a team
// @Description  Deletes an existing team from the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        team  body      models.Team  true  "Team object"
// @Success      200   {object}  successResponse
// @Failure      400   {object}  errorResponse
// @Failure      500   {object}  errorResponse
// @Router       /admin/team/delete [delete]
func DeleteTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var team models.Team

	if err := ctx.ShouldBindJSON(&team); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "delete_team",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in deleteTeam")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	if err := models.DB.Delete(&team).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "delete_team",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in deleteTeam")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "delete_team",
		"status":  "success",
		"team_id": team.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Team deleted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Team deleted successfully"})
}

// BanTeam godoc
// @Summary      Ban a team
// @Description  Bans a team from accessing the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        team  body      models.Team  true  "Team object"
// @Success      200   {object}  successResponse
// @Failure      400   {object}  errorResponse
// @Failure      500   {object}  errorResponse
// @Router       /admin/team/ban [post]
func BanTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var team models.Team

	if err := ctx.ShouldBindJSON(&team); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "ban_team",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in banTeam")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	if err := models.DB.Model(&team).Where("id = ?", team.ID).Update("ban", true).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "ban_team",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in banTeam")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "ban_team",
		"status":  "success",
		"team_id": team.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Team banned successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Team banned successfully"})
}

// UnbanTeam godoc
// @Summary      Unban a team
// @Description  Removes the ban from a team account
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        team  body      models.Team  true  "Team object"
// @Success      200   {object}  successResponse
// @Failure      400   {object}  errorResponse
// @Failure      500   {object}  errorResponse
// @Router       /admin/team/unban [post]
func UnbanTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var team models.Team

	if err := ctx.ShouldBindJSON(&team); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "unban_team",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in unbanTeam")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	if err := models.DB.Model(&team).Where("id = ?", team.ID).Update("ban", false).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "unban_team",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in unbanTeam")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "unban_team",
		"status":  "success",
		"team_id": team.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Team unbanned successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Team unbanned successfully"})
}

// BlacklistTeam godoc
// @Summary      Blacklist a team
// @Description  Adds a team to the blacklist
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        team  body      models.Team  true  "Team object"
// @Success      200   {object}  successResponse
// @Failure      400   {object}  errorResponse
// @Failure      500   {object}  errorResponse
// @Router       /admin/team/blacklist [post]
func BlacklistTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var team models.Team

	if err := ctx.ShouldBindJSON(&team); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "blacklist_team",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in blacklistTeam")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	if err := models.DB.Model(&team).Where("id = ?", team.ID).Update("blacklist", true).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "blacklist_team",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in blacklistTeam")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "blacklist_team",
		"status":  "success",
		"team_id": team.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Team blacklisted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Team blacklisted successfully"})
}

// UnblacklistTeam godoc
// @Summary      Remove team from blacklist
// @Description  Removes a team from the blacklist
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        team  body      models.Team  true  "Team object"
// @Success      200   {object}  successResponse
// @Failure      400   {object}  errorResponse
// @Failure      500   {object}  errorResponse
// @Router       /admin/team/unblacklist [post]
func UnblacklistTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var team models.Team

	if err := ctx.ShouldBindJSON(&team); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "unblacklist_team",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in unblacklistTeam")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	if err := models.DB.Model(&team).Where("id = ?", team.ID).Update("blacklist", false).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "unblacklist_team",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in unblacklistTeam")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "unblacklist_team",
		"status":  "success",
		"team_id": team.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Team unblacklisted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Team unblacklisted successfully"})
}
