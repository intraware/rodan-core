package models

// import (
// 	"gorm.io/gorm"
// )

type Solve struct {
	TeamID   int    `json:"team_id" gorm:"index"`
	ChallengeID int    `json:"challenge_id" gorm:"index"`
	UserID   int    `json:"user_id" gorm:"index"`
	Time    int64  `json:"time"`
	ChallengeType int8   `json:"challenge_type"`
}