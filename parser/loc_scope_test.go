package parser_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/parser"
)

func TestLocContainment_ExecutableDocument(t *testing.T) {
	body := `query GetUser($id: ID! = "1" @varDir, $opts: [Input!] = [{flag: true, size: 3}]) @opDir(enabled: true) {
  me: user(id: $id, opts: {tags: ["a", null, ADMIN]}) @fieldDir {
    id
    ...UserFields @spreadDir
    ... on Admin @inlineDir { permissions }
    ... @anonDir { fallback }
  }
}

fragment UserFields on User @fragDir { name }`

	doc := mustParse(t, body)
	assertLocTree(t, body, doc)
}

func TestLocContainment_SchemaDocument(t *testing.T) {
	body := `"schema desc"
schema @schemaDir { query: Query mutation: Mutation subscription: Subscription }

"scalar desc"
scalar DateTime @specifiedBy(url: "https://example.com/date")

"type desc"
type Query implements & Node @obj {
  "field desc"
  node(
    "arg desc"
    id: ID! = "x" @arg
  ): Node @field
}

interface Node { id: ID! }
type Mutation { noop: Boolean }
type Subscription { tick: Int }
union Search = Query | Mutation
enum Color { "red desc" RED @rgb GREEN }
input Filter { "term desc" term: String = "all" @trim }
"directive desc"
directive @example(
  "input desc"
  value: String = "x"
) repeatable on FIELD | QUERY
extend schema @extra { query: Query }
extend scalar DateTime @extra
extend type Query implements Node @extra { extra: String }
extend interface Node @extra { other: String }
extend union Search @extra = Subscription
extend enum Color @extra { BLUE }
extend input Filter @extra { limit: Int }`

	doc, err := parser.ParseSchema(body)
	if err != nil {
		t.Fatal(err)
	}
	assertLocTree(t, body, doc)
}

func TestLocContainment_PartialParseRoots(t *testing.T) {
	valueBody := `{input: [{name: "Ada", ok: true}, null], enum: ADMIN}`
	value, err := parser.ParseValue(valueBody)
	if err != nil {
		t.Fatal(err)
	}
	assertLocTree(t, valueBody, value)

	constBody := `[1, {field: "value"}]`
	constValue, err := parser.ParseConstValue(constBody)
	if err != nil {
		t.Fatal(err)
	}
	assertLocTree(t, constBody, constValue)

	typeBody := `[Int!]!`
	typ, err := parser.ParseType(typeBody)
	if err != nil {
		t.Fatal(err)
	}
	assertLocTree(t, typeBody, typ)
}

func TestLocWithComments_TopLevelCommentDoesNotWidenDefinition(t *testing.T) {
	body := `
# type comment
type Query { x: Int }`
	doc, err := parser.Parse(body, parser.WithComments())
	if err != nil {
		t.Fatal(err)
	}

	def := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	assertLocText(t, body, def, `type Query { x: Int }`)
	if def.Comments == nil || len(def.Comments.Leading) != 1 {
		t.Fatalf("definition comments = %+v; want one leading comment", def.Comments)
	}
	assertLocText(t, body, def.Comments.Leading[0], `# type comment`)
	assertLocTree(t, body, doc)
}

func TestLocWithComments_MemberCommentDoesNotWidenParentOrChild(t *testing.T) {
	body := `type Query {
  # field comment
  x: Int
}`
	doc, err := parser.Parse(body, parser.WithComments())
	if err != nil {
		t.Fatal(err)
	}

	def := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	field := def.Fields.ForName("x")
	assertLocText(t, body, def, body)
	assertLocText(t, body, field, `x: Int`)
	if def.Comments != nil {
		t.Fatalf("definition comments = %+v; want nil", def.Comments)
	}
	if field.Comments == nil || len(field.Comments.Leading) != 1 {
		t.Fatalf("field comments = %+v; want one leading comment", field.Comments)
	}
	assertLocText(t, body, field.Comments.Leading[0], `# field comment`)
	assertLocTree(t, body, doc)
}

