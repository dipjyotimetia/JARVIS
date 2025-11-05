// Package perf provides performance analysis and regression detection
package perf

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"
)

// PerformanceMetric represents a single performance measurement
type PerformanceMetric struct {
	Timestamp  time.Time     `json:"timestamp"`
	Endpoint   string        `json:"endpoint"`
	Method     string        `json:"method"`
	Latency    time.Duration `json:"latency_ms"`
	StatusCode int           `json:"status_code"`
	Success    bool          `json:"success"`
}

// Baseline represents historical performance baseline
type Baseline struct {
	Endpoint         string        `json:"endpoint"`
	Method           string        `json:"method"`
	MeanLatency      time.Duration `json:"mean_latency_ms"`
	StdDevLatency    time.Duration `json:"stddev_latency_ms"`
	P50Latency       time.Duration `json:"p50_latency_ms"`
	P95Latency       time.Duration `json:"p95_latency_ms"`
	P99Latency       time.Duration `json:"p99_latency_ms"`
	SampleCount      int           `json:"sample_count"`
	SuccessRate      float64       `json:"success_rate"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// RegressionDetector detects performance regressions
type RegressionDetector struct {
	baselines          map[string]*Baseline // key: method + endpoint
	threshold          float64             // standard deviations for anomaly detection
	minSamples         int                 // minimum samples needed for baseline
	baselineUpdateFreq time.Duration       // how often to update baseline
}

// Anomaly represents a detected performance anomaly
type Anomaly struct {
	Timestamp   time.Time     `json:"timestamp"`
	Endpoint    string        `json:"endpoint"`
	Method      string        `json:"method"`
	Metric      string        `json:"metric"`
	Current     float64       `json:"current_value"`
	Expected    float64       `json:"expected_value"`
	Deviation   float64       `json:"deviation_percent"`
	Severity    string        `json:"severity"`
	Description string        `json:"description"`
}

// NewRegressionDetector creates a new regression detector
func NewRegressionDetector(threshold float64, minSamples int) *RegressionDetector {
	return &RegressionDetector{
		baselines:          make(map[string]*Baseline),
		threshold:          threshold,
		minSamples:         minSamples,
		baselineUpdateFreq: 24 * time.Hour,
	}
}

// LoadBaseline loads baseline data from a file
func (rd *RegressionDetector) LoadBaseline(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read baseline file: %w", err)
	}

	var baselines []*Baseline
	if err := json.Unmarshal(data, &baselines); err != nil {
		return fmt.Errorf("failed to unmarshal baseline: %w", err)
	}

	for _, baseline := range baselines {
		key := baseline.Method + " " + baseline.Endpoint
		rd.baselines[key] = baseline
	}

	return nil
}

// SaveBaseline saves baseline data to a file
func (rd *RegressionDetector) SaveBaseline(path string) error {
	baselines := make([]*Baseline, 0, len(rd.baselines))
	for _, baseline := range rd.baselines {
		baselines = append(baselines, baseline)
	}

	data, err := json.MarshalIndent(baselines, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baseline: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write baseline file: %w", err)
	}

	return nil
}

// CheckAnomaly checks if a metric represents an anomaly
func (rd *RegressionDetector) CheckAnomaly(metric PerformanceMetric) *Anomaly {
	key := metric.Method + " " + metric.Endpoint
	baseline, exists := rd.baselines[key]

	if !exists || baseline.SampleCount < rd.minSamples {
		// Not enough data for baseline
		return nil
	}

	// Check latency anomaly
	if baseline.StdDevLatency > 0 {
		zScore := float64(metric.Latency-baseline.MeanLatency) / float64(baseline.StdDevLatency)

		if math.Abs(zScore) > rd.threshold {
			severity := "medium"
			if math.Abs(zScore) > rd.threshold*2 {
				severity = "high"
			}
			if math.Abs(zScore) > rd.threshold*3 {
				severity = "critical"
			}

			deviation := (float64(metric.Latency) - float64(baseline.MeanLatency)) / float64(baseline.MeanLatency) * 100

			return &Anomaly{
				Timestamp:   metric.Timestamp,
				Endpoint:    metric.Endpoint,
				Method:      metric.Method,
				Metric:      "latency",
				Current:     float64(metric.Latency.Milliseconds()),
				Expected:    float64(baseline.MeanLatency.Milliseconds()),
				Deviation:   deviation,
				Severity:    severity,
				Description: fmt.Sprintf("Latency %.0fms is %.1f%% %s than baseline %.0fms (%.1f std devs)",
					metric.Latency.Seconds()*1000,
					math.Abs(deviation),
					map[bool]string{true: "higher", false: "lower"}[metric.Latency > baseline.MeanLatency],
					baseline.MeanLatency.Seconds()*1000,
					math.Abs(zScore)),
			}
		}
	}

	return nil
}

// UpdateBaseline updates the baseline with new metrics
func (rd *RegressionDetector) UpdateBaseline(metrics []PerformanceMetric) {
	// Group metrics by endpoint
	grouped := make(map[string][]PerformanceMetric)
	for _, metric := range metrics {
		key := metric.Method + " " + metric.Endpoint
		grouped[key] = append(grouped[key], metric)
	}

	// Calculate baseline for each endpoint
	for key, endpointMetrics := range grouped {
		if len(endpointMetrics) < rd.minSamples {
			continue
		}

		baseline := calculateBaseline(endpointMetrics)
		rd.baselines[key] = baseline
	}
}

// GetBaseline returns the baseline for an endpoint
func (rd *RegressionDetector) GetBaseline(method, endpoint string) *Baseline {
	key := method + " " + endpoint
	return rd.baselines[key]
}

// calculateBaseline calculates statistical baseline from metrics
func calculateBaseline(metrics []PerformanceMetric) *Baseline {
	if len(metrics) == 0 {
		return nil
	}

	// Extract latencies and count successes
	latencies := make([]float64, len(metrics))
	successCount := 0

	for i, m := range metrics {
		latencies[i] = float64(m.Latency.Nanoseconds())
		if m.Success {
			successCount++
		}
	}

	// Calculate statistics
	mean, stddev := calculateMeanStdDev(latencies)
	p50, p95, p99 := calculatePercentiles(latencies)

	return &Baseline{
		Endpoint:      metrics[0].Endpoint,
		Method:        metrics[0].Method,
		MeanLatency:   time.Duration(int64(mean)),
		StdDevLatency: time.Duration(int64(stddev)),
		P50Latency:    time.Duration(int64(p50)),
		P95Latency:    time.Duration(int64(p95)),
		P99Latency:    time.Duration(int64(p99)),
		SampleCount:   len(metrics),
		SuccessRate:   float64(successCount) / float64(len(metrics)) * 100,
		LastUpdated:   time.Now(),
	}
}

// calculateMeanStdDev calculates mean and standard deviation
func calculateMeanStdDev(values []float64) (mean, stddev float64) {
	if len(values) == 0 {
		return 0, 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))

	// Calculate standard deviation
	varianceSum := 0.0
	for _, v := range values {
		diff := v - mean
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(len(values))
	stddev = math.Sqrt(variance)

	return mean, stddev
}

// calculatePercentiles calculates p50, p95, p99 percentiles
func calculatePercentiles(values []float64) (p50, p95, p99 float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}

	// Sort values
	sorted := make([]float64, len(values))
	copy(sorted, values)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Calculate percentile indices
	p50Idx := int(float64(len(sorted)) * 0.50)
	p95Idx := int(float64(len(sorted)) * 0.95)
	p99Idx := int(float64(len(sorted)) * 0.99)

	// Ensure indices are within bounds
	if p50Idx >= len(sorted) {
		p50Idx = len(sorted) - 1
	}
	if p95Idx >= len(sorted) {
		p95Idx = len(sorted) - 1
	}
	if p99Idx >= len(sorted) {
		p99Idx = len(sorted) - 1
	}

	return sorted[p50Idx], sorted[p95Idx], sorted[p99Idx]
}

// Report generates a performance analysis report
type Report struct {
	Timestamp   time.Time   `json:"timestamp"`
	Baselines   []*Baseline `json:"baselines"`
	Anomalies   []*Anomaly  `json:"anomalies"`
	Summary     Summary     `json:"summary"`
}

// Summary provides a summary of the report
type Summary struct {
	TotalEndpoints      int     `json:"total_endpoints"`
	TotalAnomalies      int     `json:"total_anomalies"`
	CriticalAnomalies   int     `json:"critical_anomalies"`
	HighAnomalies       int     `json:"high_anomalies"`
	AverageSuccessRate  float64 `json:"average_success_rate"`
	OverallHealthStatus string  `json:"overall_health_status"`
}

// GenerateReport generates a comprehensive performance report
func (rd *RegressionDetector) GenerateReport(anomalies []*Anomaly) *Report {
	report := &Report{
		Timestamp: time.Now(),
		Baselines: make([]*Baseline, 0, len(rd.baselines)),
		Anomalies: anomalies,
	}

	// Collect baselines
	totalSuccessRate := 0.0
	for _, baseline := range rd.baselines {
		report.Baselines = append(report.Baselines, baseline)
		totalSuccessRate += baseline.SuccessRate
	}

	// Calculate summary
	criticalCount := 0
	highCount := 0
	for _, anomaly := range anomalies {
		if anomaly.Severity == "critical" {
			criticalCount++
		} else if anomaly.Severity == "high" {
			highCount++
		}
	}

	avgSuccessRate := 0.0
	if len(rd.baselines) > 0 {
		avgSuccessRate = totalSuccessRate / float64(len(rd.baselines))
	}

	healthStatus := "healthy"
	if criticalCount > 0 {
		healthStatus = "critical"
	} else if highCount > 0 {
		healthStatus = "degraded"
	} else if len(anomalies) > 0 {
		healthStatus = "warning"
	}

	report.Summary = Summary{
		TotalEndpoints:      len(rd.baselines),
		TotalAnomalies:      len(anomalies),
		CriticalAnomalies:   criticalCount,
		HighAnomalies:       highCount,
		AverageSuccessRate:  avgSuccessRate,
		OverallHealthStatus: healthStatus,
	}

	return report
}
