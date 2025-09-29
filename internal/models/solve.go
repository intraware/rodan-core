package models

import "gorm.io/gorm"

type Solve struct {
	gorm.Model
	TeamID        uint `json:"team_id" gorm:"index"`
	ChallengeID   uint `json:"challenge_id" gorm:"index"`
	UserID        uint `json:"user_id" gorm:"index"`
	ChallengeType int8 `json:"challenge_type"`
	BloodCount    uint `json:"blood_type"`
}
