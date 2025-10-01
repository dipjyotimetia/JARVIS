package analyzer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dipjyotimetia/jarvis/internal/db"
	"github.com/dipjyotimetia/jarvis/pkg/engine/ollama"
)

// FailureAnalyzer analyzes API failures and provides AI-powered diagnosis
type FailureAnalyzer struct {
	aiClient ollama.Client
	database *sql.DB
}

// New creates a new failure analyzer
func New(ctx context.Context, database *sql.DB) (*FailureAnalyzer, error) {
	aiClient, err := ollama.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	return &FailureAnalyzer{
		aiClient: aiClient,
		database: database,
	}, nil
}

// FailureReport represents a failure analysis report
type FailureReport struct {
	ID              string                 `json:"id"`
	URL             string                 `json:"url"`
	Method          string                 `json:"method"`
	StatusCode      int                    `json:"status_code"`
	Timestamp       time.Time              `json:"timestamp"`
	RequestHeaders  map[string]interface{} `json:"request_headers"`
	RequestBody     string                 `json:"request_body"`
	ResponseHeaders map[string]interface{} `json:"response_headers"`
	ResponseBody    string                 `json:"response_body"`
	Duration        int64                  `json:"duration_ms"`
	Analysis        AnalysisResult         `json:"analysis"`
}

// AnalysisResult contains AI-powered analysis of a failure
type AnalysisResult struct {
	RootCause       string   `json:"root_cause"`
	ErrorCategory   string   `json:"error_category"`
	Suggestions     []string `json:"suggestions"`
	RelatedIssues   []string `json:"related_issues"`
	Severity        string   `json:"severity"`
	FixEstimate     string   `json:"fix_estimate"`
	Documentation   []string `json:"documentation"`
	ReproSteps      []string `json:"repro_steps"`
	PossibleFixes   []string `json:"possible_fixes"`
	PreventionTips  []string `json:"prevention_tips"`
}

