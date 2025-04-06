package db

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func TestInitialize(t *testing.T) {
	// Create a temporary file for the database
	tempFile, err := os.CreateTemp("", "traffic_inspector_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Test initialization
	database, stmt, err := Initialize(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	defer stmt.Close()

	// Check that the database connection works
	err = database.Ping()
	if err != nil {
		t.Errorf("Failed to ping database after initialization: %v", err)
	}
}

func TestTrafficRecordStorage(t *testing.T) {
	// Create a temporary file for the database
	tempFile, err := os.CreateTemp("", "traffic_inspector_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Initialize database
	database, stmt, err := Initialize(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	defer stmt.Close()

	// Test data
	testRecords := []TrafficRecord{
		{
			ID:              "test-record-1",
			Timestamp:       time.Now().UTC(),
			Protocol:        "HTTP",
			Method:          "GET",
			URL:             "https://example.com/api/users",
			RequestHeaders:  `{"Content-Type": ["application/json"]}`,
			RequestBody:     []byte(`{"query": "users"}`),
			ResponseStatus:  200,
			ResponseHeaders: `{"Content-Type": ["application/json"]}`,
			ResponseBody:    []byte(`["user1", "user2"]`),
			Duration:        150,
			ClientIP:        "192.168.1.1",
			TestID:          "test-123",
			SessionID:       "session-abc",
		},
		{
			ID:              "test-record-2",
			Timestamp:       time.Now().UTC(),
			Protocol:        "HTTP",
			Method:          "POST",
			URL:             "https://example.com/api/products",
			RequestHeaders:  `{"Content-Type": ["application/json"]}`,
			RequestBody:     []byte(`{"name": "Product 1"}`),
			ResponseStatus:  201,
			ResponseHeaders: `{"Content-Type": ["application/json"]}`,
			ResponseBody:    []byte(`{"id": "p1", "name": "Product 1"}`),
			Duration:        120,
			ClientIP:        "192.168.1.2",
			TestID:          "test-456",
			SessionID:       "session-def",
		},
		// Add WebSocket records for testing
		{
			ID:              "ws-connect-1",
			Timestamp:       time.Now().UTC(),
			Protocol:        "WebSocket",
			Method:          "CONNECT",
			URL:             "wss://example.com/ws/chat",
			RequestHeaders:  `{"Upgrade": ["websocket"], "Connection": ["Upgrade"]}`,
			ResponseStatus:  101,
			ResponseHeaders: `{"Upgrade": ["websocket"], "Connection": ["Upgrade"]}`,
			Duration:        50,
			ClientIP:        "192.168.1.3",
			TestID:          "test-ws-1",
			SessionID:       "session-ws-1",
			ConnectionID:    "conn-abc-123",
		},
		{
			ID:             "ws-message-1",
			Timestamp:      time.Now().UTC(),
			Protocol:       "WebSocket",
			Method:         "MESSAGE",
			URL:            "wss://example.com/ws/chat",
			RequestBody:    []byte(`{"type": "chat", "message": "Hello"}`),
			ResponseStatus: 200,
			Duration:       10,
			ClientIP:       "192.168.1.3",
			TestID:         "test-ws-1",
			SessionID:      "session-ws-1",
			ConnectionID:   "conn-abc-123",
			MessageType:    1, // Text message
			Direction:      "outbound",
		},
		{
			ID:             "ws-message-2",
			Timestamp:      time.Now().UTC(),
			Protocol:       "WebSocket",
			Method:         "MESSAGE",
			URL:            "wss://example.com/ws/chat",
			ResponseBody:   []byte(`{"type": "chat", "message": "Hi there!"}`),
			ResponseStatus: 200,
			Duration:       15,
			ClientIP:       "192.168.1.3",
			TestID:         "test-ws-1",
			SessionID:      "session-ws-1",
			ConnectionID:   "conn-abc-123",
			MessageType:    1, // Text message
			Direction:      "inbound",
		},
		{
			ID:             "ws-close-1",
			Timestamp:      time.Now().UTC(),
			Protocol:       "WebSocket",
			Method:         "CLOSE",
			URL:            "wss://example.com/ws/chat",
			ResponseStatus: 0,
			Duration:       5,
			ClientIP:       "192.168.1.3",
			TestID:         "test-ws-1",
			SessionID:      "session-ws-1",
			ConnectionID:   "conn-abc-123",
		},
	}

	// Insert test records
	for _, record := range testRecords {
		_, err := stmt.Exec(
			record.ID,
			record.Timestamp,
			record.Protocol,
			record.Method,
			record.URL,
			"", // service field
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
			t.Fatalf("Failed to insert test record: %v", err)
		}
	}

	// Test HTTP queries
	t.Run("QueryByID", func(t *testing.T) {
		var id, protocol, method, url string
		var status int

		err := database.QueryRow("SELECT id, protocol, method, url, response_status FROM traffic_records WHERE id = ?", "test-record-1").
			Scan(&id, &protocol, &method, &url, &status)
		if err != nil {
			t.Fatalf("Failed to query record: %v", err)
		}

		if id != "test-record-1" || protocol != "HTTP" || method != "GET" ||
			url != "https://example.com/api/users" || status != 200 {
			t.Errorf("Retrieved record doesn't match: got id=%s, protocol=%s, method=%s, url=%s, status=%d",
				id, protocol, method, url, status)
		}
	})

	t.Run("QueryByMethodAndURL", func(t *testing.T) {
		var id string
		var status int

		err := database.QueryRow(
			"SELECT id, response_status FROM traffic_records WHERE method = ? AND url = ?",
			"POST", "https://example.com/api/products",
		).Scan(&id, &status)
		if err != nil {
			t.Fatalf("Failed to query record by method and URL: %v", err)
		}

		if id != "test-record-2" || status != 201 {
			t.Errorf("Retrieved record doesn't match: got id=%s, status=%d", id, status)
		}
	})

	// Test WebSocket queries
	t.Run("QueryWebSocketByConnectionID", func(t *testing.T) {
		var count int

		err := database.QueryRow(
			"SELECT COUNT(*) FROM traffic_records WHERE protocol = ? AND connection_id = ?",
			"WebSocket", "conn-abc-123",
		).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count WebSocket records by connection ID: %v", err)
		}

		if count != 4 {
			t.Errorf("Wrong count for WebSocket connection: got %d, want 4", count)
		}
	})

	t.Run("QueryWebSocketMessagesByDirection", func(t *testing.T) {
		var count int

		err := database.QueryRow(
			"SELECT COUNT(*) FROM traffic_records WHERE protocol = ? AND method = ? AND direction = ?",
			"WebSocket", "MESSAGE", "outbound",
		).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count WebSocket outbound messages: %v", err)
		}

		if count != 1 {
			t.Errorf("Wrong count for WebSocket outbound messages: got %d, want 1", count)
		}
	})

	t.Run("QueryWebSocketMessagesByType", func(t *testing.T) {
		var count int

		err := database.QueryRow(
			"SELECT COUNT(*) FROM traffic_records WHERE protocol = ? AND method = ? AND message_type = ?",
			"WebSocket", "MESSAGE", 1,
		).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count WebSocket text messages: %v", err)
		}

		if count != 2 {
			t.Errorf("Wrong count for WebSocket text messages: got %d, want 2", count)
		}
	})

	t.Run("QueryBySession", func(t *testing.T) {
		var count int

		err := database.QueryRow(
			"SELECT COUNT(*) FROM traffic_records WHERE session_id = ?",
			"session-ws-1",
		).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count records by WebSocket session: %v", err)
		}

		if count != 4 {
			t.Errorf("Wrong count for session-ws-1: got %d, want 4", count)
		}
	})
}

