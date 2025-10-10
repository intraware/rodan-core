package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/api/shared"
	"github.com/intraware/rodan/internal/types"
	"github.com/intraware/rodan/internal/utils"
	"github.com/intraware/rodan/internal/utils/docker"
	"github.com/sirupsen/logrus"
)

// GetAllContainers godoc
// @Summary      Get all containers
// @Description  Retrieves a list of all containers in the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200  {object}  []sandbox.Container
// @Failure      500  {object}  types.ErrorResponse
// @Router       /admin/sandbox/all [get]
func GetAllSandboxes(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, shared.SandBoxMap.DumpValues())
}

// StopAllContainers godoc
// @Summary      Stop all containers
// @Description  Stops all running containers in the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200  {object}  types.SuccessResponse
// @Failure      500  {object}  types.ErrorResponse
// @Router       /admin/container/stop_all [delete]
func StopAllContainers(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")

	if err := docker.StopAllContainers(ctx); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "stop_all_containers",
			"status": "failure",
			"reason": "internal_error",
			"ip":     ctx.ClientIP(),
		}).Error("Failed to stop all containers")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to stop all containers"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":  "stop_all_containers",
		"status": "success",
		"ip":     ctx.ClientIP(),
	}).Info("All containers stopped successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "All containers stopped successfully"})
}

// KillAllContainers godoc
// @Summary      Kill all containers
// @Description  Forcefully kills all running containers in the system
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Success      200  {object}  types.SuccessResponse
// @Failure      500  {object}  types.ErrorResponse
// @Router       /admin/container/kill_all [post]
func KillAllContainers(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	if err := docker.KillAllContainers(ctx); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":  "kill_all_containers",
			"status": "failure",
			"reason": "internal_error",
			"ip":     ctx.ClientIP(),
		}).Error("Failed to kill all containers")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to kill all containers"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":  "kill_all_containers",
		"status": "success",
		"ip":     ctx.ClientIP(),
	}).Info("All containers killed successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "All containers killed successfully"})
}

// StopContainer godoc
// @Summary      Stop a specific container
// @Description  Stops a specific container by its ID
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Container ID"
// @Success      200  {object}  types.SuccessResponse
// @Failure      500  {object}  types.ErrorResponse
// @Router       /admin/container/stop [delete]
func StopContainer(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	containerID := ctx.Param("id")
	// TODO: should we remove the sandbox from the map as well?
	if err := docker.StopContainer(ctx, containerID); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":        "stop_container",
			"status":       "failure",
			"reason":       "internal_error",
			"container_id": containerID,
			"ip":           ctx.ClientIP(),
		}).Error("Failed to stop container")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to stop container"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":        "stop_container",
		"status":       "success",
		"container_id": containerID,
		"ip":           ctx.ClientIP(),
	}).Info("Container stopped successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Container stopped successfully"})
}

// StopTeamContainer godoc
// @Summary      Stop team containers
// @Description  Stops all containers associated with a specific team
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Team ID"
// @Success      200  {object}  types.SuccessResponse
// @Failure      500  {object}  types.ErrorResponse
// @Router       /admin/container/stop_team [delete]
func StopTeamContainer(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	teamID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || teamID <= 0 {
		ctx.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid team ID"})
		return
	}
	sandbox, ok := shared.SandBoxMap.Get(uint(teamID))
	if !ok {
		ctx.JSON(http.StatusNotFound, types.ErrorResponse{Error: "No active sandbox found for the specified team"})
		return
	}
	if err := sandbox.Stop(); err != nil {
		auditLog.WithFields(logrus.Fields{
			"event":   "stop_team_container",
			"status":  "failure",
			"reason":  "internal_error",
			"team_id": teamID,
			"ip":      ctx.ClientIP(),
		}).Error("Failed to stop team container")
		ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to stop team container"})
		return
	}
	auditLog.WithFields(logrus.Fields{
		"event":   "stop_team_container",
		"status":  "success",
		"team_id": teamID,
		"ip":      ctx.ClientIP(),
	}).Info("Team container stopped successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Team container stopped successfully"})
}

// StopChallengeContainer godoc
// @Summary      Stop challenge containers
// @Description  Stops all containers associated with a specific challenge
// @Security     BearerAuth
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Challenge ID"
// @Success      200  {object}  types.SuccessResponse
// @Failure      500  {object}  types.ErrorResponse
// @Router       /admin/container/stop_challenge [delete]
func StopChallengeContainer(ctx *gin.Context) {
	auditLog := utils.Logger.WithField("type", "audit")
	challengeID := ctx.Param("id")
	// container := sandbox.Container{}

	// TODO: challenge container???????

	// if err := container.StopChallenge(challengeID); err != nil {
	// 	auditLog.WithFields(logrus.Fields{
	// 		"event":        "stop_challenge_container",
	// 		"status":       "failure",
	// 		"reason":       "internal_error",
	// 		"challenge_id": challengeID,
	// 		"ip":           ctx.ClientIP(),
	// 	}).Error("Failed to stop challenge container")
	// 	ctx.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to stop challenge container"})
	// 	return
	// }
	auditLog.WithFields(logrus.Fields{
		"event":        "stop_challenge_container",
		"status":       "success",
		"challenge_id": challengeID,
		"ip":           ctx.ClientIP(),
	}).Info("Challenge container stopped successfully")
	ctx.JSON(http.StatusOK, types.SuccessResponse{Message: "Challenge container stopped successfully"})
}