// AnalyzeFailures analyzes all failed requests from the traffic database
func (fa *FailureAnalyzer) AnalyzeFailures(ctx context.Context, opts AnalysisOptions) ([]FailureReport, error) {
	// Query failed requests (4xx and 5xx status codes)
	query := `
		SELECT id, timestamp, protocol, method, url,
		       request_headers, request_body,
		       response_status, response_headers, response_body,
		       duration, client_ip
		FROM traffic_records
		WHERE response_status >= 400
	`

	// Add filters based on options
	args := []interface{}{}
	if opts.TimeWindow > 0 {
		query += " AND timestamp >= ?"
		args = append(args, time.Now().Add(-opts.TimeWindow))
	}
	if opts.StatusCode > 0 {
		query += " AND response_status = ?"
		args = append(args, opts.StatusCode)
	}
	if opts.Endpoint != "" {
		query += " AND url LIKE ?"
		args = append(args, "%"+opts.Endpoint+"%")
	}

	query += " ORDER BY timestamp DESC"
	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}

	rows, err := fa.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query failures: %w", err)
	}
	defer rows.Close()

	var reports []FailureReport

	for rows.Next() {
		var record db.TrafficRecord
		var reqHeaders, respHeaders string

		err := rows.Scan(
			&record.ID,
			&record.Timestamp,
			&record.Protocol,
			&record.Method,
			&record.URL,
			&reqHeaders,
			&record.RequestBody,
			&record.ResponseStatus,
			&respHeaders,
			&record.ResponseBody,
			&record.Duration,
			&record.ClientIP,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Parse headers
		var reqHeadersMap, respHeadersMap map[string]interface{}
		json.Unmarshal([]byte(reqHeaders), &reqHeadersMap)
		json.Unmarshal([]byte(respHeaders), &respHeadersMap)

		// Perform AI analysis
		analysis, err := fa.analyzeFailure(ctx, record, reqHeadersMap, respHeadersMap)
		if err != nil {
			// If AI analysis fails, provide basic analysis
			analysis = fa.basicAnalysis(record)
		}

		report := FailureReport{
			ID:              record.ID,
			URL:             record.URL,
			Method:          record.Method,
			StatusCode:      record.ResponseStatus,
			Timestamp:       record.Timestamp,
			RequestHeaders:  reqHeadersMap,
			RequestBody:     string(record.RequestBody),
			ResponseHeaders: respHeadersMap,
			ResponseBody:    string(record.ResponseBody),
			Duration:        record.Duration,
			Analysis:        analysis,
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// AnalysisOptions configures failure analysis
type AnalysisOptions struct {
	TimeWindow time.Duration // Only analyze failures within this time window
	StatusCode int           // Filter by specific status code (0 for all)
	Endpoint   string        // Filter by endpoint pattern
	Limit      int           // Maximum number of failures to analyze
}

// analyzeFailure performs AI-powered analysis of a single failure
func (fa *FailureAnalyzer) analyzeFailure(
	ctx context.Context,
	record db.TrafficRecord,
	reqHeaders, respHeaders map[string]interface{},
) (AnalysisResult, error) {
	prompt := fa.buildAnalysisPrompt(record, reqHeaders, respHeaders)

	response, err := fa.aiClient.GenerateText(ctx, prompt)
	if err != nil {
		return AnalysisResult{}, fmt.Errorf("AI analysis failed: %w", err)
	}

	// Parse AI response
	return fa.parseAnalysisResponse(response.Response)
}

// buildAnalysisPrompt creates a detailed prompt for AI analysis
func (fa *FailureAnalyzer) buildAnalysisPrompt(
	record db.TrafficRecord,
	reqHeaders, respHeaders map[string]interface{},
) string {
	var builder strings.Builder

	builder.WriteString("Analyze this API failure and provide detailed diagnosis:\n\n")
	builder.WriteString("## Request Details\n")
	builder.WriteString(fmt.Sprintf("Method: %s\n", record.Method))
	builder.WriteString(fmt.Sprintf("URL: %s\n", record.URL))
	builder.WriteString(fmt.Sprintf("Timestamp: %s\n", record.Timestamp.Format(time.RFC3339)))

	if len(reqHeaders) > 0 {
		builder.WriteString("\nRequest Headers:\n")
		headersJSON, _ := json.MarshalIndent(reqHeaders, "", "  ")
		builder.WriteString(string(headersJSON))
	}

	if len(record.RequestBody) > 0 {
		builder.WriteString("\nRequest Body:\n")
		if len(record.RequestBody) > 1000 {
			builder.WriteString(string(record.RequestBody[:1000]))
			builder.WriteString("...(truncated)")
		} else {
			builder.WriteString(string(record.RequestBody))
		}
	}

	builder.WriteString("\n\n## Response Details\n")
	builder.WriteString(fmt.Sprintf("Status Code: %d\n", record.ResponseStatus))
	builder.WriteString(fmt.Sprintf("Duration: %dms\n", record.Duration))

	if len(respHeaders) > 0 {
		builder.WriteString("\nResponse Headers:\n")
		headersJSON, _ := json.MarshalIndent(respHeaders, "", "  ")
		builder.WriteString(string(headersJSON))
	}

	if len(record.ResponseBody) > 0 {
		builder.WriteString("\nResponse Body:\n")
		if len(record.ResponseBody) > 1000 {
			builder.WriteString(string(record.ResponseBody[:1000]))
			builder.WriteString("...(truncated)")
		} else {
			builder.WriteString(string(record.ResponseBody))
		}
	}

	builder.WriteString("\n\n## Analysis Required\n")
	builder.WriteString("Provide a comprehensive analysis in JSON format with the following structure:\n")
	builder.WriteString("{\n")
	builder.WriteString("  \"root_cause\": \"Brief explanation of the root cause\",\n")
	builder.WriteString("  \"error_category\": \"Category (e.g., Authentication, Validation, Server Error, Rate Limit, etc.)\",\n")
	builder.WriteString("  \"suggestions\": [\"Actionable suggestion 1\", \"Actionable suggestion 2\"],\n")
	builder.WriteString("  \"related_issues\": [\"Common issue 1\", \"Common issue 2\"],\n")
	builder.WriteString("  \"severity\": \"Low/Medium/High/Critical\",\n")
	builder.WriteString("  \"fix_estimate\": \"Estimated time to fix (e.g., '5 minutes', '1 hour', '1 day')\",\n")
	builder.WriteString("  \"documentation\": [\"Relevant doc link or reference\"],\n")
	builder.WriteString("  \"repro_steps\": [\"Step 1\", \"Step 2\"],\n")
	builder.WriteString("  \"possible_fixes\": [\"Fix option 1\", \"Fix option 2\"],\n")
	builder.WriteString("  \"prevention_tips\": [\"Tip 1\", \"Tip 2\"]\n")
	builder.WriteString("}\n\n")
	builder.WriteString("Return ONLY the JSON object, no additional text.")

	return builder.String()
}

// parseAnalysisResponse parses AI analysis response into structured data
func (fa *FailureAnalyzer) parseAnalysisResponse(response string) (AnalysisResult, error) {
	// Clean up response
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var result AnalysisResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return AnalysisResult{}, fmt.Errorf("failed to parse analysis response: %w", err)
	}

	return result, nil
}

// basicAnalysis provides basic analysis when AI is unavailable
func (fa *FailureAnalyzer) basicAnalysis(record db.TrafficRecord) AnalysisResult {
	category := "Unknown"
	severity := "Medium"
	rootCause := fmt.Sprintf("HTTP %d error", record.ResponseStatus)

	switch {
	case record.ResponseStatus == 400:
		category = "Validation Error"
		severity = "Low"
		rootCause = "Bad request - invalid input data"
	case record.ResponseStatus == 401:
		category = "Authentication"
		severity = "High"
		rootCause = "Authentication required or failed"
	case record.ResponseStatus == 403:
		category = "Authorization"
		severity = "High"
		rootCause = "Insufficient permissions"
	case record.ResponseStatus == 404:
		category = "Not Found"
		severity = "Low"
		rootCause = "Resource not found"
	case record.ResponseStatus == 429:
		category = "Rate Limit"
		severity = "Medium"
		rootCause = "Too many requests - rate limit exceeded"
	case record.ResponseStatus >= 500:
		category = "Server Error"
		severity = "Critical"
		rootCause = "Server-side error"
	}

	return AnalysisResult{
		RootCause:     rootCause,
		ErrorCategory: category,
		Severity:      severity,
		Suggestions: []string{
			"Review the response body for detailed error information",
			"Check API documentation for correct usage",
		},
	}
}

// AnalyzeEndpoint analyzes failures for a specific endpoint
func (fa *FailureAnalyzer) AnalyzeEndpoint(ctx context.Context, endpoint string, limit int) ([]FailureReport, error) {
	opts := AnalysisOptions{
		Endpoint: endpoint,
		Limit:    limit,
	}
	return fa.AnalyzeFailures(ctx, opts)
}

// GetFailureStatistics returns statistics about failures
func (fa *FailureAnalyzer) GetFailureStatistics(ctx context.Context, timeWindow time.Duration) (FailureStats, error) {
	query := `
		SELECT
			COUNT(*) as total,
			response_status,
			COUNT(DISTINCT url) as unique_endpoints
		FROM traffic_records
		WHERE response_status >= 400
	`

	args := []interface{}{}
	if timeWindow > 0 {
		query += " AND timestamp >= ?"
		args = append(args, time.Now().Add(-timeWindow))
	}

	query += " GROUP BY response_status ORDER BY total DESC"

	rows, err := fa.database.QueryContext(ctx, query, args...)
	if err != nil {
		return FailureStats{}, fmt.Errorf("failed to get statistics: %w", err)
	}
	defer rows.Close()

	stats := FailureStats{
		ByStatusCode: make(map[int]StatusCodeStats),
	}

	for rows.Next() {
		var total, statusCode, uniqueEndpoints int
		if err := rows.Scan(&total, &statusCode, &uniqueEndpoints); err != nil {
			return stats, err
		}

		stats.Total += total
		stats.ByStatusCode[statusCode] = StatusCodeStats{
			Count:           total,
			UniqueEndpoints: uniqueEndpoints,
		}
	}

	return stats, nil
}

// FailureStats contains failure statistics
type FailureStats struct {
	Total        int                      `json:"total"`
	ByStatusCode map[int]StatusCodeStats  `json:"by_status_code"`
}

// StatusCodeStats contains statistics for a specific status code
type StatusCodeStats struct {
	Count           int `json:"count"`
	UniqueEndpoints int `json:"unique_endpoints"`
}
