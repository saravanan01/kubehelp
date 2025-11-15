# kubehelp - AI-Powered Kubernetes Troubleshooting CLI

A command-line tool that uses Large Language Models (LLMs) to help troubleshoot Kubernetes deployments by analyzing cluster state and providing actionable remediation steps.

## Features

- 🔍 **Automated Diagnostics**: Collects pod status, container states, and events from K8s namespaces
- 🤖 **AI Analysis**: Uses LLMs (Ollama, Google Gemini, Vertex AI, OpenAI GPT-4) to analyze issues and suggest fixes
- 🎯 **Targeted Troubleshooting**: Focus on specific workloads or entire namespaces
- 📋 **Actionable Output**: Get root cause analysis, remediation steps, and kubectl commands
- ⚡ **Easy Setup**: Works with your existing kubeconfig
- 🆓 **Multiple LLM Options**: Free local (Ollama) or cloud-based (Gemini free tier, Vertex AI, OpenAI paid)

## Quick Start

### Prerequisites

- Go 1.24 or later
- Access to a Kubernetes cluster (kubeconfig configured)
- **Choose one LLM provider**:
  - **Ollama** (recommended) - Free, local, no API key needed
  - **Google Gemini** - Free tier available, requires API key
  - **Google Vertex AI** - Enterprise GCP option, requires gcloud auth
  - **OpenAI** - Paid, requires API key

### Install Ollama (Recommended)

```bash
# macOS
brew install ollama

# Start Ollama service
ollama serve

# Pull a model (in another terminal)
ollama pull llama2
# or for better results:
ollama pull mistral
```

### Installation

```bash
# Clone the repository
git clone <your-repo-url>
cd kubehelp

# Build the CLI
go build -o kubehelp ./cmd/...

# Move to PATH (optional)
sudo mv kubehelp /usr/local/bin/
```

### Configuration

**Option 1: Use Ollama (Local, Free, No API Key)**

No configuration needed! Just make sure Ollama is running:

```bash
# Check if Ollama is running
curl http://localhost:11434

# If not running, start it
ollama serve
```

**Option 2: Use OpenAI (Cloud, Requires API Key)**

Set your OpenAI API key:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

**Option 3: Use Google Gemini (Cloud, Requires API Key)**

Set your Gemini API key:

```bash
export GEMINI_API_KEY="your-api-key-here"
```

**Option 4: Use Google Vertex AI (Enterprise GCP)**

Set up Google Cloud authentication:

```bash
gcloud auth application-default login
export VERTEX_AI_PROJECT_ID="your-project-id"
```

See [docs/VERTEXAI.md](docs/VERTEXAI.md) for complete setup.

### Usage

```bash
# Analyze an entire namespace (uses Ollama by default)
kubehelp diagnose -n production

# Focus on specific workloads
kubehelp diagnose -n staging -w api-server,worker

# Use a different Ollama model
OLLAMA_MODEL=mistral kubehelp diagnose -n dev

# Use OpenAI instead of Ollama
kubehelp diagnose -n prod --llm openai

# Use Google Gemini
kubehelp diagnose -n prod --llm gemini

# Use Google Vertex AI (enterprise)
kubehelp diagnose -n prod --llm vertexai

# Show verbose diagnostic data
kubehelp diagnose -n dev --verbose

# Use a different kubeconfig or context
kubehelp diagnose -n prod --kubeconfig ~/.kube/prod-config --context prod-cluster
```

## How It Works

1. **Data Collection**: The tool connects to your Kubernetes cluster and collects:
   - Pod status and ready state
   - Container states and restart counts
   - Recent Warning/Error events (last hour)
   - Pod conditions and error messages

2. **Smart Filtering**: Only includes pods and events with actual issues to minimize LLM token usage

3. **LLM Analysis**: Sends structured diagnostic data to the LLM with a prompt requesting:
   - Issue summary
   - Root cause analysis
   - Remediation steps
   - Helpful kubectl commands
   - Prevention strategies

4. **Results**: Displays the AI analysis with actionable insights

## Example Output

