package proxy

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics holds runtime metrics for the proxy
type Metrics struct {
	totalRequests   atomic.Uint64
	totalErrors     atomic.Uint64
	requestDurations sync.Map // map[string]*DurationStats
	statusCodes     sync.Map // map[int]*atomic.Uint64
	mu              sync.RWMutex
}

// DurationStats holds duration statistics
type DurationStats struct {
	count    atomic.Uint64
	total    atomic.Int64 // Total duration in nanoseconds
	min      atomic.Int64
	max      atomic.Int64
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{}
}

// IncrementRequests increments the total request counter
func (m *Metrics) IncrementRequests() {
	m.totalRequests.Add(1)
}

// IncrementErrors increments the error counter
func (m *Metrics) IncrementErrors() {
	m.totalErrors.Add(1)
}

// RecordDuration records request duration for a given method and path
func (m *Metrics) RecordDuration(method, path string, duration time.Duration) {
	key := method + " " + path

	statsInterface, _ := m.requestDurations.LoadOrStore(key, &DurationStats{})
	stats := statsInterface.(*DurationStats)

	nanos := duration.Nanoseconds()
	stats.count.Add(1)
	stats.total.Add(nanos)

	// Update min
	for {
		current := stats.min.Load()
		if current == 0 || nanos < current {
			if stats.min.CompareAndSwap(current, nanos) {
				break
			}
		} else {
			break
		}
	}

	// Update max
	for {
		current := stats.max.Load()
		if nanos > current {
			if stats.max.CompareAndSwap(current, nanos) {
				break
			}
		} else {
			break
		}
	}
}

// RecordStatusCode records a status code occurrence
func (m *Metrics) RecordStatusCode(statusCode int) {
	counterInterface, _ := m.statusCodes.LoadOrStore(statusCode, &atomic.Uint64{})
	counter := counterInterface.(*atomic.Uint64)
	counter.Add(1)
}

// GetTotalRequests returns the total number of requests
func (m *Metrics) GetTotalRequests() uint64 {
	return m.totalRequests.Load()
}

// GetTotalErrors returns the total number of errors
func (m *Metrics) GetTotalErrors() uint64 {
	return m.totalErrors.Load()
}

// GetDurationStats returns duration statistics for a given endpoint
func (m *Metrics) GetDurationStats(method, path string) *DurationStats {
	key := method + " " + path
	statsInterface, ok := m.requestDurations.Load(key)
	if !ok {
		return nil
	}
	return statsInterface.(*DurationStats)
}

// GetStatusCodeCount returns the count for a specific status code
func (m *Metrics) GetStatusCodeCount(statusCode int) uint64 {
	counterInterface, ok := m.statusCodes.Load(statusCode)
	if !ok {
		return 0
	}
	return counterInterface.(*atomic.Uint64).Load()
}

// GetAllStatusCodes returns all recorded status codes and their counts
func (m *Metrics) GetAllStatusCodes() map[int]uint64 {
	result := make(map[int]uint64)
	m.statusCodes.Range(func(key, value interface{}) bool {
		statusCode := key.(int)
		counter := value.(*atomic.Uint64)
		result[statusCode] = counter.Load()
		return true
	})
	return result
}

// GetAllDurationStats returns all duration statistics
func (m *Metrics) GetAllDurationStats() map[string]*DurationStats {
	result := make(map[string]*DurationStats)
	m.requestDurations.Range(func(key, value interface{}) bool {
		endpoint := key.(string)
		stats := value.(*DurationStats)
		result[endpoint] = stats
		return true
	})
	return result
}

// Count returns the number of recorded durations
func (ds *DurationStats) Count() uint64 {
	return ds.count.Load()
}

// Average returns the average duration
func (ds *DurationStats) Average() time.Duration {
	count := ds.count.Load()
	if count == 0 {
		return 0
	}
	total := ds.total.Load()
	return time.Duration(total / int64(count))
}

// Min returns the minimum duration
func (ds *DurationStats) Min() time.Duration {
	return time.Duration(ds.min.Load())
}

// Max returns the maximum duration
func (ds *DurationStats) Max() time.Duration {
	return time.Duration(ds.max.Load())
}

// Total returns the total duration
func (ds *DurationStats) Total() time.Duration {
	return time.Duration(ds.total.Load())
}

// MetricsSummary returns a summary of all metrics
type MetricsSummary struct {
	TotalRequests uint64                  `json:"total_requests"`
	TotalErrors   uint64                  `json:"total_errors"`
	StatusCodes   map[int]uint64          `json:"status_codes"`
	Endpoints     map[string]EndpointStat `json:"endpoints"`
}

// EndpointStat represents statistics for a single endpoint
type EndpointStat struct {
	Count      uint64        `json:"count"`
	AvgLatency time.Duration `json:"avg_latency_ms"`
	MinLatency time.Duration `json:"min_latency_ms"`
	MaxLatency time.Duration `json:"max_latency_ms"`
}

// GetSummary returns a complete metrics summary
func (m *Metrics) GetSummary() *MetricsSummary {
	summary := &MetricsSummary{
		TotalRequests: m.GetTotalRequests(),
		TotalErrors:   m.GetTotalErrors(),
		StatusCodes:   m.GetAllStatusCodes(),
		Endpoints:     make(map[string]EndpointStat),
	}

	durationStats := m.GetAllDurationStats()
	for endpoint, stats := range durationStats {
		summary.Endpoints[endpoint] = EndpointStat{
			Count:      stats.Count(),
			AvgLatency: stats.Average(),
			MinLatency: stats.Min(),
			MaxLatency: stats.Max(),
		}
	}

	return summary
}

// Reset clears all metrics
func (m *Metrics) Reset() {
	m.totalRequests.Store(0)
	m.totalErrors.Store(0)
	m.requestDurations = sync.Map{}
	m.statusCodes = sync.Map{}
}
