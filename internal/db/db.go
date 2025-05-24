package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// TrafficRecord represents a captured API call
type TrafficRecord struct {
	ID              string    `json:"id"`
	Timestamp       time.Time `json:"timestamp"`
	Protocol        string    `json:"protocol"` // HTTP or WebSocket
	Method          string    `json:"method"`   // HTTP method or WS event type (connect/message/close)
	URL             string    `json:"url"`
	Service         string    `json:"service"` // Added service field
	RequestHeaders  string    `json:"request_headers"`
	RequestBody     []byte    `json:"request_body"`
	ResponseStatus  int       `json:"response_status"` // HTTP status code
	ResponseHeaders string    `json:"response_headers"`
	ResponseBody    []byte    `json:"response_body"`
	Duration        int64     `json:"duration_ms"` // Duration in milliseconds
	ClientIP        string    `json:"client_ip"`
	TestID          string    `json:"test_id"`       // Optional: Link to test case ID
	SessionID       string    `json:"session_id"`    // For grouping related requests
	ConnectionID    string    `json:"connection_id"` // WebSocket connection identifier
	MessageType     int       `json:"message_type"`  // For WebSocket: text/binary/etc.
	Direction       string    `json:"direction"`     // For WebSocket: inbound/outbound
}

// Initialize sets up the database connection and schema
func Initialize(dbPath string) (*sql.DB, *sql.Stmt, error) {
	// Initialize SQLite client with improved concurrency settings
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=30000&_timeout=30000&cache=shared&_pragma=mmap_size=30000000000")
	if err != nil {
		return nil, nil, fmt.Errorf("opening SQLite database: %w", err)
	}

	// Set optimized SQLite connection pool settings
	db.SetMaxOpenConns(25)                  // Increased from 1 to handle more concurrent operations
	db.SetMaxIdleConns(10)                  // Increased from 1 to maintain a pool of idle connections
	db.SetConnMaxLifetime(30 * time.Minute) // Increased from 5 minutes to reduce connection churn
	db.SetConnMaxIdleTime(10 * time.Minute) // Added idle timeout

	if err := db.Ping(); err != nil {
		return nil, nil, fmt.Errorf("pinging SQLite database: %w", err)
	}
	log.Printf("🔗 Connected to SQLite database at %s", dbPath)

	// Create schema and prepare statements
	stmt, err := setupDatabase(db)
	if err != nil {
		return nil, nil, fmt.Errorf("setting up database: %w", err)
	}

	return db, stmt, nil
}

// setupDatabase creates schema and prepares insert statement
func setupDatabase(db *sql.DB) (*sql.Stmt, error) {
	// Schema creation SQL with WebSocket support and optimized indexes
	query := `
    CREATE TABLE IF NOT EXISTS traffic_records (
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

    -- Optimized indexes for common queries
    CREATE INDEX IF NOT EXISTS idx_http_lookup ON traffic_records(protocol, method, url) WHERE protocol = 'HTTP';
    CREATE INDEX IF NOT EXISTS idx_ws_lookup ON traffic_records(protocol, connection_id) WHERE protocol = 'WebSocket';
    CREATE INDEX IF NOT EXISTS idx_timestamp ON traffic_records(timestamp DESC);
    CREATE INDEX IF NOT EXISTS idx_session_id ON traffic_records(session_id) WHERE session_id IS NOT NULL;
    CREATE INDEX IF NOT EXISTS idx_test_id ON traffic_records(test_id) WHERE test_id IS NOT NULL;
    
    -- Add covering index for common query patterns
    CREATE INDEX IF NOT EXISTS idx_http_common ON traffic_records(protocol, method, url, timestamp, response_status) 
    WHERE protocol = 'HTTP';
    `

	_, err := db.Exec(query)
	if err != nil {
		return nil, fmt.Errorf("creating schema: %w", err)
	}

	// Prepare statement for inserts with optimized parameters
	insertSQL := `INSERT INTO traffic_records (
        id, timestamp, protocol, method, url, service,
        request_headers, request_body, response_status,
        response_headers, response_body, duration_ms,
        client_ip, test_id, session_id, connection_id,
        message_type, direction
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		return nil, fmt.Errorf("preparing insert statement: %w", err)
	}

	log.Println("✅ Database schema verified and statement prepared")
	return stmt, nil
}
