package certs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateSelfSignedCert(t *testing.T) {
	// Create temp directory for test certificates
	tempDir, err := os.MkdirTemp("", "cert_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test certificate paths
	certPath := filepath.Join(tempDir, "test.crt")
	keyPath := filepath.Join(tempDir, "test.key")

	// Generate certificates
	err = GenerateSelfSignedCert(certPath, keyPath)
	if err != nil {
		t.Fatalf("Certificate generation failed: %v", err)
	}

	// Check if files were created
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Errorf("Certificate file was not created at %s", certPath)
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("Key file was not created at %s", keyPath)
	}

	// Verify contents (basic check)
	certData, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("Failed to read certificate file: %v", err)
	}
	if len(certData) == 0 {
		t.Error("Certificate file is empty")
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("Failed to read key file: %v", err)
	}
	if len(keyData) == 0 {
		t.Error("Key file is empty")
	}
}

func TestGenerateSelfSignedCert_InvalidPath(t *testing.T) {
	// Test with an invalid path
	err := GenerateSelfSignedCert("/invalid/path/cert.pem", "/invalid/path/key.pem")
	if err == nil {
		t.Error("Expected error for invalid certificate path, but got nil")
	}
}
