package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils"
	"github.com/sirupsen/logrus"
)

func getAllUsers(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var users []models.User

	if err := models.DB.Find(&users).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "get_all_users",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in getAllUsers")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func updateUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "update_user",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in updateUser")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	if err := models.DB.Save(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "update_user",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in updateUser")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
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

func deleteUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "delete_user",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in deleteUser")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	if err := models.DB.Delete(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "delete_user",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in deleteUser")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
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

func banUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "ban_user",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in banUser")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	user.Ban = true
	if err := models.DB.Save(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "ban_user",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in banUser")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
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

func unbanUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "unban_user",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in unbanUser")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	user.Ban = false
	if err := models.DB.Save(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "unban_user",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in unbanUser")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
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

func blacklistUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "blacklist_user",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in blacklistUser")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	user.Blacklist = true
	if err := models.DB.Save(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "blacklist_user",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in blacklistUser")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
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

func unblacklistUser(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "unblacklist_user",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in unblacklistUser")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	user.Blacklist = false
	if err := models.DB.Save(&user).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "unblacklist_user",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in unblacklistUser")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
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