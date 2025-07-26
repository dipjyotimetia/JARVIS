package ollama

import (
	"context"
)

// Heartbeat checks if the Ollama service is responsive
func (c *client) Heartbeat(ctx context.Context) error {
	return c.apiClient.Heartbeat(ctx)
}

// Version returns the Ollama service version
func (c *client) Version(ctx context.Context) (string, error) {
	return c.apiClient.Version(ctx)
}