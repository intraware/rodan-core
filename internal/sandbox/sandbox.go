package sandbox

import (
	"context"
	"errors"
	"fmt"
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
	Context       context.Context
}

func NewSandBox(userID, teamID int, challenge *models.Challenge) *SandBox {
	ctx := context.Background() // change it
	return &SandBox{
		UserID:        userID,
		TeamID:        teamID,
		ChallengeMeta: *challenge,
		Container:     nil,
		CreatedAt:     time.Now(),
		Context:       ctx,
	}
}

func (s *SandBox) Start() (err error) {
	var ctr *container
	ctr, err = containerPool.Aquire(s.ChallengeMeta.ID)
	if err != nil && !errors.Is(err, errNoContainers) {
		return ErrFailedToCreateContainer
	}
	if ctr == nil && errors.Is(err, errNoContainers) {
		container_name := fmt.Sprintf("%d-%d-%d", s.UserID, s.TeamID, s.ChallengeMeta.ID)
		ctr, err = newContainer(
			s.Context,
			s.ChallengeMeta.ID,
			container_name,
			s.ChallengeMeta.DynamicConfig.DockerImage,
			time.Duration(s.ChallengeMeta.DynamicConfig.TTL),
		)
	}
	if err != nil {
		return ErrFailedToCreateContainer
	}
	err = ctr.Start()
	if err != nil {
		ctr.Discard()
		return ErrFailedToStartContainer
	}
	s.Container = ctr
	s.Active = true
	return nil
}

func (s *SandBox) Stop() error {
	if s.Container == nil {
		return ErrContainerNotFound
	}
	if s.ChallengeMeta.DynamicConfig.Reusable {
		s.Container.Stop()
		go s.Container.Reset()
		if err := containerPool.Release(s.Container); err != nil {
			if errors.Is(err, errPoolFull) {
				if derr := s.Container.Discard(); derr != nil {
					return ErrFailedToDiscardContainer
				}
			} else {
				return ErrFailedToStopContainer
			}
		}
	} else {
		if derr := s.Container.Discard(); derr != nil {
			return ErrFailedToDiscardContainer
		}
	}
	s.Container = nil
	s.Active = false
	return nil
}

func (s *SandBox) Regenerate(challenge *models.Challenge) (err error) {
	err = nil
	if s.Container == nil {
		err = ErrContainerNotFound
		return
	}
	s.Container.Stop()
	err = s.Container.Discard()
	if err != nil {
		err = ErrFailedToDiscardContainer
		return
	}
	var ctr *container
	container_name := fmt.Sprintf("%d-%d-%d", s.UserID, s.TeamID, s.ChallengeMeta.ID)
	ctr, err = newContainer(
		s.Context,
		s.ChallengeMeta.ID,
		container_name,
		s.ChallengeMeta.DynamicConfig.DockerImage,
		time.Duration(s.ChallengeMeta.DynamicConfig.TTL),
	)
	if err != nil {
		err = ErrFailedToCreateContainer
		return
	}
	err = ctr.Start()
	if err != nil {
		ctr.Stop()
		err = ErrFailedToStartContainer
		return
	}
	s.Container = ctr
	return
}

func (s *SandBox) ExtendTTL() {
	s.Container.StartedAt = time.Now() // this should be increased
} // extend the ttl of sandbox

func (s *SandBox) GetMeta() {
	// gotta look into these
} // details like ports exposed etc and time left
