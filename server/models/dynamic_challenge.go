package models

import (
	"gorm.io/gorm"
)

type DynamicChallenge struct {
	ID 	  int    `json:"id" gorm:"unique;index"`
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Category int8 `json:"category"`
	PointsMin int `json:"points_min"`
	PointsMax int `json:"points_max"`
	Difficulty int8 `json:"difficulty"`
}