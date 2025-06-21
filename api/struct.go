package api

type SignUpRequest struct {
	Username       string `json:"username" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6"`
	GitHubUsername string `json:"github_username" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}

type UserInfo struct {
	ID             int    `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	GitHubUsername string `json:"github_username"`
	TeamID         *int   `json:"team_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateTeamRequest struct {
	Name string `json:"name" binding:"required"`
}

type JoinTeamRequest struct {
	Code string `json:"code" binding:"required"`
}

type TeamResponse struct {
	ID       int        `json:"id"`
	Name     string     `json:"name"`
	Code     string     `json:"code"`
	LeaderID int        `json:"leader_id"`
	Members  []UserInfo `json:"members"`
}
