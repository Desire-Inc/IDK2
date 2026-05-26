package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func gitCmd(ctx context.Context, dir string, args ...string) (string, error) {
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
				"dir": map[string]interface{}{"type": "string"},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return gitCmd(ctx, sarg(args, "dir"), "status", "--short")
		},
	})

	r.Register(Tool{
		Name:        "git_diff",
		Description: "Show git diff (staged or unstaged).",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"dir":    map[string]interface{}{"type": "string"},
				"staged": map[string]interface{}{"type": "boolean"},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			if staged, _ := args["staged"].(bool); staged {
				return gitCmd(ctx, sarg(args, "dir"), "diff", "--cached")
			}
			return gitCmd(ctx, sarg(args, "dir"), "diff")
		},
	})

	r.Register(Tool{
		Name:        "git_commit",
		Description: "Stage all changes and commit.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"dir":     map[string]interface{}{"type": "string"},
				"message": map[string]interface{}{"type": "string"},
			},
			"required": []string{"message"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			dir := sarg(args, "dir")
			if _, err := gitCmd(ctx, dir, "add", "-A"); err != nil {
				return "", err
			}
			return gitCmd(ctx, dir, "commit", "-m", sarg(args, "message"))
		},
	})

	r.Register(Tool{
		Name:        "git_push",
		Description: "Push commits to remote.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"dir":    map[string]interface{}{"type": "string"},
				"remote": map[string]interface{}{"type": "string"},
				"branch": map[string]interface{}{"type": "string"},
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
			return gitCmd(ctx, sarg(args, "dir"), "push", remote, branch)
		},
	})

	r.Register(Tool{
		Name:        "git_log",
		Description: "Show recent commit log.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"dir":   map[string]interface{}{"type": "string"},
				"limit": map[string]interface{}{"type": "integer"},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			limit := "10"
			if n, ok := args["limit"].(float64); ok {
				limit = fmt.Sprintf("%d", int(n))
			}
			return gitCmd(ctx, sarg(args, "dir"), "log", "--oneline", "-"+limit)
		},
	})
}
