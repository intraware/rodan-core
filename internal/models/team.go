package models

import (
	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	Name      string `json:"name"`
	Code      string `json:"code" gorm:"unique"`
	Ban       bool   `json:"ban" gorm:"default:false"`
	Blacklist bool   `json:"blacklist" gorm:"default:false"`
	LeaderID  uint   `json:"leader" gorm:"not null"`
	Leader    User   `gorm:"foreignKey:LeaderID;constraint:OnUpdate:CASCADE"`
	Members   []User `gorm:"foreignKey:TeamID"`
}

func (Team) TableName() string {
	return "teams"
}
