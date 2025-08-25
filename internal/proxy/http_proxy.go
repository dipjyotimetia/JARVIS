package proxy

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dipjyotimetia/jarvis/config"
	"github.com/dipjyotimetia/jarvis/internal/db"
	"github.com/dipjyotimetia/jarvis/internal/validator"
	"github.com/google/uuid"
)

// Server interface allows for mocking in tests
type Server interface {
	Shutdown(ctx context.Context) error
}

// Buffer pools for optimization
var (
	jsonBufferPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 4096))
		},
	}
	
	recordPool = sync.Pool{
		New: func() interface{} {
			return &db.TrafficRecord{}
		},
	}
)

// Configuration constants
const (
	maxRequestSize  = 32 * 1024 * 1024 // 32MB
	streamThreshold = 1024 * 1024       // 1MB - stream bodies larger than this
)

// StartHTTPProxy starts the HTTP proxy server
func StartHTTPProxy(ctx context.Context, cfg *config.Config, db *sql.DB, insertStmt *sql.Stmt) Server {
	// Create a custom director for path-based routing
	director := func(req *http.Request) {
		// Determine target URL based on request path
		targetURLStr := cfg.GetTargetURL(req.URL.Path)

		// Parse the target URL for this request
		target, err := url.Parse(targetURLStr)
		if err != nil {
			slog.Error("Invalid target URL", "url", targetURLStr, "error", err)
			return
		}

		// Update request URL with correct scheme, host, etc. but keep the original path
		originalPath := req.URL.Path
		originalQuery := req.URL.RawQuery
		originalHost := req.Host

		// Set the scheme, host, etc. from the target
		*req.URL = *target

		// Restore original path and query
		req.URL.Path = originalPath
		req.URL.RawQuery = originalQuery

		// Set host header to target host
		req.Host = target.Host
		// Forwarding headers
		if ip, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			if prior := req.Header.Get("X-Forwarded-For"); prior != "" {
				req.Header.Set("X-Forwarded-For", prior+", "+ip)
			} else {
				req.Header.Set("X-Forwarded-For", ip)
			}
		}
		// Preserve original host in X-Forwarded-Host
		req.Header.Set("X-Forwarded-Host", originalHost)
		if req.TLS != nil {
			req.Header.Set("X-Forwarded-Proto", "https")
		} else {
			req.Header.Set("X-Forwarded-Proto", "http")
		}
	}

	// Create a custom ReverseProxy with our director
	proxy := &httputil.ReverseProxy{
		Director: director,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          200,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ResponseHeaderTimeout: 20 * time.Second,
			// Allow outbound HTTPS targets to respect TLS settings
			TLSClientConfig: cfg.GetTLSConfig(),
		},
		ErrorHandler: func(rw http.ResponseWriter, r *http.Request, err error) {
			slog.Error("HTTP proxy error", "error", err)
			rw.WriteHeader(http.StatusBadGateway)
		},
	}

	// Buffer pool for the response writer wrapper
	responseBufPool := sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	// Create handler function with all dependencies
	handler := createHTTPHandler(proxy, cfg, db, insertStmt, &responseBufPool)

	// Create the HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: http.HandlerFunc(handler),
		// Set timeouts
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Starting HTTP proxy server with path-based routing", "port", cfg.HTTPPort)
		// Log the routing table
		if len(cfg.TargetRoutes) > 0 {
			slog.Info("Routing configuration:")
			for _, route := range cfg.TargetRoutes {
				slog.Info("Route mapping", "path_prefix", route.PathPrefix, "target_url", route.TargetURL)
			}
		}
		if cfg.HTTPTargetURL != "" {
			slog.Info("Default route mapping", "target_url", cfg.HTTPTargetURL)
		}

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	return server
}

