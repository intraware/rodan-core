package sandbox

import (
	"time"

	"github.com/intraware/rodan/models"
)

var containerPool = newPool()

type SandBox struct {
	UserID        int
	TeamID        int
	ChallengeMeta models.Challenge
	Container     *container
	CreatedAt     time.Time
	Active        bool
}

func NewSandBox(userID, teamID int, challenge *models.Challenge) *SandBox {
	return &SandBox{
		UserID:        userID,
		TeamID:        teamID,
		ChallengeMeta: *challenge,
		Container:     nil,
		CreatedAt:     time.Now(),
	}
}

func (s *SandBox) Start() {
	containerPool.Aquire(s.ChallengeMeta.ID)
	s.Active = true
} // start a sandbox

func (s *SandBox) Stop() {
	s.Active = false
} // release a sandbox

func (s *SandBox) Regenerate(challenge *models.Challenge) {} // remove existing container and make a new one

func (s *SandBox) ExtendTTL() {
	s.CreatedAt = time.Now() // just change the created at .. and it will extend the time naa
} // extend the ttl of sandbox

func (s *SandBox) GetMeta() {
	// gotta look into these
} // details like ports exposed etc and time left
