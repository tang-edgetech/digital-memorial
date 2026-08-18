package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	burst    int
}

func newIPLimiter(r rate.Limit, burst int) *ipLimiter {
	return &ipLimiter{limiters: make(map[string]*rate.Limiter), r: r, burst: burst}
}

func (l *ipLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	lim, ok := l.limiters[ip]
	if !ok {
		lim = rate.NewLimiter(l.r, l.burst)
		l.limiters[ip] = lim
	}
	return lim
}

// RateLimit caps requests per client IP — used on login/setup endpoints to
// slow down credential-guessing and setup-abuse attempts. In-memory and
// per-instance; swap for a distributed limiter if horizontally scaled.
func RateLimit(perSecond float64, burst int) gin.HandlerFunc {
	limiter := newIPLimiter(rate.Limit(perSecond), burst)
	return func(c *gin.Context) {
		if !limiter.get(c.ClientIP()).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests", "code": "rate_limited"})
			return
		}
		c.Next()
	}
}
