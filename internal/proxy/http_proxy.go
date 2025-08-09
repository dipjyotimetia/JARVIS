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
	"log"
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

// (removed custom []byte buffer pool used for request cloning to avoid unsafe aliasing)

// StartHTTPProxy starts the HTTP proxy server
func StartHTTPProxy(ctx context.Context, cfg *config.Config, db *sql.DB, insertStmt *sql.Stmt) Server {
	// Create a custom director for path-based routing
	director := func(req *http.Request) {
		// Determine target URL based on request path
		targetURLStr := cfg.GetTargetURL(req.URL.Path)

		// Parse the target URL for this request
		target, err := url.Parse(targetURLStr)
		if err != nil {
			log.Printf("🚨 ERROR: Invalid target URL %s: %v", targetURLStr, err)
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
			log.Printf("🚨 HTTP proxy error: %v", err)
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
		log.Printf("🚀 Starting HTTP proxy server on port %d with path-based routing", cfg.HTTPPort)
		// Log the routing table
		if len(cfg.TargetRoutes) > 0 {
			log.Println("📝 Routing configuration:")
			for _, route := range cfg.TargetRoutes {
				log.Printf("  %s -> %s", route.PathPrefix, route.TargetURL)
			}
		}
		if cfg.HTTPTargetURL != "" {
			log.Printf("  default -> %s", cfg.HTTPTargetURL)
		}

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("🚨 HTTP server error: %v", err)
		}
	}()

	return server
}

