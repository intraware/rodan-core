package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/leaderboard"
	"github.com/intraware/rodan/api/shared"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func GetChallengeList(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	var challenges []models.Challenge
	if err := models.DB.Select("id, name").Find(&challenges).Error; err != nil {
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
		if err := models.DB.First(&challenge, challengeID).Error; err != nil {
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
	if challenge.IsStatic {
		if err := models.DB.Model(&challenge).Association("StaticConfig").Find(&challenge.StaticConfig); err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to get static metadata from DB"})
			return
		}
		response.Links = challenge.StaticConfig.Links
	} else {
		response.Links = []string{}
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
		if err := models.DB.First(&challenge, challengeID).Error; err != nil {
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
