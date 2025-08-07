package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils"
	"github.com/sirupsen/logrus"
	"github.com/intraware/rodan/sandbox"
)

func GetAllContainers(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	container := sandbox.Container{}
	containers, err := container.GetAll()
	if err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "get_all_containers",
			"status":  "failure",
			"reason":  "internal_error",
			"ip":      ctx.ClientIP(),
		}).Error("Failed to get all containers")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to get all containers"})
		return
	}
	ctx.JSON(http.StatusOK, containers)
}

func StopAllContainers(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	container := sandbox.Container{}
	if err := container.StopAll(); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "stop_all_containers",
			"status":  "failure",
			"reason":  "internal_error",
			"ip":      ctx.ClientIP(),
		}).Error("Failed to stop all containers")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to stop all containers"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "stop_all_containers",
		"status":  "success",
		"ip":      ctx.ClientIP(),
	}).Info("All containers stopped successfully")
	ctx.JSON(http.StatusOK, successResponse{Message: "All containers stopped successfully"})
}

func KillAllContainers(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	container := sandbox.Container{}
	if err := container.KillAll(); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "kill_all_containers",
			"status":  "failure",
			"reason":  "internal_error",
			"ip":      ctx.ClientIP(),
		}).Error("Failed to kill all containers")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to kill all containers"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "kill_all_containers",
		"status":  "success",
		"ip":      ctx.ClientIP(),
	}).Info("All containers killed successfully")
	ctx.JSON(http.StatusOK, successResponse{Message: "All containers killed successfully"})
}

func StopContainer(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	containerID := ctx.Param("id")
	container := sandbox.Container{}
	if err := container.Stop(containerID); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "stop_container",
			"status":  "failure",
			"reason":  "internal_error",
			"container_id": containerID,
			"ip":      ctx.ClientIP(),
		}).Error("Failed to stop container")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to stop container"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "stop_container",
		"status":  "success",
		"container_id": containerID,
		"ip":      ctx.ClientIP(),
	}).Info("Container stopped successfully")
	ctx.JSON(http.StatusOK, successResponse{Message: "Container stopped successfully"})
}

func StopTeamContainer(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	teamID := ctx.Param("id")
	container := sandbox.Container{}
	if err := container.StopTeam(teamID); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "stop_team_container",
			"status":  "failure",
			"reason":  "internal_error",
			"team_id": teamID,
			"ip":      ctx.ClientIP(),
		}).Error("Failed to stop team container")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to stop team container"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "stop_team_container",
		"status":  "success",
		"team_id": teamID,
		"ip":      ctx.ClientIP(),
	}).Info("Team container stopped successfully")
	ctx.JSON(http.StatusOK, successResponse{Message: "Team container stopped successfully"})
}

func StopChallengeContainer(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	challengeID := ctx.Param("id")
	container := sandbox.Container{}
	if err := container.StopChallenge(challengeID); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "stop_challenge_container",
			"status":  "failure",
			"reason":  "internal_error",
			"challenge_id": challengeID,
			"ip":      ctx.ClientIP(),
		}).Error("Failed to stop challenge container")
		ctx.JSON(http.StatusInternalServerError, errorResponse{Error: "Failed to stop challenge container"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "stop_challenge_container",
		"status":  "success",
		"challenge_id": challengeID,
		"ip":      ctx.ClientIP(),
	}).Info("Challenge container stopped successfully")
	ctx.JSON(http.StatusOK, successResponse{Message: "Challenge container stopped successfully"})
}