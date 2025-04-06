# *JARVIS*

🚀 A generative AI-driven CLI for testing 🚀

<img src="docs/assets/jarvis.jpg" width="400">

Jarvis is a powerful CLI tool that leverages advanced generative AI technologies (such as Google's Gemini Pro LLM and Gemini Vision Pro) to streamline and enhance various software testing activities. It aims to revolutionize how we approach test case generation and scenario creation
powerful HTTP proxy for recording, replaying, and analyzing API traffic., making it easier for developers and testers to ensure the quality and reliability of their applications.

## Features

- **Multiple Operation Modes**:
  - Record Mode: Capture all requests and responses
  - Replay Mode: Serve recorded responses without contacting backend servers
  - Passthrough Mode: Forward requests to target servers

- **Path-Based Routing**: Route different paths to different target servers
- **SQLite Storage**: Efficiently store and query captured traffic
- **Graceful Shutdown**: Handle process termination properly
- **Low Memory Overhead**: Efficient buffer management with sync.Pool
- **CLI Interface**: Simple command-line operations for all modes
- **Confluence and Jira Integration**: Jarvis can read from Confluence and Jira to suggest test cases using Google Gemini. This feature allows you to integrate your documentation and issue tracking systems with Jarvis to generate relevant test cases.
- **Test Case Generation**: Jarvis can generate test cases based on the provided API specifications, ensuring comprehensive coverage and adherence to best practices.
- **Scenario Generation**: Jarvis can generate test scenarios based on the provided API specifications, ensuring comprehensive coverage and adherence to best practices.

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `http_port` | Port for the HTTP proxy server | 8080 |
| `http_target_url` | Default target URL for proxying | (required) |
| `target_routes` | Array of path-based routing rules | [] |
| `sqlite_db_path` | Path to SQLite database file | traffic_inspector.db |
| `recording_mode` | Enable recording mode | false |
| `replay_mode` | Enable replay mode | false |
