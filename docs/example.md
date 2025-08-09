# Jarvis Usage Examples

This document provides comprehensive examples of using Jarvis for API testing and traffic inspection.

## Getting Started

Before running these examples, ensure you have:
1. Jarvis installed and configured (see [setup.md](setup.md))
2. A local model provider: either Ollama (native) or Docker Model Runner (OpenAI-compatible). See Setup for both options.
3. Sample specification files in your `specs/` directory

## Generation Commands

### Generate Test Scenarios

Generate AI-powered test scenarios from OpenAPI specifications:

```bash
# Basic scenario generation
jarvis gen generate-scenarios --path="specs/openapi/v3.0/mini_blog.yaml"

# Generate from multiple specs
jarvis gen generate-scenarios --path="specs/openapi/"

# Interactive mode (prompts for spec type)
jarvis gen generate-scenarios --path="specs/"
```

**Example Output:**

**Test Scenario 1: Retrieve All Blog Posts**

* Precondition:
 The API server is up and running.
* Action: Send a GET request to the "/posts" endpoint.
* Expected Result: The server responds with
 a status code of 200 and a list of all blog posts in JSON format.

**Test Scenario 2: Create a New Blog Post**

* Precondition: The API server is up and running.
* Action: Send a POST request to the "/posts" endpoint with a valid JSON payload
 representing a new blog post.
* Expected Result: The server responds with a status code of 201 and the newly created blog post in JSON format.

**Test Scenario 3: Retrieve a Specific Blog Post**

* Precondition: The API server is up and running.
* Action: Send a GET request to the "/posts/{postId}" endpoint with a valid postId.
* Expected Result: The server responds with a status code of 200 and the details of the requested blog post in JSON format.

**Test Scenario 4: Update a Blog Post**

* Precondition: The API
 server is up and running.
* Action: Send a PATCH request to the "/posts/{postId}" endpoint with a valid postId and a JSON payload containing the updated fields.
* Expected Result: The server responds with a status code of 200 and the updated blog post in JSON format.

**Test Scenario 5: Delete a Blog Post**

* Precondition: The API server is up and running.
* Action: Send a DELETE request to the "/posts/{postId}" endpoint with a valid postId.
* Expected Result: The server responds with a status code of 204 and no content in the response body.

**Test Scenario 6: Retrieve Comments for a Blog Post**

* Precondition: The API server is up and running.
* Action: Send a GET request to the "/posts/{postId}/comments" endpoint with a valid postId.
* Expected Result: The server responds with a status code of 200 and a list of comments for the specified blog post in JSON format.

**Test Scenario 7: Add a New Comment**

* Precondition: The API server is up and running.
* Action: Send a POST request to the "/posts/{postId}/comments" endpoint
 with a valid postId and a JSON payload representing a new comment.
* Expected Result: The server responds with a status code of 201 and the newly created comment in JSON format.

**Test Scenario 8: Retrieve User Profile**

* Precondition: The API server is up and running.
* Action: Send a GET request to the "/users/{userId}" endpoint with a valid userId.
* Expected Result: The server responds with a status code of 200 and the user profile details in JSON format.

**Test Scenario 9: Handle Invalid Input**

* Precondition: The API server is up and running.
* Action: Send a request to an endpoint with invalid input (e.g., an invalid postId or userId).
* Expected Result: The server responds with a status code of 400 (Bad Request) and an error message in the response body.

**Test Scenario 10: Handle Non-existent Resources**

* Precondition: The API server is up and running.
* Action: Send a request to an endpoint with a non-existent resource (e.g., a postId or userId that does not exist).
* Expected Result: The server responds with a status code of 404 (Not Found) and an error message in the response body.
```


### Generate Test Code

Generate complete test code from Protobuf specifications:

```bash
# Generate Go test code from Protobuf specs
jarvis gen generate-test --path="specs/proto" --output="output"

# Generate with interactive language selection
jarvis gen generate-test --path="specs/proto/notify.proto" --output="./tests"

# Batch generation from multiple proto files
jarvis gen generate-test --path="specs/proto/" --output="./generated-tests"
```

**Example Generated Test Code:**
```go
import (
	"context"
	"fmt"

	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go
-resty/resty/v2"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"

	notifypb "github.com/username
/project/notify/notifypb"
)

func TestCreateNotification(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/notifications", r.URL.Path)

		var req notifypb.CreateNotificationRequest
		assert.NoError(t, jsonpb.Unmarshal(r.Body, &req))

		resp := &notifypb
.CreateNotificationResponse{
			Notification: &notifypb.Notification{
				Id:           1,
				Title:        "Test Notification",
				Message:      "This is a test notification.",
				RecipientId:  "user-1",
				Timestamp:    "2023-03-08 15:06:30",
				Status:       "sent",
			},
		}
		jsonpb.Marshal(w, resp)
	}))
	defer srv.Close()

	client := resty.New()
	client.SetBaseURL(srv.URL)

	now := time.Now()
	timestamp, err := ptypes.TimestampProto(now)
	assert.NoError(t, err)

	req := &notifypb.CreateNotificationRequest{
		Notification: &notifypb.Notification{
			Title:        "Test Notification",
			Message:      "This is a test notification.",
			RecipientId:  "user-1",
			Timestamp:    timestamp.String(),
			Status:       "sent",
		},
	}
	resp, err := client.
