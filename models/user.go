package models

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/intraware/rodan/utils/values"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
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
	ID             int    `json:"id" gorm:"primaryKey"`
	Username       string `json:"username" gorm:"unique"`
	Password       string `json:"password"`
	Email          string `json:"email"`
	GitHubUsername string `json:"github_username" gorm:"column:github_username;unique"`
	Ban            bool   `json:"ban" gorm:"default:false"`
	BackupCode     string `gorm:"unique"`
	TOTPSecret     string `gorm:"unique"`
	TeamID         *int   `json:"team_id" gorm:"column:team_id"`
	Team           *Team  `json:"team" gorm:"foreignKey:TeamID"`
}

func (User) TableName() string {
	return "users"
}

func generateResetCode(length int) (string, error) {
	code := ""
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += n.String()
	}
	return code, nil
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	err = u.SetPassword(u.Password)
	if err != nil {
		return
	}
	if code, err := generateResetCode(12); err != nil {
		return err
	} else {
		u.BackupCode = code
	}
	u.CreatedAt = time.Now()
	if key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      values.GetConfig().App.TOTPIssuer,
		AccountName: u.Username,
		Period:      60,
		Digits:      otp.Digits(otp.DigitsEight),
	}); err != nil {
		return err
	} else {
		u.TOTPSecret = key.Secret()
	}
	return
}

func (u *User) SetPassword(password string) (err error) {
	err = nil
	salt := make([]byte, saltLength)
	if _, err = rand.Read(salt); err != nil {
		return
	}
	hash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, uint8(parallelism), keyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encoded := fmt.Sprintf("$argon2id$v=19$t=%d$m=%d$p=%d$%s$%s",
		timeCost, memoryCost, parallelism, b64Salt, b64Hash)
	u.Password = encoded
	return
}

func (u *User) TOTPUrl() (string, error) {
	issuer := values.GetConfig().App.TOTPIssuer
	if key, err := otp.NewKeyFromURL(
		fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
			issuer, u.Email, u.TOTPSecret, issuer),
	); err != nil {
		return "", err
	} else {
		return key.URL(), nil
	}
}

func (u *User) VerifyTOTP(otp string) bool {
	return totp.Validate(otp, u.TOTPSecret)
}

func (u *User) ComparePassword(password string) (bool, error) {
	parts := strings.Split(u.Password, "$")
	if len(parts) != 7 {
		return false, fmt.Errorf("invalid hash format")
	}
	var t, m uint32
	var p uint8
	_, err := fmt.Sscanf(parts[3], "t=%d", &t)
	if err != nil {
		return false, fmt.Errorf("error parsing time: %w", err)
	}
	_, err = fmt.Sscanf(parts[4], "m=%d", &m)
	if err != nil {
		return false, fmt.Errorf("error parsing memory: %w", err)
	}
	_, err = fmt.Sscanf(parts[5], "p=%d", &p)
	if err != nil {
		return false, fmt.Errorf("error parsing parallelism: %w", err)
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[6])
	if err != nil {
		return false, fmt.Errorf("error decoding salt: %w", err)
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[7])
	if err != nil {
		return false, fmt.Errorf("error decoding hash: %w", err)
	}
	actualHash := argon2.IDKey([]byte(password), salt, t, m, p, uint32(len(expectedHash)))
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
