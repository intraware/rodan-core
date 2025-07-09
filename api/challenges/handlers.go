package challenges

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/config"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils"
	"gorm.io/gorm"
)

func getChallengeList(ctx *gin.Context) {
	var challenges []models.Challenge

	// Get all challenges with just id and name
	if err := models.DB.Select("id, name").Find(&challenges).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch challenges"})
		return
	}

	// Convert to response format
	var challengeList []ChallengeListItem
	for _, challenge := range challenges {
		challengeList = append(challengeList, ChallengeListItem{
			ID:    challenge.ID,
			Title: challenge.Name,
		})
	}

	ctx.JSON(http.StatusOK, challengeList)
}

func getChallengeDetail(ctx *gin.Context) {
	challengeIDStr := ctx.Param("id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid challenge ID"})
		return
	}

	var challenge models.Challenge
	if err := models.DB.First(&challenge, challengeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Challenge not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Prepare response based on challenge type
	response := ChallengeDetail{
		ID:         challenge.ID,
		Name:       challenge.Name,
		Desc:       challenge.Desc,
		Category:   challenge.Category,
		Difficulty: challenge.Difficulty,
		PointsMin:  challenge.PointsMin,
		PointsMax:  challenge.PointsMax,
	}

	// Add links only for static challenges
	if challenge.IsStatic {
		response.Links = challenge.Links
	} else {
		response.Links = []string{} // Dynamic challenges don't have static links
	}

	ctx.JSON(http.StatusOK, response)
}

func submitFlag(ctx *gin.Context) {
	challengeIDStr := ctx.Param("id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid challenge ID"})
		return
	}

	var req SubmitFlagRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID := ctx.GetInt("user_id")

	// Get user's team
	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	if user.TeamID == nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "User must be in a team to submit flags"})
		return
	}

	// Get challenge details
	var challenge models.Challenge
	if err := models.DB.First(&challenge, challengeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Challenge not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	teamID := *user.TeamID

	// Check if team has already solved this challenge
	var existingSolve models.Solve
	err = models.DB.Where("team_id = ? AND challenge_id = ?", teamID, challengeID).First(&existingSolve).Error
	if err == nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Challenge already solved by your team"})
		return
	}

	var correctFlag string
	var challengeType int8

	if challenge.IsStatic {
		// For static challenges, use the flag stored in the challenge
		correctFlag = challenge.Flag
		challengeType = 0 // Static challenge type
	} else {
		// For dynamic challenges, get the generated flag from container
		var container models.Container
		err = models.DB.Where("team_id = ? AND challenge_id = ?", teamID, challengeID).First(&container).Error
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "No running container found for this challenge. Please start the challenge first"})
			return
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
			return
		}
		correctFlag = container.Flag
		challengeType = 1 // Dynamic challenge type
	}

	// Check if submitted flag is correct
	if req.Flag != correctFlag {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Incorrect flag"})
		return
	}

	// Record the solve
	solve := models.Solve{
		TeamID:        teamID,
		ChallengeID:   challengeID,
		UserID:        userID,
		Time:          time.Now().Unix(),
		ChallengeType: challengeType,
	}

	if err := models.DB.Create(&solve).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to record solve"})
		return
	}

	// Calculate points based on challenge difficulty and timing
	points := calculatePoints(challenge.PointsMin, challenge.PointsMax, solve.Time)

	ctx.JSON(http.StatusOK, SubmitFlagResponse{
		Correct: true,
		Points:  points,
		Message: "Congratulations! Flag accepted.",
	})
}

func calculatePoints(minPoints, maxPoints int, solveTime int64) int {
	// Simple point calculation - could be made more sophisticated
	// For now, just return max points for correct submissions
	// TODO: Implement decay based on solve time, number of solves, etc.
	return maxPoints
}