```
🔍 Collecting diagnostic data from namespace 'production'...
✅ Collected data: 12 pods, 5 events

🤖 Analyzing with ollama...

=== AI Analysis ===

**Summary of Issues:**
- 2 pods in CrashLoopBackOff state (api-server-7d8f9b)
- Image pull failures for worker pods
- Memory pressure on node-1

**Root Cause Analysis:**
The api-server pods are failing due to a missing database connection
configuration. The worker pods cannot pull images because the registry
credentials secret has expired.

**Remediation Steps:**
1. Update database connection string in api-server ConfigMap
2. Refresh registry credentials: `kubectl create secret docker-registry...`
3. Restart affected deployments

**kubectl Commands:**
```bash
kubectl edit configmap api-server-config -n production
kubectl rollout restart deployment/api-server -n production
```

=== End Analysis ===
```

## Project Structure

```
kubehelp/
├── cmd/
│   ├── main.go          # CLI entry point
│   ├── commands.go      # Basic subcommands (status, logs, events)
│   ├── diagnose.go      # Main AI diagnostic command
│   └── server/
│       └── main.go      # HTTP server for web service deployment
├── internal/
│   ├── k8s/
│   │   ├── client.go    # Kubernetes client wrapper
│   │   └── aggregator.go # Diagnostic data collector
   └── llm/
       ├── provider.go  # LLM provider interface
       ├── openai.go    # OpenAI implementation
       ├── gemini.go    # Google Gemini implementation
       ├── vertexai.go  # Google Vertex AI implementation
       ├── ollama.go    # Ollama implementation
       └── prompts.go   # Prompt engineering
├── k8s/
│   └── deployment.yaml  # Kubernetes deployment manifests
├── docs/
│   ├── GEMINI.md        # Gemini setup guide
│   ├── VERTEXAI.md      # Vertex AI setup guide
│   └── SERVER.md        # Server deployment guide
├── Dockerfile           # Container image definition
└── .github/
    └── copilot-instructions.md # AI coding agent guidelines
```

## Development

### Building

```bash
# Build CLI
go mod tidy
go build -o kubehelp ./cmd/...

# Or use Makefile
make build

# Build server
make build-server
```

### Running as Web Service

```bash
# Run server locally
make run-server

# Or run directly
go run ./cmd/server

# Test the API
curl http://localhost:8080/api/health
curl -X POST http://localhost:8080/api/diagnose \
   -H "Content-Type: application/json" \
   -d '{"namespace": "default", "llm": "ollama"}'
```
# Run with Docker (build first)
docker build -t kubehelp-server:latest .
docker run -p 8080:8080 \
   -v $HOME/.kube:/root/.kube:ro \
   -e OLLAMA_MODEL=mistral \
   kubehelp-server:latest

# Use different LLM provider (example: gemini)
docker run -p 8080:8080 \
   -v $HOME/.kube:/root/.kube:ro \
   -e GEMINI_API_KEY=$GEMINI_API_KEY \
   -e KUBEHELP_LLM_PROVIDER=gemini \
   kubehelp-server:latest

See [docs/SERVER.md](docs/SERVER.md) for complete server deployment guide.

### Testing

```bash
# Run tests (when implemented)
go test ./...

# Test with a real cluster (requires kubeconfig and API key)
./kubehelp diagnose -n kube-system --verbose
```

### Adding New LLM Providers

1. Implement the `llm.Provider` interface:
   ```go
   type Provider interface {
       Analyze(ctx context.Context, prompt string) (string, error)
       Name() string
   }
   ```

2. Add provider creation in `cmd/diagnose.go`:
   ```go
   case "anthropic":
       provider = llm.NewAnthropicProvider(apiKey, "claude-3")
   ```

3. Update documentation and environment variables

## Environment Variables

| Variable               | Description                             | Default                  |
| ---------------------- | --------------------------------------- | ------------------------ |
| `OLLAMA_MODEL`         | Ollama model to use                     | `mistral` (recommended)  |
| `OLLAMA_BASE_URL`      | Ollama server URL                       | `http://localhost:11434` |
| `KUBEHELP_API_KEY`     | Generic API key for cloud LLM providers | -                        |
| `OPENAI_API_KEY`       | OpenAI-specific API key                 | -                        |
| `GEMINI_API_KEY`       | Google Gemini API key                   | -                        |
| `GEMINI_MODEL`         | Gemini model to use                     | `gemini-pro`             |
| `VERTEX_AI_PROJECT_ID` | GCP project ID for Vertex AI            | Auto-detected            |
| `VERTEX_AI_LOCATION`   | Vertex AI location/region               | `us-central1`            |
| `VERTEX_AI_MODEL`      | Vertex AI model name                    | `gemini-pro`             |
| `KUBECONFIG`           | Path to kubeconfig file                 | `~/.kube/config`         |

