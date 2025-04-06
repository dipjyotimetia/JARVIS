# Jarvis

🧠 An AI-powered CLI tool for intelligent API testing and traffic inspection

<img src="docs/assets/jarvis.jpg" width="400">

[![goreleaser](https://github.com/dipjyotimetia/jarvis/actions/workflows/release.yml/badge.svg)](https://github.com/dipjyotimetia/jarvis/actions/workflows/release.yml)

## Overview

Jarvis is a powerful CLI tool that leverages Google's Gemini AI models to revolutionize API testing workflows. It combines intelligent test generation capabilities with HTTP traffic inspection to streamline development and testing processes.

## Features

### 🤖 AI-Powered Test Generation
- **API Spec Analysis**: Generate comprehensive test scenarios from OpenAPI and Protobuf specifications
- **Gemini Integration**: Utilize Google's advanced AI models for intelligent test case creation
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

# Start the traffic inspector proxy
jarvis proxy --port=8080
```

## Gemini API Setup

Jarvis requires a Google Gemini API key to utilize AI features:

1. Visit [Google AI Studio](https://ai.google.dev/)
2. Create an API key
3. Set as environment variable:
   ```bash
   # Linux/macOS
   export API_KEY="your_api_key"
   
   # Windows
   $Env:API_KEY = "your_api_key"
   ```

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