package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Desire-Inc/notion-agent/internal/llm"
)

func (r *Registry) registerGitTools() {
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "git_status",
			Description: "Get the current git status of a repository",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Repository path (default: current dir)"},
				},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			path, _ := args["path"].(string)
			return runGit(ctx, path, "status", "--short")
		},
	})

	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "git_diff",
			Description: "Get the diff of uncommitted changes in a git repository",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{"type": "string", "description": "Repository path"},
					"file": map[string]interface{}{"type": "string", "description": "Specific file to diff"},
				},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			path, _ := args["path"].(string)
			gitArgs := []string{"diff"}
			if file, ok := args["file"].(string); ok && file != "" {
				gitArgs = append(gitArgs, "--", file)
			}
			return runGit(ctx, path, gitArgs...)
		},
	})

	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "git_log",
			Description: "Get recent commit history of a git repository",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":  map[string]interface{}{"type": "string", "description": "Repository path"},
					"limit": map[string]interface{}{"type": "number", "description": "Number of commits (default: 10)"},
				},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			path, _ := args["path"].(string)
			limit := 10
			if l, ok := args["limit"].(float64); ok && l > 0 {
				limit = int(l)
			}
			return runGit(ctx, path, "log", "--oneline", fmt.Sprintf("-%d", limit))
		},
	})

	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "git_commit",
			Description: "Stage all changes and create a git commit",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":    map[string]interface{}{"type": "string", "description": "Repository path"},
					"message": map[string]interface{}{"type": "string", "description": "Commit message"},
				},
				"required": []string{"message"},
			},
		},
		Dangerous: true,
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			path, _ := args["path"].(string)
			message, _ := args["message"].(string)
			if _, err := runGit(ctx, path, "add", "-A"); err != nil {
				return "", fmt.Errorf("git add: %w", err)
			}
			return runGit(ctx, path, "commit", "-m", message)
		},
	})

	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "git_push",
			Description: "Push commits to the remote git repository",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path":   map[string]interface{}{"type": "string", "description": "Repository path"},
					"branch": map[string]interface{}{"type": "string", "description": "Branch to push"},
				},
			},
		},
		Dangerous: true,
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			path, _ := args["path"].(string)
			gitArgs := []string{"push"}
			if branch, ok := args["branch"].(string); ok && branch != "" {
				gitArgs = append(gitArgs, "origin", branch)
			}
			return runGit(ctx, path, gitArgs...)
		},
	})
}

func runGit(ctx context.Context, repoPath string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if repoPath != "" {
		cmd.Dir = repoPath
	}
	out, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(out))
	if err != nil {
		return result, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return result, nil
}
