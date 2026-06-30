package lexer_test

import (
	"strings"
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/lexer"
)

func lex(t *testing.T, body string) *lexer.Lexer {
	t.Helper()
	return lexer.New(&ast.Source{Body: body, Name: "test.graphql"})
}

func mustNext(t *testing.T, l *lexer.Lexer) lexer.Token {
	t.Helper()
	tok, err := l.Next()
	if err != nil {
		t.Fatalf("Next: unexpected error: %v", err)
	}
	return tok
}

func TestLexer_Empty(t *testing.T) {
	l := lex(t, "")
	tok := mustNext(t, l)
	if tok.Kind != lexer.EOF {
		t.Errorf("got %v; want EOF", tok.Kind)
	}
	if tok.Start != 0 || tok.End != 0 {
		t.Errorf("EOF token has Start=%d End=%d; want 0,0", tok.Start, tok.End)
	}
	// Calling Next at EOF returns EOF again.
	tok2 := mustNext(t, l)
	if tok2.Kind != lexer.EOF {
		t.Errorf("second call: got %v; want EOF", tok2.Kind)
	}
}

func TestLexer_Punctuators(t *testing.T) {
	cases := []struct {
		body string
		kind lexer.Kind
	}{
		{"!", lexer.BANG},
		{"$", lexer.DOLLAR},
		{"&", lexer.AMP},
		{"(", lexer.LPAREN},
		{")", lexer.RPAREN},
		{"...", lexer.SPREAD},
		{":", lexer.COLON},
		{"=", lexer.EQUALS},
		{"@", lexer.AT},
		{"[", lexer.LBRACKET},
		{"]", lexer.RBRACKET},
		{"{", lexer.LBRACE},
		{"|", lexer.PIPE},
		{"}", lexer.RBRACE},
	}
	for _, c := range cases {
		t.Run(c.body, func(t *testing.T) {
			l := lex(t, c.body)
			tok := mustNext(t, l)
			if tok.Kind != c.kind {
				t.Errorf("%q: kind = %v; want %v", c.body, tok.Kind, c.kind)
			}
			if tok.Start != 0 || tok.End != len(c.body) {
				t.Errorf("%q: range = [%d, %d); want [0, %d)", c.body, tok.Start, tok.End, len(c.body))
			}
			if eof := mustNext(t, l); eof.Kind != lexer.EOF {
				t.Errorf("after %q: got %v; want EOF", c.body, eof.Kind)
			}
		})
	}
}

func TestLexer_SpreadInsufficientDots(t *testing.T) {
	for _, body := range []string{".", ".."} {
		l := lex(t, body)
		if _, err := l.Next(); err == nil {
			t.Errorf("%q: expected error", body)
		}
	}
}

func TestLexer_Name(t *testing.T) {
	cases := []string{"a", "abc", "_under", "_", "A1", "X9_Y", "query"}
	for _, body := range cases {
		t.Run(body, func(t *testing.T) {
			l := lex(t, body)
			tok := mustNext(t, l)
			if tok.Kind != lexer.NAME {
				t.Errorf("kind = %v; want NAME", tok.Kind)
			}
			if tok.Value != body {
				t.Errorf("Value = %q; want %q", tok.Value, body)
			}
			if tok.Start != 0 || tok.End != len(body) {
				t.Errorf("range = [%d, %d); want [0, %d)", tok.Start, tok.End, len(body))
			}
		})
	}
}

func TestLexer_NameNotStartingWithDigit(t *testing.T) {
	// "1abc" should lex as INT 1, then error at NAME continuation, since the
	// number cannot be immediately followed by a NameStart character.
	l := lex(t, "1abc")
	if _, err := l.Next(); err == nil {
		t.Error("expected error for '1abc'")
	}
}

func TestLexer_Int(t *testing.T) {
	cases := []string{"0", "1", "12", "123", "9876543210", "-0", "-1", "-12"}
	for _, body := range cases {
		t.Run(body, func(t *testing.T) {
			l := lex(t, body)
			tok := mustNext(t, l)
			if tok.Kind != lexer.INT {
				t.Errorf("kind = %v; want INT", tok.Kind)
			}
			if tok.Value != body {
				t.Errorf("Value = %q; want %q", tok.Value, body)
			}
		})
	}
}

