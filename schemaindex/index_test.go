package schemaindex_test

import (
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/parser"
	"github.com/brian-bell/graphql-parser/schemaindex"
)

func TestNewIndexesNamedTypeDefinitionsByName(t *testing.T) {
	doc := mustParseSchema(t, `
		scalar URL
		type Query { search: Search }
		interface Node { id: ID! }
		type User implements Node { id: ID! }
		union Search = User
		enum Color { RED GREEN }
		input UserInput { id: ID }
	`)

	idx := schemaindex.New(doc)
	cases := []struct {
		name      string
		assertDef func(t *testing.T, def ast.Definition)
	}{
		{
			name: "URL",
			assertDef: func(t *testing.T, def ast.Definition) {
				t.Helper()
				if _, ok := def.(*ast.ScalarTypeDefinition); !ok {
					t.Fatalf("definition = %T; want *ast.ScalarTypeDefinition", def)
				}
			},
		},
		{
			name: "Query",
			assertDef: func(t *testing.T, def ast.Definition) {
				t.Helper()
				if _, ok := def.(*ast.ObjectTypeDefinition); !ok {
					t.Fatalf("definition = %T; want *ast.ObjectTypeDefinition", def)
				}
			},
		},
		{
			name: "Node",
			assertDef: func(t *testing.T, def ast.Definition) {
				t.Helper()
				if _, ok := def.(*ast.InterfaceTypeDefinition); !ok {
					t.Fatalf("definition = %T; want *ast.InterfaceTypeDefinition", def)
				}
			},
		},
		{
			name: "Search",
			assertDef: func(t *testing.T, def ast.Definition) {
				t.Helper()
				if _, ok := def.(*ast.UnionTypeDefinition); !ok {
					t.Fatalf("definition = %T; want *ast.UnionTypeDefinition", def)
				}
			},
		},
		{
			name: "Color",
			assertDef: func(t *testing.T, def ast.Definition) {
				t.Helper()
				if _, ok := def.(*ast.EnumTypeDefinition); !ok {
					t.Fatalf("definition = %T; want *ast.EnumTypeDefinition", def)
				}
			},
		},
		{
			name: "UserInput",
			assertDef: func(t *testing.T, def ast.Definition) {
				t.Helper()
				if _, ok := def.(*ast.InputObjectTypeDefinition); !ok {
					t.Fatalf("definition = %T; want *ast.InputObjectTypeDefinition", def)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entry := idx.LookupType(tc.name)
			if entry == nil {
				t.Fatalf("LookupType(%q) = nil", tc.name)
			}
			if entry.Name() != tc.name {
				t.Fatalf("entry.Name() = %q; want %q", entry.Name(), tc.name)
			}
			defs := entry.BaseDefinitions()
			if len(defs) != 1 {
				t.Fatalf("BaseDefinitions() length = %d; want 1", len(defs))
			}
			tc.assertDef(t, defs[0])
			if got := len(entry.Extensions()); got != 0 {
				t.Fatalf("Extensions() length = %d; want 0", got)
			}
		})
	}

	if got := idx.LookupType("Missing"); got != nil {
		t.Fatalf("LookupType(%q) = %#v; want nil", "Missing", got)
	}
}

func TestNewRecordsExtensionsSeparatelyInSourceOrder(t *testing.T) {
	doc := mustParseSchema(t, `
		extend scalar URL @specifiedBy(url: "https://example.com")
		extend type Query { second: String }
		type Query { first: String }
		extend interface Node { name: String }
		extend union Search = User
		extend enum Status { ARCHIVED }
		extend input UserInput { slug: String }
		extend type Query { third: String }
	`)

	idx := schemaindex.New(doc)
	query := idx.LookupType("Query")
	if query == nil {
		t.Fatal(`LookupType("Query") = nil`)
	}
	if got := len(query.BaseDefinitions()); got != 1 {
		t.Fatalf("Query BaseDefinitions() length = %d; want 1", got)
	}
	extensions := query.Extensions()
	if got := len(extensions); got != 2 {
		t.Fatalf("Query Extensions() length = %d; want 2", got)
	}
	firstExt := requireObjectTypeExtension(t, extensions[0])
	secondExt := requireObjectTypeExtension(t, extensions[1])
	if firstExt.Fields[0].Name != "second" {
		t.Fatalf("first Query extension field = %q; want second", firstExt.Fields[0].Name)
	}
	if secondExt.Fields[0].Name != "third" {
		t.Fatalf("second Query extension field = %q; want third", secondExt.Fields[0].Name)
	}

	cases := []struct {
		name      string
		assertExt func(t *testing.T, def ast.Definition)
	}{
		{
			name: "URL",
			assertExt: func(t *testing.T, def ast.Definition) {
				t.Helper()
				if _, ok := def.(*ast.ScalarTypeExtension); !ok {
					t.Fatalf("extension = %T; want *ast.ScalarTypeExtension", def)
				}
			},
		},
		{
			name: "Node",
			assertExt: func(t *testing.T, def ast.Definition) {
				t.Helper()
				if _, ok := def.(*ast.InterfaceTypeExtension); !ok {
					t.Fatalf("extension = %T; want *ast.InterfaceTypeExtension", def)
				}
			},
		},
		{
			name: "Search",
			assertExt: func(t *testing.T, def ast.Definition) {
				t.Helper()
				if _, ok := def.(*ast.UnionTypeExtension); !ok {
					t.Fatalf("extension = %T; want *ast.UnionTypeExtension", def)
				}
			},
		},
		{
			name: "Status",
			assertExt: func(t *testing.T, def ast.Definition) {
				t.Helper()
				if _, ok := def.(*ast.EnumTypeExtension); !ok {
					t.Fatalf("extension = %T; want *ast.EnumTypeExtension", def)
				}
			},
		},
		{
			name: "UserInput",
			assertExt: func(t *testing.T, def ast.Definition) {
				t.Helper()
				if _, ok := def.(*ast.InputObjectTypeExtension); !ok {
					t.Fatalf("extension = %T; want *ast.InputObjectTypeExtension", def)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entry := idx.LookupType(tc.name)
			if entry == nil {
				t.Fatalf("LookupType(%q) = nil", tc.name)
			}
			if got := len(entry.BaseDefinitions()); got != 0 {
				t.Fatalf("BaseDefinitions() length = %d; want 0", got)
			}
			extensions := entry.Extensions()
			if got := len(extensions); got != 1 {
				t.Fatalf("Extensions() length = %d; want 1", got)
			}
			tc.assertExt(t, extensions[0])
		})
	}
}

func TestNewPreservesDuplicateBaseDefinitionsInSourceOrder(t *testing.T) {
	doc := mustParseSchema(t, `
		type Query { first: String }
		type Query { second: String }
	`)

	idx := schemaindex.New(doc)
	entry := idx.LookupType("Query")
	if entry == nil {
		t.Fatal(`LookupType("Query") = nil`)
	}
	defs := entry.BaseDefinitions()
	if got := len(defs); got != 2 {
		t.Fatalf("BaseDefinitions() length = %d; want 2", got)
	}
	first := requireObjectTypeDefinition(t, defs[0])
	second := requireObjectTypeDefinition(t, defs[1])
	if first.Fields[0].Name != "first" {
		t.Fatalf("first Query base definition field = %q; want first", first.Fields[0].Name)
	}
	if second.Fields[0].Name != "second" {
		t.Fatalf("second Query base definition field = %q; want second", second.Fields[0].Name)
	}
}

func TestNewDoesNotValidateSchemaSemantics(t *testing.T) {
	doc := mustParseSchema(t, `
		type Query { missing: Missing @unknown }
		interface Query { id: ID! }
		extend type Product @key(fields: "id") { id: ID! related: Missing }
	`)

	idx := schemaindex.New(doc)
	query := idx.LookupType("Query")
	if query == nil {
		t.Fatal(`LookupType("Query") = nil`)
	}
	defs := query.BaseDefinitions()
	if got := len(defs); got != 2 {
		t.Fatalf("Query BaseDefinitions() length = %d; want 2", got)
	}
	if _, ok := defs[0].(*ast.ObjectTypeDefinition); !ok {
		t.Fatalf("Query definition[0] = %T; want *ast.ObjectTypeDefinition", defs[0])
	}
	if _, ok := defs[1].(*ast.InterfaceTypeDefinition); !ok {
		t.Fatalf("Query definition[1] = %T; want *ast.InterfaceTypeDefinition", defs[1])
	}

	product := idx.LookupType("Product")
	if product == nil {
		t.Fatal(`LookupType("Product") = nil`)
	}
	if got := len(product.BaseDefinitions()); got != 0 {
		t.Fatalf("Product BaseDefinitions() length = %d; want 0", got)
	}
	extensions := product.Extensions()
	if got := len(extensions); got != 1 {
		t.Fatalf("Product Extensions() length = %d; want 1", got)
	}
	if _, ok := extensions[0].(*ast.ObjectTypeExtension); !ok {
		t.Fatalf("Product extension = %T; want *ast.ObjectTypeExtension", extensions[0])
	}
}

func TestNewIgnoresNonTypeDefinitions(t *testing.T) {
	doc, err := parser.Parse(`
		schema { query: Root }
		extend schema @tag
		directive @tag on SCHEMA
		query Get { viewer }
		fragment Fields on Query { viewer }
		type Query { viewer: String }
	`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	doc.Definitions = append([]ast.Definition{&ast.BadDefinition{}}, doc.Definitions...)

	idx := schemaindex.New(doc)
	if idx.LookupType("Query") == nil {
		t.Fatal(`LookupType("Query") = nil`)
	}
	for _, name := range []string{"Root", "tag", "Get", "Fields"} {
		if got := idx.LookupType(name); got != nil {
			t.Fatalf("LookupType(%q) = %#v; want nil", name, got)
		}
	}
}

func TestNewNilDocumentIsEmpty(t *testing.T) {
	idx := schemaindex.New(nil)
	if got := idx.LookupType("Query"); got != nil {
		t.Fatalf("LookupType(%q) = %#v; want nil", "Query", got)
	}
}

func TestEntryDefinitionSlicesAreCopies(t *testing.T) {
	doc := mustParseSchema(t, `
		type Query { name: String }
		extend type Query { age: Int }
	`)

	entry := schemaindex.New(doc).LookupType("Query")
	if entry == nil {
		t.Fatal(`LookupType("Query") = nil`)
	}

	wantBase := doc.Definitions[0]
	wantExtension := doc.Definitions[1]
	baseDefs := entry.BaseDefinitions()
	if got := baseDefs[0]; got != wantBase {
		t.Fatalf("BaseDefinitions()[0] = %T; want original parsed node %T", got, wantBase)
	}
	baseDefs[0] = nil
	extensions := entry.Extensions()
	if got := extensions[0]; got != wantExtension {
		t.Fatalf("Extensions()[0] = %T; want original parsed node %T", got, wantExtension)
	}
	extensions[0] = nil

	if got := entry.BaseDefinitions()[0]; got != wantBase {
		t.Fatalf("BaseDefinitions()[0] = %T after mutating returned slice; want original parsed node %T", got, wantBase)
	}
	if got := entry.Extensions()[0]; got != wantExtension {
		t.Fatalf("Extensions()[0] = %T after mutating returned slice; want original parsed node %T", got, wantExtension)
	}
}

func requireObjectTypeDefinition(t *testing.T, def ast.Definition) *ast.ObjectTypeDefinition {
	t.Helper()
	objectDef, ok := def.(*ast.ObjectTypeDefinition)
	if !ok {
		t.Fatalf("definition = %T; want *ast.ObjectTypeDefinition", def)
	}
	return objectDef
}

func requireObjectTypeExtension(t *testing.T, def ast.Definition) *ast.ObjectTypeExtension {
	t.Helper()
	objectExt, ok := def.(*ast.ObjectTypeExtension)
	if !ok {
		t.Fatalf("extension = %T; want *ast.ObjectTypeExtension", def)
	}
	return objectExt
}

func mustParseSchema(t *testing.T, body string) *ast.Document {
	t.Helper()
	doc, err := parser.ParseSchema(body)
	if err != nil {
		t.Fatalf("ParseSchema returned error: %v", err)
	}
	return doc
}
