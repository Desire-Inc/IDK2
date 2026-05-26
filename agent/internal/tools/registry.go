package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Desire-Inc/notion-agent/internal/llm"
)

// Tool is a callable unit the agent can invoke.
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
	Handler     func(ctx context.Context, args map[string]interface{}) (string, error)
}

// Registry holds all registered tools.
type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	r := &Registry{tools: map[string]Tool{}}
	registerNotion(r)
	registerCode(r)
	registerGit(r)
	registerFilesystem(r)
	return r
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Name] = t
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) Execute(ctx context.Context, name, argsJSON string) (string, error) {
	t, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("tool not found: %s", name)
	}
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("invalid args JSON: %w", err)
	}
	return t.Handler(ctx, args)
}

func (r *Registry) Definitions() []llm.ToolDefinition {
	defs := make([]llm.ToolDefinition, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, llm.ToolDefinition{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  t.Parameters,
		})
	}
	return defs
}
