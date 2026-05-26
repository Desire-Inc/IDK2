package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Desire-Inc/notion-agent/internal/llm"
)

type RiskLevel string

const (
	RiskSafe      RiskLevel = "safe"
	RiskMedium    RiskLevel = "medium"
	RiskDangerous RiskLevel = "dangerous"
)

type ToolMeta struct {
	Name        string
	Description string
	Risk        RiskLevel
}

func Definitions() []llm.ToolDefinition {
	return []llm.ToolDefinition{
		tool("run_command", "Run a shell command and return stdout+stderr. Prefer specialized tools when available.", RiskDangerous, map[string]any{"command": str("The command to run"), "workdir": str("Working directory (optional)")}, []string{"command"}),
		tool("read_file", "Read a local file. Use max_bytes to limit output.", RiskSafe, map[string]any{"path": str("Absolute or relative file path"), "max_bytes": num("Optional max bytes to return")}, []string{"path"}),
		tool("write_file", "Write content to a file, creating directories as needed.", RiskMedium, map[string]any{"path": str("File path to write"), "content": str("File content")}, []string{"path", "content"}),
		tool("list_dir", "List files and directories at a path.", RiskSafe, map[string]any{"path": str("Directory path")}, []string{"path"}),
		tool("search_files", "Find files by name under a directory.", RiskSafe, map[string]any{"root": str("Root directory"), "query": str("Case-insensitive filename substring"), "max_results": num("Max results")}, []string{"root", "query"}),
		tool("search_code", "Search text in code files under a directory.", RiskSafe, map[string]any{"root": str("Root directory"), "query": str("Text to search"), "max_results": num("Max matching lines")}, []string{"root", "query"}),
		tool("apply_patch", "Replace text in a file using old_string/new_string. Safer than rewriting whole files.", RiskMedium, map[string]any{"path": str("File path"), "old_string": str("Exact text to replace"), "new_string": str("Replacement text"), "replace_all": boolProp("Replace all matches")}, []string{"path", "old_string", "new_string"}),
		tool("run_tests", "Run a test/check command in a project directory.", RiskMedium, map[string]any{"command": str("Test command, e.g. go test ./... or npm test"), "workdir": str("Project directory")}, []string{"command"}),
		tool("git_status", "Show git status for a repository.", RiskSafe, map[string]any{"repo_path": str("Repository path")}, []string{"repo_path"}),
		tool("git_diff", "Show git diff for a repository.", RiskSafe, map[string]any{"repo_path": str("Repository path"), "staged": boolProp("Show staged diff")}, []string{"repo_path"}),
		tool("git_log", "Show recent git commits.", RiskSafe, map[string]any{"repo_path": str("Repository path"), "limit": num("Number of commits")}, []string{"repo_path"}),
		tool("git_branch", "Show current branch and local branches.", RiskSafe, map[string]any{"repo_path": str("Repository path")}, []string{"repo_path"}),
		tool("git_checkout", "Checkout an existing git branch.", RiskMedium, map[string]any{"repo_path": str("Repository path"), "branch": str("Branch name")}, []string{"repo_path", "branch"}),
		tool("git_create_branch", "Create and checkout a new git branch.", RiskMedium, map[string]any{"repo_path": str("Repository path"), "branch": str("New branch name")}, []string{"repo_path", "branch"}),
		tool("git_commit", "Stage files and create a git commit.", RiskMedium, map[string]any{"repo_path": str("Repository path"), "message": str("Commit message"), "all": boolProp("Stage all changes")}, []string{"repo_path", "message"}),
		tool("git_push", "Push current branch to a remote.", RiskDangerous, map[string]any{"repo_path": str("Repository path"), "remote": str("Remote name, default origin"), "branch": str("Branch name, optional")}, []string{"repo_path"}),
		tool("notion_search", "Search Notion pages/databases using the Notion API. Requires NOTION_TOKEN env var.", RiskSafe, map[string]any{"query": str("Search query")}, []string{"query"}),
		tool("notion_page_read", "Read a Notion page metadata and block children. Requires NOTION_TOKEN env var.", RiskSafe, map[string]any{"page_id": str("Notion page ID")}, []string{"page_id"}),
		tool("notion_page_create", "Create a Notion page under a parent page. Requires NOTION_TOKEN env var.", RiskMedium, map[string]any{"parent_page_id": str("Parent page ID"), "title": str("Page title"), "content": str("Plain text paragraph content")}, []string{"parent_page_id", "title"}),
		tool("notion_page_update", "Update a Notion page title or archive state. Requires NOTION_TOKEN env var.", RiskMedium, map[string]any{"page_id": str("Page ID"), "title": str("New title"), "archived": boolProp("Archive/unarchive page")}, []string{"page_id"}),
		tool("notion_db_query", "Query a Notion database. Requires NOTION_TOKEN env var.", RiskSafe, map[string]any{"database_id": str("Database ID"), "filter_json": str("Optional Notion filter JSON")}, []string{"database_id"}),
		tool("fetch_url", "Fetch a public URL and return text body. Use for documentation pages and APIs.", RiskSafe, map[string]any{"url": str("URL to fetch"), "max_bytes": num("Optional max bytes")}, []string{"url"}),
	}
}

