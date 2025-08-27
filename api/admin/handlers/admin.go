package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/shared"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils"
	"github.com/sirupsen/logrus"
)

// GetAdmin godoc
// @Summary      Get admin profile
// @Description  Retrieves the profile information of the currently authenticated admin
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200  {object}  models.Admin
// @Failure      404  {object}  types.ErrorResponse
// @Router       /admin/me [get]
func GetAdmin(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	adminID := ctx.GetInt("admin_id")
	var admin models.Admin

	if err := models.DB.First(&admin, adminID).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":    "get_admin_profile",
			"status":   "failure",
			"reason":   "admin_not_found",
			"admin_id": adminID,
			"ip":       ctx.ClientIP(),
		}).Warn("Admin not found in getAdmin")
		ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Admin not found"})
		return
	}
}

// AddAdmin godoc
// @Summary      Add a new admin
// @Description  Creates a new admin account in the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        admin  body      models.Admin  true  "Admin object"
// @Success      201    {object}  models.Admin
// @Failure      400    {object}  types.ErrorResponse
// @Failure      500    {object}  types.ErrorResponse
// @Router       /admin/add [post]
func AddAdmin(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var admin models.Admin

	if err := ctx.ShouldBindJSON(&admin); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "add_admin",
			"status": "failure",
			"reason": "invalid_request",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid request in addAdmin")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid request"})
		return
	}

	if err := models.DB.Create(&admin).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "add_admin",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in addAdmin")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":    "add_admin",
		"status":   "success",
		"admin_id": admin.ID,
		"ip":       ctx.ClientIP(),
	}).Info("Admin added successfully")
	ctx.JSON(http.StatusCreated, admin)
}

// UpdateAdmin godoc
// @Summary      Update admin information
// @Description  Updates the profile information of the currently authenticated admin
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        admin  body      models.Admin  true  "Admin object"
// @Success      200    {object}  models.Admin
// @Failure      400    {object}  types.ErrorResponse
// @Failure      500    {object}  types.ErrorResponse
// @Router       /admin/edit [patch]
func UpdateAdmin(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	adminID := ctx.GetInt("admin_id")
	var admin models.Admin

	if err := ctx.ShouldBindJSON(&admin); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "update_admin",
			"status": "failure",
			"reason": "invalid_request",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid request in updateAdmin")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid request"})
		return
	}

	if err := models.DB.Model(&admin).Where("id = ?", adminID).Updates(admin).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "update_admin",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in updateAdmin")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":    "update_admin",
		"status":   "success",
		"admin_id": adminID,
		"ip":       ctx.ClientIP(),
	}).Info("Admin updated successfully")
	ctx.JSON(http.StatusOK, admin)
}

// DeleteAdmin godoc
// @Summary      Delete admin account
// @Description  Deletes the currently authenticated admin account
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      204
// @Failure      500  {object}  types.ErrorResponse
// @Router       /admin/delete [delete]
func DeleteAdmin(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	adminID := ctx.GetInt("admin_id")

	if err := models.DB.Delete(&models.Admin{}, adminID).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "delete_admin",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in deleteAdmin")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}

	auditLog.WithFields(logrus.Fields{
		"event":    "delete_admin",
		"status":   "success",
		"admin_id": adminID,
		"ip":       ctx.ClientIP(),
	}).Info("Admin deleted successfully")
	ctx.JSON(http.StatusNoContent, nil)
}

// FlushCache godoc
// @Summary      Flush system caches
// @Description  Flushes specific cache types or all caches in the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        cache_type  query     string  false  "Cache type to flush (user/team/challenge/login/static_config/team_solved/reset_password/all)"
// @Success      200         {object}  types.SuccessResponse
// @Failure      400         {object}  types.ErrorResponse
// @Router       /admin/flush_cache [post]
func flush_cache(ctx *gin.Context) {
	// take a parameter to flush specific cache objesct or all cache
	auditLog := utils.Logger.WithField("type", "audit")
	cacheType := ctx.Query("type")
	if cacheType == "" {
		auditLog.WithFields(logrus.Fields{
			"event":  "flush_cache",
			"status": "failure",
			"reason": "invalid_request",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid request in flushCache")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid request"})
		return
	}
	if cacheType == "all" {
		shared.UserCache.Reset()
		shared.TeamCache.Reset()
		shared.ChallengeCache.Reset()
		shared.LoginCache.Reset()
		shared.StaticConfig.Reset()
		shared.TeamSolvedCache.Reset()
		auditLog.WithFields(logrus.Fields{
			"event":  "flush_cache",
			"status": "success",
			"cache":  "all",
			"ip":     ctx.ClientIP(),
		}).Info("All caches flushed successfully")
		ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "All caches flushed successfully"})
		return
	}
	switch cacheType {
	case "user":
		shared.UserCache.Reset()
		auditLog.WithFields(logrus.Fields{
			"event":  "flush_cache",
			"status": "success",
			"cache":  "user",
			"ip":     ctx.ClientIP(),
		}).Info("User cache flushed successfully")
		ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "User cache flushed successfully"})
	case "team":
		shared.TeamCache.Reset()
		auditLog.WithFields(logrus.Fields{
			"event":  "flush_cache",
			"status": "success",
			"cache":  "team",
			"ip":     ctx.ClientIP(),
		}).Info("Team cache flushed successfully")
		ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Team cache flushed successfully"})
	case "challenge":
		shared.ChallengeCache.Reset()
		auditLog.WithFields(logrus.Fields{
			"event":  "flush_cache",
			"status": "success",
			"cache":  "challenge",
			"ip":     ctx.ClientIP(),
		}).Info("Challenge cache flushed successfully")
		ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Challenge cache flushed successfully"})
	case "login":
		shared.LoginCache.Reset()
		auditLog.WithFields(logrus.Fields{
			"event":  "flush_cache",
			"status": "success",
			"cache":  "login",
			"ip":     ctx.ClientIP(),
		}).Info("Login cache flushed successfully")
		ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Login cache flushed successfully"})
	case "static_config":
		shared.StaticConfig.Reset()
		auditLog.WithFields(logrus.Fields{
			"event":  "flush_cache",
			"status": "success",
			"cache":  "static_config",
			"ip":     ctx.ClientIP(),
		}).Info("Static config cache flushed successfully")
		ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Static config cache flushed successfully"})
	case "team_solved":
		shared.TeamSolvedCache.Reset()
		auditLog.WithFields(logrus.Fields{
			"event":  "flush_cache",
			"status": "success",
			"cache":  "team_solved",
			"ip":     ctx.ClientIP(),
		}).Info("Team solved cache flushed successfully")
		ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Team solved cache flushed successfully"})
	case "reset_password":
		shared.ResetPasswordCache.Reset()
		auditLog.WithFields(logrus.Fields{
			"event":  "flush_cache",
			"status": "success",
			"cache":  "reset_password",
			"ip":     ctx.ClientIP(),
		}).Info("Reset password cache flushed successfully")
		ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Reset password cache flushed successfully"})
	default:
		auditLog.WithFields(logrus.Fields{
			"event":  "flush_cache",
			"status": "failure",
			"reason": "invalid_cache_type",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid cache type in flushCache")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid cache type"})
		return
	}
}

