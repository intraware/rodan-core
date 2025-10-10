package shared

import (
	"time"

	"github.com/intraware/rodan/internal/cache"
	"github.com/intraware/rodan/internal/config"
	"github.com/intraware/rodan/internal/models"
)

func ptr[T any](v T) *T { return &v }

func Init(config *config.Config) {
	UserBlackList = config.App.Leaderboard.UserBlackList
	TeamBlackList = config.App.Leaderboard.TeamBlackList

	UserCache = cache.NewCache[uint, models.User](&cache.CacheOpts{
		TimeToLive:    3 * time.Minute,
		CleanInterval: ptr(time.Hour * 2),
		Revaluate:     ptr(true),
		Prefix:        "user-cache",
	})
	TeamCache = cache.NewCache[uint, models.Team](&cache.CacheOpts{
		TimeToLive:    3 * time.Minute,
		CleanInterval: ptr(time.Hour * 2),
		Revaluate:     ptr(true),
		Prefix:        "team-cache",
	})
	ChallengeCache = cache.NewCache[uint, models.Challenge](&cache.CacheOpts{
		TimeToLive:    3 * time.Minute,
		CleanInterval: ptr(time.Hour * 2),
		Revaluate:     ptr(true),
		Prefix:        "challenge-cache",
	})
	TeamSolvedCache = cache.NewCache[string, bool](&cache.CacheOpts{
		TimeToLive:    0,
		CleanInterval: ptr(time.Hour * 2),
		Revaluate:     ptr(false),
		Prefix:        "team-solved-cache",
	})
	StaticConfig = cache.NewCache[uint, models.StaticConfig](&cache.CacheOpts{
		TimeToLive:    3 * time.Minute,
		CleanInterval: ptr(time.Hour * 2),
		Revaluate:     ptr(true),
		Prefix:        "static-config-cache",
	})
	BanHistoryCache = cache.NewCache[string, models.BanHistory](&cache.CacheOpts{
		TimeToLive:    10 * time.Minute,
		CleanInterval: ptr(time.Hour * 2),
		Revaluate:     ptr(true),
		Prefix:        "ban-history-cache",
	})
}

func init() {
	allowSubmissions.Store(true)
}
