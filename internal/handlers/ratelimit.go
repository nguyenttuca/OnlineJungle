package handlers

import (
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	LoginLimiter  = NewIPRateLimiter(5, 1*time.Minute)  // 5 login/phút/IP
	SubmitLimiter = NewIPRateLimiter(2, 30*time.Second) // 2 submit/30s/IP
)

type IPRateLimiter struct {
	visitors sync.Map
	rate     int
	window   time.Duration
}

type visitorEntry struct {
	count   int
	resetAt time.Time
	mu      sync.Mutex
}

func NewIPRateLimiter(rate int, window time.Duration) *IPRateLimiter {
	rl := &IPRateLimiter{
		rate:   rate,
		window: window,
	}

	// Background cleanup goroutine
	go func() {
		for {
			time.Sleep(window)
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *IPRateLimiter) cleanup() {
	now := time.Now()
	rl.visitors.Range(func(key, value interface{}) bool {
		entry := value.(*visitorEntry)
		entry.mu.Lock()
		if now.After(entry.resetAt) {
			rl.visitors.Delete(key)
		}
		entry.mu.Unlock()
		return true
	})
}

func (rl *IPRateLimiter) Allow(ip string) bool {
	val, _ := rl.visitors.LoadOrStore(ip, &visitorEntry{
		resetAt: time.Now().Add(rl.window),
	})

	entry := val.(*visitorEntry)
	entry.mu.Lock()
	defer entry.mu.Unlock()

	now := time.Now()
	if now.After(entry.resetAt) {
		entry.count = 0
		entry.resetAt = now.Add(rl.window)
	}

	entry.count++
	return entry.count <= rl.rate
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		return ip
	}
	return host
}

func (rl *IPRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		if !rl.Allow(ip) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
