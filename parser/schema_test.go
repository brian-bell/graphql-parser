package parser_test

import (
	"strings"
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/parser"
)

func TestParse_SchemaDefinition(t *testing.T) {
	doc := mustParse(t, `schema { query: Query mutation: Mutation }`)
	sd := doc.Definitions[0].(*ast.SchemaDefinition)
	if len(sd.OperationTypes) != 2 {
		t.Fatalf("ots = %d; want 2", len(sd.OperationTypes))
	}
	if sd.OperationTypes[0].Operation != ast.OperationQuery {
		t.Errorf("op[0] = %q; want query", sd.OperationTypes[0].Operation)
	}
	if sd.OperationTypes[0].Type.Name != "Query" {
		t.Errorf("op[0].Type = %q; want Query", sd.OperationTypes[0].Type.Name)
	}
}

func TestParse_SchemaWithDescriptionAndDirectives(t *testing.T) {
	body := `"the schema" schema @secured { query: Q }`
	doc := mustParse(t, body)
	sd := doc.Definitions[0].(*ast.SchemaDefinition)
	if sd.Description == nil || sd.Description.Value != "the schema" {
		t.Errorf("desc = %v", sd.Description)
	}
	if len(sd.Directives) != 1 || sd.Directives[0].Name != "secured" {
		t.Errorf("directives = %v", sd.Directives)
	}
}

func TestParse_ScalarType(t *testing.T) {
	doc := mustParse(t, `scalar URL @specifiedBy(url: "https://example.com")`)
	sd := doc.Definitions[0].(*ast.ScalarTypeDefinition)
	if sd.Name != "URL" {
		t.Errorf("name = %q; want URL", sd.Name)
	}
	if len(sd.Directives) != 1 {
		t.Errorf("directives = %d; want 1", len(sd.Directives))
	}
}

func TestParse_ObjectType(t *testing.T) {
	body := `
		"User account."
		type User implements Node & HasName @auth(role: ADMIN) {
			"the id"
			id: ID!
			name(format: NameFormat = LONG): String!
			friends(first: Int): [User!]!
		}`
	doc := mustParse(t, body)
	od := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	if od.Description == nil || od.Description.Value != "User account." {
		t.Errorf("desc = %v", od.Description)
	}
	if od.Name != "User" {
		t.Errorf("name = %q", od.Name)
	}
	if len(od.Interfaces) != 2 {
		t.Errorf("interfaces = %d; want 2", len(od.Interfaces))
	}
	if od.Interfaces[0].Name != "Node" || od.Interfaces[1].Name != "HasName" {
		t.Errorf("interfaces = %v", od.Interfaces)
	}
	if len(od.Directives) != 1 || od.Directives[0].Name != "auth" {
		t.Errorf("directives = %v", od.Directives)
	}
	if len(od.Fields) != 3 {
		t.Fatalf("fields = %d; want 3", len(od.Fields))
	}
	idField := od.Fields.ForName("id")
	if idField == nil || idField.Description == nil || idField.Description.Value != "the id" {
		t.Errorf("id field wrong: %+v", idField)
	}
	nameField := od.Fields.ForName("name")
	if nameField == nil || len(nameField.Arguments) != 1 {
		t.Errorf("name args = %v", nameField)
	}
	formatArg := nameField.Arguments.ForName("format")
	if formatArg == nil || formatArg.DefaultValue == nil {
		t.Errorf("format arg has no default: %v", formatArg)
	}
}

func TestParse_ObjectType_LeadingAmpInImplements(t *testing.T) {
	doc := mustParse(t, `type T implements & A & B { x: Int }`)
	od := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	if len(od.Interfaces) != 2 {
		t.Errorf("interfaces = %d; want 2", len(od.Interfaces))
	}
}

func TestParse_InterfaceType(t *testing.T) {
	doc := mustParse(t, `interface Node implements Identifiable { id: ID! }`)
	id := doc.Definitions[0].(*ast.InterfaceTypeDefinition)
	if id.Name != "Node" {
		t.Errorf("name = %q", id.Name)
	}
	if len(id.Interfaces) != 1 || id.Interfaces[0].Name != "Identifiable" {
		t.Errorf("interfaces = %v", id.Interfaces)
	}
}

