package models

// import (
// 	"gorm.io/gorm"
// )

type StaticChallenge struct {
	ID         int      `json:"id" gorm:"unique;index"`
	Name       string   `json:"name"`
	Desc       string   `json:"desc"`
	Category   int8     `json:"category"`
	Flag       string   `json:"flag"`
	Ports      []int    `json:"ports"`
	Links      []string `json:"links"`
	PointsMin  int      `json:"points_min"`
	PointsMax  int      `json:"points_max"`
	Difficulty int8     `json:"difficulty"`
}