func TestLexer_IntInvalid(t *testing.T) {
	cases := []string{"00", "01", "-01", "0123", "+1", "-", "- 1"}
	for _, body := range cases {
		t.Run(body, func(t *testing.T) {
			l := lex(t, body)
			if _, err := l.Next(); err == nil {
				t.Errorf("%q: expected error", body)
			}
		})
	}
}

func TestLexer_Float(t *testing.T) {
	cases := []string{
		"0.0", "1.0", "1.23", "-1.0", "0.123",
		"1e0", "1E0", "1e+5", "1e-5", "1.0e2",
		"-1.0e-2", "123.456e789",
	}
	for _, body := range cases {
		t.Run(body, func(t *testing.T) {
			l := lex(t, body)
			tok := mustNext(t, l)
			if tok.Kind != lexer.FLOAT {
				t.Errorf("kind = %v; want FLOAT (body %q)", tok.Kind, body)
			}
			if tok.Value != body {
				t.Errorf("Value = %q; want %q", tok.Value, body)
			}
		})
	}
}

func TestLexer_FloatInvalid(t *testing.T) {
	cases := []string{
		"1.", "1.e1", "1.0.0", "1e", "1e+", "1ea",
		".0", "-.0",
	}
	for _, body := range cases {
		t.Run(body, func(t *testing.T) {
			l := lex(t, body)
			if _, err := l.Next(); err == nil {
				t.Errorf("%q: expected error", body)
			}
		})
	}
}

func TestLexer_NumberFollowedBy(t *testing.T) {
	// A number followed by punctuation or whitespace is fine.
	l := lex(t, "1 2.0,3")
	expect := []struct {
		kind  lexer.Kind
		value string
	}{
		{lexer.INT, "1"},
		{lexer.FLOAT, "2.0"},
		{lexer.INT, "3"},
		{lexer.EOF, ""},
	}
	for i, want := range expect {
		tok := mustNext(t, l)
		if tok.Kind != want.kind || tok.Value != want.value {
			t.Errorf("token %d: got %v %q; want %v %q", i, tok.Kind, tok.Value, want.kind, want.value)
		}
	}
}

func TestLexer_IgnoreWhitespace(t *testing.T) {
	l := lex(t, "  \t\n\r\n  foo  ")
	tok := mustNext(t, l)
	if tok.Kind != lexer.NAME || tok.Value != "foo" {
		t.Errorf("got %v %q; want NAME foo", tok.Kind, tok.Value)
	}
	if eof := mustNext(t, l); eof.Kind != lexer.EOF {
		t.Errorf("got %v; want EOF", eof.Kind)
	}
}

func TestLexer_IgnoreCommas(t *testing.T) {
	l := lex(t, "a,b,,,c")
	for _, want := range []string{"a", "b", "c"} {
		tok := mustNext(t, l)
		if tok.Kind != lexer.NAME || tok.Value != want {
			t.Errorf("got %v %q; want NAME %q", tok.Kind, tok.Value, want)
		}
	}
	if eof := mustNext(t, l); eof.Kind != lexer.EOF {
		t.Errorf("got %v; want EOF", eof.Kind)
	}
}

func TestLexer_IgnoreBOM(t *testing.T) {
	l := lex(t, "\ufefffoo")
	tok := mustNext(t, l)
	if tok.Kind != lexer.NAME || tok.Value != "foo" {
		t.Errorf("got %v %q; want NAME foo", tok.Kind, tok.Value)
	}
}

func TestLexer_CommentSkippedByDefault(t *testing.T) {
	l := lex(t, "# leading comment\nfoo # trailing\n# another\nbar")
	tok := mustNext(t, l)
	if tok.Kind != lexer.NAME || tok.Value != "foo" {
		t.Errorf("got %v %q; want NAME foo", tok.Kind, tok.Value)
	}
	tok = mustNext(t, l)
	if tok.Kind != lexer.NAME || tok.Value != "bar" {
		t.Errorf("got %v %q; want NAME bar", tok.Kind, tok.Value)
	}
}

func TestLexer_DefaultSkipsComments(t *testing.T) {
	l := lexer.New(&ast.Source{Body: "# hello\nfoo", Name: "test.graphql"})
	tok := mustNext(t, l)
	if tok.Kind != lexer.NAME || tok.Value != "foo" {
		t.Errorf("default New should skip comments; got %v %q", tok.Kind, tok.Value)
	}
}