func TestParse_UnionType(t *testing.T) {
	doc := mustParse(t, `union SearchResult = User | Post | Comment`)
	ud := doc.Definitions[0].(*ast.UnionTypeDefinition)
	if len(ud.Members) != 3 {
		t.Errorf("members = %d; want 3", len(ud.Members))
	}
}

func TestParse_UnionType_LeadingPipe(t *testing.T) {
	doc := mustParse(t, `union U = | A | B`)
	ud := doc.Definitions[0].(*ast.UnionTypeDefinition)
	if len(ud.Members) != 2 {
		t.Errorf("members = %d; want 2", len(ud.Members))
	}
}

func TestParse_UnionType_NoMembers(t *testing.T) {
	doc := mustParse(t, `union Empty`)
	ud := doc.Definitions[0].(*ast.UnionTypeDefinition)
	if len(ud.Members) != 0 {
		t.Errorf("members = %d; want 0", len(ud.Members))
	}
}

func TestParse_EnumType(t *testing.T) {
	body := `
		enum Color {
			"the color of fire"
			RED
			GREEN
			BLUE @deprecated(reason: "use INDIGO")
		}`
	doc := mustParse(t, body)
	ed := doc.Definitions[0].(*ast.EnumTypeDefinition)
	if len(ed.Values) != 3 {
		t.Fatalf("values = %d; want 3", len(ed.Values))
	}
	if ed.Values.ForName("RED").Description.Value != "the color of fire" {
		t.Error("RED desc wrong")
	}
	if len(ed.Values.ForName("BLUE").Directives) != 1 {
		t.Error("BLUE has no directive")
	}
}

func TestParse_EnumValueRejectsReserved(t *testing.T) {
	for _, body := range []string{`enum E { true }`, `enum E { false }`, `enum E { null }`} {
		if _, err := parser.Parse(body); err == nil {
			t.Errorf("%q: expected error", body)
		}
	}
}

func TestParse_InputObjectType(t *testing.T) {
	body := `input UserInput { name: String! age: Int = 18 @validate(min: 0) }`
	doc := mustParse(t, body)
	id := doc.Definitions[0].(*ast.InputObjectTypeDefinition)
	if len(id.Fields) != 2 {
		t.Fatalf("fields = %d; want 2", len(id.Fields))
	}
	age := id.Fields.ForName("age")
	if age == nil || age.DefaultValue == nil {
		t.Error("age field has no default")
	}
	if iv, ok := age.DefaultValue.(*ast.IntValue); !ok || iv.Value != "18" {
		t.Errorf("age default = %v; want IntValue 18", age.DefaultValue)
	}
	if len(age.Directives) != 1 {
		t.Errorf("age directives = %d; want 1", len(age.Directives))
	}
}

func TestParse_DirectiveDefinition(t *testing.T) {
	body := `directive @auth(role: Role!) on FIELD | OBJECT | FIELD_DEFINITION`
	doc := mustParse(t, body)
	dd := doc.Definitions[0].(*ast.DirectiveDefinition)
	if dd.Name != "auth" {
		t.Errorf("name = %q", dd.Name)
	}
	if len(dd.Arguments) != 1 {
		t.Errorf("args = %d; want 1", len(dd.Arguments))
	}
	if dd.Repeatable {
		t.Error("Repeatable = true; want false")
	}
	want := []string{"FIELD", "OBJECT", "FIELD_DEFINITION"}
	if len(dd.Locations) != len(want) {
		t.Fatalf("locations = %v; want %v", dd.Locations, want)
	}
	for i, w := range want {
		if dd.Locations[i] != w {
			t.Errorf("locations[%d] = %q; want %q", i, dd.Locations[i], w)
		}
	}
}

func TestParse_DirectiveDefinition_Repeatable(t *testing.T) {
	doc := mustParse(t, `directive @tag(name: String!) repeatable on OBJECT`)
	dd := doc.Definitions[0].(*ast.DirectiveDefinition)
	if !dd.Repeatable {
		t.Error("Repeatable = false; want true")
	}
}

func TestParse_DirectiveDefinition_LeadingPipe(t *testing.T) {
	doc := mustParse(t, `directive @x on | FIELD | OBJECT`)
	dd := doc.Definitions[0].(*ast.DirectiveDefinition)
	if len(dd.Locations) != 2 {
		t.Errorf("locations = %d; want 2", len(dd.Locations))
	}
}

