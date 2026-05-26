package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Desire-Inc/notion-agent/internal/llm"
)

func (r *Registry) registerNotionTools() {
	// Search
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "notion_search",
			Description: "Search pages and databases in the Notion workspace",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{"type": "string", "description": "Search query"},
					"type":  map[string]interface{}{"type": "string", "description": "Filter: 'page' or 'database'", "enum": []string{"page", "database"}},
				},
				"required": []string{"query"},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			query, _ := args["query"].(string)
			cmdArgs := []string{"search", query, "--format", "json"}
			if t, ok := args["type"].(string); ok && t != "" {
				cmdArgs = append(cmdArgs, "--type", t)
			}
			return runNotion(ctx, cmdArgs...)
		},
	})

	// Page view
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "notion_page_view",
			Description: "View the content and properties of a Notion page",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id": map[string]interface{}{"type": "string", "description": "Page ID or Notion URL"},
				},
				"required": []string{"page_id"},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			pageID, _ := args["page_id"].(string)
			return runNotion(ctx, "page", "view", pageID, "--format", "json")
		},
	})

	// Page create
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "notion_page_create",
			Description: "Create a new Notion page",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"parent_id": map[string]interface{}{"type": "string", "description": "Parent page or database ID"},
					"title":     map[string]interface{}{"type": "string", "description": "Page title"},
					"content":   map[string]interface{}{"type": "string", "description": "Page content in Markdown"},
				},
				"required": []string{"parent_id", "title"},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			parentID, _ := args["parent_id"].(string)
			title, _ := args["title"].(string)
			cmdArgs := []string{"page", "create", parentID, "--title", title, "--format", "json"}
			if content, ok := args["content"].(string); ok && content != "" {
				cmdArgs = append(cmdArgs, "--body", content)
			}
			return runNotion(ctx, cmdArgs...)
		},
	})

	// Page delete (dangerous)
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "notion_page_delete",
			Description: "Archive (soft-delete) a Notion page",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id": map[string]interface{}{"type": "string", "description": "Page ID or URL"},
				},
				"required": []string{"page_id"},
			},
		},
		Dangerous: true,
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			pageID, _ := args["page_id"].(string)
			return runNotion(ctx, "page", "delete", pageID, "--yes")
		},
	})

	// Page set properties
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "notion_page_set",
			Description: "Update properties of an existing Notion page",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id":    map[string]interface{}{"type": "string", "description": "Page ID or URL"},
					"properties": map[string]interface{}{"type": "object", "description": "Key=value pairs to update"},
				},
				"required": []string{"page_id", "properties"},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			pageID, _ := args["page_id"].(string)
			cmdArgs := []string{"page", "set", pageID}
			if props, ok := args["properties"].(map[string]interface{}); ok {
				for k, v := range props {
					cmdArgs = append(cmdArgs, fmt.Sprintf("%s=%v", k, v))
				}
			}
			cmdArgs = append(cmdArgs, "--format", "json")
			return runNotion(ctx, cmdArgs...)
		},
	})

	// DB list
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "notion_db_list",
			Description: "List all accessible databases in the Notion workspace",
			Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return runNotion(ctx, "db", "list", "--format", "json")
		},
	})

	// DB query
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "notion_db_query",
			Description: "Query a Notion database with optional filters and sorting",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"db_id":  map[string]interface{}{"type": "string", "description": "Database ID or URL"},
					"filter": map[string]interface{}{"type": "string", "description": "Filter, e.g. 'Status=Done'"},
					"sort":   map[string]interface{}{"type": "string", "description": "Sort, e.g. 'Date:desc'"},
					"limit":  map[string]interface{}{"type": "number", "description": "Max results"},
				},
				"required": []string{"db_id"},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			dbID, _ := args["db_id"].(string)
			cmdArgs := []string{"db", "query", dbID, "--format", "json"}
			if filter, ok := args["filter"].(string); ok && filter != "" {
				cmdArgs = append(cmdArgs, "--filter", filter)
			}
			if sort, ok := args["sort"].(string); ok && sort != "" {
				cmdArgs = append(cmdArgs, "--sort", sort)
			}
			if limit, ok := args["limit"].(float64); ok && limit > 0 {
				cmdArgs = append(cmdArgs, "--limit", fmt.Sprintf("%d", int(limit)))
			}
			return runNotion(ctx, cmdArgs...)
		},
	})

	// DB create
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "notion_db_create",
			Description: "Create a new Notion database inside a page",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"parent_id": map[string]interface{}{"type": "string", "description": "Parent page ID"},
					"title":     map[string]interface{}{"type": "string", "description": "Database title"},
				},
				"required": []string{"parent_id", "title"},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			parentID, _ := args["parent_id"].(string)
			title, _ := args["title"].(string)
			return runNotion(ctx, "db", "create", parentID, "--title", title, "--format", "json")
		},
	})

	// DB add row
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "notion_db_add",
			Description: "Add a row (page) to a Notion database with properties",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"db_id":      map[string]interface{}{"type": "string", "description": "Database ID or URL"},
					"properties": map[string]interface{}{"type": "object", "description": "Row properties, e.g. {\"Name\": \"Task 1\", \"Status\": \"Todo\"}"},
				},
				"required": []string{"db_id", "properties"},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			dbID, _ := args["db_id"].(string)
			cmdArgs := []string{"db", "add", dbID, "--format", "json"}
			if props, ok := args["properties"].(map[string]interface{}); ok {
				for k, v := range props {
					cmdArgs = append(cmdArgs, fmt.Sprintf("%s=%v", k, v))
				}
			}
			return runNotion(ctx, cmdArgs...)
		},
	})

	// Block list
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "notion_block_list",
			Description: "List content blocks of a Notion page (optionally as Markdown)",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id":  map[string]interface{}{"type": "string", "description": "Page ID or URL"},
					"depth":    map[string]interface{}{"type": "number", "description": "Block recursion depth (1-10)"},
					"markdown": map[string]interface{}{"type": "boolean", "description": "Return as Markdown"},
				},
				"required": []string{"page_id"},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			pageID, _ := args["page_id"].(string)
			cmdArgs := []string{"block", "list", pageID}
			if depth, ok := args["depth"].(float64); ok {
				cmdArgs = append(cmdArgs, "--depth", fmt.Sprintf("%d", int(depth)))
			}
			if md, ok := args["markdown"].(bool); ok && md {
				cmdArgs = append(cmdArgs, "--md")
			} else {
				cmdArgs = append(cmdArgs, "--format", "json")
			}
			return runNotion(ctx, cmdArgs...)
		},
	})

	// Block append
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "notion_block_append",
			Description: "Append Markdown content to a Notion page",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id": map[string]interface{}{"type": "string", "description": "Page ID or URL"},
					"content": map[string]interface{}{"type": "string", "description": "Markdown content to append"},
				},
				"required": []string{"page_id", "content"},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			pageID, _ := args["page_id"].(string)
			content, _ := args["content"].(string)
			cmd := exec.CommandContext(ctx, "notion", "block", "append", pageID, "--format", "json")
			cmd.Stdin = strings.NewReader(content)
			out, err := cmd.CombinedOutput()
			return strings.TrimSpace(string(out)), err
		},
	})

	// User me
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "notion_user_me",
			Description: "Get info about the currently authenticated Notion user",
			Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return runNotion(ctx, "user", "me", "--format", "json")
		},
	})

	// Comment list
	r.Register(&Tool{
		Definition: llm.ToolDefinition{
			Name:        "notion_comment_list",
			Description: "List comments on a Notion page",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"page_id": map[string]interface{}{"type": "string", "description": "Page ID or URL"},
				},
				"required": []string{"page_id"},
			},
		},
		Execute: func(ctx context.Context, args map[string]interface{}) (string, error) {
			pageID, _ := args["page_id"].(string)
			return runNotion(ctx, "comment", "list", pageID, "--format", "json")
		},
	})
}

// runNotion executes a notion CLI command and returns its output
func runNotion(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "notion", args...)
	out, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(out))
	if err != nil {
		return result, fmt.Errorf("notion %s: %w | output: %s", strings.Join(args, " "), err, result)
	}
	return result, nil
}
