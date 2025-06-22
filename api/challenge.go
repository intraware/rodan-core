package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/models"
	"gorm.io/gorm"
)

func LoadChallenges(r *gin.RouterGroup) {
	challengeRouter := r.Group("/challenge")
	challengeRouter.GET("/list", getChallengeList)
	challengeRouter.GET("/:id", getChallengeDetail)
	challengeRouter.POST("/:id/submit", func(ctx *gin.Context) {})
}

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
