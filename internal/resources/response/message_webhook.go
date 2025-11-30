package response

type WebhookResponse struct {
	Message   string `json:"message"`
	MessageID string `json:"messageId"`
}
