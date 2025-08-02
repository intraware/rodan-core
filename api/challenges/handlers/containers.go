package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/shared"
	"github.com/intraware/rodan/internal/sandbox"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func StartDynamicChallenge(ctx *gin.Context) {
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
				"event":    "start_dynamic_challenge",
				"status":   "failure",
				"reason":   "db_error_user_lookup",
				"user_id":  userID,
				"ip":       ctx.ClientIP(),
				"error":    err.Error(),
				"user_hit": userCacheHit,
			}).Error("Failed to fetch user from DB")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database Error"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	challengeIDStr := ctx.Param("id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "start_dynamic_challenge",
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
			"event":    "start_dynamic_challenge",
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
					"event":     "start_dynamic_challenge",
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
	if solved {
		auditLog.WithFields(logrus.Fields{
			"event":     "start_dynamic_challenge",
			"status":    "failure",
			"reason":    "already_solved",
			"user_id":   user.ID,
			"team_id":   *user.TeamID,
			"challenge": challengeID,
			"user_hit":  userCacheHit,
			"solve_hit": solveCacheHit,
			"ip":        ctx.ClientIP(),
		}).Warn("Team has already solved the challenge")
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "The team has already solved the challenge"})
		return
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
					"event":     "start_dynamic_challenge",
					"status":    "failure",
					"reason":    "challenge_not_found",
					"user_id":   user.ID,
					"team_id":   *user.TeamID,
					"challenge": challengeID,
					"ip":        ctx.ClientIP(),
				}).Warn("Challenge not found")
				ctx.JSON(http.StatusNotFound, errorResponse{Error: "Challenge not found"})
				return
			}
			auditLog.WithFields(logrus.Fields{
				"event":     "start_dynamic_challenge",
				"status":    "failure",
				"reason":    "db_error_challenge_lookup",
				"user_id":   user.ID,
				"team_id":   *user.TeamID,
				"challenge": challengeID,
				"ip":        ctx.ClientIP(),
				"error":     err.Error(),
			}).Error("Error fetching challenge from DB")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
			return
		}
		shared.ChallengeCache.Set(challengeID, challenge)
	}
	if challenge.IsStatic {
		auditLog.WithFields(logrus.Fields{
			"event":         "start_dynamic_challenge",
			"status":        "failure",
			"reason":        "is_static",
			"user_id":       user.ID,
			"team_id":       *user.TeamID,
			"challenge":     challengeID,
			"ip":            ctx.ClientIP(),
			"user_hit":      userCacheHit,
			"solve_hit":     solveCacheHit,
			"challenge_hit": challengeCacheHit,
		}).Warn("Static challenges cannot spawn dynamic containers")
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "Static challenges cannot spawn dynamic containers."})
		return
	}
	if err := models.DB.Model(&challenge).Association("DynamicConfig").Find(&challenge.DynamicConfig); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "start_dynamic_challenge",
			"status":    "failure",
			"reason":    "dynamic_metadata_db_error",
			"user_id":   user.ID,
			"team_id":   *user.TeamID,
			"challenge": challengeID,
			"ip":        ctx.ClientIP(),
		}).Error("Failed to get dynamic metadata from DB")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to get dynamic metadata from DB"})
		return
	}
	if challenge.DynamicConfig.DockerImage == "" {
		auditLog.WithFields(logrus.Fields{
			"event":     "start_dynamic_challenge",
			"status":    "failure",
			"reason":    "invalid_docker_image",
			"user_id":   user.ID,
			"team_id":   *user.TeamID,
			"challenge": challengeID,
			"ip":        ctx.ClientIP(),
		}).Error("Invalid Docker Image in challenge config")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Invalid Docker Image is added to the list"})
		return
	}
	var challenge_sandbox *sandbox.SandBox
	flag := generateHashedFlag(challengeID, *user.TeamID)
	if val, ok := shared.SandBoxMap[*user.TeamID]; ok {
		challenge_sandbox = val
	} else {
		challenge_sandbox = sandbox.NewSandBox(userID, *user.TeamID, &challenge, flag)
		shared.SandBoxMap[*user.TeamID] = challenge_sandbox
	}
	if challenge_sandbox.Active {
		auditLog.WithFields(logrus.Fields{
			"event":     "start_dynamic_challenge",
			"status":    "failure",
			"reason":    "container_already_running",
			"user_id":   user.ID,
			"team_id":   *user.TeamID,
			"challenge": challengeID,
			"ip":        ctx.ClientIP(),
		}).Warn("Container is already running")
		ctx.JSON(http.StatusConflict, errorResponse{Error: "Container is already running"})
		return
	}
	if err := challenge_sandbox.Start(); err != nil {
		if errors.Is(err, sandbox.ErrFailedToCreateContainer) {
			auditLog.WithFields(logrus.Fields{
				"event":     "start_dynamic_challenge",
				"status":    "failure",
				"reason":    "failed_to_create_container",
				"user_id":   user.ID,
				"team_id":   *user.TeamID,
				"challenge": challengeID,
				"ip":        ctx.ClientIP(),
				"error":     err.Error(),
			}).Error("Failed to create the container")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to create the container"})
			return
		} else if errors.Is(err, sandbox.ErrFailedToStartContainer) {
			auditLog.WithFields(logrus.Fields{
				"event":     "start_dynamic_challenge",
				"status":    "failure",
				"reason":    "failed_to_start_container",
				"user_id":   user.ID,
				"team_id":   *user.TeamID,
				"challenge": challengeID,
				"ip":        ctx.ClientIP(),
				"error":     err.Error(),
			}).Error("Failed to start the container")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to start the container"})
			return
		}
	}
	auditLog.WithFields(logrus.Fields{
		"event":         "start_dynamic_challenge",
		"status":        "success",
		"user_id":       user.ID,
		"team_id":       *user.TeamID,
		"challenge":     challengeID,
		"user_hit":      userCacheHit,
		"solve_hit":     solveCacheHit,
		"challenge_hit": challengeCacheHit,
		"ip":            ctx.ClientIP(),
	}).Info("Started dynamic challenge container successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Container started successfully"})
}

