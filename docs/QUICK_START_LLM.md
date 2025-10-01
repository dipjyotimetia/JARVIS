# Quick Start: Multi-LLM Support

## Choose Your Provider

### Option 1: Ollama (FREE, Local, Private) ⭐ Recommended for Development

```bash
# 1. Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# 2. Pull a model
ollama pull llama3.2

# 3. Use Jarvis (works automatically)
jarvis gen generate-testdata --spec api.yaml
jarvis analyze analyze-failures
```

**Pros**: Free, fast, runs offline, privacy-preserving
**Cons**: Requires local resources (GPU recommended)

---

### Option 2: OpenAI ($$, Cloud, Easy) ⭐ Recommended for Production

```bash
# 1. Get API key from https://platform.openai.com/api-keys

# 2. Set environment variable
export OPENAI_API_KEY="sk-proj-..."

# 3. Use Jarvis (automatically detects OpenAI)
jarvis gen generate-testdata --spec api.yaml
jarvis analyze analyze-failures
```

**Pros**: High quality, fast, reliable, good code understanding
**Cons**: Costs money, requires internet

**Pricing**: ~$0.15 per 1M input tokens (gpt-4o-mini)

---

### Option 3: Google Gemini ($, Cloud, Cheap) ⭐ Recommended for CI/CD

```bash
# 1. Get API key from https://aistudio.google.com/app/apikey

# 2. Set environment variable
export GOOGLE_API_KEY="AIza..."

# 3. Use Jarvis (automatically detects Google)
jarvis gen generate-testdata --spec api.yaml
jarvis analyze analyze-failures
```

**Pros**: Very cheap, fast, good quality
**Cons**: Requires internet

**Pricing**: ~$0.075 per 1M input tokens (gemini-1.5-flash)

---

### Option 4: Anthropic Claude ($$$, Cloud, Best Quality) ⭐ Recommended for Complex Analysis

```bash
# 1. Get API key from https://console.anthropic.com

# 2. Set environment variable
export ANTHROPIC_API_KEY="sk-ant-..."

# 3. Use Jarvis (automatically detects Anthropic)
jarvis gen generate-testdata --spec api.yaml
jarvis analyze analyze-failures
```

**Pros**: Best reasoning, excellent analysis, large context window
**Cons**: More expensive, requires internet

**Pricing**: ~$3.00 per 1M input tokens (claude-3-5-sonnet)

---

## Common Tasks

### Generate Test Data

```bash
# Basic usage (uses auto-detected provider)
jarvis gen generate-testdata --spec api.yaml

# Generate 10 test cases per endpoint
jarvis gen generate-testdata --spec api.yaml --count 10

# Generate only valid test cases
jarvis gen generate-testdata --spec api.yaml --invalid=false

# Use specific provider
export LLM_PROVIDER=openai
jarvis gen generate-testdata --spec api.yaml
```

### Analyze API Failures

```bash
# Analyze recent failures
jarvis analyze analyze-failures

# Analyze specific status codes
jarvis analyze analyze-failures --status 500

# Analyze specific endpoint
jarvis analyze analyze-endpoint /api/users

# Save detailed report
jarvis analyze analyze-failures --output report.json
```

### Generate Test Scenarios

```bash
# Generate test scenarios from spec
jarvis gen generate-scenarios --path specs/openapi/

# Use specific provider
export LLM_PROVIDER=anthropic
jarvis gen generate-scenarios --path specs/
```

## Advanced Configuration

### Change Model

```bash
# OpenAI
export LLM_MODEL="gpt-4o"

# Anthropic
export LLM_MODEL="claude-3-5-sonnet-20241022"

# Google
export LLM_MODEL="gemini-1.5-pro"

# Ollama
export LLM_MODEL="codellama"
```

### Adjust Creativity

```bash
# More deterministic (good for test generation)
export LLM_TEMPERATURE=0.1

# More creative (good for scenario generation)
export LLM_TEMPERATURE=0.9

# Balanced (default)
export LLM_TEMPERATURE=0.7
```

### Control Response Length

```bash
# Shorter (faster, cheaper)
export LLM_MAX_TOKENS=1024

# Longer (more detailed)
export LLM_MAX_TOKENS=4096
```

## Switching Providers

### On-the-Fly

```bash
# Use Ollama for data generation
export LLM_PROVIDER=ollama
jarvis gen generate-testdata --spec api.yaml

# Use Claude for failure analysis
export LLM_PROVIDER=anthropic
export ANTHROPIC_API_KEY="sk-ant-..."
jarvis analyze analyze-failures
```

