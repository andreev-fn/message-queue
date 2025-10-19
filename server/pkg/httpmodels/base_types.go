package httpmodels

import "time"

type QueueName = string

type MessageID = string

type MessageStatus string

const (
	MsgStatusPrepared   MessageStatus = "PREPARED"
	MsgStatusAvailable  MessageStatus = "AVAILABLE"
	MsgStatusProcessing MessageStatus = "PROCESSING"
	MsgStatusDelayed    MessageStatus = "DELAYED"
	MsgStatusDelivered  MessageStatus = "DELIVERED"
	MsgStatusDropped    MessageStatus = "DROPPED"
)

type MessageChapter struct {
	Generation   int       `json:"generation"`
	Queue        QueueName `json:"queue"`
	RedirectedAt time.Time `json:"redirected_at"`
	Priority     int       `json:"priority"`
	Retries      int       `json:"retries"`
}

type Message struct {
	ID          MessageID        `json:"id"`
	Queue       QueueName        `json:"queue"`
	CreatedAt   time.Time        `json:"created_at"`
	FinalizedAt *time.Time       `json:"finalized_at"`
	Status      MessageStatus    `json:"status"`
	Priority    int              `json:"priority"`
	Retries     int              `json:"retries"`
	Generation  int              `json:"generation"`
	History     []MessageChapter `json:"history"`
	Payload     string           `json:"payload"`
}