func StartHTTPSProxy(ctx context.Context, cfg *config.Config, db *sql.DB, insertStmt *sql.Stmt) Server {
	if !cfg.TLS.Enabled {
		log.Println("⚠️ TLS is not enabled in configuration, skipping HTTPS proxy")
		return nil
	}

	// Create a custom director for path-based routing
	director := func(req *http.Request) {
		// Determine target URL based on request path
		targetURLStr := cfg.GetTargetURL(req.URL.Path)

		// Parse the target URL for this request
		target, err := url.Parse(targetURLStr)
		if err != nil {
			log.Printf("🚨 ERROR: Invalid target URL %s: %v", targetURLStr, err)
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
			log.Printf("🚨 HTTPS proxy error: %v", err)
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
			log.Printf("🚨 Failed to read client CA certificate: %v", err)
		} else {
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				log.Printf("🚨 Failed to parse client CA certificate")
			} else {
				// Set client certificate verification
				tlsConfig.ClientCAs = caCertPool
				tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
				log.Printf("🔒 mTLS enabled: Client certificates will be verified")
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
		log.Printf("🚀 Starting HTTPS proxy server on port %d with TLS", cfg.TLS.Port)
		if tlsConfig.ClientAuth == tls.RequireAndVerifyClientCert {
			log.Printf("🔒 mTLS is enabled - client certificates will be verified")
		}

		// Log the routing table
		if len(cfg.TargetRoutes) > 0 {
			log.Println("📝 Routing configuration:")
			for _, route := range cfg.TargetRoutes {
				log.Printf("  %s -> %s", route.PathPrefix, route.TargetURL)
			}
		}
		if cfg.HTTPTargetURL != "" {
			log.Printf("  default -> %s", cfg.HTTPTargetURL)
		}

		if err := server.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("🚨 HTTPS server error: %v", err)
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
		log.Printf("🔍 Initializing OpenAPI validator from spec: %s", cfg.APIValidation.SpecPath)
		validatorOptions := validator.APIValidatorOptions{
			EnableRequestValidation:  cfg.APIValidation.ValidateRequests,
			EnableResponseValidation: cfg.APIValidation.ValidateResponses,
			StrictMode:               cfg.APIValidation.StrictMode,
		}

		var err error
		apiValidator, err = validator.NewAPIValidator(cfg.APIValidation.SpecPath, validatorOptions)
		if err != nil {
			log.Printf("⚠️ Failed to initialize OpenAPI validator: %v", err)
		} else {
			// Log API details
			apiInfo := apiValidator.GetOpenAPIInfo()
			log.Printf("🔍 Loaded OpenAPI spec: %s v%s", apiInfo["title"], apiInfo["version"])
			log.Printf("🔍 Number of API paths: %s", apiInfo["paths"])

			// Log validation modes
			log.Printf("🔍 Request validation: %v, Response validation: %v",
				cfg.APIValidation.ValidateRequests, cfg.APIValidation.ValidateResponses)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		handleHTTPRequest(w, r, proxy, cfg, database, insertStmt, responseBufPool, apiValidator)
	}
}

// responseRecorder wrapper captures status code, headers, and body
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	header     http.Header
	body       *bytes.Buffer
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

// Write captures body and writes to underlying writer
func (r *responseRecorder) Write(b []byte) (int, error) {
	// Write to our buffer first
	n, err := r.body.Write(b)
	if err != nil {
		return n, err // Return error from buffer write if any
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
			log.Printf("🕵️ No replay record found for HTTP %s %s", r.Method, r.URL.String())
			http.Error(w, "No matching replay record found", http.StatusNotFound)
		} else {
			log.Printf("🚨 DB error during HTTP replay lookup for %s %s: %v", r.Method, r.URL.String(), err)
			http.Error(w, "Database error during replay", http.StatusInternalServerError)
		}
		return
	}

	// Parse and set headers
	var headers http.Header
	if err := json.Unmarshal([]byte(headersStr), &headers); err != nil {
		log.Printf("⚠️ Error parsing stored headers for HTTP %s %s: %v", r.Method, r.URL.String(), err)
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
			log.Printf("⚠️ Error writing replayed HTTP response for %s %s: %v", r.Method, r.URL.String(), err)
		}
	}
	log.Printf("🔁 Replayed HTTP %d for %s %s", status, r.Method, r.URL.String())
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
	log.Printf("💾 Attempting to save record %s to database...", record.ID)

	// Log record details in a structured way
	log.Printf("📊 Record details: Method=%s, URL=%s, Status=%d, Size=%d bytes",
		record.Method, record.URL, record.ResponseStatus, len(record.ResponseBody))

	_, err := insertStmt.Exec(
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

	log.Printf("✅ Record %s saved successfully", record.ID)
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

	// --- Request Handling ---
	reqHeadersBytes, _ := json.Marshal(r.Header)
	clientIP := getClientIP(r)

	// Enhanced logging in record mode
	if cfg.RecordingMode {
		log.Printf("📥 Recording request: %s %s from %s", r.Method, r.URL.String(), clientIP)
		log.Printf("📋 Request headers: %s", string(reqHeadersBytes))
	}

	// Clone request body if needed for recording/replay matching later
	var reqBodyBytes []byte
	var reqBodyErr error
	if (cfg.RecordingMode || (apiValidator != nil && cfg.APIValidation.ValidateRequests)) && r.Body != nil && r.ContentLength != 0 {
		body, errRead := io.ReadAll(r.Body)
		if errRead != nil {
			reqBodyErr = fmt.Errorf("reading request body for recording: %w", errRead)
			log.Printf("⚠️ Error cloning request body for %s %s: %v", r.Method, r.URL.String(), reqBodyErr)
		} else {
			reqBodyBytes = body
			if cfg.RecordingMode && len(reqBodyBytes) > 0 {
				// Log the request body in a readable format
				if len(reqBodyBytes) > 1024 {
					log.Printf("📄 Request body (truncated): %s...", string(reqBodyBytes[:1024]))
				} else {
					log.Printf("📄 Request body: %s", string(reqBodyBytes))
				}
			}
			r.Body = io.NopCloser(bytes.NewReader(reqBodyBytes))
		}
	} else if r.Body != nil {
		defer r.Body.Close()
	}

	// --- API Validation for Request ---
	if apiValidator != nil && cfg.APIValidation.ValidateRequests {
		// Create a copy of the request to avoid modifying the original
		reqCopy := r.Clone(r.Context())
		if len(reqBodyBytes) > 0 {
			reqCopy.Body = io.NopCloser(bytes.NewReader(reqBodyBytes))
		}

		if err := apiValidator.ValidateRequest(reqCopy); err != nil {
			log.Printf("⚠️ OpenAPI request validation failed for %s %s: %v", r.Method, r.URL.Path, err)

			// If we're not continuing on validation errors, return immediately
			if !cfg.APIValidation.ContinueOnValidation {
				http.Error(w, fmt.Sprintf("Request validation error: %v", err), http.StatusBadRequest)
				return
			}

			// Add validation error header if continuing
			w.Header().Set("X-API-Validation-Error", "request")
		} else {
			log.Printf("✅ Request passed OpenAPI validation for %s %s", r.Method, r.URL.Path)
		}
	}

	// --- Replay Mode ---
	if cfg.ReplayMode {
		replayHTTPTraffic(w, r, database)
		return
	}

	// --- Recording or Passthrough Mode ---
	var recorder *responseRecorder
	writer := w

	// Always use recorder if we need to validate the response
	if cfg.RecordingMode || (apiValidator != nil && cfg.APIValidation.ValidateResponses) {
		responseBuf := responseBufPool.Get().(*bytes.Buffer)
		responseBuf.Reset()
		defer responseBufPool.Put(responseBuf)

		recorder = &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           responseBuf,
			header:         http.Header{},
		}
		writer = recorder

		// Log target URL in record mode
		if cfg.RecordingMode {
			targetURL := cfg.GetTargetURL(r.URL.Path)
			log.Printf("🔄 Proxying request to target: %s", targetURL)
		}
	}

	// Serve the request using the proxy
	proxy.ServeHTTP(writer, r)

	// --- API Validation for Response ---
	if apiValidator != nil && cfg.APIValidation.ValidateResponses && recorder != nil {
		// Validate response
		respBody := recorder.body.Bytes()
		err := apiValidator.ValidateResponse(r, recorder.statusCode, recorder.header, respBody)
		if err != nil {
			log.Printf("⚠️ OpenAPI response validation failed for %s %s: %v", r.Method, r.URL.Path, err)

			// If not continuing on validation errors and response isn't sent yet, return error
			if !cfg.APIValidation.ContinueOnValidation {
				// At this point the response has already been sent to the client
				// We can only log the error and add it to the record if recording
				log.Printf("⚠️ Response already sent to client, cannot return validation error")
			}

			// If the response was already sent, add validation error header for the record
			if recorder != nil {
				recorder.Header().Set("X-API-Validation-Error", "response")
			}
		} else {
			log.Printf("✅ Response passed OpenAPI validation for %s %s", r.Method, r.URL.Path)
		}
	}

	// --- Recording (after response) ---
	if cfg.RecordingMode && recorder != nil {
		// Calculate duration
		duration := time.Since(startTime).Milliseconds()

		// Capture response details
		respHeadersBytes, _ := json.Marshal(recorder.Header())
		respBodyBytes := recorder.body.Bytes()

		// Enhanced logging for response
		log.Printf("📤 Received response with status: %d in %d ms", recorder.statusCode, duration)
		log.Printf("📋 Response headers: %s", string(respHeadersBytes))

		// Log response body in a readable format (truncate if too large)
		if len(respBodyBytes) > 0 {
			if len(respBodyBytes) > 1024 {
				log.Printf("📄 Response body (truncated): %s...", string(respBodyBytes[:1024]))
			} else {
				log.Printf("📄 Response body: %s", string(respBodyBytes))
			}
		} else {
			log.Printf("📄 Response body: <empty>")
		}

		// Save the record asynchronously
		go func() {
			recordID := generateID()
			log.Printf("💾 Saving traffic record ID: %s", recordID)

			record := db.TrafficRecord{
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

			if err := saveTrafficRecord(record, insertStmt); err != nil {
				log.Printf("⚠️ Error saving recorded HTTP traffic: %v", err)
			} else {
				log.Printf("✅ Successfully saved record %s to database", recordID)
			}
		}()
	}
}
