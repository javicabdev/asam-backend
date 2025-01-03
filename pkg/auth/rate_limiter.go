package auth

import (
	"github.com/javicabdev/asam-backend/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	visitors map[string]*visitor
	mtx      sync.RWMutex
	// Límites configurables
	limit           rate.Limit // Peticiones por segundo
	burst           int        // Máximo de peticiones en burst
	cleanupInterval time.Duration
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewRateLimiter(rps rate.Limit, burst int, cleanup time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors:        make(map[string]*visitor),
		limit:           rps,
		burst:           burst,
		cleanupInterval: cleanup,
	}

	// Iniciar rutina de limpieza
	go rl.cleanupVisitors()
	return rl
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mtx.Lock()
	defer rl.mtx.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.limit, rl.burst)
		rl.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	// Actualizar última visita
	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(rl.cleanupInterval)

		rl.mtx.Lock()
		initial := len(rl.visitors)
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.cleanupInterval {
				delete(rl.visitors, ip)
			}
		}
		cleaned := initial - len(rl.visitors)
		if cleaned > 0 {
			logger.Debug("Rate limiter cleanup completed",
				zap.Int("cleaned", cleaned),
				zap.Int("remaining", len(rl.visitors)),
			)
		}
		rl.mtx.Unlock()
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtener IP del visitante
		ip := r.RemoteAddr
		if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
			ip = forwardedFor
		}

		// Obtener limiter para esta IP
		limiter := rl.getVisitor(ip)

		if !limiter.Allow() {
			logger.Error("Rate limit exceeded",
				zap.String("ip", ip),
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
				zap.String("user_agent", r.UserAgent()),
			)

			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
