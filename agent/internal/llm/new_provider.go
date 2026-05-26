package llm

import "net/http"

// LLMConfig is a simple config struct used to create a Provider.
type LLMConfig struct {
	Provider string
	Model    string
	APIKey   string
	BaseURL  string
}

// NewProviderFromConfig creates a Provider from a LLMConfig.
func NewProviderFromConfig(cfg LLMConfig) Provider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		switch cfg.Provider {
		case "openai", "openai_compatible":
			baseURL = "https://api.openai.com/v1"
		case "anthropic":
			baseURL = "https://api.anthropic.com"
		default:
			baseURL = "https://api.openai.com/v1"
		}
	}
	return &openAICompatProvider{
		model:   cfg.Model,
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		client:  &http.Client{},
	}
}
