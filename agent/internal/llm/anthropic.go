package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	anthropicAPIURL = "https://api.anthropic.com/v1/messages"
	anthropicVersion = "2023-06-01"
)

// AnthropicProvider implements Provider for Anthropic Claude models
type AnthropicProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewAnthropic(apiKey, model string) *AnthropicProvider {
	if model == "" {
		model = "claude-sonnet-4-5"
	}
	return &AnthropicProvider{apiKey: apiKey, model: model, client: &http.Client{}}
}

func (a *AnthropicProvider) Name() string  { return "anthropic" }
func (a *AnthropicProvider) Model() string { return a.model }

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Tools     []anthropicTool    `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string or []block
}

type anthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type anthropicResponse struct {
	Content []struct {
		Type  string          `json:"type"`
		Text  string          `json:"text,omitempty"`
		ID    string          `json:"id,omitempty"`
		Name  string          `json:"name,omitempty"`
		Input json.RawMessage `json:"input,omitempty"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
}

func (a *AnthropicProvider) Complete(ctx context.Context, messages []Message, tools []ToolDefinition) (*Response, error) {
	var systemPrompt string
	var anthropicMsgs []anthropicMessage

	for _, m := range messages {
		if m.Role == RoleSystem {
			systemPrompt = m.Content
			continue
		}
		anthropic Msgs = append(anthropicMsgs, anthropicMessage{Role: m.Role, Content: m.Content})
	}

	var anthropicTools []anthropicTool
	for _, t := range tools {
		anthropic Tools = append(anthropicTools, anthropicTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Parameters,
		})
	}

	req := anthropicRequest{
		Model:     a.model,
		MaxTokens: 8192,
		System:    systemPrompt,
		Messages:  anthropicMsgs,
		Tools:     anthropicTools,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("x-api-key", a.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)
	httpReq.Header.Set("content-type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic API %d: %s", resp.StatusCode, respBody)
	}

	var ar anthropicResponse
	if err := json.Unmarshal(respBody, &ar); err != nil {
		return nil, err
	}

	result := &Response{Done: ar.StopReason == "end_turn"}
	for _, block := range ar.Content {
		switch block.Type {
		case "text":
			result.Content += block.Text
		case "tool_use":
			var args map[string]interface{}
			if block.Input != nil {
				_ = json.Unmarshal(block.Input, &args)
			}
			result.ToolCalls = append(result.ToolCalls, ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: args,
			})
		}
	}
	return result, nil
}