func StopDynamicChallenge(ctx *gin.Context) {
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
				"event":    "stop_dynamic_challenge",
				"status":   "failure",
				"reason":   "db_error_user_lookup",
				"user_id":  userID,
				"ip":       ctx.ClientIP(),
				"error":    err.Error(),
				"user_hit": userCacheHit,
			}).Error("Failed to fetch user from DB")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database Error"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	challengeIDStr := ctx.Param("id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "stop_dynamic_challenge",
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
			"event":    "stop_dynamic_challenge",
			"status":   "failure",
			"reason":   "no_team",
			"user_id":  user.ID,
			"ip":       ctx.ClientIP(),
			"user_hit": userCacheHit,
		}).Warn("User is not part of a team")
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "User should belong to a team"})
		return
	}
	var challenge models.Challenge
	challengeCacheHit := false
	if val, ok := shared.ChallengeCache.Get(challengeID); ok {
		challenge = val
		challengeCacheHit = true
	} else {
		if err := models.DB.First(&challenge, challengeID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":     "stop_dynamic_challenge",
				"status":    "failure",
				"reason":    "db_error_challenge_lookup",
				"user_id":   user.ID,
				"team_id":   *user.TeamID,
				"challenge": challengeID,
				"ip":        ctx.ClientIP(),
				"error":     err.Error(),
			}).Error("Error fetching challenge from DB")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database Error"})
			return
		}
		shared.ChallengeCache.Set(challengeID, challenge)
	}
	if challenge.IsStatic {
		auditLog.WithFields(logrus.Fields{
			"event":         "stop_dynamic_challenge",
			"status":        "failure",
			"reason":        "is_static",
			"user_id":       user.ID,
			"team_id":       *user.TeamID,
			"challenge":     challengeID,
			"user_hit":      userCacheHit,
			"challenge_hit": challengeCacheHit,
			"ip":            ctx.ClientIP(),
		}).Warn("Static challenges cannot spawn dynamic containers")
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "Static challenges cannot spawn dynamic containers."})
		return
	}
	var challenge_sandbox *sandbox.SandBox
	if val, ok := shared.SandBoxMap[*user.TeamID]; ok {
		challenge_sandbox = val
	} else {
		auditLog.WithFields(logrus.Fields{
			"event":         "stop_dynamic_challenge",
			"status":        "failure",
			"reason":        "sandbox_not_created",
			"user_id":       user.ID,
			"team_id":       *user.TeamID,
			"challenge":     challengeID,
			"user_hit":      userCacheHit,
			"challenge_hit": challengeCacheHit,
			"ip":            ctx.ClientIP(),
		}).Warn("Sandbox is not created")
		ctx.JSON(http.StatusNotFound, errorResponse{Error: "Sanbox is not created"})
		return
	}
	if !challenge_sandbox.Active {
		auditLog.WithFields(logrus.Fields{
			"event":         "stop_dynamic_challenge",
			"status":        "failure",
			"reason":        "sandbox_not_running",
			"user_id":       user.ID,
			"team_id":       *user.TeamID,
			"challenge":     challengeID,
			"user_hit":      userCacheHit,
			"challenge_hit": challengeCacheHit,
			"ip":            ctx.ClientIP(),
		}).Warn("Sandbox is not running")
		ctx.JSON(http.StatusConflict, errorResponse{Error: "Sandbox is not runnning"})
		return
	}
	if challenge_sandbox.UserID != user.ID {
		auditLog.WithFields(logrus.Fields{
			"event":         "stop_dynamic_challenge",
			"status":        "failure",
			"reason":        "not_container_owner",
			"user_id":       user.ID,
			"team_id":       *user.TeamID,
			"challenge":     challengeID,
			"user_hit":      userCacheHit,
			"challenge_hit": challengeCacheHit,
			"ip":            ctx.ClientIP(),
		}).Warn("Only the user who started the container can stop it")
		ctx.JSON(http.StatusConflict, errorResponse{Error: "Only the user who started the container can stop it"})
		return
	}
	if err := challenge_sandbox.Stop(); err != nil {
		if errors.Is(err, sandbox.ErrContainerNotFound) {
			auditLog.WithFields(logrus.Fields{
				"event":         "stop_dynamic_challenge",
				"status":        "failure",
				"reason":        "container_not_found",
				"user_id":       user.ID,
				"team_id":       *user.TeamID,
				"challenge":     challengeID,
				"user_hit":      userCacheHit,
				"challenge_hit": challengeCacheHit,
				"ip":            ctx.ClientIP(),
				"error":         err.Error(),
			}).Error("Failed to find the sandbox container")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to find the sandbox container"})
			return
		} else if errors.Is(err, sandbox.ErrFailedToDiscardContainer) {
			auditLog.WithFields(logrus.Fields{
				"event":         "stop_dynamic_challenge",
				"status":        "failure",
				"reason":        "failed_to_discard_container",
				"user_id":       user.ID,
				"team_id":       *user.TeamID,
				"challenge":     challengeID,
				"user_hit":      userCacheHit,
				"challenge_hit": challengeCacheHit,
				"ip":            ctx.ClientIP(),
				"error":         err.Error(),
			}).Error("Failed to discard the sandbox container")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to discard the sandbox container"})
			return
		} else if errors.Is(err, sandbox.ErrFailedToStopContainer) {
			auditLog.WithFields(logrus.Fields{
				"event":         "stop_dynamic_challenge",
				"status":        "failure",
				"reason":        "failed_to_stop_container",
				"user_id":       user.ID,
				"team_id":       *user.TeamID,
				"challenge":     challengeID,
				"user_hit":      userCacheHit,
				"challenge_hit": challengeCacheHit,
				"ip":            ctx.ClientIP(),
				"error":         err.Error(),
			}).Error("Failed to stop the sandbox container")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to stop the sandbox container"})
			return
		}
	}
	auditLog.WithFields(logrus.Fields{
		"event":         "stop_dynamic_challenge",
		"status":        "success",
		"user_id":       user.ID,
		"team_id":       *user.TeamID,
		"challenge":     challengeID,
		"user_hit":      userCacheHit,
		"challenge_hit": challengeCacheHit,
		"ip":            ctx.ClientIP(),
	}).Info("Stopped dynamic challenge container successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Stopped container successfully"})
}

