package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/internal/models"
	"github.com/intraware/rodan/internal/types"
	"github.com/intraware/rodan/internal/utils"
	"github.com/sirupsen/logrus"
)

func ToTeamResponse(t models.Team) TeamResponse {
	return TeamResponse{
		ID:        t.ID,
		Name:      t.Name,
		Code:      t.Code,
		Ban:       t.Ban,
		Blacklist: t.Blacklist,
		LeaderID:  t.LeaderID,
	}
}

// GetAllTeams godoc
// @Summary      Get all teams
// @Description  Retrieves a list of all teams in the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200  {array}   TeamResponse
// @Failure      500  {object}  types.ErrorResponse
// @Router       /admin/teams [get]
func GetAllTeams(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var teams []TeamResponse
	if err := models.DB.Table("teams").Find(&teams).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "get_all_teams",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in getAllTeams")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":  "get_all_teams",
		"status": "success",
		"count":  len(teams),
		"ip":     ctx.ClientIP(),
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
// @Param        team  body      TeamResponse  true  "Team object"
// @Success      200   {object}  TeamResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/teams/{id} [patch]
func UpdateTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var req TeamResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "update_team",
			"status": "failure",
			"reason": "invalid_request",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid request in updateTeam")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid request"})
		return
	}
	var team models.Team
	if err := models.DB.First(&team, req.ID).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Team not found"})
		return
	}
	team.Name = req.Name
	team.Code = req.Code
	team.Ban = req.Ban
	team.Blacklist = req.Blacklist
	team.LeaderID = req.LeaderID
	if err := models.DB.Save(&team).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "update_team",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in updateTeam")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "update_team",
		"status":  "success",
		"team_id": team.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Team updated successfully")
	ctx.JSON(http.StatusOK, ToTeamResponse(team))
}

// DeleteTeam godoc
// @Summary      Delete a team
// @Description  Deletes an existing team from the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "Team ID"
// @Success      200   {object}  types.SuccessResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/teams/{id} [delete]
func DeleteTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	id := ctx.Param("id")
	var team models.Team
	if err := models.DB.First(&team, id).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Team not found"})
		return
	}
	if err := models.DB.Delete(&team).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "delete_team",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in deleteTeam")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "delete_team",
		"status":  "success",
		"team_id": team.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Team deleted successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Team deleted successfully"})
}

// BanTeam godoc
// @Summary      Ban a team
// @Description  Bans a team from accessing the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "Team ID"
// @Success      200   {object}  types.SuccessResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/teams/{id}/ban [post]
func BanTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	id := ctx.Param("id")
	var team models.Team
	if err := models.DB.First(&team, id).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Team not found"})
		return
	}
	if err := models.DB.Model(&team).Update("ban", true).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "ban_team",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in banTeam")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "ban_team",
		"status":  "success",
		"team_id": team.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Team banned successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Team banned successfully"})
}

// UnbanTeam godoc
// @Summary      Unban a team
// @Description  Removes the ban from a team account
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "Team ID"
// @Success      200   {object}  types.SuccessResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/teams/{id}/unban [post]
func UnbanTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	id := ctx.Param("id")
	var team models.Team
	if err := models.DB.First(&team, id).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Team not found"})
		return
	}
	if err := models.DB.Model(&team).Update("ban", false).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "unban_team",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in unbanTeam")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "unban_team",
		"status":  "success",
		"team_id": team.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Team unbanned successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Team unbanned successfully"})
}

// BlacklistTeam godoc
// @Summary      Blacklist a team
// @Description  Adds a team to the blacklist
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "Team ID"
// @Success      200   {object}  types.SuccessResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/teams/{id}/blacklist [post]
func BlacklistTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	id := ctx.Param("id")
	var team models.Team
	if err := models.DB.First(&team, id).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Team not found"})
		return
	}
	if err := models.DB.Model(&team).Update("blacklist", true).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "blacklist_team",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in blacklistTeam")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "blacklist_team",
		"status":  "success",
		"team_id": team.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Team blacklisted successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Team blacklisted successfully"})
}

// UnblacklistTeam godoc
// @Summary      Remove team from blacklist
// @Description  Removes a team from the blacklist
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "Team ID"
// @Success      200   {object}  types.SuccessResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/teams/{id}/unblacklist [post]
func UnblacklistTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	id := ctx.Param("id")
	var team models.Team
	if err := models.DB.First(&team, id).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Team not found"})
		return
	}
	if err := models.DB.Model(&team).Update("blacklist", false).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "unblacklist_team",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in unblacklistTeam")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "unblacklist_team",
		"status":  "success",
		"team_id": team.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Team unblacklisted successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Team unblacklisted successfully"})
}
