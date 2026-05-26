package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Desire-Inc/notion-agent/internal/llm"
	"github.com/Desire-Inc/notion-agent/internal/tools"
)

const systemPrompt = `Você é o Notion Agent — um assistente de IA com controle total sobre um workspace do Notion.
Você pode ler páginas, criar databases, adicionar conteúdo, executar código e gerenciar repositórios git.

Quando receber uma tarefa:
1. Divida em passos concretos
2. Use as ferramentas disponíveis para executar cada passo
3. Observe os resultados e ajuste sua abordagem se necessário
4. Relate claramente quando concluir

Sempre prefira usar ferramentas em vez de apenas explicar o que faria.
Para ações destrutivas (deletar páginas, commitar no git), o sistema pedirá aprovação automaticamente.
Seja conciso — o usuário já vê cada chamada de ferramenta e resultado em tempo real.`

// LoopConfig configures the ReAct loop
type LoopConfig struct {
	MaxIterations int
	LLMProvider   llm.Provider
	Registry      *tools.Registry
	EventChan     chan<- Event
	ApprovalFn    func(ApprovalData) bool
}

// Run executes the ReAct loop for a given task
func Run(ctx context.Context, cfg LoopConfig, history []Message, userInput string) error {
	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 30
	}

	// Build LLM message history
	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: systemPrompt},
	}
	for _, m := range history {
		messages = append(messages, llm.Message{Role: m.Role, Content: m.Content})
	}
	messages = append(messages, llm.Message{Role: llm.RoleUser, Content: userInput})

	toolDefs := cfg.Registry.Definitions()

	for i := 0; i < cfg.MaxIterations; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		cfg.EventChan <- NewEvent(EventThinking, fmt.Sprintf("Pensando... (passo %d)", i+1), nil)

		resp, err := cfg.LLMProvider.Complete(ctx, messages, toolDefs)
		if err != nil {
			cfg.EventChan <- NewEvent(EventError, fmt.Sprintf("Erro do LLM: %s", err), nil)
			return err
		}

		if resp.Content != "" {
			cfg.EventChan <- NewEvent(EventMessage, resp.Content, nil)
		}

		// No tool calls means the agent is done
		if len(resp.ToolCalls) == 0 {
			cfg.EventChan <- NewEvent(EventDone, "Concluído.", nil)
			return nil
		}

		// Add assistant turn to message history
		assistantMsg, _ := json.Marshal(map[string]interface{}{
			"content":    resp.Content,
			"tool_calls": resp.ToolCalls,
		})
		messages = append(messages, llm.Message{
			Role:    llm.RoleAssistant,
			Content: string(assistantMsg),
		})

		// Execute each tool call
		for _, tc := range resp.ToolCalls {
			cfg.EventChan <- NewEvent(EventToolCall, fmt.Sprintf("→ %s", tc.Name), ToolCallData{
				ToolName:  tc.Name,
				Arguments: tc.Arguments,
			})

			// Check if tool requires approval
			tool, err := cfg.Registry.Get(tc.Name)
			if err != nil {
				output := fmt.Sprintf("Erro: ferramenta '%s' não encontrada", tc.Name)
				cfg.EventChan <- NewEvent(EventToolResult, output, ToolResultData{ToolName: tc.Name, Output: output, Success: false})
				continue
			}

			if tool.Dangerous && cfg.ApprovalFn != nil {
				argsDesc, _ := json.MarshalIndent(tc.Arguments, "", "  ")
				approval := ApprovalData{
					ApprovalID:  tc.ID,
					Action:      tc.Name,
					Description: fmt.Sprintf("Executar `%s`:\n%s", tc.Name, string(argsDesc)),
					Data:        tc.Arguments,
				}
				cfg.EventChan <- NewEvent(EventApprovalNeeded, "Aprovação necessária antes de continuar", approval)
				if !cfg.ApprovalFn(approval) {
					output := "Ação cancelada pelo usuário."
					cfg.EventChan <- NewEvent(EventToolResult, output, ToolResultData{ToolName: tc.Name, Output: output, Success: false})
					messages = append(messages, llm.Message{Role: "tool", Content: output, ToolCallID: tc.ID})
					continue
				}
			}

			output, _, execErr := cfg.Registry.Execute(ctx, tc.Name, tc.Arguments)
			success := execErr == nil
			if execErr != nil {
				output = fmt.Sprintf("Erro: %s", execErr)
			}

			// Truncate large outputs to keep context window manageable
			if len(output) > 4000 {
				output = output[:4000] + "\n... (output truncado — use filtros para refinar)"
			}

			cfg.EventChan <- NewEvent(EventToolResult, output, ToolResultData{
				ToolName: tc.Name,
				Output:   output,
				Success:  success,
			})

			messages = append(messages, llm.Message{
				Role:       "tool",
				Content:    output,
				ToolCallID: tc.ID,
			})
		}
	}

	cfg.EventChan <- NewEvent(EventDone, "Limite de iterações atingido.", nil)
	return nil
}

// buildSummaryTitle creates a short title for a thread based on the user input
func buildSummaryTitle(input string) string {
	words := strings.Fields(input)
	if len(words) > 6 {
		words = words[:6]
	}
	return strings.Join(words, " ")
}
