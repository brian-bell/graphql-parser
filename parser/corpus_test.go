package parser_test

import (
	"strings"
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/parser"
)

// This file ports a representative slice of the graphql-js parser /
// lexer / schema-parser test corpus. See
// parser/testdata/graphql-js/README.md for the conformance bar and the
// process for adding cases.

// ----- Parser ---------------------------------------------------------------

type parseAcceptCase struct {
	name string
	body string
}

var parserAcceptCases = []parseAcceptCase{
	{"anonymous query shorthand", "{ field }"},
	{"named query", "query MyQuery { field }"},
	{"variables", "query Q($x: Int) { field(x: $x) }"},
	{"variable with default", "query Q($x: Int = 1) { field(x: $x) }"},
	{"directive on operation", "query Q @cached { field }"},
	{"mutation", "mutation { update }"},
	{"subscription", "subscription { events }"},
	{"fragment definition", "fragment F on T { x }"},
	{"fragment spread", "{ ...F }"},
	{"inline fragment with type", "{ ... on T { x } }"},
	{"inline fragment without type", "{ ... { x } }"},
	{"inline fragment with directive only", "{ ... @skip(if: true) { x } }"},
	{"alias", "{ short: longName }"},
	{"nested selections", "{ a { b { c } } }"},
	{"arguments mixed types", `{ f(int: 1, float: 1.5, str: "hi", bool: true, n: null, en: ENUM, lst: [1,2], obj: {a: 1}) }`},
	{"comma is whitespace", "{ a, b, c }"},
	{"newlines OK", "{\n  a\n  b\n}"},
	{"BOM at start", "\ufeff{ x }"},
	{"comment ignored", "# leading\n{ x }"},
	{"empty list", "{ f(x: []) }"},
	{"empty object", "{ f(x: {}) }"},
	{"variable only at value position", "{ f(x: $v) }"},
	{"directive args use values", `{ f @dir(a: 1, b: "x") }`},
	{"multiple definitions", "{ x } query Q { y } fragment F on T { z }"},
}

func TestCorpus_ParserAccept(t *testing.T) {
	for _, c := range parserAcceptCases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := parser.Parse(c.body); err != nil {
				t.Errorf("Parse(%q) errored: %v", c.body, err)
			}
		})
	}
}

type parseRejectCase struct {
	name       string
	body       string
	wantPhrase string // substring of the error message
}

var parserRejectCases = []parseRejectCase{
	{"empty document", "", "Expected"},
	{"unexpected character", "{ a $ }", "Expected"},
	{"unexpected EOF in selection", "{ a", "Expected"},
	{"empty selection set", "{}", "Expected"},
	{"empty arguments", "{ a() }", "Expected"},
	{"empty variable definitions", "query () { x }", "Expected"},
	{"variable in default", "query ($a: Int = $b) { x }", "Unexpected variable"},
	{"reserved fragment name 'on'", "fragment on on T { x }", "Unexpected"},
	{"missing colon in argument", "{ a(b 1) }", "Expected :"},
	{"missing colon in object field", `{ f(x: { a 1 }) }`, "Expected :"},
	{"trailing tokens", "query Q { x } extra", "Unexpected"},
	{"non-null on non-null", "query Q($x: Int!!) { f }", "Expected"},
	{"unmatched bracket", "query Q($x: [Int) { f }", "Expected"},
	{"unmatched paren", "query Q($x: Int { f }", "Expected"},
	{"bad number", "{ a(x: 01) }", "Invalid number"},
}

func TestCorpus_ParserReject(t *testing.T) {
	for _, c := range parserRejectCases {
		t.Run(c.name, func(t *testing.T) {
			_, err := parser.Parse(c.body)
			if err == nil {
				t.Fatalf("Parse(%q) accepted; expected an error", c.body)
			}
			if !strings.Contains(err.Error(), c.wantPhrase) {
				t.Errorf("Parse(%q) error %q does not contain %q",
					c.body, err.Error(), c.wantPhrase)
			}
		})
	}
}

// ----- Lexer (via the parser, since lexer tests are in lexer_test.go) ------

type lexAcceptCase struct {
	name  string
	body  string
	check func(t *testing.T, doc *ast.Document)
}

