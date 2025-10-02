package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/leaderboard"
	"github.com/intraware/rodan/api/shared"
	"github.com/intraware/rodan/internal/models"
	"github.com/intraware/rodan/internal/notification"
	"github.com/intraware/rodan/internal/sandbox"
	"github.com/intraware/rodan/internal/types"
	"github.com/intraware/rodan/internal/utils"
	"github.com/intraware/rodan/internal/utils/values"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GetChallengeList godoc
// @Summary      Get challenge list
// @Description  Retrieves a list of visible challenges with basic information
// @Security     BearerAuth
// @Tags         challenges
// @Accept       json
// @Produce      json
// @Success      200  {object}  []models.Challenge
// @Failure      500  {object}  types.ErrorResponse
// @Router       /challenges [get]
func GetChallengeList(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var challenges []models.Challenge
	if err := models.DB.Select("id, name").Where("is_visible = ?", true).Find(&challenges).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "get_challenge_list",
			"status": "failure",
			"reason": "db_error",
			"ip":     ctx.ClientIP(),
			"error":  err.Error(),
		}).Error("Failed to fetch challenges")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to fetch challenges"})
		return
	}
	var challengeList = make([]challengeItem, len(challenges))
	for idx, challenge := range challenges {
		challengeList[idx] = challengeItem{
			ID:    challenge.ID,
			Title: challenge.Name,
		}
	}
	auditLog.WithFields(logrus.Fields{
		"event":  "get_challenge_list",
		"status": "success",
		"ip":     ctx.ClientIP(),
		"count":  len(challengeList),
	}).Info("Fetched challenge list successfully")
	ctx.JSON(http.StatusOK, challengeList)
}

// GetChallengeDetail godoc
// @Summary      Get challenge details
// @Description  Retrieves detailed information about a specific challenge
// @Security     BearerAuth
// @Tags         challenges
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Challenge ID"
// @Success      200  {object}  models.Challenge
// @Failure      404  {object}  types.ErrorResponse
// @Failure      500  {object}  types.ErrorResponse
// @Router       /challenges/{id} [get]
func GetChallengeDetail(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	userID := ctx.GetUint("user_id")
	var user models.User
	user, userCacheHit := shared.UserCache.Get(userID)
	if !userCacheHit {
		if err := models.DB.First(&user, userID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":    "get_challenge_detail",
				"status":   "failure",
				"reason":   "db_error_user_lookup",
				"user_id":  userID,
				"ip":       ctx.ClientIP(),
				"error":    err.Error(),
				"user_hit": userCacheHit,
			}).Error("Failed to fetch user from DB")
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	challengeIDStr := ctx.Param("id")
	id, err := strconv.ParseUint(challengeIDStr, 10, 64)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "get_challenge_detail",
			"status":    "failure",
			"reason":    "invalid_challenge_id",
			"user_id":   user.ID,
			"challenge": challengeIDStr,
			"ip":        ctx.ClientIP(),
			"user_hit":  userCacheHit,
		}).Warn("Invalid challenge ID in request")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid challenge ID"})
		return
	}
	challengeID := uint(id)
	if user.TeamID == nil {
		auditLog.WithFields(logrus.Fields{
			"event":    "get_challenge_detail",
			"status":   "failure",
			"reason":   "no_team",
			"user_id":  user.ID,
			"ip":       ctx.ClientIP(),
			"user_hit": userCacheHit,
		}).Warn("User is not part of a team")
		ctx.JSON(http.StatusForbidden, types.ErrorResponse{Error: "User should belong to a team"})
		return
	}
	key := fmt.Sprintf("%d:%d", *user.TeamID, challengeID)
	solved, solveCacheHit := shared.TeamSolvedCache.Get(key)
	if !solveCacheHit {
		err := models.DB.Where("team_id = ? AND challenge_id = ?", *user.TeamID, challengeID).First(&models.Solve{}).Error
		if err == nil {
			solved = true
			shared.TeamSolvedCache.Set(key, true)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			auditLog.WithFields(logrus.Fields{
				"event":     "get_challenge_detail",
				"status":    "partial_failure",
				"reason":    "db_error_solve_lookup",
				"user_id":   user.ID,
				"team_id":   *user.TeamID,
				"challenge": challengeID,
				"ip":        ctx.ClientIP(),
				"error":     err.Error(),
			}).Warn("Error checking solve status")
		}
	}
	challenge, challengeCacheHit := shared.ChallengeCache.Get(challengeID)
	if !challengeCacheHit {
		if err := models.DB.Where("is_visible = ?", true).First(&challenge, challengeID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				auditLog.WithFields(logrus.Fields{
					"event":     "get_challenge_detail",
					"status":    "failure",
					"reason":    "challenge_not_found",
					"user_id":   user.ID,
					"challenge": challengeID,
					"ip":        ctx.ClientIP(),
				}).Warn("Challenge not found")
				ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Challenge not found"})
				return
			}
			auditLog.WithFields(logrus.Fields{
				"event":     "get_challenge_detail",
				"status":    "failure",
				"reason":    "db_error_challenge_lookup",
				"user_id":   user.ID,
				"challenge": challengeID,
				"ip":        ctx.ClientIP(),
				"error":     err.Error(),
			}).Error("Error fetching challenge from DB")
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
			return
		}
		shared.ChallengeCache.Set(challengeID, challenge)
	}
	var points int
	if val, err := calcPoints(challenge.PointsMin, challenge.PointsMax, challenge.ID); err != nil {
		points = challenge.PointsMax
	} else {
		points = val
	}
	response := challengeDetail{
		ID:         challenge.ID,
		Name:       challenge.Name,
		Author:     challenge.Author,
		Desc:       challenge.Desc,
		Category:   challenge.Category,
		Difficulty: challenge.Difficulty,
		Points:     points,
		Solved:     solved,
	}
	auditLog.WithFields(logrus.Fields{
		"event":         "get_challenge_detail",
		"status":        "success",
		"user_id":       user.ID,
		"team_id":       *user.TeamID,
		"challenge":     challenge.ID,
		"solved":        solved,
		"user_hit":      userCacheHit,
		"solve_hit":     solveCacheHit,
		"challenge_hit": challengeCacheHit,
		"ip":            ctx.ClientIP(),
	}).Info("Fetched challenge detail successfully")
	ctx.JSON(http.StatusOK, response)
}

