# Jarvis

🧠 An AI-powered CLI tool for intelligent API testing and traffic inspection

<img src="docs/assets/jarvis.jpg" width="400">

[![goreleaser](https://github.com/dipjyotimetia/jarvis/actions/workflows/release.yml/badge.svg)](https://github.com/dipjyotimetia/jarvis/actions/workflows/release.yml)

## Overview

Jarvis is a powerful CLI tool that leverages Open Weight AI models to revolutionize API testing workflows. It combines intelligent test generation capabilities with HTTP traffic inspection to streamline development and testing processes.

## Features

### 🤖 AI-Powered Test Generation
- **API Spec Analysis**: Generate comprehensive test scenarios from OpenAPI and Protobuf specifications
- **Contract Testing**: Generate Pact contracts for consumer-driven contract testing
- **Ollama Integration**: Leverage local AI models for intelligent test case creation
- **File Processing**: Process specification files to identify edge cases and testing requirements

### 🔍 Traffic Inspector
- **HTTP Proxy**: Record, replay, and analyze API traffic with minimal configuration
- **Multiple Modes**: Record mode, replay mode, and passthrough mode
- **Path-Based Routing**: Route different API paths to different target servers
- **Interactive UI**: Review captured traffic through a clean web interface

### 🔄 Integration
- **Jira & Confluence**: Connect to your existing documentation and issue tracking
- **GitHub**: Stay updated with automatic version checks
- **Customizable Output**: Generate output in formats that suit your workflow

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
# Check version
jarvis version

# Generate test scenarios from OpenAPI spec
jarvis generate-scenarios --path="specs/openapi/v3.0/my_api.yaml"

# Generate test scenarios from Protobuf spec
jarvis generate-test --path="specs/proto" --output="output"

# Generate Pact contracts from OpenAPI spec
jarvis generate-contracts --path="specs/openapi/v3.0/my_api.yaml" --consumer="web-app" --provider="api-service"

# Generate Pact contracts with test code
jarvis generate-contracts --path="specs/openapi" --consumer="mobile-app" --provider="backend-api" --language="javascript" --framework="jest" --examples

# Start the traffic inspector proxy
jarvis proxy --port=8080
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

## Documentation

- Design Document
- Setup Guide
- Example Usage

## Configuration

Jarvis can be configured via command-line flags or a config file:

| Option | Description | Default |
|--------|-------------|---------|
| `http_port` | Port for the HTTP proxy server | 8080 |
| `http_target_url` | Default target URL for proxying | (required) |
| `target_routes` | Array of path-based routing rules | [] |
| `sqlite_db_path` | Path to SQLite database file | traffic_inspector.db |

## Contributing

We welcome contributions! Please see our PR template for more details.

## License

MIT