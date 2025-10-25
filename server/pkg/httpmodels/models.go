package httpmodels

import (
	"errors"
	"time"
)

type AckRequest []AckRequestItem

type AckRequestItem struct {
	ID      MessageID   `json:"id"`
	Release []MessageID `json:"release,omitempty"`
}

func (items AckRequest) Validate() error {
	if len(items) == 0 {
		return errors.New("at least one message must be specified")
	}

	for _, item := range items {
		if item.ID == "" {
			return errors.New("field 'id' must not be empty")
		}

		for _, id := range item.Release {
			if id == "" {
				return errors.New("every element inside 'release' must be non-empty string")
			}
		}
	}

	return nil
}

type CheckRequest []MessageID

func (items CheckRequest) Validate() error {
	if len(items) == 0 {
		return errors.New("at least one message must be specified")
	}

	for _, id := range items {
		if id == "" {
			return errors.New("id must not be empty string")
		}
	}

	return nil
}

type CheckResponse = []Message

type ConsumeRequest struct {
	Queue QueueName `json:"queue"`
	Limit *int      `json:"limit,omitempty"`
	Poll  *int      `json:"poll,omitempty"`
}

func (r ConsumeRequest) Validate() error {
	if r.Queue == "" {
		return errors.New("field 'queue' required")
	}

	if r.Limit != nil && *r.Limit < 1 {
		return errors.New("field 'limit' must be greater than 0")
	}

	if r.Poll != nil && *r.Poll < 0 {
		return errors.New("field 'poll' must be >= 0")
	}

	return nil
}

type ConsumeResponse = []ConsumeResponseItem

type ConsumeResponseItem struct {
	ID      MessageID `json:"id"`
	Payload string    `json:"payload"`
}

type NackRequest []NackRequestItem

type NackRequestItem struct {
	ID        MessageID `json:"id"`
	Redeliver *bool     `json:"redeliver,omitempty"`
}

func (items NackRequest) Validate() error {
	if len(items) == 0 {
		return errors.New("at least one message must be specified")
	}

	for _, el := range items {
		if el.ID == "" {
			return errors.New("field 'id' must not be empty")
		}
	}

	return nil
}

type PublishRequest []PublishRequestItem

type PublishRequestItem struct {
	Queue    QueueName  `json:"queue"`
	Payload  string     `json:"payload"`
	Priority *int       `json:"priority,omitempty"`
	StartAt  *time.Time `json:"startAt,omitempty"`
}

func (items PublishRequest) Validate() error {
	if len(items) == 0 {
		return errors.New("at least one message must be specified")
	}

	for _, el := range items {
		if el.Queue == "" {
			return errors.New("field 'queue' must be non-empty string")
		}

		if el.Priority != nil && (*el.Priority < 0 || *el.Priority > 255) {
			return errors.New("priority must be between 0 and 255")
		}
	}

	return nil
}

type PublishResponse = []MessageID

type RedirectRequest []RedirectRequestItem

type RedirectRequestItem struct {
	ID          MessageID `json:"id"`
	Destination string    `json:"destination"`
}

func (items RedirectRequest) Validate() error {
	if len(items) == 0 {
		return errors.New("at least one message must be specified")
	}

	for _, el := range items {
		if el.ID == "" {
			return errors.New("field 'id' must not be empty")
		}

		if el.Destination == "" {
			return errors.New("field 'destination' must not be empty")
		}
	}

	return nil
}

type ReleaseRequest []MessageID

func (items ReleaseRequest) Validate() error {
	if len(items) == 0 {
		return errors.New("at least one message must be specified")
	}

	for _, id := range items {
		if id == "" {
			return errors.New("id must not be empty string")
		}
	}

	return nil
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type OkResponse struct {
	Ok bool `json:"ok"`
}
