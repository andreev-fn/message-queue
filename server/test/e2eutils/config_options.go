package e2eutils

type configOptions struct {
	deadLetteringOn bool
}

type ConfigOption func(*configOptions)

func WithDeadLettering() ConfigOption {
	return func(o *configOptions) {
		o.deadLetteringOn = true
	}
}

func buildConfigOptions(optArgs []ConfigOption) *configOptions {
	opts := configOptions{}
	for _, fn := range optArgs {
		fn(&opts)
	}
	return &opts
}