func TestLocWithComments_DescriptionsAndCommentsStayOnIntendedNodes(t *testing.T) {
	body := `# type comment
"Query description"
type Query {
  # field comment
  "Field description"
  name: String
}`
	doc, err := parser.ParseSchema(body, parser.WithComments())
	if err != nil {
		t.Fatal(err)
	}

	def := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	field := def.Fields.ForName("name")
	assertLocText(t, body, def, `"Query description"
type Query {
  # field comment
  "Field description"
  name: String
}`)
	assertLocText(t, body, def.Description, `"Query description"`)
	assertLocText(t, body, field, `"Field description"
  name: String`)
	assertLocText(t, body, field.Description, `"Field description"`)
	if def.Comments == nil || len(def.Comments.Leading) != 1 {
		t.Fatalf("definition comments = %+v; want one leading comment", def.Comments)
	}
	if field.Comments == nil || len(field.Comments.Leading) != 1 {
		t.Fatalf("field comments = %+v; want one leading comment", field.Comments)
	}
	if def.Comments.Leading[0].Text != " type comment" {
		t.Fatalf("definition comment = %q; want %q", def.Comments.Leading[0].Text, " type comment")
	}
	if field.Comments.Leading[0].Text != " field comment" {
		t.Fatalf("field comment = %q; want %q", field.Comments.Leading[0].Text, " field comment")
	}
	assertLocTree(t, body, doc)
}

func TestLocWithRecovery_BadDefinitionCoversSkippedRegion(t *testing.T) {
	body := `bogus_keyword stuff
query Q { ok }`
	doc, err := parser.Parse(body, parser.WithRecovery())
	var parseErrs parser.ParseErrors
	if !errors.As(err, &parseErrs) {
		t.Fatalf("err = %T; want parser.ParseErrors", err)
	}
	if len(parseErrs) == 0 {
		t.Fatal("expected at least one parse error")
	}
	if len(doc.Definitions) != 2 {
		t.Fatalf("definitions = %d; want 2", len(doc.Definitions))
	}
	bad := doc.Definitions[0].(*ast.BadDefinition)
	assertLocText(t, body, bad, `bogus_keyword stuff`)
	assertLocTree(t, body, doc)
}

func TestLocWithRecovery_BadFieldCoversSkippedRegion(t *testing.T) {
	body := `{ a $bogus b }`
	doc, err := parser.Parse(body, parser.WithRecovery())
	var parseErrs parser.ParseErrors
	if !errors.As(err, &parseErrs) {
		t.Fatalf("err = %T; want parser.ParseErrors", err)
	}

	op := doc.Definitions[0].(*ast.OperationDefinition)
	var bad *ast.BadField
	for _, sel := range op.SelectionSet.Selections {
		if b, ok := sel.(*ast.BadField); ok {
			bad = b
			break
		}
	}
	if bad == nil {
		t.Fatal("BadField not found")
	}
	assertLocText(t, body, bad, `$`)
	assertLocTree(t, body, doc)
}

func TestLocExactSpans_ExecutableEdges(t *testing.T) {
	operationBody := `query Q($id: ID! = "1" @vd) @op { user: node(id: $id, filter: {tags: ["a", ADMIN]}) @field { ...Frag @spread ... on User @inline { name } ... @anon { id } } }`
	body := operationBody + ` fragment Frag on User { id }`
	doc := mustParse(t, body)

	op := doc.Definitions[0].(*ast.OperationDefinition)
	fragDef := doc.Definitions[1].(*ast.FragmentDefinition)
	field := op.SelectionSet.Selections[0].(*ast.Field)
	filterArg := field.Arguments.ForName("filter")
	objectValue := filterArg.Value.(*ast.ObjectValue)
	objectField := objectValue.Fields[0]
	listValue := objectField.Value.(*ast.ListValue)
	nestedSet := field.SelectionSet
	fragmentSpread := nestedSet.Selections[0].(*ast.FragmentSpread)
	inlineWithType := nestedSet.Selections[1].(*ast.InlineFragment)
	inlineWithoutType := nestedSet.Selections[2].(*ast.InlineFragment)

	assertLocText(t, body, op, operationBody)
	assertLocText(t, body, op.VariableDefinitions[0], `$id: ID! = "1" @vd`)
	assertLocText(t, body, op.VariableDefinitions[0].Type, `ID!`)
	assertLocText(t, body, field, `user: node(id: $id, filter: {tags: ["a", ADMIN]}) @field { ...Frag @spread ... on User @inline { name } ... @anon { id } }`)
	assertLocText(t, body, filterArg, `filter: {tags: ["a", ADMIN]}`)
	assertLocText(t, body, objectValue, `{tags: ["a", ADMIN]}`)
	assertLocText(t, body, objectField, `tags: ["a", ADMIN]`)
	assertLocText(t, body, listValue, `["a", ADMIN]`)
	assertLocText(t, body, fragmentSpread, `...Frag @spread`)
	assertLocText(t, body, inlineWithType, `... on User @inline { name }`)
	assertLocText(t, body, inlineWithoutType, `... @anon { id }`)
	assertLocText(t, body, fragDef, `fragment Frag on User { id }`)

	typ, err := parser.ParseType(`[Int!]!`)
	if err != nil {
		t.Fatal(err)
	}
	assertLocText(t, `[Int!]!`, typ, `[Int!]!`)
}

