package ast_test

import (
	"reflect"
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/parser"
)

// recordVisitor records every node it visits.
type recordVisitor struct {
	visited []ast.Node
}

func (r *recordVisitor) Visit(n ast.Node) ast.Visitor {
	r.visited = append(r.visited, n)
	return r
}

func TestWalk_NilNode(t *testing.T) {
	r := &recordVisitor{}
	ast.Walk(r, nil)
	if len(r.visited) != 0 {
		t.Errorf("expected no visits, got %d", len(r.visited))
	}
}

func TestWalk_VisitsRoot(t *testing.T) {
	doc, err := parser.Parse("{ x }")
	if err != nil {
		t.Fatal(err)
	}
	r := &recordVisitor{}
	ast.Walk(r, doc)
	if len(r.visited) == 0 || r.visited[0] != ast.Node(doc) {
		t.Errorf("first visit should be the root document; got %v", r.visited)
	}
}

func TestWalk_StopOnNil(t *testing.T) {
	doc, err := parser.Parse("{ x { y } }")
	if err != nil {
		t.Fatal(err)
	}
	// Visit the root, return nil → no descent.
	stopped := &stoppingVisitor{}
	ast.Walk(stopped, doc)
	if stopped.count != 1 {
		t.Errorf("expected 1 visit, got %d", stopped.count)
	}
}

type stoppingVisitor struct{ count int }

func (s *stoppingVisitor) Visit(ast.Node) ast.Visitor {
	s.count++
	return nil
}

func TestWalk_DescendsThroughExecutable(t *testing.T) {
	body := `
		query Q($v: Int = 1) @dir {
			field(arg: 2) {
				...frag
				... on T @incl { y }
			}
		}
		fragment frag on T { z }
	`
	doc, err := parser.Parse(body)
	if err != nil {
		t.Fatal(err)
	}
	r := &recordVisitor{}
	ast.Walk(r, doc)

	expected := map[reflect.Type]bool{
		reflect.TypeOf(&ast.Document{}):            true,
		reflect.TypeOf(&ast.OperationDefinition{}): true,
		reflect.TypeOf(&ast.VariableDefinition{}):  true,
		reflect.TypeOf(&ast.Variable{}):            true,
		reflect.TypeOf(&ast.NamedType{}):           true,
		reflect.TypeOf(&ast.IntValue{}):            true,
		reflect.TypeOf(&ast.Directive{}):           true,
		reflect.TypeOf(&ast.SelectionSet{}):        true,
		reflect.TypeOf(&ast.Field{}):               true,
		reflect.TypeOf(&ast.Argument{}):            true,
		reflect.TypeOf(&ast.FragmentSpread{}):      true,
		reflect.TypeOf(&ast.InlineFragment{}):      true,
		reflect.TypeOf(&ast.FragmentDefinition{}):  true,
	}
	for ty := range expected {
		expected[ty] = false
	}
	for _, n := range r.visited {
		expected[reflect.TypeOf(n)] = true
	}
	for ty, seen := range expected {
		if !seen {
			t.Errorf("walk did not visit a node of type %v", ty)
		}
	}
}

func TestWalk_DescendsThroughSchema(t *testing.T) {
	body := `
		"the schema" schema { query: Q }
		"the type" type T implements I @x { f("arg desc" a: Int = 1 @arg): String }
		interface I { id: ID }
		union U = A | B
		enum E { A B }
		input In { f: Int }
		directive @x repeatable on FIELD
		extend type T { g: Int }
	`
	doc, err := parser.Parse(body)
	if err != nil {
		t.Fatal(err)
	}
	r := &recordVisitor{}
	ast.Walk(r, doc)

	want := []reflect.Type{
		reflect.TypeOf(&ast.SchemaDefinition{}),
		reflect.TypeOf(&ast.OperationTypeDefinition{}),
		reflect.TypeOf(&ast.ObjectTypeDefinition{}),
		reflect.TypeOf(&ast.FieldDefinition{}),
		reflect.TypeOf(&ast.InputValueDefinition{}),
		reflect.TypeOf(&ast.InterfaceTypeDefinition{}),
		reflect.TypeOf(&ast.UnionTypeDefinition{}),
		reflect.TypeOf(&ast.EnumTypeDefinition{}),
		reflect.TypeOf(&ast.EnumValueDefinition{}),
		reflect.TypeOf(&ast.InputObjectTypeDefinition{}),
		reflect.TypeOf(&ast.DirectiveDefinition{}),
		reflect.TypeOf(&ast.ObjectTypeExtension{}),
		reflect.TypeOf(&ast.StringValue{}), // descriptions
	}
	seen := map[reflect.Type]bool{}
	for _, n := range r.visited {
		seen[reflect.TypeOf(n)] = true
	}
	for _, ty := range want {
		if !seen[ty] {
			t.Errorf("walk did not visit a node of type %v", ty)
		}
	}
}

func TestInspect_StopOnFalse(t *testing.T) {
	doc, err := parser.Parse("{ a { b { c } } }")
	if err != nil {
		t.Fatal(err)
	}
	var depth int
	ast.Inspect(doc, func(n ast.Node) bool {
		if _, ok := n.(*ast.Field); ok {
			depth++
		}
		// Stop at the first Field to prevent descent into nested ones.
		if _, ok := n.(*ast.Field); ok {
			return false
		}
		return true
	})
	if depth != 1 {
		t.Errorf("expected to count exactly one Field; got %d", depth)
	}
}

func TestWalk_DescendsThroughValues(t *testing.T) {
	v, err := parser.ParseValue(`{a: [1, "hi", $v, ENUM, true, null], b: {nested: 1}}`)
	if err != nil {
		t.Fatal(err)
	}
	r := &recordVisitor{}
	ast.Walk(r, v)

	want := []reflect.Type{
		reflect.TypeOf(&ast.ObjectValue{}),
		reflect.TypeOf(&ast.ObjectField{}),
		reflect.TypeOf(&ast.ListValue{}),
		reflect.TypeOf(&ast.IntValue{}),
		reflect.TypeOf(&ast.StringValue{}),
		reflect.TypeOf(&ast.Variable{}),
		reflect.TypeOf(&ast.EnumValue{}),
		reflect.TypeOf(&ast.BooleanValue{}),
		reflect.TypeOf(&ast.NullValue{}),
	}
	seen := map[reflect.Type]bool{}
	for _, n := range r.visited {
		seen[reflect.TypeOf(n)] = true
	}
	for _, ty := range want {
		if !seen[ty] {
			t.Errorf("walk did not visit a node of type %v", ty)
		}
	}
}

func TestWalk_DescendsThroughTypes(t *testing.T) {
	t1, err := parser.ParseType("[Int!]!")
	if err != nil {
		t.Fatal(err)
	}
	r := &recordVisitor{}
	ast.Walk(r, t1)
	// NonNull → List → NonNull → Named
	want := []reflect.Type{
		reflect.TypeOf(&ast.NonNullType{}),
		reflect.TypeOf(&ast.ListType{}),
		reflect.TypeOf(&ast.NamedType{}),
	}
	seen := map[reflect.Type]bool{}
	for _, n := range r.visited {
		seen[reflect.TypeOf(n)] = true
	}
	for _, ty := range want {
		if !seen[ty] {
			t.Errorf("did not visit %v", ty)
		}
	}
}
