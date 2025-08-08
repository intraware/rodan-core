package models

import "gorm.io/gorm"

type Admin struct {
	gorm.Model
	ID           int `json:"id" gorm:"unique;index;primaryKey"`
	Name      string `json:"name"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"password"`
	Moderator   bool `json:"moderator"`
	Active      bool `json:"active"`
}

