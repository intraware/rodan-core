package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/intraware/rodan/internal/utils/values"
	redis_cache "github.com/intraware/rodan/pkg/cache"
	"github.com/redis/go-redis/v9"
)

type redisCache[K comparable, V any] struct {
	prefix  string
	client  RedisClient
	opts    *CacheOpts
	version int
}

type RedisClient struct {
	redis *redis_cache.Cache
	ctx   context.Context
}

var redisObj RedisClient

func InitRedis(ctx context.Context) {
	cacheCfg := values.GetConfig().App.AppCache
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{"redis-server": cacheCfg.ServiceUrl},
	})
	redisCache := redis_cache.New(&redis_cache.Options{
		Redis:      ring,
		LocalCache: redis_cache.NewTinyLFU(cacheCfg.InternalCacheSize, cacheCfg.InternalCacheDuration),
	})
	redisTemp := RedisClient{
		redis: redisCache,
		ctx:   ctx,
	}
	redisObj = redisTemp
}

func newRedisCache[K comparable, V any](opts *CacheOpts) Cache[K, V] {
	var prefixRandom string
	var versionInt int
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		prefixRandom = "prefix_fallback"
	} else {
		prefixRandom = "prefix_" + hex.EncodeToString(b)
	}
	if opts.TimeToLive.Seconds() == 0 {
		versionInt = 0
	} else {
		versionInt = 1
	}
	return &redisCache[K, V]{
		opts:    opts,
		client:  redisObj,
		prefix:  prefixRandom,
		version: versionInt,
	}
}

func (r *redisCache[K, V]) Get(key K) (val V, exists bool) {
	keyStr := fmt.Sprintf("%s_%d_%v", r.prefix, r.version, key)
	err := r.client.redis.Get(r.client.ctx, keyStr, &val)
	exists = (err == nil) || (err == redis.Nil)
	return
}

func (r *redisCache[K, V]) Set(key K, val V) {
	keyStr := fmt.Sprintf("%s_%d_%v", r.prefix, r.version, key)
	r.client.redis.DeleteFromLocalCache(keyStr)
	r.client.redis.Set(&redis_cache.Item{ // TODO: change the set function to handle the error
		Ctx:   r.client.ctx,
		Key:   keyStr,
		Value: val,
		TTL:   r.opts.TimeToLive,
	})
}

func (r *redisCache[K, V]) Delete(key K) {
	keyStr := fmt.Sprintf("%s_%d_%v", r.prefix, r.version, key)
	r.client.redis.Delete(r.client.ctx, keyStr)
}

func (r *redisCache[K, V]) Reset() { // TODO: change the function to handle the errror
	if r.opts.TimeToLive.Seconds() == 0 {
		r.version += 1
		return
	} else {
		prefixString := fmt.Sprintf("%s_", r.prefix)
		r.client.redis.DeletePrefix(r.client.ctx, prefixString)
		return
	}
}
