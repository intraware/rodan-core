package models

import "gorm.io/gorm"

type Solve struct {
	gorm.Model
	TeamID        uint `json:"team_id" gorm:"column:team_id;index"`
	ChallengeID   uint `json:"challenge_id" gorm:"column:challenge_id;index"`
	UserID        uint `json:"user_id" gorm:"column:user_id;index"`
	ChallengeType int8 `json:"challenge_type" gorm:"column:challenge_type"`
	BloodCount    uint `json:"blood_type" gorm:"column:blood_count"`
}
