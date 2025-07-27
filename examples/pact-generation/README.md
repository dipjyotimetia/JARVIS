# Pact Contract Generation Examples

This directory contains examples of how to use Jarvis for generating Pact contracts from OpenAPI specifications.

## Prerequisites

1. **Install Ollama**: Download from [https://ollama.ai/](https://ollama.ai/)
2. **Download AI Model**: 
   ```bash
   ollama pull llama3.2
   ```
3. **Start Ollama Service**:
   ```bash
   ollama serve
   ```

## Example 1: Basic Contract Generation

Generate a simple Pact contract from an OpenAPI specification:

```bash
jarvis gen generate-contracts \
  --path="../specs/openapi/v3.0/mini_blog.yaml" \
  --consumer="blog-frontend" \
  --provider="blog-api"
```

This will create:
- `./contracts/blog-frontend-blog-api-pact.json`

## Example 2: Generate with Test Code

Generate a contract with JavaScript/Jest test code:

```bash
jarvis gen generate-contracts \
  --path="../specs/openapi/v3.0/mini_blog.yaml" \
  --consumer="web-app" \
  --provider="blog-service" \
  --language="javascript" \
  --framework="jest" \
  --examples \
  --output="./pact-contracts"
```

This will create:
- `./pact-contracts/web-app-blog-service-pact.json`
- `./pact-contracts/web_app_blog_service_test.js`

## Example 3: Python with Pytest

Generate a contract with Python/Pytest test code:

```bash
jarvis gen generate-contracts \
  --path="../specs/openapi/v3.0/mini_blog.yaml" \
  --consumer="python-client" \
  --provider="blog-api" \
  --language="python" \
  --framework="pytest" \
  --examples
```

## Example 4: Java with JUnit

Generate a contract with Java/JUnit test code:

```bash
jarvis gen generate-contracts \
  --path="../specs/openapi/v3.0/mini_blog.yaml" \
  --consumer="android-app" \
  --provider="backend-service" \
  --language="java" \
  --framework="junit" \
  --examples
```

## Example 5: Go Testing

Generate a contract with Go test code:

```bash
jarvis gen generate-contracts \
  --path="../specs/openapi/v3.0/mini_blog.yaml" \
  --consumer="go-client" \
  --provider="api-gateway" \
  --language="go" \
  --framework="testing" \
  --examples
```

## Understanding the Generated Files

### Pact Contract File
The generated Pact contract follows the [Pact Specification v3.0.0](https://github.com/pact-foundation/pact-specification) and includes:
- Consumer and provider names
- Interactions with requests and responses
- Metadata about the generation process

### Test Code Files
The generated test files include:
- Proper imports and setup
- Pact mock provider configuration
- Test cases for each interaction
- Assertions and verification

## Validation and Quality

Jarvis provides comprehensive validation for generated contracts:

```bash
# The validation happens automatically and includes:
# ✅ Required fields validation
# ✅ HTTP method and status code validation
# ✅ URL path validation
# ✅ Best practices checking
# 💡 Suggestions for improvement
```

## Next Steps

1. **Review Generated Contracts**: Check the generated files for accuracy
2. **Customize Test Code**: Modify the generated test code to fit your needs
3. **Integrate into CI/CD**: Add contract tests to your build pipeline
4. **Provider Verification**: Set up provider tests to verify the contracts

## Troubleshooting

### Common Issues

1. **Ollama Not Running**:
   ```bash
   # Start Ollama service
   ollama serve
   ```

2. **Model Not Available**:
   ```bash
   # Download the required model
   ollama pull llama3.2
   ```

3. **Invalid OpenAPI Spec**:
   - Ensure your OpenAPI specification is valid
   - Use tools like [Swagger Editor](https://editor.swagger.io/) to validate

4. **Permission Issues**:
   ```bash
   # Ensure output directory is writable
   mkdir -p ./contracts
   chmod 755 ./contracts
   ```

## Advanced Configuration

### Environment Variables

```bash
# Customize AI model
export OLLAMA_MODEL=llama3.2
export OLLAMA_HOST=http://localhost:11434

# Run generation
jarvis gen generate-contracts --path="api.yaml" --consumer="app" --provider="api"
```

### Custom Templates

You can extend Jarvis with custom templates for specific languages or frameworks by modifying the template system in the codebase.

## Integration Examples

### JavaScript/Jest Package.json

```json
{
  "devDependencies": {
    "@pact-foundation/pact": "^10.4.1",
    "jest": "^29.0.0"
  },
  "scripts": {
    "test:pact": "jest --testPathPattern=.*pact.*"
  }
}
```

### Python Requirements

```txt
pact-python==1.7.0
pytest==7.4.0
requests==2.31.0
```

### Java Maven Dependencies

```xml
<dependency>
    <groupId>au.com.dius.pact.consumer</groupId>
    <artifactId>junit5</artifactId>
    <version>4.6.2</version>
    <scope>test</scope>
</dependency>
```

### Go Modules

```bash
go mod init example
go get github.com/pact-foundation/pact-go/v2
go get github.com/stretchr/testify
```