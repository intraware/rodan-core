package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AnimeKaizoku/cacher"
	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils"
	"github.com/intraware/rodan/utils/values"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var LoginCache = cacher.NewCacher[string, models.User](&cacher.NewCacherOpts{
	TimeToLive:    time.Minute * 3,
	CleanInterval: time.Hour * 1,
	CleanerMode:   cacher.CleaningCentral,
})

func signUp(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var req signUpRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "sign_up",
			"status": "failure",
			"reason": "invalid_json",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid signup input")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Failed to parse the body"})
		return
	}
	user := models.User{
		Username:       req.Username,
		Email:          req.Email,
		Password:       req.Password,
		GitHubUsername: req.GitHubUsername,
	}
	if err := models.DB.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
			auditLog.WithFields(logrus.Fields{
				"event":    "sign_up",
				"status":   "failure",
				"reason":   "duplicate_user",
				"username": req.Username,
				"email":    req.Email,
				"github":   req.GitHubUsername,
				"ip":       ctx.ClientIP(),
			}).Warn("User already exists during signup")
			ctx.JSON(http.StatusConflict, errorResponse{Error: "User with same email or username or github username exists"})
			return
		}
		auditLog.WithFields(logrus.Fields{
			"event":    "sign_up",
			"status":   "failure",
			"reason":   "db_error",
			"username": req.Username,
			"email":    req.Email,
			"github":   req.GitHubUsername,
			"ip":       ctx.ClientIP(),
			"error":    err.Error(),
		}).Error("Failed to create user in DB during signup")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to create user"})
		return
	}
	token, err := utils.GenerateJWT(user.ID, user.Username, values.GetConfig().Security.JWTSecret)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":    "sign_up",
			"status":   "failure",
			"reason":   "token_generation_failed",
			"user_id":  user.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
			"error":    err.Error(),
		}).Error("Failed to generate JWT token during signup")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to generate token"})
		return
	}
	userInfo := userInfo{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		GitHubUsername: user.GitHubUsername,
		TeamID:         user.TeamID,
	}
	auditLog.WithFields(logrus.Fields{
		"event":    "sign_up",
		"status":   "success",
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"github":   user.GitHubUsername,
		"ip":       ctx.ClientIP(),
	}).Info("User signed up successfully")
	ctx.JSON(http.StatusCreated, authResponse{
		Token: token,
		User:  userInfo,
	})
}

func login(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var req loginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "login",
			"status": "failure",
			"reason": "invalid_json",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid login input")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Failed to parse request body"})
		return
	}
	var user models.User
	cacheHit := false
	if val, ok := LoginCache.Get(req.Username); ok {
		user = val
		cacheHit = true
	} else {
		if err := models.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				auditLog.WithFields(logrus.Fields{
					"event":    "login",
					"status":   "failure",
					"reason":   "user_not_found",
					"username": req.Username,
					"ip":       ctx.ClientIP(),
				}).Warn("User not found during login")
				ctx.JSON(http.StatusUnauthorized, errorResponse{Error: "Invalid username or password"})
				return
			}
			auditLog.WithFields(logrus.Fields{
				"event":    "login",
				"status":   "failure",
				"reason":   "db_error",
				"username": req.Username,
				"ip":       ctx.ClientIP(),
				"error":    err.Error(),
			}).Error("Database error during login")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
			return
		} else {
			LoginCache.Set(req.Username, user)
		}
	}
	if user.Ban {
		auditLog.WithFields(logrus.Fields{
			"event":    "login",
			"status":   "failure",
			"reason":   "banned",
			"user_id":  user.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
		}).Warn("Banned user attempted login")
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "Account is banned"})
		return
	}
	isValid, err := user.ComparePassword(req.Password)
	if err != nil || !isValid {
		auditLog.WithFields(logrus.Fields{
			"event":    "login",
			"status":   "failure",
			"reason":   "invalid_password",
			"user_id":  user.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
			"error": func() string {
				if err != nil {
					return err.Error()
				} else {
					return ""
				}
			},
		}).Warn("Invalid password during login")
		ctx.JSON(http.StatusUnauthorized, errorResponse{Error: "Invalid username or password"})
		return
	}
	token, err := utils.GenerateJWT(user.ID, user.Username, values.GetConfig().Security.JWTSecret)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":    "login",
			"status":   "failure",
			"reason":   "token_generation_failed",
			"user_id":  user.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
			"error":    err.Error(),
		}).Error("Failed to generate JWT token during login")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to generate token"})
		return
	}
	userInfo := userInfo{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		GitHubUsername: user.GitHubUsername,
		TeamID:         user.TeamID,
	}
	auditLog.WithFields(logrus.Fields{
		"event":    "login",
		"status":   "success",
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"github":   user.GitHubUsername,
		"ip":       ctx.ClientIP(),
		"cache":    cacheHit,
	}).Info("User logged in successfully")
	ctx.JSON(http.StatusOK, authResponse{
		Token: token,
		User:  userInfo,
	})
}

