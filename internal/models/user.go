package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username  string `json:"username" gorm:"unique"`
	Password  string `json:"-"`
	Email     string `json:"email" gorm:"unique"`
	AvatarURL string `json:"avatar_url" gorm:"column:avatar_url;unique"`
	Active    bool   `json:"active" gorm:"default:false"`
	Ban       bool   `json:"ban" gorm:"default:false"`
	Blacklist bool   `json:"blacklist" gorm:"default:false"`
	TeamID    *uint  `json:"team_id" gorm:"column:team_id"`
	Team      *Team  `json:"team" gorm:"foreignKey:TeamID"`
}

func (User) TableName() string {
	return "users"
}
