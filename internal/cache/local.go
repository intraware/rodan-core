package cache

import (
	"github.com/AnimeKaizoku/cacher"
)

func newAppCache[K comparable, V any](opts *CacheOpts) Cache[K, V] {
	newOpts := &cacher.NewCacherOpts{
		TimeToLive:  opts.TimeToLive,
		CleanerMode: cacher.CleaningCentral,
	}
	if opts.CleanInterval != nil {
		newOpts.CleanInterval = *opts.CleanInterval
	}
	if opts.Revaluate != nil {
		newOpts.Revaluate = *opts.Revaluate
	}
	return cacher.NewCacher[K, V](newOpts)
}