func StartHTTPSProxy(ctx context.Context, cfg *config.Config, db *sql.DB, insertStmt *sql.Stmt) Server {
	if !cfg.TLS.Enabled {
		slog.Warn("TLS is not enabled in configuration, skipping HTTPS proxy")
		return nil
	}

	// Create a custom director for path-based routing
	director := func(req *http.Request) {
		// Determine target URL based on request path
		targetURLStr := cfg.GetTargetURL(req.URL.Path)

		// Parse the target URL for this request
		target, err := url.Parse(targetURLStr)
		if err != nil {
			slog.Error("Invalid target URL", "url", targetURLStr, "error", err)
			return
		}

		// Update request URL with correct scheme, host, etc. but keep the original path
		originalPath := req.URL.Path
		originalQuery := req.URL.RawQuery
		originalHost := req.Host

		// Set the scheme, host, etc. from the target
		*req.URL = *target

		// Restore original path and query
		req.URL.Path = originalPath
		req.URL.RawQuery = originalQuery

		// Set host header to target host
		req.Host = target.Host
		// Forwarding headers
		if ip, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			if prior := req.Header.Get("X-Forwarded-For"); prior != "" {
				req.Header.Set("X-Forwarded-For", prior+", "+ip)
			} else {
				req.Header.Set("X-Forwarded-For", ip)
			}
		}
		req.Header.Set("X-Forwarded-Host", originalHost)
		if req.TLS != nil {
			req.Header.Set("X-Forwarded-Proto", "https")
		} else {
			req.Header.Set("X-Forwarded-Proto", "http")
		}
	}

	// Create a custom ReverseProxy with our director
	proxy := &httputil.ReverseProxy{
		Director: director,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          200,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ResponseHeaderTimeout: 20 * time.Second,
			// Apply TLS config for outbound connections to target servers
			TLSClientConfig: cfg.GetTLSConfig(),
		},
		ErrorHandler: func(rw http.ResponseWriter, r *http.Request, err error) {
			slog.Error("HTTPS proxy error", "error", err)
			rw.WriteHeader(http.StatusBadGateway)
		},
	}

	// Buffer pool for the response writer wrapper
	responseBufPool := sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	// Create handler function with all dependencies
	handler := createHTTPHandler(proxy, cfg, db, insertStmt, &responseBufPool)

	// Configure TLS for the server (inbound connections)
	tlsConfig := &tls.Config{}

	// Configure client certificate verification for inbound connections (mTLS)
	if cfg.TLS.ClientAuth && cfg.TLS.ClientCACert != "" {
		// Load CA certificate for client verification
		caCert, err := os.ReadFile(cfg.TLS.ClientCACert)
		if err != nil {
			slog.Error("Failed to read client CA certificate", "error", err)
		} else {
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				slog.Error("Failed to parse client CA certificate")
			} else {
				// Set client certificate verification
				tlsConfig.ClientCAs = caCertPool
				tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
				slog.Info("mTLS enabled: Client certificates will be verified")
			}
		}
	}

	// Create HTTPS server with TLS configuration
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.TLS.Port),
		Handler: http.HandlerFunc(handler),
		// Set timeouts
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		TLSConfig:    tlsConfig,
	}

	// Start HTTPS server in a goroutine
	go func() {
		slog.Info("Starting HTTPS proxy server with TLS", "port", cfg.TLS.Port)
		if tlsConfig.ClientAuth == tls.RequireAndVerifyClientCert {
			slog.Info("mTLS is enabled - client certificates will be verified")
		}

		// Log the routing table
		if len(cfg.TargetRoutes) > 0 {
			slog.Info("Routing configuration:")
			for _, route := range cfg.TargetRoutes {
				slog.Info("Route mapping", "path_prefix", route.PathPrefix, "target_url", route.TargetURL)
			}
		}
		if cfg.HTTPTargetURL != "" {
			slog.Info("Default route mapping", "target_url", cfg.HTTPTargetURL)
		}

		if err := server.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTPS server error", "error", err)
		}
	}()

	return server
}

// createHTTPHandler returns the HTTP handler function
func createHTTPHandler(
	proxy *httputil.ReverseProxy,
	cfg *config.Config,
	database *sql.DB,
	insertStmt *sql.Stmt,
	responseBufPool *sync.Pool,
) func(http.ResponseWriter, *http.Request) {
	// Initialize API validator if enabled
	var apiValidator *validator.APIValidator
	if cfg.APIValidation.Enabled {
		slog.Info("Initializing OpenAPI validator from spec", "spec_path", cfg.APIValidation.SpecPath)
		validatorOptions := validator.APIValidatorOptions{
			EnableRequestValidation:  cfg.APIValidation.ValidateRequests,
			EnableResponseValidation: cfg.APIValidation.ValidateResponses,
			StrictMode:               cfg.APIValidation.StrictMode,
		}

		var err error
		apiValidator, err = validator.NewAPIValidator(cfg.APIValidation.SpecPath, validatorOptions)
		if err != nil {
			slog.Warn("Failed to initialize OpenAPI validator", "error", err)
		} else {
			// Log API details
			apiInfo := apiValidator.GetOpenAPIInfo()
			slog.Info("Loaded OpenAPI spec", "title", apiInfo["title"], "version", apiInfo["version"])
			slog.Info("API paths loaded", "count", apiInfo["paths"])

			// Log validation modes
			slog.Info("API validation configuration", "request_validation", cfg.APIValidation.ValidateRequests, "response_validation", cfg.APIValidation.ValidateResponses)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		handleHTTPRequest(w, r, proxy, cfg, database, insertStmt, responseBufPool, apiValidator)
	}
}

// responseRecorder wrapper captures status code, headers, and body
type responseRecorder struct {
	http.ResponseWriter
	statusCode    int
	header        http.Header
	body          *bytes.Buffer
	streamMode    bool // Enable streaming mode for large responses
	maxBufferSize int64 // Maximum size to buffer
	bytesWritten  int64 // Track total bytes written
}

// Header captures headers
func (r *responseRecorder) Header() http.Header {
	// Ensure headers are initialized from the underlying writer if not already set
	if len(r.header) == 0 {
		originalHeaders := r.ResponseWriter.Header()
		for k, v := range originalHeaders {
			r.header[k] = v
		}
	}
	return r.header
}

// WriteHeader captures status code and writes header to underlying writer
func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	// Write headers captured *before* writing the status code
	for k, v := range r.header {
		r.ResponseWriter.Header()[k] = v
	}
	r.ResponseWriter.WriteHeader(statusCode)
}