func tool(name, desc string, risk RiskLevel, props map[string]any, required []string) llm.ToolDefinition {
	props["risk"] = map[string]any{"type": "string", "description": "Tool risk level: " + string(risk)}
	return llm.ToolDefinition{Name: name, Description: desc, Parameters: map[string]any{"type": "object", "properties": props, "required": required}}
}
func str(description string) map[string]any { return map[string]any{"type": "string", "description": description} }
func num(description string) map[string]any { return map[string]any{"type": "number", "description": description} }
func boolProp(description string) map[string]any { return map[string]any{"type": "boolean", "description": description} }

func Risk(name string, arguments string) RiskLevel {
	switch name {
	case "read_file", "list_dir", "search_files", "search_code", "git_status", "git_diff", "git_log", "git_branch", "notion_search", "notion_page_read", "notion_db_query", "fetch_url":
		return RiskSafe
	case "git_push", "run_command":
		return RiskDangerous
	default:
		return RiskMedium
	}
}

func Execute(ctx context.Context, name string, arguments string) (string, error) {
	var args map[string]any
	if strings.TrimSpace(arguments) != "" {
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return "", fmt.Errorf("invalid args JSON: %w", err)
		}
	}
	if args == nil { args = map[string]any{} }

	switch name {
	case "run_command": return runCommand(ctx, args)
	case "read_file": return readFile(args)
	case "write_file": return writeFile(args)
	case "list_dir": return listDir(args)
	case "search_files": return searchFiles(args)
	case "search_code": return searchCode(args)
	case "apply_patch": return applyPatch(args)
	case "run_tests": return runCommand(ctx, args)
	case "git_status": return git(ctx, args, "status", "--short", "--branch")
	case "git_diff":
		if getBool(args, "staged") { return git(ctx, args, "diff", "--staged") }
		return git(ctx, args, "diff")
	case "git_log": return git(ctx, args, "log", "--oneline", fmt.Sprintf("-%d", intWithDefault(args, "limit", 12)))
	case "git_branch": return git(ctx, args, "branch", "--all")
	case "git_checkout": return git(ctx, args, "checkout", getString(args, "branch"))
	case "git_create_branch": return git(ctx, args, "checkout", "-b", getString(args, "branch"))
	case "git_commit": return gitCommit(ctx, args)
	case "git_push": return gitPush(ctx, args)
	case "notion_search": return notion(ctx, "POST", "/v1/search", map[string]any{"query": getString(args, "query")})
	case "notion_page_read": return notionPageRead(ctx, getString(args, "page_id"))
	case "notion_page_create": return notionPageCreate(ctx, args)
	case "notion_page_update": return notionPageUpdate(ctx, args)
	case "notion_db_query": return notionDBQuery(ctx, args)
	case "fetch_url": return fetchURL(ctx, args)
	default: return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func runCommand(ctx context.Context, args map[string]any) (string, error) {
	cmd := getString(args, "command")
	workdir := getString(args, "workdir")
	if cmd == "" { return "", fmt.Errorf("command is required") }
	var c *exec.Cmd
	if isWindows() { c = exec.CommandContext(ctx, "cmd", "/C", cmd) } else { c = exec.CommandContext(ctx, "sh", "-c", cmd) }
	if workdir != "" { c.Dir = workdir }
	out, err := c.CombinedOutput()
	result := strings.TrimSpace(string(out))
	if err != nil { if result != "" { return result, fmt.Errorf("exit error: %w", err) }; return "", err }
	return result, nil
}

func readFile(args map[string]any) (string, error) {
	path := getString(args, "path")
	if path == "" { return "", fmt.Errorf("path is required") }
	data, err := os.ReadFile(path)
	if err != nil { return "", err }
	max := intWithDefault(args, "max_bytes", 120000)
	if max > 0 && len(data) > max { return string(data[:max]) + "\n... [truncated]", nil }
	return string(data), nil
}

func writeFile(args map[string]any) (string, error) {
	path := getString(args, "path")
	content := getString(args, "content")
	if path == "" { return "", fmt.Errorf("path is required") }
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil { return "", err }
	if err := os.WriteFile(path, []byte(content), 0644); err != nil { return "", err }
	return fmt.Sprintf("Written %d bytes to %s", len(content), path), nil
}

func listDir(args map[string]any) (string, error) {
	path := getString(args, "path")
	if path == "" { path = "." }
	entries, err := os.ReadDir(path)
	if err != nil { return "", err }
	var lines []string
	for _, e := range entries {
		if e.IsDir() { lines = append(lines, "[dir]  "+e.Name()) } else { info, _ := e.Info(); size := int64(0); if info != nil { size = info.Size() }; lines = append(lines, fmt.Sprintf("[file] %s (%d bytes)", e.Name(), size)) }
	}
	if len(lines) == 0 { return "(empty directory)", nil }
	return strings.Join(lines, "\n"), nil
}

func searchFiles(args map[string]any) (string, error) {
	root, query := getString(args, "root"), strings.ToLower(getString(args, "query"))
	if root == "" { root = "." }
	if query == "" { return "", fmt.Errorf("query is required") }
	max := intWithDefault(args, "max_results", 80)
	var results []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil { return nil }
		if shouldSkip(path, d) { if d.IsDir() { return filepath.SkipDir }; return nil }
		if strings.Contains(strings.ToLower(d.Name()), query) { results = append(results, path); if len(results) >= max { return io.EOF } }
		return nil
	})
	if err != nil && err != io.EOF { return "", err }
	if len(results) == 0 { return "no files found", nil }
	return strings.Join(results, "\n"), nil
}