### In Scripts

```bash
#!/bin/bash

# Generate test data (free with Ollama)
export LLM_PROVIDER=ollama
jarvis gen generate-testdata --spec api.yaml --count 20

# Analyze failures (best quality with Claude)
export LLM_PROVIDER=anthropic
export ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY"
jarvis analyze analyze-failures --limit 5

# Generate scenarios (cost-effective with Gemini)
export LLM_PROVIDER=google
export GOOGLE_API_KEY="$GOOGLE_API_KEY"
jarvis gen generate-scenarios --path specs/
```

## Cost Optimization

### Development Workflow

```bash
# Free for development
export LLM_PROVIDER=ollama
ollama pull llama3.2
```

### CI/CD Workflow

```bash
# Cheap for CI/CD
export LLM_PROVIDER=google
export GOOGLE_API_KEY="$GOOGLE_API_KEY"
export LLM_MODEL="gemini-1.5-flash"
```

### Production Workflow

```bash
# High quality for production
export LLM_PROVIDER=anthropic
export ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY"
export LLM_MODEL="claude-3-5-sonnet-20241022"
```

## Troubleshooting

### Check Current Provider

```bash
# See environment variables
env | grep -E '(LLM|OPENAI|ANTHROPIC|GOOGLE|OLLAMA)'
```

### Test Provider

```bash
# Try generation with debug
jarvis --debug gen generate-testdata --spec api.yaml
```

### Common Issues

**"No API key found"**
```bash
# Set appropriate key
export OPENAI_API_KEY="sk-..."
# or
export ANTHROPIC_API_KEY="sk-ant-..."
# or
export GOOGLE_API_KEY="AIza..."
```

**"Model not found"**
```bash
# Check available models
ollama list  # for Ollama

# Or set specific model
export LLM_MODEL="gpt-4o-mini"
```

**"Rate limit exceeded"**
```bash
# Switch to different provider
export LLM_PROVIDER=google
```

## Best Practices

1. **Development**: Use Ollama (free, fast)
2. **CI/CD**: Use Gemini (cheap, good quality)
3. **Production**: Use Claude or GPT-4 (best quality)
4. **Testing**: Mix providers to ensure robustness

## Examples

### Complete Workflow

```bash
# 1. Generate test data locally
export LLM_PROVIDER=ollama
jarvis gen generate-testdata --spec petstore.yaml --count 10 --output ./testdata

# 2. Run your tests (not shown)
# ...

# 3. Analyze failures with cloud AI
export LLM_PROVIDER=anthropic
export ANTHROPIC_API_KEY="sk-ant-..."
jarvis analyze analyze-failures --time-window 1h --output failures.json

# 4. Generate report
cat failures.json | jq .
```

### Multi-Provider Script

```bash
#!/bin/bash
set -e

# Function to check if provider is available
check_provider() {
    case $1 in
        ollama)
            ollama list &>/dev/null && return 0 || return 1
            ;;
        openai)
            [ -n "$OPENAI_API_KEY" ] && return 0 || return 1
            ;;
        anthropic)
            [ -n "$ANTHROPIC_API_KEY" ] && return 0 || return 1
            ;;
        google)
            [ -n "$GOOGLE_API_KEY" ] && return 0 || return 1
            ;;
    esac
}

# Try providers in order of preference
if check_provider ollama; then
    export LLM_PROVIDER=ollama
    echo "Using Ollama (local)"
elif check_provider google; then
    export LLM_PROVIDER=google
    echo "Using Google Gemini (cloud)"
elif check_provider openai; then
    export LLM_PROVIDER=openai
    echo "Using OpenAI (cloud)"
elif check_provider anthropic; then
    export LLM_PROVIDER=anthropic
    echo "Using Anthropic Claude (cloud)"
else
    echo "No LLM provider available!"
    exit 1
fi

# Run Jarvis
jarvis gen generate-testdata --spec api.yaml
```

## Next Steps

1. Choose a provider from above
2. Set environment variables
3. Run Jarvis commands
4. See full documentation in `docs/LLM_INTEGRATION.md`

## Support

- Full Guide: `docs/LLM_INTEGRATION.md`
- Migration: `docs/MIGRATION_SUMMARY.md`
- Features: `docs/NEW_FEATURES.md`
- Issues: https://github.com/dipjyotimetia/jarvis/issues
