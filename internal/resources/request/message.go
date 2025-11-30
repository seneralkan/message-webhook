package request

type ListSentMessagesRequest struct {
	Limit int `json:"limit" validate:"omitempty,gte=1,lte=1000" default:"10"`
}
