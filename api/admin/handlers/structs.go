package handlers

import "github.com/intraware/rodan/internal/models"

// swagger:model
type UserResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	Active    bool   `json:"active"`
	Ban       bool   `json:"ban"`
	Blacklist bool   `json:"blacklist"`
	TeamID    *uint  `json:"team_id,omitempty"`
}

// swagger:model
type TeamResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Ban       bool   `json:"ban"`
	Blacklist bool   `json:"blacklist"`
	LeaderID  uint   `json:"leader"`
}

// swagger:model
type ChallengeResponse struct {
	ID            uint                  `json:"id"`
	Name          string                `json:"name"`
	Author        string                `json:"author"`
	Desc          string                `json:"desc"`
	Category      int8                  `json:"category"`
	PointsMin     int                   `json:"points_min"`
	PointsMax     int                   `json:"points_max"`
	Difficulty    int8                  `json:"difficulty"`
	IsStatic      bool                  `json:"is_static"`
	IsVisible     bool                  `json:"is_visible"`
	StaticConfig  *models.StaticConfig  `json:"static_config,omitempty"`
	DynamicConfig *models.DynamicConfig `json:"dynamic_config,omitempty"`
	Hints         []models.Hint         `json:"hints,omitempty"`
}
