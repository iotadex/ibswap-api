package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter .
type IPRateLimiter struct {
	ips         map[string]*rate.Limiter
	mu          *sync.RWMutex
	r           rate.Limit
	b           int
	expiredTime time.Duration // count of hours
}

// NewIPRateLimiter .
func NewIPRateLimiter(r rate.Limit, b int, exp time.Duration) *IPRateLimiter {
	i := &IPRateLimiter{
		ips:         make(map[string]*rate.Limiter),
		mu:          &sync.RWMutex{},
		r:           r,
		b:           b,
		expiredTime: exp,
	}
	go i.expired()
	return i
}

// AddIP creates a new rate limiter and adds it to the ips map,
// using the IP address as the key
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)

	i.ips[ip] = limiter

	return limiter
}

// GetLimiter returns the rate limiter for the provided IP address if it exists.
// Otherwise calls AddIP to add IP address to the map
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	limiter, exists := i.ips[ip]

	if !exists {
		i.mu.Unlock()
		return i.AddIP(ip)
	}

	i.mu.Unlock()

	return limiter
}

func (i *IPRateLimiter) expired() {
	ticker := time.NewTicker(time.Hour * i.expiredTime)
	for range ticker.C {
		i.mu.Lock()
		i.ips = make(map[string]*rate.Limiter)
		i.mu.Unlock()
	}
}

var signLimiter = NewIPRateLimiter(1, 10, 6)
var publicLimiter = NewIPRateLimiter(20, 40, 6)

func SignIpRateLimiterWare(c *gin.Context) {
	limiter := signLimiter.GetLimiter(c.ClientIP())
	if !limiter.Allow() {
		c.Abort()
		c.String(http.StatusOK, "too many requests.")
		return
	}
	c.Next()
}

func PublicIpRateLimiterWare(c *gin.Context) {
	limiter := publicLimiter.GetLimiter(c.ClientIP())
	if !limiter.Allow() {
		c.Abort()
		c.String(http.StatusOK, "too many requests.")
		return
	}
	c.Next()
}