// Write captures body and writes to underlying writer with streaming support
func (r *responseRecorder) Write(b []byte) (int, error) {
	// Track total bytes written
	r.bytesWritten += int64(len(b))
	
	// If we're in stream mode or would exceed buffer size, only capture limited data
	if r.streamMode || (r.maxBufferSize > 0 && r.body.Len() >= int(r.maxBufferSize)) {
		// Only capture first chunk for metadata if buffer is empty
		if r.body.Len() == 0 && len(b) > 0 {
			// Capture first 1KB for content-type detection and basic inspection
			chunkSize := len(b)
			if chunkSize > 1024 {
				chunkSize = 1024
			}
			r.body.Write(b[:chunkSize])
		}
		// Write directly to response without buffering
		return r.ResponseWriter.Write(b)
	}
	
	// Normal buffering mode - write to our buffer first
	n, err := r.body.Write(b)
	if err != nil {
		return n, err
	}
	// Then write to the original ResponseWriter
	return r.ResponseWriter.Write(b)
}

// replayHTTPTraffic serves a response from the database
func replayHTTPTraffic(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	// Consider matching on headers or body hash for more accuracy
	query := `SELECT response_status, response_headers, response_body 
              FROM traffic_records 
              WHERE protocol = 'HTTP' AND method = ? AND url = ?
              ORDER BY timestamp DESC LIMIT 1`

	row := database.QueryRow(query, r.Method, r.URL.String())

	var status int
	var headersStr string
	var respBody []byte

	err := row.Scan(&status, &headersStr, &respBody)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("No replay record found", "method", r.Method, "url", r.URL.String())
			http.Error(w, "No matching replay record found", http.StatusNotFound)
		} else {
			slog.Error("DB error during HTTP replay lookup", "method", r.Method, "url", r.URL.String(), "error", err)
			http.Error(w, "Database error during replay", http.StatusInternalServerError)
		}
		return
	}

	// Parse and set headers
	var headers http.Header
	if err := json.Unmarshal([]byte(headersStr), &headers); err != nil {
		slog.Warn("Error parsing stored headers", "method", r.Method, "url", r.URL.String(), "error", err)
		// Proceed without headers
	} else {
		for name, values := range headers {
			// Header().Set overrides, Add appends. Use Add for multi-value headers.
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}
	}

	// Set status code and write response body
	w.WriteHeader(status)
	if len(respBody) > 0 {
		_, err := w.Write(respBody)
		if err != nil {
			// Log error if writing response fails (e.g., client disconnected)
			slog.Warn("Error writing replayed HTTP response", "method", r.Method, "url", r.URL.String(), "error", err)
		}
	}
	slog.Info("Replayed HTTP response", "status", status, "method", r.Method, "url", r.URL.String())
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check common headers first (useful behind load balancers)
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For can be a list, take the first one
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return strings.TrimSpace(ip)
	}
	// Fallback to remote address
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // Return full address if split fails
	}
	return ip
}

// saveTrafficRecord saves a traffic record to SQLite
func saveTrafficRecord(record db.TrafficRecord, insertStmt *sql.Stmt) error {
	slog.Info("Attempting to save record to database", "record_id", record.ID)

	// Log record details in a structured way
	slog.Info("Record details", "method", record.Method, "url", record.URL, "status", record.ResponseStatus, "size_bytes", len(record.ResponseBody))

	// Add timeout to database operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := insertStmt.ExecContext(ctx,
		record.ID,
		record.Timestamp,
		record.Protocol,
		record.Method,
		record.URL,
		record.Service,
		record.RequestHeaders,
		record.RequestBody,
		record.ResponseStatus,
		record.ResponseHeaders,
		record.ResponseBody,
		record.Duration,
		record.ClientIP,
		record.TestID,
		record.SessionID,
		record.ConnectionID,
		record.MessageType,
		record.Direction,
	)
	if err != nil {
		return fmt.Errorf("saving record %s: %w", record.ID, err)
	}

	slog.Info("Record saved successfully", "record_id", record.ID)
	return nil
}

