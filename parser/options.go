package parser

// Option configures a parse operation. Future additions (e.g. WithComments,
// WithExperimental) will land here without API churn — every Parse* entry
// point already accepts variadic options.
type Option func(*config)

type config struct {
	recovery         bool
	preserveComments bool
}

// WithRecovery enables error recovery: the parser collects syntax errors
// instead of aborting on the first one, inserting Bad{Definition, Field, Value,
// Type} placeholder nodes where it had to resynchronize. Document parsing can
// return a partial [ast.Document]. Value and type entry points recover only at
// the root entry: a malformed value/type input returns one BadValue or BadType
// for the whole partial input rather than nested placeholders. The returned
// error is a [ParseErrors] aggregating every error found.
//
// When this option is not set, the parser fails fast on the first syntax
// error and the conformance corpus runs in this default mode.
func WithRecovery() Option {
	return func(c *config) { c.recovery = true }
}

// WithComments enables comment preservation. When set, the parser collects
// '#' comments and attaches them as Leading trivia on the next AST node it
// produces — at top-level Definitions and inside type-system definition
// bodies (field, enum-value, and input-value definitions). The
// Comments field on every AST node is otherwise nil.
//
// Trailing comments and value/type-level comments are not yet attributed;
// callers needing the full comment stream can read raw COMMENT tokens
// directly from a [lexer.Lexer] constructed with [lexer.WithComments].
func WithComments() Option {
	return func(c *config) { c.preserveComments = true }
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
