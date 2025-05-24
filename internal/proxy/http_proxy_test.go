package proxy

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dipjyotimetia/jarvis/config"
)

// MockServer creates a test HTTP server that returns predefined responses
func createMockServer() *httptest.Server {
	handler := http.NewServeMux()

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Default response"))
	})

	handler.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"users": ["user1", "user2"]}`))
	})

	handler.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server error"))
	})

	return httptest.NewServer(handler)
}

func TestStartHTTPProxy_PassthroughMode(t *testing.T) {
	// Create mock target server
	targetServer := createMockServer()
	defer targetServer.Close()

	// Create test database
	tempDB, err := os.CreateTemp("", "test_db_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer os.Remove(tempDB.Name())
	tempDB.Close()

	db, err := sql.Open("sqlite3", tempDB.Name())
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
        CREATE TABLE traffic_records (
            id TEXT PRIMARY KEY,
            timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
            protocol TEXT NOT NULL,
            method TEXT NOT NULL,
            url TEXT,
            service TEXT,
            request_headers TEXT,
            request_body BLOB,
            response_status INTEGER NOT NULL,
            response_headers TEXT,
            response_body BLOB,
            duration_ms INTEGER,
            client_ip TEXT,
            test_id TEXT,
            session_id TEXT,
			connection_id TEXT,
        	message_type INTEGER,
        	direction TEXT
        );
    `)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	stmt, err := db.Prepare(`INSERT INTO traffic_records VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		t.Fatalf("Failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	// Create config for test
	cfg := &config.Config{
		HTTPPort:      8080,
		HTTPTargetURL: targetServer.URL,
		RecordingMode: false,
		ReplayMode:    false,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start proxy server
	proxyServer := StartHTTPProxy(ctx, cfg, db, stmt)
	defer proxyServer.Shutdown(context.Background())

	// Wait a moment for server to start
	time.Sleep(100 * time.Millisecond)

	// Use httptest.NewServer with our handler to avoid needing a real port
	// Create a test HTTP proxy server using the handler function
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handleHTTPRequest(w, r, httputil.NewSingleHostReverseProxy(
			&url.URL{Scheme: "http", Host: strings.TrimPrefix(targetServer.URL, "http://")},
		), cfg, db, stmt, nil)
	}

	proxyTestServer := httptest.NewServer(http.HandlerFunc(testHandler))
	defer proxyTestServer.Close()

	// Test basic request
	resp, err := http.Get(proxyTestServer.URL + "/users")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !bytes.Contains(body, []byte(`"users": ["user1", "user2"]`)) {
		t.Errorf("Unexpected response body: %s", body)
	}
}

// More tests can be added for recording mode, replay mode, etc.
