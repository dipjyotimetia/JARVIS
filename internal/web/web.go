package web

import (
	"database/sql"
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//go:embed index.html
var templateFS embed.FS

// UIHandler manages the web interface for browsing recorded transactions
type UIHandler struct {
	database *sql.DB
	tmpl     *template.Template
}

// TransactionListResponse represents the response structure for transaction listings
type TransactionListResponse struct {
	Transactions []TransactionSummary `json:"transactions"`
	Total        int                  `json:"total"`
	Page         int                  `json:"page"`
	PageSize     int                  `json:"pageSize"`
}

// TransactionSummary contains a summarized view of a transaction
type TransactionSummary struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Protocol    string    `json:"protocol"`
	Method      string    `json:"method"`
	URL         string    `json:"url"`
	Status      int       `json:"status"`
	Duration    int64     `json:"duration_ms"`
	ContentType string    `json:"content_type"`
}

// TransactionDetail contains complete transaction details
type TransactionDetail struct {
	ID                  string    `json:"id"`
	Timestamp           time.Time `json:"timestamp"`
	Protocol            string    `json:"protocol"`
	Method              string    `json:"method"`
	URL                 string    `json:"url"`
	RequestHeaders      string    `json:"request_headers"`
	RequestBody         []byte    `json:"request_body"`
	ResponseStatus      int       `json:"response_status"`
	ResponseHeaders     string    `json:"response_headers"`
	ResponseBody        []byte    `json:"response_body"`
	Duration            int64     `json:"duration_ms"`
	ClientIP            string    `json:"client_ip"`
	TestID              string    `json:"test_id"`
	SessionID           string    `json:"session_id"`
	ConnectionID        string    `json:"connection_id,omitempty"`
	MessageType         int       `json:"message_type,omitempty"`
	Direction           string    `json:"direction,omitempty"`
	ValidationError     string    `json:"validation_error,omitempty"`
	ValidationErrorType string    `json:"validation_error_type,omitempty"`
}

// NewUIHandler creates a new web interface handler
func NewUIHandler(database *sql.DB) *UIHandler {
	// Load HTML template from embedded filesystem
	tmpl := template.Must(template.ParseFS(templateFS, "index.html"))
	return &UIHandler{
		database: database,
		tmpl:     tmpl,
	}
}

// RegisterRoutes sets up the HTTP routes for the web interface
func (h *UIHandler) RegisterRoutes(mux *http.ServeMux) {
	// Static assets
	mux.HandleFunc("/ui/", h.handleIndex)

	// API endpoints
	mux.HandleFunc("/api/transactions", h.handleTransactionsList)
	mux.HandleFunc("/api/transactions/", h.handleTransactionDetail)
}

// handleIndex renders the main UI page
func (h *UIHandler) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	h.tmpl.Execute(w, nil)
}

// handleTransactionsList returns a paginated list of transactions
func (h *UIHandler) handleTransactionsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	protocol := r.URL.Query().Get("protocol")
	method := r.URL.Query().Get("method")
	url := r.URL.Query().Get("url")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize

	// Build query
	queryParams := []any{}
	query := `SELECT 
        id, timestamp, protocol, method, url, response_status, duration_ms, response_headers
        FROM traffic_records WHERE 1=1`

	if protocol != "" {
		query += " AND protocol = ?"
		queryParams = append(queryParams, protocol)
	}
	if method != "" {
		query += " AND method = ?"
		queryParams = append(queryParams, method)
	}
	if url != "" {
		query += " AND url LIKE ?"
		queryParams = append(queryParams, "%"+url+"%")
	}

	// Add count query
	countQuery := "SELECT COUNT(*) FROM traffic_records WHERE 1=1"
	if protocol != "" {
		countQuery += " AND protocol = ?"
	}
	if method != "" {
		countQuery += " AND method = ?"
	}
	if url != "" {
		countQuery += " AND url LIKE ?"
	}

	// Add pagination
	query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	queryParams = append(queryParams, pageSize, offset)

	// Get total count
	var total int
	err := h.database.QueryRow(countQuery, queryParams[:len(queryParams)-2]...).Scan(&total)
	if err != nil {
		log.Printf("🚨 Error counting transactions: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Execute query
	rows, err := h.database.Query(query, queryParams...)
	if err != nil {
		log.Printf("🚨 Error querying transactions: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Parse results
	var transactions []TransactionSummary
	for rows.Next() {
		var t TransactionSummary
		var respHeaders string
		err := rows.Scan(&t.ID, &t.Timestamp, &t.Protocol, &t.Method, &t.URL, &t.Status, &t.Duration, &respHeaders)
		if err != nil {
			log.Printf("⚠️ Error scanning transaction row: %v", err)
			continue
		}

		// Extract content-type from headers if available
		var headers map[string][]string
		if err := json.Unmarshal([]byte(respHeaders), &headers); err == nil {
			if contentTypes, ok := headers["Content-Type"]; ok && len(contentTypes) > 0 {
				t.ContentType = contentTypes[0]
			}
		}

		transactions = append(transactions, t)
	}

	// Prepare and send response
	response := TransactionListResponse{
		Transactions: transactions,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
	}

	json.NewEncoder(w).Encode(response)
}

// handleTransactionDetail returns details of a specific transaction
func (h *UIHandler) handleTransactionDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract ID from path
	path := r.URL.Path
	prefix := "/api/transactions/"
	if !strings.HasPrefix(path, prefix) {
		http.Error(w, "Invalid transaction path", http.StatusBadRequest)
		return
	}

	id := strings.TrimSuffix(path[len(prefix):], "/")
	if id == "" {
		http.Error(w, "Transaction ID is required", http.StatusBadRequest)
		return
	}

	// Query transaction details
	query := `SELECT 
        id, timestamp, protocol, method, url, request_headers, request_body,
        response_status, response_headers, response_body, duration_ms,
        client_ip, test_id, session_id, connection_id, message_type, direction
        FROM traffic_records WHERE id = ?`

	var t TransactionDetail
	err := h.database.QueryRow(query, id).Scan(
		&t.ID, &t.Timestamp, &t.Protocol, &t.Method, &t.URL, &t.RequestHeaders, &t.RequestBody,
		&t.ResponseStatus, &t.ResponseHeaders, &t.ResponseBody, &t.Duration,
		&t.ClientIP, &t.TestID, &t.SessionID, &t.ConnectionID, &t.MessageType, &t.Direction,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Transaction not found", http.StatusNotFound)
		} else {
			log.Printf("🚨 Error querying transaction details: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Extract API validation error information from headers
	var respHeaders map[string][]string
	if err := json.Unmarshal([]byte(t.ResponseHeaders), &respHeaders); err == nil {
		if validationErrorType, ok := respHeaders["X-Api-Validation-Error"]; ok && len(validationErrorType) > 0 {
			t.ValidationErrorType = validationErrorType[0]

			// Look for detailed validation error in body
			if t.ValidationErrorType == "request" {
				// Request validation errors would have been returned from the proxy to the client
				t.ValidationError = "Request failed OpenAPI schema validation"
			} else if t.ValidationErrorType == "response" {
				// Response validation errors would be logged but not visible in the response
				t.ValidationError = "Response failed OpenAPI schema validation"
			}
		}
	}

	json.NewEncoder(w).Encode(t)
}
