package lexer_test

import (
	"strings"
	"testing"

	"github.com/bellbm/graphql-parser/ast"
	"github.com/bellbm/graphql-parser/lexer"
)

func TestLexer_String_Empty(t *testing.T) {
	l := lex(t, `""`)
	tok := mustNext(t, l)
	if tok.Kind != lexer.STRING {
		t.Fatalf("kind = %v; want STRING", tok.Kind)
	}
	if tok.Value != "" {
		t.Errorf("Value = %q; want empty", tok.Value)
	}
	if tok.Block {
		t.Error("Block = true; want false")
	}
}

func TestLexer_String_Simple(t *testing.T) {
	l := lex(t, `"hello, world"`)
	tok := mustNext(t, l)
	if tok.Kind != lexer.STRING {
		t.Fatalf("kind = %v; want STRING", tok.Kind)
	}
	if tok.Value != "hello, world" {
		t.Errorf("Value = %q; want %q", tok.Value, "hello, world")
	}
	if tok.Start != 0 || tok.End != 14 {
		t.Errorf("range = [%d, %d); want [0, 14)", tok.Start, tok.End)
	}
}

func TestLexer_String_Escapes(t *testing.T) {
	cases := []struct {
		body string
		want string
	}{
		{`"\""`, `"`},
		{`"\\"`, `\`},
		{`"\/"`, `/`},
		{`"\b"`, "\b"},
		{`"\f"`, "\f"},
		{`"\n"`, "\n"},
		{`"\r"`, "\r"},
		{`"\t"`, "\t"},
		{`"\"escaped\""`, `"escaped"`},
		{`"a\tb\nc"`, "a\tb\nc"},
	}
	for _, c := range cases {
		t.Run(c.body, func(t *testing.T) {
			l := lex(t, c.body)
			tok := mustNext(t, l)
			if tok.Kind != lexer.STRING {
				t.Fatalf("kind = %v; want STRING", tok.Kind)
			}
			if tok.Value != c.want {
				t.Errorf("Value = %q; want %q", tok.Value, c.want)
			}
		})
	}
}

func TestLexer_String_UnicodeFixed(t *testing.T) {
	cases := []struct {
		body string
		want string
	}{
		{`"é"`, "é"},
		{`"A"`, "A"},
		{`"épique"`, "épique"},
		// Surrogate pair: 😀 = U+1F600 = high D83D, low DE00.
		{`"😀"`, "\U0001F600"},
	}
	for _, c := range cases {
		t.Run(c.body, func(t *testing.T) {
			l := lex(t, c.body)
			tok := mustNext(t, l)
			if tok.Kind != lexer.STRING {
				t.Fatalf("kind = %v; want STRING", tok.Kind)
			}
			if tok.Value != c.want {
				t.Errorf("Value = %q (% x); want %q (% x)", tok.Value, tok.Value, c.want, c.want)
			}
		})
	}
}

func TestLexer_String_UnicodeVariable(t *testing.T) {
	cases := []struct {
		body string
		want string
	}{
		{`"\u{0041}"`, "A"},
		{`"\u{1F600}"`, "\U0001F600"},
		{`"\u{0}"`, "\x00"},
		{`"\u{10FFFF}"`, "\U0010FFFF"},
	}
	for _, c := range cases {
		t.Run(c.body, func(t *testing.T) {
			l := lex(t, c.body)
			tok := mustNext(t, l)
			if tok.Kind != lexer.STRING {
				t.Fatalf("kind = %v; want STRING", tok.Kind)
			}
			if tok.Value != c.want {
				t.Errorf("Value = %q (% x); want %q (% x)", tok.Value, tok.Value, c.want, c.want)
			}
		})
	}
}

func TestLexer_String_Errors(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"unterminated", `"abc`},
		{"unterminated empty", `"`},
		{"newline in string", "\"abc\ndef\""},
		{"CR in string", "\"abc\rdef\""},
		{"control character", "\"abc\x01def\""},
		{"invalid escape", `"\x"`},
		{"bad fixed unicode hex", `"\u00G0"`},
		{"truncated fixed unicode", `"\u12"`},
		{"lone high surrogate", `"\uD83D"`},
		{"lone low surrogate", `"\uDE00"`},
		{"high surrogate with non-low after", `"\uD83DA"`},
		{"variable unicode empty", `"\u{}"`},
		{"variable unicode too long", `"\u{1234567}"`},
		{"variable unicode out of range", `"\u{110000}"`},
		{"variable unicode surrogate", `"\u{D800}"`},
		{"variable unicode unterminated", `"\u{1F600`},
		{"variable unicode bad hex", `"\u{ZZ}"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			l := lex(t, c.body)
			if _, err := l.Next(); err == nil {
				t.Errorf("body %q: expected error", c.body)
			}
		})
	}
}

func TestLexer_String_TabAllowedInString(t *testing.T) {
	// Tab is allowed; only the other control characters (< 0x20 except \t) are
	// rejected mid-string per spec.
	l := lex(t, "\"a\tb\"")
	tok := mustNext(t, l)
	if tok.Kind != lexer.STRING {
		t.Fatalf("kind = %v; want STRING", tok.Kind)
	}
	if tok.Value != "a\tb" {
		t.Errorf("Value = %q; want %q", tok.Value, "a\tb")
	}
}

func TestLexer_BlockString_Empty(t *testing.T) {
	l := lex(t, `""""""`)
	tok := mustNext(t, l)
	if tok.Kind != lexer.STRING {
		t.Fatalf("kind = %v; want STRING", tok.Kind)
	}
	if !tok.Block {
		t.Error("Block = false; want true")
	}
	if tok.Value != "" {
		t.Errorf("Value = %q; want empty", tok.Value)
	}
}

func TestLexer_BlockString_Simple(t *testing.T) {
	l := lex(t, `"""hello"""`)
	tok := mustNext(t, l)
	if !tok.Block {
		t.Error("Block = false; want true")
	}
	if tok.Value != "hello" {
		t.Errorf("Value = %q; want %q", tok.Value, "hello")
	}
}

func TestLexer_BlockString_Dedent(t *testing.T) {
	body := "\"\"\"\n    Hello,\n      World!\n\"\"\""
	l := lex(t, body)
	tok := mustNext(t, l)
	if !tok.Block {
		t.Error("Block = false; want true")
	}
	if tok.Value != "Hello,\n  World!" {
		t.Errorf("Value = %q; want %q", tok.Value, "Hello,\n  World!")
	}
}

func TestLexer_BlockString_EscapedTripleQuote(t *testing.T) {
	body := `"""contains \""" inside"""`
	l := lex(t, body)
	tok := mustNext(t, l)
	if tok.Value != `contains """ inside` {
		t.Errorf("Value = %q; want %q", tok.Value, `contains """ inside`)
	}
}

func TestLexer_BlockString_NoOtherEscapes(t *testing.T) {
	// Inside a block string, only \""" is special. \n is the literal two
	// characters, not a newline.
	body := `"""\n"""`
	l := lex(t, body)
	tok := mustNext(t, l)
	if tok.Value != `\n` {
		t.Errorf("Value = %q; want %q (literal backslash-n)", tok.Value, `\n`)
	}
}

func TestLexer_BlockString_Unterminated(t *testing.T) {
	body := `"""hello`
	l := lex(t, body)
	if _, err := l.Next(); err == nil {
		t.Error("expected unterminated block string error")
	}
}

func TestLexer_BlockString_PositionsCoverDelimiters(t *testing.T) {
	body := `"""hi"""`
	l := lex(t, body)
	tok := mustNext(t, l)
	if tok.Start != 0 || tok.End != len(body) {
		t.Errorf("range = [%d, %d); want [0, %d)", tok.Start, tok.End, len(body))
	}
}

func TestLexer_String_PositionsCoverQuotes(t *testing.T) {
	body := `"hi"`
	l := lex(t, body)
	tok := mustNext(t, l)
	if tok.Start != 0 || tok.End != 4 {
		t.Errorf("range = [%d, %d); want [0, 4)", tok.Start, tok.End)
	}
}

func TestLexer_String_ErrorMessageMentionsString(t *testing.T) {
	l := lex(t, `"\x"`)
	_, err := l.Next()
	if err == nil {
		t.Fatal("expected error")
	}
	var se *ast.SyntaxError
	if !errAs(err, &se) {
		t.Fatalf("expected *ast.SyntaxError, got %T", err)
	}
	if !strings.Contains(strings.ToLower(se.Message), "escape") {
		t.Errorf("message %q does not mention escape", se.Message)
	}
}
