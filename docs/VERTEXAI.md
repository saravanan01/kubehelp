# Google Vertex AI Setup Guide

## Overview

Vertex AI is Google Cloud's enterprise AI platform. Unlike the consumer-facing Gemini API, Vertex AI:
- Uses Google Cloud authentication (Application Default Credentials)
- Integrates with GCP IAM and billing
- Offers enterprise features like VPC-SC, CMEK, and audit logging
- No API key needed - uses OAuth 2.0

## Prerequisites

1. **Google Cloud Project** with Vertex AI API enabled
2. **gcloud CLI** installed and configured
3. **Application Default Credentials** set up

## Setup Steps

### 1. Install gcloud CLI

```bash
# macOS
brew install google-cloud-sdk

# Or download from https://cloud.google.com/sdk/docs/install
```

### 2. Authenticate

```bash
# Login to your Google account
gcloud auth login

# Set up Application Default Credentials
gcloud auth application-default login

# Set your project
gcloud config set project YOUR_PROJECT_ID
```

### 3. Enable Vertex AI API

```bash
gcloud services enable aiplatform.googleapis.com
```

### 4. Set Environment Variables

```bash
# Optional: Explicitly set project ID (auto-detected from gcloud if not set)
export VERTEX_AI_PROJECT_ID="your-project-id"

# Optional: Set location (default: us-central1)
export VERTEX_AI_LOCATION="us-central1"

# Optional: Set model (default: gemini-pro)
export VERTEX_AI_MODEL="gemini-pro"
```

## Usage

```bash
# Build the tool
go mod tidy
go build -o kubehelp ./cmd/...

# Use Vertex AI for diagnosis
./kubehelp diagnose -n production --llm vertexai

# With verbose output
./kubehelp diagnose -n staging --llm vertexai --verbose

# Specify workloads
./kubehelp diagnose -n prod -w api-server,worker --llm vertexai
```

## Available Models

Vertex AI supports various Gemini models:

- **gemini-pro** (default) - Best for text generation
- **gemini-1.5-pro** - Latest with 1M token context
- **gemini-1.5-flash** - Faster, lower cost
- **gemini-ultra** - Most capable (preview)

```bash
# Use a specific model
export VERTEX_AI_MODEL="gemini-1.5-pro"
./kubehelp diagnose -n prod --llm vertexai
```

## Pricing

Vertex AI pricing is based on:
- **Input tokens**: Characters processed
- **Output tokens**: Characters generated
- **Region**: Pricing varies by location

Example pricing (us-central1):
- Gemini Pro: $0.00025/1K input tokens, $0.0005/1K output tokens
- Gemini Flash: ~50% cheaper

See latest pricing: https://cloud.google.com/vertex-ai/pricing

## Authentication Methods

### Method 1: gcloud CLI (Recommended for local development)

```bash
gcloud auth application-default login
```

### Method 2: Service Account (For production/CI/CD)

```bash
# Create service account
gcloud iam service-accounts create kubehelp \
    --display-name="kubehelp service account"

# Grant Vertex AI User role
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
    --member="serviceAccount:kubehelp@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/aiplatform.user"

# Create and download key
gcloud iam service-accounts keys create ~/kubehelp-key.json \
    --iam-account=kubehelp@YOUR_PROJECT_ID.iam.gserviceaccount.com

# Set credentials
export GOOGLE_APPLICATION_CREDENTIALS=~/kubehelp-key.json
```

### Method 3: Workload Identity (For Kubernetes)

When running in GKE, use Workload Identity:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubehelp
  annotations:
    iam.gke.io/gcp-service-account: kubehelp@PROJECT_ID.iam.gserviceaccount.com
```

## Server Deployment

### Docker

```bash
# Build image
docker build -t kubehelp-server .

# Run with Vertex AI
docker run -p 8080:8080 \
  -v ~/.config/gcloud:/root/.config/gcloud:ro \
  -e VERTEX_AI_PROJECT_ID=your-project \
  kubehelp-server
```

### Kubernetes with Workload Identity

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubehelp-server
spec:
  template:
    spec:
      serviceAccountName: kubehelp
      containers:
      - name: server
        image: kubehelp-server:latest
        env:
        - name: VERTEX_AI_PROJECT_ID
          value: "your-project-id"
        - name: VERTEX_AI_LOCATION
          value: "us-central1"
```

## Comparison: Gemini API vs Vertex AI

| Feature                 | Gemini API               | Vertex AI               |
| ----------------------- | ------------------------ | ----------------------- |
| **Auth**                | API Key                  | OAuth 2.0 / ADC         |
| **Setup**               | Instant                  | Requires GCP project    |
| **Cost**                | Free tier                | Pay-as-you-go           |
| **Enterprise Features** | No                       | Yes (VPC-SC, CMEK, etc) |
| **Rate Limits**         | 60 RPM free tier         | Higher with billing     |
| **Best For**            | Development, prototyping | Production, enterprise  |

## Troubleshooting

### "Failed to find default credentials"

```bash
# Run this to set up credentials
gcloud auth application-default login

# Verify credentials exist
ls ~/.config/gcloud/application_default_credentials.json
```

### "Project ID not specified"

Either:
```bash
# Set in gcloud config
gcloud config set project YOUR_PROJECT_ID

# Or set environment variable
export VERTEX_AI_PROJECT_ID="your-project-id"
```

### "Permission denied" errors

```bash
# Grant yourself Vertex AI User role
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
    --member="user:YOUR_EMAIL@example.com" \
    --role="roles/aiplatform.user"
```

### "API not enabled"

```bash
gcloud services enable aiplatform.googleapis.com
```

### Check current configuration

```bash
# Show current project
gcloud config get-value project

# Test authentication
gcloud auth application-default print-access-token

# List available models
gcloud ai models list --region=us-central1
```

## Example API Request

Test the server with Vertex AI:

```bash
# Run server
make run-server

# Test with Vertex AI
curl -X POST http://localhost:8080/diagnose \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "production",
    "llm": "vertexai"
  }'
```

## Best Practices

1. **Use Workload Identity** in GKE instead of service account keys
2. **Set quotas** to prevent unexpected costs
3. **Use cheaper models** (gemini-flash) for non-critical workloads
4. **Enable audit logging** for compliance
5. **Use VPC-SC** for data residency requirements
6. **Cache results** to reduce API calls

## Security Considerations

- Never commit service account keys to git
- Use Workload Identity in Kubernetes
- Rotate service account keys regularly
- Use least-privilege IAM roles
- Enable VPC Service Controls for sensitive data
- Monitor usage with Cloud Billing alerts

## Cost Optimization

```bash
# Use faster, cheaper model for development
export VERTEX_AI_MODEL="gemini-1.5-flash"

# Or use free Ollama for local development
./kubehelp diagnose -n dev --llm ollama

# Reserve Vertex AI for production
./kubehelp diagnose -n prod --llm vertexai
```

## Further Reading

- [Vertex AI Documentation](https://cloud.google.com/vertex-ai/docs)
- [Gemini API on Vertex AI](https://cloud.google.com/vertex-ai/docs/generative-ai/model-reference/gemini)
- [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials)
- [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
