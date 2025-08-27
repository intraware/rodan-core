package shared

import (
	"time"

	"github.com/intraware/rodan/internal/cache"
	"github.com/intraware/rodan/internal/models"
)

var UserCache = cache.NewCache[int, models.User](&cache.CacheOpts{
	TimeToLive:    3 * time.Minute,
	CleanInterval: ptr(time.Hour * 2),
	Revaluate:     ptr(true),
})

var TeamCache = cache.NewCache[int, models.Team](&cache.CacheOpts{
	TimeToLive:    3 * time.Minute,
	CleanInterval: ptr(time.Hour * 2),
	Revaluate:     ptr(true),
})

var LoginCache = cache.NewCache[string, models.User](&cache.CacheOpts{
	TimeToLive:    2 * time.Minute,
	CleanInterval: ptr(time.Hour * 2),
	Revaluate:     ptr(true),
})

var ResetPasswordCache cache.Cache[string, models.User]

var ChallengeCache = cache.NewCache[int, models.Challenge](&cache.CacheOpts{
	TimeToLive:    3 * time.Minute,
	CleanInterval: ptr(time.Hour * 2),
	Revaluate:     ptr(true),
})

var TeamSolvedCache = cache.NewCache[string, bool](&cache.CacheOpts{
	TimeToLive:    0,
	CleanInterval: ptr(time.Hour * 2),
	Revaluate:     ptr(false),
})

var StaticConfig = cache.NewCache[int, models.StaticConfig](&cache.CacheOpts{
	TimeToLive:    3 * time.Minute,
	CleanInterval: ptr(time.Hour * 2),
	Revaluate:     ptr(true),
})

var BanHistoryCache = cache.NewCache[string, models.BanHistory](&cache.CacheOpts{
	TimeToLive:    10 * time.Minute,
	CleanInterval: ptr(time.Hour * 2),
	Revaluate:     ptr(true),
})

func ptr[T any](v T) *T { return &v }
