package models

import "gorm.io/gorm"

type Challenge struct {
	gorm.Model
	ID         int    `json:"id" gorm:"unique;index"`
	Name       string `json:"name"`
	Author     string `json:"author"`
	Desc       string `json:"desc"`
	Category   int8   `json:"category"`
	PointsMin  int    `json:"points_min"`
	PointsMax  int    `json:"points_max"`
	Difficulty int8   `json:"difficulty"`
	IsStatic   bool   `json:"is_static"`
	IsVisible  bool   `json:"is_visible"`

	StaticConfig  *StaticConfig  `gorm:"foreignKey:ChallengeID;constraint:OnDelete:CASCADE"`
	DynamicConfig *DynamicConfig `gorm:"foreignKey:ChallengeID;constraint:OnDelete:CASCADE"`
	Hints         []Hint         `json:"hints" gorm:"foreignKey:ChallengeID"`
}

type StaticConfig struct {
	ChallengeID int      `json:"challenge_id"`
	Flag        string   `json:"flag,omitempty"`
	Ports       []int    `json:"ports,omitempty" gorm:"type:integer[]"`
	Links       []string `json:"links,omitempty" gorm:"type:text[]"`
}

type DynamicConfig struct {
	gorm.Model
	ChallengeID  int      `json:"challenge_id"`
	DockerImage  string   `json:"docker_image,omitempty"`
	ExposedPorts []string `json:"exposed_ports,omitempty" gorm:"type:text[]"`
	TTL          int64    `json:"ttl"`
	Reusable     bool     `json:"reusable"`
}

type Hint struct {
	gorm.Model
	Context     string `json:"context"`
	Points      int    `json:"points"`
	ChallengeID int    `json:"challenge_id"`
}