func ExtendDynamicChallenge(ctx *gin.Context) {
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
				"event":    "extend_dynamic_challenge",
				"status":   "failure",
				"reason":   "db_error_user_lookup",
				"user_id":  userID,
				"ip":       ctx.ClientIP(),
				"error":    err.Error(),
				"user_hit": userCacheHit,
			}).Error("Failed to fetch user from DB")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database Error"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	challengeIDStr := ctx.Param("id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "extend_dynamic_challenge",
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
			"event":    "extend_dynamic_challenge",
			"status":   "failure",
			"reason":   "no_team",
			"user_id":  user.ID,
			"ip":       ctx.ClientIP(),
			"user_hit": userCacheHit,
		}).Warn("User is not part of a team")
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "User should belong to a team"})
		return
	}
	var challenge models.Challenge
	challengeCacheHit := false
	if val, ok := shared.ChallengeCache.Get(challengeID); ok {
		challenge = val
		challengeCacheHit = true
	} else {
		if err := models.DB.First(&challenge, challengeID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":     "extend_dynamic_challenge",
				"status":    "failure",
				"reason":    "db_error_challenge_lookup",
				"user_id":   user.ID,
				"team_id":   *user.TeamID,
				"challenge": challengeID,
				"ip":        ctx.ClientIP(),
				"error":     err.Error(),
			}).Error("Error fetching challenge from DB")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database Error"})
			return
		}
		shared.ChallengeCache.Set(challengeID, challenge)
	}
	if challenge.IsStatic {
		auditLog.WithFields(logrus.Fields{
			"event":         "extend_dynamic_challenge",
			"status":        "failure",
			"reason":        "is_static",
			"user_id":       user.ID,
			"team_id":       *user.TeamID,
			"challenge":     challengeID,
			"user_hit":      userCacheHit,
			"challenge_hit": challengeCacheHit,
			"ip":            ctx.ClientIP(),
		}).Warn("Static challenges cannot spawn dynamic containers")
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "Static challenges cannot spawn dynamic containers."})
		return
	}
	var challenge_sandbox *sandbox.SandBox
	if val, ok := shared.SandBoxMap[*user.TeamID]; ok {
		challenge_sandbox = val
	} else {
		auditLog.WithFields(logrus.Fields{
			"event":         "extend_dynamic_challenge",
			"status":        "failure",
			"reason":        "sandbox_not_created",
			"user_id":       user.ID,
			"team_id":       *user.TeamID,
			"challenge":     challengeID,
			"user_hit":      userCacheHit,
			"challenge_hit": challengeCacheHit,
			"ip":            ctx.ClientIP(),
		}).Warn("Sandbox is not created")
		ctx.JSON(http.StatusNotFound, errorResponse{Error: "Sanbox is not created"})
		return
	}
	if !challenge_sandbox.Active {
		auditLog.WithFields(logrus.Fields{
			"event":         "extend_dynamic_challenge",
			"status":        "failure",
			"reason":        "sandbox_not_running",
			"user_id":       user.ID,
			"team_id":       *user.TeamID,
			"challenge":     challengeID,
			"user_hit":      userCacheHit,
			"challenge_hit": challengeCacheHit,
			"ip":            ctx.ClientIP(),
		}).Warn("Sandbox is not running")
		ctx.JSON(http.StatusConflict, errorResponse{Error: "Sandbox is not runnning"})
		return
	}
	challenge_sandbox.ExtendTTL() // test this
	auditLog.WithFields(logrus.Fields{
		"event":         "extend_dynamic_challenge",
		"status":        "success",
		"user_id":       user.ID,
		"team_id":       *user.TeamID,
		"challenge":     challengeID,
		"user_hit":      userCacheHit,
		"challenge_hit": challengeCacheHit,
		"ip":            ctx.ClientIP(),
	}).Info("Extended TTL of dynamic challenge container successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "TTL extended of the container"})
}

