package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Config holds provider settings (mirrors app.LLMConfigJSON).
type Config interface {
	GetProvider() string
	GetModel() string
	GetAPIKey() string
	GetBaseURL() string
}

// openAICompatProvider works with any OpenAI-compatible API.
type openAICompatProvider struct {
	model   string
	apiKey  string
	baseURL string
	client  *http.Client
}

func (p *openAICompatProvider) GetProvider() string { return "openai_compatible" }
func (p *openAICompatProvider) GetModel() string    { return p.model }
func (p *openAICompatProvider) GetAPIKey() string   { return p.apiKey }
func (p *openAICompatProvider) GetBaseURL() string  { return p.baseURL }

// cfgAdapter adapts LLMConfigJSON (from app package) without creating a circular import.
type cfgAdapter struct {
	provider string
	model    string
	apiKey   string
	baseURL  string
}

// NewProvider creates the appropriate Provider from a config struct.
// cfg must have fields: Provider, Model, APIKey, BaseURL (string fields).
func NewProvider(cfg interface{ 
	GetProviderStr() string
	GetModelStr() string
	GetAPIKeyStr() string
	GetBaseURLStr() string
}) Provider {
	p := &openAICompatProvider{
		model:   cfg.GetModelStr(),
		apiKey:  cfg.GetAPIKeyStr(),
		baseURL: cfg.GetBaseURLStr(),
		client:  &http.Client{},
	}
	if p.baseURL == "" {
		p.baseURL = "https://api.openai.com/v1"
	}
	return p
}
