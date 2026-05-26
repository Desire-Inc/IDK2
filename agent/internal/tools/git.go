package tools

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

func git(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return strings.TrimRight(out.String(), "\n"), err
}

func registerGit(r *Registry) {
	r.Register(Tool{
		Name:        "git_status",
		Description: "Show git status of a repository.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"dir": map[string]interface{}{"type": "string", "description": "Repository directory"},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return git(ctx, sarg(args, "dir"), "status", "--short")
		},
	})

	r.Register(Tool{
		Name:        "git_diff",
		Description: "Show git diff for staged or unstaged changes.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"dir":    map[string]interface{}{"type": "string"},
				"staged": map[string]interface{}{"type": "boolean", "description": "Show staged diff"},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			if staged, _ := args["staged"].(bool); staged {
				return git(ctx, sarg(args, "dir"), "diff", "--cached")
			}
			return git(ctx, sarg(args, "dir"), "diff")
		},
	})

	r.Register(Tool{
		Name:        "git_commit",
		Description: "Stage all changes and create a git commit.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"dir":     map[string]interface{}{"type": "string"},
				"message": map[string]interface{}{"type": "string", "description": "Commit message"},
			},
			"required": []string{"message"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			dir := sarg(args, "dir")
			if _, err := git(ctx, dir, "add", "-A"); err != nil {
				return "", err
			}
			return git(ctx, dir, "commit", "-m", sarg(args, "message"))
		},
	})

	r.Register(Tool{
		Name:        "git_push",
		Description: "Push commits to remote.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"dir":    map[string]interface{}{"type": "string"},
				"remote": map[string]interface{}{"type": "string", "description": "Remote name (default: origin)"},
				"branch": map[string]interface{}{"type": "string", "description": "Branch name (default: HEAD)"},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			remote := sarg(args, "remote")
			if remote == "" {
				remote = "origin"
			}
			branch := sarg(args, "branch")
			if branch == "" {
				branch = "HEAD"
			}
			return git(ctx, sarg(args, "dir"), "push", remote, branch)
		},
	})

	r.Register(Tool{
		Name:        "git_log",
		Description: "Show recent git commit log.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"dir":   map[string]interface{}{"type": "string"},
				"limit": map[string]interface{}{"type": "integer", "description": "Number of commits (default 10)"},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			limit := "10"
			if n, ok := args["limit"].(float64); ok {
				limit = fmt.Sprintf("%d", int(n))
			}
			return git(ctx, sarg(args, "dir"), "log", "--oneline", "-"+limit)
		},
	})
}

func fmt_Sprintf(format string, a ...interface{}) string {
	return fmt_Sprintf(format, a...)
}
