package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Desire-Inc/notion-agent/internal/llm"
	"github.com/Desire-Inc/notion-agent/internal/tools"
)

// LoopConfig configures a single agent run.
type LoopConfig struct {
	MaxIterations int
	LLMProvider   llm.Provider
	Registry      *tools.Registry
	EventChan     chan<- Event
	ApprovalFn    func(ApprovalData) bool
}

const systemPrompt = `Você é um agente de IA especializado em controlar workspaces do Notion.
Seu objetivo é entender a tarefa do usuário e executá-la usando as ferramentas disponíveis.

Regras:
- Sempre explique brevemente o que vai fazer antes de chamar uma ferramenta.
- Depois de cada ferramenta, analise o resultado e decida o próximo passo.
- Se precisar de informações que não tem, use notion_search ou notion_page_view primeiro.
- Quando concluir a tarefa, envie uma mensagem final resumindo o que foi feito.
- Responda sempre em português.
`

// Run executes the ReAct loop for a single user input.
func Run(ctx context.Context, cfg LoopConfig, history []Message, userInput string) error {
	if cfg.MaxIterations == 0 {
		cfg.MaxIterations = 20
	}

	// Build LLM message history
	var msgs []llm.Message
	for _, m := range history {
		msgs = append(msgs, llm.Message{
			Role:       llm.Role(m.Role),
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		})
	}
	msgs = append(msgs, llm.Message{
		Role:    llm.RoleUser,
		Content: userInput,
	})

	cfgLLM := llm.Config{
		Temperature:  0.7,
		MaxTokens:    8192,
		SystemPrompt: systemPrompt,
	}

	emit := func(e Event) {
		e.Timestamp = time.Now().UnixMilli()
		select {
		case cfg.EventChan <- e:
		default:
		}
	}

	for i := 0; i < cfg.MaxIterations; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Call LLM
		deltaCh, err := cfg.LLMProvider.Chat(ctx, cfgLLM, msgs, cfg.Registry.Definitions())
		if err != nil {
			emit(Event{Type: EventError, Content: err.Error()})
			return err
		}

		// Collect streaming response
		var (
			textBuf    strings.Builder
			toolCalls  []llm.ToolCall
		)

		for delta := range deltaCh {
			if delta.Error != nil {
				emit(Event{Type: EventError, Content: delta.Error.Error()})
				return delta.Error
			}
			switch delta.Type {
			case llm.DeltaThinking:
				emit(Event{Type: EventThinking, Content: delta.Text})
			case llm.DeltaText:
				textBuf.WriteString(delta.Text)
			case llm.DeltaToolCall:
				if delta.ToolCall != nil {
					toolCalls = append(toolCalls, *delta.ToolCall)
				}
			}
		}

		assistantText := textBuf.String()

		// Add assistant turn to history
		assistantMsg := llm.Message{
			Role:      llm.RoleAssistant,
			Content:   assistantText,
			ToolCalls: toolCalls,
		}
		msgs = append(msgs, assistantMsg)

		// Emit text message if any
		if assistantText != "" {
			emit(Event{Type: EventMessage, Content: assistantText})
		}

		// If no tool calls, we're done
		if len(toolCalls) == 0 {
			emit(Event{Type: EventDone, Content: "Concluído"})
			return nil
		}

		// Execute each tool call
		for _, tc := range toolCalls {
			// Parse args for display
			var displayArgs map[string]interface{}
			_ = json.Unmarshal([]byte(tc.Arguments), &displayArgs)

			emit(Event{
				Type:    EventToolCall,
				Content: fmt.Sprintf("%s(%s)", tc.Name, summarizeArgs(displayArgs)),
				Data: ToolCallData{
					ToolName:  tc.Name,
					Arguments: displayArgs,
				},
			})

			// Check if tool requires approval
			if isDestructive(tc.Name) && cfg.ApprovalFn != nil {
				approvalData := ApprovalData{
					ApprovalID:  tc.ID,
					Action:      tc.Name,
					Description: fmt.Sprintf("O agente quer executar '%s' com os argumentos: %s", tc.Name, tc.Arguments),
					Data:        displayArgs,
				}
				emit(Event{Type: EventApprovalRequired, Content: tc.Name, Data: approvalData})
				if !cfg.ApprovalFn(approvalData) {
					// Denied — inject a tool result saying so
					msgs = append(msgs, llm.Message{
						Role:       llm.RoleTool,
						Content:    "Ação recusada pelo usuário.",
						ToolCallID: tc.ID,
					})
					continue
				}
			}

			// Execute
			output, execErr := cfg.Registry.Execute(ctx, tc.Name, tc.Arguments)
			success := execErr == nil
			if execErr != nil {
				output = execErr.Error()
			}

			emit(Event{
				Type:    EventToolResult,
				Content: output,
				Data: ToolResultData{
					ToolName: tc.Name,
					Output:   output,
					Success:  success,
				},
			})

			// Add tool result to history
			msgs = append(msgs, llm.Message{
				Role:       llm.RoleTool,
				Content:    output,
				ToolCallID: tc.ID,
			})
		}
	}

	emit(Event{Type: EventDone, Content: "Número máximo de iterações atingido."})
	return nil
}

// isDestructive returns true for tools that modify or delete data.
func isDestructive(toolName string) bool {
	destructive := map[string]bool{
		"notion_page_delete":          true,
		"notion_page_set_properties":  true,
		"notion_db_create":            true,
		"git_commit":                  true,
		"git_push":                    true,
		"write_file":                  true,
	}
	return destructive[toolName]
}

func summarizeArgs(args map[string]interface{}) string {
	if args == nil {
		return ""
	}
	parts := make([]string, 0, 2)
	for k, v := range args {
		s := fmt.Sprintf("%v", v)
		if len(s) > 40 {
			s = s[:40] + "..."
		}
		parts = append(parts, fmt.Sprintf("%s=%q", k, s))
		if len(parts) >= 2 {
			break
		}
	}
	return strings.Join(parts, ", ")
}
