package parser_test

import (
	"errors"
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/parser"
)

var (
	_ func(string, ...parser.Option) (*ast.Document, error)      = parser.Parse
	_ func(string, ...parser.Option) (*ast.Document, error)      = parser.ParseSchema
	_ func(*ast.Source, ...parser.Option) (*ast.Document, error) = parser.ParseSource
	_ func(*ast.Source, ...parser.Option) (*ast.Document, error) = parser.ParseSchemaSource
	_ func(string, ...parser.Option) (ast.Value, error)          = parser.ParseValue
	_ func(string, ...parser.Option) (ast.Value, error)          = parser.ParseConstValue
	_ func(string, ...parser.Option) (ast.Type, error)           = parser.ParseType
	_ func(...parser.Option) *parser.Parser                      = parser.New
)

func TestParserReuseDoesNotLeakState(t *testing.T) {
	p := parser.New(parser.WithComments(), parser.WithRecovery())

	first, err := p.Parse(&ast.Source{Body: "# leading\nquery First { one }", Name: "first.graphql"})
	if err != nil {
		t.Fatalf("first Parse returned error: %v", err)
	}
	if len(first.Definitions) != 1 {
		t.Fatalf("first definition count = %d; want 1", len(first.Definitions))
	}
	if comments := first.Definitions[0].(ast.CommentedNode).CommentSlot(); comments == nil || *comments == nil || len((*comments).Leading) != 1 {
		t.Fatalf("first definition did not receive its leading comment")
	}

	bad, err := p.Parse(&ast.Source{Body: "query Bad { ... }", Name: "bad.graphql"})
	if err == nil {
		t.Fatal("bad Parse returned nil error")
	}
	if bad == nil {
		t.Fatal("bad Parse returned nil document in recovery mode")
	}
	var errs parser.ParseErrors
	if !errors.As(err, &errs) || len(errs) == 0 {
		t.Fatalf("bad Parse error = %T; want non-empty parser.ParseErrors", err)
	}

	v, err := p.ParseValue(&ast.Source{Body: "$x", Name: "value.graphql"})
	if err != nil {
		t.Fatalf("ParseValue after recovered document returned error: %v", err)
	}
	if variable, ok := v.(*ast.Variable); !ok || variable.Name != "x" {
		t.Fatalf("ParseValue after recovered document = %T; want variable $x", v)
	}

	second, err := p.Parse(&ast.Source{Body: "query Second { two }", Name: "second.graphql"})
	if err != nil {
		t.Fatalf("second Parse returned leaked error: %v", err)
	}
	if comments := second.Definitions[0].(ast.CommentedNode).CommentSlot(); comments != nil && *comments != nil {
		t.Fatalf("second definition received leaked comments: %#v", *comments)
	}
}

func TestParserMethodsPreserveSourceMetadata(t *testing.T) {
	p := parser.New()

	t.Run("document", func(t *testing.T) {
		src := &ast.Source{Body: "query Get { field }", Name: "doc.graphql", LocationOffset: ast.Position{Line: 10, Column: 4}}
		doc, err := p.Parse(src)
		if err != nil {
			t.Fatal(err)
		}
		assertLocSource(t, doc.Loc, src)
	})

	t.Run("schema", func(t *testing.T) {
		src := &ast.Source{Body: "type User { id: ID }", Name: "schema.graphql", LocationOffset: ast.Position{Line: 20, Column: 7}}
		doc, err := p.ParseSchema(src)
		if err != nil {
			t.Fatal(err)
		}
		assertLocSource(t, doc.Loc, src)
	})

	t.Run("value", func(t *testing.T) {
		src := &ast.Source{Body: "{name: \"Ada\"}", Name: "value.graphql", LocationOffset: ast.Position{Line: 30, Column: 2}}
		v, err := p.ParseValue(src)
		if err != nil {
			t.Fatal(err)
		}
		assertLocSource(t, v.GetLoc(), src)
	})

	t.Run("const value", func(t *testing.T) {
		src := &ast.Source{Body: "[1, 2, 3]", Name: "const-value.graphql", LocationOffset: ast.Position{Line: 40, Column: 9}}
		v, err := p.ParseConstValue(src)
		if err != nil {
			t.Fatal(err)
		}
		assertLocSource(t, v.GetLoc(), src)
	})

	t.Run("type", func(t *testing.T) {
		src := &ast.Source{Body: "[User!]!", Name: "type.graphql", LocationOffset: ast.Position{Line: 50, Column: 3}}
		typ, err := p.ParseType(src)
		if err != nil {
			t.Fatal(err)
		}
		assertLocSource(t, typ.GetLoc(), src)
	})
}