func TestLocExactSpans_SchemaEdges(t *testing.T) {
	body := `"type desc"
type Query @obj {
  "field desc"
  node(
    "arg desc"
    id: ID! = "x" @arg
  ): [Node!]! @field
}
"directive desc"
directive @example(
  "input desc"
  value: String = "x"
) repeatable on FIELD | QUERY
extend type Query @extra { extra: String }`

	doc, err := parser.ParseSchema(body)
	if err != nil {
		t.Fatal(err)
	}

	objectDef := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	field := objectDef.Fields.ForName("node")
	arg := field.Arguments.ForName("id")
	directiveDef := doc.Definitions[1].(*ast.DirectiveDefinition)
	extension := doc.Definitions[2].(*ast.ObjectTypeExtension)

	assertLocText(t, body, objectDef, `"type desc"
type Query @obj {
  "field desc"
  node(
    "arg desc"
    id: ID! = "x" @arg
  ): [Node!]! @field
}`)
	assertLocText(t, body, field, `"field desc"
  node(
    "arg desc"
    id: ID! = "x" @arg
  ): [Node!]! @field`)
	assertLocText(t, body, arg, `"arg desc"
    id: ID! = "x" @arg`)
	assertLocText(t, body, arg.Type, `ID!`)
	assertLocText(t, body, field.Type, `[Node!]!`)
	assertLocText(t, body, directiveDef, `"directive desc"
directive @example(
  "input desc"
  value: String = "x"
) repeatable on FIELD | QUERY`)
	assertLocText(t, body, directiveDef.Arguments.ForName("value"), `"input desc"
  value: String = "x"`)
	assertLocText(t, body, extension, `extend type Query @extra { extra: String }`)
	assertLocText(t, body, extension.Fields.ForName("extra"), `extra: String`)
}

func assertLocTree(t *testing.T, body string, root ast.Node) {
	t.Helper()
	assertLocTreeUnder(t, body, nil, root)
}

func assertLocTreeUnder(t *testing.T, body string, parent ast.Node, node ast.Node) {
	t.Helper()
	if isNilNode(node) {
		return
	}
	loc := node.GetLoc()
	if loc == nil {
		t.Fatalf("%T has nil Loc", node)
	}
	if loc.Source == nil || loc.Source.Body != body {
		t.Fatalf("%T Loc.Source = %+v; want source for test body", node, loc.Source)
	}
	if loc.Start < 0 || loc.Start > loc.End || loc.End > len(body) {
		t.Fatalf("%T Loc = [%d, %d), outside body len %d", node, loc.Start, loc.End, len(body))
	}
	if parent != nil {
		parentLoc := parent.GetLoc()
		if loc.Start < parentLoc.Start || loc.End > parentLoc.End {
			t.Fatalf("%T Loc [%d, %d) not contained in %T Loc [%d, %d)", node, loc.Start, loc.End, parent, parentLoc.Start, parentLoc.End)
		}
	}
	for _, child := range childrenOf(t, node) {
		assertLocTreeUnder(t, body, node, child)
	}
}

