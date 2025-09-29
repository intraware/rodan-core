package models

import "gorm.io/gorm"

type Container struct {
	gorm.Model
	TeamID      uint     `json:"team_id" gorm:"index"`
	ChallengeID uint     `json:"challenge_id" gorm:"index"`
	ContainerID string   `json:"container_id" gorm:"unique"` // Docker container ID
	Flag        string   `json:"flag"`
	Ports       []int    `json:"ports" gorm:"type:integer[]"`
	Links       []string `json:"links" gorm:"type:text[]"`
}
