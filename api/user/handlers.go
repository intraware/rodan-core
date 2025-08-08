package user

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/shared"
	"github.com/intraware/rodan/internal/models"
	"github.com/intraware/rodan/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

func getMyProfile(ctx *gin.Context) {
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
				"event":   "get_my_profile",
				"status":  "failure",
				"reason":  "user_not_found",
				"user_id": userID,
				"ip":      ctx.ClientIP(),
			}).Warn("User not found in getMyProfile")
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "User not found"})
			return
		} else {
			shared.UserCache.Set(user.ID, user)
		}
	}
	userInfo := userInfo{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		GitHubUsername: user.GitHubUsername,
		TeamID:         user.TeamID,
	}
	auditLog.WithFields(logrus.Fields{
		"event":    "get_my_profile",
		"status":   "success",
		"user_id":  user.ID,
		"username": user.Username,
		"ip":       ctx.ClientIP(),
		"cache":    cacheHit,
	}).Info("Fetched own profile")
	ctx.JSON(http.StatusOK, userInfo)
}

func getUserProfile(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	userIDStr := ctx.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "get_user_profile",
			"status": "failure",
			"reason": "invalid_user_id",
			"input":  userIDStr,
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid user ID in getUserProfile")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid user ID"})
		return
	}
	var user models.User
	cacheHit := false
	if val, ok := shared.UserCache.Get(userID); ok {
		user = val
		cacheHit = true
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				auditLog.WithFields(logrus.Fields{
					"event":   "get_user_profile",
					"status":  "failure",
					"reason":  "user_not_found",
					"user_id": userID,
					"ip":      ctx.ClientIP(),
				}).Warn("User not found in getUserProfile")
				ctx.JSON(http.StatusNotFound, errorResponse{Error: "User not found"})
				return
			}
			auditLog.WithFields(logrus.Fields{
				"event":   "get_user_profile",
				"status":  "failure",
				"reason":  "db_error",
				"user_id": userID,
				"ip":      ctx.ClientIP(),
				"error":   err.Error(),
			}).Error("Database error in getUserProfile")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
			return
		} else {
			shared.UserCache.Set(userID, user)
		}
	}
	userInfo := userInfo{
		ID:             user.ID,
		Username:       user.Username,
		GitHubUsername: user.GitHubUsername,
		TeamID:         user.TeamID,
	}
	auditLog.WithFields(logrus.Fields{
		"event":    "get_user_profile",
		"status":   "success",
		"user_id":  user.ID,
		"username": user.Username,
		"ip":       ctx.ClientIP(),
		"cache":    cacheHit,
	}).Info("Fetched other user's profile")
	ctx.JSON(http.StatusOK, userInfo)
}

func updateProfile(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	userID := ctx.GetInt("user_id")
	var input updateUserRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "update_profile",
			"status":  "failure",
			"reason":  "invalid_json",
			"user_id": userID,
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid input in updateProfile")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid input"})
		return
	}
	var user models.User
	cacheHit := false
	if val, ok := shared.UserCache.Get(userID); ok {
		user = val
		cacheHit = true
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":   "update_profile",
				"status":  "failure",
				"reason":  "user_not_found",
				"user_id": userID,
				"ip":      ctx.ClientIP(),
			}).Warn("User not found in updateProfile")
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "User not found"})
			return
		}
	}
	oldUsername := user.Username
	oldGitHub := user.GitHubUsername
	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.GitHubUsername != nil {
		user.GitHubUsername = *input.GitHubUsername
	}
	if err := models.DB.Save(&user).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "duplicate") {
			auditLog.WithFields(logrus.Fields{
				"event":        "update_profile",
				"status":       "failure",
				"reason":       "duplicate_username_or_github",
				"user_id":      user.ID,
				"old_username": oldUsername,
				"new_username": user.Username,
				"old_github":   oldGitHub,
				"new_github":   user.GitHubUsername,
				"ip":           ctx.ClientIP(),
				"cache":        cacheHit,
				"error":        err.Error(),
			}).Warn("Username or GitHub username already in use in updateProfile")
			ctx.JSON(http.StatusConflict, errorResponse{Error: "Username or GitHub username already in use"})
			return
		}
		auditLog.WithFields(logrus.Fields{
			"event":   "update_profile",
			"status":  "failure",
			"reason":  "db_error",
			"user_id": user.ID,
			"ip":      ctx.ClientIP(),
			"error":   err.Error(),
		}).Error("Failed to update profile in DB")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to update profile"})
		return
	}
	shared.UserCache.Delete(userID)
	auditLog.WithFields(logrus.Fields{
		"event":        "update_profile",
		"status":       "success",
		"user_id":      user.ID,
		"old_username": oldUsername,
		"new_username": user.Username,
		"old_github":   oldGitHub,
		"new_github":   user.GitHubUsername,
		"ip":           ctx.ClientIP(),
		"cache":        cacheHit,
	}).Info("Profile updated successfully")
	ctx.JSON(http.StatusOK, successResponse{Message: "Profile updated successfully"})
}

