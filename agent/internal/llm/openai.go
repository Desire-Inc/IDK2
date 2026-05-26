package llm

import (
	"context"
	"encoding/json"

	openai "github.com/sashabaranov/go-openai"
)

const defaultGPTModel = "gpt-4o"

type openaiProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAI(apiKey, model string) Provider {
	if model == "" {
		model = defaultGPTModel
	}
	return &openaiProvider{client: openai.NewClient(apiKey), model: model}
}

func (o *openaiProvider) Name() string  { return "openai" }
func (o *openaiProvider) Model() string { return o.model }

func (o *openaiProvider) Chat(ctx context.Context, cfg Config, messages []Message, tools []ToolDefinition) (<-chan Delta, error) {
	req := openai.ChatCompletionRequest{
		Model:       o.model,
		MaxTokens:   cfg.MaxTokens,
		Temperature: float32(cfg.Temperature),
		Stream:      true,
		Messages:    convertMessagesOpenAI(messages, cfg.SystemPrompt),
	}
	if len(tools) > 0 {
		req.Tools = convertToolsOpenAI(tools)
	}

	stream, err := o.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	ch := make(chan Delta, 64)
	go func() {
		defer close(ch)
		defer stream.Close()

		// Accumulate tool call arguments per index
		tcArgs := map[int]string{}
		tcID   := map[int]string{}
		tcName := map[int]string{}

		for {
			resp, err := stream.Recv()
			if err != nil {
				if err.Error() != "EOF" {
					ch <- Delta{Error: err}
				}
				break
			}
			for _, choice := range resp.Choices {
				delta := choice.Delta
				if delta.Content != "" {
					ch <- Delta{Type: DeltaText, Text: delta.Content}
				}
				for _, tc := range delta.ToolCalls {
					idx := tc.Index
					if tc.ID != "" {
						tcID[*idx]   = tc.ID
						tcName[*idx] = tc.Function.Name
					}
					tcArgs[*idx] += tc.Function.Arguments
				}
				if choice.FinishReason == "tool_calls" {
					for idx := range tcID {
						ch <- Delta{Type: DeltaToolCall, ToolCall: &ToolCall{
							ID:        tcID[idx],
							Name:      tcName[idx],
							Arguments: tcArgs[idx],
						}}
					}
				}
				if choice.FinishReason == "stop" {
					ch <- Delta{Type: DeltaDone}
				}
			}
		}
	}()
	return ch, nil
}

func convertMessagesOpenAI(msgs []Message, system string) []openai.ChatCompletionMessage {
	out := []openai.ChatCompletionMessage{}
	if system != "" {
		out = append(out, openai.ChatCompletionMessage{Role: "system", Content: system})
	}
	for _, m := range msgs {
		switch m.Role {
		case RoleUser:
			out = append(out, openai.ChatCompletionMessage{Role: "user", Content: m.Content})
		case RoleAssistant:
			msg := openai.ChatCompletionMessage{Role: "assistant", Content: m.Content}
			for _, tc := range m.ToolCalls {
				msg.ToolCalls = append(msg.ToolCalls, openai.ToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: openai.FunctionCall{Name: tc.Name, Arguments: tc.Arguments},
				})
			}
			out = append(out, msg)
		case RoleTool:
			out = append(out, openai.ChatCompletionMessage{
				Role:       "tool",
				Content:    m.Content,
				ToolCallID: m.ToolCallID,
			})
		}
	}
	return out
}

func convertToolsOpenAI(tools []ToolDefinition) []openai.Tool {
	out := make([]openai.Tool, len(tools))
	for i, t := range tools {
		paramsJSON, _ := json.Marshal(t.Parameters)
		out[i] = openai.Tool{
			Type: "function",
			Function: &openai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  json.RawMessage(paramsJSON),
			},
		}
	}
	return out
}
