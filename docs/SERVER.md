# kubehelp Server Deployment Guide

This guide covers deploying the kubehelp server (with web UI) locally, via Docker, and to Kubernetes. The server now serves:

- Static Web UI at `/` (HTML/JS single-page form)
- API endpoints at `/api/diagnose` and `/api/health`

## Quick Start

### 1. Build the Server

```bash
# Build locally
go build -o kubehelp-server ./cmd/server

# Run locally
./kubehelp-server
```

Server will start on port 8080 by default.

### 2. Test the Server

```bash
```bash
# Health check
curl http://localhost:8080/api/health

# Run diagnosis (default provider ollama)
curl -X POST http://localhost:8080/api/diagnose \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "default",
    "llm": "ollama"
  }'

# Diagnose specific workloads
### Example Request (Workload scoped)

```bash
curl -X POST http://localhost:8080/api/diagnose \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "production",
    "workloads": ["api-server", "worker"],
    "llm": "ollama"
  }'
```
```

## Docker Deployment

### Build Docker Image

```bash
# Build the image
docker build -t kubehelp-server:latest .

# Run with Docker
docker run -p 8080:8080 \
  -v ~/.kube:/root/.kube:ro \
  -e OLLAMA_BASE_URL=http://host.docker.internal:11434 \
  kubehelp-server:latest
```

### Docker Compose (with Ollama)

Create `docker-compose.yml`:

```yaml
version: '3.8'
services:
  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama-data:/root/.ollama
    
  kubehelp:
    build: .
    ports:
      - "8080:8080"
    environment:
      - OLLAMA_BASE_URL=http://ollama:11434
      - OLLAMA_MODEL=mistral
    volumes:
      - ~/.kube:/root/.kube:ro
    depends_on:
      - ollama

volumes:
  ollama-data:
```

Run with:
```bash
docker-compose up -d
```

## Kubernetes Deployment

### 1. Create Secrets (if using cloud LLMs)

```bash
# For Gemini
kubectl create secret generic llm-secrets \
  --from-literal=gemini-api-key=YOUR_GEMINI_KEY

# For OpenAI
kubectl create secret generic llm-secrets \
  --from-literal=openai-api-key=YOUR_OPENAI_KEY
```

### 2. Build and Push Image

```bash
# Build
docker build -t your-registry/kubehelp-server:v1.0.0 .

# Push to registry
docker push your-registry/kubehelp-server:v1.0.0

# Update deployment.yaml with your image
```

### 3. Deploy to Kubernetes

```bash
# Apply the deployment
kubectl apply -f k8s/deployment.yaml

# Check status
kubectl get pods -l app=kubehelp
kubectl get svc kubehelp-service

# View logs
kubectl logs -l app=kubehelp -f
```

### 4. Access the Service

```bash
# Port-forward to test
kubectl port-forward svc/kubehelp-service 8080:80

# Test
curl http://localhost:8080/health
```

## API Reference

### POST /api/diagnose

Analyze a Kubernetes namespace and return AI-powered insights.

**Request Body:**
```json
{
  "namespace": "string",      // Required: K8s namespace to analyze
  "workloads": ["string"],    // Optional: Specific workload names
  "llm": "string",            // Optional: "ollama"|"gemini"|"openai" (default: ollama)
  "kubeconfig": "string",     // Optional: Path to kubeconfig
  "context": "string"         // Optional: K8s context name
}
```

**Response:**
```json
{
  "analysis": "string",           // LLM analysis with recommendations
  "diagnosticData": {             // Collected K8s data
    "namespace": "string",
    "pods": [...],
    "events": [...]
  }
}
```

**Error Response:**
```json
{
  "error": "Error message"
}
```

### GET /api/health

Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

## Environment Variables

| Variable          | Description           | Default                  |
| ----------------- | --------------------- | ------------------------ |
| `PORT`            | Server port           | `8080`                   |
| `OLLAMA_MODEL`    | Ollama model name     | `mistral`                |
| `OLLAMA_BASE_URL` | Ollama server URL     | `http://localhost:11434` |
| `GEMINI_API_KEY`  | Google Gemini API key | -                        |
| `GEMINI_MODEL`    | Gemini model          | `gemini-pro`             |
| `OPENAI_API_KEY`  | OpenAI API key        | -                        |

