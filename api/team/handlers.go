package team

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/shared"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils"
	"github.com/intraware/rodan/utils/values"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func createTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var req createTeamRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "create_team",
			"status": "failure",
			"reason": "invalid_request_body",
			"ip":     ctx.ClientIP(),
		}).Warn("Failed to parse team creation request")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	userID := ctx.GetInt("user_id")
	var user models.User
	if val, ok := shared.UserCache.Get(userID); ok {
		user = val
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":   "create_team",
				"status":  "failure",
				"reason":  "user_not_found",
				"user_id": userID,
				"ip":      ctx.ClientIP(),
			}).Warn("User not found during team creation")
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "User not found"})
			return
		}
	}
	if user.TeamID != nil {
		auditLog.WithFields(logrus.Fields{
			"event":    "create_team",
			"status":   "failure",
			"reason":   "user_already_in_team",
			"user_id":  user.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
		}).Warn("User already in a team")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "User is already in a team"})
		return
	}
	team := models.Team{
		Name:     req.Name,
		LeaderID: userID,
	}
	if err := models.DB.Create(&team).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":    "create_team",
			"status":   "failure",
			"reason":   "db_team_create_failed",
			"user_id":  user.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
		}).Error("Failed to create team")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to create team"})
		return
	}
	user.TeamID = &team.ID
	if err := models.DB.Save(&user).Error; err != nil {
		_ = models.DB.Delete(&team) // rollback team
		auditLog.WithFields(logrus.Fields{
			"event":    "create_team",
			"status":   "failure",
			"reason":   "user_save_failed",
			"user_id":  user.ID,
			"username": user.Username,
			"team_id":  team.ID,
			"ip":       ctx.ClientIP(),
		}).Error("Failed to associate user with team after creation")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to add user to team"})
		return
	} else {
		shared.UserCache.Delete(user.ID)
	}
	var teamResponse models.Team
	if err := models.DB.First(&teamResponse, team.ID).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":    "create_team",
			"status":   "failure",
			"reason":   "team_reload_failed",
			"user_id":  user.ID,
			"username": user.Username,
			"team_id":  team.ID,
			"ip":       ctx.ClientIP(),
		}).Error("Failed to reload created team")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to load team"})
		return
	}
	shared.TeamCache.Set(team.ID, teamResponse)
	auditLog.WithFields(logrus.Fields{
		"event":     "create_team",
		"status":    "success",
		"user_id":   user.ID,
		"username":  user.Username,
		"team_id":   team.ID,
		"team_name": team.Name,
		"ip":        ctx.ClientIP(),
	}).Info("Team created successfully")
	ctx.JSON(http.StatusCreated, buildTeamResponse(teamResponse))
}

func joinTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	idStr := ctx.Param("id")
	teamID, err := strconv.Atoi(idStr)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "join_team",
			"status": "failure",
			"reason": "invalid_team_id",
			"param":  idStr,
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid team ID format")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid Team ID"})
		return
	}
	var req joinTeamRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "join_team",
			"status": "failure",
			"reason": "invalid_request_body",
			"ip":     ctx.ClientIP(),
		}).Warn("Failed to bind joinTeam request")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	userID := ctx.GetInt("user_id")
	var user models.User
	if val, ok := shared.UserCache.Get(userID); ok {
		user = val
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":   "join_team",
				"status":  "failure",
				"reason":  "user_not_found",
				"user_id": userID,
				"ip":      ctx.ClientIP(),
			}).Warn("User not found")
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "User not found"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	if user.TeamID != nil {
		auditLog.WithFields(logrus.Fields{
			"event":    "join_team",
			"status":   "failure",
			"reason":   "already_in_team",
			"user_id":  user.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
		}).Warn("User is already in a team")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "User is already in a team"})
		return
	}
	var team models.Team
	if val, ok := shared.TeamCache.Get(teamID); ok {
		team = val
	}
	if err := models.DB.Where("id = ?", teamID).First(&team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			auditLog.WithFields(logrus.Fields{
				"event":   "join_team",
				"status":  "failure",
				"reason":  "invalid_team_code_or_id",
				"user_id": user.ID,
				"team_id": teamID,
				"code":    req.Code,
				"ip":      ctx.ClientIP(),
			}).Warn("Invalid team join code or team not found")
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "Invalid team code"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}
	shared.TeamCache.Set(team.ID, team)
	if team.Code != req.Code {
		auditLog.WithFields(logrus.Fields{
			"event":    "join_team",
			"status":   "failure",
			"reason":   "wrong_code",
			"user_id":  user.ID,
			"team_id":  team.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
		}).Warn("Invalid code provided")
		ctx.JSON(http.StatusUnauthorized, errorResponse{Error: "Invalid Code provided"})
		return
	}
	if team.Ban {
		auditLog.WithFields(logrus.Fields{
			"event":    "join_team",
			"status":   "failure",
			"reason":   "team_banned",
			"user_id":  user.ID,
			"team_id":  team.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
		}).Warn("Attempt to join a banned team")
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "Team is banned"})
		return
	}
	teamMaxCount := values.GetConfig().App.TeamSize
	if err := models.DB.Preload("Members").First(&team).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to retrieve team members"})
		return
	}
	if len(team.Members) >= teamMaxCount {
		auditLog.WithFields(logrus.Fields{
			"event":    "join_team",
			"status":   "failure",
			"reason":   "team_full",
			"user_id":  user.ID,
			"team_id":  team.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
		}).Warn("Team max size reached")
		ctx.JSON(http.StatusConflict, errorResponse{Error: "Team max size reached"})
		return
	}
	user.TeamID = &team.ID
	if err := models.DB.Save(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "join_team",
			"status":  "failure",
			"reason":  "user_save_failed",
			"user_id": user.ID,
			"team_id": team.ID,
			"ip":      ctx.ClientIP(),
		}).Error("Failed to join team")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to join team"})
		return
	}
	shared.UserCache.Delete(user.ID)
	var teamResponse models.Team
	if err := models.DB.Preload("Members").First(&teamResponse, team.ID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to load team"})
		return
	}
	shared.TeamCache.Set(team.ID, teamResponse)
	auditLog.WithFields(logrus.Fields{
		"event":     "join_team",
		"status":    "success",
		"user_id":   user.ID,
		"username":  user.Username,
		"team_id":   team.ID,
		"team_name": team.Name,
		"ip":        ctx.ClientIP(),
	}).Info("User joined team successfully")
	ctx.JSON(http.StatusOK, buildTeamResponse(teamResponse))
}

func getMyTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	userID := ctx.GetInt("user_id")
	var user models.User
	cacheHit := false
	if val, ok := shared.UserCache.Get(userID); ok {
		user = val
		cacheHit = true
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":   "get_my_team",
				"status":  "failure",
				"reason":  "user_not_found",
				"user_id": userID,
				"ip":      ctx.ClientIP(),
			}).Warn("User not found in getMyTeam")
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "User not found"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	if user.TeamID == nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "get_my_team",
			"status":  "failure",
			"reason":  "no_team",
			"user_id": user.ID,
			"ip":      ctx.ClientIP(),
		}).Warn("User is not in a team")
		ctx.JSON(http.StatusNotFound, errorResponse{Error: "User is not in a team"})
		return
	}
	var team models.Team
	teamCacheHit := false
	if val, ok := shared.TeamCache.Get(*user.TeamID); ok {
		team = val
		teamCacheHit = true
	} else {
		if err := models.DB.Preload("Members").First(&team, *user.TeamID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":   "get_my_team",
				"status":  "failure",
				"reason":  "team_not_found",
				"user_id": user.ID,
				"team_id": *user.TeamID,
				"ip":      ctx.ClientIP(),
			}).Warn("Team not found")
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "Team not found"})
			return
		}
		shared.TeamCache.Set(*user.TeamID, team)
	}
	auditLog.WithFields(logrus.Fields{
		"event":      "get_my_team",
		"status":     "success",
		"user_id":    user.ID,
		"team_id":    team.ID,
		"ip":         ctx.ClientIP(),
		"user_cache": cacheHit,
		"team_cache": teamCacheHit,
	}).Info("Fetched user's team successfully")
	ctx.JSON(http.StatusOK, buildTeamResponse(team))
}

func getTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	teamIDStr := ctx.Param("id")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "get_team",
			"status": "failure",
			"reason": "invalid_team_id",
			"param":  teamIDStr,
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid team ID")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid team ID"})
		return
	}
	var team models.Team
	cacheHit := false
	if val, ok := shared.TeamCache.Get(teamID); ok {
		team = val
		cacheHit = true
	} else {
		if err := models.DB.Preload("Members").First(&team, teamID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				auditLog.WithFields(logrus.Fields{
					"event":   "get_team",
					"status":  "failure",
					"reason":  "team_not_found",
					"team_id": teamID,
					"ip":      ctx.ClientIP(),
				}).Warn("Team not found")
				ctx.JSON(http.StatusNotFound, errorResponse{Error: "Team not found"})
				return
			}
			auditLog.WithFields(logrus.Fields{
				"event":   "get_team",
				"status":  "failure",
				"reason":  "db_error",
				"team_id": teamID,
				"ip":      ctx.ClientIP(),
			}).Error("Failed to query team from DB")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
			return
		}
		shared.TeamCache.Set(teamID, team)
	}
	userID, exists := ctx.Get("user_id")
	isMember := false
	userIDInt := 0
	if exists {
		userIDInt = userID.(int)
		for _, member := range team.Members {
			if member.ID == userIDInt {
				isMember = true
				break
			}
		}
	}
	response := buildTeamResponse(team)
	if !isMember {
		response.Code = ""
	}
	auditLog.WithFields(logrus.Fields{
		"event":     "get_team",
		"status":    "success",
		"user_id":   userIDInt,
		"team_id":   team.ID,
		"is_member": isMember,
		"ip":        ctx.ClientIP(),
		"cache":     cacheHit,
	}).Info("Fetched team successfully")
	ctx.JSON(http.StatusOK, response)
}

