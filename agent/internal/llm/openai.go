package llm

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements Provider for OpenAI GPT models
type OpenAIProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAI(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAIProvider{client: openai.NewClient(apiKey), model: model}
}

func (o *OpenAIProvider) Name() string  { return "openai" }
func (o *OpenAIProvider) Model() string { return o.model }

func (o *OpenAIProvider) Complete(ctx context.Context, messages []Message, tools []ToolDefinition) (*Response, error) {
	oaiMessages := make([]openai.ChatCompletionMessage, 0, len(messages))
	for _, m := range messages {
		msg := openai.ChatCompletionMessage{Role: m.Role, Content: m.Content}
		if m.ToolCallID != "" {
			msg.ToolCallID = m.ToolCallID
		}
		oaiMessages = append(oaiMessages, msg)
	}

	req := openai.ChatCompletionRequest{
		Model:    o.model,
		Messages: oaiMessages,
	}

	if len(tools) > 0 {
		for _, t := range tools {
			params, _ := json.Marshal(t.Parameters)
			req.Tools = append(req.Tools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  json.RawMessage(params),
				},
			})
		}
	}

	resp, err := o.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai: empty response")
	}

	choice := resp.Choices[0]
	result := &Response{
		Content: choice.Message.Content,
		Done:    choice.FinishReason == openai.FinishReasonStop,
	}
	for _, tc := range choice.Message.ToolCalls {
		var args map[string]interface{}
		_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
		result.ToolCalls = append(result.ToolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: args,
		})
	}
	return result, nil
}
