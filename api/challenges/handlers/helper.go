package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils/values"
)

const (
	initialBackoff = time.Minute
	maxBackoff     = 1 * time.Hour
)

var (
	userCount     atomic.Int64
	userUpdatedAt atomic.Int64
	userBackoff   atomic.Int64
)

type challengeSolveStat struct {
	SolveCount     atomic.Int64
	LastUpdateUnix atomic.Int64
	Backoff        atomic.Int64
}

var challengeStats sync.Map

func getUserCount() int {
	now := time.Now()
	userBlackList := values.GetConfig().App.Leaderboard.UserBlackList
	teamBlackList := values.GetConfig().App.Leaderboard.TeamBlackList
	last := time.Unix(0, userUpdatedAt.Load())
	backoff := time.Duration(userBackoff.Load())
	if backoff == 0 {
		backoff = initialBackoff
	}
	if time.Since(last) < backoff {
		return int(userCount.Load())
	}
	var count int64
	err := models.DB.Where("id NOT IN ? AND team_id NOT IN ?", userBlackList, teamBlackList).Model(&models.User{}).Count(&count).Error
	if err != nil {
		return int(userCount.Load())
	}

	current := userCount.Load()
	if int64(count) == current {
		newBackoff := backoff * 2
		newBackoff = max(maxBackoff, newBackoff)
		userBackoff.Store(int64(newBackoff))
	} else {
		userCount.Store(int64(count))
		userBackoff.Store(int64(initialBackoff))
	}
	userUpdatedAt.Store(now.UnixNano())
	return int(count)
}

func getSolveCount(challengeID int) int {
	now := time.Now()
	val, _ := challengeStats.LoadOrStore(challengeID, &challengeSolveStat{})
	stat := val.(*challengeSolveStat)
	userBlackList := values.GetConfig().App.Leaderboard.UserBlackList
	teamBlackList := values.GetConfig().App.Leaderboard.TeamBlackList

	last := time.Unix(0, stat.LastUpdateUnix.Load())
	backoff := time.Duration(stat.Backoff.Load())
	if backoff == 0 {
		backoff = initialBackoff
	}
	if time.Since(last) < backoff {
		return int(stat.SolveCount.Load())
	}
	var count int64
	err := models.DB.Model(&models.Solve{}).
		Where("challenge_id = ? AND user_id NOT IN ? AND team_id NOT IN ?", challengeID, userBlackList, teamBlackList).
		Count(&count).Error
	if err != nil {
		return int(stat.SolveCount.Load())
	}
	current := stat.SolveCount.Load()
	if int64(count) == current {
		newBackoff := backoff * 2
		newBackoff = max(maxBackoff, newBackoff)
		stat.Backoff.Store(int64(newBackoff))
	} else {
		stat.SolveCount.Store(int64(count))
		stat.Backoff.Store(int64(initialBackoff))
	}
	stat.LastUpdateUnix.Store(now.UnixNano())
	return int(count)
}

func calcPoints(minPoints, maxPoints, challengeID int) (int, error) {
	solves := getSolveCount(challengeID)
	users := getUserCount()
	score := smoothScore(solves, maxPoints, minPoints, users, 3, 2.25)
	return score, nil
}

//solves	Current number of solves
//maxPoints	Score when 0–offset users solved
//minPoints	Floor score (e.g. 100)
//total	Expected max solves (e.g. 300)
//offset	First N solvers get full score
//power	Controls how sharp the decay is

// First `offset` solvers get full score. After that, score decays smoothly.
func smoothScore(solves int, maxPoints int, minPoints int, total int, offset int, power float64) int {
	if solves <= offset {
		return maxPoints
	}
	if solves > total {
		solves = total
	}
	x := float64(solves-offset) / float64(total-offset)
	score := float64(maxPoints) - float64(maxPoints-minPoints)*math.Pow(x, power)
	score = max(float64(minPoints), score)
	return int(math.Round(score))
}

var dynFlagMap sync.Map

func getDynamicFlag(challengeID, teamID int) string {
	key := fmt.Sprintf("%d:%d", challengeID, teamID)
	if val, ok := dynFlagMap.Load(key); ok {
		return val.(string)
	}
	flag := generateHashedFlag(challengeID, teamID)
	dynFlagMap.Store(key, flag)
	return flag
}

func generateHashedFlag(challengeID, teamID int) string {
	cfg := values.GetConfig()
	input := fmt.Sprintf("%d%s%d", teamID, cfg.Server.Security.FlagSecret, challengeID)
	hash := sha256.Sum256([]byte(input))
	hashHex := hex.EncodeToString(hash[:])
	return fmt.Sprintf("%s{%s}", cfg.App.FlagFormat, hashHex[:32])
}
