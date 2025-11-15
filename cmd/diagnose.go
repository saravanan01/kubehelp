package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"kubehelp/internal/k8s"
	"kubehelp/internal/llm"

	"github.com/spf13/cobra"
)

var (
	diagNamespace   string
	diagWorkloads   []string
	diagVerbose     bool
	diagLLMProvider string
	diagKubeconfig  string
	diagContext     string
)

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "AI-powered troubleshooting for Kubernetes workloads",
	Long: `Diagnose collects diagnostic data from a Kubernetes namespace and uses 
an LLM to analyze issues and provide actionable remediation steps.

The command gathers pod status, container states, and recent events, then 
sends this information to an LLM for analysis.

Environment variables:
  KUBEHELP_LLM_PROVIDER - LLM provider (openai, gemini, ollama)
  KUBEHELP_API_KEY      - API key for cloud LLM providers
  GEMINI_API_KEY        - Google Gemini API key
  GEMINI_MODEL          - Gemini model to use (default: gemini-pro)
  OLLAMA_MODEL          - Ollama model to use (default: llama2)
  OLLAMA_BASE_URL       - Ollama server URL (default: http://localhost:11434)
  KUBECONFIG            - Path to kubeconfig file`,
	Example: `  # Analyze entire namespace
  kubehelp diagnose -n production

  # Focus on specific workloads
  kubehelp diagnose -n staging -w api-server,worker

  # Use Ollama (local, no API key needed)
  kubehelp diagnose -n dev --llm ollama

  # Use Google Gemini
  kubehelp diagnose -n prod --llm gemini

  # Use custom Ollama model
  OLLAMA_MODEL=mistral kubehelp diagnose -n prod

  # Show verbose diagnostic data
  kubehelp diagnose -n prod --verbose`,
	RunE: runDiagnose,
}

func init() {
	diagnoseCmd.Flags().StringVarP(&diagNamespace, "namespace", "n", "default", "Target namespace to diagnose")
	diagnoseCmd.Flags().StringSliceVarP(&diagWorkloads, "workload", "w", []string{}, "Specific workloads to analyze (comma-separated)")
	diagnoseCmd.Flags().BoolVar(&diagVerbose, "verbose", false, "Show raw diagnostic data before analysis")
	diagnoseCmd.Flags().StringVar(&diagLLMProvider, "llm", "ollama", "LLM provider: openai, gemini, ollama")
	diagnoseCmd.Flags().StringVar(&diagKubeconfig, "kubeconfig", "", "Path to kubeconfig file (default: $KUBECONFIG or ~/.kube/config)")
	diagnoseCmd.Flags().StringVar(&diagContext, "context", "", "Kubernetes context to use")
}

func runDiagnose(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create Kubernetes client
	k8sClient, err := k8s.NewClient(diagKubeconfig, diagContext)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	fmt.Printf("🔍 Collecting diagnostic data from namespace '%s'...\n", diagNamespace)

	// Create aggregator and collect data
	aggregator := k8s.NewAggregator(k8sClient)
	data, err := aggregator.CollectDiagnostics(ctx, diagNamespace, diagWorkloads)
	if err != nil {
		return fmt.Errorf("failed to collect diagnostics: %w", err)
	}

	fmt.Printf("✅ Collected data: %d pods, %d events\n\n", len(data.Pods), len(data.Events))

	// Build diagnostic prompt
	prompt := llm.BuildDiagnosticPrompt(data)

	// Show verbose output if requested
	if diagVerbose {
		fmt.Println("=== Raw Diagnostic Data ===")
		fmt.Println(prompt)
		fmt.Println("=== End Raw Data ===\n")
	}

	// Get LLM provider configuration
	apiKey := os.Getenv("KUBEHELP_API_KEY")
	if apiKey == "" {
		// Try provider-specific env vars
		switch diagLLMProvider {
		case "openai":
			apiKey = os.Getenv("OPENAI_API_KEY")
		case "anthropic":
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		case "gemini":
			apiKey = os.Getenv("GEMINI_API_KEY")
		}
	}

	// API key not required for Ollama (local)
	if apiKey == "" && diagLLMProvider != "ollama" {
		return fmt.Errorf("API key not found. Set KUBEHELP_API_KEY or %s_API_KEY environment variable",
			strings.ToUpper(diagLLMProvider))
	}

	// Create LLM provider
	var provider llm.Provider
	switch diagLLMProvider {
	case "openai":
		provider = llm.NewOpenAIProvider(apiKey, "gpt-4")
	case "gemini":
		// Get model from env or use default
		model := os.Getenv("GEMINI_MODEL")
		if model == "" {
			model = "gemini-pro" // default model
		}
		provider = llm.NewGeminiProvider(apiKey, model)
	case "ollama":
		// Get model and base URL from env or use defaults
		model := os.Getenv("OLLAMA_MODEL")
		if model == "" {
			model = "llama2" // default model
		}
		baseURL := os.Getenv("OLLAMA_BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:11434" // default Ollama URL
		}
		provider = llm.NewOllamaProvider(model, baseURL)
	default:
		return fmt.Errorf("unsupported LLM provider: %s (supported: openai, gemini, ollama)", diagLLMProvider)
	}

	fmt.Printf("🤖 Analyzing with %s...\n\n", provider.Name())

	// Get analysis from LLM
	analysis, err := provider.Analyze(ctx, prompt)
	if err != nil {
		return fmt.Errorf("LLM analysis failed: %w", err)
	}

	// Display results
	fmt.Println("=== AI Analysis ===")
	fmt.Println(analysis)
	fmt.Println("=== End Analysis ===")

	return nil
}
