package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"kubehelp/internal/k8s"
	"kubehelp/internal/llm"
)

type DiagnoseRequest struct {
	Namespace   string   `json:"namespace"`
	Workloads   []string `json:"workloads,omitempty"`
	LLMProvider string   `json:"llm,omitempty"` // defaults to "ollama"
	Context     string   `json:"context,omitempty"`
}

type DiagnoseResponse struct {
	Analysis       string              `json:"analysis"`
	DiagnosticData *k8s.DiagnosticData `json:"diagnosticData,omitempty"`
	Error          string              `json:"error,omitempty"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func diagnoseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DiagnoseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if req.LLMProvider == "" {
		req.LLMProvider = "ollama"
	}

	log.Printf("Diagnosing namespace: %s, workloads: %v, llm: %s", req.Namespace, req.Workloads, req.LLMProvider)

	// Create K8s client
	client, err := k8s.NewClient("", req.Context)
	if err != nil {
		respondWithError(w, "Failed to create Kubernetes client: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Collect diagnostics
	aggregator := k8s.NewAggregator(client)
	data, err := aggregator.CollectDiagnostics(context.Background(), req.Namespace, req.Workloads)
	if err != nil {
		respondWithError(w, "Failed to collect diagnostics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Collected data: %d pods, %d events", len(data.Pods), len(data.Events))

	// Build prompt
	prompt := llm.BuildDiagnosticPrompt(data)

	// Get LLM provider
	provider, err := createLLMProvider(req.LLMProvider)
	if err != nil {
		respondWithError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Analyzing with %s...", provider.Name())

	// Get analysis from LLM
	analysis, err := provider.Analyze(context.Background(), prompt)
	if err != nil {
		respondWithError(w, "LLM analysis failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send successful response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DiagnoseResponse{
		Analysis:       analysis,
		DiagnosticData: data,
	})
}

func createLLMProvider(providerName string) (llm.Provider, error) {
	switch providerName {
	case "ollama":
		model := getEnv("OLLAMA_MODEL", "mistral")
		baseURL := getEnv("OLLAMA_BASE_URL", "http://localhost:11434")
		return llm.NewOllamaProvider(model, baseURL), nil

	case "gemini":
		apiKey := getEnv("GEMINI_API_KEY", "")
		if apiKey == "" {
			return nil, jsonError("GEMINI_API_KEY environment variable not set")
		}
		model := getEnv("GEMINI_MODEL", "gemini-pro")
		return llm.NewGeminiProvider(apiKey, model), nil

	case "openai":
		apiKey := getEnv("OPENAI_API_KEY", "")
		if apiKey == "" {
			return nil, jsonError("OPENAI_API_KEY environment variable not set")
		}
		return llm.NewOpenAIProvider(apiKey, "gpt-4"), nil

	case "vertexai":
		vertexProvider, err := llm.NewVertexAIProviderFromEnv()
		if err != nil {
			return nil, jsonError("Failed to create Vertex AI provider: " + err.Error())
		}
		return vertexProvider, nil

	default:
		return nil, jsonError("Unsupported LLM provider: " + providerName + " (supported: ollama, gemini, openai, vertexai)")
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{
		Status:  "healthy",
		Version: "1.0.0",
	})
}

func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(DiagnoseResponse{
		Error: message,
	})
}

func jsonError(message string) error {
	return &ErrorWithMessage{message}
}

type ErrorWithMessage struct {
	message string
}

func (e *ErrorWithMessage) Error() string {
	return e.message
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent XSS attacks
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:")
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		// Control referrer information
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Prevent XSS in older browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/diagnose", diagnoseHandler)
	mux.HandleFunc("/api/health", healthHandler)

	// Serve static web UI at root
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	// Wrap with middlewares (security headers applied first)
	handler := loggingMiddleware(corsMiddleware(securityHeadersMiddleware(mux)))

	port := getEnv("PORT", "8080")
	log.Printf("🚀 kubehelp server starting on port %s", port)
	log.Printf("📍 Endpoints:")
	log.Printf("   Web UI:  http://localhost:%s/", port)
	log.Printf("   POST     http://localhost:%s/api/diagnose - Run diagnosis", port)
	log.Printf("   GET      http://localhost:%s/api/health - Health check", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
