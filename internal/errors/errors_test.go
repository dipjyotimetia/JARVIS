package errors

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestJarvisError(t *testing.T) {
	t.Run("Basic error creation", func(t *testing.T) {
		err := New(ErrCodeDatabase, "database connection failed")
		if err.Code != ErrCodeDatabase {
			t.Errorf("Expected code %s, got %s", ErrCodeDatabase, err.Code)
		}
		if err.Message != "database connection failed" {
			t.Errorf("Expected message 'database connection failed', got '%s'", err.Message)
		}
	})

	t.Run("Error wrapping", func(t *testing.T) {
		cause := errors.New("connection refused")
		err := Wrap(ErrCodeDatabase, "failed to connect", cause)

		if err.Cause != cause {
			t.Error("Cause not properly wrapped")
		}

		unwrapped := errors.Unwrap(err)
		if unwrapped != cause {
			t.Error("Unwrap did not return original cause")
		}
	})

	t.Run("Error with context", func(t *testing.T) {
		err := New(ErrCodeLLM, "API call failed").
			WithContext("provider", "openai").
			WithContext("attempt", 3)

		if val, ok := err.Context["provider"]; !ok || val != "openai" {
			t.Error("Context not properly set")
		}

		if val, ok := GetContext(err, "attempt"); !ok || val != 3 {
			t.Error("Failed to get context value")
		}
	})

	t.Run("Error code checking", func(t *testing.T) {
		err := New(ErrCodeValidation, "invalid input")

		if !IsErrorCode(err, ErrCodeValidation) {
			t.Error("IsErrorCode failed to match correct code")
		}

		if IsErrorCode(err, ErrCodeDatabase) {
			t.Error("IsErrorCode matched incorrect code")
		}
	})

	t.Run("JSON serialization", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := Wrap(ErrCodeProxy, "proxy failed", cause).
			WithContext("url", "http://example.com")

		data, jsonErr := json.Marshal(err)
		if jsonErr != nil {
			t.Fatalf("Failed to marshal error: %v", jsonErr)
		}

		var decoded map[string]interface{}
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal error: %v", err)
		}

		if decoded["code"] != string(ErrCodeProxy) {
			t.Error("Code not properly serialized")
		}
	})

	t.Run("Predefined constructors", func(t *testing.T) {
		tests := []struct {
			name        string
			constructor func(string, error) *JarvisError
			expectedCode ErrorCode
		}{
			{"Database", NewDatabaseError, ErrCodeDatabase},
			{"LLM", NewLLMError, ErrCodeLLM},
			{"Proxy", NewProxyError, ErrCodeProxy},
			{"Validation", NewValidationError, ErrCodeValidation},
			{"Config", NewConfigError, ErrCodeConfig},
			{"Certificate", NewCertificateError, ErrCodeCertificate},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.constructor("test error", nil)
				if err.Code != tt.expectedCode {
					t.Errorf("Expected code %s, got %s", tt.expectedCode, err.Code)
				}
			})
		}
	})
}
