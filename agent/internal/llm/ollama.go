package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultOllamaBase  = "http://localhost:11434"
const defaultOllamaModel = "llama3"

type ollamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllama(baseURL, model string) Provider {
	if baseURL == "" {
		baseURL = defaultOllamaBase
	}
	if model == "" {
		model = defaultOllamaModel
	}
	return &ollamaProvider{baseURL: baseURL, model: model, client: &http.Client{}}
}

func (o *ollamaProvider) Name() string  { return "ollama" }
func (o *ollamaProvider) Model() string { return o.model }

func (o *ollamaProvider) Chat(ctx context.Context, cfg Config, messages []Message, tools []ToolDefinition) (<-chan Delta, error) {
	// Inject tool definitions into system prompt (Ollama may not support native tool_call)
	sysPrompt := cfg.SystemPrompt
	if len(tools) > 0 {
		var sb strings.Builder
		sb.WriteString(sysPrompt)
		sb.WriteString("\n\n## Available tools\n")
		for _, t := range tools {
			paramsJSON, _ := json.Marshal(t.Parameters)
			sb.WriteString(fmt.Sprintf("### %s\n%s\nSchema: %s\n\n", t.Name, t.Description, string(paramsJSON)))
		}
		sb.WriteString("When you want to call a tool, respond ONLY with JSON:\n{\"tool\": \"name\", \"arguments\": {...}}\n")
		sysPrompt = sb.String()
	}

	ollaMsgs := make([]map[string]interface{}, 0, len(messages)+1)
	if sysPrompt != "" {
		ollaMsgs = append(ollaMsgs, map[string]interface{}{"role": "system", "content": sysPrompt})
	}
	for _, m := range messages {
		switch m.Role {
		case RoleUser:
			ollaMsgs = append(ollaMsgs, map[string]interface{}{"role": "user", "content": m.Content})
		case RoleAssistant:
			ollaMsgs = append(ollaMsgs, map[string]interface{}{"role": "assistant", "content": m.Content})
		case RoleTool:
			ollaMsgs = append(ollaMsgs, map[string]interface{}{"role": "user", "content": "Tool result: " + m.Content})
		}
	}

	payload := map[string]interface{}{
		"model":    o.model,
		"messages": ollaMsgs,
		"stream":   true,
		"options":  map[string]interface{}{"temperature": cfg.Temperature, "num_predict": cfg.MaxTokens},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("ollama: status %d: %s", resp.StatusCode, string(b))
	}

	ch := make(chan Delta, 64)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		var fullText strings.Builder

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			var evt map[string]interface{}
			if err := json.Unmarshal([]byte(line), &evt); err != nil {
				continue
			}
			if msg, ok := evt["message"].(map[string]interface{}); ok {
				token := str(msg["content"])
				fullText.WriteString(token)
				ch <- Delta{Type: DeltaText, Text: token}
			}
			if done, _ := evt["done"].(bool); done {
				// Check if full text is a tool call JSON
				trimmed := strings.TrimSpace(fullText.String())
				if strings.HasPrefix(trimmed, "{") {
					var tc struct {
						Tool      string                 `json:"tool"`
						Arguments map[string]interface{} `json:"arguments"`
					}
					if err := json.Unmarshal([]byte(trimmed), &tc); err == nil && tc.Tool != "" {
						argsJSON, _ := json.Marshal(tc.Arguments)
						ch <- Delta{Type: DeltaToolCall, ToolCall: &ToolCall{
							ID:        "ollama-1",
							Name:      tc.Tool,
							Arguments: string(argsJSON),
						}}
					}
				}
				ch <- Delta{Type: DeltaDone}
			}
		}
	}()
	return ch, nil
}