func TestParserMethodErrorsPreserveSourceMetadata(t *testing.T) {
	p := parser.New()
	cases := []struct {
		name  string
		src   *ast.Source
		parse func(*ast.Source) error
	}{
		{
			name: "document",
			src:  &ast.Source{Body: "query", Name: "doc-error.graphql", LocationOffset: ast.Position{Line: 3, Column: 5}},
			parse: func(src *ast.Source) error {
				_, err := p.Parse(src)
				return err
			},
		},
		{
			name: "schema",
			src:  &ast.Source{Body: "query", Name: "schema-error.graphql", LocationOffset: ast.Position{Line: 4, Column: 6}},
			parse: func(src *ast.Source) error {
				_, err := p.ParseSchema(src)
				return err
			},
		},
		{
			name: "value",
			src:  &ast.Source{Body: "!", Name: "value-error.graphql", LocationOffset: ast.Position{Line: 5, Column: 7}},
			parse: func(src *ast.Source) error {
				_, err := p.ParseValue(src)
				return err
			},
		},
		{
			name: "const value",
			src:  &ast.Source{Body: "$x", Name: "const-value-error.graphql", LocationOffset: ast.Position{Line: 6, Column: 8}},
			parse: func(src *ast.Source) error {
				_, err := p.ParseConstValue(src)
				return err
			},
		},
		{
			name: "type",
			src:  &ast.Source{Body: "!", Name: "type-error.graphql", LocationOffset: ast.Position{Line: 7, Column: 9}},
			parse: func(src *ast.Source) error {
				_, err := p.ParseType(src)
				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.parse(tc.src)
			if err == nil {
				t.Fatal("parse returned nil error")
			}
			var syntaxErr *parser.ParseError
			if !errors.As(err, &syntaxErr) {
				t.Fatalf("error = %T; want *parser.ParseError", err)
			}
			if syntaxErr.Source != tc.src {
				t.Fatalf("error source = %p; want %p", syntaxErr.Source, tc.src)
			}
			if pos := syntaxErr.Position(); pos.Line != tc.src.LocationOffset.Line {
				t.Fatalf("error line = %d; want shifted line %d", pos.Line, tc.src.LocationOffset.Line)
			}
		})
	}
}

func TestParserNewNilOptionsMatchesPackageHelpers(t *testing.T) {
	p := parser.New(nil)

	doc, err := p.Parse(&ast.Source{Body: "query Get { field }", Name: "custom.graphql"})
	if err != nil {
		t.Fatalf("Parser.Parse returned error: %v", err)
	}
	compatDoc, err := parser.Parse("query Get { field }", nil)
	if err != nil {
		t.Fatalf("parser.Parse returned error: %v", err)
	}
	if len(doc.Definitions) != len(compatDoc.Definitions) {
		t.Fatalf("definition count = %d; want %d", len(doc.Definitions), len(compatDoc.Definitions))
	}

	value, err := p.ParseValue(&ast.Source{Body: "123", Name: "value.graphql"})
	if err != nil {
		t.Fatalf("Parser.ParseValue returned error: %v", err)
	}
	compatValue, err := parser.ParseValue("123", nil)
	if err != nil {
		t.Fatalf("parser.ParseValue returned error: %v", err)
	}
	if value.GetLoc().End != compatValue.GetLoc().End {
		t.Fatalf("value end = %d; want %d", value.GetLoc().End, compatValue.GetLoc().End)
	}
}

func TestParserValueAndTypeRecoveryReturnsRootBadNodes(t *testing.T) {
	p := parser.New(parser.WithRecovery())

	t.Run("value", func(t *testing.T) {
		src := &ast.Source{Body: "!", Name: "value-recovery.graphql"}
		v, err := p.ParseValue(src)
		assertParseErrors(t, err, 1)
		bad, ok := v.(*ast.BadValue)
		if !ok {
			t.Fatalf("value = %T; want *ast.BadValue", v)
		}
		assertBadLoc(t, bad.Err, bad.Loc, src, 0, 1)
	})

	t.Run("const value", func(t *testing.T) {
		src := &ast.Source{Body: "$x", Name: "const-value-recovery.graphql"}
		v, err := p.ParseConstValue(src)
		assertParseErrors(t, err, 1)
		bad, ok := v.(*ast.BadValue)
		if !ok {
			t.Fatalf("const value = %T; want *ast.BadValue", v)
		}
		assertBadLoc(t, bad.Err, bad.Loc, src, 0, 2)
	})

	t.Run("type", func(t *testing.T) {
		src := &ast.Source{Body: "!", Name: "type-recovery.graphql"}
		typ, err := p.ParseType(src)
		assertParseErrors(t, err, 1)
		bad, ok := typ.(*ast.BadType)
		if !ok {
			t.Fatalf("type = %T; want *ast.BadType", typ)
		}
		assertBadLoc(t, bad.Err, bad.Loc, src, 0, 1)
	})
}

