package web

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// setupTestDB creates a temporary SQLite database for testing
func setupTestDB(t *testing.T) (*sql.DB, string) {
	// Create a temporary file for the database
	tempFile, err := os.CreateTemp("", "traffic_inspector_web_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	tempFile.Close()

	dbPath := tempFile.Name()

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		os.Remove(dbPath)
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS traffic_records (
		id TEXT PRIMARY KEY,
		timestamp TIMESTAMP,
		protocol TEXT,
		method TEXT,
		url TEXT,
		service TEXT,
		request_headers TEXT,
		request_body BLOB,
		response_status INTEGER,
		response_headers TEXT,
		response_body BLOB,
		duration_ms INTEGER,
		client_ip TEXT,
		test_id TEXT,
		session_id TEXT,
		connection_id TEXT,
		message_type INTEGER,
		direction TEXT
	)`)
	if err != nil {
		db.Close()
		os.Remove(dbPath)
		t.Fatalf("Failed to create table: %v", err)
	}

	// Add some test data
	insertTestData(t, db)

	return db, dbPath
}

// insertTestData adds sample records to the test database
func insertTestData(t *testing.T, db *sql.DB) {
	records := []struct {
		id              string
		timestamp       time.Time
		protocol        string
		method          string
		url             string
		requestHeaders  string
		requestBody     []byte
		responseStatus  int
		responseHeaders string
		responseBody    []byte
		duration        int64
		clientIP        string
		testID          string
		sessionID       string
		connectionID    string
		messageType     int
		direction       string
	}{
		{
			id:              "http-1",
			timestamp:       time.Now().Add(-2 * time.Hour),
			protocol:        "HTTP",
			method:          "GET",
			url:             "https://example.com/api/users",
			requestHeaders:  `{"Accept": ["application/json"]}`,
			requestBody:     nil,
			responseStatus:  200,
			responseHeaders: `{"Content-Type": ["application/json"]}`,
			responseBody:    []byte(`{"users": ["user1", "user2"]}`),
			duration:        150,
			clientIP:        "192.168.1.1",
			testID:          "test-1",
			sessionID:       "session-1",
		},
		{
			id:              "http-2",
			timestamp:       time.Now().Add(-1 * time.Hour),
			protocol:        "HTTP",
			method:          "POST",
			url:             "https://example.com/api/users",
			requestHeaders:  `{"Content-Type": ["application/json"]}`,
			requestBody:     []byte(`{"name": "New User"}`),
			responseStatus:  201,
			responseHeaders: `{"Content-Type": ["application/json"]}`,
			responseBody:    []byte(`{"id": "user3", "name": "New User"}`),
			duration:        120,
			clientIP:        "192.168.1.2",
			testID:          "test-2",
			sessionID:       "session-2",
		},
		{
			id:              "ws-1",
			timestamp:       time.Now().Add(-30 * time.Minute),
			protocol:        "WebSocket",
			method:          "CONNECT",
			url:             "wss://example.com/ws",
			requestHeaders:  `{"Upgrade": ["websocket"]}`,
			requestBody:     nil,
			responseStatus:  101,
			responseHeaders: `{"Upgrade": ["websocket"]}`,
			responseBody:    nil,
			duration:        50,
			clientIP:        "192.168.1.3",
			testID:          "test-3",
			sessionID:       "session-3",
			connectionID:    "conn-1",
		},
		{
			id:             "ws-2",
			timestamp:      time.Now().Add(-20 * time.Minute),
			protocol:       "WebSocket",
			method:         "MESSAGE",
			url:            "wss://example.com/ws",
			requestBody:    []byte(`{"type": "message", "text": "Hello"}`),
			responseStatus: 200,
			responseBody:   nil,
			duration:       25,
			clientIP:       "192.168.1.3",
			testID:         "test-3",
			sessionID:      "session-3",
			connectionID:   "conn-1",
			messageType:    1,
			direction:      "outbound",
		},
	}

	for _, r := range records {
		_, err := db.Exec(
			`INSERT INTO traffic_records 
			(id, timestamp, protocol, method, url, service, request_headers, request_body, 
			 response_status, response_headers, response_body, duration_ms, 
			 client_ip, test_id, session_id, connection_id, message_type, direction)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			r.id, r.timestamp, r.protocol, r.method, r.url, "", r.requestHeaders, r.requestBody,
			r.responseStatus, r.responseHeaders, r.responseBody, r.duration,
			r.clientIP, r.testID, r.sessionID, r.connectionID, r.messageType, r.direction,
		)
		if err != nil {
			t.Fatalf("Failed to insert test record: %v", err)
		}
	}
}

