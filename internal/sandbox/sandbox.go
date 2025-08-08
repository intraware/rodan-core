package sandbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/intraware/rodan/internal/models"
	"github.com/intraware/rodan/internal/utils/docker"
)

var containerPool = newPool()

type SandBox struct {
	UserID        int
	TeamID        int
	ChallengeMeta models.Challenge
	Container     *container
	CreatedAt     time.Time
	Active        bool
	Flag          string
	Context       context.Context
	CancelFunc    context.CancelFunc
}

func NewSandBox(userID, teamID int, challenge *models.Challenge, flag string) *SandBox {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(challenge.DynamicConfig.TTL))
	return &SandBox{
		UserID:        userID,
		TeamID:        teamID,
		ChallengeMeta: *challenge,
		Container:     nil,
		CreatedAt:     time.Now(),
		Context:       ctx,
		CancelFunc:    cancel,
		Flag:          flag,
	}
}

func (s *SandBox) Start() error {
	if s.CancelFunc != nil {
		s.CancelFunc()
	}
	ttl := time.Duration(s.ChallengeMeta.DynamicConfig.TTL)
	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	s.Context = ctx
	s.CancelFunc = cancel
	ctr, err := containerPool.Aquire(s.ChallengeMeta.ID)
	if err != nil && !errors.Is(err, errNoContainers) {
		cancel()
		s.CancelFunc = nil
		return ErrFailedToCreateContainer
	}
	if ctr == nil && errors.Is(err, errNoContainers) {
		containerName := fmt.Sprintf("%d-%d-%d", s.UserID, s.TeamID, s.ChallengeMeta.ID)
		ctr, err = newContainer(
			ctx,
			s.ChallengeMeta.ID,
			containerName,
			s.ChallengeMeta.DynamicConfig.DockerImage,
			ttl,
			s.ChallengeMeta.DynamicConfig.ExposedPorts,
		)
		if err != nil {
			cancel()
			s.CancelFunc = nil
			return ErrFailedToCreateContainer
		}
	}
	err = ctr.GenerateFlag(s.Flag)
	if err != nil {
		ctr.Discard()
		cancel()
		s.CancelFunc = nil
		return ErrFailedToGenerateFlag
	}
	err = ctr.Start()
	if err != nil {
		ctr.Discard()
		cancel()
		s.CancelFunc = nil
		return ErrFailedToStartContainer
	}
	s.Container = ctr
	s.Active = true
	return nil
}

func (s *SandBox) Stop() error {
	if s.CancelFunc != nil {
		s.CancelFunc()
		s.CancelFunc = nil
	}
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
	ttl := time.Duration(challenge.DynamicConfig.TTL)
	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	s.Context = ctx
	s.CancelFunc = cancel
	ctr, err = newContainer(
		s.Context,
		s.ChallengeMeta.ID,
		container_name,
		s.ChallengeMeta.DynamicConfig.DockerImage,
		time.Duration(s.ChallengeMeta.DynamicConfig.TTL),
		s.ChallengeMeta.DynamicConfig.ExposedPorts,
	)
	if err != nil {
		err = ErrFailedToCreateContainer
		return
	}
	err = ctr.GenerateFlag(s.Flag)
	if err != nil {
		ctr.Stop()
		err = ErrFailedToGenerateFlag
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
	if s.CancelFunc != nil {
		s.CancelFunc()
	}
	ttl := time.Duration(s.ChallengeMeta.DynamicConfig.TTL)
	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	s.Context = ctx
	s.CancelFunc = cancel
	s.Container.StartedAt = time.Now()
}

func (s *SandBox) GetMeta() (SandBoxResponse, error) {
	var response SandBoxResponse
	ports, err := docker.GetBoundPorts(s.Context, s.Container.ContainerID)
	if err != nil {
		return response, err
	}
	response.Ports = make([]string, 0, len(ports))
	for _, port := range ports {
		response.Ports = append(response.Ports, port)
	}
	expiryTime := s.Container.StartedAt.Add(s.Container.TTL)
	timeLeft := time.Until(expiryTime).Seconds()
	timeLeft = max(timeLeft, 0)
	response.TimeLeft = int64(timeLeft)
	return response, nil
}
