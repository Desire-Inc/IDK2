package app

import (
	"database/sql"
	"strings"
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

type Memory struct { db *sql.DB }

func NewMemory(dbPath string) (*Memory, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil { return nil, err }
	if err := migrate(db); err != nil { return nil, err }
	return &Memory{db: db}, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
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
	`)
	if err != nil { return err }
	_, _ = db.Exec(`ALTER TABLE messages ADD COLUMN tool_calls TEXT NOT NULL DEFAULT ''`)
	return nil
}

func (m *Memory) NewThread(title string) (string, error) {
	id := uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := m.db.Exec(`INSERT INTO threads (id, title, created_at, updated_at) VALUES (?, ?, ?, ?)`, id, title, now, now)
	if err != nil { return "", err }
	return id, nil
}

func (m *Memory) ListThreads() ([]Thread, error) {
	rows, err := m.db.Query(`SELECT id, title, created_at, updated_at FROM threads ORDER BY updated_at DESC`)
	if err != nil { return nil, err }
	defer rows.Close()
	var threads []Thread
	for rows.Next() {
		var t Thread
		if err := rows.Scan(&t.ID, &t.Title, &t.CreatedAt, &t.UpdatedAt); err != nil { return nil, err }
		threads = append(threads, t)
	}
	return threads, nil
}

func (m *Memory) GetThread(id string) (Thread, error) {
	var t Thread
	err := m.db.QueryRow(`SELECT id, title, created_at, updated_at FROM threads WHERE id = ?`, id).Scan(&t.ID, &t.Title, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (m *Memory) SetThreadTitle(id, title string) error {
	title = strings.TrimSpace(title)
	if title == "" { return nil }
	if len(title) > 60 { title = title[:60] }
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := m.db.Exec(`UPDATE threads SET title = ?, updated_at = ? WHERE id = ?`, title, now, id)
	return err
}

func (m *Memory) DeleteThread(id string) error {
	_, err := m.db.Exec(`DELETE FROM messages WHERE thread_id = ?`, id)
	if err != nil { return err }
	_, err = m.db.Exec(`DELETE FROM threads WHERE id = ?`, id)
	return err
}

func (m *Memory) GetMessages(threadID string) ([]Message, error) {
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
	return msgs, nil
}

func (m *Memory) AppendMessage(threadID string, msg Message) error {
	now := time.Now().UTC().Format(time.RFC3339)
	toolCallsJSON := jsonMarshalToolCalls(msg.ToolCalls)
	_, err := m.db.Exec(`INSERT INTO messages (thread_id, role, content, tool_call_id, tool_calls, created_at) VALUES (?, ?, ?, ?, ?, ?)`, threadID, msg.Role, msg.Content, msg.ToolCallID, toolCallsJSON, now)
	if err != nil { return err }
	_, _ = m.db.Exec(`UPDATE threads SET updated_at = ? WHERE id = ?`, now, threadID)
	return nil
}
