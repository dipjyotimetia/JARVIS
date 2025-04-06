Collecting workspace information# Traffic Inspector Design Document

## Project Overview

Traffic Inspector is a transparent HTTP proxy tool designed to record, replay, and passthrough HTTP traffic. It acts as an intermediary between clients and servers, allowing developers to:

1. **Record Mode**: Capture all HTTP traffic for later analysis or replay
2. **Replay Mode**: Serve recorded responses instead of forwarding requests to the actual server
3. **Passthrough Mode**: Simply forward requests to target servers with minimal interference

## System Architecture

The system follows a modular design with clear separation of concerns:

### Components

1. **Command Line Interface** (`cmd/`): Provides user interface through CLI commands
2. **Configuration** (`config/`): Handles loading and validating application settings
3. **HTTP Proxy** (`internal/proxy/`): Core proxy functionality for intercepting and forwarding HTTP requests
4. **Storage** (`internal/db/`): SQLite database for persisting traffic records

### Workflow Diagram

```
┌─────────────┐        ┌────────────────┐       ┌─────────────────┐       ┌──────────────┐
│             │        │                │       │                 │       │              │
│   Client    │━━━━━━━▶│  HTTP Proxy    │━━━━━━▶│  Target Server  │━━━━━━▶│   Response   │
│   Request   │        │  (Inspector)   │       │                 │       │              │
│             │        │                │       │                 │       │              │
└─────────────┘        └────────────────┘       └─────────────────┘       └──────────────┘
                               │                                                 │
                               ▼                                                 │
                        ┌────────────────┐                                      │
                        │                │                                      │
                        │  Traffic DB    │◀━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
                        │  (SQLite)     │       Record Mode: Store request/response
                        │                │
                        └────────────────┘
                               │
                               ▼
                        ┌────────────────┐
                        │                │
                        │  Replay Mode   │
                        │  (Serve from   │
                        │   database)    │
                        │                │
                        └────────────────┘
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

1. **Efficient Buffer Management**: Uses sync.Pool for efficient buffer reuse and memory management
2. **Graceful Shutdown**: Properly handles process termination with context cancellation
3. **Path-based Routing**: Routes traffic to different backend servers based on URL path prefixes
4. **SQLite with WAL**: Uses Write-Ahead Logging for better concurrency
5. **Customizable Transport**: Configurable HTTP transport with timeout settings

### Operating Modes

#### Recording Mode
- Captures both request and response data
- Stores in SQLite database with metadata (headers, body, timing, etc.)
- Continues to proxy requests to actual servers

#### Replay Mode
- Queries database for matching requests
- Returns recorded responses without contacting backend servers
- Falls back to 404 if no matching record is found

#### Passthrough Mode
- Acts as a simple reverse proxy
- Forwards requests to configured target URLs
- No database interaction

## Getting Started

### Installation

```bash
# Clone the repository
git clone https://github.com/dipjyotimetia/jarvis.git
cd traffix

# Build the application
go build
```

### Usage

```bash
# Start in passthrough mode
./traffix start

# Start in recording mode
./traffix start --record

# Start in replay mode
./traffix start --replay

# Use a custom config file
./traffix start --config my-config.json

# Display version information
./traffix version
```

### Configuration

Edit the config.json file to set:

- HTTP port (default: 8080)
- Target URLs for proxying
- Path-based routing rules
- SQLite database path
- Operating mode (recording/replay)

## Project Structure

- cmd: Command line interface implementation
- `config/`: Configuration handling
- internal: Core implementation
  - `db/`: Database operations
  - `proxy/`: Proxy server implementation 
- main.go: Application entry point