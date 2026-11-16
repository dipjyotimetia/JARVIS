# JARVIS

AI-powered CLI for API testing: generate test data and scenarios from OpenAPI, analyse failures, and inspect traffic.

<img src="docs/assets/jarvis_updated.png" width="400" alt="JARVIS">

[![goreleaser](https://github.com/dipjyotimetia/JARVIS/actions/workflows/release.yml/badge.svg)](https://github.com/dipjyotimetia/JARVIS/actions/workflows/release.yml)

## Installation

Download the binary for your platform from the [releases page](https://github.com/dipjyotimetia/JARVIS/releases).

```bash
chmod +x jarvis
sudo mv jarvis /usr/local/bin/
```

On Windows, put `jarvis.exe` on your `PATH`. To build from source, see the [setup guide](docs/setup.md).

## Quick start

```bash
jarvis setup
jarvis gen generate-testdata --spec api.yaml --count 10
jarvis gen generate-scenarios --path specs/openapi/api.yaml
jarvis analyze analyze-failures --limit 5
jarvis proxy --record
jarvis certificate --cert-dir ./certs
```

## Supported providers

Generation and failure analysis call one LLM through `pkg/engine/llm`. Set `LLM_PROVIDER` to `openai`, `anthropic`, `google`, or `ollama`. If it is unset, Jarvis checks `OPENAI_API_KEY`, then `ANTHROPIC_API_KEY`, then `GOOGLE_API_KEY`, and otherwise uses Ollama.

| Provider | How to authenticate | Default model |
| --- | --- | --- |
| OpenAI | `OPENAI_API_KEY` | `gpt-4o-mini` |
| Anthropic | `ANTHROPIC_API_KEY` | `claude-3-5-sonnet-20241022` |
| Google Gemini | `GOOGLE_API_KEY` | `gemini-1.5-flash` |
| Ollama | local server (`OLLAMA_HOST`, default `http://localhost:11434`) | `llama3.2` |

`LLM_MODEL` overrides the default model. `LLM_API_KEY` is accepted in place of the provider-specific key, and `LLM_BASE_URL` sets a custom endpoint for OpenAI-compatible APIs or for Ollama.

```bash
export OPENAI_API_KEY="sk-..."
export LLM_PROVIDER=openai
export LLM_MODEL=gpt-4o-mini
```

Provider setup, model choice, and troubleshooting: [LLM integration](docs/LLM_INTEGRATION.md) and the [provider quick start](docs/QUICK_START_LLM.md).

## Documentation

- [Commands](docs/commands.md)
- [AI features](docs/ai-features.md): test data, scenarios, and failure analysis
- [Traffic inspector](docs/traffic-inspector.md): proxy, TLS, and OpenAPI validation
- [Workflows](docs/usage.md): examples carried over from the previous README
- [Setup and configuration](docs/setup.md)
- [Further examples](docs/example.md)

## Contributing

```bash
git clone https://github.com/dipjyotimetia/JARVIS.git
cd JARVIS
go mod download
go build -o jarvis
go test ./...
```

## License

MIT. See [LICENSE](LICENSE).

## Support

- [Issues](https://github.com/dipjyotimetia/JARVIS/issues)
- [Discussions](https://github.com/dipjyotimetia/JARVIS/discussions)