## Command-Line Flags

### `diagnose` command

| Flag           | Short | Description                                     | Default         |
| -------------- | ----- | ----------------------------------------------- | --------------- |
| `--namespace`  | `-n`  | Target namespace                                | `default`       |
| `--workload`   | `-w`  | Specific workloads (comma-separated)            | All workloads   |
| `--verbose`    | -     | Show raw diagnostic data                        | `false`         |
| `--llm`        | -     | LLM provider (ollama, gemini, vertexai, openai) | `ollama`        |
| `--kubeconfig` | -     | Path to kubeconfig                              | `$KUBECONFIG`   |
| `--context`    | -     | Kubernetes context to use                       | Current context |

## Roadmap

- [x] Add support for local LLMs (Ollama)
- [x] Add support for Google Gemini
- [x] Add support for Google Vertex AI
- [ ] Add support for Anthropic Claude
- [ ] Implement caching to reduce API calls
- [ ] Add log analysis capabilities
- [ ] Support for multi-cluster diagnostics
- [ ] Export reports in JSON/Markdown format
- [ ] Add cost estimation for cloud LLM calls

## Contributing

Contributions are welcome! Please follow the existing code patterns:

- Organize code by feature, not by type
- Wrap errors with context
- Add examples to documentation
- Test with real clusters when possible

## License

[Add your license here]

## Support

For issues, questions, or contributions, please [open an issue](link-to-issues).

---

Built with ❤️ using Go, Kubernetes client-go, and LLMs

---

## Development Workflow

```bash
# Format code
go fmt ./...

# Lint / vet
go vet ./...

# Build CLI & server
make build
make build-server

# Run server
make run-server

# Run tests (when implemented)
go test ./...
```

## Recommended Ollama Models

| Model     | Size  | Notes                                   |
| --------- | ----- | --------------------------------------- |
| mistral   | ~4GB  | Balanced quality & speed (default here) |
| llama2    | ~3.8G | Stable general model                    |
| codellama | ~3.8G | Better for code / technical prompts     |
| llama3    | ~4.7G | Newer, improved reasoning               |

Pull & use:
```bash
ollama pull mistral
OLLAMA_MODEL=mistral kubehelp diagnose -n production
```

## File Overview

```
.
├── cmd/                 # CLI & server entrypoints
│   ├── main.go
│   ├── diagnose.go
│   └── server/main.go
├── internal/
│   ├── k8s/             # Kubernetes access & aggregation
│   │   ├── client.go
│   │   └── aggregator.go
│   └── llm/             # LLM providers & prompt engineering
│       ├── provider.go
│       ├── ollama.go
│       ├── openai.go
│       ├── gemini.go
│       ├── vertexai.go
│       └── prompts.go
├── web/                 # Static web UI (HTML/JS)
├── docs/                # Provider & server docs
├── Dockerfile           # Container build
├── Makefile             # Common tasks
└── .github/copilot-instructions.md
```

## Docker Usage Quick Reference

```bash
# Build image
docker build -t kubehelp-server:latest .

# Run with mounted kubeconfig (read-only)
docker run -p 8080:8080 -v $HOME/.kube:/root/.kube:ro kubehelp-server:latest

# Check health
curl http://localhost:8080/api/health

# Run diagnose
curl -X POST http://localhost:8080/api/diagnose \
   -H 'Content-Type: application/json' \
   -d '{"namespace":"default","llm":"ollama"}'
```

## Notes
- Web UI served at `/` (static files)
- API endpoints under `/api/*`
- For Vertex AI ensure `gcloud auth application-default login` completed locally before docker run (mounting ADC credentials or run inside environment with access).
- Avoid embedding secrets in the image; supply via `-e` flags.