func deleteProfile(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	userID := ctx.GetInt("user_id")
	result := models.DB.Delete(&models.User{}, userID)
	if result.Error != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "delete_profile",
			"status":  "failure",
			"reason":  "db_error",
			"user_id": userID,
			"ip":      ctx.ClientIP(),
			"error":   result.Error.Error(),
		}).Error("Failed to delete user in deleteProfile")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Internal server error"})
		return
	}
	shared.UserCache.Delete(userID)
	if result.RowsAffected == 0 {
		auditLog.WithFields(logrus.Fields{
			"event":   "delete_profile",
			"status":  "failure",
			"reason":  "not_found_or_already_deleted",
			"user_id": userID,
			"ip":      ctx.ClientIP(),
		}).Warn("User not found or already deleted in deleteProfile")
		ctx.JSON(http.StatusNotFound, errorResponse{Error: "User not found or already deleted"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "delete_profile",
		"status":  "success",
		"user_id": userID,
		"ip":      ctx.ClientIP(),
	}).Info("User deleted successfully")
	ctx.JSON(http.StatusOK, successResponse{Message: "User deleted successfully"})
}

func profileTOTP(ctx *gin.Context) {
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
				"event":   "profile_totp",
				"status":  "failure",
				"reason":  "db_error",
				"user_id": userID,
				"ip":      ctx.ClientIP(),
				"error":   err.Error(),
			}).Error("Failed to fetch user in profileTOTP")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to fetch user"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	totpURL, _ := user.TOTPUrl()
	png, err := qrcode.Encode(totpURL, qrcode.Medium, 256)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "profile_totp",
			"status":  "failure",
			"reason":  "qrcode_generation_failed",
			"user_id": userID,
			"ip":      ctx.ClientIP(),
			"error":   err.Error(),
		}).Error("Failed to generate QR code in profileTOTP")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to generate QR code"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "profile_totp",
		"status":  "success",
		"user_id": userID,
		"ip":      ctx.ClientIP(),
		"cache":   cacheHit,
	}).Info("TOTP QR code generated for profile")
	ctx.Header("Content-Type", "image/png")
	ctx.Writer.Write(png)
}

func profileBackupCode(ctx *gin.Context) {
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
				"event":   "profile_backup_code",
				"status":  "failure",
				"reason":  "db_error",
				"user_id": userID,
				"ip":      ctx.ClientIP(),
				"error":   err.Error(),
			}).Error("Failed to fetch user in profileBackupCode")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to fetch user"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	auditLog.WithFields(logrus.Fields{
		"event":       "profile_backup_code",
		"status":      "success",
		"user_id":     userID,
		"ip":          ctx.ClientIP(),
		"cache":       cacheHit,
		"backup_code": user.BackupCode,
	}).Info("Fetched backup code for profile")
	ctx.JSON(http.StatusOK, gin.H{"backup_code": user.BackupCode})
}