func RegenerateDynamicChallenge(ctx *gin.Context) {
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
				"event":    "regenerate_dynamic_challenge",
				"status":   "failure",
				"reason":   "db_error_user_lookup",
				"user_id":  userID,
				"ip":       ctx.ClientIP(),
				"error":    err.Error(),
				"user_hit": userCacheHit,
			}).Error("Failed to fetch user from DB")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database Error"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	challengeIDStr := ctx.Param("id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":     "regenerate_dynamic_challenge",
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
			"event":    "regenerate_dynamic_challenge",
			"status":   "failure",
			"reason":   "no_team",
			"user_id":  user.ID,
			"ip":       ctx.ClientIP(),
			"user_hit": userCacheHit,
		}).Warn("User is not part of a team")
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "User should belong to a team"})
		return
	}
	var challenge models.Challenge
	challengeCacheHit := false
	if val, ok := shared.ChallengeCache.Get(challengeID); ok {
		challenge = val
		challengeCacheHit = true
	} else {
		if err := models.DB.First(&challenge, challengeID).Error; err != nil {
			auditLog.WithFields(logrus.Fields{
				"event":     "regenerate_dynamic_challenge",
				"status":    "failure",
				"reason":    "db_error_challenge_lookup",
				"user_id":   user.ID,
				"team_id":   *user.TeamID,
				"challenge": challengeID,
				"ip":        ctx.ClientIP(),
				"error":     err.Error(),
			}).Error("Error fetching challenge from DB")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database Error"})
			return
		}
		shared.ChallengeCache.Set(challengeID, challenge)
	}
	if challenge.IsStatic {
		auditLog.WithFields(logrus.Fields{
			"event":         "regenerate_dynamic_challenge",
			"status":        "failure",
			"reason":        "is_static",
			"user_id":       user.ID,
			"team_id":       *user.TeamID,
			"challenge":     challengeID,
			"user_hit":      userCacheHit,
			"challenge_hit": challengeCacheHit,
			"ip":            ctx.ClientIP(),
		}).Warn("Static challenges cannot spawn dynamic containers")
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "Static challenges cannot spawn dynamic containers."})
		return
	}
	var challenge_sandbox *sandbox.SandBox
	if val, ok := shared.SandBoxMap[*user.TeamID]; ok {
		challenge_sandbox = val
	} else {
		auditLog.WithFields(logrus.Fields{
			"event":         "regenerate_dynamic_challenge",
			"status":        "failure",
			"reason":        "sandbox_not_created",
			"user_id":       user.ID,
			"team_id":       *user.TeamID,
			"challenge":     challengeID,
			"user_hit":      userCacheHit,
			"challenge_hit": challengeCacheHit,
			"ip":            ctx.ClientIP(),
		}).Warn("Sandbox is not created")
		ctx.JSON(http.StatusNotFound, errorResponse{Error: "Sanbox is not created"})
		return
	}
	if !challenge_sandbox.Active {
		auditLog.WithFields(logrus.Fields{
			"event":         "regenerate_dynamic_challenge",
			"status":        "failure",
			"reason":        "sandbox_not_running",
			"user_id":       user.ID,
			"team_id":       *user.TeamID,
			"challenge":     challengeID,
			"user_hit":      userCacheHit,
			"challenge_hit": challengeCacheHit,
			"ip":            ctx.ClientIP(),
		}).Warn("Sandbox is not running")
		ctx.JSON(http.StatusConflict, errorResponse{Error: "Sandbox is not runnning"})
		return
	}
	if err := challenge_sandbox.Regenerate(&challenge); err != nil {
		if errors.Is(err, sandbox.ErrContainerNotFound) {
			auditLog.WithFields(logrus.Fields{
				"event":         "regenerate_dynamic_challenge",
				"status":        "failure",
				"reason":        "container_not_found",
				"user_id":       user.ID,
				"team_id":       *user.TeamID,
				"challenge":     challengeID,
				"user_hit":      userCacheHit,
				"challenge_hit": challengeCacheHit,
				"ip":            ctx.ClientIP(),
				"error":         err.Error(),
			}).Warn("Container not found for regeneration")
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "Container not found"})
			return
		} else if errors.Is(err, sandbox.ErrFailedToDiscardContainer) {
			auditLog.WithFields(logrus.Fields{
				"event":         "regenerate_dynamic_challenge",
				"status":        "failure",
				"reason":        "failed_to_discard_container",
				"user_id":       user.ID,
				"team_id":       *user.TeamID,
				"challenge":     challengeID,
				"user_hit":      userCacheHit,
				"challenge_hit": challengeCacheHit,
				"ip":            ctx.ClientIP(),
				"error":         err.Error(),
			}).Error("Failed to discard container during regeneration")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to discard container"})
			return
		} else if errors.Is(err, sandbox.ErrFailedToCreateContainer) {
			auditLog.WithFields(logrus.Fields{
				"event":         "regenerate_dynamic_challenge",
				"status":        "failure",
				"reason":        "failed_to_create_container",
				"user_id":       user.ID,
				"team_id":       *user.TeamID,
				"challenge":     challengeID,
				"user_hit":      userCacheHit,
				"challenge_hit": challengeCacheHit,
				"ip":            ctx.ClientIP(),
				"error":         err.Error(),
			}).Error("Failed to create container during regeneration")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to create container"})
			return
		} else if errors.Is(err, sandbox.ErrFailedToStartContainer) {
			auditLog.WithFields(logrus.Fields{
				"event":         "regenerate_dynamic_challenge",
				"status":        "failure",
				"reason":        "failed_to_start_container",
				"user_id":       user.ID,
				"team_id":       *user.TeamID,
				"challenge":     challengeID,
				"user_hit":      userCacheHit,
				"challenge_hit": challengeCacheHit,
				"ip":            ctx.ClientIP(),
				"error":         err.Error(),
			}).Error("Failed to start container during regeneration")
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to start container"})
			return
		}
	}
	auditLog.WithFields(logrus.Fields{
		"event":         "regenerate_dynamic_challenge",
		"status":        "success",
		"user_id":       user.ID,
		"team_id":       *user.TeamID,
		"challenge":     challengeID,
		"user_hit":      userCacheHit,
		"challenge_hit": challengeCacheHit,
		"ip":            ctx.ClientIP(),
	}).Info("Regenerated dynamic challenge container successfully")
	ctx.JSON(http.StatusOK, gin.H{"message": "Regenerated container successfully"})
}
