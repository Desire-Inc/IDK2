package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Desire-Inc/notion-agent/internal/llm"
	"github.com/Desire-Inc/notion-agent/internal/tools"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const systemPrompt = `You are an autonomous AI agent — a powerful coding and productivity assistant.

You can use tools to inspect and modify local projects, run tests, work with Git, fetch URLs, and manage Notion through the Notion API when NOTION_TOKEN is configured.

Operating style:
- Start with a short plan before significant multi-step work.
- Prefer specialized tools over raw shell commands.
- Inspect before editing.
- Use apply_patch for precise edits when possible.
- Run relevant checks after code changes.
- Ask for approval only when the runtime requests it; otherwise keep moving.
- Summarize what changed, what was verified, and what remains.
- Always respond in the same language the user writes in.
Be concise but complete.`

func (a *App) runLoop(ctx context.Context, threadID string, userMsg string) {
	emit := func(eventType, content string, data any) {
		runtime.EventsEmit(ctx, "agent:event", map[string]any{"thread_id": threadID, "type": eventType, "content": content, "data": data})
	}

	if err := a.ensureThreadTitle(threadID, userMsg, emit); err != nil { emit("error", fmt.Sprintf("title error: %v", err), nil) }
	if err := a.mem.AppendMessage(threadID, Message{Role: "user", Content: userMsg}); err != nil { emit("error", fmt.Sprintf("memory error: %v", err), nil); return }

	emit("plan", "Vou entender o pedido, escolher as ferramentas necessárias, executar em etapas e validar o resultado.", map[string]any{"max_iterations": 30})

	for iteration := 0; iteration < 30; iteration++ {
		select { case <-ctx.Done(): emit("cancelled", "Execução cancelada.", nil); emit("done", "", nil); return; default: }

		history, err := a.mem.GetMessages(threadID)
		if err != nil { emit("error", fmt.Sprintf("memory error: %v", err), nil); return }

		msgs := make([]llm.Message, 0, len(history))
		for _, m := range history {
			role, ok := mapRole(m.Role)
			if !ok { continue }
			msgs = append(msgs, llm.Message{Role: role, Content: m.Content, ToolCallID: m.ToolCallID, ToolCalls: m.ToolCalls})
		}

		emit("thinking", fmt.Sprintf("Iteração %d: analisando próximo passo...", iteration+1), map[string]any{"iteration": iteration + 1})
		resp, err := a.llm.Chat(ctx, llm.ChatRequest{System: systemPrompt, Messages: msgs, Tools: tools.Definitions()})
		if err != nil { emit("error", fmt.Sprintf("LLM error: %v", err), nil); return }

		if len(resp.ToolCalls) == 0 {
			if strings.TrimSpace(resp.Content) != "" {
				_ = a.mem.AppendMessage(threadID, Message{Role: "assistant", Content: resp.Content})
				emit("message", resp.Content, nil)
			}
			emit("done", "", map[string]any{"iterations": iteration + 1})
			return
		}

		if strings.TrimSpace(resp.Content) != "" { emit("reflection", resp.Content, nil) }
		_ = a.mem.AppendMessage(threadID, Message{Role: "assistant", Content: resp.Content, ToolCalls: resp.ToolCalls})

		for _, tc := range resp.ToolCalls {
			select { case <-ctx.Done(): emit("cancelled", "Execução cancelada.", nil); emit("done", "", nil); return; default: }

			risk := tools.Risk(tc.Name, tc.Arguments)
			emit("step", fmt.Sprintf("Preparando %s", tc.Name), map[string]any{"tool_name": tc.Name, "risk": risk})
			if risk != tools.RiskSafe {
				emit("approval_required", "", map[string]any{"action": tc.Name, "description": approvalDescription(tc.Name, tc.Arguments, risk), "risk": risk, "args": tc.Arguments})
				approved, ok := waitApproval(ctx)
				if !ok { emit("error", "aprovação expirada ou execução cancelada", nil); return }
				if !approved {
					msg := "Usuário recusou a execução da ferramenta."
					emit("tool_result", "", map[string]any{"tool_name": tc.Name, "output": msg, "success": false, "risk": risk})
					_ = a.mem.AppendMessage(threadID, Message{Role: "tool", Content: msg, ToolCallID: tc.ID})
					continue
				}
			}

			emit("tool_call", fmt.Sprintf("Executando %s", tc.Name), map[string]any{"tool_name": tc.Name, "args": tc.Arguments, "risk": risk})
			output, execErr := tools.Execute(ctx, tc.Name, tc.Arguments)
			success := execErr == nil
			outputStr := output
			if execErr != nil { if outputStr != "" { outputStr += "\n" }; outputStr += execErr.Error() }
			if len(outputStr) > 8000 { outputStr = outputStr[:8000] + "\n... [truncado]" }
			emit("tool_result", "", map[string]any{"tool_name": tc.Name, "output": outputStr, "success": success, "risk": risk})
			_ = a.mem.AppendMessage(threadID, Message{Role: "tool", Content: outputStr, ToolCallID: tc.ID})
		}
	}

	emit("message", "A execução atingiu o limite de iterações. Resumi o progresso acima; peça para continuar se quiser avançar mais.", nil)
	emit("done", "", map[string]any{"max_iterations": true})
}

func mapRole(role string) (llm.Role, bool) {
	switch role { case "user": return llm.RoleUser, true; case "assistant": return llm.RoleAssistant, true; case "tool": return llm.RoleTool, true; default: return "", false }
}

func waitApproval(ctx context.Context) (bool, bool) {
	select {
	case approved := <-approvalCh: return approved, true
	case <-time.After(10 * time.Minute): return false, false
	case <-ctx.Done(): return false, false
	}
}

func approvalDescription(name, args string, risk tools.RiskLevel) string {
	return fmt.Sprintf("A ferramenta %s tem risco %s e quer executar com argumentos:\n%s", name, risk, args)
}

func (a *App) ensureThreadTitle(threadID, userMsg string, emit func(string, string, any)) error {
	thread, err := a.mem.GetThread(threadID)
	if err != nil { return err }
	if thread.Title != "" && thread.Title != "Nova conversa" { return nil }
	title := strings.TrimSpace(userMsg)
	title = strings.ReplaceAll(title, "\n", " ")
	if len(title) > 44 { title = title[:44] + "..." }
	if title == "" { title = "Nova conversa" }
	if err := a.mem.SetThreadTitle(threadID, title); err != nil { return err }
	emit("title_updated", title, map[string]any{"title": title})
	return nil
}
