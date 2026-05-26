package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Desire-Inc/notion-agent/internal/llm"
)

// Definitions returns the list of tools available to the LLM.
func Definitions() []llm.ToolDefinition {
	return []llm.ToolDefinition{
		{
			Name:        "run_command",
			Description: "Run a shell command and return stdout+stderr. Use for any terminal operation.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{"type": "string", "description": "The command to run"},
					"workdir": map[string]any{"type": "string", "description": "Working directory (optional)"},
				},
				"required": []string{"command"},
			},
		},
		{
			Name:        "read_file",
			Description: "Read the contents of a file.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{"type": "string", "description": "Absolute or relative file path"},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "write_file",
			Description: "Write content to a file, creating directories as needed.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":    map[string]any{"type": "string", "description": "File path to write"},
					"content": map[string]any{"type": "string", "description": "File content"},
				},
				"required": []string{"path", "content"},
			},
		},
		{
			Name:        "list_dir",
			Description: "List files and directories at a path.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{"type": "string", "description": "Directory path"},
				},
				"required": []string{"path"},
			},
		},
	}
}

// Execute runs a tool by name with the given JSON arguments string.
func Execute(ctx context.Context, name string, arguments string) (string, error) {
	var args map[string]any
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("invalid args JSON: %w", err)
	}

	switch name {
	case "run_command":
		return runCommand(ctx, args)
	case "read_file":
		return readFile(args)
	case "write_file":
		return writeFile(args)
	case "list_dir":
		return listDir(args)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func runCommand(ctx context.Context, args map[string]any) (string, error) {
	cmd, _ := args["command"].(string)
	workdir, _ := args["workdir"].(string)
	if cmd == "" {
		return "", fmt.Errorf("command is required")
	}

	var c *exec.Cmd
	if isWindows() {
		c = exec.CommandContext(ctx, "cmd", "/C", cmd)
	} else {
		c = exec.CommandContext(ctx, "sh", "-c", cmd)
	}
	if workdir != "" {
		c.Dir = workdir
	}
	out, err := c.CombinedOutput()
	result := strings.TrimSpace(string(out))
	if err != nil {
		if result != "" {
			return result, fmt.Errorf("exit error: %w", err)
		}
		return "", err
	}
	return result, nil
}

func readFile(args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func writeFile(args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	content, _ := args["content"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return fmt.Sprintf("Written %d bytes to %s", len(content), path), nil
}

func listDir(args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		path = "."
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}
	var lines []string
	for _, e := range entries {
		if e.IsDir() {
			lines = append(lines, "[dir]  "+e.Name())
		} else {
			info, _ := e.Info()
			size := int64(0)
			if info != nil {
				size = info.Size()
			}
			lines = append(lines, fmt.Sprintf("[file] %s (%d bytes)", e.Name(), size))
		}
	}
	if len(lines) == 0 {
		return "(empty directory)", nil
	}
	return strings.Join(lines, "\n"), nil
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}