// cleanupTestDB closes the database connection and removes the temporary file
func cleanupTestDB(db *sql.DB, path string) {
	db.Close()
	os.Remove(path)
}

func TestNewUIHandler(t *testing.T) {
	db, dbPath := setupTestDB(t)
	defer cleanupTestDB(db, dbPath)

	handler := NewUIHandler(db)

	if handler == nil {
		t.Fatal("NewUIHandler returned nil")
	}

	if handler.database != db {
		t.Error("Database reference not properly stored in UIHandler")
	}

	if handler.tmpl == nil {
		t.Error("Template not properly initialized in UIHandler")
	}
}

func TestRegisterRoutes(t *testing.T) {
	db, dbPath := setupTestDB(t)
	defer cleanupTestDB(db, dbPath)

	handler := NewUIHandler(db)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Create test server with our handler
	server := httptest.NewServer(mux)
	defer server.Close()

	// Test routes registration by making actual requests
	testCases := []struct {
		path       string
		statusCode int
	}{
		{"/ui/", http.StatusOK},
		{"/api/transactions", http.StatusOK},
		{"/api/transactions/http-1", http.StatusOK},
		{"/api/transactions/nonexistent", http.StatusNotFound},
	}

	for _, tc := range testCases {
		resp, err := http.Get(server.URL + tc.path)
		if err != nil {
			t.Errorf("Request to %s failed: %v", tc.path, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != tc.statusCode {
			t.Errorf("Path %s returned status %d, expected %d", tc.path, resp.StatusCode, tc.statusCode)
		}
	}
}

func TestHandleIndex(t *testing.T) {
	db, dbPath := setupTestDB(t)
	defer cleanupTestDB(db, dbPath)

	handler := NewUIHandler(db)

	// Create a request to test the handler
	req, err := http.NewRequest("GET", "/ui/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a recorder to capture the response
	rr := httptest.NewRecorder()
	handlerFunc := http.HandlerFunc(handler.handleIndex)
	handlerFunc.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handleIndex returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("handleIndex returned wrong content type: got %v want %v", contentType, "text/html")
	}

	// Check that response contains HTML
	if len(rr.Body.String()) == 0 {
		t.Error("handleIndex returned empty body")
	}
}

func TestHandleTransactionsList(t *testing.T) {
	db, dbPath := setupTestDB(t)
	defer cleanupTestDB(db, dbPath)

	handler := NewUIHandler(db)

	// Test cases for different query parameters
	tests := []struct {
		name          string
		queryParams   string
		expectedCount int
	}{
		{
			name:          "All transactions",
			queryParams:   "",
			expectedCount: 4, // All records
		},
		{
			name:          "Filter by HTTP protocol",
			queryParams:   "protocol=HTTP",
			expectedCount: 2, // Two HTTP records
		},
		{
			name:          "Filter by WebSocket protocol",
			queryParams:   "protocol=WebSocket",
			expectedCount: 2, // Two WebSocket records
		},
		{
			name:          "Filter by method",
			queryParams:   "method=GET",
			expectedCount: 1, // One GET record
		},
		{
			name:          "Filter by URL",
			queryParams:   "url=users",
			expectedCount: 2, // Two records with 'users' in URL
		},
		{
			name:          "Pagination - page 1, size 2",
			queryParams:   "page=1&pageSize=2",
			expectedCount: 2, // First page, 2 items
		},
		{
			name:          "Pagination - page 2, size 2",
			queryParams:   "page=2&pageSize=2",
			expectedCount: 2, // Second page, 2 items
		},
		{
			name:          "Pagination - page 3, size 2",
			queryParams:   "page=3&pageSize=2",
			expectedCount: 0, // Third page, no items
		},
		{
			name:          "Complex filter",
			queryParams:   "protocol=HTTP&method=POST&pageSize=10",
			expectedCount: 1, // One HTTP POST record
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with query parameters
			req, err := http.NewRequest("GET", "/api/transactions?"+tt.queryParams, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create recorder and call handler
			rr := httptest.NewRecorder()
			handler.handleTransactionsList(rr, req)

			// Check status code
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
			}

			// Parse response
			var response TransactionListResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			// Check transaction count
			if len(response.Transactions) != tt.expectedCount {
				t.Errorf("Expected %d transactions, got %d", tt.expectedCount, len(response.Transactions))
			}

			// Verify page and pageSize are correctly reflected
			page := 1
			pageSize := 50
			if p, err := strconv.Atoi(req.URL.Query().Get("page")); err == nil && p > 0 {
				page = p
			}
			if ps, err := strconv.Atoi(req.URL.Query().Get("pageSize")); err == nil && ps > 0 && ps <= 100 {
				pageSize = ps
			}

			if response.Page != page {
				t.Errorf("Expected page %d, got %d", page, response.Page)
			}
			if response.PageSize != pageSize {
				t.Errorf("Expected pageSize %d, got %d", pageSize, response.PageSize)
			}

			// For filter tests, check the total count is correct
			if tt.queryParams != "" {
				if !strings.Contains(tt.queryParams, "page=") && !strings.Contains(tt.queryParams, "pageSize=") {
					// For pure filter tests (no pagination specified), total should match returned count
					if response.Total != len(response.Transactions) {
						t.Errorf("Total count %d doesn't match returned count %d for filter: %s",
							response.Total, len(response.Transactions), tt.queryParams)
					}
				}
			}
		})
	}

	// Test invalid page/pageSize parameters
	t.Run("Invalid pagination params", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/transactions?page=0&pageSize=-10", nil)
		rr := httptest.NewRecorder()
		handler.handleTransactionsList(rr, req)

		var response TransactionListResponse
		json.Unmarshal(rr.Body.Bytes(), &response)

		// Should default to page 1, pageSize 50
		if response.Page != 1 || response.PageSize != 50 {
			t.Errorf("Expected default pagination (1,50), got (%d,%d)", response.Page, response.PageSize)
		}
	})

	// Test large pageSize parameter (should be capped)
	t.Run("Large pageSize", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/transactions?pageSize=500", nil)
		rr := httptest.NewRecorder()
		handler.handleTransactionsList(rr, req)

		var response TransactionListResponse
		json.Unmarshal(rr.Body.Bytes(), &response)

		// Should be capped at 100
		if response.PageSize != 50 {
			t.Errorf("Expected pageSize to be capped at 50, got %d", response.PageSize)
		}
	})
}

