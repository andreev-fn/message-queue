package fixtures

const DefaultMsgQueue = "test"
const DefaultMsgPayload = `{"arg": 123}`
const DefaultMsgPriority = 100

type options struct {
	queue    string
	payload  string
	priority int
	history  []string
}

type Option func(*options)

func WithQueue(queue string) Option {
	return func(o *options) {
		o.queue = queue
	}
}

func WithPayload(payload string) Option {
	return func(o *options) {
		o.payload = payload
	}
}

func WithPriority(priority int) Option {
	return func(o *options) {
		o.priority = priority
	}
}

func WithHistory(queues ...string) Option {
	return func(o *options) {
		o.history = queues
	}
}

func buildOptions(optArgs []Option) *options {
	opts := options{
		queue:    DefaultMsgQueue,
		payload:  DefaultMsgPayload,
		priority: DefaultMsgPriority,
		history:  []string{},
	}
	for _, fn := range optArgs {
		fn(&opts)
	}
	return &opts
}
