# kubehelp - AI-Powered Kubernetes Troubleshooting CLI

A command-line tool that uses Large Language Models (LLMs) to help troubleshoot Kubernetes deployments by analyzing cluster state and providing actionable remediation steps.

## Features

- 🔍 **Automated Diagnostics**: Collects pod status, container states, and events from K8s namespaces
- 🤖 **AI Analysis**: Uses LLMs (Ollama, Google Gemini, OpenAI GPT-4) to analyze issues and suggest fixes
- 🎯 **Targeted Troubleshooting**: Focus on specific workloads or entire namespaces
- 📋 **Actionable Output**: Get root cause analysis, remediation steps, and kubectl commands
- ⚡ **Easy Setup**: Works with your existing kubeconfig
- 🆓 **Multiple LLM Options**: Free local (Ollama) or cloud-based (Gemini free tier, OpenAI paid)

## Quick Start

### Prerequisites

- Go 1.24 or later
- Access to a Kubernetes cluster (kubeconfig configured)
- **Choose one LLM provider**:
  - **Ollama** (recommended) - Free, local, no API key needed
  - **Google Gemini** - Free tier available, requires API key
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
│   └── diagnose.go      # Main AI diagnostic command
├── internal/
│   ├── k8s/
│   │   ├── client.go    # Kubernetes client wrapper
│   │   └── aggregator.go # Diagnostic data collector
│   └── llm/
│       ├── provider.go  # LLM provider interface
│       ├── openai.go    # OpenAI implementation
│       └── prompts.go   # Prompt engineering
└── .github/
    └── copilot-instructions.md # AI coding agent guidelines
```

## Development

### Building

```bash
go mod tidy
go build -o kubehelp ./cmd/...
```

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

| Variable           | Description                             | Default                  |
| ------------------ | --------------------------------------- | ------------------------ |
| `OLLAMA_MODEL`     | Ollama model to use                     | `llama2`                 |
| `OLLAMA_BASE_URL`  | Ollama server URL                       | `http://localhost:11434` |
| `KUBEHELP_API_KEY` | Generic API key for cloud LLM providers | -                        |
| `OPENAI_API_KEY`   | OpenAI-specific API key                 | -                        |
| `GEMINI_API_KEY`   | Google Gemini API key                   | -                        |
| `GEMINI_MODEL`     | Gemini model to use                     | `gemini-pro`             |
| `KUBECONFIG`       | Path to kubeconfig file                 | `~/.kube/config`         |

## Command-Line Flags

### `diagnose` command

| Flag           | Short | Description                          | Default         |
| -------------- | ----- | ------------------------------------ | --------------- |
| `--namespace`  | `-n`  | Target namespace                     | `default`       |
| `--workload`   | `-w`  | Specific workloads (comma-separated) | All workloads   |
| `--verbose`    | -     | Show raw diagnostic data             | `false`         |
| `--llm`        | -     | LLM provider (openai, ollama)        | `ollama`        |
| `--kubeconfig` | -     | Path to kubeconfig                   | `$KUBECONFIG`   |
| `--context`    | -     | Kubernetes context to use            | Current context |

## Roadmap

- [x] Add support for local LLMs (Ollama)
- [x] Add support for Google Gemini
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

Built with ❤️ using Go, Kubernetes client-go, and OpenAI
