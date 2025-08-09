# Jarvis Design Document

## Project Overview

Jarvis is a comprehensive AI-powered CLI tool that combines intelligent API testing capabilities with advanced HTTP/HTTPS traffic inspection. It leverages Ollama AI models to revolutionize API testing workflows and provides sophisticated proxy functionality for development and testing environments.

The system is designed around three core pillars:

1. **AI-Powered Generation**: Uses local Ollama models to generate test scenarios, test code, and Pact contracts from API specifications
2. **Advanced Traffic Inspection**: Provides HTTP/HTTPS proxy with TLS, mTLS, and OpenAPI validation capabilities
3. **Developer Experience**: Interactive setup, comprehensive tooling, and seamless integration with existing workflows

## System Architecture

The system follows a modular design with clear separation of concerns:

### Core Components

1. **Command Line Interface** (`cmd/`):
   - Root command with structured subcommands (gen, analyze, tools, proxy, etc.)
   - Interactive setup wizard with step-by-step configuration
   - Version management with GitHub integration for updates
   - Certificate generation for TLS/HTTPS support

2. **AI Generation Engine** (`pkg/engine/`):
   - **Ollama Integration**: Local AI model client for code and scenario generation
   - **Pact Generator**: Advanced contract testing with AI-powered interaction generation
   - **Prompt Engineering**: Sophisticated prompt templates for different output types
   - **Multi-language Support**: Template system for JavaScript, Python, Java, Go, etc.

3. **Traffic Inspector** (`internal/proxy/`, `internal/web/`):
   - **HTTP/HTTPS Proxy**: Advanced proxy with TLS and mTLS support
   - **OpenAPI Validation**: Real-time API request/response validation
   - **Web UI**: Interactive dashboard for traffic analysis
   - **Recording & Replay**: Sophisticated traffic capture and playback

4. **Configuration & Storage**:
   - **Config Management** (`config/`): YAML-based configuration with environment variable support
   - **Database Layer** (`internal/db/`): SQLite with WAL for traffic persistence
   - **Certificate Management** (`internal/certs/`): Self-signed certificate generation

5. **Analysis & Utilities**:
   - **Spec Analyzers** (`pkg/engine/utils/`): Protobuf and OpenAPI specification analysis
   - **File Processing** (`pkg/engine/files/`): Multi-format file handling and parsing
   - **Integration Tools**: GitHub, Jira, and Confluence connectivity

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                 Jarvis CLI                                      │
├─────────────────┬─────────────────┬─────────────────┬─────────────────────────┤
│   Generation    │    Analysis     │     Tools       │       Proxy             │
│   Commands      │    Commands     │   Commands      │      Commands           │
│                 │                 │                 │                         │
│ • gen test      │ • analyze spec  │ • tools grpc    │ • proxy (HTTP/HTTPS)    │
│ • gen scenarios │                 │                 │ • certificate           │
│ • gen contracts │                 │                 │                         │
└─────────────────┴─────────────────┴─────────────────┴─────────────────────────┘
         │                   │                 │                     │
         ▼                   ▼                 ▼                     ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────┐ ┌─────────────────────┐
│   AI Engine     │ │ Spec Analyzers  │ │  Utilities  │ │  Traffic Inspector  │
│                 │ │                 │ │             │ │                     │
│ • Ollama Client │ │ • Proto Parser  │ │ • gRPC Gen  │ │ • HTTP Proxy        │
│ • Pact Gen      │ │ • OpenAPI Parse │ │ • File Walk │ │ • HTTPS/TLS/mTLS    │
│ • Templates     │ │ • Validation    │ │             │ │ • OpenAPI Validate  │
│ • Prompts       │ │                 │ │             │ │ • Web UI            │
└─────────────────┘ └─────────────────┘ └─────────────┘ └─────────────────────┘
         │                   │                             │
         ▼                   ▼                             ▼
