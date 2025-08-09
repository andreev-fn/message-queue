package eventbus

import "context"

type DriverEventHandler func(channel, message string)

type PubSubDriver interface {
	Listen(ctx context.Context, channels []string, h DriverEventHandler) error
	Publish(channel string, message string) error
}
