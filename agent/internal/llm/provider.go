package llm

import "context"

// Role constants
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
	RoleSystem    = "system"
)

// Message is a single LLM conversation message
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall is a request from the LLM to call a tool
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolDefinition describes a tool available to the LLM
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Response is the LLM's full response
type Response struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Done      bool       `json:"done"`
}

// Provider is the interface all LLM adapters must implement
type Provider interface {
	Complete(ctx context.Context, messages []Message, tools []ToolDefinition) (*Response, error)
	Name() string
	Model() string
}

// Config holds LLM provider configuration
type Config struct {
	Provider string `json:"provider"` // "anthropic", "openai", "ollama"
	Model    string `json:"model"`
	APIKey   string `json:"api_key,omitempty"`
	BaseURL  string `json:"base_url,omitempty"`
}

// DefaultConfigs holds default configurations per provider
var DefaultConfigs = map[string]Config{
	"anthropic": {Provider: "anthropic", Model: "claude-sonnet-4-5"},
	"openai":    {Provider: "openai", Model: "gpt-4o"},
	"ollama":    {Provider: "ollama", Model: "llama3.1", BaseURL: "http://localhost:11434"},
}
