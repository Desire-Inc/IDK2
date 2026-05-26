package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Desire-Inc/notion-agent/internal/llm"
)

// App is the main application struct exposed to Wails.
type App struct {
	ctx     context.Context
	mem     *Memory
	llm     llm.Provider
	cfgPath string
}

// NewApp creates a new App instance.
func NewApp() *App {
	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, ".notion-agent")
	_ = os.MkdirAll(dataDir, 0700)

	mem, err := NewMemory(filepath.Join(dataDir, "memory.db"))
	if err != nil {
		panic(err)
	}

	cfgPath := filepath.Join(dataDir, "config.json")
	cfg := loadConfig(cfgPath)
	provider := llm.NewProviderFromConfig(llm.LLMConfig{
		Provider: cfg.Provider,
		Model:    cfg.Model,
		APIKey:   cfg.APIKey,
		BaseURL:  cfg.BaseURL,
	})

	return &App{
		mem:     mem,
		llm:     provider,
		cfgPath: cfgPath,
	}
}

// Startup is called by Wails when the app starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Shutdown is called by Wails when the app closes.
func (a *App) Shutdown(ctx context.Context) {}

// NewThread creates a new conversation thread.
func (a *App) NewThread(title string) (string, error) {
	return a.mem.NewThread(title)
}

// GetThreads returns all threads ordered by most recent.
func (a *App) GetThreads() ([]Thread, error) {
	return a.mem.ListThreads()
}

// DeleteThread deletes a thread and all its messages.
func (a *App) DeleteThread(id string) error {
	return a.mem.DeleteThread(id)
}

// SendMessage runs the agent loop in the background.
func (a *App) SendMessage(threadID string, message string) error {
	go a.runLoop(a.ctx, threadID, message)
	return nil
}

// ApproveAction signals approval/denial for a pending tool action.
func (a *App) ApproveAction(approved bool) {
	select {
	case approvalCh <- approved:
	default:
	}
}

var approvalCh = make(chan bool, 1)

// LLMConfigJSON is the shape persisted to disk and sent to the frontend.
type LLMConfigJSON struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url"`
}

func loadConfig(path string) LLMConfigJSON {
	def := LLMConfigJSON{
		Provider: "openai_compatible",
		Model:    "mimo-v2.5-pro",
		BaseURL:  "https://opengateway.gitlawb.com/v1",
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return def
	}
	var cfg LLMConfigJSON
	if err := json.Unmarshal(data, &cfg); err != nil {
		return def
	}
	return cfg
}

// SaveConfig saves the LLM config to disk and reloads the provider.
func (a *App) SaveConfig(cfg LLMConfigJSON) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(a.cfgPath, data, 0600); err != nil {
		return err
	}
	a.llm = llm.NewProviderFromConfig(llm.LLMConfig{
		Provider: cfg.Provider,
		Model:    cfg.Model,
		APIKey:   cfg.APIKey,
		BaseURL:  cfg.BaseURL,
	})
	return nil
}

// GetConfig returns the current LLM config.
func (a *App) GetConfig() LLMConfigJSON {
	return loadConfig(a.cfgPath)
}
