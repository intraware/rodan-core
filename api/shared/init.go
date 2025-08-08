package shared

import (
	"time"

	"github.com/AnimeKaizoku/cacher"
	"github.com/intraware/rodan/internal/config"
	"github.com/intraware/rodan/internal/models"
)

func Init(config *config.Config) {
	UserBlackList = config.App.Leaderboard.UserBlackList
	TeamBlackList = config.App.Leaderboard.TeamBlackList

	ResetPasswordCache = cacher.NewCacher[string, models.User](&cacher.NewCacherOpts{
		TimeToLive:    time.Duration(config.App.TokenExpiry) * time.Minute,
		CleanInterval: time.Hour * 2,
		CleanerMode:   cacher.CleaningCentral,
	})
}
