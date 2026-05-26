package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/Desire-Inc/notion-agent/internal/app"
	"github.com/Desire-Inc/notion-agent/internal/llm"
	"github.com/Desire-Inc/notion-agent/internal/tools"
)

var (
	provider string
	model    string
	apiKey   string
)

var runCmd = &cobra.Command{
	Use:   "run [task]",
	Short: "Run the agent on a task",
	Example: `  notion-agent run "Crie uma database de tarefas"
  notion-agent run --provider openai --model gpt-4o "Resuma meu workspace"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var input string

		if len(args) > 0 {
			input = args[0]
		} else {
			// Interactive mode
			fmt.Print("🤖 Notion Agent > ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			input = strings.TrimSpace(scanner.Text())
		}

		if input == "" {
			return fmt.Errorf("nenhuma tarefa fornecida")
		}

		if apiKey == "" {
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
			if apiKey == "" {
				apiKey = os.Getenv("OPENAI_API_KEY")
			}
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		var p llm.Provider
		switch provider {
		case "openai":
			p = llm.NewOpenAI(apiKey, model)
		case "ollama":
			p = llm.NewOllama("", model)
		default:
			p = llm.NewAnthropic(apiKey, model)
		}

		registry := tools.NewRegistry()
		eventCh := make(chan app.Event, 200)

		// Print events to terminal
		go func() {
			for event := range eventCh {
				printEvent(event)
			}
		}()

		cfg := app.LoopConfig{
			MaxIterations: 30,
			LLMProvider:   p,
			Registry:      registry,
			EventChan:     eventCh,
			ApprovalFn: func(data app.ApprovalData) bool {
				fmt.Printf("\n⚠ı  Ação perigosa detectada:\n%s\n\nAprovar? [s/N] ", data.Description)
				var answer string
				fmt.Scanln(&answer)
				return strings.ToLower(answer) == "s" || strings.ToLower(answer) == "sim"
			},
		}

		close(eventCh) // will be replaced by loop
		eventCh = make(chan app.Event, 200)
		cfg.EventChan = eventCh

		go func() {
			for event := range eventCh {
				printEvent(event)
			}
		}()

		return app.Run(ctx, cfg, nil, input)
	},
}

func printEvent(e app.Event) {
	switch e.Type {
	case app.EventThinking:
		fmt.Printf("\r⏳ %s", e.Content)
	case app.EventToolCall:
		fmt.Printf("\n🔧 %s\n", e.Content)
	case app.EventToolResult:
		if d, ok := e.Data.(app.ToolResultData); ok {
			if d.Success {
				fmt.Printf("✅ %s\n", truncate(d.Output, 300))
			} else {
				fmt.Printf("❌ %s\n", truncate(d.Output, 300))
			}
		}
	case app.EventMessage:
		fmt.Printf("\n🤖 %s\n", e.Content)
	case app.EventDone:
		fmt.Printf("\n✅ Concluído\n")
	case app.EventError:
		fmt.Printf("\n❌ Erro: %s\n", e.Content)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&provider, "provider", "p", "anthropic", "LLM provider: anthropic, openai, ollama")
	runCmd.Flags().StringVarP(&model, "model", "m", "", "Model name (default: provider's default)")
	runCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "API key (or set ANTHROPIC_API_KEY / OPENAI_API_KEY env var)")
}
