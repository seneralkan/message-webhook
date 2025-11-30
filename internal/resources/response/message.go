package response

type SentMessagesResponse struct {
	Status    string                `json:"status"`
	Timestamp int64                 `json:"timestamp"`
	Data      []SentMessageResponse `json:"data"`
}

type SentMessageResponse struct {
	MessageID         int64  `json:"message_id"`
	ExternalMessageID string `json:"external_message_id"`
	To                string `json:"to"`
	Content           string `json:"content"`
	SentAt            string `json:"sent_at"`
}
