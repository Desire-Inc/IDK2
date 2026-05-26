package tools

import (
	"context"
	"fmt"

	"github.com/Desire-Inc/notion-agent/internal/llm"
)

// Tool is a callable function the agent can use
type Tool struct {
	Definition llm.ToolDefinition
	Execute    func(ctx context.Context, args map[string]interface{}) (string, error)
	Dangerous  bool // Requires human approval before execution
}

// Registry holds all available tools
type Registry struct {
	tools map[string]*Tool
}

func NewRegistry() *Registry {
	r := &Registry{tools: make(map[string]*Tool)}
	r.registerNotionTools()
	r.registerCodeTools()
	r.registerGitTools()
	r.registerFilesystemTools()
	return r
}

func (r *Registry) Register(t *Tool) {
	r.tools[t.Definition.Name] = t
}

func (r *Registry) Get(name string) (*Tool, error) {
	t, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return t, nil
}

func (r *Registry) Definitions() []llm.ToolDefinition {
	defs := make([]llm.ToolDefinition, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, t.Definition)
	}
	return defs
}

func (r *Registry) Execute(ctx context.Context, name string, args map[string]interface{}) (string, bool, error) {
	t, err := r.Get(name)
	if err != nil {
		return "", false, err
	}
	output, err := t.Execute(ctx, args)
	return output, t.Dangerous, err
}
