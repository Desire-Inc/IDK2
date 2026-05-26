package app

import (
	"encoding/json"

	"github.com/Desire-Inc/notion-agent/internal/llm"
)

func jsonMarshalToolCalls(calls []llm.ToolCall) string {
	if len(calls) == 0 { return "" }
	b, err := json.Marshal(calls)
	if err != nil { return "" }
	return string(b)
}

func jsonUnmarshalToolCalls(raw string, calls *[]llm.ToolCall) error {
	return json.Unmarshal([]byte(raw), calls)
}
