package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type client struct {
	count    int
	lastSeen time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*client
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*client),
		limit:   limit,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(rl.window)
		rl.mu.Lock()
		for ip, c := range rl.clients {
			if time.Since(c.lastSeen) > rl.window {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) isAllowed(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	c, exists := rl.clients[ip]
	if !exists {
		rl.clients[ip] = &client{count: 1, lastSeen: time.Now()}
		return true
	}

	if time.Since(c.lastSeen) > rl.window {
		c.count = 1
		c.lastSeen = time.Now()
		return true
	}

	c.count++
	c.lastSeen = time.Now()
	return c.count <= rl.limit
}

// RateLimitMiddleware applies rate limiting per IP.
// General API: 100 requests per minute.
func RateLimitMiddleware() gin.HandlerFunc {
	limiter := NewRateLimiter(100, time.Minute)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.isAllowed(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, try again later",
			})
			return
		}
		c.Next()
	}
}

// AuthRateLimitMiddleware applies stricter rate limiting for auth endpoints.
// 10 attempts per minute per IP to prevent brute force.
func AuthRateLimitMiddleware() gin.HandlerFunc {
	limiter := NewRateLimiter(10, time.Minute)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.isAllowed(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many attempts, try again later",
			})
			return
		}
		c.Next()
	}
}
