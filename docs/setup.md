# Jarvis Setup Guide

This guide covers the setup and configuration of Jarvis, an AI-powered CLI tool for API testing and traffic inspection.

## Prerequisites

Before setting up Jarvis, ensure you have the following installed:

- **Ollama**: Required for local AI model integration
  - Install from [https://ollama.ai/](https://ollama.ai/)
  - Pull required models: `ollama pull llama2` or `ollama pull codellama`
- **Go 1.24+**: Required for building from source
- **Git**: For cloning the repository

## Installation Methods

### Method 1: Download Pre-built Binary

1. Visit the [Releases page](https://github.com/dipjyotimetia/jarvis/releases)
2. Download the appropriate binary for your platform
3. Make it executable and add to PATH:

```bash
# Linux/macOS
chmod +x jarvis
sudo mv jarvis /usr/local/bin/

# Windows
# Move jarvis.exe to a directory in your PATH
```

### Method 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/dipjyotimetia/jarvis.git
cd jarvis

# Build the application
go build -o jarvis

# Install to PATH (optional)
sudo mv jarvis /usr/local/bin/
```

## Interactive Setup

Jarvis provides an interactive setup wizard to configure your environment:

```bash
jarvis setup
```

This wizard will guide you through:
- AI API configuration (optional for offline mode)
- Output directory preferences
- Programming language and framework selection
- Proxy settings configuration
- Report format preferences

## Manual Configuration

### Environment Variables

Set these environment variables for advanced configuration:

```bash
# Optional: AI API configuration
export JARVIS_AI_KEY="your_api_key_here"

# Optional: Custom configuration file path
export JARVIS_CONFIG="./custom-config.yaml"

# Optional: Default output directory
export JARVIS_OUTPUT_DIR="./jarvis-output"
```

### Configuration File

Create a `config.yaml` file in your project directory:

```yaml
# Basic Configuration
http_port: 8080
ui_port: 9090
sqlite_db_path: "./data/traffic_inspector.db"

# AI Configuration
ai:
  provider: "ollama"
  model: "llama2"
  endpoint: "http://localhost:11434"

# Proxy Settings
recording_mode: false
replay_mode: false

# TLS Configuration
tls:
  enabled: false
  port: 8443
  cert_file: "./certs/server.crt"
  key_file: "./certs/server.key"
  client_auth: false
  allow_insecure: false

# API Validation
api_validation:
  enabled: false
  spec_path: ""
  validate_requests: true
  validate_responses: true
  strict_mode: false
  continue_on_validation: false

# Default Preferences
language:
  preferred: "javascript"
  framework: "jest"

output:
  directory: "./output"
  report_format: "HTML"

# Target Routes (Path-based routing)
target_routes:
  - path: "/api/users/*"
    target: "https://api.example.com"
  - path: "/api/products/*"
    target: "https://products.service.com"

# Default target for unmatched routes
http_target_url: "https://jsonplaceholder.typicode.com"
```

## Ollama Setup

### Installation

1. **Install Ollama**:
   ```bash
   # Linux
   curl -fsSL https://ollama.ai/install.sh | sh
   
   # macOS
   brew install ollama
   
   # Windows: Download from https://ollama.ai/
   ```

2. **Start Ollama service**:
   ```bash
   ollama serve
   ```

3. **Pull required models**:
   ```bash
   # For code generation
   ollama pull codellama
   
   # For general text generation
   ollama pull llama2
   
   # For smaller, faster responses
   ollama pull llama3.2
   ```

### Verify Ollama Integration

```bash
# Test Ollama connection
curl http://localhost:11434/api/tags

# Generate a simple test scenario
jarvis gen generate-scenarios --path="./specs/openapi/example.yaml"
```

## Certificate Setup (for HTTPS)

Generate self-signed certificates for testing:

```bash
# Create certificates directory
mkdir -p ./certs

# Generate self-signed certificate
jarvis certificate --cert-dir ./certs

# Verify certificate generation
ls -la ./certs/
```

## Directory Structure

After setup, your project structure should look like:

```
your-project/
├── config.yaml              # Configuration file
├── certs/                   # TLS certificates (if using HTTPS)
│   ├── server.crt
│   └── server.key
├── data/                    # Database and logs
│   └── traffic_inspector.db
├── output/                  # Generated test files
├── specs/                   # API specifications
│   ├── openapi/
│   └── proto/
└── jarvis                   # Jarvis binary
```

## Verification

Test your installation:

```bash
# Check version
jarvis version

# Run health check
jarvis proxy --timeout 1 &  # Start proxy in background
curl http://localhost:8080/health  # Test proxy
pkill jarvis  # Stop proxy

# Test AI integration
jarvis gen generate-scenarios --path="./specs/openapi/example.yaml"
```

## Troubleshooting

### Common Issues

1. **Ollama not found**: Ensure Ollama is installed and running on port 11434
2. **Permission denied**: Make sure the binary is executable (`chmod +x jarvis`)
3. **Port conflicts**: Change ports in config.yaml if defaults are occupied
4. **Certificate errors**: Regenerate certificates with `jarvis certificate`

### Debug Mode

Enable debug logging for troubleshooting:

```bash
jarvis --debug proxy
jarvis --debug gen generate-test --path="./specs"
```

### Log Files

Check log files for detailed error information:
- Proxy logs: Check console output when running `jarvis proxy`
- Database logs: SQLite errors are logged to console
- AI generation logs: Displayed during generation commands

## Next Steps

After setup:
1. Review the [Example Usage](example.md) documentation
2. Explore the [Design Document](design.md) for architecture details
3. Try generating your first test scenarios
4. Set up traffic inspection for your API

