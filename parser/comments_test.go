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

func TestComments_DefinitionVsFirstFieldAttribution(t *testing.T) {
	body := `
# the type
type T {
	# the first field
	id: ID!
}`
	doc, err := parser.Parse(body, parser.WithComments())
	if err != nil {
		t.Fatal(err)
	}
	def := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	if def.Comments == nil || len(def.Comments.Leading) != 1 || def.Comments.Leading[0].Text != " the type" {
		t.Fatalf("type leading comments wrong: %+v", def.Comments)
	}
	id := def.Fields.ForName("id")
	if id.Comments == nil || len(id.Comments.Leading) != 1 || id.Comments.Leading[0].Text != " the first field" {
		t.Fatalf("field leading comments wrong: %+v", id.Comments)
	}
	// The type's comment must not leak onto the field, nor the field's onto the
	// type — pins the capture-at-entry timing.
	for _, c := range def.Comments.Leading {
		if c.Text == " the first field" {
			t.Errorf("field comment leaked onto the type")
		}
	}
	for _, c := range id.Comments.Leading {
		if c.Text == " the type" {
			t.Errorf("type comment leaked onto the field")
		}
	}
}

func TestComments_WithRecovery_NoBleedAcrossBadDefinition(t *testing.T) {
	body := `
# bad one
type 123
# good one
type Good { x: Int }`
	doc, _ := parser.Parse(body, parser.WithComments(), parser.WithRecovery())
	if doc == nil {
		t.Fatal("expected a document even with a malformed definition")
	}
	var good *ast.ObjectTypeDefinition
	for _, d := range doc.Definitions {
		if obj, ok := d.(*ast.ObjectTypeDefinition); ok && obj.Name == "Good" {
			good = obj
		}
	}
	if good == nil {
		t.Fatal("surviving Good definition not found")
	}
	if good.Comments == nil || len(good.Comments.Leading) != 1 || good.Comments.Leading[0].Text != " good one" {
		t.Errorf("surviving definition lost/leaked comments: %+v", good.Comments)
	}
}

func TestComments_NestedInputValueAttribution(t *testing.T) {
	body := `
type T {
	# the field itself
	field(
		# the arg
		arg: Int
	): Boolean
}`
	doc, err := parser.Parse(body, parser.WithComments())
	if err != nil {
		t.Fatal(err)
	}
	def := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	f := def.Fields.ForName("field")
	if f.Comments == nil || len(f.Comments.Leading) != 1 || f.Comments.Leading[0].Text != " the field itself" {
		t.Fatalf("field comments wrong: %+v", f.Comments)
	}
	arg := f.Arguments.ForName("arg")
	if arg.Comments == nil || len(arg.Comments.Leading) != 1 || arg.Comments.Leading[0].Text != " the arg" {
		t.Fatalf("arg comments wrong: %+v", arg.Comments)
	}
	// The arg comment must attach to the InputValueDefinition, not the field.
	for _, c := range f.Comments.Leading {
		if c.Text == " the arg" {
			t.Errorf("arg comment leaked onto the field")
		}
	}
}

func TestComments_LatentNodesStayNil(t *testing.T) {
	body := `
type T {
	# the field
	field(
		# the arg
		arg: Int = 42
	): Boolean
}`
	doc, err := parser.Parse(body, parser.WithComments())
	if err != nil {
		t.Fatal(err)
	}
	def := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	arg := def.Fields.ForName("field").Arguments.ForName("arg")
	// The default value is an IntValue — a latent node the parser does not bind
	// comments to. Pin that scope did not widen.
	if iv, ok := arg.DefaultValue.(*ast.IntValue); ok {
		if iv.Comments != nil {
			t.Errorf("IntValue default unexpectedly carries comments: %+v", iv.Comments)
		}
	} else {
		t.Fatalf("default value is %T; want *ast.IntValue", arg.DefaultValue)
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
