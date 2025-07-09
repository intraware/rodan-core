package auth

type signUpRequest struct {
	Username       string `json:"username" binding:"required" example:"intraware"`
	Email          string `json:"email" binding:"required,email" example:"example@intraware.org"`
	Password       string `json:"password" binding:"required,min=8" example:"mystrongpassword"`
	GitHubUsername string `json:"github_username" binding:"required" example:"intraware"`
}

type loginRequest struct {
	Username string `json:"username" binding:"required" example:"intraware"`
	Password string `json:"password" binding:"required" example:"mystrongpassword"`
}

type authResponse struct {
	Token string   `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  userInfo `json:"user"`
}

type userInfo struct {
	ID             int    `json:"id" example:"42"`
	Username       string `json:"username" example:"intraware"`
	Email          string `json:"email" example:"example@intraware.org"`
	GitHubUsername string `json:"github_username" example:"intraware"`
	TeamID         *int   `json:"team_id" example:"1"`
}

type errorResponse struct {
	Error string `json:"error" example:"Something went wrong"`
}

type resetPasswordRequest struct {
	Password string `json:"password" example:"MyNewStrongPassword" binding:"required,min=8"`
}

type forgotPasswordRequest struct {
	Username   string  `json:"username" example:"intraware" binding:"required"`
	OTP        *string `json:"otp" example:"11223344"`
	BackupCode *string `json:"backup_code" example:"123456789012"`
}

type resetTokenResponse struct {
	ResetToken string `json:"reset_token" example:"abc123def456..."`
}