var lexerCases = []lexAcceptCase{
	{
		"surrogate pair fixed escape",
		`{ f(x: "😀") }`,
		func(t *testing.T, doc *ast.Document) {
			arg := doc.Definitions[0].(*ast.OperationDefinition).
				SelectionSet.Selections[0].(*ast.Field).Arguments[0]
			if sv, ok := arg.Value.(*ast.StringValue); !ok || sv.Value != "\U0001F600" {
				t.Errorf("expected emoji; got %v", arg.Value)
			}
		},
	},
	{
		"variable Unicode escape",
		`{ f(x: "\u{1F600}") }`,
		func(t *testing.T, doc *ast.Document) {
			arg := doc.Definitions[0].(*ast.OperationDefinition).
				SelectionSet.Selections[0].(*ast.Field).Arguments[0]
			if sv, ok := arg.Value.(*ast.StringValue); !ok || sv.Value != "\U0001F600" {
				t.Errorf("expected emoji; got %v", arg.Value)
			}
		},
	},
	{
		"block string dedent",
		"{ f(x: \"\"\"\n  a\n    b\n  \"\"\") }",
		func(t *testing.T, doc *ast.Document) {
			arg := doc.Definitions[0].(*ast.OperationDefinition).
				SelectionSet.Selections[0].(*ast.Field).Arguments[0]
			sv := arg.Value.(*ast.StringValue)
			if !sv.Block {
				t.Error("expected Block: true")
			}
			if sv.Value != "a\n  b" {
				t.Errorf("got %q; want %q", sv.Value, "a\n  b")
			}
		},
	},
	{
		"block string escaped triple",
		"{ f(x: \"\"\"\\\"\"\"hi\"\"\") }",
		func(t *testing.T, doc *ast.Document) {
			arg := doc.Definitions[0].(*ast.OperationDefinition).
				SelectionSet.Selections[0].(*ast.Field).Arguments[0]
			sv := arg.Value.(*ast.StringValue)
			if sv.Value != `"""hi` {
				t.Errorf("got %q", sv.Value)
			}
		},
	},
}

func TestCorpus_Lexer(t *testing.T) {
	for _, c := range lexerCases {
		t.Run(c.name, func(t *testing.T) {
			doc, err := parser.Parse(c.body)
			if err != nil {
				t.Fatalf("Parse(%q) errored: %v", c.body, err)
			}
			c.check(t, doc)
		})
	}
}

// ----- Schema parser --------------------------------------------------------

var schemaAcceptCases = []parseAcceptCase{
	{"schema definition", "schema { query: Query }"},
	{"schema with directive", "schema @secured { query: Query }"},
	{"scalar", "scalar URL"},
	{"scalar with directive", `scalar URL @specifiedBy(url: "https://x")`},
	{"object type", "type T { x: Int }"},
	{"object type with description", `"the type" type T { x: Int }`},
	{"object type with block description", "\"\"\"\nblock\n\"\"\"\ntype T { x: Int }"},
	{"object implements interfaces", "type T implements A & B { x: Int }"},
	{"object implements with leading amp", "type T implements & A & B { x: Int }"},
	{"interface type", "interface I { x: Int }"},
	{"interface implementing interface", "interface I implements N { id: ID! }"},
	{"union type", "union U = A | B"},
	{"union with leading pipe", "union U = | A | B"},
	{"union no members", "union Empty"},
	{"enum type", "enum E { A B C }"},
	{"enum with description", "enum Color { \"red\" RED \"green\" GREEN }"},
	{"input type", "input In { x: Int = 1 }"},
	{"input with directives", "input In { x: Int = 1 @check }"},
	{"directive definition", "directive @auth on FIELD"},
	{"directive with args", "directive @auth(role: Role!) on FIELD"},
	{"directive repeatable", "directive @tag(name: String!) repeatable on OBJECT"},
	{"directive multiple locations", "directive @x on FIELD | OBJECT | ARGUMENT_DEFINITION"},
	{"schema extension", "extend schema @secured { mutation: M }"},
	{"object extension", "extend type T { y: Int }"},
	{"object extension implements only", "extend type T implements I"},
	{"interface extension", "extend interface I { y: Int }"},
	{"union extension", "extend union U = C"},
	{"enum extension", "extend enum E { D }"},
	{"input extension", "extend input In { y: Int }"},
	{"scalar extension directive", "extend scalar URL @new"},
}

func TestCorpus_SchemaAccept(t *testing.T) {
	for _, c := range schemaAcceptCases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := parser.Parse(c.body); err != nil {
				t.Errorf("Parse(%q) errored: %v", c.body, err)
			}
		})
	}
}

var schemaRejectCases = []parseRejectCase{
	{"description on extension", `"x" extend type T { y: Int }`, "description"},
	{"empty schema extension", "extend schema", "extension must define"},
	{"empty type extension", "extend type T", "extension must define"},
	{"unknown directive location", "directive @x on FROBNICATE", "directive location"},
	{"reserved enum value true", "enum E { true }", "true, false, or null"},
	{"reserved enum value false", "enum E { false }", "true, false, or null"},
	{"reserved enum value null", "enum E { null }", "true, false, or null"},
	{"missing operation type", "schema { }", "operation type"},
	{"missing fields after type", "type T { }", "field definition"},
	{"empty arguments definition", "type T { f(): Int }", "argument definition"},
	{"empty input fields", "input In { }", "input field definition"},
	{"description before extension keyword", `"x" extend scalar URL @x`, "description"},
}

func TestCorpus_SchemaReject(t *testing.T) {
	for _, c := range schemaRejectCases {
		t.Run(c.name, func(t *testing.T) {
			_, err := parser.Parse(c.body)
			if err == nil {
				t.Fatalf("Parse(%q) accepted; expected an error", c.body)
			}
			if !strings.Contains(err.Error(), c.wantPhrase) {
				t.Errorf("Parse(%q) error %q does not contain %q",
					c.body, err.Error(), c.wantPhrase)
			}
		})
	}
}