func TestHandleTransactionDetail(t *testing.T) {
	db, dbPath := setupTestDB(t)
	defer cleanupTestDB(db, dbPath)

	handler := NewUIHandler(db)

	// Test cases
	tests := []struct {
		name       string
		id         string
		statusCode int
	}{
		{
			name:       "Get HTTP transaction",
			id:         "http-1",
			statusCode: http.StatusOK,
		},
		{
			name:       "Get WebSocket transaction",
			id:         "ws-1",
			statusCode: http.StatusOK,
		},
		{
			name:       "Non-existent transaction",
			id:         "nonexistent",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Empty ID",
			id:         "",
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/api/transactions/"
			if tt.id != "" {
				path += tt.id
			}

			req, err := http.NewRequest("GET", path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.handleTransactionDetail(rr, req)

			// Check status code
			if status := rr.Code; status != tt.statusCode {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tt.statusCode)
			}

			// If successful, check details
			if tt.statusCode == http.StatusOK {
				var transaction TransactionDetail
				if err := json.Unmarshal(rr.Body.Bytes(), &transaction); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}

				if transaction.ID != tt.id {
					t.Errorf("Wrong transaction ID: got %v want %v", transaction.ID, tt.id)
				}
			}
		})
	}

	// Test invalid path
	t.Run("Invalid path", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/wrong/path", nil)
		rr := httptest.NewRecorder()
		handler.handleTransactionDetail(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid path, got %d", rr.Code)
		}
	})
}

func TestDatabaseErrorHandling(t *testing.T) {
	// Create a database connection that will be closed immediately
	// to simulate database errors
	db, dbPath := setupTestDB(t)
	handler := NewUIHandler(db)

	// Close the database to cause errors
	db.Close()
	os.Remove(dbPath)

	// Test list handler with database error
	t.Run("List handler database error", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/transactions", nil)
		rr := httptest.NewRecorder()
		handler.handleTransactionsList(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500 for DB error, got %d", rr.Code)
		}
	})

	// Test detail handler with database error
	t.Run("Detail handler database error", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/transactions/http-1", nil)
		rr := httptest.NewRecorder()
		handler.handleTransactionDetail(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500 for DB error, got %d", rr.Code)
		}
	})
}
