package validator

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// APIValidator handles OpenAPI validation of HTTP requests and responses
type APIValidator struct {
	swagger *openapi3.T
	router  routers.Router
	options APIValidatorOptions
}

// APIValidatorOptions configures the behavior of the API validator
type APIValidatorOptions struct {
	// EnableRequestValidation determines if incoming requests should be validated
	EnableRequestValidation bool
	// EnableResponseValidation determines if responses should be validated
	EnableResponseValidation bool
	// StrictMode enables more rigorous validation checks
	StrictMode bool
}

// NewAPIValidator creates a new API validator from OpenAPI spec file
func NewAPIValidator(specPath string, options APIValidatorOptions) (*APIValidator, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("loading OpenAPI spec: %w", err)
	}

	// Validate the OpenAPI document
	err = doc.Validate(loader.Context)
	if err != nil {
		return nil, fmt.Errorf("validating OpenAPI spec: %w", err)
	}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, fmt.Errorf("creating router: %w", err)
	}

	return &APIValidator{
		swagger: doc,
		router:  router,
		options: options,
	}, nil
}

// ValidateRequest validates an HTTP request against the OpenAPI spec
func (v *APIValidator) ValidateRequest(req *http.Request) error {
	if !v.options.EnableRequestValidation {
		return nil
	}

	// Find route
	route, pathParams, err := v.router.FindRoute(req)
	if err != nil {
		var routeError *routers.RouteError
		if errors.As(err, &routeError) {
			if routeError.Reason == routers.ErrPathNotFound.Error() {
				return fmt.Errorf("path not found in API spec: %s", req.URL.Path)
			} else if routeError.Reason == routers.ErrMethodNotAllowed.Error() {
				return fmt.Errorf("method %s not allowed for path %s", req.Method, req.URL.Path)
			}
		}
		return fmt.Errorf("finding route: %w", err)
	}

	// Create validation input
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
		Options: &openapi3filter.Options{
			ExcludeRequestBody:    false,
			MultiError:            true,
			AuthenticationFunc:    nil, // No authentication validation
			IncludeResponseStatus: true,
		},
	}

	// Validate request
	err = openapi3filter.ValidateRequest(context.Background(), requestValidationInput)
	if err != nil {
		return fmt.Errorf("validating request: %w", err)
	}

	return nil
}

// ValidateResponse validates an HTTP response against the OpenAPI spec
func (v *APIValidator) ValidateResponse(req *http.Request, status int, header http.Header, body []byte) error {
	if !v.options.EnableResponseValidation {
		return nil
	}

	// Find route
	route, pathParams, err := v.router.FindRoute(req)
	if err != nil {
		return fmt.Errorf("finding route: %w", err)
	}

	// Create validation input
	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		},
		Status:  status,
		Header:  header,
		Options: &openapi3filter.Options{MultiError: true},
	}

	// Set response body
	if len(body) > 0 {
		responseValidationInput.SetBodyBytes(body)
	}

	// Validate response
	err = openapi3filter.ValidateResponse(context.Background(), responseValidationInput)
	if err != nil {
		return fmt.Errorf("validating response: %w", err)
	}

	return nil
}

// GetPathsWithMethods returns all paths and their HTTP methods defined in the OpenAPI spec
func (v *APIValidator) GetPathsWithMethods() map[string][]string {
	result := make(map[string][]string)

	for path, pathItem := range v.swagger.Paths.Map() {
		var methods []string
		if pathItem.Get != nil {
			methods = append(methods, http.MethodGet)
		}
		if pathItem.Post != nil {
			methods = append(methods, http.MethodPost)
		}
		if pathItem.Put != nil {
			methods = append(methods, http.MethodPut)
		}
		if pathItem.Delete != nil {
			methods = append(methods, http.MethodDelete)
		}
		if pathItem.Patch != nil {
			methods = append(methods, http.MethodPatch)
		}
		if pathItem.Head != nil {
			methods = append(methods, http.MethodHead)
		}
		if pathItem.Options != nil {
			methods = append(methods, http.MethodOptions)
		}
		if len(methods) > 0 {
			result[path] = methods
		}
	}

	return result
}

// NormalizePathForSpec converts a real URL path to an OpenAPI spec path by replacing
// path parameters with their template form
func NormalizePathForSpec(path string, specPaths map[string]struct{}) (string, bool) {
	pathParts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	for specPath := range specPaths {
		specPathParts := strings.Split(strings.TrimPrefix(specPath, "/"), "/")

		if len(pathParts) != len(specPathParts) {
			continue
		}

		match := true
		for i, part := range specPathParts {
			// Check if this part is a path parameter
			if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
				// Path parameter, always matches
				continue
			}

			// Literal path segment must match exactly
			if part != pathParts[i] {
				match = false
				break
			}
		}

		if match {
			return specPath, true
		}
	}

	return "", false
}

// GetOpenAPIInfo returns basic information about the loaded OpenAPI spec
func (v *APIValidator) GetOpenAPIInfo() map[string]string {
	info := make(map[string]string)

	if v.swagger.Info != nil {
		info["title"] = v.swagger.Info.Title
		info["version"] = v.swagger.Info.Version
		if v.swagger.Info.Description != "" {
			info["description"] = v.swagger.Info.Description
		}
	}

	info["paths"] = fmt.Sprintf("%d", v.swagger.Paths.Len())

	return info
}
