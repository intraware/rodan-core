package models

// import (
// 	"gorm.io/gorm"
// )

type Container struct {
	TeamID      int      `json:"team_id" gorm:"index"`
	ChallengeID int      `json:"challenge_id" gorm:"index"`
	ContainerID string   `json:"container_id" gorm:"unique"` // Docker container ID
	Flag        string   `json:"flag"`
	Ports       []int    `json:"ports"`
	Links       []string `json:"links"`
}
