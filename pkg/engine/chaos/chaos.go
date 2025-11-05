// Package chaos provides chaos engineering capabilities for API testing
package chaos

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

// FaultType represents the type of fault to inject
type FaultType string

const (
	FaultLatency    FaultType = "latency"
	FaultError      FaultType = "error"
	FaultDisconnect FaultType = "disconnect"
	FaultCorrupt    FaultType = "corrupt"
	FaultThrottle   FaultType = "throttle"
)

// FaultConfig defines chaos fault configuration
type FaultConfig struct {
	Type        FaultType     `json:"type"`
	Target      string        `json:"target"`       // Path pattern to target (e.g., "/api/*")
	Probability float64       `json:"probability"`  // 0.0 - 1.0
	Percentage  int           `json:"percentage"`   // 0-100, alternative to probability
	Duration    time.Duration `json:"duration"`     // How long to inject the fault
	Enabled     bool          `json:"enabled"`

	// Type-specific configs
	LatencyMs    int   `json:"latency_ms,omitempty"`     // For latency faults
	StatusCode   int   `json:"status_code,omitempty"`    // For error faults
	ResponseBody string `json:"response_body,omitempty"` // For error faults
	CorruptRate  float64 `json:"corrupt_rate,omitempty"`  // For corrupt faults
	ThrottleRate int   `json:"throttle_rate,omitempty"`  // Requests per second for throttle
}

// ChaosEngine manages fault injection
type ChaosEngine struct {
	faults    []*FaultConfig
	mu        sync.RWMutex
	enabled   bool
	startTime time.Time
}

// NewChaosEngine creates a new chaos engine
func NewChaosEngine() *ChaosEngine {
	return &ChaosEngine{
		faults:  make([]*FaultConfig, 0),
		enabled: true,
	}
}

// AddFault adds a fault configuration
func (ce *ChaosEngine) AddFault(fault *FaultConfig) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	if fault.Percentage > 0 && fault.Probability == 0 {
		fault.Probability = float64(fault.Percentage) / 100.0
	}

	fault.Enabled = true
	ce.faults = append(ce.faults, fault)
}

// RemoveFault removes a fault by index
func (ce *ChaosEngine) RemoveFault(index int) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	if index < 0 || index >= len(ce.faults) {
		return fmt.Errorf("invalid fault index: %d", index)
	}

	ce.faults = append(ce.faults[:index], ce.faults[index+1:]...)
	return nil
}

// ClearFaults removes all faults
func (ce *ChaosEngine) ClearFaults() {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.faults = make([]*FaultConfig, 0)
}

// Enable enables the chaos engine
func (ce *ChaosEngine) Enable() {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.enabled = true
	ce.startTime = time.Now()
}

// Disable disables the chaos engine
func (ce *ChaosEngine) Disable() {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.enabled = false
}

// IsEnabled returns whether the engine is enabled
func (ce *ChaosEngine) IsEnabled() bool {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return ce.enabled
}

// ShouldInjectFault determines if a fault should be injected for the given request
func (ce *ChaosEngine) ShouldInjectFault(path string) *FaultConfig {
	if !ce.IsEnabled() {
		return nil
	}

	ce.mu.RLock()
	defer ce.mu.RUnlock()

	for _, fault := range ce.faults {
		if !fault.Enabled {
			continue
		}

		// Check if duration expired
		if fault.Duration > 0 && time.Since(ce.startTime) > fault.Duration {
			fault.Enabled = false
			continue
		}

		// Check if path matches target pattern
		if !matchesPattern(path, fault.Target) {
			continue
		}

		// Check probability
		if rand.Float64() > fault.Probability {
			continue
		}

		return fault
	}

	return nil
}

// InjectFault injects the specified fault into the request/response
func (ce *ChaosEngine) InjectFault(w http.ResponseWriter, r *http.Request, fault *FaultConfig) bool {
	switch fault.Type {
	case FaultLatency:
		return ce.injectLatency(fault)

	case FaultError:
		return ce.injectError(w, fault)

	case FaultDisconnect:
		return ce.injectDisconnect(w)

	case FaultCorrupt:
		return ce.injectCorrupt(w, fault)

	case FaultThrottle:
		return ce.injectThrottle(fault)

	default:
		return false
	}
}

