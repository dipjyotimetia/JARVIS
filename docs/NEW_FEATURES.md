# New AI-Powered Testing Features

## Overview

Added two major AI-powered testing assistance features to Jarvis:

1. **Test Data Generator** - Generate realistic test data from OpenAPI specifications
2. **Failure Analyzer** - AI-powered diagnosis of API failures from recorded traffic

---

## 1. Test Data Generator

### Description
Leverages Ollama AI to generate realistic, schema-compliant test data from OpenAPI specifications. Automatically creates both valid and invalid test cases with proper boundary conditions.

### Features
- ✅ Generates test data from OpenAPI specifications
- ✅ Creates valid test cases that should succeed
- ✅ Creates invalid/boundary test cases for validation testing
- ✅ Supports multiple locales (en-US, en-GB, etc.)
- ✅ Configurable number of test cases per endpoint
- ✅ Schema-aware generation for complex data types
- ✅ JSON schema support for standalone schemas

### Commands

#### Generate from OpenAPI Spec
```bash
# Basic usage
jarvis gen generate-testdata --spec api.yaml

# Generate 10 test cases per endpoint
jarvis gen generate-testdata --spec api.yaml --count 10

# Generate only valid test cases
jarvis gen generate-testdata --spec api.yaml --invalid=false

# Use UK locale for realistic data
jarvis gen generate-testdata --spec api.yaml --locale en-GB

# Custom output directory
jarvis gen generate-testdata --spec api.yaml --output ./my-testdata
```

#### Generate from JSON Schema
```bash
# Generate from standalone schema
jarvis gen generate-from-schema --schema user.schema.json --count 10
```

### Output Format
```json
[
  {
    "endpoint": "/api/users",
    "method": "post",
    "description": "Create a new user",
    "test_cases": [
      {
        "name": "Valid user creation",
        "type": "valid",
        "data": {
          "email": "john.doe@example.com",
          "name": "John Doe",
          "age": 30
        },
        "expected_status": 201,
        "description": "Creates a valid user with all required fields"
      },
      {
        "name": "Invalid email format",
        "type": "invalid",
        "data": {
          "email": "invalid-email",
          "name": "Jane Doe",
          "age": 25
        },
        "expected_status": 400,
        "description": "Tests email validation with invalid format"
      }
    ]
  }
]
```

### Implementation Files
- `pkg/engine/testdata/generator.go` - Core test data generation logic
- `pkg/commands/testdata.go` - CLI commands for test data generation

---

## 2. Failure Analyzer

### Description
AI-powered analysis of API failures from recorded traffic. Provides root cause analysis, actionable suggestions, and prevention tips using Ollama.

### Features
- ✅ Analyzes failed API requests (4xx, 5xx status codes)
- ✅ AI-powered root cause analysis
- ✅ Error categorization (Authentication, Validation, Server Error, etc.)
- ✅ Severity assessment (Low, Medium, High, Critical)
- ✅ Actionable fix suggestions
- ✅ Reproduction steps
- ✅ Prevention tips
- ✅ Related issues and documentation links
- ✅ Fix time estimates
- ✅ Failure statistics and patterns

### Commands

#### Analyze All Failures
```bash
# Analyze recent failures (last 24 hours)
jarvis analyze analyze-failures

# Analyze failures from last hour
jarvis analyze analyze-failures --time-window 1h

# Analyze last 20 failures
jarvis analyze analyze-failures --limit 20

# Filter by status code
jarvis analyze analyze-failures --status 500

# Filter by endpoint
jarvis analyze analyze-failures --endpoint /api/users

# Save detailed report
jarvis analyze analyze-failures --output failures-report.json

# Show only statistics
jarvis analyze analyze-failures --stats-only
```

#### Analyze Specific Endpoint
```bash
# Analyze failures for a specific endpoint
jarvis analyze analyze-endpoint /api/users --limit 5
```

