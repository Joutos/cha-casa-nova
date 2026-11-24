package middleware

import (
	"sync"
	"time"

	"github.com/chadecasanova/backend/internal/apperr"
	"github.com/gin-gonic/gin"
)

// RateLimiter é um token bucket em memória por chave (IP). Suficiente para uma
// única instância; com várias réplicas, troque por um armazenamento compartilhado.
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	capacity float64
	refill   float64 // tokens por segundo
	now      func() time.Time
}

type bucket struct {
	tokens float64
	last   time.Time
}

func NewRateLimiter(perMinute int) *RateLimiter {
	rl := &RateLimiter{
		buckets:  map[string]*bucket{},
		capacity: float64(perMinute),
		refill:   float64(perMinute) / 60,
		now:      time.Now,
	}
	return rl
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := rl.now()
	b, ok := rl.buckets[key]
	if !ok {
		b = &bucket{tokens: rl.capacity, last: now}
		rl.buckets[key] = b
	}
	b.tokens = min(rl.capacity, b.tokens+now.Sub(b.last).Seconds()*rl.refill)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Cleanup remove baldes cheios (inativos) para limitar o uso de memória.
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := rl.now()
	for k, b := range rl.buckets {
		if b.tokens+now.Sub(b.last).Seconds()*rl.refill >= rl.capacity {
			delete(rl.buckets, k)
		}
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.Allow(c.ClientIP()) {
			c.Error(apperr.TooManyRequests())
			c.Abort()
			return
		}
		c.Next()
	}
}
