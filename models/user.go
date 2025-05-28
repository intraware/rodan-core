package models

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

const (
	timeCost    = 1
	memoryCost  = 32 * 1024
	parallelism = 2
	saltLength  = 16
	keyLength   = 32
)

type User struct {
	gorm.Model
	ID             int       `json:"id" gorm:"primaryKey"`
	Username       string    `json:"username" gorm:"unique"`
	Password       string    `json:"password"`
	Email          string    `json:"email"`
	GitHubUsername string    `json:"github_username" gorm:"column:github_username;unique"`
	Ban            bool      `json:"ban" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at"`
	TeamID         *int      `json:"team_id" gorm:"column:team_id"`
	Team           *Team     `json:"team" gorm:"foreignKey:TeamID"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	salt := make([]byte, saltLength)
	if _, err = rand.Read(salt); err != nil {
		return err
	}
	hash := argon2.IDKey([]byte(u.Password), salt, timeCost, memoryCost, uint8(parallelism), keyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encoded := fmt.Sprintf("$argon2id$v=19$t=%d$m=%d$p=%d%s%s", timeCost, memoryCost, parallelism, b64Salt, b64Hash)
	u.Password = encoded
	u.CreatedAt = time.Now()
	return nil
}

func (u *User) ComparePassword(password string) (bool, error) {
	parts := strings.Split(u.Password, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}
	var t, m, p uint32
	_, err := fmt.Sscanf(parts[3], "t=%d", &t)
	if err != nil {
		return false, err
	}
	_, err = fmt.Sscanf(parts[3], "m=%d", &m)
	if err != nil {
		return false, err
	}
	_, err = fmt.Sscanf(parts[3], "p=%d", &p)
	if err != nil {
		return false, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	actualHash := argon2.IDKey([]byte(password), salt, t, m, uint8(p), uint32(len(expectedHash)))
	if bytes.Equal(actualHash, expectedHash) {
		return true, nil
	}
	return false, nil
}

func (u *User) BeforeDelete(tx *gorm.DB) (err error) {
	var team Team
	err = tx.First(&team, u.TeamID).Error
	if err != nil {
		return
	}
	if team.LeaderID == u.ID {
		err = tx.Exec(`
    UPDATE teams 
    SET leader_id = (
        SELECT id FROM users 
        WHERE team_id = ? AND id != ? AND deleted_at IS NULL
        ORDER BY created_at LIMIT 1
    ) 
    WHERE id = ?`,
			team.ID, u.ID, team.ID).Error
	}
	return
}
