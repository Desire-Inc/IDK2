package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func registerCode(r *Registry) {
	r.Register(Tool{
		Name:        "run_code",
		Description: "Execute code in a sandboxed environment. Supports python, javascript (node), and bash.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"language": map[string]interface{}{"type": "string", "enum": []string{"python", "javascript", "bash"}},
				"code":     map[string]interface{}{"type": "string"},
			},
			"required": []string{"language", "code"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			lang := sarg(args, "language")
			code := sarg(args, "code")
			if dockerAvailable() {
				return runInDocker(ctx, lang, code)
			}
			return runLocal(ctx, lang, code)
		},
	})
}

func dockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func runInDocker(ctx context.Context, lang, code string) (string, error) {
	image := map[string]string{
		"python":     "python:3.12-slim",
		"javascript": "node:20-slim",
		"bash":       "alpine:latest",
	}[lang]
	if image == "" {
		return "", fmt.Errorf("unsupported language: %s", lang)
	}
	var command []string
	switch lang {
	case "python":
		command = []string{"python", "-c", code}
	case "javascript":
		command = []string{"node", "-e", code}
	case "bash":
		command = []string{"sh", "-c", code}
	}
	dockerArgs := append([]string{"run", "--rm", "--network=none", "--memory=256m", "--cpus=0.5", image}, command...)
	cmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s\nstderr: %s", err, errBuf.String())
	}
	return out.String(), nil
}

func runLocal(ctx context.Context, lang, code string) (string, error) {
	var cmd *exec.Cmd
	switch lang {
	case "python":
		cmd = exec.CommandContext(ctx, "python3", "-c", code)
	case "javascript":
		cmd = exec.CommandContext(ctx, "node", "-e", code)
	case "bash":
		cmd = exec.CommandContext(ctx, "sh", "-c", code)
	default:
		return "", fmt.Errorf("unsupported language: %s", lang)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("exit: %w\n%s", err, out.String())
	}
	return strings.TrimRight(out.String(), "\n"), nil
}
