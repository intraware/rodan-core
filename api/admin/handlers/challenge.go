package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils"
	"github.com/sirupsen/logrus"
)

func GetAllChallenges(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var challenges []models.Challenge

	if err := models.DB.Find(&challenges).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "get_all_challenges",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in getAllChallenges")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "get_all_challenges",
		"status":  "success",
		"count":   len(challenges),
		"ip":      ctx.ClientIP(),
	}).Info("Retrieved all challenges successfully")
	ctx.JSON(http.StatusOK, challenges)
}

func AddChallenge(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var challenge models.Challenge

	if err := ctx.ShouldBindJSON(&challenge); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "add_challenge",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in addChallenge")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	if err := models.DB.Create(&challenge).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "add_challenge",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in addChallenge")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "add_challenge",
		"status":  "success",
		"challenge_id": challenge.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Challenge added successfully")
	ctx.JSON(http.StatusOK, challenge)
}

func UpdateChallenge(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var challenge models.Challenge

	if err := ctx.ShouldBindJSON(&challenge); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "update_challenge",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in updateChallenge")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	if err := models.DB.Save(&challenge).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "update_challenge",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in updateChallenge")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "update_challenge",
		"status":  "success",
		"challenge_id": challenge.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Challenge updated successfully")
	ctx.JSON(http.StatusOK, challenge)
}

func DeleteChallenge(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var challenge models.Challenge

	if err := ctx.ShouldBindJSON(&challenge); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "delete_challenge",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in deleteChallenge")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	if err := models.DB.Delete(&challenge).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "delete_challenge",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in deleteChallenge")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "delete_challenge",
		"status":  "success",
		"challenge_id": challenge.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Challenge deleted successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Challenge deleted successfully"})
}