func TestPartialEntryTrailingTokenSemantics(t *testing.T) {
	failFast := parser.New()
	recovery := parser.New(parser.WithRecovery())

	cases := []struct {
		name      string
		body      string
		failFast  func(*ast.Source) (ast.Node, error)
		recovered func(*ast.Source) (ast.Node, error)
		wantBad   string
	}{
		{
			name: "document",
			body: "query Get { field } !",
			failFast: func(src *ast.Source) (ast.Node, error) {
				return failFast.Parse(src)
			},
			recovered: func(src *ast.Source) (ast.Node, error) {
				return recovery.Parse(src)
			},
			wantBad: "definition",
		},
		{
			name: "schema",
			body: "type Query { field: String } !",
			failFast: func(src *ast.Source) (ast.Node, error) {
				return failFast.ParseSchema(src)
			},
			recovered: func(src *ast.Source) (ast.Node, error) {
				return recovery.ParseSchema(src)
			},
			wantBad: "definition",
		},
		{
			name: "value",
			body: "1 2",
			failFast: func(src *ast.Source) (ast.Node, error) {
				return failFast.ParseValue(src)
			},
			recovered: func(src *ast.Source) (ast.Node, error) {
				return recovery.ParseValue(src)
			},
			wantBad: "value",
		},
		{
			name: "const value",
			body: "1 2",
			failFast: func(src *ast.Source) (ast.Node, error) {
				return failFast.ParseConstValue(src)
			},
			recovered: func(src *ast.Source) (ast.Node, error) {
				return recovery.ParseConstValue(src)
			},
			wantBad: "value",
		},
		{
			name: "type",
			body: "Int String",
			failFast: func(src *ast.Source) (ast.Node, error) {
				return failFast.ParseType(src)
			},
			recovered: func(src *ast.Source) (ast.Node, error) {
				return recovery.ParseType(src)
			},
			wantBad: "type",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := &ast.Source{Body: tc.body, Name: tc.name + ".graphql"}
			node, err := tc.failFast(src)
			if err == nil {
				t.Fatal("fail-fast parse returned nil error")
			}
			if tc.wantBad != "definition" && node != nil {
				t.Fatalf("fail-fast node = %T; want nil", node)
			}

			src = &ast.Source{Body: tc.body, Name: tc.name + "-recovery.graphql"}
			node, err = tc.recovered(src)
			assertParseErrors(t, err, 1)
			switch tc.wantBad {
			case "definition":
				doc, ok := node.(*ast.Document)
				if !ok {
					t.Fatalf("recovered node = %T; want *ast.Document", node)
				}
				if _, ok := doc.Definitions[len(doc.Definitions)-1].(*ast.BadDefinition); !ok {
					t.Fatalf("last definition = %T; want *ast.BadDefinition", doc.Definitions[len(doc.Definitions)-1])
				}
			case "value":
				if _, ok := node.(*ast.BadValue); !ok {
					t.Fatalf("recovered node = %T; want *ast.BadValue", node)
				}
			case "type":
				if _, ok := node.(*ast.BadType); !ok {
					t.Fatalf("recovered node = %T; want *ast.BadType", node)
				}
			}
		})
	}
}

