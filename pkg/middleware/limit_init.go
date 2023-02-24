package middleware

import (
	"sync"
	"time"

	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/viper"
)

type TokenBucket struct {
	Rate         int64 // 固定的token放入速率，r/s
	Capacity     int64 // 桶的容量
	Tokens       int64 // 桶中当前token数量
	LastTokenSec int64 // 桶上次放token的时间

	lock sync.Mutex
}

func (tb *TokenBucket) Allow() bool {
	tb.lock.Lock()
	defer tb.lock.Unlock()
	now := time.Now().Unix()
	tb.Tokens = tb.Tokens + (now-tb.LastTokenSec)*tb.Rate
	if tb.Tokens > tb.Capacity {
		tb.Tokens = tb.Capacity
	}
	tb.LastTokenSec = now
	if tb.Tokens > 0 {
		tb.Tokens--
		return true
	}
	return false
}

func MakeTokenBucket(c, r int64) *TokenBucket {
	return &TokenBucket{
		Rate:         r,
		Capacity:     c,
		Tokens:       int64(limiterTokenInit),
		LastTokenSec: time.Now().Unix(),
	}
}

type TokenBuckets struct {
	buckets  map[string]*TokenBucket
	capacity int64
	rate     int64

	lock sync.Mutex
}

func MakeTokenBuckets(capacity, rate int64) *TokenBuckets {
	return &TokenBuckets{
		buckets:  make(map[string]*TokenBucket),
		capacity: capacity,
		rate:     rate,
	}
}

func (tbs *TokenBuckets) Allow(token string) bool {
	tbs.lock.Lock()
	defer tbs.lock.Unlock()
	if bucket, ok := tbs.buckets[token]; ok {
		return bucket.Allow()
	} else {
		tbs.buckets[token] = MakeTokenBucket(tbs.capacity, tbs.rate)
		return tbs.buckets[token].Allow()
	}
}

// 对每个token限流，不管它请求的API
var (
	apiConfig        = viper.Init("api")
	limiterCapacity  = apiConfig.Viper.GetInt("server.limit.capacity")
	limiterRate      = apiConfig.Viper.GetInt("server.limit.rate")
	limiterTokenInit = apiConfig.Viper.GetInt("server.limit.tokenInit")
	CurrentLimiter   = MakeTokenBuckets(int64(limiterCapacity), int64(limiterRate)) // 后续用redis缓存每个token对应的令牌桶
)

func init() {

}