R().
		SetBody(req).
		Post("/v1/notifications")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var respProto notifypb.CreateNotificationResponse
	assert.NoError(t, jsonpb.Unmarshal(resp.Body(), &respProto))
	assert.Equal(t, int32(1), respProto.Notification.Id)
	assert.Equal(t, "Test Notification", respProto.Notification.Title)
	assert.Equal(t, "This is a test notification.", respProto.Notification.Message)
	assert.Equal(t, "user-1", respProto.Notification.RecipientId)
	assert.Equal(t, timestamp.String(), respProto.Notification.Timestamp)
	assert.Equal(t, "sent", respProto.Notification.Status)
}

func TestGetNotifications(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/notifications/user-1", r.URL.Path)

		resp := &notifypb.GetNotificationsResponse{
			Notifications: &notifypb.NotificationList{
				Notifications: []*notifypb.Notification{
					{
						Id:           1,
						Title:        "Test Notification 1",
						Message:      "This is a test notification 1.",
						RecipientId:  "user-1",
						Timestamp:    "2023-03-08 15:06:30",
						Status:       "sent",
					},
					{
						Id:           2,
						Title:        "Test Notification 2",
						Message:      "This is a test notification 2.",
						RecipientId:  "user-1",
						Timestamp:    "2023-03-08 15:07:00",
						Status:       "delivered",
					},
				},
			},
		}
		jsonpb.Marshal(w, resp)
	}))
	defer srv.
Close()

	client := resty.New()
	client.SetBaseURL(srv.URL)

	resp, err := client.R().
		Get("/v1/notifications/user-1")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode())

	var respProto notifypb.GetNotificationsResponse
	assert.NoError(t, jsonpb.Unmarshal(resp.Body(), &respProto))
	assert.Equal(t, 2, len(respProto.Notifications.Notifications))
	for _, n := range respProto.Notifications.Notifications {
		assert.Contains(t, []string{"sent", "delivered"}, n.Status)
	}
}
```

### Generate Pact Contracts

Create consumer-driven contract tests from OpenAPI specifications:

```bash
# Basic contract generation
jarvis gen generate-contracts \
  --path="specs/openapi/v3.0/api.yaml" \
  --consumer="web-frontend" \
  --provider="user-service"

# Generate with test code examples
jarvis gen generate-contracts \
  --path="specs/openapi/petstore.yaml" \
  --consumer="mobile-app" \
  --provider="petstore-api" \
  --language="javascript" \
  --framework="jest" \
  --examples

# Custom output directory
jarvis gen generate-contracts \
  --path="specs/openapi/" \
  --consumer="client" \
  --provider="server" \
  --output="./pact-contracts"

# Generate with Python/pytest
jarvis gen generate-contracts \
  --path="specs/openapi/blog.yaml" \
  --consumer="blog-ui" \
  --provider="blog-service" \
  --language="python" \
  --framework="pytest" \
  --examples
```

**Example Generated Files:**
- `web-frontend-user-service-pact.json`: Pact contract file
- `web_frontend_user_service_test.js`: Jest test file (with --examples)

## Analysis Commands

### Spec Analysis

Analyze API specifications for completeness and potential issues:

```bash
# Analyze Protobuf specifications
jarvis analyze spec-analyzer --path="specs/proto/"

# Analyze OpenAPI specifications
jarvis analyze spec-analyzer --path="specs/openapi/v3.0/"

# Analyze specific file
jarvis analyze spec-analyzer --path="specs/proto/user.proto"
```

**Example Output:**
```
Analyzing protobuf spec files...
✅ Found 3 service definitions
✅ Found 15 message types
⚠️  Missing documentation for UserService.DeleteUser method
✅ All required fields are properly defined
📊 Analysis Summary:
   - Services: 3
   - Methods: 12
   - Messages: 15
   - Issues: 1 warning
```

## Utility Tools

### gRPC Curl Generator

Generate curl commands for testing gRPC services:

```bash
# Generate gRPC curl command
jarvis tools grpc-curl \
  --proto="specs/proto/user.proto" \
  --service="UserService" \
  --method="GetUser"

# Generate for complex method
jarvis tools grpc-curl \
  --proto="specs/proto/notification.proto" \
  --service="NotificationService" \
  --method="CreateNotification"
```

**Example Output:**
```bash
# Generated gRPC curl command:
grpcurl -plaintext \
  -d '{"user_id": "12345"}' \
  localhost:9000 \
  user.UserService/GetUser
```

## Traffic Inspector Examples

### Basic Proxy Usage

Start the traffic inspector to monitor API calls:

```bash
# Start basic HTTP proxy
jarvis proxy

