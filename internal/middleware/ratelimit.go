package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitConfig configura los rate limiters.
type RateLimitConfig struct {
	// Global: máx requests por segundo a nivel de servidor.
	GlobalRPS float64
	// GlobalBurst: ráfaga máxima permitida.
	GlobalBurst int
	// PerIPRPS: requests por segundo por IP (endpoints generales).
	PerIPRPS float64
	// PerIPBurst: ráfaga por IP.
	PerIPBurst int
	// CleanupInterval: cada cuánto limpiar IPs inactivas.
	CleanupInterval time.Duration
}

// DefaultRateLimitConfig devuelve una configuración por defecto para ~500 rps.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		GlobalRPS:       600, // un poco de margen sobre 500
		GlobalBurst:     100,
		PerIPRPS:        50,
		PerIPBurst:      20,
		CleanupInterval: 5 * time.Minute,
	}
}

// ipLimiter guarda el rate limiter y la última vez que se usó.
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter maneja rate limiting global + per-IP.
type RateLimiter struct {
	global  *rate.Limiter
	mu      sync.Mutex
	ips     map[string]*ipLimiter
	perIPR  rate.Limit
	perIPB  int
	cleanup time.Duration
}

// NewRateLimiter crea un nuevo RateLimiter.
func NewRateLimiter(cfg RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		global:  rate.NewLimiter(rate.Limit(cfg.GlobalRPS), cfg.GlobalBurst),
		ips:     make(map[string]*ipLimiter),
		perIPR:  rate.Limit(cfg.PerIPRPS),
		perIPB:  cfg.PerIPBurst,
		cleanup: cfg.CleanupInterval,
	}

	// Limpiar IPs inactivas periódicamente para evitar leak de memoria.
	go rl.cleanupLoop()

	return rl
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, l := range rl.ips {
			if time.Since(l.lastSeen) > rl.cleanup {
				delete(rl.ips, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) getIPLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	l, exists := rl.ips[ip]
	if !exists {
		l = &ipLimiter{
			limiter: rate.NewLimiter(rl.perIPR, rl.perIPB),
		}
		rl.ips[ip] = l
	}
	l.lastSeen = time.Now()
	return l.limiter
}

// GlobalRateLimitMiddleware aplica rate limiting global + per-IP.
func GlobalRateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check global limit first
		if !rl.global.Allow() {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   true,
				"message": "server is busy, please try again later",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			return
		}

		// Check per-IP limit
		ip := c.ClientIP()
		if !rl.getIPLimiter(ip).Allow() {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   true,
				"message": "too many requests from your IP, please slow down",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			return
		}

		c.Next()
	}
}

// StrictRateLimitMiddleware aplica un límite más estricto por IP.
// Ideal para endpoints sensibles como /auth/login.
func StrictRateLimitMiddleware(requestsPerSecond float64, burst int) gin.HandlerFunc {
	var mu sync.Mutex
	ips := make(map[string]*ipLimiter)

	// Cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			for ip, l := range ips {
				if time.Since(l.lastSeen) > 5*time.Minute {
					delete(ips, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		l, exists := ips[ip]
		if !exists {
			l = &ipLimiter{
				limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), burst),
			}
			ips[ip] = l
		}
		l.lastSeen = time.Now()
		mu.Unlock()

		if !l.limiter.Allow() {
			c.Header("Retry-After", "5")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   true,
				"message": "too many login attempts, please wait",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			return
		}

		c.Next()
	}
}
