package httpmodels

import "time"

type AckRequest = []AckRequestItem

type AckRequestItem struct {
	ID      MessageID   `json:"id"`
	Release []MessageID `json:"release,omitempty"`
}

type CheckRequest = []MessageID

type CheckResponse = []Message

type ConsumeRequest struct {
	Queue QueueName `json:"queue"`
	Limit *int      `json:"limit,omitempty"`
	Poll  *int      `json:"poll,omitempty"`
}

type ConsumeResponse = []ConsumeResponseItem

type ConsumeResponseItem struct {
	ID      MessageID `json:"id"`
	Payload string    `json:"payload"`
}

type NackRequest = []NackRequestItem

type NackRequestItem struct {
	ID        MessageID `json:"id"`
	Redeliver *bool     `json:"redeliver,omitempty"`
}

type PublishRequest = []PublishRequestItem

type PublishRequestItem struct {
	Queue    QueueName  `json:"queue"`
	Payload  string     `json:"payload"`
	Priority *int       `json:"priority,omitempty"`
	StartAt  *time.Time `json:"startAt,omitempty"`
}

type PublishResponse = []MessageID

type RedirectRequest = []RedirectRequestItem

type RedirectRequestItem struct {
	ID          MessageID `json:"id"`
	Destination string    `json:"destination"`
}

type ReleaseRequest = []MessageID

type ErrorResponse struct {
	Error string `json:"error"`
}

type OkResponse struct {
	Ok bool `json:"ok"`
}
