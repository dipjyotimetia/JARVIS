package ollama

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dipjyotimetia/jarvis/pkg/engine/files"
	"github.com/ollama/ollama/api"
)

// GenerateText generates content with a single response
func (c *client) GenerateText(ctx context.Context, prompt string) (*api.GenerateResponse, error) {
	req := &api.GenerateRequest{
		Model:   getDefaultModel(),
		Prompt:  prompt,
		Stream:  &[]bool{false}[0], // Disable streaming for single response
		Options: getDefaultOptions(),
	}

	var response *api.GenerateResponse
	err := c.apiClient.Generate(ctx, req, func(resp api.GenerateResponse) error {
		response = &resp
		return nil
	})

	if err != nil {
		return nil, err
	}

	return response, nil
}

// GenerateWithOptions provides full control over generation parameters
func (c *client) GenerateWithOptions(ctx context.Context, req *api.GenerateRequest, fn api.GenerateResponseFunc) error {
	return c.apiClient.Generate(ctx, req, fn)
}

// GenerateTextStream generates content from specs with streaming response
func (c *client) GenerateTextStream(ctx context.Context, specs []string, specType string) error {
	prompt := buildPrompt(specs, fmt.Sprintf("Generate all possible positive and negative test scenarios in simple english for the provided %s spec file.", specType))
	
	req := &api.GenerateRequest{
		Model:   getDefaultModel(),
		Prompt:  prompt,
		Stream:  &[]bool{true}[0], // Pointer to true
		Options: getDefaultOptions(),
	}

	return c.apiClient.Generate(ctx, req, func(resp api.GenerateResponse) error {
		fmt.Print(resp.Response)
		return nil
	})
}

// GenerateTextStreamWriter generates content and writes to file
func (c *client) GenerateTextStreamWriter(ctx context.Context, specs []string, language, specType string, outputFolder string) error {
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
		Options: getDefaultOptions(),
	}

	return c.apiClient.Generate(ctx, req, func(resp api.GenerateResponse) error {
		_, err := fmt.Fprint(writer, resp.Response)
		return err
	})
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