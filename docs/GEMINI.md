# Google Gemini Setup Guide

## Getting Your API Key

1. Visit [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Sign in with your Google account
3. Click "Create API Key"
4. Copy your API key

## Configuration

```bash
# Set your Gemini API key
export GEMINI_API_KEY="your-api-key-here"

# Optional: Use a different model (default is gemini-pro)
export GEMINI_MODEL="gemini-pro"
```

## Available Models

- **gemini-pro** (default) - Best for text generation and analysis
- **gemini-pro-vision** - For image and text inputs (not needed for kubehelp)

## Usage Examples

```bash
# Use Gemini for diagnosis
kubehelp diagnose -n production --llm gemini

# Use Gemini with verbose output
kubehelp diagnose -n staging --llm gemini --verbose

# Focus on specific workloads with Gemini
kubehelp diagnose -n prod -w api-server,worker --llm gemini
```

## Pricing

Google Gemini offers:
- **Free tier**: 60 requests per minute
- Generous quota for development and testing
- Check current pricing at: https://ai.google.dev/pricing

## Advantages

- ✅ **Free tier** with good limits
- ✅ **Fast responses** (comparable to GPT-4)
- ✅ **Good at technical analysis**
- ✅ **Easy to get started**
- ✅ **No credit card required** for free tier

## Comparison with Other Providers

| Feature     | Ollama                 | Gemini           | OpenAI       |
| ----------- | ---------------------- | ---------------- | ------------ |
| **Cost**    | Free (local)           | Free tier + paid | Paid only    |
| **Privacy** | 100% local             | Cloud-based      | Cloud-based  |
| **Speed**   | Depends on hardware    | Fast             | Fast         |
| **Quality** | Model-dependent        | High             | High         |
| **Setup**   | Requires local install | API key only     | API key only |

## Troubleshooting

### "API key not found" error

Make sure you've exported the environment variable:
```bash
echo $GEMINI_API_KEY
```

If empty, set it:
```bash
export GEMINI_API_KEY="your-key-here"
```

### Rate limit errors

If you hit rate limits, either:
1. Wait a minute and try again
2. Switch to Ollama for unlimited local usage
3. Upgrade to a paid Gemini plan

### "Invalid API key" error

1. Verify your API key at https://makersuite.google.com/app/apikey
2. Make sure you copied the entire key
3. Check for extra spaces or quotes

## Best Practices

1. **Use Ollama for development** - Save API calls, work offline
2. **Use Gemini for production** - More consistent, faster responses
3. **Cache results** - Don't analyze the same namespace repeatedly
4. **Use verbose mode** - See what data is sent to Gemini

## Example Workflow

```bash
# Development: Use Ollama (free, local)
./kubehelp diagnose -n dev --llm ollama

# Staging: Use Gemini free tier
export GEMINI_API_KEY="your-key"
./kubehelp diagnose -n staging --llm gemini

# Production: Same as staging
./kubehelp diagnose -n production --llm gemini --verbose
```
