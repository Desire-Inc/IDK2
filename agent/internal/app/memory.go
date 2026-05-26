package app

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/Desire-Inc/notion-agent/internal/llm"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type Thread struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Message struct {
	Role       string         `json:"role"`
	Content    string         `json:"content"`
	ToolCallID string         `json:"tool_id,omitempty"`
	ToolCalls  []llm.ToolCall `json:"tool_calls,omitempty"`
}

type Memory struct {
	db *sql.DB
	mu sync.Mutex
}

func NewMemory(dbPath string) (*Memory, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil { return nil, err }

	// SQLite allows only one writer at a time. The agent can emit messages,
	// update titles and refresh threads concurrently from UI events, so keep the
	// connection serialized and make SQLite wait briefly instead of immediately
	// returning "database is locked (SQLITE_BUSY)".
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err := db.Exec(`PRAGMA busy_timeout = 10000;`); err != nil { return nil, err }
	if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil { return nil, err }
	if _, err := db.Exec(`PRAGMA synchronous = NORMAL;`); err != nil { return nil, err }
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil { return nil, err }

	m := &Memory{db: db}
	if err := m.migrate(); err != nil { return nil, err }
	return m, nil
}

func (m *Memory) migrate() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS threads (
			id         TEXT PRIMARY KEY,
			title      TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS messages (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			thread_id    TEXT NOT NULL,
			role         TEXT NOT NULL,
			content      TEXT NOT NULL DEFAULT '',
			tool_call_id TEXT NOT NULL DEFAULT '',
			tool_calls   TEXT NOT NULL DEFAULT '',
			created_at   TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_messages_thread_id_id ON messages(thread_id, id);
	`)
	if err != nil { return err }
	_, _ = m.db.Exec(`ALTER TABLE messages ADD COLUMN tool_calls TEXT NOT NULL DEFAULT ''`)
	return nil
}

func (m *Memory) NewThread(title string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := m.db.Exec(`INSERT INTO threads (id, title, created_at, updated_at) VALUES (?, ?, ?, ?)`, id, title, now, now)
	if err != nil { return "", err }
	return id, nil
}

func (m *Memory) ListThreads() ([]Thread, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rows, err := m.db.Query(`SELECT id, title, created_at, updated_at FROM threads ORDER BY updated_at DESC`)
	if err != nil { return nil, err }
	defer rows.Close()
	var threads []Thread
	for rows.Next() {
		var t Thread
		if err := rows.Scan(&t.ID, &t.Title, &t.CreatedAt, &t.UpdatedAt); err != nil { return nil, err }
		threads = append(threads, t)
	}
	return threads, rows.Err()
}

func (m *Memory) GetThread(id string) (Thread, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var t Thread
	err := m.db.QueryRow(`SELECT id, title, created_at, updated_at FROM threads WHERE id = ?`, id).Scan(&t.ID, &t.Title, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (m *Memory) SetThreadTitle(id, title string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	title = strings.TrimSpace(title)
	if title == "" { return nil }
	if len(title) > 60 { title = title[:60] }
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := m.db.Exec(`UPDATE threads SET title = ?, updated_at = ? WHERE id = ?`, title, now, id)
	return err
}

func (m *Memory) DeleteThread(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := m.db.Exec(`DELETE FROM messages WHERE thread_id = ?`, id)
	if err != nil { return err }
	_, err = m.db.Exec(`DELETE FROM threads WHERE id = ?`, id)
	return err
}

func (m *Memory) GetMessages(threadID string) ([]Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rows, err := m.db.Query(`SELECT role, content, tool_call_id, tool_calls FROM messages WHERE thread_id = ? ORDER BY id ASC`, threadID)
	if err != nil { return nil, err }
	defer rows.Close()
	var msgs []Message
	for rows.Next() {
		var msg Message
		var toolCallsJSON string
		if err := rows.Scan(&msg.Role, &msg.Content, &msg.ToolCallID, &toolCallsJSON); err != nil { return nil, err }
		if toolCallsJSON != "" { _ = jsonUnmarshalToolCalls(toolCallsJSON, &msg.ToolCalls) }
		msgs = append(msgs, msg)
	}
	return msgs, rows.Err()
}

func (m *Memory) AppendMessage(threadID string, msg Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC().Format(time.RFC3339)
	toolCallsJSON := jsonMarshalToolCalls(msg.ToolCalls)
	_, err := m.db.Exec(`INSERT INTO messages (thread_id, role, content, tool_call_id, tool_calls, created_at) VALUES (?, ?, ?, ?, ?, ?)`, threadID, msg.Role, msg.Content, msg.ToolCallID, toolCallsJSON, now)
	if err != nil { return err }
	_, _ = m.db.Exec(`UPDATE threads SET updated_at = ? WHERE id = ?`, now, threadID)
	return nil
}
