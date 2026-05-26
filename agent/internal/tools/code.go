package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Desire-Inc/notion-agent/internal/llm"
)

func (r *Registry) registerCodeTools() {
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "run_code",
			Description: "Execute code in a sandboxed environment. Supports Python, JavaScript (Node), and Bash.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"language": map[string]interface{}{
						"type":        "string",
						"description": "Programming language",
						"enum":        []string{"python", "javascript", "bash"},
					},
					"code":    map[string]interface{}{"type": "string", "description": "Code to execute"},
					"timeout": map[string]interface{}{"type": "number", "description": "Timeout in seconds (default: 30, max: 120)"},
				},
				"required": []string{"language", "code"},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			language, _ := args["language"].(string)
			code, _ := args["code"].(string)
			timeoutSecs := 30.0
			if t, ok := args["timeout"].(float64); ok && t > 0 {
				if t > 120 {
					t = 120
				}
				timeoutSecs = t
			}
			return executeCode(ctx, language, code, time.Duration(timeoutSecs)*time.Second)
		},
	})
}

func executeCode(ctx context.Context, language, code string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if isDockerAvailable() {
		return executeInDocker(ctx, language, code)
	}
	return executeLocally(ctx, language, code)
}

func executeInDocker(ctx context.Context, language, code string) (string, error) {
	var image, interpreter string
	switch language {
	case "python":
		image, interpreter = "python:3.12-slim", "python3"
	case "javascript":
		image, interpreter = "node:20-slim", "node"
	case "bash":
		image, interpreter = "bash:5", "bash"
	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}

	cmd := exec.CommandContext(ctx,
		"docker", "run", "--rm",
		"--network=none",
		"--memory=256m",
		"--cpus=0.5",
		"--security-opt=no-new-privileges",
		image, interpreter, "-c", code,
	)
	return combinedOutput(cmd)
}

func executeLocally(ctx context.Context, language, code string) (string, error) {
	var cmd *exec.Cmd
	switch language {
	case "python":
		cmd = exec.CommandContext(ctx, "python3", "-c", code)
	case "javascript":
		cmd = exec.CommandContext(ctx, "node", "-e", code)
	case "bash":
		cmd = exec.CommandContext(ctx, "bash", "-c", code)
	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}
	return combinedOutput(cmd)
}

func combinedOutput(cmd *exec.Cmd) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	result := strings.TrimSpace(stdout.String())
	if stderr.Len() > 0 {
		result += "\n[stderr]\n" + strings.TrimSpace(stderr.String())
	}
	return result, err
}

func isDockerAvailable() bool {
	return exec.Command("docker", "info").Run() == nil
}
