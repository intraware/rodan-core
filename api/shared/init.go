package shared

import (
	"github.com/intraware/rodan/internal/config"
)

func Init(config *config.Config) {
	UserBlackList = config.App.Leaderboard.UserBlackList
	TeamBlackList = config.App.Leaderboard.TeamBlackList
}
