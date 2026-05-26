package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Desire-Inc/notion-agent/internal/llm"
	"github.com/Desire-Inc/notion-agent/internal/tools"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const systemPrompt = `You are an autonomous AI agent — a powerful coding and productivity assistant similar to Claude or Codex.

You can:
- Write, read, edit and execute code in any language
- Create, read, update and delete files on the local filesystem
- Run terminal commands and scripts
- Search the web and browse URLs
- Manage Notion workspaces: create pages, databases, query data
- Interact with Git repositories
- Plan and execute multi-step tasks autonomously

You think step by step. When given a task, you break it down, use the available tools, and complete it fully.
Always respond in the same language the user writes in.
Be concise but complete. Show your work when it's helpful.`

func (a *App) runLoop(ctx context.Context, threadID string, userMsg string) {
	emit := func(eventType, content string, data any) {
		payload := map[string]any{
			"thread_id": threadID,
			"type":      eventType,
			"content":   content,
			"data":       data,
		}
		runtime.EventsEmit(ctx, "agent:event", payload)
	}

	// Append user message to memory
	if err := a.mem.AppendMessage(threadID, Message{Role: "user", Content: userMsg}); err != nil {
		emit("error", fmt.Sprintf("memory error: %v", err), nil)
		return
	}

	for iteration := 0; iteration < 20; iteration++ {
		// Build message history
		history, err := a.mem.GetMessages(threadID)
		if err != nil {
			emit("error", fmt.Sprintf("memory error: %v", err), nil)
			return
		}

		var msgs []llm.Message
		for _, m := range history {
			var role llm.Role
			switch m.Role {
			case "user":
				role = llm.RoleUser
			case "assistant":
				role = llm.RoleAssistant
			case "tool":
				role = llm.RoleTool
			default:
				continue
			}
			msgs = append(msgs, llm.Message{Role: role, Content: m.Content, ToolCallID: m.ToolCallID})
		}

		// Call LLM
		emit("thinking", "Pensando...", nil)
		resp, err := a.llm.Chat(ctx, llm.ChatRequest{
			System:   systemPrompt,
			Messages: msgs,
			Tools:    tools.Definitions(),
		})
		if err != nil {
			emit("error", fmt.Sprintf("LLM error: %v", err), nil)
			return
		}

		// No tool calls — final answer
		if len(resp.ToolCalls) == 0 {
			if resp.Content != "" {
				_ = a.mem.AppendMessage(threadID, Message{Role: "assistant", Content: resp.Content})
				emit("message", resp.Content, nil)
			}
			emit("done", "", nil)
			return
		}

		// Save assistant turn (tool calls as JSON)
		assistantJSON, _ := json.Marshal(resp.ToolCalls)
		_ = a.mem.AppendMessage(threadID, Message{Role: "assistant", Content: string(assistantJSON)})

		// Execute each tool call
		for _, tc := range resp.ToolCalls {
			emit("tool_call", fmt.Sprintf("Executando %s", tc.Name), map[string]any{
				"tool_name": tc.Name,
				"args":      tc.Arguments,
			})

			output, execErr := tools.Execute(ctx, tc.Name, tc.Arguments)
			success := execErr == nil
			outputStr := output
			if execErr != nil {
				outputStr = execErr.Error()
			}

			// Truncate very long outputs
			if len(outputStr) > 4000 {
				outputStr = outputStr[:4000] + "\n... [truncado]"
			}

			emit("tool_result", "", map[string]any{
				"tool_name": tc.Name,
				"output":    outputStr,
				"success":   success,
			})

			_ = a.mem.AppendMessage(threadID, Message{
				Role:       "tool",
				Content:    outputStr,
				ToolCallID: tc.ID,
			})
		}
	}

	emit("done", "", nil)
}