func forgotPassword(ctx *gin.Context) {
	var input forgotPasswordRequest
	var user models.User
	auditLog := utils.Logger.WithField("type", "audit")
	if err := ctx.ShouldBindJSON(&input); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "forgot_password",
			"status": "failure",
			"reason": "invalid_json",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid forgot password input")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	otpSet := input.OTP != nil && *input.OTP != ""
	backupSet := input.BackupCode != nil && *input.BackupCode != ""
	if (otpSet && backupSet) || (!otpSet && !backupSet) {
		auditLog.WithFields(logrus.Fields{
			"event":    "forgot_password",
			"status":   "failure",
			"reason":   "invalid_auth_method_selection",
			"username": input.Username,
			"ip":       ctx.ClientIP(),
		}).Warn("Invalid OTP/Backup Code usage")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Provide either OTP or Backup Code, not both"})
		return
	}
	if value, ok := LoginCache.Get(input.Username); ok {
		ctx.Set("message", fmt.Sprintf("User %d loaded from login cache", value.ID))
		user = value
		auditLog.WithFields(logrus.Fields{
			"event":    "forgot_password",
			"status":   "info",
			"reason":   "cache_hit",
			"user_id":  user.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
		}).Info("User loaded from cache for password reset")
	} else {
		if err := models.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
			ctx.Set("message", err.Error())
			auditLog.WithFields(logrus.Fields{
				"event":    "forgot_password",
				"status":   "failure",
				"reason":   "user_not_found",
				"username": input.Username,
				"ip":       ctx.ClientIP(),
			}).Warn("User not found for forgot password")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
	}
	if otpSet && user.VerifyTOTP(*input.OTP) {
		ctx.Set("message", fmt.Sprintf("User %d resetting password using TOTP", user.ID))
		auditLog.WithFields(logrus.Fields{
			"event":    "forgot_password_auth",
			"status":   "success",
			"method":   "totp",
			"user_id":  user.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
		}).Info("Password reset authorized via TOTP")
	} else if backupSet && user.BackupCode == *input.BackupCode {
		ctx.Set("message", fmt.Sprintf("User %d resetting password using Backup code", user.ID))
		auditLog.WithFields(logrus.Fields{
			"event":    "forgot_password_auth",
			"status":   "success",
			"method":   "backup_code",
			"user_id":  user.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
		}).Info("Password reset authorized via Backup Code")
	} else {
		errMsg := "Invalid credentials"
		method := "unknown"
		if otpSet {
			errMsg = "Invalid OTP"
			method = "totp"
		} else if backupSet {
			errMsg = "Invalid Backup Code"
			method = "backup_code"
		}
		auditLog.WithFields(logrus.Fields{
			"event":    "forgot_password_auth",
			"status":   "failure",
			"reason":   "invalid_" + method,
			"user_id":  user.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
		}).Warn("Failed password reset authentication")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": errMsg})
		return
	}
	random := make([]byte, 20)
	_, err := rand.Read(random)
	if err != nil {
		ctx.Set("message", err.Error())
		auditLog.WithFields(logrus.Fields{
			"event":    "forgot_password",
			"status":   "failure",
			"reason":   "token_generation_failed",
			"user_id":  user.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
			"error":    err.Error(),
		}).Error("Failed to generate password reset token")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	token := hex.EncodeToString(random)
	ResetPasswordCache.Set(token, user)
	auditLog.WithFields(logrus.Fields{
		"event":    "forgot_password_token_issued",
		"status":   "success",
		"user_id":  user.ID,
		"username": user.Username,
		"ip":       ctx.ClientIP(),
	}).Info("Password reset token successfully issued")
	ctx.JSON(http.StatusOK, resetTokenResponse{
		ResetToken: token,
	})
}

func resetPassword(ctx *gin.Context) {
	token := ctx.Param("token")
	auditLog := utils.Logger.WithField("type", "audit")
	if token == "" {
		auditLog.WithFields(logrus.Fields{
			"event":  "reset_password",
			"status": "failure",
			"reason": "missing_token",
			"ip":     ctx.ClientIP(),
		}).Warn("Reset password request missing token")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing reset token"})
		return
	}
	user, ok := ResetPasswordCache.Get(token)
	if !ok {
		auditLog.WithFields(logrus.Fields{
			"event":  "reset_password",
			"status": "failure",
			"reason": "invalid_or_expired_token",
			"token":  token,
			"ip":     ctx.ClientIP(),
		}).Warn("Reset password token invalid or expired")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}
	var input resetPasswordRequest
	if err := ctx.ShouldBindJSON(&input); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":    "reset_password",
			"status":   "failure",
			"reason":   "invalid_json",
			"user_id":  user.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
		}).Warn("Invalid input during password reset")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := user.SetPassword(input.Password); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":    "reset_password",
			"status":   "failure",
			"reason":   "set_password_failed",
			"user_id":  user.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
			"error":    err.Error(),
		}).Error("Failed to hash new password")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set password"})
		return
	}
	if err := models.DB.Model(&user).Update("password", user.Password).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":    "reset_password",
			"status":   "failure",
			"reason":   "db_update_failed",
			"user_id":  user.ID,
			"username": user.Username,
			"ip":       ctx.ClientIP(),
			"error":    err.Error(),
		}).Error("Failed to update password in DB")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}
	ResetPasswordCache.Delete(token)
	LoginCache.Delete(user.Email)
	auditLog.WithFields(logrus.Fields{
		"event":    "reset_password",
		"status":   "success",
		"user_id":  user.ID,
		"username": user.Username,
		"ip":       ctx.ClientIP(),
	}).Info("Password reset successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully"})
}
