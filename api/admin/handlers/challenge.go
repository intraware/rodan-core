package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/internal/models"
	"github.com/intraware/rodan/internal/types"
	"github.com/intraware/rodan/internal/utils"
	"github.com/sirupsen/logrus"
)

func ToChallengeResponse(c models.Challenge) ChallengeResponse {
	return ChallengeResponse{
		ID:            c.ID,
		Name:          c.Name,
		Author:        c.Author,
		Desc:          c.Desc,
		Category:      c.Category,
		PointsMin:     c.PointsMin,
		PointsMax:     c.PointsMax,
		Difficulty:    c.Difficulty,
		IsStatic:      c.IsStatic,
		IsVisible:     c.IsVisible,
		StaticConfig:  c.StaticConfig,
		DynamicConfig: c.DynamicConfig,
		Hints:         c.Hints,
	}
}

// GetAllChallenges godoc
// @Summary      Get all challenges
// @Description  Retrieves a list of all challenges from the database
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200  {array}   ChallengeResponse
// @Failure      500  {object}  types.ErrorResponse
// @Router       /admin/challenges [get]
func GetAllChallenges(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var challenges []models.Challenge
	if err := models.DB.Preload("Hints").Preload("StaticConfig").Preload("DynamicConfig").Find(&challenges).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "get_all_challenges",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in getAllChallenges")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	var resp []ChallengeResponse
	for _, c := range challenges {
		resp = append(resp, ToChallengeResponse(c))
	}
	auditLog.WithFields(logrus.Fields{
		"event":  "get_all_challenges",
		"status": "success",
		"count":  len(resp),
		"ip":     ctx.ClientIP(),
	}).Info("Retrieved all challenges successfully")
	ctx.JSON(http.StatusOK, resp)
}

// AddChallenge godoc
// @Summary      Add a new challenge
// @Description  Creates a new challenge in the database
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        challenge  body      ChallengeResponse  true  "Challenge object"
// @Success      200        {object}  ChallengeResponse
// @Failure      400        {object}  types.ErrorResponse
// @Failure      500        {object}  types.ErrorResponse
// @Router       /admin/challenges [post]
func AddChallenge(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var req ChallengeResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "add_challenge",
			"status": "failure",
			"reason": "invalid_request",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid request in addChallenge")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid request"})
		return
	}
	challenge := models.Challenge{
		Name:          req.Name,
		Author:        req.Author,
		Desc:          req.Desc,
		Category:      req.Category,
		PointsMin:     req.PointsMin,
		PointsMax:     req.PointsMax,
		Difficulty:    req.Difficulty,
		IsStatic:      req.IsStatic,
		IsVisible:     req.IsVisible,
		StaticConfig:  req.StaticConfig,
		DynamicConfig: req.DynamicConfig,
		Hints:         req.Hints,
	}
	if err := models.DB.Create(&challenge).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "add_challenge",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in addChallenge")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":        "add_challenge",
		"status":       "success",
		"challenge_id": challenge.ID,
		"ip":           ctx.ClientIP(),
	}).Info("Challenge added successfully")
	ctx.JSON(http.StatusOK, ToChallengeResponse(challenge))
}

// UpdateChallenge godoc
// @Summary      Update a challenge
// @Description  Updates an existing challenge in the database
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id         path      int                true  "Challenge ID"
// @Param        challenge  body      ChallengeResponse  true  "Challenge object"
// @Success      200        {object}  ChallengeResponse
// @Failure      400        {object}  types.ErrorResponse
// @Failure      500        {object}  types.ErrorResponse
// @Router       /admin/challenges/{id} [patch]
func UpdateChallenge(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	id := ctx.Param("id")
	var req ChallengeResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "update_challenge",
			"status": "failure",
			"reason": "invalid_request",
			"ip":     ctx.ClientIP(),
		}).Warn("Invalid request in updateChallenge")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid request"})
		return
	}
	var challenge models.Challenge
	if err := models.DB.Preload("Hints").Preload("StaticConfig").Preload("DynamicConfig").First(&challenge, id).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Challenge not found"})
		return
	}
	// Update fields
	challenge.Name = req.Name
	challenge.Author = req.Author
	challenge.Desc = req.Desc
	challenge.Category = req.Category
	challenge.PointsMin = req.PointsMin
	challenge.PointsMax = req.PointsMax
	challenge.Difficulty = req.Difficulty
	challenge.IsStatic = req.IsStatic
	challenge.IsVisible = req.IsVisible
	challenge.StaticConfig = req.StaticConfig
	challenge.DynamicConfig = req.DynamicConfig
	challenge.Hints = req.Hints

	if err := models.DB.Save(&challenge).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "update_challenge",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in updateChallenge")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":        "update_challenge",
		"status":       "success",
		"challenge_id": challenge.ID,
		"ip":           ctx.ClientIP(),
	}).Info("Challenge updated successfully")
	ctx.JSON(http.StatusOK, ToChallengeResponse(challenge))
}

// DeleteChallenge godoc
// @Summary      Delete a challenge
// @Description  Deletes an existing challenge from the database
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id  path      int  true  "Challenge ID"
// @Success      200  {object}  types.SuccessResponse
// @Failure      400  {object}  types.ErrorResponse
// @Failure      500  {object}  types.ErrorResponse
// @Router       /admin/challenges/{id} [delete]
func DeleteChallenge(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	id := ctx.Param("id")
	var challenge models.Challenge
	if err := models.DB.First(&challenge, id).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Challenge not found"})
		return
	}
	if err := models.DB.Delete(&challenge).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "delete_challenge",
			"status": "failure",
			"reason": "database_error",
			"ip":     ctx.ClientIP(),
		}).Error("Database error in deleteChallenge")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":        "delete_challenge",
		"status":       "success",
		"challenge_id": challenge.ID,
		"ip":           ctx.ClientIP(),
	}).Info("Challenge deleted successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Challenge deleted successfully"})
}
