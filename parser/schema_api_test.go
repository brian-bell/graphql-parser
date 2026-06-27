package parser_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/parser"
)

func TestParseSchema_AcceptsSchemaDefinitionsAndExtensions(t *testing.T) {
	body := `
		type Query { name: String }
		extend type Query { age: Int }
	`

	doc, err := parser.ParseSchema(body)
	if err != nil {
		t.Fatalf("ParseSchema returned error: %v", err)
	}
	if len(doc.Definitions) != 2 {
		t.Fatalf("definitions = %d; want 2", len(doc.Definitions))
	}
	if _, ok := doc.Definitions[0].(*ast.ObjectTypeDefinition); !ok {
		t.Errorf("definition[0] = %T; want *ast.ObjectTypeDefinition", doc.Definitions[0])
	}
	if _, ok := doc.Definitions[1].(*ast.ObjectTypeExtension); !ok {
		t.Errorf("definition[1] = %T; want *ast.ObjectTypeExtension", doc.Definitions[1])
	}
}

func TestParseSchema_RejectsExecutableDefinitions(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{name: "named query", body: `query MyQuery { hello }`},
		{name: "named mutation", body: `mutation Update { update }`},
		{name: "named subscription", body: `subscription Events { events }`},
		{name: "shorthand query", body: `{ hello }`},
		{name: "fragment", body: `fragment UserFields on User { id }`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parser.Parse(tc.body); err != nil {
				t.Fatalf("Parse returned unexpected error: %v", err)
			}

			doc, err := parser.ParseSchema(tc.body)
			if err == nil {
				t.Fatal("ParseSchema returned nil error")
			}
			if doc != nil {
				t.Fatalf("ParseSchema document = %v; want nil", doc)
			}
			var parseErr *parser.ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("error = %T; want *parser.ParseError", err)
			}
			msg := strings.ToLower(err.Error())
			if !strings.Contains(msg, "schema") || !strings.Contains(msg, "executable") {
				t.Errorf("error %q should mention schema and executable definitions", err.Error())
			}
		})
	}
}

func TestParseSchema_RejectsMixedSchemaAndExecutableDocument(t *testing.T) {
	body := `type Query { x: String } query Q { x }`
	if _, err := parser.Parse(body); err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}

	doc, err := parser.ParseSchema(body)
	if err == nil {
		t.Fatal("ParseSchema returned nil error")
	}
	if doc != nil {
		t.Fatalf("ParseSchema document = %v; want nil", doc)
	}
}

func TestParseSchema_AcceptsUnknownDirectiveUses(t *testing.T) {
	doc, err := parser.ParseSchema(`type User @custom { id: ID @fieldDirective }`)
	if err != nil {
		t.Fatalf("ParseSchema returned error: %v", err)
	}

	def := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	if def.Directives.ForName("custom") == nil {
		t.Fatal("missing @custom directive on type")
	}
	field := def.Fields.ForName("id")
	if field == nil {
		t.Fatal("missing id field")
	}
	if field.Directives.ForName("fieldDirective") == nil {
		t.Fatal("missing @fieldDirective directive on field")
	}
}

func TestParseSchemaSource_RejectionUsesProvidedSourceMetadata(t *testing.T) {
	src := &ast.Source{
		Body:           "\nquery Q { x }",
		Name:           "schema.graphql",
		LocationOffset: ast.Position{Line: 10, Column: 5},
	}

	_, err := parser.ParseSchemaSource(src)
	if err == nil {
		t.Fatal("ParseSchemaSource returned nil error")
	}
	var parseErr *parser.ParseError
	if !errors.As(err, &parseErr) {
		t.Fatalf("error = %T; want *parser.ParseError", err)
	}
	if parseErr.Source != src {
		t.Fatalf("error source = %v; want provided source", parseErr.Source)
	}
	if pos := parseErr.Position(); pos.Line != 11 || pos.Column != 1 {
		t.Fatalf("error position = %s; want 11:1", pos)
	}
	if !strings.Contains(err.Error(), "schema.graphql:11:1") {
		t.Fatalf("error %q does not include provided source position", err.Error())
	}
}

func TestParseSchema_WithCommentsPreservesCommentAttachment(t *testing.T) {
	body := `
# query type
type Query {
	# field comment
	name: String
}`
	doc, err := parser.ParseSchema(body, parser.WithComments())
	if err != nil {
		t.Fatalf("ParseSchema returned error: %v", err)
	}

	def := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	if def.Comments == nil || len(def.Comments.Leading) != 1 {
		t.Fatalf("definition comments = %+v; want one leading comment", def.Comments)
	}
	if def.Comments.Leading[0].Text != " query type" {
		t.Fatalf("definition comment = %q; want %q", def.Comments.Leading[0].Text, " query type")
	}
	field := def.Fields.ForName("name")
	if field.Comments == nil || len(field.Comments.Leading) != 1 {
		t.Fatalf("field comments = %+v; want one leading comment", field.Comments)
	}
	if field.Comments.Leading[0].Text != " field comment" {
		t.Fatalf("field comment = %q; want %q", field.Comments.Leading[0].Text, " field comment")
	}
}

func TestParseSchema_WithRecoveryAggregatesSchemaOnlyErrors(t *testing.T) {
	src := &ast.Source{
		Body: `type Query { x: String } query Q { x } fragment F on Query { x }`,
		Name: "schema.graphql",
	}

	doc, err := parser.ParseSchemaSource(src, parser.WithRecovery())
	if err == nil {
		t.Fatal("ParseSchemaSource returned nil error")
	}
	if doc == nil {
		t.Fatal("ParseSchemaSource returned nil document; want recovered document")
	}
	var parseErrs parser.ParseErrors
	if !errors.As(err, &parseErrs) {
		t.Fatalf("error = %T; want parser.ParseErrors", err)
	}
	if len(parseErrs) != 2 {
		t.Fatalf("errors = %d; want 2", len(parseErrs))
	}
	if len(doc.Definitions) != 3 {
		t.Fatalf("definitions = %d; want 3", len(doc.Definitions))
	}
	wantPositions := []ast.Position{
		{Line: 1, Column: 26},
		{Line: 1, Column: 40},
	}
	for i, parseErr := range parseErrs {
		if parseErr.Source != src {
			t.Fatalf("error[%d] source = %v; want provided source", i, parseErr.Source)
		}
		if pos := parseErr.Position(); pos != wantPositions[i] {
			t.Fatalf("error[%d] position = %s; want %s", i, pos, wantPositions[i])
		}
	}
}
