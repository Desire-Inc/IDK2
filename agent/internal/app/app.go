package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/Desire-Inc/notion-agent/internal/llm"
)

// App is the main application struct exposed to Wails.
type App struct {
	ctx     context.Context
	mem     *Memory
	llm     llm.Provider
	cfgPath string
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
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
		cancels: map[string]context.CancelFunc{},
	}
}

func (a *App) Startup(ctx context.Context) { a.ctx = ctx }
func (a *App) Shutdown(ctx context.Context) { a.StopAllRuns() }

func (a *App) NewThread(title string) (string, error) { return a.mem.NewThread(title) }
func (a *App) GetThreads() ([]Thread, error) { return a.mem.ListThreads() }
func (a *App) DeleteThread(id string) error { a.StopRun(id); return a.mem.DeleteThread(id) }

// SendMessage runs the agent loop in the background.
func (a *App) SendMessage(threadID string, message string) error {
	a.StopRun(threadID)
	ctx, cancel := context.WithCancel(a.ctx)
	a.mu.Lock()
	a.cancels[threadID] = cancel
	a.mu.Unlock()
	go func() {
		defer func() {
			a.mu.Lock()
			delete(a.cancels, threadID)
			a.mu.Unlock()
		}()
		a.runLoop(ctx, threadID, message)
	}()
	return nil
}

func (a *App) StopRun(threadID string) {
	a.mu.Lock()
	cancel := a.cancels[threadID]
	delete(a.cancels, threadID)
	a.mu.Unlock()
	if cancel != nil { cancel() }
}

func (a *App) StopAllRuns() {
	a.mu.Lock()
	cancels := a.cancels
	a.cancels = map[string]context.CancelFunc{}
	a.mu.Unlock()
	for _, cancel := range cancels { cancel() }
}

// ApproveAction signals approval/denial for a pending tool action.
func (a *App) ApproveAction(approved bool) {
	select { case approvalCh <- approved: default: }
}

var approvalCh = make(chan bool, 1)

type LLMConfigJSON struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url"`
}

func loadConfig(path string) LLMConfigJSON {
	def := LLMConfigJSON{Provider: "openai_compatible", Model: "mimo-v2.5-pro", BaseURL: "https://opengateway.gitlawb.com/v1"}
	data, err := os.ReadFile(path)
	if err != nil { return def }
	var cfg LLMConfigJSON
	if err := json.Unmarshal(data, &cfg); err != nil { return def }
	if cfg.Provider == "" { cfg.Provider = def.Provider }
	if cfg.Model == "" { cfg.Model = def.Model }
	if cfg.BaseURL == "" && cfg.Provider == "openai_compatible" { cfg.BaseURL = def.BaseURL }
	return cfg
}

func (a *App) SaveConfig(cfg LLMConfigJSON) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil { return err }
	if err := os.WriteFile(a.cfgPath, data, 0600); err != nil { return err }
	a.llm = llm.NewProviderFromConfig(llm.LLMConfig{Provider: cfg.Provider, Model: cfg.Model, APIKey: cfg.APIKey, BaseURL: cfg.BaseURL})
	return nil
}

func (a *App) GetConfig() LLMConfigJSON { return loadConfig(a.cfgPath) }
