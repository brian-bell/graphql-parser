// Package ast defines the GraphQL abstract syntax tree, including node types,
// source-position tracking, and traversal helpers.
package ast

import (
	"fmt"
	"sync"
	"unicode/utf8"
)

// Position is a 1-based line and column into a [Source]. Columns count
// codepoints (not bytes), matching graphql-js. The zero value (Position{}) is
// interpreted as Position{Line: 1, Column: 1} when used as a [Source.LocationOffset].
type Position struct {
	Line   int // 1-based
	Column int // 1-based, in codepoints
}

// String returns "line:column".
func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// Source is a GraphQL document together with an identifying name (used in
// error messages) and an optional offset describing its location within a
// larger file (e.g. an embedded `gql` template tag). A Source must not be
// copied after first use; pass it by pointer.
type Source struct {
	// Body is the document text. Tokens hold byte offsets into this string.
	Body string

	// Name is a human-readable identifier (typically a filename) used in error
	// messages. It is not interpreted otherwise.
	Name string

	// LocationOffset shifts reported positions so that they reference a larger
	// containing file. The zero value is treated as Position{Line: 1, Column: 1}
	// (no shift). Column is applied only to the first line of Body.
	LocationOffset Position

	once       sync.Once
	lineStarts []int // byte offsets where each line begins; lineStarts[0] == 0
}

// Loc covers a half-open byte range [Start, End) into Source.Body. It describes
// the full extent of an AST node, not just the start of its first token.
type Loc struct {
	Start, End int
	Source     *Source
}

// StartPosition returns the 1-based line/column of [Loc.Start].
func (l *Loc) StartPosition() Position { return l.Source.PositionAt(l.Start) }

// EndPosition returns the 1-based line/column of [Loc.End].
func (l *Loc) EndPosition() Position { return l.Source.PositionAt(l.End) }

// PositionAt returns the 1-based {Line, Column} of the byte at offset.
// Negative offsets are treated as 0; offsets greater than len(Body) are
// clamped to len(Body). The first call on a given Source builds a line-start
// index lazily; subsequent calls reuse it.
func (s *Source) PositionAt(offset int) Position {
	if offset < 0 {
		offset = 0
	}
	if offset > len(s.Body) {
		offset = len(s.Body)
	}
	s.once.Do(s.computeLineStarts)

	// Binary search: find the largest index i such that lineStarts[i] <= offset.
	// The 1-based line number is i+1.
	lo, hi := 0, len(s.lineStarts)
	for lo < hi {
		mid := (lo + hi) / 2
		if s.lineStarts[mid] <= offset {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	line := lo // count of starts <= offset == 1-based line number
	lineStart := s.lineStarts[line-1]
	column := utf8.RuneCountInString(s.Body[lineStart:offset]) + 1

	return s.applyLocationOffset(Position{Line: line, Column: column})
}

func (s *Source) computeLineStarts() {
	starts := []int{0}
	for i := 0; i < len(s.Body); i++ {
		switch s.Body[i] {
		case '\n':
			starts = append(starts, i+1)
		case '\r':
			if i+1 < len(s.Body) && s.Body[i+1] == '\n' {
				starts = append(starts, i+2)
				i++
			} else {
				starts = append(starts, i+1)
			}
		}
	}
	s.lineStarts = starts
}

func (s *Source) applyLocationOffset(p Position) Position {
	off := s.LocationOffset
	if off.Line == 0 && off.Column == 0 {
		return p
	}
	line := p.Line + off.Line - 1
	column := p.Column
	if p.Line == 1 {
		column = p.Column + off.Column - 1
	}
	return Position{Line: line, Column: column}
}
