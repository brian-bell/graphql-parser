package parser

// Option configures a parse operation. Future additions (e.g. WithRecovery,
// WithComments, WithExperimental) will land here without API churn —
// every Parse* entry point already accepts variadic options.
type Option func(*config)

type config struct {
	// recovery and preserveComments will be wired up by phases 11 and 12.
	recovery         bool
	preserveComments bool
}

func newConfig(opts []Option) *config {
	cfg := &config{}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}
	return cfg
}
