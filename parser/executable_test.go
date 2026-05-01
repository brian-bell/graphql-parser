package parser_test

import (
	"strings"
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/parser"
)

func mustParse(t *testing.T, body string) *ast.Document {
	t.Helper()
	doc, err := parser.Parse(body)
	if err != nil {
		t.Fatalf("Parse(%q): %v", body, err)
	}
	return doc
}

func TestParse_ShorthandQuery(t *testing.T) {
	doc := mustParse(t, "{ hello }")
	if len(doc.Definitions) != 1 {
		t.Fatalf("definitions = %d; want 1", len(doc.Definitions))
	}
	op, ok := doc.Definitions[0].(*ast.OperationDefinition)
	if !ok {
		t.Fatalf("got %T; want *ast.OperationDefinition", doc.Definitions[0])
	}
	if op.Operation != ast.OperationQuery {
		t.Errorf("op = %q; want query", op.Operation)
	}
	if op.Name != "" {
		t.Errorf("Name = %q; want empty", op.Name)
	}
	if op.SelectionSet == nil || len(op.SelectionSet.Selections) != 1 {
		t.Fatalf("selection set wrong: %+v", op.SelectionSet)
	}
	f, ok := op.SelectionSet.Selections[0].(*ast.Field)
	if !ok {
		t.Fatalf("got %T; want *ast.Field", op.SelectionSet.Selections[0])
	}
	if f.Name != "hello" {
		t.Errorf("Name = %q; want hello", f.Name)
	}
}

func TestParse_NamedQuery(t *testing.T) {
	doc := mustParse(t, "query MyQuery { hello }")
	op := doc.Definitions[0].(*ast.OperationDefinition)
	if op.Operation != ast.OperationQuery {
		t.Errorf("op = %q; want query", op.Operation)
	}
	if op.Name != "MyQuery" {
		t.Errorf("Name = %q; want MyQuery", op.Name)
	}
}

func TestParse_Mutation(t *testing.T) {
	doc := mustParse(t, "mutation { update }")
	op := doc.Definitions[0].(*ast.OperationDefinition)
	if op.Operation != ast.OperationMutation {
		t.Errorf("op = %q; want mutation", op.Operation)
	}
}

func TestParse_Subscription(t *testing.T) {
	doc := mustParse(t, "subscription Sub { events }")
	op := doc.Definitions[0].(*ast.OperationDefinition)
	if op.Operation != ast.OperationSubscription {
		t.Errorf("op = %q; want subscription", op.Operation)
	}
	if op.Name != "Sub" {
		t.Errorf("Name = %q; want Sub", op.Name)
	}
}

func TestParse_VariablesAndDefaults(t *testing.T) {
	body := `query Q($id: ID!, $count: Int = 10, $tags: [String!]! = ["a","b"]) { x }`
	doc := mustParse(t, body)
	op := doc.Definitions[0].(*ast.OperationDefinition)
	if len(op.VariableDefinitions) != 3 {
		t.Fatalf("vars = %d; want 3", len(op.VariableDefinitions))
	}
	id := op.VariableDefinitions.ForName("id")
	if id == nil {
		t.Fatal("missing $id")
	}
	if _, ok := id.Type.(*ast.NonNullType); !ok {
		t.Errorf("$id type = %T; want *ast.NonNullType", id.Type)
	}
	if id.DefaultValue != nil {
		t.Error("$id has unexpected default")
	}
	count := op.VariableDefinitions.ForName("count")
	if count == nil || count.DefaultValue == nil {
		t.Fatal("missing $count or its default")
	}
	if iv, ok := count.DefaultValue.(*ast.IntValue); !ok || iv.Value != "10" {
		t.Errorf("count default = %T %v; want IntValue 10", count.DefaultValue, count.DefaultValue)
	}
}

func TestParse_VariableDefaultRejectsVariable(t *testing.T) {
	// Default values are const, so $other is not allowed.
	if _, err := parser.Parse(`query ($a: Int = $other) { x }`); err == nil {
		t.Error("expected error: variable in default value")
	}
}

func TestParse_Directives(t *testing.T) {
	doc := mustParse(t, `query @deprecated(reason: "old") @cacheControl { x @include(if: true) }`)
	op := doc.Definitions[0].(*ast.OperationDefinition)
	if len(op.Directives) != 2 {
		t.Fatalf("op directives = %d; want 2", len(op.Directives))
	}
	if op.Directives.ForName("deprecated") == nil {
		t.Error("missing @deprecated")
	}
	if op.Directives.ForName("cacheControl") == nil {
		t.Error("missing @cacheControl")
	}
	dep := op.Directives.ForName("deprecated")
	if len(dep.Arguments) != 1 || dep.Arguments[0].Name != "reason" {
		t.Errorf("deprecated args = %v", dep.Arguments)
	}

	f := op.SelectionSet.Selections[0].(*ast.Field)
	inc := f.Directives.ForName("include")
	if inc == nil {
		t.Fatal("missing @include on field")
	}
	if iv := inc.Arguments.ForName("if"); iv == nil {
		t.Fatal("missing if: arg")
	}
}

func TestParse_FieldAliasAndArguments(t *testing.T) {
	doc := mustParse(t, `{ short: longFieldName(a: 1, b: $v) }`)
	op := doc.Definitions[0].(*ast.OperationDefinition)
	f := op.SelectionSet.Selections[0].(*ast.Field)
	if f.Alias != "short" {
		t.Errorf("alias = %q; want short", f.Alias)
	}
	if f.Name != "longFieldName" {
		t.Errorf("name = %q; want longFieldName", f.Name)
	}
	if len(f.Arguments) != 2 {
		t.Errorf("args = %d; want 2", len(f.Arguments))
	}
}

