package validator

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAPIValidator(t *testing.T) {
	// Create a temporary test spec file
	specDir := t.TempDir()
	specPath := filepath.Join(specDir, "test-openapi.yaml")

	// Simple test OpenAPI spec
	specContent := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get users
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    id:
                      type: string
                    name:
                      type: string
    post:
      summary: Create user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - name
              properties:
                name:
                  type: string
      responses:
        '201':
          description: User created
  /users/{userId}:
    get:
      summary: Get user by ID
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: User found
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  name:
                    type: string
        '404':
          description: User not found
`

	if err := os.WriteFile(specPath, []byte(specContent), 0o644); err != nil {
		t.Fatalf("Failed to write test spec file: %v", err)
	}

	// Create validator with test options
	options := APIValidatorOptions{
		EnableRequestValidation:  true,
		EnableResponseValidation: true,
		StrictMode:               true,
	}

	validator, err := NewAPIValidator(specPath, options)
	if err != nil {
		t.Fatalf("Failed to create API validator: %v", err)
	}

	// Test getting OpenAPI info
	info := validator.GetOpenAPIInfo()
	if info["title"] != "Test API" {
		t.Errorf("Expected title 'Test API', got %s", info["title"])
	}
	if info["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", info["version"])
	}

	// Test getting paths with methods
	paths := validator.GetPathsWithMethods()
	if len(paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(paths))
	}
	if len(paths["/users"]) != 2 || !contains(paths["/users"], "GET") || !contains(paths["/users"], "POST") {
		t.Errorf("Expected /users path to have GET and POST methods, got %v", paths["/users"])
	}
	if len(paths["/users/{userId}"]) != 1 || !contains(paths["/users/{userId}"], "GET") {
		t.Errorf("Expected /users/{userId} path to have GET method, got %v", paths["/users/{userId}"])
	}

	// Test request validation - valid request
	t.Run("ValidRequest", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users", nil)
		req.Header.Set("Accept", "application/json")

		err := validator.ValidateRequest(req)
		if err != nil {
			t.Errorf("Expected valid request to pass validation, got error: %v", err)
		}
	})

	// Test request validation - invalid method
	t.Run("InvalidMethod", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/users", nil)

		err := validator.ValidateRequest(req)
		if err == nil || !strings.Contains(err.Error(), "method DELETE not allowed") {
			t.Errorf("Expected error about method not allowed, got: %v", err)
		}
	})

	// Test response validation - valid response
	t.Run("ValidResponse", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users", nil)
		resp := httptest.NewRecorder()
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte(`[{"id":"1","name":"Test User"}]`))

		err := validator.ValidateResponse(req, resp.Code, resp.Header(), resp.Body.Bytes())
		if err != nil {
			t.Errorf("Expected valid response to pass validation, got error: %v", err)
		}
	})

	// Test response validation - invalid response
	t.Run("InvalidResponse", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users", nil)
		resp := httptest.NewRecorder()
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte(`{"invalid":"response"}`)) // Should be an array

		err := validator.ValidateResponse(req, resp.Code, resp.Header(), resp.Body.Bytes())
		if err == nil {
			t.Errorf("Expected invalid response to fail validation")
		}
	})
}

// Test helper for path normalization
func TestNormalizePathForSpec(t *testing.T) {
	specPaths := map[string]struct{}{
		"/api/v1/users":         {},
		"/api/v1/users/{id}":    {},
		"/api/v1/products/{id}": {},
	}

	tests := []struct {
		actualPath       string
		expectedSpecPath string
		shouldMatch      bool
	}{
		{"/api/v1/users", "/api/v1/users", true},
		{"/api/v1/users/123", "/api/v1/users/{id}", true},
		{"/api/v1/products/456", "/api/v1/products/{id}", true},
		{"/api/v2/users", "", false},
		{"/api/v1/orders", "", false},
	}

	for _, test := range tests {
		t.Run(test.actualPath, func(t *testing.T) {
			specPath, matched := NormalizePathForSpec(test.actualPath, specPaths)
			if matched != test.shouldMatch {
				t.Errorf("Path %s: expected match=%v, got %v", test.actualPath, test.shouldMatch, matched)
			}

			if matched && specPath != test.expectedSpecPath {
				t.Errorf("Path %s: expected spec path %s, got %s", test.actualPath, test.expectedSpecPath, specPath)
			}
		})
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
