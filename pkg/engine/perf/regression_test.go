package perf

import (
	"math"
	"os"
	"testing"
	"time"
)

func TestNewRegressionDetector(t *testing.T) {
	detector := NewRegressionDetector(3.0, 100)

	if detector == nil {
		t.Fatal("NewRegressionDetector returned nil")
	}

	if detector.threshold != 3.0 {
		t.Errorf("Expected threshold 3.0, got %f", detector.threshold)
	}

	if detector.minSamples != 100 {
		t.Errorf("Expected minSamples 100, got %d", detector.minSamples)
	}

	if len(detector.baselines) != 0 {
		t.Error("Expected empty baselines initially")
	}
}

func TestCheckAnomaly(t *testing.T) {
	detector := NewRegressionDetector(2.0, 10)

	// Create a baseline
	baseline := &Baseline{
		Endpoint:      "/api/users",
		Method:        "GET",
		MeanLatency:   100 * time.Millisecond,
		StdDevLatency: 10 * time.Millisecond,
		SampleCount:   100,
	}

	detector.baselines["GET /api/users"] = baseline

	t.Run("No anomaly - normal latency", func(t *testing.T) {
		metric := PerformanceMetric{
			Timestamp: time.Now(),
			Endpoint:  "/api/users",
			Method:    "GET",
			Latency:   105 * time.Millisecond,
			Success:   true,
		}

		anomaly := detector.CheckAnomaly(metric)
		if anomaly != nil {
			t.Error("Expected no anomaly for normal latency")
		}
	})

	t.Run("Anomaly detected - high latency", func(t *testing.T) {
		metric := PerformanceMetric{
			Timestamp: time.Now(),
			Endpoint:  "/api/users",
			Method:    "GET",
			Latency:   170 * time.Millisecond, // 7 std devs above mean (threshold*3 = 6)
			Success:   true,
		}

		anomaly := detector.CheckAnomaly(metric)
		if anomaly == nil {
			t.Fatal("Expected anomaly for high latency")
		}

		if anomaly.Metric != "latency" {
			t.Error("Expected latency metric")
		}

		if anomaly.Severity != "critical" {
			t.Errorf("Expected critical severity, got %s", anomaly.Severity)
		}
	})

	t.Run("No baseline - no anomaly", func(t *testing.T) {
		metric := PerformanceMetric{
			Endpoint: "/api/posts",
			Method:   "GET",
			Latency:  1000 * time.Millisecond,
		}

		anomaly := detector.CheckAnomaly(metric)
		if anomaly != nil {
			t.Error("Expected no anomaly when baseline doesn't exist")
		}
	})
}

func TestUpdateBaseline(t *testing.T) {
	detector := NewRegressionDetector(3.0, 10)

	metrics := []PerformanceMetric{
		{Endpoint: "/api/users", Method: "GET", Latency: 100 * time.Millisecond, Success: true},
		{Endpoint: "/api/users", Method: "GET", Latency: 110 * time.Millisecond, Success: true},
		{Endpoint: "/api/users", Method: "GET", Latency: 90 * time.Millisecond, Success: true},
		{Endpoint: "/api/users", Method: "GET", Latency: 105 * time.Millisecond, Success: true},
		{Endpoint: "/api/users", Method: "GET", Latency: 95 * time.Millisecond, Success: true},
		{Endpoint: "/api/users", Method: "GET", Latency: 100 * time.Millisecond, Success: true},
		{Endpoint: "/api/users", Method: "GET", Latency: 110 * time.Millisecond, Success: true},
		{Endpoint: "/api/users", Method: "GET", Latency: 90 * time.Millisecond, Success: true},
		{Endpoint: "/api/users", Method: "GET", Latency: 105 * time.Millisecond, Success: true},
		{Endpoint: "/api/users", Method: "GET", Latency: 95 * time.Millisecond, Success: true},
	}

	detector.UpdateBaseline(metrics)

	baseline := detector.GetBaseline("GET", "/api/users")
	if baseline == nil {
		t.Fatal("Expected baseline to be created")
	}

	if baseline.SampleCount != 10 {
		t.Errorf("Expected 10 samples, got %d", baseline.SampleCount)
	}

	if baseline.SuccessRate != 100.0 {
		t.Errorf("Expected 100%% success rate, got %f", baseline.SuccessRate)
	}

	// Mean should be around 100ms
	if baseline.MeanLatency < 95*time.Millisecond || baseline.MeanLatency > 105*time.Millisecond {
		t.Errorf("Expected mean around 100ms, got %v", baseline.MeanLatency)
	}
}

