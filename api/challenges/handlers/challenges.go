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
	"github.com/intraware/rodan/internal/sandbox"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils"
	"github.com/intraware/rodan/utils/values"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

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
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to fetch challenges"})
		return
	}
	var challengeList []challengeItem
	for _, challenge := range challenges {
		challengeList = append(challengeList, challengeItem{
			ID:    challenge.ID,
			Title: challenge.Name,
		})
	}
	auditLog.WithFields(logrus.Fields{
		"event":  "get_challenge_list",
		"status": "success",
		"ip":     ctx.ClientIP(),
		"count":  len(challengeList),
	}).Info("Fetched challenge list successfully")
	ctx.JSON(http.StatusOK, challengeList)
}

func GetChallengeDetail(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	userID := ctx.GetInt("user_id")
	var user models.User
	userCacheHit := false
	if val, ok := shared.UserCache.Get(userID); ok {
		user = val
		userCacheHit = true
	} else {
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
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	challengeIDStr := ctx.Param("id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "get_challenge_detail",
			"status":    "failure",
			"reason":    "invalid_challenge_id",
			"user_id":   user.ID,
			"challenge": challengeIDStr,
			"ip":        ctx.ClientIP(),
		}).Warn("Invalid challenge ID in request")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid challenge ID"})
		return
	}
	if user.TeamID == nil {
		auditLog.WithFields(logrus.Fields{
			"event":    "get_challenge_detail",
			"status":   "failure",
			"reason":   "no_team",
			"user_id":  user.ID,
			"ip":       ctx.ClientIP(),
			"user_hit": userCacheHit,
		}).Warn("User is not part of a team")
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "User should belong to a team"})
		return
	}
	solved := false
	solveCacheHit := false
	{
		key := fmt.Sprintf("%d:%d", *user.TeamID, challengeID)
		var solve models.Solve
		if val, ok := shared.TeamSolvedCache.Get(key); ok && val {
			solved = true
			solveCacheHit = true
		} else {
			err := models.DB.Where("team_id = ? AND challenge_id = ?", *user.TeamID, challengeID).First(&solve).Error
			if err == nil {
				solved = true
				shared.TeamSolvedCache.Set(key, true)
			} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
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
	}
	var challenge models.Challenge
	challengeCacheHit := false
	if val, ok := shared.ChallengeCache.Get(challengeID); ok {
		challenge = val
		challengeCacheHit = true
	} else {
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
				ctx.JSON(http.StatusNotFound, errorResponse{Error: "Challenge not found"})
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
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
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

func GetChallengeConfig(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	userID := ctx.GetInt("user_id")
	var user models.User
	userCacheHit := false
	if val, ok := shared.UserCache.Get(userID); ok {
		user = val
		userCacheHit = true
	} else {
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
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	challengeIDStr := ctx.Param("id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "get_challenge_config",
			"status":    "failure",
			"reason":    "invalid_challenge_id",
			"user_id":   user.ID,
			"challenge": challengeIDStr,
			"ip":        ctx.ClientIP(),
		}).Warn("Invalid challenge ID in request")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid challenge ID"})
		return
	}
	if user.TeamID == nil {
		auditLog.WithFields(logrus.Fields{
			"event":    "get_challenge_config",
			"status":   "failure",
			"reason":   "no_team",
			"user_id":  user.ID,
			"ip":       ctx.ClientIP(),
			"user_hit": userCacheHit,
		}).Warn("User is not part of a team")
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "User should belong to a team"})
		return
	}
	solved := false
	solveCacheHit := false
	{
		key := fmt.Sprintf("%d:%d", *user.TeamID, challengeID)
		var solve models.Solve
		if val, ok := shared.TeamSolvedCache.Get(key); ok && val {
			solved = true
			solveCacheHit = true
		} else {
			err := models.DB.Where("team_id = ? AND challenge_id = ?", *user.TeamID, challengeID).First(&solve).Error
			if err == nil {
				solved = true
				shared.TeamSolvedCache.Set(key, true)
			} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
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
	}
	var challenge models.Challenge
	challengeCacheHit := false
	if val, ok := shared.ChallengeCache.Get(challengeID); ok {
		challenge = val
		challengeCacheHit = true
	} else {
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
				ctx.JSON(http.StatusNotFound, errorResponse{Error: "Challenge not found"})
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
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
			return
		}
		shared.ChallengeCache.Set(challengeID, challenge)
	}
	// for files .. store in links
	var response challengeConfigResponse
	if challenge.IsStatic {
		var static_config models.StaticConfig
		staticConfigCacheHit := false
		if val, ok := shared.StaticConfig.Get(challenge.ID); ok {
			static_config = val
			staticConfigCacheHit = true
		} else {
			if err := models.DB.Where("challenge_id = ?", challenge.ID).First(&static_config).Error; err != nil {
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
				ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to retrieve data from DB"})
				return
			} else {
				shared.StaticConfig.Set(challenge.ID, static_config)
			}
		}
		response = challengeConfigResponse{
			ID:       challenge.ID,
			Links:    static_config.Links,
			Ports:    static_config.Ports,
			IsStatic: true,
		}
		cache_time := values.GetConfig().App.CacheDuration
		ctx.Header("Cache-Control", fmt.Sprintf("public,max-age=%.0f", cache_time.Seconds()))
		ctx.Header("Expires", time.Now().Add(cache_time).Format(http.TimeFormat))
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
		var challenge_sandbox *sandbox.SandBox
		if val, ok := shared.SandBoxMap[*user.TeamID]; ok {
			challenge_sandbox = val
		} else {
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
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "The user doesnt have a sandbox created"})
			return
		}
		if !challenge_sandbox.Active {
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
			ctx.JSON(http.StatusForbidden, errorResponse{Error: "The user doesnt have sandbox active"})
			return
		}
		var sandbox_meta sandbox.SandBoxResponse
		if val, err := challenge_sandbox.GetMeta(); err != nil {
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
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to get meta of sandbox"})
			return
		} else {
			sandbox_meta = val
		}
		response = challengeConfigResponse{
			ID:       challenge.ID,
			Links:    sandbox_meta.Links,
			TimeLeft: sandbox_meta.TimeLeft,
			IsStatic: false,
		}
		for v := range sandbox_meta.Ports {
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

func SubmitFlag(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	challengeIDStr := ctx.Param("id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "submit_flag",
			"status":    "failure",
			"reason":    "invalid_challenge_id",
			"challenge": challengeIDStr,
			"ip":        ctx.ClientIP(),
		}).Warn("Invalid challenge ID format")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid challenge ID"})
		return
	}
	var req submitFlagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "submit_flag",
			"status":    "failure",
			"reason":    "invalid_json",
			"challenge": challengeID,
			"ip":        ctx.ClientIP(),
		}).Warn("Failed to parse request body")
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Failed to parse the body"})
		return
	}
	userID := ctx.GetInt("user_id")
	var user models.User
	userCacheHit := false
	if val, ok := shared.UserCache.Get(userID); ok {
		user = val
		userCacheHit = true
	} else {
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
				ctx.JSON(http.StatusNotFound, errorResponse{Error: "User not found"})
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
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to get user data"})
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
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "User must be in a team to submit flags"})
		return
	}
	var challenge models.Challenge
	challengeCacheHit := false
	if val, ok := shared.ChallengeCache.Get(challengeID); ok {
		challenge = val
		challengeCacheHit = true
	} else {
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
				ctx.JSON(http.StatusNotFound, errorResponse{Error: "Challenge not found"})
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
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to get challenge data"})
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
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Challenge already solved by your team"})
		return
	}
	var correctFlag string
	var challengeType int8
	if challenge.IsStatic {
		if err := models.DB.Model(&challenge).Association("StaticConfig").Find(&challenge.StaticConfig); err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to get static metadata from DB"})
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
	var flag_values []string
	dynFlagMap.Range(func(key, value any) bool {
		if req.Flag == value {
			flag_values = append(flag_values, value.(string))
		}
		return true
	})
	if len(flag_values) > 0 {
		if values.GetConfig().App.Ban.UserBan {
			user.Ban = true
			if err := models.DB.Save(&user).Error; err != nil {
				auditLog.WithFields(logrus.Fields{
					"event":     "user_ban",
					"status":    "failure",
					"reason":    "db_error_solve",
					"user_id":   user.ID,
					"team_id":   teamID,
					"challenge": challengeID,
					"ip":        ctx.ClientIP(),
					"error":     err.Error(),
				}).Error("Failed to record User ban")
				ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to update user status"})
				return
			} else {
				auditLog.WithFields(logrus.Fields{
					"event":     "submit_flag",
					"status":    "failure",
					"reason":    "submit_other_flag",
					"user_id":   user.ID,
					"team_id":   teamID,
					"challenge": challengeID,
					"ip":        ctx.ClientIP(),
				}).Error("User has submitted someone else's flag")
				shared.UserCache.Delete(user.ID)
				ctx.JSON(http.StatusForbidden, errorResponse{Error: "Account got banned"})
				return
			}
		} else if values.GetConfig().App.Ban.TeamBan {
			if err := models.DB.Where("id = ?", *user.TeamID).Update("ban", true).Error; err != nil {
				auditLog.WithFields(logrus.Fields{
					"event":     "team_ban",
					"status":    "failure",
					"reason":    "db_error_solve",
					"user_id":   user.ID,
					"team_id":   teamID,
					"challenge": challengeID,
					"ip":        ctx.ClientIP(),
					"error":     err.Error(),
				}).Error("Failed to record Team ban")
				ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to update team status"})
				return
			} else {
				auditLog.WithFields(logrus.Fields{
					"event":     "submit_flag",
					"status":    "failure",
					"reason":    "submit_other_flag",
					"user_id":   user.ID,
					"team_id":   teamID,
					"challenge": challengeID,
					"ip":        ctx.ClientIP(),
				}).Error("Team has submitted someone else's flag")
				shared.TeamCache.Delete(*user.TeamID)
				ctx.JSON(http.StatusForbidden, errorResponse{Error: "Team account got banned"})
				return
			}
		}
	}
	solve := models.Solve{
		TeamID:        teamID,
		ChallengeID:   challengeID,
		UserID:        userID,
		ChallengeType: challengeType,
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
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to record solve"})
		return
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
