package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OllamaProvider implements Provider for local Ollama models
type OllamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllama(baseURL, model string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.1"
	}
	return &OllamaProvider{baseURL: baseURL, model: model, client: &http.Client{}}
}

func (o *OllamaProvider) Name() string  { return "ollama" }
func (o *OllamaProvider) Model() string { return o.model }

type ollamaRequest struct {
	Model    string           `json:"model"`
	Messages []ollamaMessage  `json:"messages"`
	Stream   bool             `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

func (o *OllamaProvider) Complete(ctx context.Context, messages []Message, tools []ToolDefinition) (*Response, error) {
	ollamaMsgs := make([]ollamaMessage, 0)

	// Inject tool descriptions into the system message (Ollama has no native tool support)
	if len(tools) > 0 {
		toolsJSON, _ := json.MarshalIndent(tools, "", "  ")
		sysContent := fmt.Sprintf("You have access to these tools. Call them by responding with JSON:\n%s", string(toolsJSON))
		ollamaMsgs = append(ollamaMsgs, ollamaMessage{Role: "system", Content: sysContent})
	}

	for _, m := range messages {
		if m.Role == RoleSystem && len(tools) > 0 {
			// Already injected tools into system; append original system prompt
			ollamaMsgs[0].Content += "\n\n" + m.Content
			continue
		}
		ollamaMsgs = append(ollamaMsgs, ollamaMessage{Role: m.Role, Content: m.Content})
	}

	req := ollamaRequest{Model: o.model, Messages: ollamaMsgs, Stream: false}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("content-type", "application/json")

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var or ollamaResponse
	if err := json.Unmarshal(respBody, &or); err != nil {
		return nil, fmt.Errorf("ollama parse: %w", err)
	}
	return &Response{Content: or.Message.Content, Done: or.Done}, nil
}