// injectLatency introduces artificial latency
func (ce *ChaosEngine) injectLatency(fault *FaultConfig) bool {
	if fault.LatencyMs > 0 {
		delay := time.Duration(fault.LatencyMs) * time.Millisecond
		time.Sleep(delay)
		return false // Continue processing
	}
	return false
}

// injectError returns an error response
func (ce *ChaosEngine) injectError(w http.ResponseWriter, fault *FaultConfig) bool {
	statusCode := fault.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	body := fault.ResponseBody
	if body == "" {
		body = "Chaos fault injected: simulated error"
	}

	w.Header().Set("X-Chaos-Fault", "error")
	http.Error(w, body, statusCode)
	return true // Stop processing
}

// injectDisconnect simulates connection drop
func (ce *ChaosEngine) injectDisconnect(w http.ResponseWriter) bool {
	// Close connection without response
	if hj, ok := w.(http.Hijacker); ok {
		conn, _, err := hj.Hijack()
		if err == nil {
			conn.Close()
			return true
		}
	}

	// Fallback: send 503
	w.Header().Set("X-Chaos-Fault", "disconnect")
	http.Error(w, "Connection dropped (chaos fault)", http.StatusServiceUnavailable)
	return true
}

// injectCorrupt corrupts the response body
func (ce *ChaosEngine) injectCorrupt(w http.ResponseWriter, fault *FaultConfig) bool {
	// Mark for downstream corruption
	w.Header().Set("X-Chaos-Fault", "corrupt")
	w.Header().Set("X-Chaos-Corrupt-Rate", fmt.Sprintf("%.2f", fault.CorruptRate))
	return false // Continue processing, corruption happens in response
}

// injectThrottle implements rate limiting
func (ce *ChaosEngine) injectThrottle(fault *FaultConfig) bool {
	if fault.ThrottleRate > 0 {
		// Simple throttling: delay based on rate
		delay := time.Second / time.Duration(fault.ThrottleRate)
		time.Sleep(delay)
	}
	return false
}

// matchesPattern checks if a path matches a pattern
func matchesPattern(path, pattern string) bool {
	// Simple wildcard matching
	if pattern == "*" {
		return true
	}

	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(path, suffix)
	}

	return path == pattern
}

// GetActiveFaults returns all active faults
func (ce *ChaosEngine) GetActiveFaults() []*FaultConfig {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	active := make([]*FaultConfig, 0)
	for _, fault := range ce.faults {
		if fault.Enabled {
			active = append(active, fault)
		}
	}

	return active
}

// Stats tracks chaos fault statistics
type Stats struct {
	TotalRequests    int64            `json:"total_requests"`
	FaultsInjected   int64            `json:"faults_injected"`
	FaultsByType     map[FaultType]int64 `json:"faults_by_type"`
	InjectionRate    float64          `json:"injection_rate"`
}

// GetStats returns chaos engine statistics
func (ce *ChaosEngine) GetStats() *Stats {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	stats := &Stats{
		FaultsByType: make(map[FaultType]int64),
	}

	// Calculate stats from active faults
	for _, fault := range ce.faults {
		if fault.Enabled {
			stats.FaultsByType[fault.Type]++
		}
	}

	return stats
}

// Preset fault configurations

// CreateLatencyFault creates a latency fault
func CreateLatencyFault(target string, latencyMs int, probability float64) *FaultConfig {
	return &FaultConfig{
		Type:        FaultLatency,
		Target:      target,
		Probability: probability,
		LatencyMs:   latencyMs,
		Enabled:     true,
	}
}

// CreateErrorFault creates an error fault
func CreateErrorFault(target string, statusCode int, probability float64) *FaultConfig {
	return &FaultConfig{
		Type:        FaultError,
		Target:      target,
		Probability: probability,
		StatusCode:  statusCode,
		Enabled:     true,
	}
}

// CreateDisconnectFault creates a disconnect fault
func CreateDisconnectFault(target string, probability float64) *FaultConfig {
	return &FaultConfig{
		Type:        FaultDisconnect,
		Target:      target,
		Probability: probability,
		Enabled:     true,
	}
}

// CreateThrottleFault creates a throttle fault
func CreateThrottleFault(target string, rps int) *FaultConfig {
	return &FaultConfig{
		Type:         FaultThrottle,
		Target:       target,
		Probability:  1.0, // Always throttle
		ThrottleRate: rps,
		Enabled:      true,
	}
}
