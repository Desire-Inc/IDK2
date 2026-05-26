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

const anthropicBaseURL  = "https://api.anthropic.com/v1/messages"
const anthropicVersion  = "2023-06-01"
const defaultClaudeModel = "claude-sonnet-4-5"

type anthropicProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewAnthropic(apiKey, model string) Provider {
	if model == "" {
		model = defaultClaudeModel
	}
	return &anthropicProvider{apiKey: apiKey, model: model, client: &http.Client{}}
}

func (a *anthropicProvider) Name() string  { return "anthropic" }
func (a *anthropicProvider) Model() string { return a.model }

func (a *anthropicProvider) Chat(
	ctx context.Context,
	cfg Config,
	messages []Message,
	tools []ToolDefinition,
) (<-chan Delta, error) {
	payload := map[string]interface{}{
		"model":      a.model,
		"max_tokens": cfg.MaxTokens,
		"stream":     true,
		"messages":   convertMessagesAnthropic(messages),
	}
	if cfg.SystemPrompt != "" {
		payload["system"] = cfg.SystemPrompt
	}
	if len(tools) > 0 {
		payload["tools"] = convertToolsAnthropic(tools)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", anthropicBaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("anthropic-beta", "interleaved-thinking-2025-05-14")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("anthropic: status %d: %s", resp.StatusCode, string(b))
	}

	ch := make(chan Delta, 64)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		var (
			currentToolID   string
			currentToolName string
			currentToolArgs string
			inTool          bool
		)

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var evt map[string]interface{}
			if err := json.Unmarshal([]byte(data), &evt); err != nil {
				continue
			}

			switch evt["type"] {
			case "content_block_start":
				if cb, ok := evt["content_block"].(map[string]interface{}); ok {
					if cb["type"] == "tool_use" {
						currentToolID   = strVal(cb["id"])
						currentToolName = strVal(cb["name"])
						currentToolArgs = ""
						inTool = true
					}
				}
			case "content_block_delta":
				if delta, ok := evt["delta"].(map[string]interface{}); ok {
					switch delta["type"] {
					case "text_delta":
						ch <- Delta{Type: DeltaText, Text: strVal(delta["text"])}
					case "thinking_delta":
						ch <- Delta{Type: DeltaThinking, Text: strVal(delta["thinking"])}
					case "input_json_delta":
						currentToolArgs += strVal(delta["partial_json"])
					}
				}
			case "content_block_stop":
				if inTool {
					ch <- Delta{Type: DeltaToolCall, ToolCall: &ToolCall{
						ID:        currentToolID,
						Name:      currentToolName,
						Arguments: currentToolArgs,
					}}
					inTool = false
				}
			case "message_stop":
				ch <- Delta{Type: DeltaDone}
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- Delta{Error: err}
		}
	}()
	return ch, nil
}

// ---- helpers ----------------------------------------------------------------

func strVal(v interface{}) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func convertMessagesAnthropic(msgs []Message) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case RoleUser:
			out = append(out, map[string]interface{}{"role": "user", "content": m.Content})
		case RoleAssistant:
			if len(m.ToolCalls) > 0 {
				content := []map[string]interface{}{}
				if m.Content != "" {
					content = append(content, map[string]interface{}{"type": "text", "text": m.Content})
				}
				for _, tc := range m.ToolCalls {
					var inputArgs map[string]interface{}
					_ = json.Unmarshal([]byte(tc.Arguments), &inputArgs)
					content = append(content, map[string]interface{}{
						"type":  "tool_use",
						"id":    tc.ID,
						"name":  tc.Name,
						"input": inputArgs,
					})
				}
				out = append(out, map[string]interface{}{"role": "assistant", "content": content})
			} else {
				out = append(out, map[string]interface{}{"role": "assistant", "content": m.Content})
			}
		case RoleTool:
			// Anthropic expects tool_result inside a user turn
			out = append(out, map[string]interface{}{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type":       "tool_result",
						"tool_use_id": m.ToolCallID,
						"content":    m.Content,
					},
				},
			})
		}
	}
	return out
}

func convertToolsAnthropic(tools []ToolDefinition) []map[string]interface{} {
	out := make([]map[string]interface{}, len(tools))
	for i, t := range tools {
		out[i] = map[string]interface{}{
			"name":         t.Name,
			"description":  t.Description,
			"input_schema": t.Parameters,
		}
	}
	return out
}