# Start with custom ports
jarvis proxy --ui-port=9999

# Recording mode - capture all traffic
jarvis proxy --record

# Replay previously recorded traffic
jarvis proxy --replay
```

### HTTPS and TLS Examples

```bash
# Generate certificates first
jarvis certificate --cert-dir ./certs

# Start HTTPS proxy
jarvis proxy --tls --cert ./certs/server.crt --key ./certs/server.key

# Start with mutual TLS
jarvis proxy --mtls \
  --client-ca ./certs/ca.crt \
  --client-cert ./certs/client.crt \
  --client-key ./certs/client.key
```

### OpenAPI Validation

```bash
# Enable API validation
jarvis proxy --api-validate --api-spec ./specs/openapi/api.yaml

# Strict validation mode
jarvis proxy \
  --api-validate \
  --api-spec ./specs/openapi/strict-api.yaml \
  --strict-validation

# Validate only requests
jarvis proxy \
  --api-validate \
  --api-spec ./specs/openapi/api.yaml \
  --validate-req \
  --validate-resp=false
```

### Testing with the Proxy

Once the proxy is running, test it with your API calls:

```bash
# Test through the proxy (proxy running on localhost:8080)
curl -x http://localhost:8080 https://jsonplaceholder.typicode.com/posts/1

# Direct API call through proxy
curl http://localhost:8080/posts/1

# POST request through proxy
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -d '{"title": "Test Post", "body": "This is a test", "userId": 1}'

# View captured traffic in web UI
open http://localhost:9090/ui/
```

## Configuration Examples

### Using Config File

Create a `config.yaml` file for persistent settings:

```yaml
# config.yaml
http_port: 8080
ui_port: 9090
recording_mode: true
sqlite_db_path: "./data/traffic.db"

target_routes:
  - path: "/api/users/*"
    target: "https://jsonplaceholder.typicode.com"
  - path: "/api/posts/*"
    target: "https://jsonplaceholder.typicode.com"

http_target_url: "https://httpbin.org"

language:
  preferred: "javascript"
  framework: "jest"
```

Then run:
```bash
jarvis proxy --config config.yaml
```

### Environment Variables

```bash
# Set environment variables
export JARVIS_OUTPUT_DIR="./custom-output"
export JARVIS_CONFIG="./custom-config.yaml"

# Commands will use these settings
jarvis gen generate-test --path="specs/proto"
```

## Interactive Setup

Use the setup wizard for guided configuration:

```bash
jarvis setup
```

This will prompt you through:
1. AI API configuration
2. Output directory selection
3. Language and framework preferences
4. Proxy settings
5. Report format selection

## Real-world Workflow Example

Here's a complete workflow for API testing:

```bash
# 1. Initial setup
jarvis setup

# 2. Generate certificates for HTTPS testing
jarvis certificate --cert-dir ./certs

# 3. Analyze existing API specifications
jarvis analyze spec-analyzer --path="specs/openapi/"

# 4. Generate test scenarios
jarvis gen generate-scenarios --path="specs/openapi/user-api.yaml"

# 5. Generate Pact contracts
jarvis gen generate-contracts \
  --path="specs/openapi/user-api.yaml" \
  --consumer="frontend" \
  --provider="user-service" \
  --language="javascript" \
  --framework="jest" \
  --examples

# 6. Start traffic inspector with validation
jarvis proxy \
  --record \
  --api-validate \
  --api-spec ./specs/openapi/user-api.yaml \
  --tls --cert ./certs/server.crt --key ./certs/server.key

# 7. Run your tests against the proxy
npm test  # or your test command

# 8. View results in web UI
open http://localhost:9090/ui/

# 9. Stop recording and replay traffic
# Stop with Ctrl+C, then:
jarvis proxy --replay
```

## Advanced Examples

### Custom Target Routing

Configure path-based routing in `config.yaml`:

```yaml
target_routes:
  - path: "/api/v1/users/*"
    target: "https://users.service.com"
  - path: "/api/v1/orders/*"
    target: "https://orders.service.com"
  - path: "/api/v1/payments/*"
    target: "https://payments.service.com"
  - path: "/health/*"
    target: "http://localhost:8081"
http_target_url: "https://api.fallback.com"  # Default for unmatched paths
```

### Batch Contract Generation

Generate contracts for multiple services:

```bash
#!/bin/bash
# batch-contracts.sh

services=("user-service" "order-service" "payment-service")
specs=("user-api.yaml" "order-api.yaml" "payment-api.yaml")

for i in "${!services[@]}"; do
  jarvis gen generate-contracts \
    --path="specs/openapi/${specs[$i]}" \
    --consumer="web-frontend" \
    --provider="${services[$i]}" \
    --language="typescript" \
    --framework="jest" \
    --examples \
    --output="./contracts/${services[$i]}"
done
```

This comprehensive guide demonstrates the full capability of Jarvis for API testing and traffic inspection workflows.