package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// initStartFlags initializes the flags for the start command

func initStartFlags() {
	proxyCmd.Flags().BoolP("record", "r", false, "Enable recording mode")
	proxyCmd.Flags().BoolP("replay", "p", false, "Enable replay mode")
	proxyCmd.Flags().BoolP("tls", "t", false, "Enable TLS")
	proxyCmd.Flags().String("cert", "", "TLS certificate file path")
	proxyCmd.Flags().String("key", "", "TLS key file path")
	proxyCmd.Flags().Int("tls-port", 443, "TLS port")

	// Add WebSocket flag initialization for tests
	proxyCmd.Flags().Bool("websocket", false, "Enable WebSocket support")
	proxyCmd.Flags().Int("ws-read-buffer", 4096, "WebSocket read buffer size in bytes")
	proxyCmd.Flags().Int("ws-write-buffer", 4096, "WebSocket write buffer size in bytes")
}

func TestVersionCommand(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute version command
	rootCmd.SetArgs([]string{"version"})
	err := rootCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if err != nil {
		t.Errorf("version command failed: %v", err)
	}

	// Check that output contains version info
	if !strings.Contains(output, "Traffic Inspector v") {
		t.Errorf("version command output does not contain version info: %s", output)
	}
}

func TestCertificateCommand(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "cert_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Capture stdout
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	// Execute certificate command with temp directory
	rootCmd.SetArgs([]string{
		"certificate",
		"--cert-dir", tempDir,
	})

	err = rootCmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Errorf("certificate command failed: %v", err)
	}

	// Check that files were created
	certPath := tempDir + "/server.crt"
	keyPath := tempDir + "/server.key"

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Errorf("Certificate file was not created at %s", certPath)
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("Key file was not created at %s", keyPath)
	}
}

func TestRootCommand(t *testing.T) {
	// Test with no args (should not error)
	rootCmd.SetArgs([]string{})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("root command failed: %v", err)
	}

	// Test with help flag
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"--help"})
	rootCmd.Execute()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "traffix") {
		t.Errorf("help output does not contain expected text: %s", output)
	}
}
