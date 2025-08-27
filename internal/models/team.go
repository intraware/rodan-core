package models

import (
	"crypto/rand"
	"encoding/hex"

	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	ID       int    `json:"id" gorm:"primaryKey"`
	Name     string `json:"name"`
	Code     string `json:"code" gorm:"unique"`
	Ban      bool   `json:"ban" gorm:"default:false"`
	Blacklist bool   `json:"blacklist" gorm:"default:false"`
	LeaderID int    `json:"leader" gorm:"not null"`
	Leader   User   `gorm:"foreignKey:LeaderID;constraint:OnUpdate:CASCADE"`
	Members  []User `gorm:"foreignKey:TeamID"`
}

func (Team) TableName() string {
	return "teams"
}

func (t *Team) BeforeCreate(tx *gorm.DB) (err error) {
	random := make([]byte, 6) // this much size to avoid collisions
	_, err = rand.Read(random)
	if err != nil {
		return err
	}
	t.Code = hex.EncodeToString(random)
	return nil
}

func (t *Team) BeforeDelete(tx *gorm.DB) (err error) {
	err = tx.Model(&User{}).Where("team_id = ?", t.ID).Update("team_id", nil).Error
	return
}