func TestDatabaseIndexes(t *testing.T) {
	// Create a temporary file for the database
	tempFile, err := os.CreateTemp("", "traffic_inspector_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Initialize database
	database, stmt, err := Initialize(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	defer stmt.Close()

	// Check that indexes exist
	rows, err := database.Query("SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = 'traffic_records'")
	if err != nil {
		t.Fatalf("Failed to query indexes: %v", err)
	}
	defer rows.Close()

	indexes := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("Failed to scan index name: %v", err)
		}
		indexes[name] = true
	}

	// Verify expected indexes
	expectedIndexes := []string{
		"idx_http_lookup",
		"idx_ws_lookup", // WebSocket index
		"idx_timestamp",
		"idx_session_id",
		"idx_test_id",
	}

	for _, idx := range expectedIndexes {
		if !indexes[idx] {
			t.Errorf("Expected index %s not found", idx)
		}
	}
}

func TestDatabaseFailure(t *testing.T) {
	// Test with an invalid database path
	_, _, err := Initialize("/nonexistent/directory/db.sqlite")
	if err == nil {
		t.Error("Expected error with invalid database path, but got nil")
	}

	// Test with a read-only directory
	if os.Getuid() == 0 {
		t.Skip("Skipping read-only test when running as root")
	}

	tempDir, err := os.MkdirTemp("", "readonly_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Make directory read-only
	err = os.Chmod(tempDir, 0o500)
	if err != nil {
		t.Fatalf("Failed to set directory permissions: %v", err)
	}

	dbPath := tempDir + "/test.db"
	_, _, err = Initialize(dbPath)
	if err == nil {
		t.Error("Expected error with read-only directory, but got nil")
	}
}

func TestDatabaseConcurrency(t *testing.T) {
	// Create a temporary file for the database
	tempFile, err := os.CreateTemp("", "traffic_inspector_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Initialize database
	database, stmt, err := Initialize(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()
	defer stmt.Close()

	// Number of concurrent operations
	const numConcurrent = 10
	errCh := make(chan error, numConcurrent)
	doneCh := make(chan bool, numConcurrent)

	// Use a mutex to protect ID generation
	var idMutex sync.Mutex
	idCounter := 0

	// Insert records concurrently - test both HTTP and WebSocket records
	for i := 0; i < numConcurrent; i++ {
		go func(idx int) {
			// Generate a truly unique ID
			idMutex.Lock()
			idCounter++
			uniqueID := fmt.Sprintf("concurrent-%d-%d", idx, idCounter)
			idMutex.Unlock()

			// Alternate between HTTP and WebSocket records
			var record TrafficRecord
			if idx%2 == 0 {
				// HTTP record
				record = TrafficRecord{
					ID:              uniqueID,
					Timestamp:       time.Now().UTC(),
					Protocol:        "HTTP",
					Method:          "GET",
					URL:             fmt.Sprintf("https://example.com/api/concurrent/%d", idx),
					RequestHeaders:  `{"Content-Type": ["application/json"]}`,
					RequestBody:     []byte(fmt.Sprintf(`{"worker": %d}`, idx)),
					ResponseStatus:  200,
					ResponseHeaders: `{"Content-Type": ["application/json"]}`,
					ResponseBody:    []byte(`{"result": "ok"}`),
					Duration:        int64(100 + idx),
					ClientIP:        fmt.Sprintf("192.168.1.%d", idx),
					TestID:          fmt.Sprintf("test-%d", idx),
					SessionID:       fmt.Sprintf("session-%d", idx),
				}
			} else {
				// WebSocket record
				record = TrafficRecord{
					ID:             uniqueID,
					Timestamp:      time.Now().UTC(),
					Protocol:       "WebSocket",
					Method:         "MESSAGE",
					URL:            fmt.Sprintf("wss://example.com/ws/concurrent/%d", idx),
					RequestBody:    []byte(fmt.Sprintf(`{"worker": %d, "msg": "hello"}`, idx)),
					ResponseStatus: 200,
					Duration:       int64(50 + idx),
					ClientIP:       fmt.Sprintf("192.168.1.%d", idx),
					TestID:         fmt.Sprintf("test-ws-%d", idx),
					SessionID:      fmt.Sprintf("session-ws-%d", idx),
					ConnectionID:   fmt.Sprintf("conn-%d", idx),
					MessageType:    1, // Text message
					Direction:      "outbound",
				}
			}

			_, err := stmt.Exec(
				record.ID,
				record.Timestamp,
				record.Protocol,
				record.Method,
				record.URL,
				"", // service field
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
				errCh <- err
				return
			}

			doneCh <- true
		}(i)
	}

	// Collect results
	successCount := 0
	errorCount := 0

	for i := 0; i < numConcurrent; i++ {
		select {
		case <-doneCh:
			successCount++
		case err := <-errCh:
			errorCount++
			t.Logf("Concurrent insert error: %v", err)
		}
	}

	if errorCount > 0 {
		t.Errorf("Got %d errors during concurrent inserts", errorCount)
	}

	// Verify total count
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM traffic_records").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}

	if count != numConcurrent {
		t.Errorf("Expected %d records, got %d", numConcurrent, count)
	}

	// Verify protocol split
	var wsCount int
	err = database.QueryRow("SELECT COUNT(*) FROM traffic_records WHERE protocol = 'WebSocket'").Scan(&wsCount)
	if err != nil {
		t.Fatalf("Failed to count WebSocket records: %v", err)
	}

	var httpCount int
	err = database.QueryRow("SELECT COUNT(*) FROM traffic_records WHERE protocol = 'HTTP'").Scan(&httpCount)
	if err != nil {
		t.Fatalf("Failed to count HTTP records: %v", err)
	}

	// We should have approximately half WebSocket and half HTTP records
	if wsCount < numConcurrent/4 || httpCount < numConcurrent/4 {
		t.Errorf("Unexpected protocol distribution: HTTP=%d, WebSocket=%d", httpCount, wsCount)
	}
}
