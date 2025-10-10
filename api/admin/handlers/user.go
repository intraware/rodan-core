package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/leaderboard"
	"github.com/intraware/rodan/internal/models"
	"github.com/intraware/rodan/internal/types"
	"github.com/intraware/rodan/internal/utils"
	"github.com/sirupsen/logrus"
)

func ToUserResponse(u models.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
		Active:    u.Active,
		Ban:       u.Ban,
		Blacklist: u.Blacklist,
		TeamID:    u.TeamID,
	}
}

// GetAllUsers godoc
// @Summary      Get all users
// @Description  Retrieves a list of all users in the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200  {array}   UserResponse
// @Failure      500  {object}  types.ErrorResponse
// @Router       /admin/users [get]
func GetAllUsers(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var users []UserResponse
	if err := models.DB.Table("users").Find(&users).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "get_all_users",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in getAllUsers")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	ctx.JSON(http.StatusOK, users)
}

// UpdateUser godoc
// @Summary      Update user information
// @Description  Updates an existing user's information in the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        user  body      UserResponse  true  "User object"
// @Success      200   {object}  UserResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/users/{id} [patch]
func UpdateUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var req UserResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "update_user",
			"status": "failure",
			"reason": "invalid_request",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid request in updateUser")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid request"})
		return
	}
	var user models.User
	if err := models.DB.First(&user, req.ID).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "User not found"})
		return
	}
	user.Username = req.Username
	user.Email = req.Email
	user.AvatarURL = req.AvatarURL
	user.Active = req.Active
	user.Ban = req.Ban
	user.Blacklist = req.Blacklist
	user.TeamID = req.TeamID

	if err := models.DB.Save(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "update_user",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in updateUser")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "update_user",
		"status":  "success",
		"user_id": user.ID,
		"ip":      ctx.ClientIP(),
	}).Info("User updated successfully")
	ctx.JSON(http.StatusOK, ToUserResponse(user))
}

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Deletes an existing user from the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "User ID"
// @Success      200   {object}  types.SuccessResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/users/{id} [delete]
func DeleteUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	id := ctx.Param("id")
	var user models.User
	if err := models.DB.First(&user, id).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "User not found"})
		return
	}
	if err := models.DB.Delete(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "delete_user",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in deleteUser")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "delete_user",
		"status":  "success",
		"user_id": user.ID,
		"ip":      ctx.ClientIP(),
	}).Info("User deleted successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "User deleted successfully"})
}

// BanUser godoc
// @Summary      Ban a user
// @Description  Bans a user from accessing the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "User ID"
// @Success      200   {object}  types.SuccessResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/users/{id}/ban [post]
func BanUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	id := ctx.Param("id")
	var user models.User

	if err := models.DB.First(&user, id).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "User not found"})
		return
	}
	user.Ban = true
	if err := models.DB.Save(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "ban_user",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in banUser")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "ban_user",
		"status":  "success",
		"user_id": user.ID,
		"ip":      ctx.ClientIP(),
	}).Info("User banned successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "User banned successfully"})
}

// UnbanUser godoc
// @Summary      Unban a user
// @Description  Removes the ban from a user account
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "User ID"
// @Success      200   {object}  types.SuccessResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/users/{id}/unban [post]
func UnbanUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	id := ctx.Param("id")
	var user models.User

	if err := models.DB.First(&user, id).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "User not found"})
		return
	}
	user.Ban = false
	if err := models.DB.Save(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "unban_user",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in unbanUser")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "unban_user",
		"status":  "success",
		"user_id": user.ID,
		"ip":      ctx.ClientIP(),
	}).Info("User unbanned successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "User unbanned successfully"})
}

// BlacklistUser godoc
// @Summary      Blacklist a user
// @Description  Adds a user to the blacklist
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "User ID"
// @Success      200   {object}  types.SuccessResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/users/{id}/blacklist [post]
func BlacklistUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	id := ctx.Param("id")
	var user models.User

	if err := models.DB.First(&user, id).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "User not found"})
		return
	}
	user.Blacklist = true
	if err := models.DB.Save(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "blacklist_user",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in blacklistUser")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	leaderboard.MarkLeaderboardDirty()
	auditLog.WithFields(logrus.Fields{
		"event":   "blacklist_user",
		"status":  "success",
		"user_id": user.ID,
		"ip":      ctx.ClientIP(),
	}).Info("User blacklisted successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "User blacklisted successfully"})
}

// UnblacklistUser godoc
// @Summary      Remove user from blacklist
// @Description  Removes a user from the blacklist
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "User ID"
// @Success      200   {object}  types.SuccessResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/users/{id}/unblacklist [post]
func UnblacklistUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	id := ctx.Param("id")
	var user models.User

	if err := models.DB.First(&user, id).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "User not found"})
		return
	}
	user.Blacklist = false
	if err := models.DB.Save(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "unblacklist_user",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in unblacklistUser")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	leaderboard.MarkLeaderboardDirty()
	auditLog.WithFields(logrus.Fields{
		"event":   "unblacklist_user",
		"status":  "success",
		"user_id": user.ID,
		"ip":      ctx.ClientIP(),
	}).Info("User unblacklisted successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "User unblacklisted successfully"})
}

// RemoveUserFromTeam godoc
// @Summary      Remove user from team
// @Description  Removes a user from their current team
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "User ID"
// @Success      200   {object}  types.SuccessResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/users/{id}/remove-from-team [post]
func RemoveUserFromTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	id := ctx.Param("id")
	var user models.User

	if err := models.DB.First(&user, id).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "User not found"})
		return
	}
	user.TeamID = nil
	if err := models.DB.Save(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "remove_user_from_team",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in removeUserFromTeam")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "remove_user_from_team",
		"status":  "success",
		"user_id": user.ID,
		"ip":      ctx.ClientIP(),
	}).Info("User removed from team successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "User removed from team successfully"})
}

// AddUserToTeam godoc
// @Summary      Add user to team
// @Description  Adds a user to a team
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id      path      int   true  "User ID"
// @Param        team_id body      int   true  "Team ID"
// @Success      200     {object}  types.SuccessResponse
// @Failure      400     {object}  types.ErrorResponse
// @Failure      500     {object}  types.ErrorResponse
// @Router       /admin/users/{id}/add-to-team [post]
func AddUserToTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	id := ctx.Param("id")
	var user models.User

	if err := models.DB.First(&user, id).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "User not found"})
		return
	}

	var teamData struct {
		TeamID uint `json:"team_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&teamData); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "add_user_to_team",
			"status": "failure",
			"reason": "no_team_id",
			"ip":     ctx.ClientIP(),
		}).Warn("No team ID provided in addUserToTeam")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "No team ID provided"})
		return
	}

	user.TeamID = &teamData.TeamID
	if err := models.DB.Save(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "add_user_to_team",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in addUserToTeam")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "add_user_to_team",
		"status":  "success",
		"user_id": user.ID,
		"ip":      ctx.ClientIP(),
	}).Info("User added to team successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "User added to team successfully"})
}
