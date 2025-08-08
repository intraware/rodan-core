package shared

import (
	"time"

	"github.com/AnimeKaizoku/cacher"
	"github.com/intraware/rodan/internal/models"
)

var UserCache = cacher.NewCacher[int, models.User](&cacher.NewCacherOpts{
	TimeToLive:    time.Minute * 3,
	CleanInterval: time.Hour * 2,
	CleanerMode:   cacher.CleaningCentral,
	Revaluate:     true,
})

var TeamCache = cacher.NewCacher[int, models.Team](&cacher.NewCacherOpts{
	TimeToLive:    time.Minute * 3,
	CleanInterval: time.Hour * 2,
	CleanerMode:   cacher.CleaningCentral,
	Revaluate:     true,
})

var LoginCache = cacher.NewCacher[string, models.User](&cacher.NewCacherOpts{
	TimeToLive:    time.Minute * 2,
	CleanInterval: time.Hour * 2,
	CleanerMode:   cacher.CleaningCentral,
	Revaluate:     true,
})

var ResetPasswordCache *cacher.Cacher[string, models.User]

var ChallengeCache = cacher.NewCacher[int, models.Challenge](&cacher.NewCacherOpts{
	TimeToLive:    time.Minute * 3,
	CleanInterval: time.Hour * 2,
	CleanerMode:   cacher.CleaningCentral,
	Revaluate:     true,
})

var TeamSolvedCache = cacher.NewCacher[string, bool](nil)

var StaticConfig = cacher.NewCacher[int, models.StaticConfig](&cacher.NewCacherOpts{
	TimeToLive:    time.Minute * 3,
	CleanInterval: time.Hour * 2,
	CleanerMode:   cacher.CleaningCentral,
	Revaluate:     true,
})
