package proxy

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/dipjyotimetia/jarvis/config"
	"github.com/dipjyotimetia/jarvis/internal/validator"
)

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares to a handler in order
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// LoggingMiddleware logs HTTP requests and responses
func LoggingMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Log request
			slog.Info("Incoming request",
				"method", r.Method,
				"url", r.URL.String(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)

			// Create response recorder
			rec := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call next handler
			next.ServeHTTP(rec, r)

			// Log response
			duration := time.Since(start)
			slog.Info("Request completed",
				"method", r.Method,
				"url", r.URL.String(),
				"status", rec.statusCode,
				"duration_ms", duration.Milliseconds(),
			)
		})
	}
}

// RecoveryMiddleware recovers from panics and returns 500
func RecoveryMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					slog.Error("Panic recovered",
						"error", err,
						"url", r.URL.String(),
						"method", r.Method,
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// ValidationMiddleware validates requests against OpenAPI spec
func ValidationMiddleware(cfg *config.Config, apiValidator *validator.APIValidator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if apiValidator != nil && cfg.APIValidation.ValidateRequests {
				if err := apiValidator.ValidateRequest(r); err != nil {
					slog.Warn("OpenAPI request validation failed",
						"method", r.Method,
						"path", r.URL.Path,
						"error", err,
					)

					if !cfg.APIValidation.ContinueOnValidation {
						http.Error(w, fmt.Sprintf("Request validation error: %v", err), http.StatusBadRequest)
						return
					}

					w.Header().Set("X-API-Validation-Error", "request")
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// MetricsMiddleware collects metrics about requests
func MetricsMiddleware(metrics *Metrics) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Increment total requests
			metrics.IncrementRequests()

			rec := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			// Record duration
			duration := time.Since(start)
			metrics.RecordDuration(r.Method, r.URL.Path, duration)

			// Record status code
			metrics.RecordStatusCode(rec.statusCode)
		})
	}
}

// RateLimitMiddleware implements simple rate limiting
func RateLimitMiddleware(requestsPerSecond int) Middleware {
	// Simple token bucket implementation
	ticker := time.NewTicker(time.Second / time.Duration(requestsPerSecond))
	tokens := make(chan struct{}, requestsPerSecond)

	// Fill bucket initially
	for i := 0; i < requestsPerSecond; i++ {
		tokens <- struct{}{}
	}

	// Refill tokens
	go func() {
		for range ticker.C {
			select {
			case tokens <- struct{}{}:
			default:
			}
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-tokens:
				next.ServeHTTP(w, r)
			default:
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			}
		})
	}
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := len(allowedOrigins) == 0 || contains(allowedOrigins, "*")
			if !allowed {
				for _, allowedOrigin := range allowedOrigins {
					if allowedOrigin == origin {
						allowed = true
						break
					}
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateID()
			}

			w.Header().Set("X-Request-ID", requestID)
			r.Header.Set("X-Request-ID", requestID)

			next.ServeHTTP(w, r)
		})
	}
}

// CompressionMiddleware adds gzip compression support
func CompressionMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if client accepts gzip
			if !contains(r.Header.Values("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			// For now, just pass through - full implementation would wrap writer with gzip.Writer
			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware provides basic authentication
func AuthMiddleware(username, password string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RecordingMiddleware handles traffic recording
func RecordingMiddleware(cfg *config.Config, db *sql.DB, insertStmt *sql.Stmt) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.RecordingMode {
				next.ServeHTTP(w, r)
				return
			}

			// Recording logic would go here
			// For now, just pass through
			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