func childrenOf(t *testing.T, node ast.Node) []ast.Node {
	t.Helper()
	var children []ast.Node
	add := func(nodes ...ast.Node) {
		for _, n := range nodes {
			if !isNilNode(n) {
				children = append(children, n)
			}
		}
	}

	switch n := node.(type) {
	case *ast.Document:
		for _, def := range n.Definitions {
			add(def)
		}
	case *ast.OperationDefinition:
		for _, vd := range n.VariableDefinitions {
			add(vd)
		}
		for _, d := range n.Directives {
			add(d)
		}
		add(n.SelectionSet)
	case *ast.FragmentDefinition:
		add(n.TypeCondition)
		for _, d := range n.Directives {
			add(d)
		}
		add(n.SelectionSet)
	case *ast.VariableDefinition:
		add(n.Variable, n.Type, n.DefaultValue)
		for _, d := range n.Directives {
			add(d)
		}
	case *ast.SelectionSet:
		for _, sel := range n.Selections {
			add(sel)
		}
	case *ast.Field:
		for _, a := range n.Arguments {
			add(a)
		}
		for _, d := range n.Directives {
			add(d)
		}
		add(n.SelectionSet)
	case *ast.FragmentSpread:
		for _, d := range n.Directives {
			add(d)
		}
	case *ast.InlineFragment:
		add(n.TypeCondition)
		for _, d := range n.Directives {
			add(d)
		}
		add(n.SelectionSet)
	case *ast.Argument:
		add(n.Value)
	case *ast.Directive:
		for _, a := range n.Arguments {
			add(a)
		}
	case *ast.ListValue:
		for _, v := range n.Values {
			add(v)
		}
	case *ast.ObjectValue:
		for _, f := range n.Fields {
			add(f)
		}
	case *ast.ObjectField:
		add(n.Value)
	case *ast.ListType:
		add(n.OfType)
	case *ast.NonNullType:
		add(n.OfType)
	case *ast.SchemaDefinition:
		add(n.Description)
		for _, d := range n.Directives {
			add(d)
		}
		for _, ot := range n.OperationTypes {
			add(ot)
		}
	case *ast.SchemaExtension:
		for _, d := range n.Directives {
			add(d)
		}
		for _, ot := range n.OperationTypes {
			add(ot)
		}
	case *ast.OperationTypeDefinition:
		add(n.Type)
	case *ast.ScalarTypeDefinition:
		add(n.Description)
		for _, d := range n.Directives {
			add(d)
		}
	case *ast.ScalarTypeExtension:
		for _, d := range n.Directives {
			add(d)
		}
	case *ast.ObjectTypeDefinition:
		add(n.Description)
		for _, i := range n.Interfaces {
			add(i)
		}
		for _, d := range n.Directives {
			add(d)
		}
		for _, f := range n.Fields {
			add(f)
		}
	case *ast.ObjectTypeExtension:
		for _, i := range n.Interfaces {
			add(i)
		}
		for _, d := range n.Directives {
			add(d)
		}
		for _, f := range n.Fields {
			add(f)
		}
	case *ast.InterfaceTypeDefinition:
		add(n.Description)
		for _, i := range n.Interfaces {
			add(i)
		}
		for _, d := range n.Directives {
			add(d)
		}
		for _, f := range n.Fields {
			add(f)
		}
	case *ast.InterfaceTypeExtension:
		for _, i := range n.Interfaces {
			add(i)
		}
		for _, d := range n.Directives {
			add(d)
		}
		for _, f := range n.Fields {
			add(f)
		}
	case *ast.UnionTypeDefinition:
		add(n.Description)
		for _, d := range n.Directives {
			add(d)
		}
		for _, m := range n.Members {
			add(m)
		}
	case *ast.UnionTypeExtension:
		for _, d := range n.Directives {
			add(d)
		}
		for _, m := range n.Members {
			add(m)
		}
	case *ast.EnumTypeDefinition:
		add(n.Description)
		for _, d := range n.Directives {
			add(d)
		}
		for _, v := range n.Values {
			add(v)
		}
	case *ast.EnumTypeExtension:
		for _, d := range n.Directives {
			add(d)
		}
		for _, v := range n.Values {
			add(v)
		}
	case *ast.InputObjectTypeDefinition:
		add(n.Description)
		for _, d := range n.Directives {
			add(d)
		}
		for _, f := range n.Fields {
			add(f)
		}
	case *ast.InputObjectTypeExtension:
		for _, d := range n.Directives {
			add(d)
		}
		for _, f := range n.Fields {
			add(f)
		}
	case *ast.FieldDefinition:
		add(n.Description)
		for _, a := range n.Arguments {
			add(a)
		}
		add(n.Type)
		for _, d := range n.Directives {
			add(d)
		}
	case *ast.InputValueDefinition:
		add(n.Description, n.Type, n.DefaultValue)
		for _, d := range n.Directives {
			add(d)
		}
	case *ast.EnumValueDefinition:
		add(n.Description)
		for _, d := range n.Directives {
			add(d)
		}
	case *ast.DirectiveDefinition:
		add(n.Description)
		for _, a := range n.Arguments {
			add(a)
		}
	case *ast.IntValue, *ast.FloatValue, *ast.StringValue, *ast.BooleanValue,
		*ast.NullValue, *ast.EnumValue, *ast.Variable, *ast.NamedType,
		*ast.Comment, *ast.BadValue, *ast.BadType, *ast.BadField,
		*ast.BadDefinition:
	default:
		t.Fatalf("childrenOf missing AST node type %T", node)
	}
	return children
}

