# Commands

Command tree for the Jarvis CLI, taken from the commands registered in `cmd/root.go`.

```text
jarvis
├── setup
├── version
├── update
├── certificate
├── proxy
├── gen
│   ├── generate-test
│   ├── generate-scenarios
│   ├── generate-testdata
│   └── generate-from-schema
├── analyze
│   ├── spec-analyzer
│   ├── analyze-failures
│   ├── analyze-endpoint
│   └── diff
└── tools
    └── grpc-curl
```

## Root

| Command | What it does |
| --- | --- |
| `setup` | Interactive setup wizard |
| `version` | Print version, commit, build date, and Go version |
| `update` | Check GitHub for a newer release and update the CLI |
| `certificate` | Write a self-signed certificate and key for the HTTPS proxy |
| `proxy` | Start the traffic inspector |

`setup` asks for an API key (optional), an output directory, a language, a framework, optional proxy settings, and a report format (`HTML`, `JSON`, `JUnit XML`, or `Text`). Languages offered are Go, Python, JavaScript, Java, and TypeScript. The wizard can also accept another language.

## gen

| Command | What it does |
| --- | --- |
| `generate-test` | Prompt for a language and a spec type, then ask the LLM to write tests for the spec files under `--path` |
| `generate-scenarios` | Prompt for a spec type, then ask the LLM for positive and negative scenarios for the spec files under `--path` |
| `generate-testdata` | Generate test data from an OpenAPI file (`--spec`) |
| `generate-from-schema` | Generate test data from a JSON Schema file (`--schema`) |

Spec types offered by the prompt are Avro, Protobuf, Swagger, and OpenAPI. `spec-analyzer` only runs its analyzers for Protobuf and OpenAPI.

## analyze

| Command | What it does |
| --- | --- |
| `spec-analyzer` | Analyze Protobuf or OpenAPI files at `--path` |
| `analyze-failures` | Diagnose failed requests (HTTP 4xx and 5xx) stored in the traffic database |
| `analyze-endpoint` | Diagnose failures for one endpoint path |
| `diff` | Compare two OpenAPI files and report breaking changes |

## tools

| Command | What it does |
| --- | --- |
| `grpc-curl` | Print a gRPC curl command from `--proto`, `--service`, and `--method` |

Examples for the AI commands are in [AI features](ai-features.md). Proxy flags are in [Traffic inspector](traffic-inspector.md).