func TestLoadSaveBaseline(t *testing.T) {
	detector := NewRegressionDetector(3.0, 10)

	// Create some baselines
	detector.baselines["GET /api/users"] = &Baseline{
		Endpoint:      "/api/users",
		Method:        "GET",
		MeanLatency:   100 * time.Millisecond,
		StdDevLatency: 10 * time.Millisecond,
		SampleCount:   100,
		SuccessRate:   99.5,
	}

	// Save to file
	tmpFile := "/tmp/test-baseline.json"
	defer os.Remove(tmpFile)

	err := detector.SaveBaseline(tmpFile)
	if err != nil {
		t.Fatalf("Failed to save baseline: %v", err)
	}

	// Load into new detector
	detector2 := NewRegressionDetector(3.0, 10)
	err = detector2.LoadBaseline(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load baseline: %v", err)
	}

	// Verify loaded baseline
	baseline := detector2.GetBaseline("GET", "/api/users")
	if baseline == nil {
		t.Fatal("Expected baseline to be loaded")
	}

	if baseline.MeanLatency != 100*time.Millisecond {
		t.Errorf("Expected mean 100ms, got %v", baseline.MeanLatency)
	}

	if baseline.SampleCount != 100 {
		t.Errorf("Expected 100 samples, got %d", baseline.SampleCount)
	}
}

func TestCalculateMeanStdDev(t *testing.T) {
	values := []float64{10, 20, 30, 40, 50}

	mean, stddev := calculateMeanStdDev(values)

	if mean != 30.0 {
		t.Errorf("Expected mean 30, got %f", mean)
	}

	// Standard deviation should be around 14.14
	expectedStdDev := 14.14
	if math.Abs(stddev-expectedStdDev) > 0.5 {
		t.Errorf("Expected stddev around %f, got %f", expectedStdDev, stddev)
	}

	t.Run("Empty values", func(t *testing.T) {
		mean, stddev := calculateMeanStdDev([]float64{})
		if mean != 0 || stddev != 0 {
			t.Error("Expected 0 for empty values")
		}
	})

	t.Run("Single value", func(t *testing.T) {
		mean, stddev := calculateMeanStdDev([]float64{42})
		if mean != 42 {
			t.Errorf("Expected mean 42, got %f", mean)
		}
		if stddev != 0 {
			t.Errorf("Expected stddev 0 for single value, got %f", stddev)
		}
	})
}

func TestCalculatePercentiles(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	p50, p95, p99 := calculatePercentiles(values)

	// P50 should be around 5
	if p50 < 4 || p50 > 6 {
		t.Errorf("Expected p50 around 5, got %f", p50)
	}

	// P95 should be around 9-10
	if p95 < 9 {
		t.Errorf("Expected p95 >= 9, got %f", p95)
	}

	// P99 should be close to 10
	if p99 < 9 {
		t.Errorf("Expected p99 >= 9, got %f", p99)
	}

	t.Run("Empty values", func(t *testing.T) {
		p50, p95, p99 := calculatePercentiles([]float64{})
		if p50 != 0 || p95 != 0 || p99 != 0 {
			t.Error("Expected 0 for empty values")
		}
	})

	t.Run("Single value", func(t *testing.T) {
		p50, p95, p99 := calculatePercentiles([]float64{42})
		if p50 != 42 || p95 != 42 || p99 != 42 {
			t.Error("Expected all percentiles to equal single value")
		}
	})
}

func TestCalculateBaseline(t *testing.T) {
	metrics := []PerformanceMetric{
		{Endpoint: "/api/users", Method: "GET", Latency: 100 * time.Millisecond, Success: true},
		{Endpoint: "/api/users", Method: "GET", Latency: 200 * time.Millisecond, Success: true},
		{Endpoint: "/api/users", Method: "GET", Latency: 150 * time.Millisecond, Success: false},
	}

	baseline := calculateBaseline(metrics)

	if baseline == nil {
		t.Fatal("Expected baseline, got nil")
	}

	if baseline.SampleCount != 3 {
		t.Errorf("Expected 3 samples, got %d", baseline.SampleCount)
	}

	if baseline.SuccessRate < 66 || baseline.SuccessRate > 67 {
		t.Errorf("Expected success rate ~66.67%%, got %f", baseline.SuccessRate)
	}

	// Mean should be 150ms
	if baseline.MeanLatency < 140*time.Millisecond || baseline.MeanLatency > 160*time.Millisecond {
		t.Errorf("Expected mean around 150ms, got %v", baseline.MeanLatency)
	}

	t.Run("Empty metrics", func(t *testing.T) {
		baseline := calculateBaseline([]PerformanceMetric{})
		if baseline != nil {
			t.Error("Expected nil for empty metrics")
		}
	})
}

