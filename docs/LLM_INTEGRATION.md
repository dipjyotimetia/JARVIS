# LLM Integration with langchaingo

## Overview

Jarvis now uses [langchaingo](https://github.com/tmc/langchaingo) for unified access to multiple LLM providers. This allows you to use OpenAI, Anthropic Claude, Google Gemini, or Ollama (local) models interchangeably.

## Supported Providers

| Provider | Models | API Key Required | Local/Cloud |
|----------|--------|------------------|-------------|
| **OpenAI** | gpt-4o, gpt-4o-mini, gpt-4-turbo, gpt-3.5-turbo | Yes | Cloud |
| **Anthropic** | claude-3-5-sonnet, claude-3-opus, claude-3-sonnet, claude-3-haiku | Yes | Cloud |
| **Google** | gemini-1.5-pro, gemini-1.5-flash, gemini-pro | Yes | Cloud |
| **Ollama** | llama3.2, codellama, mistral, phi, etc. | No | Local |

## Configuration

### Environment Variables

Jarvis automatically detects which provider to use based on available API keys. Priority order:

```bash
# Option 1: OpenAI (recommended for production)
export OPENAI_API_KEY="sk-..."
export LLM_MODEL="gpt-4o-mini"  # Optional, defaults to gpt-4o-mini

# Option 2: Anthropic Claude (recommended for complex reasoning)
export ANTHROPIC_API_KEY="sk-ant-..."
export LLM_MODEL="claude-3-5-sonnet-20241022"  # Optional

# Option 3: Google Gemini (recommended for cost-effectiveness)
export GOOGLE_API_KEY="AIza..."
export LLM_MODEL="gemini-1.5-flash"  # Optional

# Option 4: Ollama (recommended for privacy/offline use)
export OLLAMA_HOST="http://localhost:11434"  # Optional, default
export LLM_MODEL="llama3.2"  # Optional
```

### Explicit Provider Selection

```bash
# Force a specific provider
export LLM_PROVIDER="openai"  # or "anthropic", "google", "ollama"
export LLM_MODEL="gpt-4o"
export LLM_API_KEY="sk-..."  # Only for cloud providers
```

### Advanced Configuration

```bash
# Fine-tune generation parameters
export LLM_TEMPERATURE="0.7"     # Creativity (0.0 - 1.0)
export LLM_MAX_TOKENS="2048"     # Maximum response length
export LLM_TOP_P="0.9"           # Nucleus sampling
export LLM_TOP_K="40"            # Top-K sampling

# Custom endpoints
export LLM_BASE_URL="https://api.openai.com/v1"  # For OpenAI-compatible APIs
```

## Quick Start

### 1. Using Ollama (Local, No API Key)

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull a model
ollama pull llama3.2

# Use Jarvis (automatically detects Ollama)
jarvis gen generate-testdata --spec api.yaml

# Or explicitly set Ollama
export LLM_PROVIDER=ollama
jarvis analyze analyze-failures
```

### 2. Using OpenAI

```bash
# Set API key
export OPENAI_API_KEY="sk-proj-..."

# Use default model (gpt-4o-mini)
jarvis gen generate-testdata --spec api.yaml

# Or specify a different model
export LLM_MODEL="gpt-4o"
jarvis analyze analyze-failures
```

### 3. Using Anthropic Claude

```bash
# Set API key
export ANTHROPIC_API_KEY="sk-ant-..."

# Use default model (claude-3-5-sonnet)
jarvis gen generate-testdata --spec api.yaml

# Or specify a different model
export LLM_MODEL="claude-3-opus-20240229"
jarvis analyze analyze-failures
```

### 4. Using Google Gemini

```bash
# Set API key (get from https://aistudio.google.com/app/apikey)
export GOOGLE_API_KEY="AIza..."

# Use default model (gemini-1.5-flash)
jarvis gen generate-testdata --spec api.yaml

# Or specify pro model
export LLM_MODEL="gemini-1.5-pro"
jarvis analyze analyze-failures
```

## Features Comparison

### Test Data Generation

All providers support test data generation. Recommended models:

- **Best Quality**: `claude-3-5-sonnet-20241022` (Anthropic)
- **Best Speed**: `gemini-1.5-flash` (Google)
- **Best Cost**: `gpt-4o-mini` (OpenAI)
- **Offline/Privacy**: `llama3.2` (Ollama)

```bash
# High-quality test data with Claude
export ANTHROPIC_API_KEY="sk-ant-..."
jarvis gen generate-testdata --spec api.yaml --count 10

# Fast generation with Gemini
export GOOGLE_API_KEY="AIza..."
export LLM_MODEL="gemini-1.5-flash"
jarvis gen generate-testdata --spec api.yaml --count 10
```

### Failure Analysis

All providers support failure analysis. Recommended models:

- **Best Reasoning**: `claude-3-5-sonnet-20241022` (Anthropic)
- **Best Code**: `gpt-4o` (OpenAI)
- **Best Cost**: `gpt-4o-mini` (OpenAI)
- **Offline**: `llama3.2` (Ollama)

```bash
# Deep analysis with Claude
export ANTHROPIC_API_KEY="sk-ant-..."
jarvis analyze analyze-failures --limit 5

# Code-focused analysis with GPT-4
export OPENAI_API_KEY="sk-proj-..."
export LLM_MODEL="gpt-4o"
jarvis analyze analyze-failures
```

## Cost Comparison

| Provider | Model | Cost per 1M Input Tokens | Cost per 1M Output Tokens |
|----------|-------|--------------------------|---------------------------|
| OpenAI | gpt-4o-mini | $0.15 | $0.60 |
| OpenAI | gpt-4o | $2.50 | $10.00 |
| Anthropic | claude-3-haiku | $0.25 | $1.25 |
| Anthropic | claude-3-5-sonnet | $3.00 | $15.00 |
| Google | gemini-1.5-flash | $0.075 | $0.30 |
| Google | gemini-1.5-pro | $1.25 | $5.00 |
| Ollama | Any model | **FREE** | **FREE** |

*Prices as of January 2025. Check provider websites for current pricing.*

## Model Selection Guide

### For Test Data Generation

1. **Development/Testing**: Ollama (`llama3.2`)
   - Free, fast, runs locally
   - Good enough for most scenarios

2. **Production/CI**: Google Gemini (`gemini-1.5-flash`)
   - Very low cost ($0.075/1M tokens)
   - Fast response times
   - Good quality

3. **High Quality Needs**: Anthropic (`claude-3-5-sonnet`)
   - Best understanding of complex schemas
   - Most accurate edge cases
   - Higher cost justified for critical tests

### For Failure Analysis

1. **Quick Debugging**: OpenAI (`gpt-4o-mini`)
   - Good balance of speed and quality
   - Affordable for frequent use

2. **Complex Issues**: Anthropic (`claude-3-5-sonnet`)
   - Best reasoning capabilities
   - Most detailed root cause analysis

3. **Code-Heavy Errors**: OpenAI (`gpt-4o`)
   - Excellent code understanding
   - Great for stack traces

## Provider-Specific Notes

### OpenAI
- Requires API key from https://platform.openai.com/api-keys
- Best general-purpose choice
- Fast and reliable
- Good code understanding

### Anthropic
- Requires API key from https://console.anthropic.com
- Excellent reasoning and analysis
- Longer context windows (200K+ tokens)
- Best for complex scenarios

### Google Gemini
- Requires API key from https://aistudio.google.com/app/apikey
- Most cost-effective cloud option
- Fast inference
- Good multimodal capabilities

### Ollama
- No API key required
- Runs completely locally
- Privacy-preserving
- Requires local resources (GPU recommended)
- Model quality varies

## Troubleshooting

### Provider Not Detected

```bash
# Check environment variables
env | grep -E '(OPENAI|ANTHROPIC|GOOGLE|OLLAMA|LLM)'

# Force provider selection
export LLM_PROVIDER=ollama
```

### API Key Errors

```bash
# Verify API key is valid
echo $OPENAI_API_KEY
echo $ANTHROPIC_API_KEY
echo $GOOGLE_API_KEY

# Test API key
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
  https://api.openai.com/v1/models
```

### Model Not Found

```bash
# List available models for provider
# OpenAI
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"

# Ollama
ollama list
```

### Rate Limits

If you hit rate limits:

```bash
# Use a different provider
export LLM_PROVIDER=google

# Or reduce concurrent requests
jarvis gen generate-testdata --spec api.yaml --count 3  # Instead of 10
```

## Migration from Ollama-only

Previous Jarvis versions only supported Ollama. To migrate:

### No Changes Required

If you were using Ollama, everything continues to work:

```bash
# This still works exactly the same
ollama pull llama3.2
jarvis gen generate-testdata --spec api.yaml
```

### Switch to Cloud Providers

```bash
# Just add an API key
export OPENAI_API_KEY="sk-..."

# Jarvis automatically switches
jarvis gen generate-testdata --spec api.yaml
```

### Force Ollama (if multiple providers configured)

```bash
export LLM_PROVIDER=ollama
export LLM_MODEL=llama3.2
```

## Best Practices

1. **Development**: Use Ollama for fast iteration
2. **CI/CD**: Use Gemini Flash for cost-effectiveness
3. **Production**: Use Claude or GPT-4 for quality
4. **Testing**: Mix providers to ensure robustness

## Examples

### Multi-Provider Workflow

```bash
# Generate test data with Ollama (free, fast)
export LLM_PROVIDER=ollama
jarvis gen generate-testdata --spec api.yaml --count 20

# Analyze failures with Claude (best quality)
export LLM_PROVIDER=anthropic
export ANTHROPIC_API_KEY="sk-ant-..."
jarvis analyze analyze-failures --limit 5

# Generate scenarios with Gemini (cost-effective)
export LLM_PROVIDER=google
export GOOGLE_API_KEY="AIza..."
jarvis gen generate-scenarios --path specs/
```

### Provider Fallback Script

```bash
#!/bin/bash
# Try providers in order of preference

if [ -n "$ANTHROPIC_API_KEY" ]; then
    export LLM_PROVIDER=anthropic
elif [ -n "$OPENAI_API_KEY" ]; then
    export LLM_PROVIDER=openai
elif [ -n "$GOOGLE_API_KEY" ]; then
    export LLM_PROVIDER=google
else
    export LLM_PROVIDER=ollama
fi

jarvis gen generate-testdata --spec api.yaml
```

## Performance Tuning

### Temperature Settings

```bash
# More deterministic (good for test generation)
export LLM_TEMPERATURE=0.1

# More creative (good for scenario generation)
export LLM_TEMPERATURE=0.9

# Balanced (default)
export LLM_TEMPERATURE=0.7
```

### Token Limits

```bash
# Shorter responses (faster, cheaper)
export LLM_MAX_TOKENS=1024

# Longer responses (more detailed)
export LLM_MAX_TOKENS=4096

# Maximum (for complex analysis)
export LLM_MAX_TOKENS=8192
```

## License

Same as Jarvis (MIT)
