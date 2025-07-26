# Ollama Integration

This package provides offline AI capabilities using [Ollama](https://ollama.ai/) with the official Go API library, replacing the previous Gemini cloud-based integration.

## Prerequisites

1. **Install Ollama**: Download and install Ollama from [https://ollama.ai/](https://ollama.ai/)

2. **Download Models**: Pull the models you want to use:
   ```bash
   # For text generation
   ollama pull llama3.2
   
   # For vision capabilities
   ollama pull llava
   
   # List available models
   ollama list
   ```

3. **Start Ollama Service**: 
   ```bash
   ollama serve
   ```
   By default, Ollama runs on `http://localhost:11434`

## Configuration

### Environment Variables

- `OLLAMA_HOST`: Ollama server URL (default: `http://localhost:11434`)
- `OLLAMA_MODEL`: Default model for text generation (default: `llama3.2`)
- `OLLAMA_VISION_MODEL`: Default model for vision tasks (default: `llava`)

### Example Configuration

```bash
export OLLAMA_HOST=http://localhost:11434
export OLLAMA_MODEL=llama3.2
export OLLAMA_VISION_MODEL=llava
```

## Features

### Text Generation
- Stream-based text generation for real-time responses
- File-based content generation with output writing
- Configurable generation options (temperature, top-k, top-p)

### Vision Capabilities
- Image analysis using vision-enabled models like LLaVA
- Base64 image encoding for API compatibility
- Multi-image support

### Model Management
- List available models
- Get detailed model information
- Check model availability

## Usage Examples

### Basic Text Generation
```go
ctx := context.Background()
client, err := ollama.New(ctx)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Stream text generation
specs := []string{"Your specification content here"}
err = client.GenerateTextStream(ctx, specs, "openapi")
```

### Vision Analysis
```go
// Analyze images using official API types
response, err := client.GenerateVision(ctx, "Describe this image", []string{"/path/to/image.jpg"})
if err != nil {
    log.Fatal(err)
}
fmt.Println(response.Response)
```

### Model Management
```go
// List available models
models, err := client.ListModels(ctx)
if err != nil {
    log.Fatal(err)
}
for _, model := range models.Models {
    fmt.Printf("Model: %s, Size: %d bytes\n", model.Name, model.Size)
}

// Check if a specific model is available
available, err := client.IsModelAvailable(ctx, "llama3.2")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Llama 3.2 available: %v\n", available)
```

## Available Models

### Text Models
- `llama3.2` - Latest Llama 3.2 model (recommended)
- `llama3.1` - Llama 3.1 model
- `llama3` - Llama 3 base model
- `codellama` - Code-specialized model
- `mistral` - Mistral model
- `gemma` - Google's Gemma model

### Vision Models
- `llava` - Large Language and Vision Assistant
- `bakllava` - BakLLaVA vision model

## Performance Notes

- **Offline Operation**: All processing happens locally, no internet required after model download
- **Resource Usage**: Models require significant RAM (4GB+ recommended)
- **First Run**: Initial model loading may take time
- **GPU Acceleration**: Ollama supports GPU acceleration if available

## Troubleshooting

### Common Issues

1. **Connection Failed**: Ensure Ollama service is running
   ```bash
   ollama serve
   ```

2. **Model Not Found**: Download the required model
   ```bash
   ollama pull llama3.2
   ```

3. **Out of Memory**: Try smaller models or increase system RAM

4. **Slow Response**: Consider GPU acceleration or smaller models

### Checking Service Status
```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# List running models
ollama ps
```

## Migration from Gemini

The Ollama integration provides the same interface as the previous Gemini implementation but with offline capabilities:

- **No API Keys Required**: Everything runs locally
- **Same Interface**: Drop-in replacement for existing code
- **Better Privacy**: No data sent to external services
- **Cost Effective**: No per-request charges

## Supported Operations

- ✅ Text generation from specifications
- ✅ Streaming responses
- ✅ File-based generation with output writing
- ✅ Vision analysis (with compatible models)
- ✅ Model management and information
- ✅ Configurable generation parameters