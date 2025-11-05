package chaos

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewChaosEngine(t *testing.T) {
	engine := NewChaosEngine()

	if engine == nil {
		t.Fatal("NewChaosEngine returned nil")
	}

	if !engine.IsEnabled() {
		t.Error("Expected engine to be enabled by default")
	}

	if len(engine.faults) != 0 {
		t.Error("Expected no faults initially")
	}
}

func TestAddFault(t *testing.T) {
	engine := NewChaosEngine()

	fault := &FaultConfig{
		Type:        FaultLatency,
		Target:      "/api/*",
		Probability: 0.5,
		LatencyMs:   100,
	}

	engine.AddFault(fault)

	if len(engine.faults) != 1 {
		t.Errorf("Expected 1 fault, got %d", len(engine.faults))
	}

	if !fault.Enabled {
		t.Error("Expected fault to be enabled")
	}
}

func TestRemoveFault(t *testing.T) {
	engine := NewChaosEngine()

	fault := &FaultConfig{
		Type:        FaultLatency,
		Target:      "/api/*",
		Probability: 0.5,
	}

	engine.AddFault(fault)
	err := engine.RemoveFault(0)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(engine.faults) != 0 {
		t.Errorf("Expected 0 faults, got %d", len(engine.faults))
	}
}

func TestRemoveFaultInvalidIndex(t *testing.T) {
	engine := NewChaosEngine()

	err := engine.RemoveFault(0)
	if err == nil {
		t.Error("Expected error for invalid index")
	}

	engine.AddFault(&FaultConfig{Type: FaultLatency})
	err = engine.RemoveFault(10)
	if err == nil {
		t.Error("Expected error for out of bounds index")
	}
}

func TestClearFaults(t *testing.T) {
	engine := NewChaosEngine()

	engine.AddFault(&FaultConfig{Type: FaultLatency})
	engine.AddFault(&FaultConfig{Type: FaultError})

	engine.ClearFaults()

	if len(engine.faults) != 0 {
		t.Error("Expected no faults after clear")
	}
}

func TestEnableDisable(t *testing.T) {
	engine := NewChaosEngine()

	engine.Disable()
	if engine.IsEnabled() {
		t.Error("Expected engine to be disabled")
	}

	engine.Enable()
	if !engine.IsEnabled() {
		t.Error("Expected engine to be enabled")
	}
}

func TestShouldInjectFault(t *testing.T) {
	engine := NewChaosEngine()

	fault := &FaultConfig{
		Type:        FaultLatency,
		Target:      "/api/*",
		Probability: 1.0, // Always inject
		LatencyMs:   100,
	}

	engine.AddFault(fault)

	t.Run("Matching path", func(t *testing.T) {
		result := engine.ShouldInjectFault("/api/users")
		if result == nil {
			t.Error("Expected fault to be injected for matching path")
		}
	})

	t.Run("Non-matching path", func(t *testing.T) {
		result := engine.ShouldInjectFault("/health")
		if result != nil {
			t.Error("Expected no fault for non-matching path")
		}
	})

	t.Run("Disabled engine", func(t *testing.T) {
		engine.Disable()
		result := engine.ShouldInjectFault("/api/users")
		if result != nil {
			t.Error("Expected no fault when engine is disabled")
		}
		engine.Enable()
	})
}

func TestInjectLatency(t *testing.T) {
	engine := NewChaosEngine()

	fault := &FaultConfig{
		Type:      FaultLatency,
		LatencyMs: 50,
	}

	start := time.Now()
	shouldStop := engine.injectLatency(fault)
	duration := time.Since(start)

	if shouldStop {
		t.Error("Latency fault should not stop request processing")
	}

	if duration < 45*time.Millisecond {
		t.Errorf("Expected delay of at least 45ms, got %v", duration)
	}
}

func TestInjectError(t *testing.T) {
	engine := NewChaosEngine()

	fault := &FaultConfig{
		Type:         FaultError,
		StatusCode:   500,
		ResponseBody: "Test error",
	}

	w := httptest.NewRecorder()
	shouldStop := engine.injectError(w, fault)

	if !shouldStop {
		t.Error("Error fault should stop request processing")
	}

	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	if w.Header().Get("X-Chaos-Fault") != "error" {
		t.Error("Expected X-Chaos-Fault header")
	}
}

func TestInjectDisconnect(t *testing.T) {
	engine := NewChaosEngine()

	w := httptest.NewRecorder()
	shouldStop := engine.injectDisconnect(w)

	if !shouldStop {
		t.Error("Disconnect fault should stop request processing")
	}

	// Should have set disconnect header or status
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
}

func TestInjectCorrupt(t *testing.T) {
	engine := NewChaosEngine()

	fault := &FaultConfig{
		Type:        FaultCorrupt,
		CorruptRate: 0.5,
	}

	w := httptest.NewRecorder()
	shouldStop := engine.injectCorrupt(w, fault)

	if shouldStop {
		t.Error("Corrupt fault should not stop request processing")
	}

	if w.Header().Get("X-Chaos-Fault") != "corrupt" {
		t.Error("Expected X-Chaos-Fault header for corrupt")
	}
}