┌─────────────────┐ ┌─────────────────┐           ┌─────────────────────┐
│  Ollama Server  │ │ Spec Files      │           │   Traffic Storage   │
│  (localhost)    │ │                 │           │                     │
│                 │ │ • OpenAPI YAML  │           │ • SQLite DB         │
│ • llama2        │ │ • Protobuf      │           │ • Request/Response  │
│ • codellama     │ │ • Swagger       │           │ • Replay Data       │
│ • Custom Models │ │ • Avro          │           │ • Web UI Assets     │
└─────────────────┘ └─────────────────┘           └─────────────────────┘
```

### Traffic Inspector Flow

```
┌─────────────┐     ┌──────────────────────┐     ┌─────────────────┐     ┌──────────────┐
│   Client    │────▶│   Jarvis Proxy      │────▶│  Target Server  │────▶│   Response   │
│   Request   │     │                      │     │                 │     │              │
└─────────────┘     │ • TLS Termination    │     └─────────────────┘     └──────────────┘
                    │ • OpenAPI Validation │               │                      │
                    │ • Path Routing       │               │                      │
                    │ • Request/Response   │               │                      │
                    │   Recording          │               │                      │
                    └──────────────────────┘               │                      │
                             │                             │                      │
                             │  ┌─────────────────────────┘                      │
                             │  │                                                 │
                             ▼  ▼                                                 │
                    ┌──────────────────────┐                                     │
                    │   SQLite Database    │◀────────────────────────────────────┘
                    │                      │      Store Request/Response Data
                    │ • Traffic Records    │
                    │ • Validation Results │
                    │ • Performance Metrics│
                    └──────────────────────┘
                             │
                             ▼
                    ┌──────────────────────┐
                    │     Web UI           │
                    │                      │
                    │ • Traffic Analysis   │     ┌─────────────────┐
                    │ • Request Explorer   │────▶│ Developer       │
                    │ • Response Viewer    │     │ Browser         │
                    │ • API Documentation  │     │ (localhost:9090)│
                    └──────────────────────┘     └─────────────────┘
