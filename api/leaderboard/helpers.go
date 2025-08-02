package leaderboard

import (
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"sort"

	"github.com/intraware/rodan/api/shared"
	"github.com/intraware/rodan/models"
	"github.com/intraware/rodan/utils/values"
)

type UserPoints struct {
	UserID   int
	UserName string
	Points   float64
}

type TeamPoints struct {
	TeamID   int
	TeamName string
	Points   float64
}

var (
	userLeaderboardCache atomic.Pointer[[]UserPoints]
	teamLeaderboardCache atomic.Pointer[[]TeamPoints]

	cacheDirtyFlag   atomic.Bool
	dirtyTriggerLock sync.Mutex
	dirtyTimer       *time.Timer
)

func MarkLeaderboardDirty() {
	dirtyTriggerLock.Lock()
	defer dirtyTriggerLock.Unlock()
	if dirtyTimer != nil {
		dirtyTimer.Stop()
	}
	debounceDuration := values.GetConfig().App.Leaderboard.DebounceTimer
	dirtyTimer = time.AfterFunc(debounceDuration, func() {
		cacheDirtyFlag.Store(true)
	})
}

func smoothScore(solves int, maxPoints int, minPoints int, total int, offset int, power float64) int {
	if solves <= offset {
		return maxPoints
	}
	if solves > total {
		solves = total
	}
	x := float64(solves-offset) / float64(total-offset)
	score := float64(maxPoints) - float64(maxPoints-minPoints)*math.Pow(x, power)
	score = max(score, float64(minPoints))
	return int(math.Round(score))
}

func timeAdjustedScoreWithCreatedAt(
	solveTime time.Time,
	allSolveTimes []time.Time,
	minPoints, maxPoints, offset int,
	power float64,
) float64 {
	totalSolves := len(allSolveTimes)
	solvesSoFar := totalSolves
	base := float64(smoothScore(solvesSoFar, maxPoints, minPoints, totalSolves, offset, power))
	sort.Slice(allSolveTimes, func(i, j int) bool {
		return allSolveTimes[i].Before(allSolveTimes[j])
	})
	var rank int
	for i, t := range allSolveTimes {
		if t.Equal(solveTime) {
			rank = i
			break
		}
	}
	bonus := float64(totalSolves-rank) / float64(totalSolves+1)
	bonus *= 1e-9
	return base + bonus
}