func TestInjectThrottle(t *testing.T) {
	engine := NewChaosEngine()

	fault := &FaultConfig{
		Type:         FaultThrottle,
		ThrottleRate: 100, // 100 req/sec
	}

	start := time.Now()
	shouldStop := engine.injectThrottle(fault)
	duration := time.Since(start)

	if shouldStop {
		t.Error("Throttle fault should not stop request processing")
	}

	// Should have some delay, but very small (1/100 second = 10ms)
	if duration < 5*time.Millisecond {
		t.Errorf("Expected some throttling delay, got %v", duration)
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		{"Exact match", "/api/users", "/api/users", true},
		{"Wildcard all", "/api/users", "*", true},
		{"Prefix wildcard match", "/api/users", "/api/*", true},
		{"Prefix wildcard no match", "/health", "/api/*", false},
		{"Suffix wildcard match", "/api/users.json", "*.json", true},
		{"Suffix wildcard no match", "/api/users.xml", "*.json", false},
		{"No wildcard no match", "/api/users", "/api/posts", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesPattern(tt.path, tt.pattern)
			if result != tt.expected {
				t.Errorf("Expected %v for pattern %s and path %s, got %v",
					tt.expected, tt.pattern, tt.path, result)
			}
		})
	}
}

func TestGetActiveFaults(t *testing.T) {
	engine := NewChaosEngine()

	fault1 := &FaultConfig{Type: FaultLatency, Enabled: true}
	fault2 := &FaultConfig{Type: FaultError, Enabled: false}
	fault3 := &FaultConfig{Type: FaultThrottle, Enabled: true}

	engine.faults = []*FaultConfig{fault1, fault2, fault3}

	active := engine.GetActiveFaults()

	if len(active) != 2 {
		t.Errorf("Expected 2 active faults, got %d", len(active))
	}
}

func TestGetStats(t *testing.T) {
	engine := NewChaosEngine()

	engine.AddFault(&FaultConfig{Type: FaultLatency, Enabled: true})
	engine.AddFault(&FaultConfig{Type: FaultLatency, Enabled: true})
	engine.AddFault(&FaultConfig{Type: FaultError, Enabled: true})

	stats := engine.GetStats()

	if stats.FaultsByType[FaultLatency] != 2 {
		t.Errorf("Expected 2 latency faults, got %d", stats.FaultsByType[FaultLatency])
	}

	if stats.FaultsByType[FaultError] != 1 {
		t.Errorf("Expected 1 error fault, got %d", stats.FaultsByType[FaultError])
	}
}

func TestCreatePresetFaults(t *testing.T) {
	t.Run("CreateLatencyFault", func(t *testing.T) {
		fault := CreateLatencyFault("/api/*", 100, 0.5)

		if fault.Type != FaultLatency {
			t.Error("Expected latency fault type")
		}

		if fault.LatencyMs != 100 {
			t.Error("Expected latency of 100ms")
		}

		if fault.Probability != 0.5 {
			t.Error("Expected probability of 0.5")
		}
	})

	t.Run("CreateErrorFault", func(t *testing.T) {
		fault := CreateErrorFault("/api/*", 500, 0.1)

		if fault.Type != FaultError {
			t.Error("Expected error fault type")
		}

		if fault.StatusCode != 500 {
			t.Error("Expected status code 500")
		}
	})

	t.Run("CreateDisconnectFault", func(t *testing.T) {
		fault := CreateDisconnectFault("/api/*", 0.05)

		if fault.Type != FaultDisconnect {
			t.Error("Expected disconnect fault type")
		}

		if fault.Probability != 0.05 {
			t.Error("Expected probability of 0.05")
		}
	})

	t.Run("CreateThrottleFault", func(t *testing.T) {
		fault := CreateThrottleFault("/api/*", 50)

		if fault.Type != FaultThrottle {
			t.Error("Expected throttle fault type")
		}

		if fault.ThrottleRate != 50 {
			t.Error("Expected throttle rate of 50")
		}
	})
}

func TestFaultDuration(t *testing.T) {
	engine := NewChaosEngine()

	fault := &FaultConfig{
		Type:        FaultLatency,
		Target:      "*",
		Probability: 1.0,
		Duration:    50 * time.Millisecond,
	}

	engine.AddFault(fault)
	engine.Enable()

	// Should inject initially
	result := engine.ShouldInjectFault("/test")
	if result == nil {
		t.Error("Expected fault to be injected initially")
	}

	// Wait for duration to expire
	time.Sleep(60 * time.Millisecond)

	// Should not inject after duration expired
	result = engine.ShouldInjectFault("/test")
	if result != nil {
		t.Error("Expected fault to not be injected after duration expired")
	}
}

func TestPercentageConversion(t *testing.T) {
	engine := NewChaosEngine()

	fault := &FaultConfig{
		Type:       FaultLatency,
		Target:     "*",
		Percentage: 50,
	}

	engine.AddFault(fault)

	if fault.Probability != 0.5 {
		t.Errorf("Expected probability 0.5 from 50%%, got %f", fault.Probability)
	}
}
