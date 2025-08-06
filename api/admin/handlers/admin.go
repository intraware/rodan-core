package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils"
	"github.com/sirupsen/logrus"
)

func getAdmin(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	adminID := ctx.GetInt("admin_id")
	var admin models.Admin

	if err := models.DB.First(&admin, adminID).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "get_admin_profile",
			"status":  "failure",
			"reason":  "admin_not_found",
			"admin_id": adminID,
			"ip":      ctx.ClientIP(),
		}).Warn("Admin not found in getAdmin")
		ctx.JSON(http.StatusNotFound, errorResponse{Error: "Admin not found"})
		return
	}
}

func addAdmin(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var admin models.Admin

	if err := ctx.ShouldBindJSON(&admin); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "add_admin",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in addAdmin")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	if err := models.DB.Create(&admin).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "add_admin",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in addAdmin")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "add_admin",
		"status":  "success",
		"admin_id": admin.ID,
		"ip":      ctx.ClientIP(),
	}).Info("Admin added successfully")
	ctx.JSON(http.StatusCreated, admin)
}

func updateAdmin(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	adminID := ctx.GetInt("admin_id")
	var admin models.Admin

	if err := ctx.ShouldBindJSON(&admin); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "update_admin",
			"status":  "failure",
			"reason":  "invalid_request",
			"ip":      ctx.ClientIP(),
		}).Warn("Invalid request in updateAdmin")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid request"})
		return
	}

	if err := models.DB.Model(&admin).Where("id = ?", adminID).Updates(admin).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "update_admin",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in updateAdmin")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "update_admin",
		"status":  "success",
		"admin_id": adminID,
		"ip":      ctx.ClientIP(),
	}).Info("Admin updated successfully")
	ctx.JSON(http.StatusOK, admin)
}

func deleteAdmin(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	adminID := ctx.GetInt("admin_id")

	if err := models.DB.Delete(&models.Admin{}, adminID).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "delete_admin",
			"status":  "failure",
			"reason":  "database_error",
			"ip":      ctx.ClientIP(),
		}).Error("Database error in deleteAdmin")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":   "delete_admin",
		"status":  "success",
		"admin_id": adminID,
		"ip":      ctx.ClientIP(),
	}).Info("Admin deleted successfully")
	ctx.JSON(http.StatusNoContent, nil)
}