package app

// EventType defines the type of agent event.
type EventType string

const (
	EventThinking         EventType = "thinking"
	EventToolCall         EventType = "tool_call"
	EventToolResult       EventType = "tool_result"
	EventMessage          EventType = "message"
	EventDone             EventType = "done"
	EventError            EventType = "error"
	EventApprovalRequired EventType = "approval_required"
)

// Event is emitted by the agent loop and forwarded to the frontend.
type Event struct {
	Type      EventType   `json:"type"`
	Content   string      `json:"content"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// ToolCallData is attached to EventToolCall events.
type ToolCallData struct {
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResultData is attached to EventToolResult events.
type ToolResultData struct {
	ToolName string `json:"tool_name"`
	Output   string `json:"output"`
	Success  bool   `json:"success"`
}

// ApprovalData is attached to EventApprovalRequired events.
type ApprovalData struct {
	ApprovalID  string      `json:"approval_id"`
	Action      string      `json:"action"`
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
}

// PageDiff represents a before/after diff of a Notion page.
type PageDiff struct {
	PageID string `json:"page_id"`
	Before string `json:"before"`
	After  string `json:"after"`
}
