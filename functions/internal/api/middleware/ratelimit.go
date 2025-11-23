package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	maxRequests int
	window      time.Duration
	clients     map[string]*clientLimit
	mu          sync.RWMutex
}

type clientLimit struct {
	requests []time.Time
}

// NewRateLimiter creates a new rate limiter
// maxRequests: number of requests allowed
// window: time window for rate limiting (e.g., 1 minute)
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		maxRequests: maxRequests,
		window:      window,
		clients:     make(map[string]*clientLimit),
	}

	// Cleanup old entries every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return rl
}

// Allow checks if a request from the client should be allowed
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[clientID]
	if !exists {
		client = &clientLimit{
			requests: []time.Time{now},
		}
		rl.clients[clientID] = client
		return true
	}

	// Remove requests outside the window
	cutoff := now.Add(-rl.window)
	var recentRequests []time.Time
	for _, req := range client.requests {
		if req.After(cutoff) {
			recentRequests = append(recentRequests, req)
		}
	}

	// Check if limit exceeded
	if len(recentRequests) >= rl.maxRequests {
		client.requests = recentRequests
		return false
	}

	// Add new request
	recentRequests = append(recentRequests, now)
	client.requests = recentRequests
	return true
}

// cleanup removes old client entries
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window * 2)

	for clientID, client := range rl.clients {
		hasRecent := false
		for _, req := range client.requests {
			if req.After(cutoff) {
				hasRecent = true
				break
			}
		}
		if !hasRecent {
			delete(rl.clients, clientID)
		}
	}
}

// Middleware returns a Chi middleware for rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := r.RemoteAddr
		if !rl.Allow(clientID) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