func updateLeaderboards() {
	var solves []models.Solve
	userBlackList := values.GetConfig().App.Leaderboard.UserBlackList
	teamBlackList := values.GetConfig().App.Leaderboard.TeamBlackList
	err := models.DB.Where("user_id NOT IN ? AND team_id NOT IN ?", userBlackList, teamBlackList).Order("challenge_id, created_at").Find(&solves).Error
	if err != nil {
		log.Println("[leaderboard] DB error:", err)
		return
	}
	challengeSolves := make(map[int][]models.Solve)
	solveTimes := make(map[int][]time.Time)
	for _, s := range solves {
		challengeSolves[s.ChallengeID] = append(challengeSolves[s.ChallengeID], s)
		solveTimes[s.ChallengeID] = append(solveTimes[s.ChallengeID], s.CreatedAt)
	}
	userScores := make(map[int]float64)
	userToTeam := make(map[int]int)

	challengeToMeta := make(map[int]models.Challenge)
	var missingChallengeIDs []int
	for cid := range challengeSolves {
		if val, ok := shared.ChallengeCache.Get(cid); ok {
			challenge := val
			challengeToMeta[cid] = challenge
		} else {
			missingChallengeIDs = append(missingChallengeIDs, cid)
		}
	}
	if len(missingChallengeIDs) > 0 {
		var challenges []models.Challenge
		if err := models.DB.Where("id IN ?", missingChallengeIDs).Find(&challenges).Error; err == nil {
			for _, ch := range challenges {
				shared.ChallengeCache.Set(ch.ID, ch)
				challengeToMeta[ch.ID] = ch
			}
		}
	}
	offset := values.GetConfig().App.Leaderboard.FullPointsThreshold
	power := values.GetConfig().App.Leaderboard.DecaySharpness
	for cid, solveList := range challengeSolves {
		times := solveTimes[cid]
		challenge := challengeToMeta[cid]
		for _, s := range solveList {
			points := timeAdjustedScoreWithCreatedAt(
				s.CreatedAt,
				times,
				challenge.PointsMin,
				challenge.PointsMax,
				offset,
				power,
			)
			userScores[s.UserID] += points
			userToTeam[s.UserID] = s.TeamID
		}
	}
	userToName := make(map[int]string)
	var missingIDs []int
	for uid := range userScores {
		if val, ok := shared.UserCache.Get(uid); ok {
			user := val
			userToName[uid] = user.Username
		} else {
			missingIDs = append(missingIDs, uid)
		}
	}
	if len(missingIDs) > 0 {
		var users []models.User
		if err := models.DB.Where("id IN ? AND id NOT IN ? AND team_id NOT IN ?", missingIDs, userBlackList, teamBlackList).Find(&users).Error; err == nil {
			for _, u := range users {
				shared.UserCache.Set(u.ID, u)
				userToName[u.ID] = u.Username
			}
		}
	}
	var userLeaderboard []UserPoints
	for uid, pts := range userScores {
		userLeaderboard = append(userLeaderboard, UserPoints{
			UserID:   uid,
			UserName: userToName[uid],
			Points:   pts,
		})
	}
	sort.Slice(userLeaderboard, func(i, j int) bool {
		return userLeaderboard[i].Points > userLeaderboard[j].Points
	})
	userLeaderboardCache.Store(&userLeaderboard)

	teamScores := make(map[int]float64)
	teamIDSet := make(map[int]struct{})
	for uid, pts := range userScores {
		tid := userToTeam[uid]
		teamScores[tid] += pts
		teamIDSet[tid] = struct{}{}
	}
	teamToName := make(map[int]string)
	var missingTeamIDs []int
	for tid := range teamIDSet {
		if val, ok := shared.TeamCache.Get(tid); ok {
			team := val
			teamToName[tid] = team.Name
		} else {
			missingTeamIDs = append(missingTeamIDs, tid)
		}
	}
	if len(missingTeamIDs) > 0 {
		var teams []models.Team
		if err := models.DB.Where("id IN ? AND id NOT in ?", missingTeamIDs, teamBlackList).Find(&teams).Error; err == nil {
			for _, t := range teams {
				shared.TeamCache.Set(t.ID, t)
				teamToName[t.ID] = t.Name
			}
		}
	}
	var teamLeaderboard []TeamPoints
	for tid, pts := range teamScores {
		teamLeaderboard = append(teamLeaderboard, TeamPoints{
			TeamID:   tid,
			TeamName: teamToName[tid],
			Points:   pts,
		})
	}
	sort.Slice(teamLeaderboard, func(i, j int) bool {
		return teamLeaderboard[i].Points > teamLeaderboard[j].Points
	})
	teamLeaderboardCache.Store(&teamLeaderboard)
	LastModified.Store(time.Now().UTC())
	log.Println("[leaderboard] cache updated")
}

func maybeRefreshLeaderboard() {
	if cacheDirtyFlag.Load() {
		if cacheDirtyFlag.Swap(false) {
			go updateLeaderboards()
		}
	}
}

func GetCachedUserLeaderboard() []UserPoints {
	maybeRefreshLeaderboard()
	ptr := userLeaderboardCache.Load()
	if ptr == nil {
		return nil
	}
	return *ptr
}

func GetCachedTeamLeaderboard() []TeamPoints {
	maybeRefreshLeaderboard()
	ptr := teamLeaderboardCache.Load()
	if ptr == nil {
		return nil
	}
	return *ptr
}