func isNilNode(node ast.Node) bool {
	if node == nil {
		return true
	}
	switch n := node.(type) {
	case *ast.Document:
		return n == nil
	case *ast.OperationDefinition:
		return n == nil
	case *ast.FragmentDefinition:
		return n == nil
	case *ast.VariableDefinition:
		return n == nil
	case *ast.SelectionSet:
		return n == nil
	case *ast.Field:
		return n == nil
	case *ast.FragmentSpread:
		return n == nil
	case *ast.InlineFragment:
		return n == nil
	case *ast.Argument:
		return n == nil
	case *ast.Directive:
		return n == nil
	case *ast.IntValue:
		return n == nil
	case *ast.FloatValue:
		return n == nil
	case *ast.StringValue:
		return n == nil
	case *ast.BooleanValue:
		return n == nil
	case *ast.NullValue:
		return n == nil
	case *ast.EnumValue:
		return n == nil
	case *ast.ListValue:
		return n == nil
	case *ast.ObjectValue:
		return n == nil
	case *ast.ObjectField:
		return n == nil
	case *ast.Variable:
		return n == nil
	case *ast.NamedType:
		return n == nil
	case *ast.ListType:
		return n == nil
	case *ast.NonNullType:
		return n == nil
	case *ast.SchemaDefinition:
		return n == nil
	case *ast.SchemaExtension:
		return n == nil
	case *ast.OperationTypeDefinition:
		return n == nil
	case *ast.ScalarTypeDefinition:
		return n == nil
	case *ast.ScalarTypeExtension:
		return n == nil
	case *ast.ObjectTypeDefinition:
		return n == nil
	case *ast.ObjectTypeExtension:
		return n == nil
	case *ast.InterfaceTypeDefinition:
		return n == nil
	case *ast.InterfaceTypeExtension:
		return n == nil
	case *ast.UnionTypeDefinition:
		return n == nil
	case *ast.UnionTypeExtension:
		return n == nil
	case *ast.EnumTypeDefinition:
		return n == nil
	case *ast.EnumTypeExtension:
		return n == nil
	case *ast.InputObjectTypeDefinition:
		return n == nil
	case *ast.InputObjectTypeExtension:
		return n == nil
	case *ast.FieldDefinition:
		return n == nil
	case *ast.InputValueDefinition:
		return n == nil
	case *ast.EnumValueDefinition:
		return n == nil
	case *ast.DirectiveDefinition:
		return n == nil
	case *ast.Comment:
		return n == nil
	case *ast.BadValue:
		return n == nil
	case *ast.BadType:
		return n == nil
	case *ast.BadField:
		return n == nil
	case *ast.BadDefinition:
		return n == nil
	}
	return false
}

func assertLocText(t *testing.T, body string, node ast.Node, want string) {
	t.Helper()
	loc := node.GetLoc()
	if loc == nil {
		t.Fatalf("%T has nil Loc", node)
	}
	if loc.Start < 0 || loc.Start > loc.End || loc.End > len(body) {
		t.Fatalf("%T Loc = [%d, %d), outside body len %d", node, loc.Start, loc.End, len(body))
	}
	got := body[loc.Start:loc.End]
	if got != want {
		t.Fatalf("%T Loc text = %s; want %s", node, quoteForLoc(got), quoteForLoc(want))
	}
	if idx := strings.Index(body, want); idx >= 0 && (loc.Start != idx || loc.End != idx+len(want)) {
		t.Fatalf("%T Loc = [%d, %d); want first %q at [%d, %d)", node, loc.Start, loc.End, want, idx, idx+len(want))
	}
}

func quoteForLoc(s string) string {
	return strconv.Quote(s)
}
