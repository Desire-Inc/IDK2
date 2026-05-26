package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Desire-Inc/notion-agent/internal/llm"
)

func (r *Registry) registerFilesystemTools() {
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "read_file",
			Description: "Read the contents of a local file",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "File path"},
				},
				"required": []string{"path"},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			path, _ := args["path"].(string)
			b, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("read_file %s: %w", path, err)
			}
			return string(b), nil
		},
	})

	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "write_file",
			Description: "Write content to a local file (creates or overwrites)",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":    map[string]interface{}{"type": "string", "description": "File path"},
					"content": map[string]interface{}{"type": "string", "description": "Content to write"},
				},
				"required": []string{"path", "content"},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			path, _ := args["path"].(string)
			content, _ := args["content"].(string)
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return "", err
			}
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return "", fmt.Errorf("write_file %s: %w", path, err)
			}
			return fmt.Sprintf("Arquivo escrito: %s (%d bytes)", path, len(content)), nil
		},
	})

	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "list_directory",
			Description: "List files and directories at a given path",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Directory path (default: current directory)"},
				},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			path, _ := args["path"].(string)
			if path == "" {
				path = "."
			}
			entries, err := os.ReadDir(path)
			if err != nil {
				return "", fmt.Errorf("list_directory %s: %w", path, err)
			}
			var lines []string
			for _, e := range entries {
				if e.IsDir() {
					lines = append(lines, e.Name()+"/")
				} else {
					lines = append(lines, e.Name())
				}
			}
			return strings.Join(lines, "\n"), nil
		},
	})
}
