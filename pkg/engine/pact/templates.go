package pact

import (
	"fmt"
	"strings"
)

// PactTemplate defines a template for generating specific types of contracts
type PactTemplate struct {
	Name        string
	Description string
	Language    string
	Framework   string
	Template    string
	TestCode    string
}

// GetDefaultTemplates returns a collection of predefined Pact templates
func GetDefaultTemplates() map[string]*PactTemplate {
	templates := make(map[string]*PactTemplate)

	// JavaScript Jest Template
	templates["javascript-jest"] = &PactTemplate{
		Name:        "JavaScript Jest",
		Description: "Pact contract testing with Jest framework",
		Language:    "javascript",
		Framework:   "jest",
		Template:    getJavaScriptJestTemplate(),
		TestCode:    getJavaScriptJestTestCode(),
	}

	// Python Pytest Template
	templates["python-pytest"] = &PactTemplate{
		Name:        "Python Pytest",
		Description: "Pact contract testing with Pytest framework",
		Language:    "python",
		Framework:   "pytest",
		Template:    getPythonPytestTemplate(),
		TestCode:    getPythonPytestTestCode(),
	}

	// Java JUnit Template
	templates["java-junit"] = &PactTemplate{
		Name:        "Java JUnit",
		Description: "Pact contract testing with JUnit framework",
		Language:    "java",
		Framework:   "junit",
		Template:    getJavaJUnitTemplate(),
		TestCode:    getJavaJUnitTestCode(),
	}

	// Go Testing Template
	templates["go-testing"] = &PactTemplate{
		Name:        "Go Testing",
		Description: "Pact contract testing with Go testing package",
		Language:    "go",
		Framework:   "testing",
		Template:    getGoTestingTemplate(),
		TestCode:    getGoTestingTestCode(),
	}

	return templates
}

// GenerateTestCodeFromTemplate generates test code using a specific template
func GenerateTestCodeFromTemplate(template *PactTemplate, contract *PactContract) (string, error) {
	if template == nil {
		return "", fmt.Errorf("template is nil")
	}

	testCode := template.TestCode

	// Replace placeholders in the template
	replacements := map[string]string{
		"{{CONSUMER_NAME}}": contract.Consumer.Name,
		"{{PROVIDER_NAME}}": contract.Provider.Name,
		"{{INTERACTIONS}}": generateInteractionsCode(contract, template.Language),
		"{{PACKAGE_NAME}}":  strings.ToLower(strings.ReplaceAll(contract.Consumer.Name, "-", "_")),
		"{{CLASS_NAME}}":    toPascalCase(contract.Consumer.Name) + "ContractTest",
	}

	for placeholder, value := range replacements {
		testCode = strings.ReplaceAll(testCode, placeholder, value)
	}

	return testCode, nil
}

// JavaScript Jest Template
func getJavaScriptJestTemplate() string {
	return `{
  "consumer": {
    "name": "{{CONSUMER_NAME}}"
  },
  "provider": {
    "name": "{{PROVIDER_NAME}}"
  },
  "interactions": [
    {{INTERACTIONS}}
  ],
  "metadata": {
    "pactSpecification": {
      "version": "3.0.0"
    },
    "client": {
      "name": "jarvis-pact-generator",
      "version": "1.0.0"
    }
  }
}`
}

func getJavaScriptJestTestCode() string {
	return `const { Pact } = require('@pact-foundation/pact');
const { like, eachLike } = require('@pact-foundation/pact').Matchers;
const axios = require('axios');

describe('{{CONSUMER_NAME}} - {{PROVIDER_NAME}} Contract', () => {
  const provider = new Pact({
    consumer: '{{CONSUMER_NAME}}',
    provider: '{{PROVIDER_NAME}}',
    port: 1234,
    log: './logs/pact.log',
    dir: './pacts',
    logLevel: 'INFO',
    spec: 3
  });

  beforeAll(() => provider.setup());
  afterAll(() => provider.finalize());
  afterEach(() => provider.verify());

  {{INTERACTIONS}}

  describe('Health Check', () => {
    beforeEach(() => {
      return provider.addInteraction({
        state: 'provider is healthy',
        uponReceiving: 'a health check request',
        withRequest: {
          method: 'GET',
          path: '/health'
        },
        willRespondWith: {
          status: 200,
          headers: {
            'Content-Type': 'application/json'
          },
          body: {
            status: 'ok'
          }
        }
      });
    });

    it('should receive health status', async () => {
      const response = await axios.get('http://localhost:1234/health');
      expect(response.status).toBe(200);
      expect(response.data.status).toBe('ok');
    });
  });
});`
}

// Python Pytest Template
func getPythonPytestTemplate() string {
	return `{
  "consumer": {
    "name": "{{CONSUMER_NAME}}"
  },
  "provider": {
    "name": "{{PROVIDER_NAME}}"
  },
  "interactions": [
    {{INTERACTIONS}}
  ],
  "metadata": {
    "pactSpecification": {
      "version": "3.0.0"
    },
    "client": {
      "name": "jarvis-pact-generator",
      "version": "1.0.0"
    }
  }
}`
}

