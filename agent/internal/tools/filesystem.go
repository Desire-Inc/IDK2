package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func registerFilesystem(r *Registry) {
	r.Register(Tool{
		Name:        "read_file",
		Description: "Read the contents of a local file.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{"type": "string"},
			},
			"required": []string{"path"},
		},
		Handler: func(_ context.Context, args map[string]interface{}) (string, error) {
			b, err := os.ReadFile(filepath.Clean(sarg(args, "path")))
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
	})

	r.Register(Tool{
		Name:        "write_file",
		Description: "Write content to a local file. Creates directories as needed.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path":    map[string]interface{}{"type": "string"},
				"content": map[string]interface{}{"type": "string"},
				"append":  map[string]interface{}{"type": "boolean"},
			},
			"required": []string{"path", "content"},
		},
		Handler: func(_ context.Context, args map[string]interface{}) (string, error) {
			path := filepath.Clean(sarg(args, "path"))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return "", err
			}
			flags := os.O_CREATE | os.O_WRONLY
			if a, _ := args["append"].(bool); a {
				flags |= os.O_APPEND
			} else {
				flags |= os.O_TRUNC
			}
			f, err := os.OpenFile(path, flags, 0o644)
			if err != nil {
				return "", err
			}
			defer f.Close()
			content := sarg(args, "content")
			if _, err := f.WriteString(content); err != nil {
				return "", err
			}
			return fmt.Sprintf("Written %d bytes to %s", len(content), path), nil
		},
	})

	r.Register(Tool{
		Name:        "list_directory",
		Description: "List files and directories at a given path.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path":      map[string]interface{}{"type": "string"},
				"recursive": map[string]interface{}{"type": "boolean"},
			},
			"required": []string{"path"},
		},
		Handler: func(_ context.Context, args map[string]interface{}) (string, error) {
			path := filepath.Clean(sarg(args, "path"))
			recursive, _ := args["recursive"].(bool)
			var entries []string
			if recursive {
				_ = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
					if err != nil || p == path {
						return err
					}
					rel, _ := filepath.Rel(path, p)
					if info.IsDir() {
						rel += "/"
					}
					entries = append(entries, rel)
					return nil
				})
			} else {
				infos, err := os.ReadDir(path)
				if err != nil {
					return "", err
				}
				for _, info := range infos {
					name := info.Name()
					if info.IsDir() {
						name += "/"
					}
					entries = append(entries, name)
				}
			}
			return strings.Join(entries, "\n"), nil
		},
	})
}
