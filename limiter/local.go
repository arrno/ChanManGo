package limiter

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type LocalLimiter struct {
	mu        sync.Mutex
	visitors  map[string]*rate.Limiter
	rateLimit rate.Limit
	burst     int
}

func NewLocalLimiter() *LocalLimiter {
	return &LocalLimiter{
		visitors:  make(map[string]*rate.Limiter),
		rateLimit: rate.Every(5 * time.Second),
		burst:     10,
	}
}

func (l *LocalLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	var limiter *rate.Limiter
	var exists bool
	if limiter, exists = l.visitors[ip]; !exists {
		limiter = rate.NewLimiter(l.rateLimit, l.burst)
		l.visitors[ip] = limiter
	}

	go func() {
		time.Sleep(5 * time.Minute)
		l.mu.Lock()
		delete(l.visitors, ip)
		l.mu.Unlock()
	}()

	return limiter.Allow()
}
