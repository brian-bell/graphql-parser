package ast

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// SyntaxError is a syntax-level error tied to a byte offset in a [Source].
// It is the common error type produced by the lexer and parser; the parser's
// public ParseError is an alias for it.
//
// Error() renders a graphql-js-style message:
//
//	Syntax Error: <message>
//
//	<source.Name>:<line>:<col>
//	N - 1 | <preceding line>
//	N     | <offending line>
//	      |     ^
//	N + 1 | <following line>
//
// LocationOffset on the Source shifts the displayed line numbers and the
// first-line column, so errors in fragments parsed out of larger files
// report the original file's coordinates.
type SyntaxError struct {
	Source  *Source
	Offset  int
	Message string
}

// Error renders a multi-line, graphql-js-style error message.
func (e *SyntaxError) Error() string {
	if e.Source == nil {
		return "Syntax Error: " + e.Message
	}
	pos := e.Position()
	name := e.Source.Name
	if name == "" {
		name = "GraphQL"
	}
	snippet := e.Source.snippetAt(e.Offset)
	if snippet == "" {
		return fmt.Sprintf("Syntax Error: %s\n\n%s:%s", e.Message, name, pos)
	}
	return fmt.Sprintf("Syntax Error: %s\n\n%s:%s\n%s", e.Message, name, pos, snippet)
}

// Position returns the 1-based line/column of the error, with [Source.LocationOffset]
// applied.
func (e *SyntaxError) Position() Position {
	if e.Source == nil {
		return Position{}
	}
	return e.Source.PositionAt(e.Offset)
}

// snippetAt returns up to three lines of source context around offset, with
// reported line numbers in the gutter and a caret aligned beneath offset.
func (s *Source) snippetAt(offset int) string {
	if offset < 0 {
		offset = 0
	}
	if offset > len(s.Body) {
		offset = len(s.Body)
	}
	s.once.Do(s.computeLineStarts)
	if len(s.lineStarts) == 0 {
		return ""
	}

	// Raw (unshifted) line/column.
	rawLine := lineNumberFor(s.lineStarts, offset)
	lineStart := s.lineStarts[rawLine-1]
	rawColumn := utf8.RuneCountInString(s.Body[lineStart:offset]) + 1

	off := s.LocationOffset
	if off.Line == 0 && off.Column == 0 {
		off = Position{Line: 1, Column: 1}
	}
	displayLine := func(raw int) int { return raw + off.Line - 1 }
	displayColumn := func(raw, rawL int) int {
		if rawL == 1 {
			return raw + off.Column - 1
		}
		return raw
	}

	type ctxLine struct {
		num     int
		text    string
		isError bool
	}
	var lines []ctxLine
	if rawLine > 1 {
		lines = append(lines, ctxLine{num: displayLine(rawLine - 1), text: s.lineText(rawLine - 1)})
	}
	lines = append(lines, ctxLine{num: displayLine(rawLine), text: s.lineText(rawLine), isError: true})
	if rawLine < len(s.lineStarts) {
		lines = append(lines, ctxLine{num: displayLine(rawLine + 1), text: s.lineText(rawLine + 1)})
	}

	gutterWidth := len(strconv.Itoa(lines[len(lines)-1].num))
	caretCol := displayColumn(rawColumn, rawLine)

	var sb strings.Builder
	for _, l := range lines {
		ns := strconv.Itoa(l.num)
		sb.WriteString(strings.Repeat(" ", gutterWidth-len(ns)))
		sb.WriteString(ns)
		sb.WriteString(" | ")
		sb.WriteString(l.text)
		sb.WriteString("\n")
		if l.isError {
			sb.WriteString(strings.Repeat(" ", gutterWidth))
			sb.WriteString(" | ")
			if caretCol > 1 {
				sb.WriteString(strings.Repeat(" ", caretCol-1))
			}
			sb.WriteString("^")
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// lineNumberFor returns the 1-based line number containing offset, given a
// pre-computed line-start table.
func lineNumberFor(lineStarts []int, offset int) int {
	lo, hi := 0, len(lineStarts)
	for lo < hi {
		mid := (lo + hi) / 2
		if lineStarts[mid] <= offset {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

// lineText returns the text of a 1-based line, stripped of its terminator.
func (s *Source) lineText(line int) string {
	if line < 1 || line > len(s.lineStarts) {
		return ""
	}
	start := s.lineStarts[line-1]
	var end int
	if line < len(s.lineStarts) {
		end = s.lineStarts[line]
		for end > start && (s.Body[end-1] == '\n' || s.Body[end-1] == '\r') {
			end--
		}
	} else {
		end = len(s.Body)
	}
	return s.Body[start:end]
}
