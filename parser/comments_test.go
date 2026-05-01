package parser_test

import (
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/parser"
)

func TestComments_OffByDefault(t *testing.T) {
	body := `
# leading comment
type T {
	# field comment
	name: String
}`
	doc, err := parser.Parse(body)
	if err != nil {
		t.Fatal(err)
	}
	def := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	if def.Comments != nil {
		t.Errorf("Comments should be nil when WithComments off; got %+v", def.Comments)
	}
	field := def.Fields.ForName("name")
	if field.Comments != nil {
		t.Errorf("field Comments should be nil; got %+v", field.Comments)
	}
}

func TestComments_LeadingOnDefinition(t *testing.T) {
	body := `
# this is the type
# with a multi-line comment
type T { x: Int }`
	doc, err := parser.Parse(body, parser.WithComments())
	if err != nil {
		t.Fatal(err)
	}
	def := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	if def.Comments == nil || len(def.Comments.Leading) != 2 {
		t.Fatalf("expected 2 leading comments; got %+v", def.Comments)
	}
	if def.Comments.Leading[0].Text != " this is the type" {
		t.Errorf("leading[0].Text = %q", def.Comments.Leading[0].Text)
	}
	if def.Comments.Leading[1].Text != " with a multi-line comment" {
		t.Errorf("leading[1].Text = %q", def.Comments.Leading[1].Text)
	}
}

func TestComments_LeadingOnField(t *testing.T) {
	body := `
type T {
	# the id
	id: ID!
	# the name
	# (display only)
	name: String
}`
	doc, err := parser.Parse(body, parser.WithComments())
	if err != nil {
		t.Fatal(err)
	}
	def := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	id := def.Fields.ForName("id")
	if id.Comments == nil || len(id.Comments.Leading) != 1 || id.Comments.Leading[0].Text != " the id" {
		t.Errorf("id field leading comments wrong: %+v", id.Comments)
	}
	name := def.Fields.ForName("name")
	if name.Comments == nil || len(name.Comments.Leading) != 2 {
		t.Errorf("name field leading comments wrong: %+v", name.Comments)
	}
}

func TestComments_LeadingOnEnumValue(t *testing.T) {
	body := `
enum Color {
	# red
	RED
	# green
	GREEN
}`
	doc, err := parser.Parse(body, parser.WithComments())
	if err != nil {
		t.Fatal(err)
	}
	enum := doc.Definitions[0].(*ast.EnumTypeDefinition)
	red := enum.Values.ForName("RED")
	if red.Comments == nil || len(red.Comments.Leading) != 1 {
		t.Errorf("RED comments: %+v", red.Comments)
	}
}

func TestComments_LeadingOnArgumentDefinition(t *testing.T) {
	body := `
type T {
	field(
		# the first arg
		first: Int
		# the second arg
		second: String
	): Boolean
}`
	doc, err := parser.Parse(body, parser.WithComments())
	if err != nil {
		t.Fatal(err)
	}
	def := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	f := def.Fields.ForName("field")
	first := f.Arguments.ForName("first")
	if first.Comments == nil || len(first.Comments.Leading) != 1 {
		t.Errorf("first arg comments: %+v", first.Comments)
	}
}

func TestComments_OffInvariance_ASTUnchanged(t *testing.T) {
	body := `
# comment
type T {
	# comment
	x: Int
}`
	docOff, err := parser.Parse(body)
	if err != nil {
		t.Fatal(err)
	}
	docOn, err := parser.Parse(body, parser.WithComments())
	if err != nil {
		t.Fatal(err)
	}
	// Modulo Comments fields, the structure should be identical.
	defOff := docOff.Definitions[0].(*ast.ObjectTypeDefinition)
	defOn := docOn.Definitions[0].(*ast.ObjectTypeDefinition)
	if defOff.Name != defOn.Name {
		t.Errorf("Name differs")
	}
	if len(defOff.Fields) != len(defOn.Fields) {
		t.Errorf("Fields len differs: %d vs %d", len(defOff.Fields), len(defOn.Fields))
	}
	if defOff.Loc.Start != defOn.Loc.Start || defOff.Loc.End != defOn.Loc.End {
		t.Errorf("Loc differs: %+v vs %+v", defOff.Loc, defOn.Loc)
	}
}

func TestComments_NoComments_NoCommentGroup(t *testing.T) {
	body := `type T { x: Int }`
	doc, err := parser.Parse(body, parser.WithComments())
	if err != nil {
		t.Fatal(err)
	}
	def := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	if def.Comments != nil {
		t.Errorf("Comments should be nil when there are no comments; got %+v", def.Comments)
	}
}
