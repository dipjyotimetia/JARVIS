package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dipjyotimetia/jarvis/config"
)

func TestChain(t *testing.T) {
	called := []string{}

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = append(called, "middleware1")
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = append(called, "middleware2")
			next.ServeHTTP(w, r)
		})
	}

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = append(called, "handler")
		w.WriteHeader(http.StatusOK)
	})

	chained := Chain(finalHandler, middleware1, middleware2)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	chained.ServeHTTP(w, req)

	if len(called) != 3 {
		t.Errorf("Expected 3 calls, got %d", len(called))
	}

	// Middleware should be called in order: middleware1, middleware2, handler
	if called[0] != "middleware1" || called[1] != "middleware2" || called[2] != "handler" {
		t.Errorf("Unexpected call order: %v", called)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	middleware := LoggingMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	middleware := RecoveryMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic
	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestMetricsMiddleware(t *testing.T) {
	metrics := NewMetrics()
	middleware := MetricsMiddleware(metrics)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if metrics.GetTotalRequests() != 1 {
		t.Errorf("Expected 1 request, got %d", metrics.GetTotalRequests())
	}

	if metrics.GetStatusCodeCount(http.StatusOK) != 1 {
		t.Errorf("Expected 1 OK status, got %d", metrics.GetStatusCodeCount(http.StatusOK))
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	middleware := RateLimitMiddleware(10) // 10 requests per second

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	// First request should succeed
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}

	// Allow some time for token bucket
	time.Sleep(10 * time.Millisecond)
}

func TestCORSMiddleware(t *testing.T) {
	middleware := CORSMiddleware([]string{"http://example.com"})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	t.Run("Allowed origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		origin := w.Header().Get("Access-Control-Allow-Origin")
		if origin != "http://example.com" {
			t.Errorf("Expected CORS header to be set, got %s", origin)
		}
	})

	t.Run("OPTIONS request", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK for OPTIONS, got %d", w.Code)
		}
	})
}

func TestRequestIDMiddleware(t *testing.T) {
	middleware := RequestIDMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			t.Error("Request ID not set in request headers")
		}
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	t.Run("Generate new request ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		if requestID == "" {
			t.Error("Request ID not set in response headers")
		}
	})

	t.Run("Use existing request ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", "test-123")
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		if requestID != "test-123" {
			t.Errorf("Expected request ID test-123, got %s", requestID)
		}
	})
}

func TestAuthMiddleware(t *testing.T) {
	middleware := AuthMiddleware("admin", "secret")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	t.Run("Valid credentials", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.SetBasicAuth("admin", "secret")
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK with valid auth, got %d", w.Code)
		}
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.SetBasicAuth("admin", "wrong")
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 with invalid auth, got %d", w.Code)
		}
	})

	t.Run("No credentials", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		wrapped.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 with no auth, got %d", w.Code)
		}
	})
}

func TestRecordingMiddleware(t *testing.T) {
	cfg := &config.Config{
		RecordingMode: false,
	}

	middleware := RecordingMiddleware(cfg, nil, nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{"Found", []string{"a", "b", "c"}, "b", true},
		{"Not found", []string{"a", "b", "c"}, "d", false},
		{"Empty slice", []string{}, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
