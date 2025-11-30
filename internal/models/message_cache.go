package models

import "time"

type SentMessageCache struct {
	MessageID         int64     `json:"message_id"`
	ExternalMessageID string    `json:"external_message_id"`
	To                string    `json:"to"`
	Content           string    `json:"content"`
	SentAt            time.Time `json:"sent_at"`
}
