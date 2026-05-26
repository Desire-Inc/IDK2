package app

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Message is a single message in a conversation
type Message struct {
	Role    string `json:"role"` // "user", "assistant", "tool"
	Content string `json:"content"`
	ToolID  string `json:"tool_id,omitempty"`
}

// Thread is a conversation thread
type Thread struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Memory manages short-term and long-term memory using SQLite
type Memory struct {
	db *sql.DB
}

func NewMemory() (*Memory, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", filepath.Join(dir, "memory.db"))
	if err != nil {
		return nil, err
	}

	m := &Memory{db: db}
	if err := m.migrate(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Memory) migrate() error {
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS threads (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			thread_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			tool_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (thread_id) REFERENCES threads(id)
		);
		CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
	`)
	return err
}

func (m *Memory) CreateThread(id, title string) error {
	_, err := m.db.Exec(`INSERT INTO threads (id, title) VALUES (?, ?)`, id, title)
	return err
}

func (m *Memory) UpdateThreadTitle(id, title string) error {
	_, err := m.db.Exec(`UPDATE threads SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, title, id)
	return err
}

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

func (m *Memory) AddMessage(threadID string, msg Message) error {
	_, err := m.db.Exec(
		`INSERT INTO messages (thread_id, role, content, tool_id) VALUES (?, ?, ?, ?)`,
		threadID, msg.Role, msg.Content, nullableString(msg.ToolID),
	)
	if err != nil {
		return err
	}
	_, err = m.db.Exec(`UPDATE threads SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`, threadID)
	return err
}

func (m *Memory) GetMessages(threadID string) ([]Message, error) {
	rows, err := m.db.Query(
		`SELECT role, content, COALESCE(tool_id, '') FROM messages WHERE thread_id = ? ORDER BY created_at ASC`,
		threadID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.Role, &msg.Content, &msg.ToolID); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (m *Memory) SetConfig(key, value string) error {
	_, err := m.db.Exec(`INSERT OR REPLACE INTO config (key, value) VALUES (?, ?)`, key, value)
	return err
}

func (m *Memory) GetConfig(key string) (string, error) {
	var value string
	err := m.db.QueryRow(`SELECT value FROM config WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (m *Memory) SetConfigJSON(key string, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return m.SetConfig(key, string(b))
}

func (m *Memory) GetConfigJSON(key string, v interface{}) error {
	s, err := m.GetConfig(key)
	if err != nil || s == "" {
		return err
	}
	return json.Unmarshal([]byte(s), v)
}

func (m *Memory) Close() error {
	return m.db.Close()
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "notion-agent"), nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
