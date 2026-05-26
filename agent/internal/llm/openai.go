package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type oaiMessage struct {
	Role       string       `json:"role"`
	Content    any          `json:"content,omitempty"`
	ToolCalls  []oaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string       `json:"tool_call_id,omitempty"`
	Name       string       `json:"name,omitempty"`
}

type oaiToolCall struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Function oaiFunctionCall `json:"function"`
}

type oaiFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type oaiTool struct {
	Type     string          `json:"type"`
	Function oaiFunctionDef  `json:"function"`
}

type oaiFunctionDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type oaiRequest struct {
	Model    string       `json:"model"`
	Messages []oaiMessage `json:"messages"`
	Tools    []oaiTool    `json:"tools,omitempty"`
}

type oaiResponse struct {
	Choices []struct {
		Message oaiMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *openAICompatProvider) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	// Build messages
	var msgs []oaiMessage
	if req.System != "" {
		msgs = append(msgs, oaiMessage{Role: "system", Content: req.System})
	}
	for _, m := range req.Messages {
		msg := oaiMessage{Role: string(m.Role)}
		if m.Role == RoleTool {
			msg.Content = m.Content
			msg.ToolCallID = m.ToolCallID
		} else if len(m.ToolCalls) > 0 {
			var tcs []oaiToolCall
			for _, tc := range m.ToolCalls {
				tcs = append(tcs, oaiToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: oaiFunctionCall{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				})
			}
			msg.ToolCalls = tcs
		} else {
			msg.Content = m.Content
		}
		msgs = append(msgs, msg)
	}

	// Build tools
	var tools []oaiTool
	for _, td := range req.Tools {
		params := td.Parameters
		if params == nil {
			params = map[string]any{"type": "object", "properties": map[string]any{}}
		}
		tools = append(tools, oaiTool{
			Type: "function",
			Function: oaiFunctionDef{
				Name:        td.Name,
				Description: td.Description,
				Parameters:  params,
			},
		})
	}

	body := oaiRequest{Model: p.model, Messages: msgs, Tools: tools}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return ChatResponse{}, err
	}

	request, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return ChatResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(request)
	if err != nil {
		return ChatResponse{}, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ChatResponse{}, err
	}

	var oaiResp oaiResponse
	if err := json.Unmarshal(respBytes, &oaiResp); err != nil {
		return ChatResponse{}, fmt.Errorf("decode error: %w\nbody: %s", err, string(respBytes))
	}
	if oaiResp.Error != nil {
		return ChatResponse{}, fmt.Errorf("API error: %s", oaiResp.Error.Message)
	}
	if len(oaiResp.Choices) == 0 {
		return ChatResponse{}, fmt.Errorf("no choices in response")
	}

	choice := oaiResp.Choices[0].Message
	var tcs []ToolCall
	for _, tc := range choice.ToolCalls {
		tcs = append(tcs, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	content := ""
	if s, ok := choice.Content.(string); ok {
		content = s
	}

	return ChatResponse{Content: content, ToolCalls: tcs}, nil
}
