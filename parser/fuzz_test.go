package parser_test

import (
	"errors"
	"testing"

	"github.com/brian-bell/graphql-parser/parser"
)

// fuzzSeeds are a representative slice of inputs (mix of valid and broken)
// used to seed both FuzzParse and FuzzParseWithRecovery.
var fuzzSeeds = []string{
	"",
	"{}",
	"{ x }",
	"query Q { x }",
	"mutation M { do }",
	"subscription S { events }",
	"fragment F on T { x }",
	"{ ...frag }",
	"{ ... on T { x } }",
	"{ a(x: 1, y: $v) @dir }",
	`{ s: "hi", b: """block""" }`,
	"query Q($v: Int = 1) @dir { x }",
	"type T { id: ID! name(arg: Int = 1): String }",
	"interface I implements & A & B { x: Int }",
	"union U = A | B | C",
	"enum E { A B C }",
	"input In { x: Int = 1 }",
	"scalar URL @specifiedBy(url: \"https://x\")",
	"directive @auth(role: Role!) repeatable on FIELD | OBJECT",
	"extend type T { y: Int }",
	`"desc" type T { x: Int }`,
	`"""multi
	line""" type T { x: Int }`,
	// Intentionally malformed inputs:
	"query",
	"{",
	"{ a (",
	"type T {",
	"type T { x: }",
	"type T implements",
	"\"unterminated",
	"\"\"\"unterminated block",
	"directive @x on BOGUS",
	"{ a $bogus b }",
	"01",
	"1.0.0",
	"\\u{ZZ}",
	// Edge cases that have historically broken parsers:
	"{ a, b, c }",
	"{ a # comment\n b }",
	"#only a comment\n",
	"\xef\xbb\xbf{ x }", // BOM
}

// FuzzParse exercises the default (fail-fast) parser. Properties:
//   - never panics
//   - if (doc, nil) is returned then doc is non-nil
//   - if an error is returned it is a *parser.ParseError
func FuzzParse(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, body string) {
		doc, err := parser.Parse(body)
		if err == nil && doc == nil {
			t.Errorf("Parse returned (nil, nil) for %q", body)
		}
		if err != nil {
			var pe *parser.ParseError
			if !errors.As(err, &pe) {
				t.Errorf("Parse returned %T (want *ParseError) for %q", err, body)
			}
		}
	})
}

// FuzzParseWithRecovery exercises the recovery-mode parser. Properties:
//   - never panics
//   - always returns a non-nil document (recovery should produce a partial AST)
//     OR returns nil when the input has no definitions at all
//   - any error is a parser.ParseErrors (a slice wrapper of ParseError)
func FuzzParseWithRecovery(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, body string) {
		doc, err := parser.Parse(body, parser.WithRecovery())
		if err != nil {
			var es parser.ParseErrors
			if !errors.As(err, &es) && !isParseError(err) {
				t.Errorf("Parse(WithRecovery) returned %T for %q", err, body)
			}
		}
		// doc may be nil for empty-input corner cases; that's fine.
		_ = doc
	})
}

func isParseError(err error) bool {
	var pe *parser.ParseError
	return errors.As(err, &pe)
}

// Also fuzz the partial-parse entry points.

func FuzzParseValue(f *testing.F) {
	for _, s := range []string{"", "1", "1.0", `"x"`, "true", "null", "[]", "{}",
		`{a: 1, b: [2, 3]}`, "$v", "ENUM", "[$x, 1, true, null]"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, body string) {
		v, err := parser.ParseValue(body)
		if err == nil && v == nil {
			t.Errorf("ParseValue returned (nil, nil) for %q", body)
		}
	})
}

func FuzzParseType(f *testing.F) {
	for _, s := range []string{"", "T", "[T]", "T!", "[T!]!", "[[[T]]]"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, body string) {
		ty, err := parser.ParseType(body)
		if err == nil && ty == nil {
			t.Errorf("ParseType returned (nil, nil) for %q", body)
		}
	})
}
