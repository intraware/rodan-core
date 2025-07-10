package user

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AnimeKaizoku/cacher"
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/models"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

var UserCache = cacher.NewCacher[int, models.User](&cacher.NewCacherOpts{
	TimeToLive:    time.Minute * 3,
	CleanInterval: time.Hour * 2,
	CleanerMode:   cacher.CleaningCentral,
})

func getMyProfile(ctx *gin.Context) {
	userID := ctx.GetInt("user_id")
	var user models.User
	if val, ok := UserCache.Get(userID); ok {
		user = val
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "User not found"})
			return
		} else {
			UserCache.Set(user.ID, user)
		}
	}
	userInfo := userInfo{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		GitHubUsername: user.GitHubUsername,
		TeamID:         user.TeamID,
	}
	ctx.JSON(http.StatusOK, userInfo)
}

func getUserProfile(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid user ID"})
		return
	}
	var user models.User
	if val, ok := UserCache.Get(userID); ok {
		user = val
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				ctx.JSON(http.StatusNotFound, errorResponse{Error: "User not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
			return
		} else {
			UserCache.Set(userID, user)
		}
	}
	userInfo := userInfo{
		ID:             user.ID,
		Username:       user.Username,
		GitHubUsername: user.GitHubUsername,
		TeamID:         user.TeamID,
	}
	ctx.JSON(http.StatusOK, userInfo)
}

func updateProfile(ctx *gin.Context) {
	userID := ctx.GetInt("user_id")
	var input updateUserRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid input"})
		return
	}
	var user models.User
	if val, ok := UserCache.Get(userID); ok {
		user = val
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "User not found"})
			return
		}
	}
	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.GitHubUsername != nil {
		user.GitHubUsername = *input.GitHubUsername
	}
	if err := models.DB.Save(&user).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "duplicate") {
			ctx.JSON(http.StatusConflict, errorResponse{Error: "Username or GitHub username already in use"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to update profile"})
		return
	}
	UserCache.Delete(userID)
	ctx.JSON(http.StatusOK, successResponse{Message: "Profile updated successfully"})
}

func deleteProfile(ctx *gin.Context) {
	userID := ctx.GetInt("user_id")
	result := models.DB.Delete(&models.User{}, userID)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Internal server error"})
		return
	}
	UserCache.Delete(userID)
	if result.RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, errorResponse{Error: "User not found or already deleted"})
		return
	}
	ctx.JSON(http.StatusOK, successResponse{Message: "User deleted successfully"})
}

func profileTOTP(ctx *gin.Context) {
	userID := ctx.GetInt("user_id")
	var user models.User
	if val, ok := UserCache.Get(userID); ok {
		user = val
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to fetch user"})
			return
		}
		UserCache.Set(userID, user)
	}
	totpURL, _ := user.TOTPUrl()
	png, err := qrcode.Encode(totpURL, qrcode.Medium, 256)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to generate QR code"})
		return
	}
	ctx.Header("Content-Type", "image/png")
	ctx.Writer.Write(png)
}

func profileBackupCode(ctx *gin.Context) {
	userID := ctx.GetInt("user_id")
	var user models.User
	if val, ok := UserCache.Get(userID); ok {
		user = val
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to fetch user"})
			return
		}
		UserCache.Set(userID, user)
	}
	ctx.JSON(http.StatusOK, gin.H{"backup_code": user.BackupCode})
}
