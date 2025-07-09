package values

import (
	"sync/atomic"

	"github.com/intraware/rodan/config"
)

var cfg atomic.Value

func SetConfig(c *config.Config) {
	cfg.Store(c)
}

func GetConfig() *config.Config {
	val := cfg.Load()
	if val == nil {
		return nil
	}
	return val.(*config.Config)
}