## Examples

### Example 1: Local Development (Web UI + API)

```bash
# Terminal 1: Start Ollama
ollama serve

# Terminal 2: Start kubehelp server
go run ./cmd/server

# Terminal 3: Test
curl -X POST http://localhost:8080/api/diagnose \
  -H "Content-Type: application/json" \
  -d '{"namespace": "kube-system", "llm": "ollama"}'
```

### Example 2: Production with Gemini

```bash
# Set API key
export GEMINI_API_KEY="your-key"

# Build and run
docker build -t kubehelp-server .
docker run -p 8080:8080 \
  -e GEMINI_API_KEY=$GEMINI_API_KEY \
  -v ~/.kube:/root/.kube:ro \
  kubehelp-server

# Test
curl -X POST http://localhost:8080/api/diagnose \
  -H "Content-Type: application/json" \
  -d '{"namespace": "production", "llm": "gemini"}'
```

### Example 3: Kubernetes with Ollama Sidecar

Deploy Ollama alongside kubehelp in the same pod:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubehelp-with-ollama
spec:
  template:
    spec:
      containers:
      - name: ollama
        image: ollama/ollama:latest
        ports:
        - containerPort: 11434
      - name: kubehelp
        image: kubehelp-server:latest
        env:
        - name: OLLAMA_BASE_URL
          value: "http://localhost:11434"
```

## Troubleshooting

### Server won't start

Check if port 8080 is already in use:
```bash
lsof -i :8080
# Use different port
PORT=9090 ./kubehelp-server
```

### Can't connect to Kubernetes

Verify kubeconfig:
```bash
kubectl cluster-info
kubectl get nodes
```

### Ollama connection failed

Check Ollama is running:
```bash
curl http://localhost:11434/api/tags
```

### API key errors

Verify environment variables:
```bash
echo $GEMINI_API_KEY
echo $OPENAI_API_KEY
```

## Security Considerations

1. **RBAC**: Service account has read-only access to pods/events
2. **API Keys**: Store in Kubernetes secrets (never bake into images)
3. **Network**: Use NetworkPolicies to restrict traffic
4. **TLS**: Terminate TLS at ingress / gateway (server runs HTTP only)
5. **Rate Limiting**: Add middleware or enforce at ingress
6. **Security Headers**: The server sets CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, X-XSS-Protection
7. **XSS Protection**: Web UI sanitizes all dynamic data (LLM output, pod/event fields)
8. **Secrets**: Prefer mounting secrets as env vars via K8s Secret or using external secret manager

### Recommended Ingress Annotations (Example)
```yaml
nginx.ingress.kubernetes.io/proxy-body-size: "1m"
nginx.ingress.kubernetes.io/limit-rps: "10"
nginx.ingress.kubernetes.io/enable-cors: "true"
```

## Performance Tips

1. **Caching**: Cache recent namespace analyses (e.g., in-memory TTL cache)
2. **Timeouts**: LLM calls already capped; consider shorter timeouts for production
3. **Concurrency**: Use a request semaphore or rate limiter for heavy clusters
4. **Resource Limits**: Define CPU/memory requests/limits in deployment
5. **Pod Affinity**: Co-locate with Ollama if using local model in same node
6. **Compression**: Enable gzip at ingress for JSON responses

## Docker Quick Reference

```bash
# Build image
docker build -t kubehelp-server:latest .

# Run with kubeconfig mount and Ollama on host
docker run -p 8080:8080 \
  -v $HOME/.kube:/root/.kube:ro \
  -e OLLAMA_BASE_URL=http://host.docker.internal:11434 \
  kubehelp-server:latest

# Health
curl http://localhost:8080/api/health

# Diagnose
curl -X POST http://localhost:8080/api/diagnose \
  -H 'Content-Type: application/json' \
  -d '{"namespace":"default","llm":"ollama"}'
```

## Web UI

Access the interactive web UI at `http://localhost:8080/`:

Features:
- Namespace and workload scoped analysis
- LLM provider selector (Ollama, Gemini, Vertex AI, OpenAI)
- Real-time results with pod/event breakdown
- Error handling and status badges

All dynamic content is displayed using sanitized text to prevent injection.
