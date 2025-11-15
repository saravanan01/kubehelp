package llm

import (
	"context"
)

// Provider defines the interface for LLM providers
type Provider interface {
	// Analyze sends diagnostic data to the LLM and returns insights
	Analyze(ctx context.Context, prompt string) (string, error)
	// Name returns the provider name
	Name() string
}

// Config holds LLM provider configuration
type Config struct {
	Provider string
	APIKey   string
	Model    string
	BaseURL  string
}
