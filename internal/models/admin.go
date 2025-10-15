package models

import "gorm.io/gorm"

type Admin struct {
	gorm.Model
	AdminID   uint   `json:"admin_id" gorm:"column:admin_id;index"`
	Username  string `json:"username"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"password"`
	Moderator bool   `json:"moderator"`
	Active    bool   `json:"active"`
}
