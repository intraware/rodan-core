package shared

import (
	"github.com/intraware/rodan/internal/cache"
	"github.com/intraware/rodan/internal/models"
)

var UserCache cache.Cache[uint, models.User]
var TeamCache cache.Cache[uint, models.Team]
var ChallengeCache cache.Cache[uint, models.Challenge]
var TeamSolvedCache cache.Cache[string, bool]
var StaticConfig cache.Cache[uint, models.StaticConfig]
var BanHistoryCache cache.Cache[string, models.BanHistory]
