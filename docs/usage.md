# Workflows

These examples use commands that the CLI registers. Provider selection is described in [AI features](ai-features.md).

## Development

```bash
export LLM_PROVIDER=ollama
ollama pull llama3.2

jarvis gen generate-testdata --spec api.yaml --count 20

jarvis proxy --record

jarvis analyze analyze-failures
```

## CI

```bash
export LLM_PROVIDER=google
export GOOGLE_API_KEY="AIza..."

jarvis gen generate-testdata --spec api.yaml
jarvis proxy --api-validate --api-spec api.yaml
```

## Failure review

```bash
export LLM_PROVIDER=anthropic
export ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY"

jarvis analyze analyze-failures --limit 10 --output report.json
```

## Recorded traffic, then a report

```bash
jarvis gen generate-testdata --spec petstore.yaml --output testdata/

jarvis proxy --record --api-validate --api-spec petstore.yaml

jarvis analyze analyze-failures --time-window 1h --output failures.json
```

The JSON report's analysis object includes `root_cause` for each failure. A review one-liner:

```bash
cat failures.json | jq '.[] | {endpoint: .url, cause: .analysis.root_cause}'
```

## Switching providers in one shell

```bash
export LLM_PROVIDER=ollama
jarvis gen generate-testdata --spec api.yaml --count 50

export LLM_PROVIDER=anthropic
export ANTHROPIC_API_KEY="sk-ant-..."
jarvis analyze analyze-failures --limit 5

export LLM_PROVIDER=google
export GOOGLE_API_KEY="AIza..."
jarvis gen generate-scenarios --path specs/
```

## Performance figures from the previous README

These numbers were already published in the README. They were not re-measured when the docs were split:

- Binary size: about 48 MB
- Memory: about 100 MB, plus the AI model
- Model response: about 1–5 seconds with local Ollama, about 0.5–3 seconds for the cloud providers
- Proxy throughput: 1000+ requests/sec
