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
		Description: "Read the contents of a file from the local filesystem.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{"type": "string", "description": "Absolute or relative file path"},
			},
			"required": []string{"path"},
		},
		Handler: func(_ context.Context, args map[string]interface{}) (string, error) {
			path := sarg(args, "path")
			path = filepath.Clean(path)
			b, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
	})

	r.Register(Tool{
		Name:        "write_file",
		Description: "Write content to a file on the local filesystem. Creates parent directories as needed.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path":    map[string]interface{}{"type": "string"},
				"content": map[string]interface{}{"type": "string", "description": "File content"},
				"append":  map[string]interface{}{"type": "boolean", "description": "Append instead of overwrite"},
			},
			"required": []string{"path", "content"},
		},
		Handler: func(_ context.Context, args map[string]interface{}) (string, error) {
			path := filepath.Clean(sarg(args, "path"))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return "", err
			}
			flags := os.O_CREATE | os.O_WRONLY
			if appendMode, _ := args["append"].(bool); appendMode {
				flags |= os.O_APPEND
			} else {
				flags |= os.O_TRUNC
			}
			f, err := os.OpenFile(path, flags, 0o644)
			if err != nil {
				return "", err
			}
			defer f.Close()
			_, err = f.WriteString(sarg(args, "content"))
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Written %d bytes to %s", len(sarg(args, "content")), path), nil
		},
	})

	r.Register(Tool{
		Name:        "list_directory",
		Description: "List files and directories at a given path.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path":      map[string]interface{}{"type": "string"},
				"recursive": map[string]interface{}{"type": "boolean", "description": "List recursively"},
			},
			"required": []string{"path"},
		},
		Handler: func(_ context.Context, args map[string]interface{}) (string, error) {
			path := filepath.Clean(sarg(args, "path"))
			recursive, _ := args["recursive"].(bool)

			var entries []string
			if recursive {
				err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if p != path {
						rel, _ := filepath.Rel(path, p)
						if info.IsDir() {
							entries = append(entries, rel+"/")
						} else {
							entries = append(entries, rel)
						}
					}
					return nil
				})
				if err != nil {
					return "", err
				}
			} else {
				infos, err := os.ReadDir(path)
				if err != nil {
					return "", err
				}
				for _, info := range infos {
					if info.IsDir() {
						entries = append(entries, info.Name()+"/")
					} else {
						entries = append(entries, info.Name())
					}
				}
			}
			return strings.Join(entries, "\n"), nil
		},
	})
}
