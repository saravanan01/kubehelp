# LLM Provider Comparison

`kubehelp` supports four different LLM providers, each with different trade-offs for cost, performance, and setup complexity.

## Quick Comparison

| Provider      | Type           | Cost             | Setup Complexity | Best For                               |
| ------------- | -------------- | ---------------- | ---------------- | -------------------------------------- |
| **Ollama**    | Local          | Free             | Low              | Development, privacy, offline use      |
| **Gemini**    | Cloud API      | Free tier + paid | Low              | Quick start, free tier users           |
| **Vertex AI** | Enterprise GCP | Pay-per-use      | Medium           | Enterprise GCP users, compliance needs |
| **OpenAI**    | Cloud API      | Paid only        | Low              | Best AI quality, budget available      |

## Provider Details

### 1. Ollama (Default)

**Type**: Local model execution  
**Cost**: Completely free  
**Setup Time**: 5 minutes  

**Pros**:
- ✅ No API keys or accounts needed
- ✅ Completely free, unlimited usage
- ✅ Works offline
- ✅ Data never leaves your machine
- ✅ No rate limits
- ✅ Multiple model options (llama2, mistral, codellama, etc.)

**Cons**:
- ❌ Requires local resources (CPU/RAM)
- ❌ Slower than cloud APIs
- ❌ Quality varies by model size
- ❌ Need to download models (1-7GB each)

**When to Use**:
- Development and testing
- Privacy/security requirements
- Limited or no budget
- Offline environments
- High-volume usage (no per-request costs)

**Setup**:
```bash
# macOS
brew install ollama
ollama serve

# Pull a model
ollama pull mistral  # recommended for troubleshooting
# or
ollama pull llama2
```

**Usage**:
```bash
kubehelp diagnose -n production
# or explicitly
kubehelp diagnose -n production --llm ollama
```

---

### 2. Google Gemini

**Type**: Cloud API  
**Cost**: Free tier (60 requests/minute) + paid tier  
**Setup Time**: 2 minutes  

**Pros**:
- ✅ Generous free tier
- ✅ Fast response times
- ✅ Good quality analysis
- ✅ Simple API key setup
- ✅ No infrastructure needed

**Cons**:
- ❌ Requires Google account and API key
- ❌ Rate limits on free tier
- ❌ Data sent to Google servers
- ❌ Requires internet connection

**When to Use**:
- Quick evaluation without local setup
- Light to moderate usage within free tier
- Good balance of quality and cost
- Already using Google Cloud ecosystem

**Setup**:
```bash
# Get API key from https://makersuite.google.com/app/apikey
export GEMINI_API_KEY="your-key-here"
```

**Usage**:
```bash
kubehelp diagnose -n production --llm gemini
```

**Pricing** (as of 2024):
- Free tier: 60 requests per minute
- Paid: ~$0.00025 per 1K characters
- [Pricing details](https://ai.google.dev/pricing)

See [GEMINI.md](GEMINI.md) for complete setup guide.

---

### 3. Google Vertex AI

**Type**: Enterprise GCP service  
**Cost**: Pay-per-use (similar to Gemini pricing)  
**Setup Time**: 10 minutes  

**Pros**:
- ✅ Enterprise-grade features
- ✅ Integrates with GCP IAM and billing
- ✅ VPC Service Controls support
- ✅ Customer-managed encryption keys (CMEK)
- ✅ No API keys (uses OAuth/ADC)
- ✅ Better for compliance/governance
- ✅ Service accounts and workload identity

**Cons**:
- ❌ Requires GCP account and project
- ❌ More complex authentication setup
- ❌ Paid only (no free tier)
- ❌ Requires gcloud CLI or service account
- ❌ Data sent to Google servers

**When to Use**:
- Enterprise/production GCP environments
- Need compliance features (CMEK, VPC-SC)
- Already using GCP with billing set up
- Want fine-grained IAM controls
- Running in GKE or GCP infrastructure

**Setup**:
```bash
# Authenticate with Google Cloud
gcloud auth application-default login

# Set project
export VERTEX_AI_PROJECT_ID="your-gcp-project"

# Optional: customize region/model
export VERTEX_AI_LOCATION="us-central1"
export VERTEX_AI_MODEL="gemini-pro"
```

**Usage**:
```bash
kubehelp diagnose -n production --llm vertexai
```

**Pricing** (similar to Gemini):
- Input: ~$0.00025 per 1K characters
- Output: ~$0.0005 per 1K characters
- [Vertex AI Pricing](https://cloud.google.com/vertex-ai/pricing)

See [VERTEXAI.md](VERTEXAI.md) for complete setup guide.

---

### 4. OpenAI

**Type**: Cloud API  
**Cost**: Paid only (no free tier)  
**Setup Time**: 2 minutes  

**Pros**:
- ✅ Highest quality AI analysis
- ✅ Fast and reliable
- ✅ Well-established service
- ✅ Simple API key setup

**Cons**:
- ❌ No free tier
- ❌ Can be expensive at scale
- ❌ Data sent to OpenAI servers
- ❌ May be blocked in some organizations

**When to Use**:
- Need best possible AI quality
- Budget is available
- OpenAI not blocked by organization
- Critical production troubleshooting

**Setup**:
```bash
export OPENAI_API_KEY="sk-your-key-here"
```

**Usage**:
```bash
kubehelp diagnose -n production --llm openai
```

**Pricing** (as of 2024, GPT-4):
- ~$0.03 per 1K input tokens
- ~$0.06 per 1K output tokens
- [OpenAI Pricing](https://openai.com/pricing)

---

## Decision Tree

```
Need to troubleshoot K8s?
│
├─ Already on GCP with billing?
│  └─ YES → Use Vertex AI (enterprise features)
│
├─ Have budget and need best quality?
│  └─ YES → Use OpenAI (if not blocked)
│
├─ Want free cloud option for light use?
│  └─ YES → Use Gemini (free tier)
│
└─ Want completely free / privacy / offline?
   └─ YES → Use Ollama (default)
```

## Cost Estimation Examples

### Typical Diagnostic Analysis
- Input: ~2000 characters (pod status + events)
- Output: ~1000 characters (analysis + recommendations)
- Total: ~3000 characters or ~750 tokens

**Per-analysis costs**:
- **Ollama**: $0 (free)
- **Gemini Free**: $0 (within limits)
- **Gemini Paid**: ~$0.0008
- **Vertex AI**: ~$0.0008
- **OpenAI GPT-4**: ~$0.07

**Monthly costs** (100 analyses/month):
- **Ollama**: $0
- **Gemini Free**: $0 (within limits)
- **Gemini Paid**: ~$0.08
- **Vertex AI**: ~$0.08
- **OpenAI GPT-4**: ~$7

## Switching Providers

You can easily switch between providers using the `--llm` flag:

```bash
# Try different providers on the same namespace
kubehelp diagnose -n prod --llm ollama
kubehelp diagnose -n prod --llm gemini
kubehelp diagnose -n prod --llm vertexai
kubehelp diagnose -n prod --llm openai
```

No code changes needed - just set the appropriate environment variables for authentication.

## Recommendations

### For Development
→ **Use Ollama** - Free, fast enough, works offline

### For Personal Projects
→ **Use Gemini** - Free tier is generous, good quality

### For Enterprise/Production
→ **Use Vertex AI** - Enterprise features, GCP integration

### For Critical Issues (Budget Available)
→ **Use OpenAI** - Best quality analysis

---

## More Information

- [Gemini Setup Guide](GEMINI.md)
- [Vertex AI Setup Guide](VERTEXAI.md)
- [Server Deployment Guide](SERVER.md)
- [Main README](../README.md)
