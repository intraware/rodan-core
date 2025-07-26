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
	
}