// GetChallengeConfig godoc
// @Summary      Get challenge configuration
// @Description  Retrieves configuration details for a challenge including ports, links, and runtime information
// @Security     BearerAuth
// @Tags         challenges
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Challenge ID"
// @Success      200  {object}  challengeConfigResponse
// @Failure      403  {object}  types.ErrorResponse
// @Failure      404  {object}  types.ErrorResponse
// @Failure      500  {object}  types.ErrorResponse
// @Router       /challenges/{id}/config [get]
func GetChallengeConfig(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	userID := ctx.GetUint("user_id")
	user, userCacheHit := shared.UserCache.Get(userID)
	if !userCacheHit {
		if err := models.DB.First(&user, userID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":    "get_challenge_config",
				"status":   "failure",
				"reason":   "db_error_user_lookup",
				"user_id":  userID,
				"ip":       ctx.ClientIP(),
				"error":    err.Error(),
				"user_hit": userCacheHit,
			}).Error("Failed to fetch user from DB")
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	challengeIDStr := ctx.Param("id")
	id, err := strconv.ParseUint(challengeIDStr, 10, 64)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "get_challenge_config",
			"status":    "failure",
			"reason":    "invalid_challenge_id",
			"user_id":   user.ID,
			"challenge": challengeIDStr,
			"ip":        ctx.ClientIP(),
		}).Warn("Invalid challenge ID in request")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid challenge ID"})
		return
	}
	challengeID := uint(id)
	if user.TeamID == nil {
		auditLog.WithFields(logrus.Fields{
			"event":    "get_challenge_config",
			"status":   "failure",
			"reason":   "no_team",
			"user_id":  user.ID,
			"ip":       ctx.ClientIP(),
			"user_hit": userCacheHit,
		}).Warn("User is not part of a team")
		ctx.JSON(http.StatusForbidden, types.ErrorResponse{Error: "User should belong to a team"})
		return
	}
	key := fmt.Sprintf("%d:%d", *user.TeamID, challengeID)
	solved, solveCacheHit := shared.TeamSolvedCache.Get(key)
	if !solveCacheHit {
		err := models.DB.Where("team_id = ? AND challenge_id = ?", *user.TeamID, challengeID).First(&models.Solve{}).Error
		if err == nil {
			solved = true
			shared.TeamSolvedCache.Set(key, true)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			auditLog.WithFields(logrus.Fields{
				"event":     "get_challenge_config",
				"status":    "partial_failure",
				"reason":    "db_error_solve_lookup",
				"user_id":   user.ID,
				"team_id":   *user.TeamID,
				"challenge": challengeID,
				"ip":        ctx.ClientIP(),
				"error":     err.Error(),
			}).Warn("Error checking solve status")
		}
	}
	challenge, challengeCacheHit := shared.ChallengeCache.Get(challengeID)
	if !challengeCacheHit {
		if err := models.DB.Where("is_visible = ?", true).First(&challenge, challengeID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				auditLog.WithFields(logrus.Fields{
					"event":     "get_challenge_config",
					"status":    "failure",
					"reason":    "challenge_not_found",
					"user_id":   user.ID,
					"challenge": challengeID,
					"ip":        ctx.ClientIP(),
				}).Warn("Challenge not found")
				ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Challenge not found"})
				return
			}
			auditLog.WithFields(logrus.Fields{
				"event":     "get_challenge_config",
				"status":    "failure",
				"reason":    "db_error_challenge_lookup",
				"user_id":   user.ID,
				"challenge": challengeID,
				"ip":        ctx.ClientIP(),
				"error":     err.Error(),
			}).Error("Error fetching challenge from DB")
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
			return
		}
		shared.ChallengeCache.Set(challengeID, challenge)
	}
	// for files .. store in links
	var response challengeConfigResponse
	if challenge.IsStatic {
		staticConfig, staticConfigCacheHit := shared.StaticConfig.Get(challenge.ID)
		if !staticConfigCacheHit {
			if err := models.DB.Where("challenge_id = ?", challenge.ID).First(&staticConfig).Error; err != nil {
				auditLog.WithFields(logrus.Fields{
					"event":         "get_challenge_config",
					"status":        "failure",
					"reason":        "db_error_static_config",
					"user_id":       user.ID,
					"team_id":       *user.TeamID,
					"challenge":     challengeID,
					"user_hit":      userCacheHit,
					"solve_hit":     solveCacheHit,
					"challenge_hit": challengeCacheHit,
					"ip":            ctx.ClientIP(),
					"error":         err.Error(),
				}).Error("Failed to retrieve static config from DB")
				ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to retrieve data from DB"})
				return
			} else {
				shared.StaticConfig.Set(challenge.ID, staticConfig)
			}
		}
		response = challengeConfigResponse{
			ID:       challenge.ID,
			Links:    staticConfig.Links,
			Ports:    staticConfig.Ports,
			IsStatic: true,
		}
		cacheTime := values.GetConfig().App.CacheDuration
		ctx.Header("Cache-Control", fmt.Sprintf("public,max-age=%.0f", cacheTime.Seconds()))
		ctx.Header("Expires", time.Now().Add(cacheTime).Format(http.TimeFormat))
		auditLog.WithFields(logrus.Fields{
			"event":             "get_challenge_config",
			"status":            "success",
			"user_id":           user.ID,
			"team_id":           *user.TeamID,
			"challenge":         challengeID,
			"solved":            solved,
			"user_hit":          userCacheHit,
			"solve_hit":         solveCacheHit,
			"challenge_hit":     challengeCacheHit,
			"static_config_hit": staticConfigCacheHit,
			"is_static":         true,
			"ip":                ctx.ClientIP(),
		}).Info("Fetched static challenge config successfully")
	} else {
		challengeSandbox, ok := shared.SandBoxMap[*user.TeamID]
		if !ok {
			auditLog.WithFields(logrus.Fields{
				"event":         "get_challenge_config",
				"status":        "failure",
				"reason":        "sandbox_not_created",
				"user_id":       user.ID,
				"team_id":       *user.TeamID,
				"challenge":     challengeID,
				"user_hit":      userCacheHit,
				"solve_hit":     solveCacheHit,
				"challenge_hit": challengeCacheHit,
				"is_static":     false,
				"ip":            ctx.ClientIP(),
			}).Warn("The user doesn't have a sandbox created")
			ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "The user doesnt have a sandbox created"})
			return
		}
		if !challengeSandbox.Active {
			auditLog.WithFields(logrus.Fields{
				"event":         "get_challenge_config",
				"status":        "failure",
				"reason":        "sandbox_not_active",
				"user_id":       user.ID,
				"team_id":       *user.TeamID,
				"challenge":     challengeID,
				"user_hit":      userCacheHit,
				"solve_hit":     solveCacheHit,
				"challenge_hit": challengeCacheHit,
				"is_static":     false,
				"ip":            ctx.ClientIP(),
			}).Warn("The user doesn't have sandbox active")
			ctx.JSON(http.StatusForbidden, types.ErrorResponse{Error: "The user doesnt have sandbox active"})
			return
		}
		var sandboxMeta sandbox.SandBoxResponse
		if val, err := challengeSandbox.GetMeta(); err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":         "get_challenge_config",
				"status":        "failure",
				"reason":        "sandbox_meta_error",
				"user_id":       user.ID,
				"team_id":       *user.TeamID,
				"challenge":     challengeID,
				"user_hit":      userCacheHit,
				"solve_hit":     solveCacheHit,
				"challenge_hit": challengeCacheHit,
				"is_static":     false,
				"ip":            ctx.ClientIP(),
				"error":         err.Error(),
			}).Error("Failed to get meta of sandbox")
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to get meta of sandbox"})
			return
		} else {
			sandboxMeta = val
		}
		response = challengeConfigResponse{
			ID:       challenge.ID,
			Links:    sandboxMeta.Links,
			TimeLeft: sandboxMeta.TimeLeft,
			IsStatic: false,
		}
		for v := range sandboxMeta.Ports {
			response.Ports = append(response.Ports, int(v))
		}
		auditLog.WithFields(logrus.Fields{
			"event":         "get_challenge_config",
			"status":        "success",
			"user_id":       user.ID,
			"team_id":       *user.TeamID,
			"challenge":     challengeID,
			"solved":        solved,
			"user_hit":      userCacheHit,
			"solve_hit":     solveCacheHit,
			"challenge_hit": challengeCacheHit,
			"is_static":     false,
			"ip":            ctx.ClientIP(),
		}).Info("Fetched dynamic challenge config successfully")
	}
	ctx.JSON(http.StatusOK, response)
}

