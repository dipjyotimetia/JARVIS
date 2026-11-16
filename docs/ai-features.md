# AI features

Jarvis uses the provider selected in [Supported providers](../README.md#supported-providers) for the commands below. See [LLM integration](LLM_INTEGRATION.md) for model and cost notes, and [New features](NEW_FEATURES.md) for a longer walkthrough of test data and failure analysis.

## Test data

`generate-testdata` reads an OpenAPI file and asks the LLM for test cases. Flags:

| Flag | Default | Meaning |
| --- | --- | --- |
| `--spec`, `-s` | required | OpenAPI file |
| `--output`, `-o` | `./testdata` | Directory for the JSON result |
| `--count`, `-c` | `5` | Cases per endpoint |
| `--valid` | `true` | Include cases that should succeed |
| `--invalid` | `true` | Include invalid or boundary cases |
| `--locale` | `en-US` | Locale passed to the generator (`en-GB` is accepted) |

The command always writes a JSON file. `--format` is accepted (`json` or `yaml`) and stored on the request, but the saved file is JSON either way.

```bash
jarvis gen generate-testdata --spec api.yaml

jarvis gen generate-testdata \
  --spec api.yaml \
  --count 10 \
  --locale en-GB \
  --output ./testdata

jarvis gen generate-from-schema --schema user.schema.json --count 10
```

## Failure analysis

`analyze-failures` reads the traffic database (default `traffic_inspector.db`) and asks the LLM to diagnose HTTP 4xx and 5xx exchanges. Each report includes the recorded request and response plus an analysis with root cause, error category, severity, suggestions, possible fixes, reproduction steps, related issues, documentation links, a fix estimate, and prevention tips.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--database`, `-d` | `traffic_inspector.db` | Traffic database |
| `--endpoint`, `-e` | all | Filter by endpoint |
| `--status`, `-s` | `0` (any failure) | Filter by status code |
| `--limit`, `-l` | `10` | Maximum failures to analyze |
| `--time-window`, `-t` | `24h` | How far back to look |
| `--output`, `-o` | stdout only | Also write a JSON report |
| `--stats-only` | `false` | Print statistics and skip the LLM |

```bash
jarvis analyze analyze-failures

jarvis analyze analyze-endpoint /api/users

jarvis analyze analyze-failures --status 500

jarvis analyze analyze-failures --output report.json

jarvis analyze analyze-failures --time-window 1h --output failures.json
```

## Test scenarios and test code

Both commands prompt for the spec type, then stream the LLM response. `generate-test` also prompts for a language and writes a Markdown file under `--output`.

```bash
jarvis gen generate-scenarios --path specs/openapi/api.yaml

jarvis gen generate-test --path specs/ --output tests/
```

## Provider configuration

```bash
export LLM_PROVIDER=openai
export LLM_MODEL=gpt-4o-mini
export LLM_TEMPERATURE=0.7
export LLM_MAX_TOKENS=2048
```

`LLM_PROVIDER` accepts `openai`, `anthropic`, `google`, or `ollama`. Leave it unset to auto-detect from `OPENAI_API_KEY`, then `ANTHROPIC_API_KEY`, then `GOOGLE_API_KEY`, otherwise Ollama. `LLM_TOP_P` and `LLM_TOP_K` are also read when set.
