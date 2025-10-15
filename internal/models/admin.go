package models

import "gorm.io/gorm"

type Admin struct {
	gorm.Model
	Username  string `json:"username"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"password"`
	Moderator bool   `json:"moderator"`
	Active    bool   `json:"active"`
}