func buildTeamResponse(team models.Team) teamResponse {
	members := make([]userInfo, len(team.Members))
	for i, member := range team.Members {
		members[i] = userInfo{
			ID:             member.ID,
			Username:       member.Username,
			Email:          member.Email,
			GitHubUsername: member.GitHubUsername,
			TeamID:         member.TeamID,
		}
	}
	return teamResponse{
		ID:       team.ID,
		Name:     team.Name,
		Code:     team.Code,
		LeaderID: team.LeaderID,
		Members:  members,
	}
}

func editTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var req editTeamReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "edit_team",
			"status": "failure",
			"reason": "invalid_request_body",
			"ip":     ctx.ClientIP(),
		}).Warn("Failed to parse request body")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Failed to parse the body"})
		return
	}
	userID := ctx.GetInt("user_id")
	var user models.User
	if val, ok := shared.UserCache.Get(userID); ok {
		user = val
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":   "edit_team",
				"status":  "failure",
				"reason":  "user_not_found",
				"user_id": userID,
				"ip":      ctx.ClientIP(),
			}).Warn("User not found")
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "User not found"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	if user.TeamID == nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "edit_team",
			"status":  "failure",
			"reason":  "no_team",
			"user_id": user.ID,
			"ip":      ctx.ClientIP(),
		}).Warn("User not in a team")
		ctx.JSON(http.StatusNotFound, errorResponse{Error: "User is not in a team"})
		return
	}
	var team models.Team
	cacheHit := false
	if val, ok := shared.TeamCache.Get(*user.TeamID); ok {
		team = val
		cacheHit = true
	} else {
		if err := models.DB.Preload("Members").First(&team, *user.TeamID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":   "edit_team",
				"status":  "failure",
				"reason":  "team_not_found",
				"user_id": user.ID,
				"team_id": *user.TeamID,
				"ip":      ctx.ClientIP(),
			}).Warn("Team not found")
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "Team not found"})
			return
		}
		shared.TeamCache.Set(*user.TeamID, team)
	}
	if team.LeaderID != user.ID {
		auditLog.WithFields(logrus.Fields{
			"event":    "edit_team",
			"status":   "failure",
			"reason":   "not_team_leader",
			"user_id":  user.ID,
			"team_id":  team.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
		}).Warn("Non-leader tried to edit team")
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "Only the team leader can edit the team"})
		return
	}
	updates := logrus.Fields{
		"event":   "edit_team",
		"status":  "in_progress",
		"user_id": user.ID,
		"team_id": team.ID,
		"ip":      ctx.ClientIP(),
		"cache":   cacheHit,
	}
	if req.Name != nil {
		updates["new_name"] = *req.Name
		team.Name = *req.Name
	}
	if req.LeaderUsername != nil {
		var newLeader models.User
		if err := models.DB.Where("username = ? AND team_id = ?", *req.LeaderUsername, team.ID).First(&newLeader).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":            "edit_team",
				"status":           "failure",
				"reason":           "new_leader_not_found",
				"requested_leader": *req.LeaderUsername,
				"team_id":          team.ID,
				"user_id":          user.ID,
				"ip":               ctx.ClientIP(),
			}).Warn("Requested new leader not found in team")
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "New leader not found in the team"})
			return
		}
		team.LeaderID = newLeader.ID
		updates["new_leader_id"] = newLeader.ID
		updates["new_leader_username"] = newLeader.Username
	}
	if err := models.DB.Save(&team).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "edit_team",
			"status":  "failure",
			"reason":  "db_save_failed",
			"user_id": user.ID,
			"team_id": team.ID,
			"ip":      ctx.ClientIP(),
		}).Error("Failed to update team")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to update team"})
		return
	}
	shared.TeamCache.Set(team.ID, team)
	updates["status"] = "success"
	auditLog.WithFields(updates).Info("Team updated successfully")
	ctx.JSON(http.StatusOK, successResponse{Message: "Team updated successfully"})
}

func deleteTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	userID := ctx.GetInt("user_id")
	var user models.User
	if val, ok := shared.UserCache.Get(userID); ok {
		user = val
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":   "delete_team",
				"status":  "failure",
				"reason":  "user_not_found",
				"user_id": userID,
				"ip":      ctx.ClientIP(),
			}).Warn("User not found")
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "User not found"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	if user.TeamID == nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "delete_team",
			"status":  "failure",
			"reason":  "no_team",
			"user_id": user.ID,
			"ip":      ctx.ClientIP(),
		}).Warn("User not in a team")
		ctx.JSON(http.StatusNotFound, errorResponse{Error: "User is not in a team"})
		return
	}
	var team models.Team
	cacheHit := false
	if val, ok := shared.TeamCache.Get(*user.TeamID); ok {
		team = val
		cacheHit = true
	} else {
		if err := models.DB.Preload("Members").First(&team, *user.TeamID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":   "delete_team",
				"status":  "failure",
				"reason":  "team_not_found",
				"user_id": user.ID,
				"team_id": *user.TeamID,
				"ip":      ctx.ClientIP(),
			}).Warn("Team not found")
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "Team not found"})
			return
		}
		shared.TeamCache.Set(*user.TeamID, team)
	}
	if team.LeaderID != user.ID {
		auditLog.WithFields(logrus.Fields{
			"event":   "delete_team",
			"status":  "failure",
			"reason":  "not_team_leader",
			"user_id": user.ID,
			"team_id": team.ID,
			"ip":      ctx.ClientIP(),
		}).Warn("Non-leader tried to delete team")
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "Only the team leader can delete the team"})
		return
	}
	if err := models.DB.Delete(&team).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "delete_team",
			"status":  "failure",
			"reason":  "db_delete_failed",
			"user_id": user.ID,
			"team_id": team.ID,
			"ip":      ctx.ClientIP(),
		}).Error("Failed to delete team from DB")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to delete team from database"})
		return
	}
	shared.TeamCache.Delete(team.ID)
	for _, member := range team.Members {
		shared.UserCache.Delete(member.ID)
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "delete_team",
		"status":  "success",
		"user_id": user.ID,
		"team_id": team.ID,
		"ip":      ctx.ClientIP(),
		"cache":   cacheHit,
	}).Info("Team deleted successfully")
	ctx.JSON(http.StatusOK, successResponse{Message: "Deleted team successfully"})
}

func leaveTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	userID := ctx.GetInt("user_id")
	var user models.User
	if val, ok := shared.UserCache.Get(userID); ok {
		user = val
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":   "leave_team",
				"status":  "failure",
				"reason":  "user_not_found",
				"user_id": userID,
				"ip":      ctx.ClientIP(),
			}).Warn("User not found")
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "User not found"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	if user.TeamID == nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "leave_team",
			"status":  "failure",
			"reason":  "no_team",
			"user_id": user.ID,
			"ip":      ctx.ClientIP(),
		}).Warn("User is not in a team")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "User is not in a team"})
		return
	}
	var team models.Team
	cacheHit := false
	if val, ok := shared.TeamCache.Get(*user.TeamID); ok {
		team = val
		cacheHit = true
	} else {
		if err := models.DB.Preload("Members").First(&team, *user.TeamID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":   "leave_team",
				"status":  "failure",
				"reason":  "team_not_found",
				"user_id": user.ID,
				"team_id": *user.TeamID,
				"ip":      ctx.ClientIP(),
			}).Warn("Team not found")
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "Team not found"})
			return
		}
		shared.TeamCache.Set(*user.TeamID, team)
	}
	if team.LeaderID == user.ID {
		auditLog.WithFields(logrus.Fields{
			"event":   "leave_team",
			"status":  "failure",
			"reason":  "leader_cannot_leave",
			"user_id": user.ID,
			"team_id": team.ID,
			"ip":      ctx.ClientIP(),
		}).Warn("Leader tried to leave team")
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "Team leader cannot leave the team. Transfer leadership or delete the team."})
		return
	}
	if err := models.DB.Model(&user).Update("team_id", nil).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "leave_team",
			"status":  "failure",
			"reason":  "db_update_failed",
			"user_id": user.ID,
			"team_id": team.ID,
			"ip":      ctx.ClientIP(),
		}).Error("Failed to leave team")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to leave team"})
		return
	}
	shared.UserCache.Delete(user.ID)
	for i, member := range team.Members {
		if member.ID == user.ID {
			team.Members = append(team.Members[:i], team.Members[i+1:]...)
			break
		}
	}
	shared.TeamCache.Set(team.ID, team)
	auditLog.WithFields(logrus.Fields{
		"event":   "leave_team",
		"status":  "success",
		"user_id": user.ID,
		"team_id": team.ID,
		"ip":      ctx.ClientIP(),
		"cache":   cacheHit,
	}).Info("User left the team")
	ctx.JSON(http.StatusOK, successResponse{Message: "Successfully left the team"})
}