// generateID creates a unique ID for a traffic record
func generateID() string {
	return uuid.NewString()
}

// handleHTTPRequest contains the logic for processing each HTTP request
func handleHTTPRequest(
	w http.ResponseWriter,
	r *http.Request,
	proxy *httputil.ReverseProxy,
	cfg *config.Config,
	database *sql.DB,
	insertStmt *sql.Stmt,
	responseBufPool *sync.Pool,
	apiValidator *validator.APIValidator,
) {
	startTime := time.Now()

	// Add request size limit
	if r.ContentLength > maxRequestSize {
		http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
		return
	}

	// --- Request Handling ---
	reqHeadersBytes, _ := json.Marshal(r.Header)
	clientIP := getClientIP(r)

	// Enhanced logging in record mode
	if cfg.RecordingMode {
		slog.Info("Recording request", "method", r.Method, "url", r.URL.String(), "client_ip", clientIP)
		slog.Info("Request headers", "headers", string(reqHeadersBytes))
	}

	// Optimized request body handling with streaming support
	var reqBodyBytes []byte
	var reqBodyErr error
	var isLargeBody bool
	
	if (cfg.RecordingMode || (apiValidator != nil && cfg.APIValidation.ValidateRequests)) && r.Body != nil && r.ContentLength != 0 {
		// Check if body is too large for full buffering
		if r.ContentLength > streamThreshold {
			isLargeBody = true
			reqBodyBytes = []byte(fmt.Sprintf("<streaming-body-size:%d>", r.ContentLength))
			slog.Info("Large request body detected, using streaming mode", "size", r.ContentLength)
		} else {
			// Buffer small bodies for validation and recording
			body, errRead := io.ReadAll(io.LimitReader(r.Body, maxRequestSize))
			if errRead != nil {
				reqBodyErr = fmt.Errorf("reading request body for recording: %w", errRead)
				slog.Warn("Error reading request body", "method", r.Method, "url", r.URL.String(), "error", reqBodyErr)
			} else {
				reqBodyBytes = body
				if cfg.RecordingMode && len(reqBodyBytes) > 0 {
					// Log the request body in a readable format
					if len(reqBodyBytes) > 1024 {
						slog.Info("Request body (truncated)", "body", string(reqBodyBytes[:1024]))
					} else {
						slog.Info("Request body", "body", string(reqBodyBytes))
					}
				}
				r.Body = io.NopCloser(bytes.NewReader(reqBodyBytes))
			}
		}
	} else if r.Body != nil {
		defer r.Body.Close()
	}

	// --- API Validation for Request ---
	if apiValidator != nil && cfg.APIValidation.ValidateRequests && !isLargeBody {
		// Skip validation for large bodies to avoid memory issues
		reqCopy := r.Clone(r.Context())
		if len(reqBodyBytes) > 0 {
			reqCopy.Body = io.NopCloser(bytes.NewReader(reqBodyBytes))
		}

		if err := apiValidator.ValidateRequest(reqCopy); err != nil {
			slog.Warn("OpenAPI request validation failed", "method", r.Method, "path", r.URL.Path, "error", err)

			// If we're not continuing on validation errors, return immediately
			if !cfg.APIValidation.ContinueOnValidation {
				http.Error(w, fmt.Sprintf("Request validation error: %v", err), http.StatusBadRequest)
				return
			}

			// Add validation error header if continuing
			w.Header().Set("X-API-Validation-Error", "request")
		} else {
			slog.Info("Request passed OpenAPI validation", "method", r.Method, "path", r.URL.Path)
		}
	} else if apiValidator != nil && isLargeBody {
		slog.Info("Skipping request validation for large body", "size", r.ContentLength)
	}

	// --- Replay Mode ---
	if cfg.ReplayMode {
		replayHTTPTraffic(w, r, database)
		return
	}

	// --- Recording or Passthrough Mode ---
	var recorder *responseRecorder
	writer := w
	var needsRecording = cfg.RecordingMode
	var needsValidation = apiValidator != nil && cfg.APIValidation.ValidateResponses

	// Always use recorder if we need to validate the response or record non-large responses
	if needsRecording || needsValidation {
		responseBuf := responseBufPool.Get().(*bytes.Buffer)
		responseBuf.Reset()
		defer responseBufPool.Put(responseBuf)

		recorder = &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           responseBuf,
			header:         http.Header{},
			streamMode:     isLargeBody, // Enable streaming for large responses
			maxBufferSize:  streamThreshold,
		}
		writer = recorder

		// Log target URL in record mode
		if cfg.RecordingMode {
			targetURL := cfg.GetTargetURL(r.URL.Path)
			slog.Info("Proxying request to target", "target_url", targetURL)
		}
	}

	// Serve the request using the proxy
	proxy.ServeHTTP(writer, r)

	// --- API Validation for Response ---
	if apiValidator != nil && cfg.APIValidation.ValidateResponses && recorder != nil && !recorder.streamMode {
		// Only validate non-streaming responses
		respBody := recorder.body.Bytes()
		err := apiValidator.ValidateResponse(r, recorder.statusCode, recorder.header, respBody)
		if err != nil {
			slog.Warn("OpenAPI response validation failed", "method", r.Method, "path", r.URL.Path, "error", err)

			// If not continuing on validation errors and response isn't sent yet, return error
			if !cfg.APIValidation.ContinueOnValidation {
				// At this point the response has already been sent to the client
				// We can only log the error and add it to the record if recording
				slog.Warn("Response already sent to client, cannot return validation error")
			}

			// If the response was already sent, add validation error header for the record
			if recorder != nil {
				recorder.Header().Set("X-API-Validation-Error", "response")
			}
		} else {
			slog.Info("Response passed OpenAPI validation", "method", r.Method, "path", r.URL.Path)
		}
	} else if apiValidator != nil && recorder != nil && recorder.streamMode {
		slog.Info("Skipping response validation for streaming response")
	}

	// --- Recording (after response) ---
	if cfg.RecordingMode && recorder != nil {
		// Calculate duration
		duration := time.Since(startTime).Milliseconds()

		// Capture response details efficiently
		buf := jsonBufferPool.Get().(*bytes.Buffer)
		defer jsonBufferPool.Put(buf)
		buf.Reset()
		
		// Marshal headers using pooled buffer
		encoder := json.NewEncoder(buf)
		encoder.Encode(recorder.Header())
		respHeadersBytes := make([]byte, buf.Len())
		copy(respHeadersBytes, buf.Bytes())

		// Handle response body based on streaming mode
		var respBodyBytes []byte
		if recorder.streamMode && recorder.body.Len() > streamThreshold {
			// For large streaming responses, store metadata instead of full body
			respBodyBytes = []byte(fmt.Sprintf("<streaming-response-size:%d>", recorder.body.Len()))
			slog.Info("Large response body detected, storing metadata only", "size", recorder.body.Len())
		} else {
			respBodyBytes = recorder.body.Bytes()
		}

		// Enhanced logging for response
		slog.Info("Received response", "status", recorder.statusCode, "duration_ms", duration, "body_size", len(respBodyBytes))
		slog.Info("Response headers", "headers", string(respHeadersBytes))

		// Log response body in a readable format (truncate if too large)
		if len(respBodyBytes) > 0 && !recorder.streamMode {
			if len(respBodyBytes) > 1024 {
				slog.Info("Response body (truncated)", "body", string(respBodyBytes[:1024]))
			} else {
				slog.Info("Response body", "body", string(respBodyBytes))
			}
		} else {
			slog.Info("Response body: <empty or streaming>")
		}

		// Save the record asynchronously using pooled record
		go func() {
			recordID := generateID()
			slog.Info("Saving traffic record", "record_id", recordID)

			record := recordPool.Get().(*db.TrafficRecord)
			defer recordPool.Put(record)
			
			// Reset and populate record
			*record = db.TrafficRecord{
				ID:              recordID,
				Timestamp:       time.Now().UTC(),
				Protocol:        "HTTP",
				Method:          r.Method,
				URL:             r.URL.String(),
				RequestHeaders:  string(reqHeadersBytes),
				RequestBody:     reqBodyBytes,
				ResponseStatus:  recorder.statusCode,
				ResponseHeaders: string(respHeadersBytes),
				ResponseBody:    respBodyBytes,
				Duration:        duration,
				ClientIP:        clientIP,
				SessionID:       r.Header.Get("X-Session-ID"),
				TestID:          r.Header.Get("X-Test-ID"),
			}

			if err := saveTrafficRecord(*record, insertStmt); err != nil {
				slog.Warn("Error saving recorded HTTP traffic", "error", err)
			} else {
				slog.Info("Successfully saved record to database", "record_id", recordID)
			}
		}()
	}
}
