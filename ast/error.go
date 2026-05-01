package ast

import "fmt"

// SyntaxError is a syntax-level error tied to a byte offset in a [Source].
// It is the common error type produced by the lexer and parser; the parser's
// public ParseError wraps or aliases it. Phase 8 enhances Error() with a
// graphql-js-style source-line snippet and caret pointer.
type SyntaxError struct {
	Source  *Source
	Offset  int
	Message string
}

// Error returns "<name> (<line>:<column>): <message>".
func (e *SyntaxError) Error() string {
	name := "GraphQL"
	if e.Source != nil && e.Source.Name != "" {
		name = e.Source.Name
	}
	return fmt.Sprintf("%s (%s): %s", name, e.Position(), e.Message)
}

// Position returns the 1-based line/column of the error.
func (e *SyntaxError) Position() Position {
	if e.Source == nil {
		return Position{}
	}
	return e.Source.PositionAt(e.Offset)
}
