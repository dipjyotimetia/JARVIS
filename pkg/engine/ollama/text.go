package ollama

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dipjyotimetia/jarvis/pkg/engine/files"
	"github.com/ollama/ollama/api"
)

// GenerateText generates content with a single response and timeout handling
func (c *client) GenerateText(ctx context.Context, prompt string) (*api.GenerateResponse, error) {
	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	
	req := &api.GenerateRequest{
		Model:   getDefaultModel(),
		Prompt:  prompt,
		Stream:  &[]bool{false}[0], // Disable streaming for single response
		Options: c.getConfigurableOptionsForClient("generation"),
	}

	var response *api.GenerateResponse
	err := c.apiClient.Generate(ctx, req, func(resp api.GenerateResponse) error {
		response = &resp
		return nil
	})

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("AI generation timeout after 5 minutes: %w", err)
		}
		return nil, err
	}

	return response, nil
}

// GenerateWithOptions provides full control over generation parameters
func (c *client) GenerateWithOptions(ctx context.Context, req *api.GenerateRequest, fn api.GenerateResponseFunc) error {
	return c.apiClient.Generate(ctx, req, fn)
}

// GenerateTextStream generates content from specs with streaming response and timeout
func (c *client) GenerateTextStream(ctx context.Context, specs []string, specType string) error {
	// Add timeout to context for streaming operations
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute) // Longer timeout for streaming
	defer cancel()
	
	prompt := buildPrompt(specs, fmt.Sprintf("Generate all possible positive and negative test scenarios in simple english for the provided %s spec file.", specType))
	
	req := &api.GenerateRequest{
		Model:   getDefaultModel(),
		Prompt:  prompt,
		Stream:  &[]bool{true}[0], // Pointer to true
		Options: c.getConfigurableOptionsForClient("generation"),
	}

	err := c.apiClient.Generate(ctx, req, func(resp api.GenerateResponse) error {
		fmt.Print(resp.Response)
		return nil
	})
	
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("streaming generation timeout after 10 minutes: %w", err)
	}
	
	return err
}

// GenerateTextStreamWriter generates content and writes to file with timeout handling
func (c *client) GenerateTextStreamWriter(ctx context.Context, specs []string, language, specType string, outputFolder string) error {
	// Add timeout to context
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute) // Extended timeout for file writing
	defer cancel()
	
	prompt := buildPrompt(specs, fmt.Sprintf("Generate %s tests based on this %s spec.", language, specType))

	ct := time.Now().Format("2006-01-02-15-04-05")
	files.CheckDirectryExists(outputFolder)
	outputFile, err := os.Create(fmt.Sprintf("%s/%s_output_test.md", outputFolder, ct))
	if err != nil {
		return err
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	defer writer.Flush()

	req := &api.GenerateRequest{
		Model:   getDefaultModel(),
		Prompt:  prompt,
		Stream:  &[]bool{true}[0], // Pointer to true
		Options: c.getConfigurableOptionsForClient("generation"),
	}

	err = c.apiClient.Generate(ctx, req, func(resp api.GenerateResponse) error {
		_, err := fmt.Fprint(writer, resp.Response)
		return err
	})
	
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("file writing generation timeout after 15 minutes: %w", err)
	}
	
	return err
}

// buildPrompt combines specs with instruction text
func buildPrompt(specs []string, instruction string) string {
	var builder strings.Builder
	
	for _, spec := range specs {
		builder.WriteString(spec)
		builder.WriteString("\n")
	}
	
	builder.WriteString("\n")
	builder.WriteString(instruction)
	
	return builder.String()
}