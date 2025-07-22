package sandbox

type container struct {
	ContainerID string
	ImageName   string
	ChallengeID int
	TTL         int64 // duration in sec
}

func newContainer(challengeID int, imageName string, ttl int64) *container {
	return &container{
		ContainerID: "container_id",
		ImageName:   imageName,
		ChallengeID: challengeID,
		TTL:         ttl,
	}
}

func (c *container) Start()                   {}
func (c *container) Stop()                    {}
func (c *container) Discard()                 {}
func (c *container) Reset()                   {}
func (c *container) GenerateFlag(flag string) {}
