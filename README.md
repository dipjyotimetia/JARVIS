# Jarvis

🧠 An AI-powered CLI tool for intelligent API testing and traffic inspection

<img src="docs/assets/jarvis.jpg" width="400">

[![goreleaser](https://github.com/dipjyotimetia/jarvis/actions/workflows/release.yml/badge.svg)](https://github.com/dipjyotimetia/jarvis/actions/workflows/release.yml)

## Overview

Jarvis is a comprehensive CLI tool that leverages Ollama AI models to revolutionize API testing workflows. It combines intelligent test generation capabilities with HTTP/HTTPS traffic inspection, certificate management, and interactive setup wizards to streamline development and testing processes.

## Features

### 🤖 AI-Powered Test Generation
- **API Spec Analysis**: Generate comprehensive test scenarios from OpenAPI and Protobuf specifications
- **Contract Testing**: Generate Pact contracts for consumer-driven contract testing
- **Ollama Integration**: Leverage local AI models for intelligent test case creation
- **File Processing**: Process specification files to identify edge cases and testing requirements

### 🔍 Advanced Traffic Inspector
- **HTTP/HTTPS Proxy**: Record, replay, and analyze API traffic with TLS and mTLS support
- **Multiple Modes**: Record mode, replay mode, and passthrough mode
- **Path-Based Routing**: Route different API paths to different target servers
- **Interactive Web UI**: Review captured traffic through a clean web interface (default port: 9090)
- **OpenAPI Validation**: Real-time API validation against OpenAPI specifications
- **Certificate Management**: Built-in self-signed certificate generation

### 🔧 Developer Tools
- **Interactive Setup Wizard**: Step-by-step configuration with language and framework preferences
- **Spec Analysis**: Deep analysis of Protobuf and OpenAPI specifications
- **gRPC Tools**: Generate gRPC curl commands for service testing
- **Multi-Language Support**: JavaScript, TypeScript, Python, Java, Go, and more

### 🔄 Integration & Automation
- **Jira & Confluence**: Connect to your existing documentation and issue tracking
- **GitHub**: Stay updated with automatic version checks
- **Customizable Output**: Generate output in formats that suit your workflow
- **Template System**: Pre-built templates for various languages and frameworks

## Installation

```bash
# Download the latest release for your platform from:
# https://github.com/dipjyotimetia/jarvis/releases

# Make it executable (Linux/macOS)
chmod +x jarvis

# Move to a directory in your PATH (Linux/macOS)
sudo mv jarvis /usr/local/bin/
```

## Quick Start

```bash
# Interactive setup wizard
jarvis setup

# Check version and updates
jarvis version

# Start the traffic inspector proxy (HTTP + Web UI)
jarvis proxy

# Start HTTPS proxy with TLS
jarvis proxy --tls --cert ./certs/server.crt --key ./certs/server.key --tls-port=8443

# Generate self-signed certificates
jarvis certificate --cert-dir ./certs

# Generate test scenarios from OpenAPI spec  
jarvis gen generate-scenarios --path="specs/openapi/v3.0/my_api.yaml"

# Generate test cases from Protobuf spec
jarvis gen generate-test --path="specs/proto" --output="output"

# Generate Pact contracts from OpenAPI spec
jarvis gen generate-contracts --path="specs/openapi/v3.0/my_api.yaml" --consumer="web-app" --provider="api-service"

# Generate Pact contracts with test code
jarvis gen generate-contracts --path="specs/openapi" --consumer="mobile-app" --provider="backend-api" --language="javascript" --framework="jest" --examples

# Analyze API specifications
jarvis analyze spec-analyzer --path="specs/proto"

# Generate gRPC curl commands
jarvis tools grpc-curl --proto="user.proto" --service="UserService" --method="GetUser"
```

## Pact Contract Generation

Jarvis can generate Pact contracts from OpenAPI specifications using AI, helping you implement consumer-driven contract testing.

### Features
- **AI-Generated Contracts**: Creates realistic Pact contracts from OpenAPI specs
- **Multi-Language Support**: Generates test code for JavaScript, Python, Java, Go
- **Framework Integration**: Supports Jest, Pytest, JUnit, Go testing
- **Smart Validation**: Comprehensive validation with helpful suggestions
- **Template System**: Pre-built templates for common languages and frameworks

### Usage Examples

```bash
# Basic contract generation
jarvis gen generate-contracts \
  --path="api-spec.yaml" \
  --consumer="web-frontend" \
  --provider="user-service"

# Generate with test code
jarvis gen generate-contracts \
  --path="api-spec.yaml" \
  --consumer="mobile-app" \
  --provider="backend-api" \
  --language="javascript" \
  --framework="jest" \
  --examples

# Custom output directory
jarvis gen generate-contracts \
  --path="specs/openapi/" \
  --consumer="client" \
  --provider="server" \
  --output="./pact-contracts"
```

### Supported Languages & Frameworks
- **JavaScript**: Jest, Mocha
- **Python**: Pytest, unittest
- **Java**: JUnit, TestNG
- **Go**: testing package
- **TypeScript**: Jest, Mocha

### Generated Files
- `{consumer}-{provider}-pact.json`: Pact contract file
- `{consumer}_{provider}_test.{ext}`: Test code (when `--examples` is used)

## Traffic Inspector Proxy

The proxy component provides powerful HTTP/HTTPS traffic inspection capabilities:

### Basic Proxy Usage
```bash
# Start basic HTTP proxy on port 8080
jarvis proxy

# Start with custom ports
jarvis proxy --ui-port=9999

# Recording mode - capture all traffic
jarvis proxy --record

# Replay mode - replay captured traffic
jarvis proxy --replay
```

### HTTPS/TLS Support
```bash
# Generate self-signed certificates
jarvis certificate --cert-dir ./certs

# Start HTTPS proxy
jarvis proxy --tls --cert ./certs/server.crt --key ./certs/server.key --tls-port=8443

# Enable mutual TLS (mTLS)
jarvis proxy --mtls --client-ca ./certs/ca.crt --client-cert ./certs/client.crt --client-key ./certs/client.key
```

### OpenAPI Validation
```bash
# Enable API validation against OpenAPI spec
jarvis proxy --api-validate --api-spec ./specs/api.yaml

# Strict validation mode
jarvis proxy --api-validate --api-spec ./specs/api.yaml --strict-validation

# Validate requests only
jarvis proxy --api-validate --api-spec ./specs/api.yaml --validate-req --validate-resp=false
```

### Web UI
- Access the web interface at `http://localhost:9090/ui/` (default)
- View captured requests and responses
- Analyze traffic patterns and API behavior
- Export data for further analysis

## Command Structure

Jarvis uses a structured command hierarchy:

```
jarvis
├── setup                    # Interactive setup wizard
├── version                  # Version information and updates
├── certificate             # Certificate generation
├── proxy                   # Traffic inspector proxy
├── gen                     # Generation commands
│   ├── generate-test       # Generate test cases
│   ├── generate-scenarios  # Generate test scenarios  
│   └── generate-contracts  # Generate Pact contracts
├── analyze                 # Analysis commands
│   └── spec-analyzer       # Analyze API specifications
└── tools                   # Utility tools
    └── grpc-curl           # Generate gRPC curl commands
```

## Documentation

- [Design Document](docs/design.md) - Architecture and design decisions
- [Setup Guide](docs/setup.md) - Detailed setup and configuration
- [Example Usage](docs/example.md) - Comprehensive usage examples

## Configuration

Jarvis can be configured via command-line flags, a config file, or the interactive setup wizard:

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `http_port` | Port for the HTTP proxy server | 8080 |
| `ui_port` | Port for the web UI | 9090 |
| `http_target_url` | Default target URL for proxying | (required) |
| `target_routes` | Array of path-based routing rules | [] |
| `sqlite_db_path` | Path to SQLite database file | traffic_inspector.db |
| `recording_mode` | Enable traffic recording | false |
| `replay_mode` | Enable traffic replay | false |
| `tls.enabled` | Enable HTTPS support | false |
| `tls.port` | HTTPS port | 8443 |
| `tls.cert_file` | TLS certificate file path | "" |
| `tls.key_file` | TLS private key file path | "" |
| `api_validation.enabled` | Enable OpenAPI validation | false |
| `api_validation.spec_path` | OpenAPI specification file path | "" |

### Configuration File Example

```yaml
# config.yaml
http_port: 8080
ui_port: 9090
sqlite_db_path: "./data/traffic_inspector.db"
recording_mode: false
replay_mode: false

tls:
  enabled: true
  port: 8443
  cert_file: "./certs/server.crt"
  key_file: "./certs/server.key"
  client_auth: false

api_validation:
  enabled: true
  spec_path: "./specs/api.yaml"
  validate_requests: true
  validate_responses: true
  strict_mode: false

language:
  preferred: "javascript"
  framework: "jest"

output:
  directory: "./output"
  report_format: "HTML"
```

## Contributing

We welcome contributions! Please see our PR template for more details.

## License

MIT