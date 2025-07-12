package team

type errorResponse struct {
	Error string `json:"error" example:"Something went wrong"`
}

type successResponse struct {
	Message string `json:"message" example:"Something went right"`
}

type createTeamRequest struct {
	Name string `json:"name" binding:"required" example:"Avengers"`
}

type joinTeamRequest struct {
	Code string `json:"code" binding:"required" example:"ABC123"`
}

type teamResponse struct {
	ID       int        `json:"id" example:"1"`
	Name     string     `json:"name" example:"Avengers"`
	Code     string     `json:"code" example:"ABC123"`
	LeaderID int        `json:"leader_id" example:"42"`
	Members  []userInfo `json:"members"`
}

type userInfo struct {
	ID             int    `json:"id" example:"42"`
	Username       string `json:"username" example:"intraware"`
	Email          string `json:"email" example:"example@intraware.org"`
	GitHubUsername string `json:"github_username" example:"intraware"`
	TeamID         *int   `json:"team_id" example:"1"`
}

type editTeamReq struct {
	Name           *string `json:"name" example:"New Avengers"`
	LeaderUsername *string `json:"leader_username" example:"newleader"`
}
