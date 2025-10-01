# JARVIS

An AI-powered CLI tool for intelligent API testing and traffic inspection

<img src="docs/assets/jarvis_updated.png" width="400">

[![goreleaser](https://github.com/dipjyotimetia/jarvis/actions/workflows/release.yml/badge.svg)](https://github.com/dipjyotimetia/jarvis/actions/workflows/release.yml)

## Overview

Jarvis is a comprehensive CLI tool that leverages **multiple AI providers** (OpenAI, Anthropic, Google, Ollama) to revolutionize API testing workflows. It combines intelligent test generation capabilities with HTTP/HTTPS traffic inspection, certificate management, and interactive setup wizards to streamline development and testing processes.

## 🆕 What's New

- **🚀 Multi-LLM Support**: Choose between OpenAI, Anthropic Claude, Google Gemini, or Ollama
- **🧪 AI Test Data Generator**: Generate realistic test data from OpenAPI specifications
- **🔍 AI Failure Analyzer**: Intelligent diagnosis of API failures with root cause analysis
- **💰 Cost Optimization**: Use free local models (Ollama) or premium cloud models as needed
- **🔄 Provider Flexibility**: Switch between AI providers without code changes

## Features

### 🤖 AI-Powered Test Generation

- **Multiple AI Providers**: OpenAI GPT-4, Anthropic Claude, Google Gemini, or local Ollama models
- **Test Data Generation**: Generate realistic test data from OpenAPI schemas
- **Test Scenarios**: Generate comprehensive test scenarios from API specifications
- **Failure Analysis**: AI-powered diagnosis of API failures with actionable insights
- **Contract Testing**: Generate Pact contracts for consumer-driven contract testing
- **Smart Analysis**: Identify edge cases, boundary conditions, and validation requirements

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

### AI Provider Setup

Choose one or more AI providers:

```bash
# Option 1: Ollama (FREE, Local, Private)
ollama pull llama3.2

# Option 2: OpenAI (Cloud, High Quality)
export OPENAI_API_KEY="sk-..."

# Option 3: Anthropic Claude (Cloud, Best Reasoning)
export ANTHROPIC_API_KEY="sk-ant-..."

# Option 4: Google Gemini (Cloud, Cost-Effective)
export GOOGLE_API_KEY="AIza..."
```

See [Quick Start Guide](docs/QUICK_START_LLM.md) for detailed setup.

### Core Features

```bash
# Interactive setup wizard
jarvis setup

# Generate realistic test data from OpenAPI spec
jarvis gen generate-testdata --spec api.yaml --count 10

# Analyze API failures with AI
jarvis analyze analyze-failures --limit 5

# Generate test scenarios from OpenAPI spec
jarvis gen generate-scenarios --path="specs/openapi/api.yaml"

# Start traffic inspector proxy
jarvis proxy --record

# Generate self-signed certificates
jarvis certificate --cert-dir ./certs
```

## AI-Powered Features

### 1. Test Data Generation

Generate realistic, schema-compliant test data:

```bash
# Basic usage
jarvis gen generate-testdata --spec api.yaml

# Advanced options
jarvis gen generate-testdata \
  --spec api.yaml \
  --count 10 \
  --locale en-GB \
  --output ./testdata

# Generate from JSON schema
jarvis gen generate-from-schema --schema user.schema.json
```

**Features:**
- ✅ Valid and invalid test cases
- ✅ Boundary value testing
- ✅ Schema-aware generation
- ✅ Multiple locales support
- ✅ Edge case identification

### 2. Failure Analysis

AI-powered diagnosis of API failures:

```bash
# Analyze recent failures
jarvis analyze analyze-failures

# Analyze specific endpoint
jarvis analyze analyze-endpoint /api/users

# Filter by status code
jarvis analyze analyze-failures --status 500

# Save detailed report
jarvis analyze analyze-failures --output report.json
```

**Analysis Includes:**
- 🎯 Root cause identification
- 📊 Error categorization
- 💡 Actionable fix suggestions
- 🔄 Reproduction steps
- 🛡️ Prevention tips
- 📚 Related documentation

### 3. Test Scenarios

Generate comprehensive test scenarios:

```bash
# Generate scenarios from spec
jarvis gen generate-scenarios --path specs/api.yaml

# Generate test cases with code
jarvis gen generate-test --path specs/ --output tests/
```

## Traffic Inspector Proxy

### Basic Usage

```bash
# Start HTTP proxy on port 8080
jarvis proxy

# Recording mode - capture all traffic
jarvis proxy --record

# Replay mode - replay captured traffic
jarvis proxy --replay

# Start with custom ports
jarvis proxy --ui-port=9999
```

### HTTPS/TLS Support

```bash
# Generate self-signed certificates
jarvis certificate --cert-dir ./certs

# Start HTTPS proxy
jarvis proxy --tls --cert ./certs/server.crt --key ./certs/server.key --tls-port=8443

# Enable mutual TLS (mTLS)
jarvis proxy --mtls --client-ca ./certs/ca.crt
```

### OpenAPI Validation

```bash
# Enable API validation
jarvis proxy --api-validate --api-spec ./specs/api.yaml

# Strict validation mode
jarvis proxy --api-validate --api-spec ./specs/api.yaml --strict-validation
```

### Web UI

- Access at `http://localhost:9090/ui/`
- View captured requests and responses
- Analyze traffic patterns
- Export data for analysis

## Command Structure

```
jarvis
├── setup                        # Interactive setup wizard
├── version                      # Version information
├── certificate                  # Certificate generation
├── proxy                        # Traffic inspector proxy
├── gen                          # Generation commands
│   ├── generate-test            # Generate test cases
│   ├── generate-scenarios       # Generate test scenarios
│   ├── generate-testdata        # Generate test data (NEW)
│   └── generate-from-schema     # Generate from JSON schema (NEW)
├── analyze                      # Analysis commands
│   ├── spec-analyzer            # Analyze API specifications
│   ├── analyze-failures         # AI failure analysis (NEW)
│   └── analyze-endpoint         # Endpoint failure analysis (NEW)
└── tools                        # Utility tools
    └── grpc-curl                # Generate gRPC curl commands
```

## Configuration

### AI Provider Configuration

```bash
# Auto-detect provider (priority: OpenAI → Anthropic → Google → Ollama)
# Just set an API key

# Or explicitly set provider
export LLM_PROVIDER=openai|anthropic|google|ollama
export LLM_MODEL=gpt-4o-mini
export LLM_TEMPERATURE=0.7
export LLM_MAX_TOKENS=2048
```

### Proxy Configuration

Jarvis can be configured via command-line flags, config file, or environment variables:

**Priority:** Command-line flags > Environment variables > Config file > Defaults

```bash
# Environment variables
export JARVIS_HTTP_PORT=8080
export JARVIS_UI_PORT=9090
export JARVIS_TLS_ENABLED=true
```

**Config file example:**

```yaml
# config.yaml
http_port: 8080
ui_port: 9090
recording_mode: false

tls:
  enabled: true
  port: 8443
  cert_file: "./certs/server.crt"
  key_file: "./certs/server.key"

api_validation:
  enabled: true
  spec_path: "./specs/api.yaml"
```

## Documentation

### Getting Started
- [Quick Start Guide](docs/QUICK_START_LLM.md) - Get started in 5 minutes
- [Setup Guide](docs/setup.md) - Detailed setup and configuration
- [Example Usage](docs/example.md) - Comprehensive usage examples

### AI Features
- [LLM Integration Guide](docs/LLM_INTEGRATION.md) - Multi-provider setup and configuration
- [New Features Guide](docs/NEW_FEATURES.md) - Test data generation and failure analysis
- [Migration Summary](docs/MIGRATION_SUMMARY.md) - Migrating from Ollama-only

### Advanced
- [Design Document](docs/design.md) - Architecture and design decisions
- [Implementation Summary](IMPLEMENTATION_SUMMARY.md) - Technical implementation details

## Use Cases

### Development Workflow

```bash
# Use free local models
export LLM_PROVIDER=ollama
ollama pull llama3.2

# Generate test data
jarvis gen generate-testdata --spec api.yaml --count 20

# Run tests, record traffic
jarvis proxy --record

# Analyze failures
jarvis analyze analyze-failures
```

### CI/CD Workflow

```bash
# Use cost-effective cloud model
export LLM_PROVIDER=google
export GOOGLE_API_KEY="$GOOGLE_API_KEY"

# Generate and validate
jarvis gen generate-testdata --spec api.yaml
jarvis proxy --api-validate --api-spec api.yaml
```

### Production Workflow

```bash
# Use high-quality models for critical analysis
export LLM_PROVIDER=anthropic
export ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY"

# Deep failure analysis
jarvis analyze analyze-failures --limit 10 --output report.json
```

## Examples

### Complete Testing Workflow

```bash
# 1. Generate test data
jarvis gen generate-testdata --spec petstore.yaml --output testdata/

# 2. Start proxy in record mode
jarvis proxy --record --api-validate --api-spec petstore.yaml &

# 3. Run your tests (using generated test data)
# npm test / pytest / go test

# 4. Analyze failures
jarvis analyze analyze-failures --time-window 1h --output failures.json

# 5. Review report
cat failures.json | jq '.[] | {endpoint: .url, cause: .analysis.root_cause}'
```

### Multi-Provider Usage

```bash
# Generate test data locally (free)
export LLM_PROVIDER=ollama
jarvis gen generate-testdata --spec api.yaml --count 50

# Analyze failures with Claude (best quality)
export LLM_PROVIDER=anthropic
export ANTHROPIC_API_KEY="sk-ant-..."
jarvis analyze analyze-failures --limit 5

# Generate scenarios with Gemini (cost-effective)
export LLM_PROVIDER=google
export GOOGLE_API_KEY="AIza..."
jarvis gen generate-scenarios --path specs/
```

## Performance

- **Binary Size**: ~48 MB
- **Memory Usage**: ~100 MB (base) + AI model overhead
- **AI Response Time**:
  - Local (Ollama): 1-5 seconds
  - Cloud: 0.5-3 seconds
- **Proxy Throughput**: 1000+ req/sec

## Contributing

We welcome contributions! Please see our [contributing guidelines](CONTRIBUTING.md).

### Development Setup

```bash
# Clone repository
git clone https://github.com/dipjyotimetia/jarvis.git
cd jarvis

# Install dependencies
go mod download

# Build
go build -o jarvis

# Run tests
go test ./...
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: See `docs/` directory
- **Issues**: https://github.com/dipjyotimetia/jarvis/issues
- **Discussions**: https://github.com/dipjyotimetia/jarvis/discussions

---

**Made with ❤️ for API testers and developers**
