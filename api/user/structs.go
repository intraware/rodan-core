package user

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

type successResponse struct {
	Message string `json:"message" example:"Something went right"`
}

type updateUserRequest struct {
	Username       *string `json:"username"`
	GitHubUsername *string `json:"github_username"`
}
