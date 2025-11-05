package proxy

import (
	"sync"
	"testing"
	"time"
)

func TestNewMetrics(t *testing.T) {
	metrics := NewMetrics()

	if metrics == nil {
		t.Fatal("NewMetrics returned nil")
	}

	if metrics.GetTotalRequests() != 0 {
		t.Error("Expected initial total requests to be 0")
	}

	if metrics.GetTotalErrors() != 0 {
		t.Error("Expected initial total errors to be 0")
	}
}

func TestIncrementRequests(t *testing.T) {
	metrics := NewMetrics()

	metrics.IncrementRequests()
	if metrics.GetTotalRequests() != 1 {
		t.Errorf("Expected 1 request, got %d", metrics.GetTotalRequests())
	}

	metrics.IncrementRequests()
	if metrics.GetTotalRequests() != 2 {
		t.Errorf("Expected 2 requests, got %d", metrics.GetTotalRequests())
	}
}

func TestIncrementErrors(t *testing.T) {
	metrics := NewMetrics()

	metrics.IncrementErrors()
	if metrics.GetTotalErrors() != 1 {
		t.Errorf("Expected 1 error, got %d", metrics.GetTotalErrors())
	}

	metrics.IncrementErrors()
	if metrics.GetTotalErrors() != 2 {
		t.Errorf("Expected 2 errors, got %d", metrics.GetTotalErrors())
	}
}

func TestRecordDuration(t *testing.T) {
	metrics := NewMetrics()

	// Record some durations
	metrics.RecordDuration("GET", "/api/users", 100*time.Millisecond)
	metrics.RecordDuration("GET", "/api/users", 200*time.Millisecond)
	metrics.RecordDuration("GET", "/api/users", 150*time.Millisecond)

	stats := metrics.GetDurationStats("GET", "/api/users")
	if stats == nil {
		t.Fatal("Expected duration stats, got nil")
	}

	if stats.Count() != 3 {
		t.Errorf("Expected 3 measurements, got %d", stats.Count())
	}

	// Check average
	avg := stats.Average()
	expectedAvg := 150 * time.Millisecond
	if avg < expectedAvg-10*time.Millisecond || avg > expectedAvg+10*time.Millisecond {
		t.Errorf("Expected average around %v, got %v", expectedAvg, avg)
	}

	// Check min
	min := stats.Min()
	if min != 100*time.Millisecond {
		t.Errorf("Expected min 100ms, got %v", min)
	}

	// Check max
	max := stats.Max()
	if max != 200*time.Millisecond {
		t.Errorf("Expected max 200ms, got %v", max)
	}
}

func TestRecordStatusCode(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordStatusCode(200)
	metrics.RecordStatusCode(200)
	metrics.RecordStatusCode(404)
	metrics.RecordStatusCode(500)

	if metrics.GetStatusCodeCount(200) != 2 {
		t.Errorf("Expected 2 status 200, got %d", metrics.GetStatusCodeCount(200))
	}

	if metrics.GetStatusCodeCount(404) != 1 {
		t.Errorf("Expected 1 status 404, got %d", metrics.GetStatusCodeCount(404))
	}

	if metrics.GetStatusCodeCount(500) != 1 {
		t.Errorf("Expected 1 status 500, got %d", metrics.GetStatusCodeCount(500))
	}

	if metrics.GetStatusCodeCount(301) != 0 {
		t.Errorf("Expected 0 status 301, got %d", metrics.GetStatusCodeCount(301))
	}
}

func TestGetAllStatusCodes(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordStatusCode(200)
	metrics.RecordStatusCode(404)
	metrics.RecordStatusCode(500)

	allCodes := metrics.GetAllStatusCodes()

	if len(allCodes) != 3 {
		t.Errorf("Expected 3 different status codes, got %d", len(allCodes))
	}

	if allCodes[200] != 1 {
		t.Errorf("Expected 1 status 200, got %d", allCodes[200])
	}
}

func TestGetAllDurationStats(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordDuration("GET", "/api/users", 100*time.Millisecond)
	metrics.RecordDuration("POST", "/api/users", 200*time.Millisecond)
	metrics.RecordDuration("GET", "/api/posts", 150*time.Millisecond)

	allStats := metrics.GetAllDurationStats()

	if len(allStats) != 3 {
		t.Errorf("Expected 3 different endpoints, got %d", len(allStats))
	}

	if allStats["GET /api/users"] == nil {
		t.Error("Expected stats for GET /api/users")
	}

	if allStats["POST /api/users"] == nil {
		t.Error("Expected stats for POST /api/users")
	}

	if allStats["GET /api/posts"] == nil {
		t.Error("Expected stats for GET /api/posts")
	}
}

