# Setup Instructions

## Quick Setup

Run these commands to get the project ready:

```bash
# 1. Navigate to the project directory
cd /Users/v242928/src/devl/kubehelp

# 2. Download and install all dependencies
go mod tidy

# 3. Build the project
go build -o kubehelp ./cmd/...

# 4. Verify the build
./kubehelp --help
```

## LLM Provider Setup

### Option 1: Ollama (Recommended - Local, Free, No API Key)

```bash
# Install Ollama
brew install ollama

# Start Ollama service (in one terminal)
ollama serve

# Pull a model (in another terminal)
#ollama pull llama2

# Or for better results with troubleshooting:
ollama pull mistral
#ollama pull codellama

# Verify it's working
curl http://localhost:11434/api/tags
```

### Option 2: OpenAI (Cloud-based, Requires API Key)

```bash
export OPENAI_API_KEY="sk-your-key-here"
```

### Option 3: Google Gemini (Cloud-based, Requires API Key)

```bash
# Get your API key from https://makersuite.google.com/app/apikey
export GEMINI_API_KEY="your-gemini-key-here"

# Optional: specify model (default is gemini-pro)
export GEMINI_MODEL="gemini-pro"
```

## Next Steps

1. **Test with your cluster**:
   ```bash
   # Make sure you have a valid kubeconfig
   kubectl config current-context
   
   # Run a diagnostic on a namespace (uses Ollama by default)
   ./kubehelp diagnose -n kube-system --verbose
   
   # Or use OpenAI
   ./kubehelp diagnose -n kube-system --llm openai
   
   # Or use Google Gemini
   ./kubehelp diagnose -n kube-system --llm gemini
   ```

2. **Install globally (optional)**:
   ```bash
   make install
   # or
   sudo mv kubehelp /usr/local/bin/
   ```

## Troubleshooting

### Dependencies not found

If you see import errors, run:
```bash
go mod tidy
go mod download
```

### Build errors

Make sure you're using Go 1.21 or later:
```bash
go version
```

### Kubernetes connection issues

Verify your kubeconfig:
```bash
kubectl cluster-info
kubectl get namespaces
```

## Development Workflow

```bash
# Format code
make fmt

# Run linters
make vet

# Build
make build

# Run tests (when implemented)
make test

# Clean build artifacts
make clean
```

## Project Status

✅ Core implementation complete:
- Kubernetes client wrapper
- Diagnostic data aggregator  
- LLM provider interface with Ollama (default) and OpenAI support
- Prompt engineering
- Diagnose command with flags
- CLI structure and help

🚧 Next steps:
- Add unit tests
- Add more LLM providers (Anthropic)
- Implement caching
- Add log analysis
- Performance optimizations

## Recommended Models

### Ollama Models (Local)
- **llama2** (default) - Good general purpose, 3.8GB
- **mistral** - Better reasoning, 4.1GB (recommended)
- **codellama** - Specialized for code/technical content, 3.8GB
- **llama3** - Latest version, best quality, 4.7GB

```bash
# Pull recommended model
ollama pull mistral

# Use it
OLLAMA_MODEL=mistral ./kubehelp diagnose -n production
```

## Files Created

```
.
├── README.md
├── Makefile
├── SETUP.md (this file)
├── go.mod
├── go.sum
├── .github/
│   └── copilot-instructions.md
├── cmd/
│   ├── main.go
│   ├── commands.go
│   └── diagnose.go
└── internal/
    ├── k8s/
    │   ├── client.go
    │   └── aggregator.go
    └── llm/
        ├── provider.go
        ├── openai.go
        ├── gemini.go
        ├── ollama.go
        └── prompts.go
```
