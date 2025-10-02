package cache

import (
	"time"

	"github.com/intraware/rodan/internal/config"
	"github.com/intraware/rodan/internal/utils/values"
)

type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	Delete(key K)
	Reset()
}

type CacheOpts struct {
	TimeToLive    time.Duration
	CleanInterval *time.Duration
	Revaluate     *bool
	Prefix        string
}

var cfg *config.CacheConfig = nil

func NewCache[K comparable, V any](opts *CacheOpts) Cache[K, V] {
	if cfg == nil {
		cfg = &values.GetConfig().App.AppCache
	}
	if cfg.InApp {
		return newAppCache[K, V](opts)
	} else if cfg.ServiceType == "redis" {
		return newRedisCache[K, V](opts)
	} else {
		return newAppCache[K, V](opts)
	}
}
