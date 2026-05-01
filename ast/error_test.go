package ast_test

import (
	"strings"
	"testing"

	"github.com/bellbm/graphql-parser/ast"
)

func TestSyntaxError_Header(t *testing.T) {
	src := &ast.Source{Body: "foo bar baz", Name: "x.graphql"}
	e := &ast.SyntaxError{Source: src, Offset: 4, Message: "boom"}
	got := e.Error()
	if !strings.Contains(got, "Syntax Error: boom") {
		t.Errorf("error %q missing 'Syntax Error: boom'", got)
	}
	if !strings.Contains(got, "x.graphql:1:5") {
		t.Errorf("error %q missing 'x.graphql:1:5'", got)
	}
}

func TestSyntaxError_Snippet_SingleLine(t *testing.T) {
	src := &ast.Source{Body: "foo bar baz", Name: "x.graphql"}
	e := &ast.SyntaxError{Source: src, Offset: 4, Message: "expected EOF"}
	got := e.Error()
	if !strings.Contains(got, "1 | foo bar baz") {
		t.Errorf("error %q missing the offending line", got)
	}
	// Caret should be aligned under "b" in "bar" (column 5).
	if !strings.Contains(got, "  |     ^") {
		t.Errorf("error %q has misaligned caret", got)
	}
}

func TestSyntaxError_Snippet_WithContext(t *testing.T) {
	src := &ast.Source{
		Body: "line1\nline2 with error\nline3",
		Name: "x",
	}
	// Offset of 'w' in "with" on line 2 is 12.
	e := &ast.SyntaxError{Source: src, Offset: 12, Message: "boom"}
	got := e.Error()
	for _, want := range []string{"1 | line1", "2 | line2 with error", "3 | line3"} {
		if !strings.Contains(got, want) {
			t.Errorf("error\n%s\nmissing %q", got, want)
		}
	}
}

func TestSyntaxError_Snippet_FirstLineNoBefore(t *testing.T) {
	src := &ast.Source{Body: "abc\ndef", Name: "x"}
	e := &ast.SyntaxError{Source: src, Offset: 0, Message: "x"}
	got := e.Error()
	if !strings.Contains(got, "1 | abc") {
		t.Errorf("missing 1 | abc in:\n%s", got)
	}
	if !strings.Contains(got, "2 | def") {
		t.Errorf("missing 2 | def (after) in:\n%s", got)
	}
}

func TestSyntaxError_Snippet_LastLineNoAfter(t *testing.T) {
	src := &ast.Source{Body: "abc\ndef", Name: "x"}
	// Offset 4 = 'd' on line 2.
	e := &ast.SyntaxError{Source: src, Offset: 4, Message: "x"}
	got := e.Error()
	if !strings.Contains(got, "1 | abc") {
		t.Errorf("missing line 1 in:\n%s", got)
	}
	if !strings.Contains(got, "2 | def") {
		t.Errorf("missing line 2 in:\n%s", got)
	}
}

func TestSyntaxError_NilSource(t *testing.T) {
	e := &ast.SyntaxError{Message: "no source"}
	got := e.Error()
	if !strings.Contains(got, "no source") {
		t.Errorf("error %q missing message", got)
	}
}

func TestSyntaxError_LocationOffset(t *testing.T) {
	// Source claims to start at line 10, column 5 of a larger file.
	src := &ast.Source{
		Body:           "abc\ndef",
		Name:           "x",
		LocationOffset: ast.Position{Line: 10, Column: 5},
	}
	// Offset 0 → reported as line 10, col 5.
	e := &ast.SyntaxError{Source: src, Offset: 0, Message: "x"}
	got := e.Error()
	if !strings.Contains(got, "x:10:5") {
		t.Errorf("error %q missing shifted position 10:5", got)
	}
	// The gutter uses reported (shifted) line numbers.
	if !strings.Contains(got, "10 | abc") {
		t.Errorf("error %q does not show shifted gutter line 10", got)
	}
}

func TestSyntaxError_OffsetAtEOF(t *testing.T) {
	src := &ast.Source{Body: "abc", Name: "x"}
	e := &ast.SyntaxError{Source: src, Offset: 3, Message: "expected something"}
	// Should not panic, should include the line and a caret past the last char.
	got := e.Error()
	if !strings.Contains(got, "1 | abc") {
		t.Errorf("missing line in:\n%s", got)
	}
}

func TestSyntaxError_MultiByteRunes(t *testing.T) {
	// "héllo world": é is 2 bytes (UTF-8 0xC3 0xA9). Bytes 0..6 = "héllo "
	// (codepoints h é l l o space = 6 runes). Byte 7 = 'w', codepoint column 7.
	src := &ast.Source{Body: "héllo world", Name: "x"}
	e := &ast.SyntaxError{Source: src, Offset: 7, Message: "x"} // 'w'
	got := e.Error()
	if !strings.Contains(got, "x:1:7") {
		t.Errorf("expected codepoint column 7 in:\n%s", got)
	}
	if !strings.Contains(got, "héllo world") {
		t.Errorf("missing source in:\n%s", got)
	}
	// caretCol-1 = 6 spaces, then ^.  Format is "<gutter> | <pad><caret>".
	// Gutter width 1 (single-digit line), so prefix is "  | " then 6 spaces.
	if !strings.Contains(got, "  |       ^") {
		t.Errorf("caret misaligned in:\n%s", got)
	}
}
