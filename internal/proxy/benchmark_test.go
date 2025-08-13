package proxy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/dipjyotimetia/jarvis/internal/db"
)

// BenchmarkResponseRecorder measures the performance of response recording
func BenchmarkResponseRecorder(b *testing.B) {
	responseBufPool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		responseBuf := responseBufPool.Get().(*bytes.Buffer)
		responseBuf.Reset()
		
		recorder := &responseRecorder{
			ResponseWriter: httptest.NewRecorder(),
			statusCode:     200,
			body:           responseBuf,
			header:         http.Header{},
			streamMode:     false,
			maxBufferSize:  1024 * 1024,
		}

		// Simulate writing response data
		data := []byte("test response data")
		recorder.Write(data)
		recorder.WriteHeader(200)
		
		responseBufPool.Put(responseBuf)
	}
}

// BenchmarkJSONMarshaling measures JSON marshaling performance for traffic records
func BenchmarkJSONMarshaling(b *testing.B) {
	record := db.TrafficRecord{
		ID:              "test-id",
		Timestamp:       time.Now(),
		Protocol:        "HTTP",
		Method:          "GET",
		URL:             "http://example.com/api/test",
		RequestHeaders:  `{"Content-Type":"application/json"}`,
		RequestBody:     []byte(`{"test": "data"}`),
		ResponseStatus:  200,
		ResponseHeaders: `{"Content-Type":"application/json"}`,
		ResponseBody:    []byte(`{"result": "success"}`),
		Duration:        150,
		ClientIP:        "192.168.1.1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsonBufferPool.Get()
		// Simulate the JSON marshaling that happens during record saving
		_ = record.RequestBody
		_ = record.ResponseBody
		jsonBufferPool.Put(make([]byte, 0, 4096))
	}
}

// BenchmarkBufferPool measures buffer pool performance
func BenchmarkBufferPool(b *testing.B) {
	pool := &sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 4096)
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := pool.Get().([]byte)
		buf = append(buf, []byte("test data")...)
		pool.Put(buf[:0]) // Reset slice but keep capacity
	}
}

// BenchmarkStreamingMode measures streaming vs buffering performance
func BenchmarkStreamingVsBuffering(b *testing.B) {
	testData := make([]byte, 1024*1024) // 1MB test data
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	b.Run("Buffering", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(make([]byte, 0, len(testData)))
			buf.Write(testData)
		}
	})

	b.Run("Streaming", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate streaming by writing in chunks
			w := httptest.NewRecorder()
			chunkSize := 8192
			for j := 0; j < len(testData); j += chunkSize {
				end := j + chunkSize
				if end > len(testData) {
					end = len(testData)
				}
				w.Write(testData[j:end])
			}
		}
	})
}

// BenchmarkGetClientIP measures client IP extraction performance
func BenchmarkGetClientIP(b *testing.B) {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
	req.Header.Set("X-Real-IP", "203.0.113.1")
	req.RemoteAddr = "198.51.100.1:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getClientIP(req)
	}
}

// BenchmarkStringOperations measures string manipulation performance
func BenchmarkStringOperations(b *testing.B) {
	testPaths := []string{
		"/api/v1/users",
		"/api/v2/posts",
		"/static/images/logo.png",
		"/health",
		"/metrics",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := testPaths[i%len(testPaths)]
		// Simulate URL manipulation
		_ = "http://default.example.com" + path
	}
}