package llm

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	aiplatform "google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/option"
)

// VertexAIProvider implements the Provider interface for Google Vertex AI
type VertexAIProvider struct {
	projectID string
	location  string
	model     string
	service   *aiplatform.Service
}

// NewVertexAIProvider creates a new Vertex AI provider
func NewVertexAIProvider(projectID, location, model string) (*VertexAIProvider, error) {

	if projectID == "" {
		return nil, fmt.Errorf("project ID not specified and could not be determined from gcloud config")
	}

	if location == "" {
		location = "us-central1"
	}

	if model == "" {
		model = "gemini-pro"
	}

	ctx := context.Background()

	// Use Application Default Credentials
	creds, err := google.FindDefaultCredentials(ctx, aiplatform.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("failed to find default credentials: %w (run 'gcloud auth application-default login')", err)
	}

	service, err := aiplatform.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI service: %w", err)
	}

	return &VertexAIProvider{
		projectID: projectID,
		location:  location,
		model:     model,
		service:   service,
	}, nil
}

// Name returns the provider name
func (p *VertexAIProvider) Name() string {
	return "vertexai"
}

// Analyze sends a prompt to Vertex AI and returns the response
func (p *VertexAIProvider) Analyze(ctx context.Context, prompt string) (string, error) {
	endpoint := fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s",
		p.projectID, p.location, p.model)

	systemInstruction := "You are a Kubernetes troubleshooting expert. Analyze the provided diagnostic data and provide actionable insights."

	request := &aiplatform.GoogleCloudAiplatformV1GenerateContentRequest{
		Contents: []*aiplatform.GoogleCloudAiplatformV1Content{
			{
				Role: "user",
				Parts: []*aiplatform.GoogleCloudAiplatformV1Part{
					{
						Text: fmt.Sprintf("%s\n\n%s", systemInstruction, prompt),
					},
				},
			},
		},
		GenerationConfig: &aiplatform.GoogleCloudAiplatformV1GenerationConfig{
			Temperature:     0.7,
			MaxOutputTokens: 2048,
		},
	}

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	resp, err := p.service.Projects.Locations.Publishers.Models.GenerateContent(endpoint, request).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("Vertex AI API request failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Vertex AI")
	}

	return resp.Candidates[0].Content.Parts[0].Text, nil
}

// Helper function to get Vertex AI provider from environment
func NewVertexAIProviderFromEnv() (*VertexAIProvider, error) {
	projectID := os.Getenv("VERTEX_AI_PROJECT_ID")
	if projectID == "" {
		projectID = os.Getenv("GCP_PROJECT")
	}
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	location := os.Getenv("VERTEX_AI_LOCATION")
	if location == "" {
		location = "us-central1"
	}

	model := os.Getenv("VERTEX_AI_MODEL")
	if model == "" {
		model = "gemini-pro"
	}

	return NewVertexAIProvider(projectID, location, model)
}
