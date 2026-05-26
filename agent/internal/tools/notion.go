package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const notionAPIBase = "https://api.notion.com/v1"
const notionVersion = "2022-06-28"

func notionToken() string { return os.Getenv("NOTION_TOKEN") }

func notionRequest(ctx context.Context, method, path string, body map[string]interface{}) (map[string]interface{}, error) {
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = strings.NewReader(string(b))
	}
	req, err := http.NewRequestWithContext(ctx, method, notionAPIBase+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+notionToken())
	req.Header.Set("Notion-Version", notionVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("notion %d: %v", resp.StatusCode, result["message"])
	}
	return result, nil
}

func toJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func sarg(args map[string]interface{}, key string) string {
	v, _ := args[key].(string)
	return v
}

func registerNotion(r *Registry) {
	r.Register(Tool{
		Name: "notion_search",
		Description: "Search pages and databases in the Notion workspace.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query":  map[string]interface{}{"type": "string"},
				"filter": map[string]interface{}{"type": "string", "enum": []string{"page", "database"}},
			},
			"required": []string{"query"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			body := map[string]interface{}{"query": sarg(args, "query")}
			if f := sarg(args, "filter"); f != "" {
				body["filter"] = map[string]interface{}{"value": f, "property": "object"}
			}
			res, err := notionRequest(ctx, "POST", "/search", body)
			if err != nil {
				return "", err
			}
			return toJSON(res), nil
		},
	})

	r.Register(Tool{
		Name: "notion_page_view",
		Description: "Get a Notion page by ID, including all properties.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"page_id": map[string]interface{}{"type": "string"},
			},
			"required": []string{"page_id"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			res, err := notionRequest(ctx, "GET", "/pages/"+sarg(args, "page_id"), nil)
			if err != nil {
				return "", err
			}
			return toJSON(res), nil
		},
	})

	r.Register(Tool{
		Name: "notion_page_create",
		Description: "Create a new Notion page inside a parent page or database.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"parent_id":   map[string]interface{}{"type": "string"},
				"parent_type": map[string]interface{}{"type": "string", "enum": []string{"page_id", "database_id"}},
				"title":       map[string]interface{}{"type": "string"},
				"properties":  map[string]interface{}{"type": "object"},
				"content":     map[string]interface{}{"type": "string"},
			},
			"required": []string{"parent_id", "parent_type", "title"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			parentType := sarg(args, "parent_type")
			payload := map[string]interface{}{
				"parent": map[string]interface{}{parentType: sarg(args, "parent_id")},
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"title": []map[string]interface{}{{"text": map[string]interface{}{"content": sarg(args, "title")}}},
					},
				},
			}
			if props, ok := args["properties"].(map[string]interface{}); ok {
				for k, v := range props {
					payload["properties"].(map[string]interface{})[k] = v
				}
			}
			if content := sarg(args, "content"); content != "" {
				payload["children"] = []map[string]interface{}{
					{"object": "block", "type": "paragraph",
						"paragraph": map[string]interface{}{
							"rich_text": []map[string]interface{}{
								{"text": map[string]interface{}{"content": content}},
							},
						}},
				}
			}
			res, err := notionRequest(ctx, "POST", "/pages", payload)
			if err != nil {
				return "", err
			}
			return toJSON(res), nil
		},
	})

	r.Register(Tool{
		Name: "notion_page_delete",
		Description: "Archive (soft-delete) a Notion page.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"page_id": map[string]interface{}{"type": "string"},
			},
			"required": []string{"page_id"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			res, err := notionRequest(ctx, "PATCH", "/pages/"+sarg(args, "page_id"), map[string]interface{}{"archived": true})
			if err != nil {
				return "", err
			}
			return toJSON(res), nil
		},
	})

	r.Register(Tool{
		Name: "notion_page_set_properties",
		Description: "Update properties on an existing Notion page.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"page_id":    map[string]interface{}{"type": "string"},
				"properties": map[string]interface{}{"type": "object"},
			},
			"required": []string{"page_id", "properties"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			props, _ := args["properties"].(map[string]interface{})
			res, err := notionRequest(ctx, "PATCH", "/pages/"+sarg(args, "page_id"), map[string]interface{}{"properties": props})
			if err != nil {
				return "", err
			}
			return toJSON(res), nil
		},
	})

	r.Register(Tool{
		Name: "notion_db_list",
		Description: "Search for databases in the workspace.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{"type": "string"},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			body := map[string]interface{}{
				"filter": map[string]interface{}{"value": "database", "property": "object"},
			}
			if q := sarg(args, "query"); q != "" {
				body["query"] = q
			}
			res, err := notionRequest(ctx, "POST", "/search", body)
			if err != nil {
				return "", err
			}
			return toJSON(res), nil
		},
	})

	r.Register(Tool{
		Name: "notion_db_query",
		Description: "Query rows from a Notion database.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"database_id": map[string]interface{}{"type": "string"},
				"filter":      map[string]interface{}{"type": "object"},
				"sorts":       map[string]interface{}{"type": "array"},
				"page_size":   map[string]interface{}{"type": "integer"},
			},
			"required": []string{"database_id"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			body := map[string]interface{}{"page_size": 20}
			if f, ok := args["filter"]; ok {
				body["filter"] = f
			}
			if s, ok := args["sorts"]; ok {
				body["sorts"] = s
			}
			if ps, ok := args["page_size"].(float64); ok {
				body["page_size"] = int(ps)
			}
			res, err := notionRequest(ctx, "POST", "/databases/"+sarg(args, "database_id")+"/query", body)
			if err != nil {
				return "", err
			}
			return toJSON(res), nil
		},
	})

	r.Register(Tool{
		Name: "notion_db_create",
		Description: "Create a new Notion database inside a parent page.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"parent_page_id": map[string]interface{}{"type": "string"},
				"title":         map[string]interface{}{"type": "string"},
				"properties":    map[string]interface{}{"type": "object"},
			},
			"required": []string{"parent_page_id", "title", "properties"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			props, _ := args["properties"].(map[string]interface{})
			payload := map[string]interface{}{
				"parent":     map[string]interface{}{"type": "page_id", "page_id": sarg(args, "parent_page_id")},
				"title":      []map[string]interface{}{{"text": map[string]interface{}{"content": sarg(args, "title")}}},
				"properties": props,
			}
			res, err := notionRequest(ctx, "POST", "/databases", payload)
			if err != nil {
				return "", err
			}
			return toJSON(res), nil
		},
	})

	r.Register(Tool{
		Name: "notion_db_add_row",
		Description: "Add a new row (page) to a Notion database.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"database_id": map[string]interface{}{"type": "string"},
				"properties":  map[string]interface{}{"type": "object"},
			},
			"required": []string{"database_id", "properties"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			props, _ := args["properties"].(map[string]interface{})
			payload := map[string]interface{}{
				"parent":     map[string]interface{}{"database_id": sarg(args, "database_id")},
				"properties": props,
			}
			res, err := notionRequest(ctx, "POST", "/pages", payload)
			if err != nil {
				return "", err
			}
			return toJSON(res), nil
		},
	})

	r.Register(Tool{
		Name: "notion_block_list",
		Description: "List child blocks of a Notion page or block.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"block_id": map[string]interface{}{"type": "string"},
			},
			"required": []string{"block_id"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			res, err := notionRequest(ctx, "GET", "/blocks/"+sarg(args, "block_id")+"/children", nil)
			if err != nil {
				return "", err
			}
			return toJSON(res), nil
		},
	})

	r.Register(Tool{
		Name: "notion_block_append",
		Description: "Append blocks to a Notion page or block.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"block_id": map[string]interface{}{"type": "string"},
				"blocks":   map[string]interface{}{"type": "array"},
			},
			"required": []string{"block_id", "blocks"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			blocks, _ := args["blocks"].([]interface{})
			res, err := notionRequest(ctx, "PATCH", "/blocks/"+sarg(args, "block_id")+"/children", map[string]interface{}{"children": blocks})
			if err != nil {
				return "", err
			}
			return toJSON(res), nil
		},
	})

	r.Register(Tool{
		Name:       "notion_user_me",
		Description: "Get info about the authenticated Notion user/bot.",
		Parameters: map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		Handler: func(ctx context.Context, _ map[string]interface{}) (string, error) {
			res, err := notionRequest(ctx, "GET", "/users/me", nil)
			if err != nil {
				return "", err
			}
			return toJSON(res), nil
		},
	})

	r.Register(Tool{
		Name: "notion_comment_list",
		Description: "List comments on a Notion page.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"page_id": map[string]interface{}{"type": "string"},
			},
			"required": []string{"page_id"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			res, err := notionRequest(ctx, "GET", "/comments?block_id="+sarg(args, "page_id"), nil)
			if err != nil {
				return "", err
			}
			return toJSON(res), nil
		},
	})
}