```

### Path-Based Routing

Traffic Inspector supports sophisticated path-based routing, allowing different request paths to be proxied to different target servers:

```
/api/users/* ─────────▶ https://api.example.com/users
/api/products/* ─────▶ https://products.example.org/api
/health/* ───────────▶ http://localhost:8081
/* (default) ────────▶ https://jsonplaceholder.typicode.com
```

## Technical Implementation

### Key Features

#### AI-Powered Generation
1. **Local AI Processing**: Uses Ollama for privacy-preserving, offline AI generation
2. **Intelligent Prompt Engineering**: Context-aware prompts for different spec types and output formats
3. **Multi-format Support**: Handles OpenAPI (YAML/JSON), Protobuf, Swagger, and Avro specifications
4. **Template System**: Extensible template engine for multiple programming languages and frameworks
5. **Contract Testing**: Advanced Pact contract generation with AI-powered interaction creation

#### Advanced Proxy Capabilities
1. **TLS/HTTPS Support**: Full TLS 1.2/1.3 support with custom certificate management
2. **Mutual TLS (mTLS)**: Client certificate authentication for secure B2B communications
3. **OpenAPI Validation**: Real-time request/response validation against API specifications
4. **Intelligent Routing**: Path-based routing with wildcard support and fallback mechanisms
5. **Performance Optimization**: Efficient buffer management with sync.Pool and connection pooling

#### Developer Experience
1. **Interactive Setup**: Guided configuration wizard with sensible defaults
2. **Graceful Shutdown**: Proper signal handling and resource cleanup
3. **Comprehensive Logging**: Structured logging with debug modes and different verbosity levels
4. **Web Interface**: Rich web UI for traffic analysis and API exploration
5. **Cross-platform**: Native binaries for Linux, macOS, and Windows

### Operating Modes

#### AI Generation Modes
1. **Test Scenario Generation**: Creates comprehensive test cases from API specifications
   - Analyzes API endpoints and generates positive, negative, and edge cases
   - Considers authentication, rate limiting, and error handling
   - Outputs human-readable test scenarios

2. **Test Code Generation**: Produces executable test code in multiple languages
   - Supports JavaScript (Jest), Python (pytest), Java (JUnit), Go (testing)
   - Includes setup, teardown, and assertion logic
   - Handles complex data structures and API authentication

3. **Pact Contract Generation**: Creates consumer-driven contract tests
   - AI-powered interaction generation based on OpenAPI specs
   - Supports multiple consumer-provider relationships
   - Includes validation and suggestion features

#### Proxy Operating Modes

1. **Passthrough Mode** (Default):
   - Acts as transparent reverse proxy
   - Forwards requests to configured target URLs
   - Minimal latency with optional logging

2. **Recording Mode**:
   - Captures all request/response data with full fidelity
   - Stores in SQLite with metadata (headers, timing, validation results)
   - Continues proxying to actual servers for real-time testing

3. **Replay Mode**:
   - Serves recorded responses from database
   - Supports request matching with flexible criteria
   - Falls back to 404 for unmatched requests with option to forward

4. **Validation Mode**:
   - Real-time OpenAPI specification validation
   - Validates both requests and responses
   - Provides detailed error reporting and suggestions

## Architecture Decisions

### Why Ollama?
1. **Privacy**: All AI processing happens locally, ensuring sensitive API specifications never leave your environment
2. **Performance**: Direct integration without external API calls or rate limits
3. **Flexibility**: Support for multiple models (llama2, codellama, custom models)
4. **Cost**: No per-request costs or subscription fees

### Why SQLite?
1. **Simplicity**: Single-file database with no setup required
2. **Performance**: WAL mode provides excellent concurrency for proxy traffic
3. **Portability**: Database files can be easily shared and backed up
4. **Reliability**: ACID compliance ensures data integrity

### Why Go?
1. **Performance**: Excellent for concurrent proxy operations
2. **Cross-platform**: Single binary deployment across all platforms
3. **Networking**: Superior HTTP/TLS support with standard library
4. **Tooling**: Rich ecosystem for CLI applications and testing

## Security Considerations

### TLS/Certificate Management
- Self-signed certificates for development and testing
- Support for custom CA certificates and certificate chains
- Proper certificate validation with configurable strictness
- Secure private key handling and storage

### Data Privacy
- All AI processing happens locally via Ollama
- Traffic data stored locally in SQLite database
- No external API calls for AI generation
- Configurable data retention policies

### API Security
- Request/response validation against OpenAPI specifications
- Support for various authentication methods (Bearer tokens, API keys, etc.)
- Rate limiting and abuse prevention
- Secure proxy forwarding with header sanitization

## Performance Characteristics

### AI Generation Performance
- Local processing eliminates network latency
- Model loading is a one-time cost per session
- Generation speed depends on model size and complexity
- Typical scenarios: 2-10 seconds for comprehensive test generation

### Proxy Performance
- Low-latency forwarding with connection pooling
- Efficient memory management with buffer reuse
- Concurrent request handling with Go's goroutines
- Database operations optimized with prepared statements

### Scalability
- Handles hundreds of concurrent proxy connections
- SQLite WAL mode supports multiple readers
- Memory usage scales linearly with active connections
- Configurable timeouts and resource limits

## Extension Points

### Custom AI Models
- Support for specialized models via Ollama
- Custom prompt templates for domain-specific generation
- Model selection based on specification type or complexity

### Plugin Architecture
- Extensible template system for new languages/frameworks
- Custom validation rules for API specifications
- Integration hooks for external tools and services

### Custom Protocols
- Framework for adding new specification format support
- Extensible proxy middleware for custom processing
- Integration points for external validation services

## Monitoring and Observability

### Logging
- Structured logging with configurable levels
- Request/response logging with sanitization
- Performance metrics and timing information
- Error tracking and debugging information

### Metrics
- Proxy performance metrics (latency, throughput, error rates)
- AI generation metrics (processing time, model performance)
- Database operation metrics (query performance, storage usage)
- System resource utilization

### Web UI Analytics
- Traffic pattern analysis and visualization
- API endpoint usage statistics
- Error rate tracking and alerting
- Performance trending and capacity planning

## Future Roadmap

### Planned Features
1. **Advanced AI Models**: Support for specialized testing models
2. **Cloud Integration**: Optional cloud model support alongside local processing
3. **API Mocking**: Advanced mock server capabilities based on specifications
4. **Load Testing**: Integration with load testing frameworks
5. **CI/CD Integration**: Plugins for popular CI/CD platforms

### Extensibility Goals
1. **Custom Generators**: Plugin system for custom test generators
2. **Advanced Analytics**: Enhanced traffic analysis and reporting
3. **Multi-protocol Support**: gRPC, GraphQL, and WebSocket proxy support
4. **Distributed Testing**: Multi-node proxy deployment for large-scale testing

This design ensures Jarvis remains a comprehensive, privacy-focused, and extensible platform for modern API testing and development workflows.