func startDynamicChallenge(ctx *gin.Context) {
	challengeIDStr := ctx.Param("id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid challenge ID"})
		return
	}

	userID := ctx.GetInt("user_id")

	// Get user's team
	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	if user.TeamID == nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "User must be in a team to start challenges"})
		return
	}

	// Get challenge details
	var challenge models.Challenge
	if err := models.DB.First(&challenge, challengeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Challenge not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Check if challenge is dynamic
	if challenge.IsStatic {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "This endpoint is only for dynamic challenges"})
		return
	}

	// Validate challenge has required fields for dynamic containers
	if challenge.DockerImage == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Challenge configuration incomplete - missing Docker image"})
		return
	}

	teamID := *user.TeamID

	// Check if container already exists for this team and challenge
	var existingContainer models.Container
	err = models.DB.Where("team_id = ? AND challenge_id = ?", teamID, challengeID).First(&existingContainer).Error
	if err == nil {
		// Container already exists, return existing details
		response := ContainerResponse{
			Flag:  existingContainer.Flag,
			Ports: existingContainer.Ports,
			Links: existingContainer.Links,
		}
		ctx.JSON(http.StatusOK, response)
		return
	}

	// Generate hashed flag
	flag := generateHashedFlag(challengeID, teamID)

	// Initialize Docker service
	dockerService, err := utils.NewDockerService()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to initialize Docker service"})
		return
	}
	defer dockerService.Close()

	// Use exposed ports from challenge or default
	exposedPorts := challenge.ExposedPorts
	if len(exposedPorts) == 0 {
		exposedPorts = []string{"80/tcp"} // Default to port 80
	}

	// Create and start container
	containerInfo, err := dockerService.CreateContainer(challengeID, teamID, challenge.DockerImage, exposedPorts)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to create container: %v", err)})
		return
	}

	// Create container record
	container := models.Container{
		TeamID:      teamID,
		ChallengeID: challengeID,
		ContainerID: containerInfo.ID,
		Flag:        flag,
		Ports:       containerInfo.Ports,
		Links:       containerInfo.Links,
	}

	if err := models.DB.Create(&container).Error; err != nil {
		// Clean up the Docker container if database insert fails
		dockerService.StopContainer(containerInfo.ID)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create container record"})
		return
	}

	response := ContainerResponse{
		Flag:  container.Flag,
		Ports: container.Ports,
		Links: container.Links,
	}

	ctx.JSON(http.StatusCreated, response)
}

func stopDynamicChallenge(ctx *gin.Context) {
	challengeIDStr := ctx.Param("id")
	challengeID, err := strconv.Atoi(challengeIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid challenge ID"})
		return
	}

	userID := ctx.GetInt("user_id")

	// Get user's team
	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	if user.TeamID == nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "User must be in a team to stop challenges"})
		return
	}

	// Get challenge details
	var challenge models.Challenge
	if err := models.DB.First(&challenge, challengeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Challenge not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Check if challenge is dynamic
	if challenge.IsStatic {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "This endpoint is only for dynamic challenges"})
		return
	}

	teamID := *user.TeamID

	// Check if container exists for this team and challenge
	var container models.Container
	err = models.DB.Where("team_id = ? AND challenge_id = ?", teamID, challengeID).First(&container).Error
	if err == gorm.ErrRecordNotFound {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "No running container found for this challenge"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Initialize Docker service
	dockerService, err := utils.NewDockerService()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to initialize Docker service"})
		return
	}
	defer dockerService.Close()

	// Stop and remove the Docker container
	if err := dockerService.StopContainer(container.ContainerID); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Failed to stop container: %v", err)})
		return
	}

	// Remove container record from database
	if err := models.DB.Delete(&container).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to remove container record"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Container stopped successfully"})
}

func generateHashedFlag(challengeID, teamID int) string {
	cfg, _ := config.LoadConfig("./config.toml")

	// Create the input string: FlagSecret + TeamID + ChallengeID
	input := fmt.Sprintf("%s%d%d", cfg.FlagSecret, teamID, challengeID)

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(input))
	hashHex := hex.EncodeToString(hash[:])

	// Format as flag
	return fmt.Sprintf("flag{%s}", hashHex[:32]) // Use first 32 characters of hash
}
