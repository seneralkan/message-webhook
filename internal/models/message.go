package models

import "time"

type Status string

const (
	StatusPending Status = "PENDING"
	StatusSent    Status = "SENT"
	StatusFailed  Status = "FAILED"
)

type Message struct {
	ID                int64     `json:"id"`
	To                string    `json:"to"`
	Content           string    `json:"content"`
	Status            Status    `json:"status"`
	ExternalMessageID string    `json:"external_message_id"`
	SentAt            time.Time `json:"sent_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// GetMessageSchema returns the SQL schema for creating the message table
func GetMessageSchema() string {
	return `
    CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    "to" VARCHAR(20) NOT NULL,
    content VARCHAR(160) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    external_message_id VARCHAR(64) NOT NULL,
    sent_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

    CREATE INDEX IF NOT EXISTS idx_messages_status_created_at ON messages(status, created_at);
    `
}
