package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/internal/models"
	"github.com/intraware/rodan/internal/types"
	"github.com/intraware/rodan/internal/utils"
	"github.com/sirupsen/logrus"
)

// GetAllUsers godoc
// @Summary      Get all users
// @Description  Retrieves a list of all users in the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200  {object}  []models.User
// @Failure      500  {object}  types.ErrorResponse
// @Router       /admin/user/all [get]
func GetAllUsers(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var users []models.User

	if err := models.DB.Find(&users).Error; err != nil {
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
// @Param        user  body      models.User  true  "User object"
// @Success      200   {object}  models.User
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/user/edit [patch]
func UpdateUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "update_user",
			"status": "failure",
			"reason": "invalid_request",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid request in updateUser")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid request"})
		return
	}

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
	ctx.JSON(http.StatusOK, user)
}

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Deletes an existing user from the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "User object"
// @Success      200   {object}  successResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/user/delete [delete]
func DeleteUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "delete_user",
			"status": "failure",
			"reason": "invalid_request",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid request in deleteUser")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid request"})
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
	ctx.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// BanUser godoc
// @Summary      Ban a user
// @Description  Bans a user from accessing the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "User object"
// @Success      200   {object}  successResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/user/ban [post]
func BanUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "ban_user",
			"status": "failure",
			"reason": "invalid_request",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid request in banUser")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid request"})
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
	ctx.JSON(http.StatusOK, gin.H{"message": "User banned successfully"})
}

// UnbanUser godoc
// @Summary      Unban a user
// @Description  Removes the ban from a user account
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "User object"
// @Success      200   {object}  successResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/user/unban [post]
func UnbanUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "unban_user",
			"status": "failure",
			"reason": "invalid_request",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid request in unbanUser")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid request"})
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
	ctx.JSON(http.StatusOK, gin.H{"message": "User unbanned successfully"})
}

// BlacklistUser godoc
// @Summary      Blacklist a user
// @Description  Adds a user to the blacklist
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "User object"
// @Success      200   {object}  successResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/user/blacklist [post]
func BlacklistUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "blacklist_user",
			"status": "failure",
			"reason": "invalid_request",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid request in blacklistUser")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid request"})
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

	auditLog.WithFields(logrus.Fields{
		"event":   "blacklist_user",
		"status":  "success",
		"user_id": user.ID,
		"ip":      ctx.ClientIP(),
	}).Info("User blacklisted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "User blacklisted successfully"})
}

// UnblacklistUser godoc
// @Summary      Remove user from blacklist
// @Description  Removes a user from the blacklist
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "User object"
// @Success      200   {object}  successResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /admin/user/unblacklist [post]
func UnblacklistUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "unblacklist_user",
			"status": "failure",
			"reason": "invalid_request",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid request in unblacklistUser")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid request"})
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

	auditLog.WithFields(logrus.Fields{
		"event":   "unblacklist_user",
		"status":  "success",
		"user_id": user.ID,
		"ip":      ctx.ClientIP(),
	}).Info("User unblacklisted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "User unblacklisted successfully"})
}

func RemoveUserFromTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "remove_user_from_team",
			"status": "failure",
			"reason": "invalid_request",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid request in removeUserFromTeam")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
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
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "remove_user_from_team",
		"status":  "success",
		"user_id": user.ID,
		"ip":      ctx.ClientIP(),
	}).Info("User removed from team successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "User removed from team successfully"})
}

func AddUserToTeam(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "add_user_to_team",
			"status": "failure",
			"reason": "invalid_request",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid request in addUserToTeam")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	if user.TeamID == nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "add_user_to_team",
			"status": "failure",
			"reason": "no_team_id",
			"ip":     ctx.ClientIP(),
		}).Warn("No team ID provided in addUserToTeam")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "No team ID provided"})
		return
	}

	if err := models.DB.Save(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "add_user_to_team",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in addUserToTeam")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "add_user_to_team",
		"status":  "success",
		"user_id": user.ID,
		"ip":      ctx.ClientIP(),
	}).Info("User added to team successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "User added to team successfully"})
}