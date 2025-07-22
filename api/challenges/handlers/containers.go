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
	"gorm.io/gorm"
)

func StartDynamicChallenge(ctx *gin.Context) {
	userID := ctx.GetInt("user_id")
	var user models.User
	if val, ok := shared.UserCache.Get(userID); ok {
		user = val
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database Error"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	challengeIDStr := ctx.Param("id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid challenge ID"})
		return
	}
	if user.TeamID == nil {
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "User should belong to a team"})
		return
	}
	solved := false
	{
		key := fmt.Sprintf("%d:%d", *user.TeamID, challengeID)
		var solve models.Solve
		if val, ok := shared.TeamSolvedCache.Get(key); ok && val {
			solved = true
		} else {
			err := models.DB.Where("team_id = ? AND challenge_id = ?", *user.TeamID, challengeID).First(&solve).Error
			if err == nil {
				solved = true
				shared.TeamSolvedCache.Set(key, true)
			} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				// log it
			}
		}
	}
	if solved {
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "The team has already solved the challenge"})
		return
	}
	var challenge models.Challenge
	if val, ok := shared.ChallengeCache.Get(challengeID); ok {
		challenge = val
	} else {
		if err := models.DB.First(&challenge, challengeID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				ctx.JSON(http.StatusNotFound, errorResponse{Error: "Challenge not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database error"})
			return
		}
		shared.ChallengeCache.Set(challengeID, challenge)
	}
	if challenge.IsStatic {
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "Static challenges cannot spawn dynamic containers."})
		return
	}
	if err := models.DB.Model(&challenge).Association("DynamicConfig").Find(&challenge.DynamicConfig); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to get dynamic metadata from DB"})
		return
	}
	sandbox.StartSandBox(userID, *user.TeamID, &challenge) // not done
}

func StopDynamicChallenge(ctx *gin.Context) {
	userID := ctx.GetInt("user_id")
	var user models.User
	if val, ok := shared.UserCache.Get(userID); ok {
		user = val
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database Error"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	challengeIDStr := ctx.Param("id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid challenge ID"})
		return
	}
	if user.TeamID == nil {
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "User should belong to a team"})
		return
	}
	var challenge models.Challenge
	if val, ok := shared.ChallengeCache.Get(challengeID); ok {
		challenge = val
	} else {
		if err := models.DB.First(&challenge, challengeID).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database Error"})
			return
		}
		shared.ChallengeCache.Set(challengeID, challenge)
	}
	sandbox.ReleaseSandBox(userID, *user.TeamID, challengeID)
}

func ExtendDynamicChallenge(ctx *gin.Context) {
	userID := ctx.GetInt("user_id")
	var user models.User
	if val, ok := shared.UserCache.Get(userID); ok {
		user = val
	} else {
		if err := models.DB.First(&user, userID).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database Error"})
			return
		}
		shared.UserCache.Set(userID, user)
	}
	challengeIDStr := ctx.Param("id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse{Error: "Invalid challenge ID"})
		return
	}
	if user.TeamID == nil {
		ctx.JSON(http.StatusForbidden, errorResponse{Error: "User should belong to a team"})
		return
	}
	var challenge models.Challenge
	if val, ok := shared.ChallengeCache.Get(challengeID); ok {
		challenge = val
	} else {
		if err := models.DB.First(&challenge, challengeID).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Database Error"})
			return
		}
		shared.ChallengeCache.Set(challengeID, challenge)
	}
	sandbox.ExtendSandBoxTTL(userID, *user.TeamID, challengeID)
}
