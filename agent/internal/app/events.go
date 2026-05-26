package app

import "time"

// EventType defines the type of streaming event from the agent
type EventType string

const (
	EventThinking       EventType = "thinking"
	EventToolCall       EventType = "tool_call"
	EventToolResult     EventType = "tool_result"
	EventMessage        EventType = "message"
	EventDone           EventType = "done"
	EventError          EventType = "error"
	EventApprovalNeeded EventType = "approval_required"
)

// Event is a single streaming event emitted by the agent loop
type Event struct {
	Type      EventType   `json:"type"`
	Content   string      `json:"content"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

func NewEvent(t EventType, content string, data interface{}) Event {
	return Event{
		Type:      t,
		Content:   content,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
}

// ToolCallData carries info about a tool being called
type ToolCallData struct {
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResultData carries the result of a tool call
type ToolResultData struct {
	ToolName string `json:"tool_name"`
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
	Success  bool   `json:"success"`
}

// ApprovalData is sent when the agent needs human approval before acting
type ApprovalData struct {
	ApprovalID  string      `json:"approval_id"`
	Action      string      `json:"action"`
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
}

// DiffBlock represents a single changed block in a Notion page
type DiffBlock struct {
	Type    string `json:"type"` // "added", "removed", "unchanged"
	Content string `json:"content"`
}

// PageDiff is a diff of a Notion page's blocks
type PageDiff struct {
	PageID    string      `json:"page_id"`
	PageTitle string      `json:"page_title"`
	Added     int         `json:"added"`
	Removed   int         `json:"removed"`
	Blocks    []DiffBlock `json:"blocks"`
}