func TestGenerateReport(t *testing.T) {
	detector := NewRegressionDetector(3.0, 10)

	// Add some baselines
	detector.baselines["GET /api/users"] = &Baseline{
		Endpoint:      "/api/users",
		Method:        "GET",
		MeanLatency:   100 * time.Millisecond,
		SampleCount:   100,
		SuccessRate:   99.5,
	}

	detector.baselines["POST /api/users"] = &Baseline{
		Endpoint:      "/api/users",
		Method:        "POST",
		MeanLatency:   200 * time.Millisecond,
		SampleCount:   50,
		SuccessRate:   98.0,
	}

	// Create some anomalies
	anomalies := []*Anomaly{
		{
			Endpoint: "/api/users",
			Method:   "GET",
			Severity: "critical",
		},
		{
			Endpoint: "/api/users",
			Method:   "POST",
			Severity: "high",
		},
	}

	report := detector.GenerateReport(anomalies)

	if report == nil {
		t.Fatal("Expected report, got nil")
	}

	if len(report.Baselines) != 2 {
		t.Errorf("Expected 2 baselines in report, got %d", len(report.Baselines))
	}

	if report.Summary.TotalEndpoints != 2 {
		t.Errorf("Expected 2 total endpoints, got %d", report.Summary.TotalEndpoints)
	}

	if report.Summary.TotalAnomalies != 2 {
		t.Errorf("Expected 2 anomalies, got %d", report.Summary.TotalAnomalies)
	}

	if report.Summary.CriticalAnomalies != 1 {
		t.Errorf("Expected 1 critical anomaly, got %d", report.Summary.CriticalAnomalies)
	}

	if report.Summary.HighAnomalies != 1 {
		t.Errorf("Expected 1 high anomaly, got %d", report.Summary.HighAnomalies)
	}

	if report.Summary.OverallHealthStatus != "critical" {
		t.Errorf("Expected critical health status, got %s", report.Summary.OverallHealthStatus)
	}

	// Average success rate should be around 98.75 ((99.5 + 98.0) / 2)
	expectedAvg := 98.75
	if math.Abs(report.Summary.AverageSuccessRate-expectedAvg) > 0.1 {
		t.Errorf("Expected average success rate around %f, got %f",
			expectedAvg, report.Summary.AverageSuccessRate)
	}
}

func TestHealthStatus(t *testing.T) {
	detector := NewRegressionDetector(3.0, 10)

	tests := []struct {
		name       string
		anomalies  []*Anomaly
		expected   string
	}{
		{
			name:      "Healthy",
			anomalies: []*Anomaly{},
			expected:  "healthy",
		},
		{
			name: "Warning",
			anomalies: []*Anomaly{
				{Severity: "medium"},
			},
			expected: "warning",
		},
		{
			name: "Degraded",
			anomalies: []*Anomaly{
				{Severity: "high"},
			},
			expected: "degraded",
		},
		{
			name: "Critical",
			anomalies: []*Anomaly{
				{Severity: "critical"},
			},
			expected: "critical",
		},
		{
			name: "Critical overrides degraded",
			anomalies: []*Anomaly{
				{Severity: "high"},
				{Severity: "critical"},
			},
			expected: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := detector.GenerateReport(tt.anomalies)
			if report.Summary.OverallHealthStatus != tt.expected {
				t.Errorf("Expected %s, got %s",
					tt.expected, report.Summary.OverallHealthStatus)
			}
		})
	}
}

func TestMinSamplesRequirement(t *testing.T) {
	detector := NewRegressionDetector(3.0, 100)

	// Create metrics but not enough samples
	metrics := []PerformanceMetric{
		{Endpoint: "/api/users", Method: "GET", Latency: 100 * time.Millisecond},
		{Endpoint: "/api/users", Method: "GET", Latency: 110 * time.Millisecond},
	}

	detector.UpdateBaseline(metrics)

	// Should not create baseline with insufficient samples
	baseline := detector.GetBaseline("GET", "/api/users")
	if baseline != nil {
		t.Error("Expected no baseline with insufficient samples")
	}
}

func TestSeverityClassification(t *testing.T) {
	detector := NewRegressionDetector(2.0, 10)

	baseline := &Baseline{
		Endpoint:      "/api/users",
		Method:        "GET",
		MeanLatency:   100 * time.Millisecond,
		StdDevLatency: 10 * time.Millisecond,
		SampleCount:   100,
	}

	detector.baselines["GET /api/users"] = baseline

	tests := []struct {
		name     string
		latency  time.Duration
		severity string
	}{
		{"Medium - 2.5 std devs", 125 * time.Millisecond, "medium"},
		{"High - 4.5 std devs", 145 * time.Millisecond, "high"},
		{"Critical - 7 std devs", 170 * time.Millisecond, "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric := PerformanceMetric{
				Endpoint: "/api/users",
				Method:   "GET",
				Latency:  tt.latency,
			}

			anomaly := detector.CheckAnomaly(metric)
			if anomaly == nil {
				t.Fatal("Expected anomaly to be detected")
			}

			if anomaly.Severity != tt.severity {
				t.Errorf("Expected severity %s, got %s", tt.severity, anomaly.Severity)
			}
		})
	}
}