func TestEntryEOFRecoverySemantics(t *testing.T) {
	failFast := parser.New()
	recovery := parser.New(parser.WithRecovery())

	cases := []struct {
		name      string
		failFast  func(*ast.Source) (ast.Node, error)
		recovered func(*ast.Source) (ast.Node, error)
		wantBad   string
	}{
		{
			name: "document",
			failFast: func(src *ast.Source) (ast.Node, error) {
				return failFast.Parse(src)
			},
			recovered: func(src *ast.Source) (ast.Node, error) {
				return recovery.Parse(src)
			},
		},
		{
			name: "schema",
			failFast: func(src *ast.Source) (ast.Node, error) {
				return failFast.ParseSchema(src)
			},
			recovered: func(src *ast.Source) (ast.Node, error) {
				return recovery.ParseSchema(src)
			},
		},
		{
			name: "value",
			failFast: func(src *ast.Source) (ast.Node, error) {
				return failFast.ParseValue(src)
			},
			recovered: func(src *ast.Source) (ast.Node, error) {
				return recovery.ParseValue(src)
			},
			wantBad: "value",
		},
		{
			name: "const value",
			failFast: func(src *ast.Source) (ast.Node, error) {
				return failFast.ParseConstValue(src)
			},
			recovered: func(src *ast.Source) (ast.Node, error) {
				return recovery.ParseConstValue(src)
			},
			wantBad: "value",
		},
		{
			name: "type",
			failFast: func(src *ast.Source) (ast.Node, error) {
				return failFast.ParseType(src)
			},
			recovered: func(src *ast.Source) (ast.Node, error) {
				return recovery.ParseType(src)
			},
			wantBad: "type",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := &ast.Source{Name: tc.name + "-eof.graphql"}
			node, err := tc.failFast(src)
			if err == nil {
				t.Fatal("fail-fast parse returned nil error")
			}
			if tc.wantBad != "" && node != nil {
				t.Fatalf("fail-fast node = %T; want nil", node)
			}

			src = &ast.Source{Name: tc.name + "-eof-recovery.graphql"}
			node, err = tc.recovered(src)
			assertParseErrors(t, err, 1)
			switch tc.wantBad {
			case "":
				// Document/schema entry points may return a typed nil document
				// through the ast.Node test adapter; the ParseErrors assertion
				// above is the compatibility contract being documented here.
			case "value":
				bad, ok := node.(*ast.BadValue)
				if !ok {
					t.Fatalf("recovered node = %T; want *ast.BadValue", node)
				}
				assertBadLoc(t, bad.Err, bad.Loc, src, 0, 0)
			case "type":
				bad, ok := node.(*ast.BadType)
				if !ok {
					t.Fatalf("recovered node = %T; want *ast.BadType", node)
				}
				assertBadLoc(t, bad.Err, bad.Loc, src, 0, 0)
			}
		})
	}
}

func TestParserValueAndTypeRecoveryTreatsNestedFailureAsRootFailure(t *testing.T) {
	p := parser.New(parser.WithRecovery())

	t.Run("value", func(t *testing.T) {
		src := &ast.Source{Body: "[1, !, 2]", Name: "nested-value.graphql"}
		v, err := p.ParseValue(src)
		assertParseErrors(t, err, 1)
		bad, ok := v.(*ast.BadValue)
		if !ok {
			t.Fatalf("value = %T; want *ast.BadValue", v)
		}
		assertBadLoc(t, bad.Err, bad.Loc, src, 0, len(src.Body))
	})

	t.Run("type", func(t *testing.T) {
		src := &ast.Source{Body: "[User!!]", Name: "nested-type.graphql"}
		typ, err := p.ParseType(src)
		assertParseErrors(t, err, 1)
		bad, ok := typ.(*ast.BadType)
		if !ok {
			t.Fatalf("type = %T; want *ast.BadType", typ)
		}
		assertBadLoc(t, bad.Err, bad.Loc, src, 0, len(src.Body))
	})
}

func TestParserValueAndTypeRecoveryConsumesLexerErrors(t *testing.T) {
	p := parser.New(parser.WithRecovery())

	valueBodies := []string{"?", "01", "-", "\"unterminated"}
	for _, body := range valueBodies {
		t.Run("value/"+body, func(t *testing.T) {
			src := &ast.Source{Body: body, Name: "value-lexer-error.graphql"}
			v, err := p.ParseValue(src)
			assertParseErrors(t, err, 1)
			bad, ok := v.(*ast.BadValue)
			if !ok {
				t.Fatalf("value = %T; want *ast.BadValue", v)
			}
			assertBadLoc(t, bad.Err, bad.Loc, src, 0, len(src.Body))
		})

		t.Run("const value/"+body, func(t *testing.T) {
			src := &ast.Source{Body: body, Name: "const-value-lexer-error.graphql"}
			v, err := p.ParseConstValue(src)
			assertParseErrors(t, err, 1)
			bad, ok := v.(*ast.BadValue)
			if !ok {
				t.Fatalf("const value = %T; want *ast.BadValue", v)
			}
			assertBadLoc(t, bad.Err, bad.Loc, src, 0, len(src.Body))
		})
	}

	typeBodies := []string{"?", "01", "-", "\"unterminated"}
	for _, body := range typeBodies {
		t.Run("type/"+body, func(t *testing.T) {
			src := &ast.Source{Body: body, Name: "type-lexer-error.graphql"}
			typ, err := p.ParseType(src)
			assertParseErrors(t, err, 1)
			bad, ok := typ.(*ast.BadType)
			if !ok {
				t.Fatalf("type = %T; want *ast.BadType", typ)
			}
			assertBadLoc(t, bad.Err, bad.Loc, src, 0, len(src.Body))
		})
	}
}

