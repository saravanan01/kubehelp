# Copilot Instructions for kubehelp CLI

## Project Overview
- **kubehelp** is a CLI tool that uses AI/LLM to troubleshoot Kubernetes deployments
- Collects diagnostic data (pods, events, container states) from K8s clusters and sends it to an LLM for analysis
- Provides actionable remediation steps, root cause analysis, and kubectl commands to fix issues
- Built with Go 1.24, using k8s.io/client-go v0.34.2 and cobra CLI framework

## Architecture & Key Components

### Core Modules
- **`cmd/`** — CLI entrypoints and subcommand definitions (main.go, commands.go, diagnose.go)
- **`internal/k8s/`** — Kubernetes client wrapper and diagnostic data aggregator
  - `client.go`: Kubeconfig loading, context management, clientset creation
  - `aggregator.go`: Collects pod status, container states, events into structured DiagnosticData
- **`internal/llm/`** — LLM provider abstraction and prompt engineering
  - `provider.go`: Interface for LLM providers (OpenAI, Anthropic, etc.)
  - `openai.go`: OpenAI API implementation with retry logic
  - `prompts.go`: Converts K8s diagnostic data into structured LLM prompts

### Data Flow
1. User runs `kubehelp diagnose -n <namespace> -w <workload>`
2. K8s aggregator collects pod info, container statuses, recent warning/error events
3. Prompt builder formats data as markdown tables with issue summaries
4. LLM provider analyzes and returns: issue summary, root cause, remediation, prevention tips
5. Results displayed to user with clear formatting

## Developer Workflows

### Build & Run
```bash
go mod tidy                              # Install/update dependencies
go build -o kubehelp ./cmd/...          # Build the CLI
./kubehelp diagnose -n production       # Run diagnose command
```

### Environment Variables
- `KUBEHELP_API_KEY` or `OPENAI_API_KEY` — LLM provider API key (required for OpenAI)
- `GEMINI_API_KEY` — Google Gemini API key (required for Gemini)
- `GEMINI_MODEL` — Gemini model name (default: gemini-pro)
- `OLLAMA_MODEL` — Ollama model name (default: llama2)
- `OLLAMA_BASE_URL` — Ollama server URL (default: http://localhost:11434)
- `KUBECONFIG` — Path to kubeconfig (optional, defaults to ~/.kube/config)

### Testing Strategy
- Test K8s aggregator with mock clientsets
- Test prompt generation with sample DiagnosticData
- Test LLM provider with HTTP mock responses
- Integration tests require valid kubeconfig + API key

## Project Conventions

### Command Structure
- All commands follow: `kubehelp <command> [flags]`
- Flags: `-n/--namespace`, `-w/--workload`, `--verbose`, `--llm`, `--context`
- Use cobra for command parsing and flag management

### Error Handling
- Wrap all errors with context: `fmt.Errorf("failed to X: %w", err)`
- User-facing errors are actionable (e.g., "Set OPENAI_API_KEY environment variable")
- K8s API errors include namespace/resource context

### Code Organization
- Organize by feature, not type (all diagnose logic together)
- Keep K8s client operations in `internal/k8s/`
- LLM interactions isolated in `internal/llm/`
- Each subcommand in separate file under `cmd/`

## Integration Points

### Kubernetes Client
- Uses official k8s.io/client-go library
- Loads kubeconfig from default locations or explicit path
- Supports custom contexts via `--context` flag
- Aggregator filters events by time (last hour) and severity (Warning/Error)

### LLM Providers
- Abstract interface allows multiple providers (OpenAI, Gemini, Ollama)
- **Default provider: Ollama (local, no API key required)**
- OpenAI with GPT-4 (requires OPENAI_API_KEY)
- Google Gemini with gemini-pro (requires GEMINI_API_KEY)
- Ollama supports local models (llama2, mistral, etc.)
- Timeout: 60s for cloud, 120s for local models
- Prompts include system message defining expert role

### Diagnostic Data Collection
- Pods: name, phase, ready status, restart count, container states
- Events: type, reason, message, count, timestamps (last hour only)
- Filters: optional workload names (prefix matching on pod names)
- **Critical**: Only include pods/events with actual issues to reduce token usage

## Examples

### Basic Usage
```bash
# Analyze entire namespace (uses Ollama by default)
kubehelp diagnose -n production

# Focus on specific workloads
kubehelp diagnose -n staging -w api-server,worker

# Use different Ollama model
OLLAMA_MODEL=mistral kubehelp diagnose -n dev

# Use OpenAI instead
kubehelp diagnose -n prod --llm openai

# Use Google Gemini
kubehelp diagnose -n prod --llm gemini

# Show raw diagnostic data before LLM analysis
kubehelp diagnose -n dev --verbose

# Use different kubeconfig/context
kubehelp diagnose -n prod --kubeconfig ~/.kube/prod-config --context prod-us-west
```

### Expected Output
```
🔍 Collecting diagnostic data from namespace 'production'...
✅ Collected data: 12 pods, 5 events

🤖 Analyzing with ollama...

=== AI Analysis ===
[LLM provides structured analysis with issues, root cause, remediation]
=== End Analysis ===
```

## Key Files Reference
- `cmd/main.go` — Registers all subcommands
- `cmd/diagnose.go` — Main diagnose command implementation (~120 lines)
- `internal/k8s/aggregator.go` — Data collection logic (~250 lines)
- `internal/llm/prompts.go` — Prompt engineering and formatting (~150 lines)
- `internal/llm/openai.go` — OpenAI API client (~90 lines)

## Adding New Features

### New LLM Provider
1. Implement `llm.Provider` interface (Analyze, Name methods)
2. Add provider creation in `cmd/diagnose.go` switch statement
3. Update environment variable documentation

### New Diagnostic Data Source
1. Add collection method to `internal/k8s/aggregator.go`
2. Extend `DiagnosticData` struct with new fields
3. Update `llm/prompts.go` to include new data in prompt

### New Subcommand
1. Create new file in `cmd/` (e.g., `cmd/logs.go`)
2. Define cobra.Command with Use, Short, RunE
3. Register in `cmd/main.go`: `rootCmd.AddCommand(logsCmd)`

---
**Note**: This tool focuses on LLM-driven analysis. For raw K8s data, users should use `kubectl` directly.