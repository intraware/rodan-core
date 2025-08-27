package shared

import (
	"time"

	"github.com/intraware/rodan/internal/cache"
	"github.com/intraware/rodan/internal/config"
	"github.com/intraware/rodan/internal/models"
)

func Init(config *config.Config) {
	UserBlackList = config.App.Leaderboard.UserBlackList
	TeamBlackList = config.App.Leaderboard.TeamBlackList

	ResetPasswordCache = cache.NewCache[string, models.User](&cache.CacheOpts{
		TimeToLive:    time.Duration(config.App.TokenExpiry) * time.Minute,
		CleanInterval: ptr(time.Hour * 2),
		Revaluate:     ptr(false),
	})
}
