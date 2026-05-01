package parser

import (
	"strings"

	"github.com/bellbm/graphql-parser/ast"
)

// ParseError is the syntax-error type produced by all Parse* entry points.
// It is an alias for ast.SyntaxError; the underlying type is shared with the
// lexer so callers don't need to handle two error types.
type ParseError = ast.SyntaxError

// ParseErrors aggregates multiple ParseErrors, used when [WithRecovery] is
// enabled. It implements error and supports errors.Is / errors.As / errors.Unwrap
// for unwrapping into the underlying slice.
type ParseErrors []*ParseError

// Error renders one error per line.
func (es ParseErrors) Error() string {
	if len(es) == 0 {
		return "no syntax errors"
	}
	if len(es) == 1 {
		return es[0].Error()
	}
	var sb strings.Builder
	for i, e := range es {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(e.Error())
	}
	return sb.String()
}

// Unwrap exposes the underlying ParseError slice for errors.Is / errors.As.
func (es ParseErrors) Unwrap() []error {
	out := make([]error, len(es))
	for i, e := range es {
		out[i] = e
	}
	return out
}