// SubmitFlag godoc
// @Summary      Submit a flag for a challenge
// @Description  Submits a flag for validation against a specific challenge
// @Security     BearerAuth
// @Tags         challenges
// @Accept       json
// @Produce      json
// @Param        id    path      string              true   "Challenge ID"
// @Param        flag  body      submitFlagRequest   true   "Flag submission"
// @Success      200   {object}  submitFlagResponse
// @Failure      400   {object}  types.ErrorResponse
// @Failure      403   {object}  types.ErrorResponse
// @Failure      404   {object}  types.ErrorResponse
// @Failure      500   {object}  types.ErrorResponse
// @Router       /challenges/{id}/submit [post]
func SubmitFlag(ctx *gin.Context) {
	cfg := values.GetConfig().App
	auditLog := utils.Logger.WithField("type", "audit")
	challengeIDStr := ctx.Param("id")
	id, err := strconv.ParseUint(challengeIDStr, 10, 64)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "submit_flag",
			"status":    "failure",
			"reason":    "invalid_challenge_id",
			"challenge": challengeIDStr,
			"ip":        ctx.ClientIP(),
		}).Warn("Invalid challenge ID format")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid challenge ID"})
		return
	}
	challengeID := uint(id)
	var req submitFlagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "submit_flag",
			"status":    "failure",
			"reason":    "invalid_json",
			"challenge": challengeID,
			"ip":        ctx.ClientIP(),
		}).Warn("Failed to parse request body")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Failed to parse the body"})
		return
	}
	userID := ctx.GetUint("user_id")
	user, userCacheHit := shared.UserCache.Get(userID)
	if !userCacheHit {
		if err := models.DB.First(&user, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				auditLog.WithFields(logrus.Fields{
					"event":    "submit_flag",
					"status":   "failure",
					"reason":   "user_not_found",
					"user_id":  userID,
					"ip":       ctx.ClientIP(),
					"user_hit": userCacheHit,
				}).Warn("User not found during flag submission")
				ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "User not found"})
				return
			}
			auditLog.WithFields(logrus.Fields{
				"event":    "submit_flag",
				"status":   "failure",
				"reason":   "db_error_user",
				"user_id":  userID,
				"ip":       ctx.ClientIP(),
				"error":    err.Error(),
				"user_hit": userCacheHit,
			}).Error("DB error fetching user during flag submission")
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to get user data"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	if user.TeamID == nil {
		auditLog.WithFields(logrus.Fields{
			"event":    "submit_flag",
			"status":   "failure",
			"reason":   "no_team",
			"user_id":  user.ID,
			"ip":       ctx.ClientIP(),
			"user_hit": userCacheHit,
		}).Warn("User is not part of a team")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "User must be in a team to submit flags"})
		return
	}
	challenge, challengeCacheHit := shared.ChallengeCache.Get(challengeID)
	if !challengeCacheHit {
		if err := models.DB.Where("is_visible = ?", true).First(&challenge, challengeID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				auditLog.WithFields(logrus.Fields{
					"event":     "submit_flag",
					"status":    "failure",
					"reason":    "challenge_not_found",
					"user_id":   user.ID,
					"challenge": challengeID,
					"ip":        ctx.ClientIP(),
					"user_hit":  userCacheHit,
				}).Warn("Challenge not found during flag submission")
				ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Challenge not found"})
				return
			}
			auditLog.WithFields(logrus.Fields{
				"event":     "submit_flag",
				"status":    "failure",
				"reason":    "db_error_challenge",
				"user_id":   user.ID,
				"challenge": challengeID,
				"ip":        ctx.ClientIP(),
				"error":     err.Error(),
				"user_hit":  userCacheHit,
			}).Error("DB error fetching challenge during flag submission")
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to get challenge data"})
			return
		}
		shared.ChallengeCache.Set(challengeID, challenge)
	}
	teamID := *user.TeamID
	var existingSolve models.Solve
	err = models.DB.Where("team_id = ? AND challenge_id = ?", teamID, challengeID).First(&existingSolve).Error
	if err == nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "submit_flag",
			"status":    "failure",
			"reason":    "already_solved",
			"user_id":   user.ID,
			"team_id":   teamID,
			"challenge": challengeID,
			"ip":        ctx.ClientIP(),
		}).Warn("Team already solved the challenge")
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Challenge already solved by your team"})
		return
	}
	var correctFlag string
	var challengeType int8
	if challenge.IsStatic {
		if err := models.DB.Model(&challenge).Association("StaticConfig").Find(&challenge.StaticConfig); err != nil {
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to get static metadata from DB"})
			return
		}
		correctFlag = challenge.StaticConfig.Flag
		challengeType = 0
	} else {
		correctFlag = getDynamicFlag(challengeID, teamID)
		challengeType = 1
	}
	if req.Flag != correctFlag {
		auditLog.WithFields(logrus.Fields{
			"event":     "submit_flag",
			"status":    "failure",
			"reason":    "wrong_flag",
			"user_id":   user.ID,
			"team_id":   teamID,
			"challenge": challengeID,
			"ip":        ctx.ClientIP(),
		}).Info("Incorrect flag submitted")
		ctx.JSON(http.StatusOK, submitFlagResponse{
			Correct: false,
			Message: "Wrong flag! Try again.",
		})
		return
	}
	var flagValues []string
	dynFlagMap.Range(func(key, value any) bool {
		if req.Flag == value {
			flagValues = append(flagValues, value.(string))
		}
		return true
	})
	if len(flagValues) > 0 && (cfg.Ban.UserBan || cfg.Ban.TeamBan) {
		banReason := "submit_someone_flag"
		tx := models.DB.Begin()
		var count int64
		if cfg.Ban.UserBan {
			tx.Model(&models.BanHistory{}).
				Where("user_id = ?", user.ID).
				Count(&count)
		} else if cfg.Ban.TeamBan {
			tx.Model(&models.BanHistory{}).
				Where("team_id = ?", user.TeamID).
				Count(&count)
		}
		expiration := time.Now().Add(calcBanDuration(int(count))).Unix()
		if err := tx.Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Database error"})
			return
		}
		var key string
		var ban models.BanHistory
		ban.Context = banReason
		ban.ExpiresAt = expiration
		if cfg.Ban.UserBan {
			key = fmt.Sprintf("%d:%d", user.ID, *user.TeamID)
			user.Ban = true
			if err := tx.Save(&user).Error; err != nil {
				tx.Rollback()
				auditLog.WithFields(logrus.Fields{
					"event":     "user_ban",
					"status":    "failure",
					"reason":    "db_error_solve",
					"user_id":   user.ID,
					"team_id":   teamID,
					"challenge": challengeID,
					"ip":        ctx.ClientIP(),
					"error":     err.Error(),
				}).Error("Failed to ban user")
				ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to update user status"})
				return
			}
			ban.UserID = &userID
		}
		if cfg.Ban.TeamBan {
			key = fmt.Sprintf(":%d", teamID)
			if err := tx.Model(&models.Team{}).Where("id = ?", *user.TeamID).Update("ban", true).Error; err != nil {
				tx.Rollback()
				auditLog.WithFields(logrus.Fields{
					"event":     "team_ban",
					"status":    "failure",
					"reason":    "db_error_solve",
					"user_id":   user.ID,
					"team_id":   teamID,
					"challenge": challengeID,
					"ip":        ctx.ClientIP(),
					"error":     err.Error(),
				}).Error("Failed to ban team")
				ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to update team status"})
				return
			}
			ban.UserID = &userID
			ban.TeamID = user.TeamID
		}
		if err := tx.Create(&ban).Error; err != nil {
			tx.Rollback()
			auditLog.WithFields(logrus.Fields{
				"event":     "ban_history",
				"status":    "failure",
				"reason":    "db_error_solve",
				"user_id":   user.ID,
				"team_id":   teamID,
				"challenge": challengeID,
				"ip":        ctx.ClientIP(),
				"error":     err.Error(),
			}).Error("Failed to create ban history")
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to record ban history"})
			return
		} else {
			shared.BanHistoryCache.Set(key, ban)
			var sandboxes []*sandbox.SandBox
			if cfg.Ban.UserBan {
				for _, sandbox := range shared.SandBoxMap {
					if sandbox.UserID == *ban.UserID {
						sandboxes = append(sandboxes, sandbox)
					}
				}
			} else if cfg.Ban.TeamBan {
				for _, sandbox := range shared.SandBoxMap {
					if sandbox.TeamID == *ban.TeamID {
						sandboxes = append(sandboxes, sandbox)
					}
				}
			}
			stopAllContainers(sandboxes)
		}
		if err := tx.Commit().Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to finalize ban"})
			return
		}
		auditLog.WithFields(logrus.Fields{
			"event":     "submit_flag",
			"status":    "failure",
			"reason":    "submit_other_flag",
			"user_id":   user.ID,
			"team_id":   teamID,
			"challenge": challengeID,
			"ip":        ctx.ClientIP(),
		}).Error("Unauthorized flag submission — ban issued")
		shared.UserCache.Delete(user.ID)
		if user.TeamID != nil {
			shared.TeamCache.Delete(*user.TeamID)
		}
		if values.GetConfig().App.Ban.UserBan {
			ctx.JSON(http.StatusForbidden, types.ErrorResponse{Error: "Account got banned"})
		} else {
			ctx.JSON(http.StatusForbidden, types.ErrorResponse{Error: "Team account got banned"})
		}
	}
	var solveCount int64
	if err := models.DB.Model(&models.Solve{}).
		Where("challenge_id = ?", challengeID).
		Count(&solveCount).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "submit_flag",
			"status":    "failure",
			"reason":    "db_error_count",
			"user_id":   user.ID,
			"team_id":   teamID,
			"challenge": challengeID,
			"ip":        ctx.ClientIP(),
			"error":     err.Error(),
		}).Error("Failed to count challenge solves")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to record solve"})
		return
	}
	bloodCount := func() uint {
		if solveCount < 3 {
			return uint(solveCount + 1)
		} else {
			return 0
		}
	}()
	solve := models.Solve{
		TeamID:        teamID,
		ChallengeID:   challengeID,
		UserID:        userID,
		ChallengeType: challengeType,
		BloodCount:    bloodCount,
	}
	if err := models.DB.Create(&solve).Error; err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "submit_flag",
			"status":    "failure",
			"reason":    "db_error_solve",
			"user_id":   user.ID,
			"team_id":   teamID,
			"challenge": challengeID,
			"ip":        ctx.ClientIP(),
			"error":     err.Error(),
		}).Error("Failed to record solve")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to record solve"})
		return
	}
	if bloodCount > 0 && bloodCount <= 3 {
		auditLog.WithFields(logrus.Fields{
			"event":     "blood",
			"user_id":   user.ID,
			"team_id":   teamID,
			"challenge": challengeID,
			"blood":     bloodCount,
		}).Infof("Team got %d-blood on challenge", bloodCount)
		if bloodCount == 1 {
			notification.SendNotification(fmt.Sprintf("First Blood! %s solved %s", user.Username, challenge.Name))
		}
	}
	auditLog.WithFields(logrus.Fields{
		"event":          "submit_flag",
		"status":         "success",
		"user_id":        user.ID,
		"username":       user.Username,
		"team_id":        teamID,
		"challenge":      challengeID,
		"challenge_type": challengeType,
		"user_hit":       userCacheHit,
		"challenge_hit":  challengeCacheHit,
		"ip":             ctx.ClientIP(),
		"solved_at":      solve.CreatedAt,
	}).Info("Flag submitted successfully")
	leaderboard.MarkLeaderboardDirty()
	ctx.JSON(http.StatusOK, submitFlagResponse{
		Correct: true,
		Message: "Congratulations! Flag accepted.",
	})
}
