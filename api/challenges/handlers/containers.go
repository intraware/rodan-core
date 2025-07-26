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
	if challenge.DynamicConfig.DockerImage == "" {
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Invalid Docker Image is added to the list"})
		return
	}
	var challenge_sandbox *sandbox.SandBox
	if val, ok := shared.SandBoxMap[*user.TeamID]; ok {
		challenge_sandbox = val
	} else {
		challenge_sandbox = sandbox.NewSandBox(userID, *user.TeamID, &challenge)
		shared.SandBoxMap[*user.TeamID] = challenge_sandbox
	}
	if challenge_sandbox.Active {
		ctx.JSON(http.StatusConflict, errorResponse{Error: "Container is already running"})
		return
	}
	if err := challenge_sandbox.Start(); err != nil {
		if errors.Is(err, sandbox.ErrFailedToCreateContainer) {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to create the container"})
			return
		} else if errors.Is(err, sandbox.ErrFailedToStartContainer) {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to start the container"})
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Container started successfully"})
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
	var challenge_sandbox *sandbox.SandBox
	if val, ok := shared.SandBoxMap[*user.TeamID]; ok {
		challenge_sandbox = val
	} else {
		ctx.JSON(http.StatusNotFound, errorResponse{Error: "Sanbox is not created"})
		return
	}
	if !challenge_sandbox.Active {
		ctx.JSON(http.StatusConflict, errorResponse{Error: "Sandbox is not runnning"})
		return
	}
	if challenge_sandbox.UserID != user.ID {
		ctx.JSON(http.StatusConflict, errorResponse{Error: "Only the user who started the container can stop it"})
		return
	}
	if err := challenge_sandbox.Stop(); err != nil {
		if errors.Is(err, sandbox.ErrContainerNotFound) {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to find the sandbox container"})
			return
		} else if errors.Is(err, sandbox.ErrFailedToDiscardContainer) {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to discard the sandbox container"})
			return
		} else if errors.Is(err, sandbox.ErrFailedToStopContainer) {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to stop the sandbox container"})
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Stopped container successfully"})
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
	var challenge_sandbox *sandbox.SandBox
	if val, ok := shared.SandBoxMap[*user.TeamID]; ok {
		challenge_sandbox = val
	} else {
		ctx.JSON(http.StatusNotFound, errorResponse{Error: "Sanbox is not created"})
		return
	}
	if !challenge_sandbox.Active {
		ctx.JSON(http.StatusConflict, errorResponse{Error: "Sandbox is not runnning"})
		return
	}
	challenge_sandbox.ExtendTTL() // test this
	ctx.JSON(http.StatusOK, gin.H{"message": "TTL extended of the container"})
}

func RegenerateDynamicChallenge(ctx *gin.Context) {
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
	var challenge_sandbox *sandbox.SandBox
	if val, ok := shared.SandBoxMap[*user.TeamID]; ok {
		challenge_sandbox = val
	} else {
		ctx.JSON(http.StatusNotFound, errorResponse{Error: "Sanbox is not created"})
		return
	}
	if !challenge_sandbox.Active {
		ctx.JSON(http.StatusConflict, errorResponse{Error: "Sandbox is not runnning"})
		return
	}
	if err := challenge_sandbox.Regenerate(&challenge); err != nil {
		if errors.Is(err, sandbox.ErrContainerNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse{Error: "Container not found"})
			return
		} else if errors.Is(err, sandbox.ErrFailedToDiscardContainer) {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to discard container"})
			return
		} else if errors.Is(err, sandbox.ErrFailedToCreateContainer) {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to create container"})
			return
		} else if errors.Is(err, sandbox.ErrFailedToStartContainer) {
			ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to start container"})
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Regenerated container successfully"})
}