func TestParse_NestedSelectionSets(t *testing.T) {
	doc := mustParse(t, `{ user { name address { city } } }`)
	op := doc.Definitions[0].(*ast.OperationDefinition)
	user := op.SelectionSet.Selections[0].(*ast.Field)
	if user.SelectionSet == nil {
		t.Fatal("user has no nested selection set")
	}
	if len(user.SelectionSet.Selections) != 2 {
		t.Errorf("user children = %d; want 2", len(user.SelectionSet.Selections))
	}
	address := user.SelectionSet.Selections[1].(*ast.Field)
	if address.SelectionSet == nil || len(address.SelectionSet.Selections) != 1 {
		t.Errorf("address selection set wrong: %+v", address.SelectionSet)
	}
}

func TestParse_FragmentSpread(t *testing.T) {
	doc := mustParse(t, `{ ...UserFields ...OtherFields @skip(if: false) }`)
	op := doc.Definitions[0].(*ast.OperationDefinition)
	if len(op.SelectionSet.Selections) != 2 {
		t.Fatalf("selections = %d; want 2", len(op.SelectionSet.Selections))
	}
	a, ok := op.SelectionSet.Selections[0].(*ast.FragmentSpread)
	if !ok || a.Name != "UserFields" {
		t.Errorf("first: got %T %v; want FragmentSpread UserFields", op.SelectionSet.Selections[0], op.SelectionSet.Selections[0])
	}
	b := op.SelectionSet.Selections[1].(*ast.FragmentSpread)
	if b.Name != "OtherFields" || len(b.Directives) != 1 {
		t.Errorf("second wrong: %+v", b)
	}
}

func TestParse_InlineFragmentWithCondition(t *testing.T) {
	doc := mustParse(t, `{ ... on User { name } }`)
	op := doc.Definitions[0].(*ast.OperationDefinition)
	frag, ok := op.SelectionSet.Selections[0].(*ast.InlineFragment)
	if !ok {
		t.Fatalf("got %T; want *ast.InlineFragment", op.SelectionSet.Selections[0])
	}
	if frag.TypeCondition == nil || frag.TypeCondition.Name != "User" {
		t.Errorf("type condition = %v; want NamedType User", frag.TypeCondition)
	}
}

func TestParse_InlineFragmentWithoutCondition(t *testing.T) {
	doc := mustParse(t, `{ ... @skip(if: true) { name } }`)
	op := doc.Definitions[0].(*ast.OperationDefinition)
	frag := op.SelectionSet.Selections[0].(*ast.InlineFragment)
	if frag.TypeCondition != nil {
		t.Errorf("type condition = %v; want nil", frag.TypeCondition)
	}
	if len(frag.Directives) != 1 {
		t.Errorf("directives = %d; want 1", len(frag.Directives))
	}
}

func TestParse_FragmentDefinition(t *testing.T) {
	doc := mustParse(t, `fragment UserFields on User { id name }`)
	frag := doc.Definitions[0].(*ast.FragmentDefinition)
	if frag.Name != "UserFields" {
		t.Errorf("name = %q; want UserFields", frag.Name)
	}
	if frag.TypeCondition.Name != "User" {
		t.Errorf("cond = %v; want User", frag.TypeCondition)
	}
	if len(frag.SelectionSet.Selections) != 2 {
		t.Errorf("selections = %d; want 2", len(frag.SelectionSet.Selections))
	}
}

func TestParse_FragmentDefinitionRejectsOnAsName(t *testing.T) {
	if _, err := parser.Parse(`fragment on on Foo { x }`); err == nil {
		t.Error("expected error: fragment named 'on'")
	}
}

func TestParse_MultipleDefinitions(t *testing.T) {
	body := `query A { a } query B { b } fragment F on T { c }`
	doc := mustParse(t, body)
	if len(doc.Definitions) != 3 {
		t.Errorf("defs = %d; want 3", len(doc.Definitions))
	}
}

func TestParse_EmptyDocumentIsError(t *testing.T) {
	if _, err := parser.Parse(""); err == nil {
		t.Error("expected error for empty document")
	}
}

func TestParse_EmptySelectionSetIsError(t *testing.T) {
	if _, err := parser.Parse("{}"); err == nil {
		t.Error("expected error: empty selection set")
	}
}

func TestParse_EmptyArgumentsIsError(t *testing.T) {
	if _, err := parser.Parse("{ field() }"); err == nil {
		t.Error("expected error: empty arguments")
	}
}

func TestParse_EmptyVariableDefinitionsIsError(t *testing.T) {
	if _, err := parser.Parse("query () { x }"); err == nil {
		t.Error("expected error: empty variable definitions")
	}
}

func TestParse_DocumentLocCoversFullSource(t *testing.T) {
	body := "  query Q { x }  "
	doc := mustParse(t, body)
	loc := doc.GetLoc()
	if loc == nil {
		t.Fatal("nil Loc")
	}
	if loc.Start != 2 {
		t.Errorf("Start = %d; want 2 (skip leading whitespace)", loc.Start)
	}
	// End should land at the closing brace, not the trailing whitespace.
	if loc.End != 15 {
		t.Errorf("End = %d; want 15", loc.End)
	}
}

func TestParse_ErrorMentionsExpectedToken(t *testing.T) {
	_, err := parser.Parse("query")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "expected") {
		t.Errorf("error %q does not say 'expected'", err.Error())
	}
}

func TestParseSource_UsesProvidedName(t *testing.T) {
	src := &ast.Source{Body: "{ x }", Name: "myfile.graphql"}
	doc, err := parser.ParseSource(src)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Loc.Source.Name != "myfile.graphql" {
		t.Errorf("source name = %q; want myfile.graphql", doc.Loc.Source.Name)
	}
}
