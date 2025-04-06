package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name      string
		configMap map[string]interface{}
		wantErr   bool
	}{
		// Existing test cases...
		{
			name: "Valid configuration",
			configMap: map[string]interface{}{
				"http_port":       8080,
				"http_target_url": "http://example.com",
				"sqlite_db_path":  "test.db",
			},
			wantErr: false,
		},
		{
			name: "Missing HTTP port",
			configMap: map[string]interface{}{
				"http_target_url": "http://example.com",
			},
			wantErr: true,
		},
		{
			name: "Missing target URL",
			configMap: map[string]interface{}{
				"http_port": 8080,
			},
			wantErr: true,
		},
		{
			name: "Both recording and replay mode",
			configMap: map[string]interface{}{
				"http_port":       8080,
				"http_target_url": "http://example.com",
				"recording_mode":  true,
				"replay_mode":     true,
			},
			wantErr: true,
		},
		{
			name: "TLS enabled without cert",
			configMap: map[string]interface{}{
				"http_port":       8080,
				"http_target_url": "http://example.com",
				"tls": map[string]interface{}{
					"enabled": true,
					"port":    8443,
				},
			},
			wantErr: true,
		},
		{
			name: "Valid TLS config",
			configMap: map[string]interface{}{
				"http_port":       8080,
				"http_target_url": "http://example.com",
				"tls": map[string]interface{}{
					"enabled":   true,
					"port":      8443,
					"cert_file": "cert.pem",
					"key_file":  "key.pem",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid target routes",
			configMap: map[string]interface{}{
				"http_port": 8080,
				"target_routes": []map[string]interface{}{
					{
						"path_prefix": "/api",
						"target_url":  "http://api.example.com",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid target routes - missing path prefix",
			configMap: map[string]interface{}{
				"http_port": 8080,
				"target_routes": []map[string]interface{}{
					{
						"target_url": "http://api.example.com",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid target routes - path prefix without leading slash",
			configMap: map[string]interface{}{
				"http_port": 8080,
				"target_routes": []map[string]interface{}{
					{
						"path_prefix": "api",
						"target_url":  "http://api.example.com",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			for key, value := range tt.configMap {
				v.Set(key, value)
			}

			config, err := LoadConfig(v)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify defaults are set
				if config.SQLiteDBPath == "" {
					t.Error("Default SQLiteDBPath not set")
				}

				// Verify TLS port default
				if config.TLS.Enabled && config.TLS.Port == 0 {
					t.Error("Default TLS port not set when TLS enabled")
				}

			}
		})
	}
}

func TestGetTargetURL(t *testing.T) {
	config := &Config{
		HTTPTargetURL: "https://default.example.com",
		TargetRoutes: []TargetRoute{
			{
				PathPrefix: "/api/users",
				TargetURL:  "https://users.example.com",
			},
			{
				PathPrefix: "/api/products",
				TargetURL:  "https://products.example.com",
			},
		},
	}

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Match first route",
			path:     "/api/users/123",
			expected: "https://users.example.com",
		},
		{
			name:     "Match second route",
			path:     "/api/products/456",
			expected: "https://products.example.com",
		},
		{
			name:     "Use default for non-matching path",
			path:     "/other/path",
			expected: "https://default.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.GetTargetURL(tt.path)
			if result != tt.expected {
				t.Errorf("GetTargetURL() got = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetWebSocketTargetURL tests the GetWebSocketTargetURL method
