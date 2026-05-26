package llm

import "context"

// Provider is the interface all LLM adapters must implement.
type Provider interface {
	// Chat sends a list of messages and streams back tokens/tool calls.
	Chat(ctx context.Context, cfg Config, messages []Message, tools []ToolDefinition) (<-chan Delta, error)
	// Name returns a human-readable name for the provider.
	Name() string
	// Model returns the active model name.
	Model() string
}

// ----- Message ---------------------------------------------------------------

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role       Role
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string // only for role=tool responses
}

// ----- Tool definitions ------------------------------------------------------

type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]interface{} // JSON Schema object
}

type ToolCall struct {
	ID        string
	Name      string
	Arguments string // raw JSON
}

// ----- Streaming delta -------------------------------------------------------

type DeltaType string

const (
	DeltaText      DeltaType = "text"
	DeltaToolCall  DeltaType = "tool_call"
	DeltaThinking  DeltaType = "thinking"
	DeltaDone      DeltaType = "done"
)

type Delta struct {
	Type     DeltaType
	Text     string
	ToolCall *ToolCall
	Error    error
}

// ----- Config ----------------------------------------------------------------

type Config struct {
	Temperature float64
	MaxTokens   int
	SystemPrompt string
}

var DefaultConfig = Config{
	Temperature:  0.7,
	MaxTokens:    8192,
	SystemPrompt: "",
}