func TestLexer_WithCommentsOption(t *testing.T) {
	l := lexer.New(&ast.Source{Body: "# hello\nfoo", Name: "test.graphql"}, lexer.WithComments())
	tok := mustNext(t, l)
	if tok.Kind != lexer.COMMENT || tok.Value != " hello" {
		t.Errorf("WithComments should emit COMMENT; got %v %q", tok.Kind, tok.Value)
	}
}

func TestLexer_CommentRetained(t *testing.T) {
	l := lexer.New(&ast.Source{Body: "# hello\nfoo", Name: "test.graphql"}, lexer.WithComments())
	tok := mustNext(t, l)
	if tok.Kind != lexer.COMMENT {
		t.Fatalf("got %v; want COMMENT", tok.Kind)
	}
	if tok.Value != " hello" {
		t.Errorf("comment Value = %q; want %q", tok.Value, " hello")
	}
	tok = mustNext(t, l)
	if tok.Kind != lexer.NAME || tok.Value != "foo" {
		t.Errorf("got %v %q; want NAME foo", tok.Kind, tok.Value)
	}
}

func TestLexer_CommentToEOF(t *testing.T) {
	// Comment with no trailing newline is fine.
	l := lex(t, "foo # final comment, no newline")
	if tok := mustNext(t, l); tok.Kind != lexer.NAME {
		t.Errorf("got %v; want NAME", tok.Kind)
	}
	if eof := mustNext(t, l); eof.Kind != lexer.EOF {
		t.Errorf("got %v; want EOF", eof.Kind)
	}
}

func TestLexer_Peek(t *testing.T) {
	l := lex(t, "foo bar")
	tok1, err := l.Peek()
	if err != nil {
		t.Fatal(err)
	}
	if tok1.Value != "foo" {
		t.Errorf("Peek 1 = %q; want foo", tok1.Value)
	}
	// Peek twice — same result, no advance.
	tok1b, _ := l.Peek()
	if tok1b.Start != tok1.Start || tok1b.End != tok1.End {
		t.Errorf("Peek twice differs: %+v vs %+v", tok1, tok1b)
	}
	// Next returns the peeked token.
	got := mustNext(t, l)
	if got.Value != "foo" {
		t.Errorf("Next after Peek = %q; want foo", got.Value)
	}
	// Then bar.
	got = mustNext(t, l)
	if got.Value != "bar" {
		t.Errorf("Next = %q; want bar", got.Value)
	}
}

func TestLexer_PositionsAreByteOffsets(t *testing.T) {
	l := lex(t, "  foo\n  bar")
	tok := mustNext(t, l)
	if tok.Start != 2 || tok.End != 5 {
		t.Errorf("foo range = [%d, %d); want [2, 5)", tok.Start, tok.End)
	}
	tok = mustNext(t, l)
	if tok.Start != 8 || tok.End != 11 {
		t.Errorf("bar range = [%d, %d); want [8, 11)", tok.Start, tok.End)
	}
}

func TestLexer_UnexpectedCharacter(t *testing.T) {
	cases := []string{"?", "*", "%", "^", ";"}
	for _, body := range cases {
		t.Run(body, func(t *testing.T) {
			l := lex(t, body)
			_, err := l.Next()
			if err == nil {
				t.Errorf("%q: expected error", body)
			}
			var se *ast.SyntaxError
			if !errAs(err, &se) {
				t.Errorf("%q: expected *ast.SyntaxError, got %T", body, err)
			}
		})
	}
}

func TestLexer_ErrorPosition(t *testing.T) {
	l := lex(t, "foo\n  ?")
	if _, err := l.Next(); err != nil {
		t.Fatalf("first Next: %v", err)
	}
	_, err := l.Next()
	if err == nil {
		t.Fatal("expected error")
	}
	var se *ast.SyntaxError
	if !errAs(err, &se) {
		t.Fatalf("expected *ast.SyntaxError, got %T", err)
	}
	if pos := se.Position(); pos != (ast.Position{Line: 2, Column: 3}) {
		t.Errorf("error position = %+v; want {2, 3}", pos)
	}
	// Subsequent message should reflect the offending character.
	if !strings.Contains(se.Message, "?") {
		t.Errorf("error message %q does not mention '?'", se.Message)
	}
}

// errAs is a tiny errors.As shim to keep the test file dependency-free.
func errAs(err error, target **ast.SyntaxError) bool {
	for err != nil {
		if se, ok := err.(*ast.SyntaxError); ok {
			*target = se
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := err.(unwrapper)
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
