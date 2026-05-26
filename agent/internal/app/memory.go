package app

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// Thread represents a conversation thread.
type Thread struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Message represents a single message in a thread.
type Message struct {
	Role       string `json:"role"`        // "user" | "assistant" | "tool"
	Content    string `json:"content"`
	ToolCallID string `json:"tool_id,omitempty"`
}

// Memory manages persistent storage of threads and messages.
type Memory struct {
	db *sql.DB
}

// NewMemory opens (or creates) the SQLite database at dbPath.
func NewMemory(dbPath string) (*Memory, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}
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
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			thread_id   TEXT NOT NULL,
			role        TEXT NOT NULL,
			content     TEXT NOT NULL DEFAULT '',
			tool_call_id TEXT NOT NULL DEFAULT '',
			created_at  TEXT NOT NULL
		);
	`)
	return err
}

// NewThread creates a new thread and returns its ID.
func (m *Memory) NewThread(title string) (string, error) {
	id := uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := m.db.Exec(
		`INSERT INTO threads (id, title, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		id, title, now, now,
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

// ListThreads returns all threads, newest first.
func (m *Memory) ListThreads() ([]Thread, error) {
	rows, err := m.db.Query(`SELECT id, title, created_at, updated_at FROM threads ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var threads []Thread
	for rows.Next() {
		var t Thread
		if err := rows.Scan(&t.ID, &t.Title, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		threads = append(threads, t)
	}
	return threads, nil
}

// DeleteThread removes a thread and all its messages.
func (m *Memory) DeleteThread(id string) error {
	_, err := m.db.Exec(`DELETE FROM messages WHERE thread_id = ?`, id)
	if err != nil {
		return err
	}
	_, err = m.db.Exec(`DELETE FROM threads WHERE id = ?`, id)
	return err
}

// GetMessages returns all messages for a thread in order.
func (m *Memory) GetMessages(threadID string) ([]Message, error) {
	rows, err := m.db.Query(
		`SELECT role, content, tool_call_id FROM messages WHERE thread_id = ? ORDER BY id ASC`,
		threadID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.Role, &msg.Content, &msg.ToolCallID); err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

// AppendMessage saves a message to a thread.
func (m *Memory) AppendMessage(threadID string, msg Message) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := m.db.Exec(
		`INSERT INTO messages (thread_id, role, content, tool_call_id, created_at) VALUES (?, ?, ?, ?, ?)`,
		threadID, msg.Role, msg.Content, msg.ToolCallID, now,
	)
	if err != nil {
		return err
	}
	// Update thread timestamp
	_, _ = m.db.Exec(`UPDATE threads SET updated_at = ? WHERE id = ?`, now, threadID)
	return nil
}