func getPythonPytestTestCode() string {
	return `import pytest
import requests
from pact import Consumer, Provider, Like, EachLike

@pytest.fixture
def pact():
    return Consumer('{{CONSUMER_NAME}}').has_pact_with(
        Provider('{{PROVIDER_NAME}}'),
        host_name='localhost',
        port=1234,
        pact_dir='./pacts',
        version='3.0.0'
    )

class Test{{CLASS_NAME}}:
    def setup_method(self):
        self.base_url = 'http://localhost:1234'

    {{INTERACTIONS}}

    def test_health_check(self, pact):
        expected = {
            'status': 'ok'
        }

        (pact
         .given('provider is healthy')
         .upon_receiving('a health check request')
         .with_request('GET', '/health')
         .will_respond_with(200, body=expected, headers={'Content-Type': 'application/json'}))

        with pact:
            response = requests.get(f'{self.base_url}/health')
            assert response.status_code == 200
            assert response.json() == expected`
}

// Java JUnit Template
func getJavaJUnitTemplate() string {
	return `{
  "consumer": {
    "name": "{{CONSUMER_NAME}}"
  },
  "provider": {
    "name": "{{PROVIDER_NAME}}"
  },
  "interactions": [
    {{INTERACTIONS}}
  ],
  "metadata": {
    "pactSpecification": {
      "version": "3.0.0"
    },
    "client": {
      "name": "jarvis-pact-generator",
      "version": "1.0.0"
    }
  }
}`
}

func getJavaJUnitTestCode() string {
	return `package com.example.{{PACKAGE_NAME}};

import au.com.dius.pact.consumer.dsl.PactDslWithProvider;
import au.com.dius.pact.consumer.junit5.PactConsumerTestExt;
import au.com.dius.pact.consumer.junit5.PactTestFor;
import au.com.dius.pact.core.model.RequestResponsePact;
import au.com.dius.pact.core.model.annotations.Pact;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.springframework.web.client.RestTemplate;

import static org.junit.jupiter.api.Assertions.assertEquals;

@ExtendWith(PactConsumerTestExt.class)
@PactTestFor(providerName = "{{PROVIDER_NAME}}", port = "1234")
public class {{CLASS_NAME}} {

    private final RestTemplate restTemplate = new RestTemplate();
    private final String baseUrl = "http://localhost:1234";

    {{INTERACTIONS}}

    @Pact(consumer = "{{CONSUMER_NAME}}")
    public RequestResponsePact healthCheckPact(PactDslWithProvider builder) {
        return builder
            .given("provider is healthy")
            .uponReceiving("a health check request")
            .path("/health")
            .method("GET")
            .willRespondWith()
            .status(200)
            .body("{\"status\": \"ok\"}")
            .toPact();
    }

    @Test
    @PactTestFor(pactMethod = "healthCheckPact")
    void testHealthCheck() {
        String response = restTemplate.getForObject(baseUrl + "/health", String.class);
        assertEquals("{\"status\": \"ok\"}", response);
    }
}`
}

// Go Testing Template
func getGoTestingTemplate() string {
	return `{
  "consumer": {
    "name": "{{CONSUMER_NAME}}"
  },
  "provider": {
    "name": "{{PROVIDER_NAME}}"
  },
  "interactions": [
    {{INTERACTIONS}}
  ],
  "metadata": {
    "pactSpecification": {
      "version": "3.0.0"
    },
    "client": {
      "name": "jarvis-pact-generator",
      "version": "1.0.0"
    }
  }
}`
}

func getGoTestingTestCode() string {
	return `package {{PACKAGE_NAME}}_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/stretchr/testify/assert"
)

func TestPact{{CLASS_NAME}}(t *testing.T) {
	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Consumer: "{{CONSUMER_NAME}}",
		Provider: "{{PROVIDER_NAME}}",
		Host:     "localhost",
		Port:     1234,
		LogLevel: "INFO",
	}
	defer pact.Teardown()

	{{INTERACTIONS}}

	// Health check test
	t.Run("health check", func(t *testing.T) {
		pact.
			AddInteraction().
			Given("provider is healthy").
			UponReceiving("a health check request").
			WithRequest(dsl.Request{
				Method: "GET",
				Path:   dsl.String("/health"),
			}).
			WillRespondWith(dsl.Response{
				Status: 200,
				Body: map[string]interface{}{
					"status": "ok",
				},
			})

		// Start the mock server
		err := pact.Verify(func() error {
			// Make the actual request
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", pact.Port))
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			// Verify response
			assert.Equal(t, 200, resp.StatusCode)
			
			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			assert.Equal(t, "ok", response["status"])

			return nil
		})

		assert.NoError(t, err)
	})
}`
}

// Helper functions
func generateInteractionsCode(_ *PactContract, language string) string {
	// This would generate language-specific interaction code
	// For now, return a placeholder
	switch language {
	case "javascript":
		return "// Additional interactions will be generated here"
	case "python":
		return "# Additional interactions will be generated here"
	case "java":
		return "// Additional interactions will be generated here"
	case "go":
		return "// Additional interactions will be generated here"
	default:
		return "// Additional interactions will be generated here"
	}
}

func toPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(c rune) bool {
		return c == '-' || c == '_' || c == ' '
	})
	
	result := ""
	for _, word := range words {
		if len(word) > 0 {
			result += strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return result
}