func TestParserValueAndTypeRecoveryStartsAtFirstNonIgnoredLexerError(t *testing.T) {
	p := parser.New(parser.WithRecovery())
	body := "  ?"
	wantStart := 2

	t.Run("value", func(t *testing.T) {
		src := &ast.Source{Body: body, Name: "value-leading-lexer-error.graphql"}
		v, err := p.ParseValue(src)
		assertParseErrors(t, err, 1)
		bad, ok := v.(*ast.BadValue)
		if !ok {
			t.Fatalf("value = %T; want *ast.BadValue", v)
		}
		assertBadLoc(t, bad.Err, bad.Loc, src, wantStart, len(src.Body))
	})

	t.Run("const value", func(t *testing.T) {
		src := &ast.Source{Body: body, Name: "const-value-leading-lexer-error.graphql"}
		v, err := p.ParseConstValue(src)
		assertParseErrors(t, err, 1)
		bad, ok := v.(*ast.BadValue)
		if !ok {
			t.Fatalf("const value = %T; want *ast.BadValue", v)
		}
		assertBadLoc(t, bad.Err, bad.Loc, src, wantStart, len(src.Body))
	})

	t.Run("type", func(t *testing.T) {
		src := &ast.Source{Body: body, Name: "type-leading-lexer-error.graphql"}
		typ, err := p.ParseType(src)
		assertParseErrors(t, err, 1)
		bad, ok := typ.(*ast.BadType)
		if !ok {
			t.Fatalf("type = %T; want *ast.BadType", typ)
		}
		assertBadLoc(t, bad.Err, bad.Loc, src, wantStart, len(src.Body))
	})
}

func TestDocumentRecoveryConsumesInitialLexerErrorsOnce(t *testing.T) {
	p := parser.New(parser.WithRecovery())

	for _, body := range []string{"?", "01"} {
		t.Run("document/"+body, func(t *testing.T) {
			doc, err := p.Parse(&ast.Source{Body: body, Name: "doc-lexer-error.graphql"})
			if doc != nil {
				t.Fatalf("document = %T; want nil", doc)
			}
			assertParseErrors(t, err, 1)
		})

		t.Run("schema/"+body, func(t *testing.T) {
			doc, err := p.ParseSchema(&ast.Source{Body: body, Name: "schema-lexer-error.graphql"})
			if doc != nil {
				t.Fatalf("schema document = %T; want nil", doc)
			}
			assertParseErrors(t, err, 1)
		})
	}
}

func assertLocSource(t *testing.T, loc *ast.Loc, src *ast.Source) {
	t.Helper()
	if loc == nil {
		t.Fatal("Loc is nil")
	}
	if loc.Source != src {
		t.Fatalf("Loc.Source = %p; want %p", loc.Source, src)
	}
}

func assertParseErrors(t *testing.T, err error, wantLen int) parser.ParseErrors {
	t.Helper()
	if err == nil {
		t.Fatal("parse returned nil error")
	}
	var errs parser.ParseErrors
	if !errors.As(err, &errs) {
		t.Fatalf("error = %T; want parser.ParseErrors", err)
	}
	if len(errs) != wantLen {
		t.Fatalf("ParseErrors length = %d; want %d", len(errs), wantLen)
	}
	return errs
}

func assertBadLoc(t *testing.T, err *ast.SyntaxError, loc *ast.Loc, src *ast.Source, wantStart, wantEnd int) {
	t.Helper()
	if err == nil {
		t.Fatal("bad node Err is nil")
	}
	if err.Source != src {
		t.Fatalf("bad node Err.Source = %p; want %p", err.Source, src)
	}
	assertLocSource(t, loc, src)
	if loc.Start != wantStart || loc.End != wantEnd {
		t.Fatalf("bad node Loc = [%d, %d); want [%d, %d)", loc.Start, loc.End, wantStart, wantEnd)
	}
}
