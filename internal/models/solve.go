package models

import "gorm.io/gorm"

type Solve struct {
	gorm.Model
	TeamID        int  `json:"team_id" gorm:"index"`
	ChallengeID   int  `json:"challenge_id" gorm:"index"`
	UserID        int  `json:"user_id" gorm:"index"`
	ChallengeType int8 `json:"challenge_type"`
}