func TestGetSummary(t *testing.T) {
	metrics := NewMetrics()

	// Record some data
	metrics.IncrementRequests()
	metrics.IncrementRequests()
	metrics.IncrementErrors()
	metrics.RecordStatusCode(200)
	metrics.RecordStatusCode(500)
	metrics.RecordDuration("GET", "/api/users", 100*time.Millisecond)

	summary := metrics.GetSummary()

	if summary.TotalRequests != 2 {
		t.Errorf("Expected 2 total requests, got %d", summary.TotalRequests)
	}

	if summary.TotalErrors != 1 {
		t.Errorf("Expected 1 total error, got %d", summary.TotalErrors)
	}

	if len(summary.StatusCodes) != 2 {
		t.Errorf("Expected 2 status codes, got %d", len(summary.StatusCodes))
	}

	if len(summary.Endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(summary.Endpoints))
	}

	endpoint := summary.Endpoints["GET /api/users"]
	if endpoint.Count != 1 {
		t.Errorf("Expected 1 count for endpoint, got %d", endpoint.Count)
	}
}

func TestReset(t *testing.T) {
	metrics := NewMetrics()

	// Add some data
	metrics.IncrementRequests()
	metrics.IncrementErrors()
	metrics.RecordStatusCode(200)
	metrics.RecordDuration("GET", "/api/users", 100*time.Millisecond)

	// Reset
	metrics.Reset()

	// Verify everything is reset
	if metrics.GetTotalRequests() != 0 {
		t.Error("Expected total requests to be 0 after reset")
	}

	if metrics.GetTotalErrors() != 0 {
		t.Error("Expected total errors to be 0 after reset")
	}

	allCodes := metrics.GetAllStatusCodes()
	if len(allCodes) != 0 {
		t.Error("Expected no status codes after reset")
	}

	allStats := metrics.GetAllDurationStats()
	if len(allStats) != 0 {
		t.Error("Expected no duration stats after reset")
	}
}

func TestConcurrentAccess(t *testing.T) {
	metrics := NewMetrics()
	iterations := 1000

	var wg sync.WaitGroup

	// Concurrent increments
	wg.Add(3)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			metrics.IncrementRequests()
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			metrics.RecordStatusCode(200)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			metrics.RecordDuration("GET", "/test", time.Millisecond)
		}
	}()

	wg.Wait()

	// Verify counts
	if metrics.GetTotalRequests() != uint64(iterations) {
		t.Errorf("Expected %d requests, got %d", iterations, metrics.GetTotalRequests())
	}

	if metrics.GetStatusCodeCount(200) != uint64(iterations) {
		t.Errorf("Expected %d status codes, got %d", iterations, metrics.GetStatusCodeCount(200))
	}

	stats := metrics.GetDurationStats("GET", "/test")
	if stats.Count() != uint64(iterations) {
		t.Errorf("Expected %d duration records, got %d", iterations, stats.Count())
	}
}

func TestDurationStatsEdgeCases(t *testing.T) {
	metrics := NewMetrics()

	t.Run("No data", func(t *testing.T) {
		stats := metrics.GetDurationStats("GET", "/nonexistent")
		if stats != nil {
			t.Error("Expected nil for nonexistent endpoint")
		}
	})

	t.Run("Single measurement", func(t *testing.T) {
		metrics.RecordDuration("GET", "/single", 100*time.Millisecond)
		stats := metrics.GetDurationStats("GET", "/single")

		if stats.Count() != 1 {
			t.Error("Expected count of 1")
		}

		if stats.Average() != 100*time.Millisecond {
			t.Error("Expected average to equal single measurement")
		}

		if stats.Min() != stats.Max() {
			t.Error("Expected min and max to be equal for single measurement")
		}
	})

	t.Run("Zero duration", func(t *testing.T) {
		metrics.RecordDuration("GET", "/zero", 0)
		stats := metrics.GetDurationStats("GET", "/zero")

		if stats.Min() != 0 {
			t.Error("Expected min to be 0")
		}

		if stats.Max() != 0 {
			t.Error("Expected max to be 0")
		}
	})
}