### Output Example
```
📊 Failure Statistics
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Total Failures: 15

┌─────────────┬───────┬──────────────────┬────────────┐
│ Status Code │ Count │ Unique Endpoints │ Percentage │
├─────────────┼───────┼──────────────────┼────────────┤
│ 500         │ 8     │ 3                │ 53.3%      │
│ 404         │ 4     │ 2                │ 26.7%      │
│ 401         │ 3     │ 1                │ 20.0%      │
└─────────────┴───────┴──────────────────┴────────────┘

🔍 Analyzing Failures...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. POST /api/users
   ────────────────────────────────────────
   Status: 500 | Duration: 234ms | Time: 2025-10-01 15:30:45

   🎯 Root Cause:
      Database connection timeout during user creation

   📂 Category: Server Error | Severity: Critical | Estimate: 2 hours

   💡 Suggestions:
      • Check database connection pool settings
      • Verify database server is running
      • Review recent database schema changes

   🔧 Possible Fixes:
      • Increase database connection timeout
      • Add connection retry logic
      • Implement circuit breaker pattern

   🔄 Reproduction Steps:
      1. Send POST request to /api/users with valid data
      2. Database connection pool is exhausted
      3. Request times out after 30 seconds

   🛡️  Prevention:
      • Monitor database connection pool usage
      • Set up alerts for connection pool exhaustion
      • Implement proper connection pooling
```

### Implementation Files
- `pkg/engine/analyzer/failure.go` - Core failure analysis logic
- `pkg/commands/analyze_failures.go` - CLI commands for failure analysis

---

## Architecture

### Test Data Generator Flow
```
OpenAPI Spec → Schema Parser → AI Prompt Builder → Ollama → JSON Parser → Test Cases
```

### Failure Analyzer Flow
```
Traffic DB → Failed Requests → AI Analysis → Structured Diagnosis → Report
```

---

## Requirements

- **Ollama**: Must be running locally (default: `http://localhost:11434`)
- **Model**: Default model `llama3.2` or configure via `OLLAMA_MODEL` environment variable
- **Database**: SQLite database with recorded traffic (for failure analysis)

---

## Configuration

### Environment Variables

```bash
# Ollama Configuration
export OLLAMA_HOST=http://localhost:11434
export OLLAMA_MODEL=llama3.2

# Test Data Generation
export OLLAMA_GENERATION_TEMPERATURE=0.1
export OLLAMA_GENERATION_TOP_K=40
export OLLAMA_GENERATION_TOP_P=0.9
```

---

## Use Cases

### Test Data Generator
1. **API Testing**: Generate test data for automated API tests
2. **Load Testing**: Create realistic test data for performance testing
3. **Development**: Generate sample data for frontend development
4. **QA**: Create edge cases and boundary test scenarios
5. **Documentation**: Generate example data for API documentation

### Failure Analyzer
1. **Debugging**: Quickly identify root causes of API failures
2. **Monitoring**: Analyze patterns in production failures
3. **Post-Incident**: Generate detailed failure reports
4. **Testing**: Validate test scenarios match real failure patterns
5. **Documentation**: Document common failure scenarios

---

## Future Enhancements

### Test Data Generator
- [ ] Support for GraphQL schemas
- [ ] Data relationship management (foreign keys)
- [ ] Custom data generators (plugins)
- [ ] Integration with test frameworks
- [ ] Real-time data generation during tests

### Failure Analyzer
- [ ] Failure pattern clustering
- [ ] Predictive failure analysis
- [ ] Integration with issue trackers (Jira, GitHub)
- [ ] Automated fix suggestions with code
- [ ] Historical trend analysis
- [ ] Alert integration (PagerDuty, Slack)

---

## Examples

### End-to-End Workflow

1. **Generate Test Data**
```bash
# Generate test data from OpenAPI spec
jarvis gen generate-testdata --spec petstore.yaml --count 10
```

2. **Run Tests with Generated Data**
```bash
# Use generated test data in your test framework
# (integrate with your testing tools)
```

3. **Record Traffic**
```bash
# Start proxy to record traffic
jarvis proxy --record
```

4. **Analyze Failures**
```bash
# After tests run, analyze any failures
jarvis analyze analyze-failures --time-window 1h
```

5. **Generate Report**
```bash
# Create detailed failure report
jarvis analyze analyze-failures --output report.json
```

---

## Contributing

When extending these features:

1. **Test Data Generator**: Add new prompts in `buildPromptForOperation()`
2. **Failure Analyzer**: Enhance analysis in `buildAnalysisPrompt()`
3. **Both**: Configure AI parameters via `AIConfig` in `pkg/engine/ollama/client.go`

---

## License

MIT License - Same as Jarvis project
