package models

import (
	"gorm.io/gorm"
)

type User struct {
	ID        int    `json:"id" gorm:"unique;index"`
	Username  string `json:"username" gorm:"unique"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	GitHubUsernsme string `json:"github_username"`
	Ban bool `json:"ban" gorm:"default:false"`
}