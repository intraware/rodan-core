package models

import (
	"gorm.io/gorm"
)

type Team struct {
	ID        int    `json:"id" gorm:"unique;index"`
	Name      string `json:"name"`
	Code	  string `json:"code" gorm:"unique"`
	Ban bool `json:"ban" gorm:"default:false"`
} 