func TestParse_DirectiveDefinition_RejectsBadLocation(t *testing.T) {
	if _, err := parser.Parse(`directive @x on FROBNICATE`); err == nil {
		t.Error("expected error for unknown directive location")
	}
}

func TestParse_Extensions(t *testing.T) {
	body := `
		extend schema { mutation: Mutation }
		extend scalar URL @new
		extend type User { x: Int }
		extend interface Node { y: Int }
		extend union Result = Z
		extend enum Color { ORANGE }
		extend input UserInput { phone: String }`
	doc := mustParse(t, body)
	if len(doc.Definitions) != 7 {
		t.Fatalf("defs = %d; want 7", len(doc.Definitions))
	}
	if _, ok := doc.Definitions[0].(*ast.SchemaExtension); !ok {
		t.Error("[0] not SchemaExtension")
	}
	if _, ok := doc.Definitions[1].(*ast.ScalarTypeExtension); !ok {
		t.Error("[1] not ScalarTypeExtension")
	}
	if _, ok := doc.Definitions[2].(*ast.ObjectTypeExtension); !ok {
		t.Error("[2] not ObjectTypeExtension")
	}
	if _, ok := doc.Definitions[3].(*ast.InterfaceTypeExtension); !ok {
		t.Error("[3] not InterfaceTypeExtension")
	}
	if _, ok := doc.Definitions[4].(*ast.UnionTypeExtension); !ok {
		t.Error("[4] not UnionTypeExtension")
	}
	if _, ok := doc.Definitions[5].(*ast.EnumTypeExtension); !ok {
		t.Error("[5] not EnumTypeExtension")
	}
	if _, ok := doc.Definitions[6].(*ast.InputObjectTypeExtension); !ok {
		t.Error("[6] not InputObjectTypeExtension")
	}
}

func TestParse_Extensions_RejectsLeadingDescription(t *testing.T) {
	body := `"a description" extend type User { x: Int }`
	_, err := parser.Parse(body)
	if err == nil {
		t.Fatal("expected error: extensions cannot have description")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "description") {
		t.Errorf("error %q does not mention description", err.Error())
	}
}

func TestParse_Extensions_RequireSomeContent(t *testing.T) {
	cases := []string{
		`extend schema`,
		`extend scalar URL`,
		`extend type User`,
		`extend interface Node`,
		`extend union U`,
		`extend enum E`,
		`extend input I`,
	}
	for _, body := range cases {
		t.Run(body, func(t *testing.T) {
			if _, err := parser.Parse(body); err == nil {
				t.Errorf("%q: expected error for empty extension", body)
			}
		})
	}
}

func TestParse_DescriptionsRequiredOnDefinitions(t *testing.T) {
	// Description as a block string is fine.
	body := `"""multi-line description""" type T { x: Int }`
	doc := mustParse(t, body)
	od := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	if od.Description == nil || !od.Description.Block || od.Description.Value != "multi-line description" {
		t.Errorf("desc wrong: %+v", od.Description)
	}
}

func TestParse_DescriptionsAcceptedOnAllDefinitionKinds(t *testing.T) {
	body := `
		"a" schema { query: Q }
		"b" scalar URL
		"c" type T { f: Int }
		"d" interface I { f: Int }
		"e" union U = A
		"f" enum E { A }
		"g" input In { f: Int }
		"h" directive @foo on FIELD`
	doc := mustParse(t, body)
	if len(doc.Definitions) != 8 {
		t.Fatalf("defs = %d; want 8", len(doc.Definitions))
	}
}

func TestParse_FieldDefinitionWithDescriptionAndArguments(t *testing.T) {
	body := `
		type T {
			"the field" field(
				"first arg" first: Int = 10
				"second arg" second: String!
			): Boolean! @auth
		}`
	doc := mustParse(t, body)
	od := doc.Definitions[0].(*ast.ObjectTypeDefinition)
	f := od.Fields.ForName("field")
	if f == nil {
		t.Fatal("missing 'field'")
	}
	if f.Description == nil || f.Description.Value != "the field" {
		t.Errorf("field desc = %v", f.Description)
	}
	if len(f.Arguments) != 2 {
		t.Fatalf("args = %d; want 2", len(f.Arguments))
	}
	if f.Arguments.ForName("first").Description.Value != "first arg" {
		t.Error("first-arg description wrong")
	}
}
