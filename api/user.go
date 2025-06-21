package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/intraware/rodan/config"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils"
	"github.com/intraware/rodan/utils/middleware"
	"gorm.io/gorm"
)

func LoadUser(r *gin.RouterGroup) {
	userRouter := r.Group("/user")
	userRouter.POST("/signup", signUp)
	userRouter.POST("/login", login)
	userRouter.POST("/forgot-password", forgotPassword)

	// Protected routes
	protected := userRouter.Group("")
	protected.Use(middleware.AuthRequired())
	protected.GET("/me", getMyProfile)

	// Public routes
	userRouter.GET("/:id", getUserProfile)
}

func signUp(ctx *gin.Context) {
	var req SignUpRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := models.DB.Where("username = ? OR email = ? OR github_username = ?",
		req.Username, req.Email, req.GitHubUsername).First(&existingUser).Error; err == nil {
		ctx.JSON(http.StatusConflict, ErrorResponse{Error: "User with this username, email, or GitHub username already exists"})
		return
	}

	// Create new user
	user := models.User{
		Username:       req.Username,
		Email:          req.Email,
		Password:       req.Password, // Will be hashed in BeforeCreate hook
		GitHubUsername: req.GitHubUsername,
	}

	if err := models.DB.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user"})
		return
	}

	// Generate JWT token
	cfg, _ := config.LoadConfig("./config.toml")
	token, err := utils.GenerateJWT(user.ID, user.Username, cfg.JwtSecret)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
		return
	}

	// Prepare response
	userInfo := UserInfo{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		GitHubUsername: user.GitHubUsername,
		TeamID:         user.TeamID,
	}

	ctx.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  userInfo,
	})
}

func login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Find user by username
	var user models.User
	if err := models.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid username or password"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Check if user is banned
	if user.Ban {
		ctx.JSON(http.StatusForbidden, ErrorResponse{Error: "Account is banned"})
		return
	}

	// Verify password
	isValid, err := user.ComparePassword(req.Password)
	if err != nil || !isValid {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid username or password"})
		return
	}

	// Generate JWT token
	cfg, _ := config.LoadConfig("./config.toml")
	token, err := utils.GenerateJWT(user.ID, user.Username, cfg.JwtSecret)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
		return
	}

	// Prepare response
	userInfo := UserInfo{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		GitHubUsername: user.GitHubUsername,
		TeamID:         user.TeamID,
	}

	ctx.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  userInfo,
	})
}

func forgotPassword(ctx *gin.Context) {
	// TODO: Implement forgot password functionality
	ctx.JSON(http.StatusNotImplemented, ErrorResponse{Error: "Forgot password functionality not yet implemented"})
}

func getMyProfile(ctx *gin.Context) {
	userID := ctx.GetInt("user_id")

	var user models.User
	if err := models.DB.Preload("Team").First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
		return
	}

	userInfo := UserInfo{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		GitHubUsername: user.GitHubUsername,
		TeamID:         user.TeamID,
	}

	ctx.JSON(http.StatusOK, userInfo)
}

func getUserProfile(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "User not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Database error"})
		return
	}

	// Return public profile (excluding sensitive information like email)
	userInfo := UserInfo{
		ID:             user.ID,
		Username:       user.Username,
		Email:          "", // Don't expose email in public profile
		GitHubUsername: user.GitHubUsername,
		TeamID:         user.TeamID,
	}

	ctx.JSON(http.StatusOK, userInfo)
}
