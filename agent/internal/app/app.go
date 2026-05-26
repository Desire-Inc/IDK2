package app

import (
	"context"
	"os"
	"sync"

	"github.com/Desire-Inc/notion-agent/internal/llm"
	"github.com/Desire-Inc/notion-agent/internal/tools"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// LLMConfigJSON is the LLM configuration stored in the app.
type LLMConfigJSON struct {
	Provider string `json:"provider"` // "openai_compatible" | "anthropic" | "openai" | "ollama"
	Model    string `json:"model"`
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url"`
}

var defaultLLMConfig = LLMConfigJSON{
	Provider: "openai_compatible",
	Model:    "mimo-v2.5-pro",
	APIKey:   "",
	BaseURL:  "https://opengateway.gitlawb.com/v1",
}

// App is the main Wails application struct.
type App struct {
	ctx        context.Context
	mem        *Memory
	registry   *tools.Registry
	llmConfig  LLMConfigJSON

	mu         sync.Mutex
	cancel     context.CancelFunc
	approvalCh chan bool
}

func NewApp() *App {
	mem, err := NewMemory("notion-agent.db")
	if err != nil {
		panic(err)
	}
	return &App{
		mem:        mem,
		registry:   tools.NewRegistry(),
		llmConfig:  defaultLLMConfig,
		approvalCh: make(chan bool, 1),
	}
}

// Startup is called by Wails when the app starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	for _, env := range []string{"OPENGATEWAY_API_KEY", "ANTHROPIC_API_KEY", "OPENAI_API_KEY"} {
		if v := os.Getenv(env); v != "" {
			a.llmConfig.APIKey = v
			break
		}
	}
}

// Shutdown is called by Wails when the app closes.
func (a *App) Shutdown(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
	}
	if a.mem != nil && a.mem.db != nil {
		_ = a.mem.db.Close()
	}
}

// ---- Thread methods ---------------------------------------------------------

func (a *App) ListThreads() ([]Thread, error) {
	return a.mem.ListThreads()
}

func (a *App) NewThread(title string) (string, error) {
	if title == "" {
		title = "Nova conversa"
	}
	return a.mem.NewThread(title)
}

func (a *App) DeleteThread(threadID string) error {
	return a.mem.DeleteThread(threadID)
}

func (a *App) GetMessages(threadID string) ([]Message, error) {
	return a.mem.GetMessages(threadID)
}

// ---- Agent execution --------------------------------------------------------

func (a *App) RunAgent(threadID, input string) error {
	a.mu.Lock()
	if a.cancel != nil {
		a.cancel()
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel
	a.mu.Unlock()

	provider := a.buildProvider()

	eventCh := make(chan Event, 200)
	go func() {
		for event := range eventCh {
			runtime.EventsEmit(a.ctx, "agent:event", event)
			if event.Type == EventApprovalRequired {
				runtime.EventsEmit(a.ctx, "agent:approval_required", event.Data)
			}
		}
	}()

	history, _ := a.mem.GetMessages(threadID)

	cfg := LoopConfig{
		MaxIterations: 30,
		LLMProvider:   provider,
		Registry:      a.registry,
		EventChan:     eventCh,
		ApprovalFn: func(data ApprovalData) bool {
			select {
			case approved := <-a.approvalCh:
				return approved
			case <-ctx.Done():
				return false
			}
		},
	}

	_ = a.mem.AppendMessage(threadID, Message{Role: "user", Content: input})

	err := Run(ctx, cfg, history, input)
	close(eventCh)
	return err
}

func (a *App) StopAgent() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
}

func (a *App) ApproveAction(approved bool) {
	select {
	case a.approvalCh <- approved:
	default:
	}
}

// ---- LLM config -------------------------------------------------------------

func (a *App) GetLLMConfig() LLMConfigJSON {
	return a.llmConfig
}

func (a *App) SetLLMConfig(cfg LLMConfigJSON) {
	a.llmConfig = cfg
}

func (a *App) GetAvailableModels() map[string][]string {
	return map[string][]string{
		"openai_compatible": {"mimo-v2.5-pro", "gpt-4o", "gpt-4o-mini"},
		"anthropic":         {"claude-sonnet-4-5", "claude-opus-4-5", "claude-haiku-4-5"},
		"openai":            {"gpt-4o", "gpt-4o-mini", "o3-mini"},
		"ollama":            {"llama3", "mixtral", "codestral"},
	}
}

func (a *App) buildProvider() llm.Provider {
	switch a.llmConfig.Provider {
	case "openai_compatible":
		return llm.NewOpenAICompatible(a.llmConfig.APIKey, a.llmConfig.BaseURL, a.llmConfig.Model)
	case "openai":
		return llm.NewOpenAI(a.llmConfig.APIKey, a.llmConfig.Model)
	case "anthropic":
		return llm.NewAnthropic(a.llmConfig.APIKey, a.llmConfig.Model)
	case "ollama":
		return llm.NewOllama(a.llmConfig.BaseURL, a.llmConfig.Model)
	default:
		return llm.NewOpenAICompatible(a.llmConfig.APIKey, a.llmConfig.BaseURL, a.llmConfig.Model)
	}
}