func searchCode(args map[string]any) (string, error) {
	root, query := getString(args, "root"), getString(args, "query")
	if root == "" { root = "." }
	if query == "" { return "", fmt.Errorf("query is required") }
	max := intWithDefault(args, "max_results", 120)
	var results []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() { if d != nil && shouldSkip(path, d) { return filepath.SkipDir }; return nil }
		if shouldSkip(path, d) || isProbablyBinary(path) { return nil }
		data, err := os.ReadFile(path); if err != nil { return nil }
		lines := strings.Split(string(data), "\n")
		for i, line := range lines { if strings.Contains(line, query) { results = append(results, fmt.Sprintf("%s:%d: %s", path, i+1, strings.TrimSpace(line))); if len(results) >= max { return io.EOF } } }
		return nil
	})
	if err != nil && err != io.EOF { return "", err }
	if len(results) == 0 { return "no matches found", nil }
	return strings.Join(results, "\n"), nil
}

func applyPatch(args map[string]any) (string, error) {
	path, oldStr, newStr := getString(args, "path"), getString(args, "old_string"), getString(args, "new_string")
	if path == "" || oldStr == "" { return "", fmt.Errorf("path and old_string are required") }
	data, err := os.ReadFile(path); if err != nil { return "", err }
	content := string(data)
	count := strings.Count(content, oldStr)
	if count == 0 { return "", fmt.Errorf("old_string not found") }
	if count > 1 && !getBool(args, "replace_all") { return "", fmt.Errorf("old_string appears %d times; set replace_all=true or provide a more specific string", count) }
	limit := 1; if getBool(args, "replace_all") { limit = -1 }
	updated := strings.Replace(content, oldStr, newStr, limit)
	if err := os.WriteFile(path, []byte(updated), 0644); err != nil { return "", err }
	return fmt.Sprintf("Patched %s (%d replacement(s))", path, count), nil
}

func git(ctx context.Context, args map[string]any, gitArgs ...string) (string, error) {
	repo := getString(args, "repo_path"); if repo == "" { repo = "." }
	c := exec.CommandContext(ctx, "git", gitArgs...); c.Dir = repo
	out, err := c.CombinedOutput(); result := strings.TrimSpace(string(out))
	if err != nil { return result, fmt.Errorf("git %s failed: %w", strings.Join(gitArgs, " "), err) }
	return result, nil
}
func gitCommit(ctx context.Context, args map[string]any) (string, error) { if getBool(args, "all") { if out, err := git(ctx, args, "add", "-A"); err != nil { return out, err } }; return git(ctx, args, "commit", "-m", getString(args, "message")) }
func gitPush(ctx context.Context, args map[string]any) (string, error) { remote := getString(args, "remote"); if remote == "" { remote = "origin" }; branch := getString(args, "branch"); if branch == "" { return git(ctx, args, "push", remote) }; return git(ctx, args, "push", remote, branch) }

