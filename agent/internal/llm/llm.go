package llm

import "context"

// Role represents the role of a message in a conversation.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message is a single turn in the conversation.
type Message struct {
	Role       Role
	Content    string
	ToolCallID string
	ToolCalls  []ToolCall
}

// ToolCall represents a tool invocation requested by the model.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string // raw JSON
}

// ToolDefinition describes a tool the model can call.
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]any // JSON Schema object
}

// ChatRequest is the input to the Chat method.
type ChatRequest struct {
	System   string
	Messages []Message
	Tools    []ToolDefinition
}

// ChatResponse is the output from the Chat method.
type ChatResponse struct {
	Content   string
	ToolCalls []ToolCall
}

// Provider is the interface all LLM backends implement.
type Provider interface {
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}
