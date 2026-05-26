package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/Desire-Inc/notion-agent/internal/llm"
	"github.com/Desire-Inc/notion-agent/internal/tools"
)

// App is the main Wails application struct bound to the frontend
type App struct {
	ctx      context.Context
	memory   *Memory
	registry *tools.Registry
	mu       sync.Mutex

	// Current agent run
	cancelFn   context.CancelFunc
	approvalCh chan bool

	// LLM configuration
	llmConfig llm.Config
}

func NewApp() *App {
	return &App{
		approvalCh: make(chan bool, 1),
		llmConfig:  llm.DefaultConfigs["anthropic"],
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.registry = tools.NewRegistry()

	mem, err := NewMemory()
	if err != nil {
		runtime.LogErrorf(ctx, "Failed to init memory: %s", err)
		return
	}
	a.memory = mem

	// Load saved LLM config
	_ = a.memory.GetConfigJSON("llm_config", &a.llmConfig)
}

func (a *App) Shutdown(ctx context.Context) {
	if a.memory != nil {
		_ = a.memory.Close()
	}
	if a.cancelFn != nil {
		a.cancelFn()
	}
}

// ─── Thread Management ───────────────────────────────────────────────────────

func (a *App) ListThreads() ([]Thread, error) {
	return a.memory.ListThreads()
}

func (a *App) GetMessages(threadID string) ([]Message, error) {
	return a.memory.GetMessages(threadID)
}

func (a *App) NewThread(title string) (string, error) {
	id := uuid.New().String()
	if title == "" {
		title = "Nova conversa"
	}
	return id, a.memory.CreateThread(id, title)
}

func (a *App) DeleteThread(threadID string) error {
	_, err := a.memory.db.Exec(`DELETE FROM messages WHERE thread_id = ?`, threadID)
	if err != nil {
		return err
	}
	_, err = a.memory.db.Exec(`DELETE FROM threads WHERE id = ?`, threadID)
	return err
}

// ─── Agent Execution ─────────────────────────────────────────────────────────

// RunAgent starts the ReAct agent loop for a given thread and user message
func (a *App) RunAgent(threadID, userInput string) error {
	a.mu.Lock()
	if a.cancelFn != nil {
		a.cancelFn()
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.cancelFn = cancel
	a.mu.Unlock()

	// Persist user message
	if err := a.memory.AddMessage(threadID, Message{Role: "user", Content: userInput}); err != nil {
		return err
	}

	// Auto-title the thread on first message
	msgs, _ := a.memory.GetMessages(threadID)
	if len(msgs) == 1 {
		_ = a.memory.UpdateThreadTitle(threadID, buildSummaryTitle(userInput))
	}

	// Build LLM provider from current config
	provider, err := a.buildProvider()
	if err != nil {
		return fmt.Errorf("LLM provider: %w", err)
	}

	// Event channel — forwards events to frontend via Wails runtime.EventsEmit
	eventCh := make(chan Event, 200)
	go func() {
		for event := range eventCh {
			runtime.EventsEmit(a.ctx, "agent:event", event)
		}
	}()

	// Get history (all messages before the one we just added)
	history := msgs[:len(msgs)-1]
	if msgs == nil {
		history = nil
	}

	cfg := LoopConfig{
		MaxIterations: 30,
		LLMProvider:   provider,
		Registry:      a.registry,
		EventChan:     eventCh,
		ApprovalFn: func(data ApprovalData) bool {
			runtime.EventsEmit(a.ctx, "agent:approval_required", data)
			select {
			case approved := <-a.approvalCh:
				return approved
			case <-ctx.Done():
				return false
			}
		},
	}

	go func() {
		defer close(eventCh)
		if runErr := Run(ctx, cfg, history, userInput); runErr != nil && runErr != context.Canceled {
			runtime.EventsEmit(a.ctx, "agent:event", NewEvent(EventError, runErr.Error(), nil))
		}
		// Persist assistant response in DB
		_ = a.memory.AddMessage(threadID, Message{Role: "assistant", Content: "[agent run complete]"})
	}()

	return nil
}

// StopAgent cancels the currently running agent
func (a *App) StopAgent() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancelFn != nil {
		a.cancelFn()
		a.cancelFn = nil
	}
}

// ApproveAction responds to a pending dangerous-action approval request
func (a *App) ApproveAction(approved bool) {
	select {
	case a.approvalCh <- approved:
	default:
	}
}

// ─── LLM Configuration ───────────────────────────────────────────────────────

func (a *App) GetLLMConfig() llm.Config {
	return a.llmConfig
}

func (a *App) SetLLMConfig(config llm.Config) error {
	a.llmConfig = config
	return a.memory.SetConfigJSON("llm_config", config)
}

func (a *App) GetAvailableModels() map[string][]string {
	return map[string][]string{
		"anthropic": {"claude-sonnet-4-5", "claude-opus-4-5", "claude-haiku-4-5"},
		"openai":    {"gpt-4o", "gpt-4o-mini", "o3-mini", "gpt-4-turbo"},
		"ollama":    {"llama3.1", "llama3.2", "qwen2.5-coder", "deepseek-coder-v2", "mistral"},
	}
}

func (a *App) buildProvider() (llm.Provider, error) {
	switch a.llmConfig.Provider {
	case "anthropic":
		if a.llmConfig.APIKey == "" {
			return nil, fmt.Errorf("Anthropic API key não configurada — vá em Configurações > Modelo")
		}
		return llm.NewAnthropic(a.llmConfig.APIKey, a.llmConfig.Model), nil
	case "openai":
		if a.llmConfig.APIKey == "" {
			return nil, fmt.Errorf("OpenAI API key não configurada — vá em Configurações > Modelo")
		}
		return llm.NewOpenAI(a.llmConfig.APIKey, a.llmConfig.Model), nil
	case "ollama":
		return llm.NewOllama(a.llmConfig.BaseURL, a.llmConfig.Model), nil
	default:
		return nil, fmt.Errorf("provedor LLM desconhecido: %s", a.llmConfig.Provider)
	}
}