func notionPageRead(ctx context.Context, pageID string) (string, error) { if pageID == "" { return "", fmt.Errorf("page_id is required") }; page, err := notionRaw(ctx, "GET", "/v1/pages/"+pageID, nil); if err != nil { return "", err }; blocks, err := notionRaw(ctx, "GET", "/v1/blocks/"+pageID+"/children?page_size=50", nil); if err != nil { return "", err }; return truncate(page+"\n\nBLOCKS:\n"+blocks, 60000), nil }
func notionPageCreate(ctx context.Context, args map[string]any) (string, error) { body := map[string]any{"parent": map[string]any{"page_id": getString(args, "parent_page_id")}, "properties": map[string]any{"title": map[string]any{"title": []any{map[string]any{"text": map[string]any{"content": getString(args, "title")}}}}}}; if c := getString(args, "content"); c != "" { body["children"] = []any{map[string]any{"object":"block", "type":"paragraph", "paragraph": map[string]any{"rich_text": []any{map[string]any{"type":"text", "text": map[string]any{"content": c}}}}}} }; return notion(ctx, "POST", "/v1/pages", body) }
func notionPageUpdate(ctx context.Context, args map[string]any) (string, error) { body := map[string]any{}; if title := getString(args, "title"); title != "" { body["properties"] = map[string]any{"title": map[string]any{"title": []any{map[string]any{"text": map[string]any{"content": title}}}}} }; if _, ok := args["archived"]; ok { body["archived"] = getBool(args, "archived") }; return notion(ctx, "PATCH", "/v1/pages/"+getString(args, "page_id"), body) }
func notionDBQuery(ctx context.Context, args map[string]any) (string, error) { body := map[string]any{}; if f := getString(args, "filter_json"); f != "" { var filter any; if err := json.Unmarshal([]byte(f), &filter); err != nil { return "", err }; body["filter"] = filter }; return notion(ctx, "POST", "/v1/databases/"+getString(args, "database_id")+"/query", body) }
func notion(ctx context.Context, method, path string, body any) (string, error) { raw, err := notionRaw(ctx, method, path, body); if err != nil { return "", err }; return truncate(raw, 60000), nil }
func notionRaw(ctx context.Context, method, path string, body any) (string, error) { token := os.Getenv("NOTION_TOKEN"); if token == "" { return "", fmt.Errorf("NOTION_TOKEN env var is not configured") }; var r io.Reader; if body != nil { b, _ := json.Marshal(body); r = bytes.NewReader(b) }; req, err := http.NewRequestWithContext(ctx, method, "https://api.notion.com"+path, r); if err != nil { return "", err }; req.Header.Set("Authorization", "Bearer "+token); req.Header.Set("Notion-Version", "2022-06-28"); req.Header.Set("Content-Type", "application/json"); resp, err := (&http.Client{Timeout: 45*time.Second}).Do(req); if err != nil { return "", err }; defer resp.Body.Close(); data, _ := io.ReadAll(resp.Body); if resp.StatusCode >= 300 { return string(data), fmt.Errorf("notion API returned %s", resp.Status) }; return string(data), nil }
func fetchURL(ctx context.Context, args map[string]any) (string, error) { url := getString(args, "url"); if url == "" { return "", fmt.Errorf("url is required") }; req, _ := http.NewRequestWithContext(ctx, "GET", url, nil); resp, err := (&http.Client{Timeout: 30*time.Second}).Do(req); if err != nil { return "", err }; defer resp.Body.Close(); data, _ := io.ReadAll(resp.Body); return truncate(string(data), intWithDefault(args, "max_bytes", 60000)), nil }

func getString(args map[string]any, key string) string { if v, ok := args[key].(string); ok { return v }; return "" }
func getBool(args map[string]any, key string) bool { if v, ok := args[key].(bool); ok { return v }; return false }
func intWithDefault(args map[string]any, key string, def int) int { if v, ok := args[key].(float64); ok && v > 0 { return int(v) }; if v, ok := args[key].(int); ok && v > 0 { return v }; return def }
func shouldSkip(path string, d os.DirEntry) bool { name := d.Name(); return name == ".git" || name == "node_modules" || name == "dist" || name == "build" || name == ".next" || name == "vendor" }
func isProbablyBinary(path string) bool { ext := strings.ToLower(filepath.Ext(path)); return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".pdf" || ext == ".exe" || ext == ".dll" || ext == ".zip" }
func truncate(s string, max int) string { if max > 0 && len(s) > max { return s[:max] + "\n... [truncated]" }; return s }
func isWindows() bool { return os.PathSeparator == '\\' }
