package api

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

type ChallengeListItem struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type ChallengeDetail struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Desc       string   `json:"desc"`
	Category   int8     `json:"category"`
	Difficulty int8     `json:"difficulty"`
	PointsMin  int      `json:"points_min"`
	PointsMax  int      `json:"points_max"`
	Links      []string `json:"links,omitempty"`
}

type ContainerResponse struct {
	Flag  string   `json:"flag"`
	Ports []int    `json:"ports"`
	Links []string `json:"links"`
}

type SubmitFlagRequest struct {
	Flag string `json:"flag" binding:"required"`
}

type SubmitFlagResponse struct {
	Correct bool   `json:"correct"`
	Points  int    `json:"points"`
	Message string `json:"message"`
}
