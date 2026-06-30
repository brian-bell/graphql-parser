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

func TestTypeEntryObjectFieldsFoldBaseAndExtensions(t *testing.T) {
	doc := mustParseSchema(t, `
		type Query { base: String }
		extend type Query { firstExtension: String }
		extend type Product { id: ID }
		extend type Query { secondExtension: String }
	`)

	idx := schemaindex.New(doc)
	query := idx.LookupType("Query")
	if query == nil {
		t.Fatal(`LookupType("Query") = nil`)
	}
	assertFieldNames(t, query.ObjectFields(), "base", "firstExtension", "secondExtension")

	product := idx.LookupType("Product")
	if product == nil {
		t.Fatal(`LookupType("Product") = nil`)
	}
	assertFieldNames(t, product.ObjectFields(), "id")
}

func TestTypeEntryInterfacesFoldBaseAndExtensions(t *testing.T) {
	doc := mustParseSchema(t, `
		type User implements Node { id: ID }
		extend type User implements Resource { resource: String }
		interface Resource implements Node { id: ID }
		extend interface Resource implements Named { name: String }
		extend type Product implements Node { id: ID }
		extend interface Entity implements Node { id: ID }
	`)

	idx := schemaindex.New(doc)
	user := idx.LookupType("User")
	if user == nil {
		t.Fatal(`LookupType("User") = nil`)
	}
	assertNamedTypeNames(t, user.ObjectInterfaces(), "Node", "Resource")

	product := idx.LookupType("Product")
	if product == nil {
		t.Fatal(`LookupType("Product") = nil`)
	}
	assertNamedTypeNames(t, product.ObjectInterfaces(), "Node")

	resource := idx.LookupType("Resource")
	if resource == nil {
		t.Fatal(`LookupType("Resource") = nil`)
	}
	assertNamedTypeNames(t, resource.InterfaceInterfaces(), "Node", "Named")

	entity := idx.LookupType("Entity")
	if entity == nil {
		t.Fatal(`LookupType("Entity") = nil`)
	}
	assertNamedTypeNames(t, entity.InterfaceInterfaces(), "Node")
}

func TestTypeEntryInterfaceFieldsFoldBaseAndExtensions(t *testing.T) {
	doc := mustParseSchema(t, `
		interface Node { id: ID! }
		extend interface Node { createdAt: String }
		extend interface Entity { id: ID! }
		extend interface Node { updatedAt: String }
	`)

	idx := schemaindex.New(doc)
	node := idx.LookupType("Node")
	if node == nil {
		t.Fatal(`LookupType("Node") = nil`)
	}
	assertFieldNames(t, node.InterfaceFields(), "id", "createdAt", "updatedAt")

	entity := idx.LookupType("Entity")
	if entity == nil {
		t.Fatal(`LookupType("Entity") = nil`)
	}
	assertFieldNames(t, entity.InterfaceFields(), "id")
}

func TestTypeEntryInputFieldsFoldBaseAndExtensions(t *testing.T) {
	doc := mustParseSchema(t, `
		input UserInput { id: ID }
		extend input UserInput { name: String }
		extend input ProductInput { sku: String }
		extend input UserInput { email: String }
	`)

	idx := schemaindex.New(doc)
	userInput := idx.LookupType("UserInput")
	if userInput == nil {
		t.Fatal(`LookupType("UserInput") = nil`)
	}
	assertInputValueNames(t, userInput.InputFields(), "id", "name", "email")

	productInput := idx.LookupType("ProductInput")
	if productInput == nil {
		t.Fatal(`LookupType("ProductInput") = nil`)
	}
	assertInputValueNames(t, productInput.InputFields(), "sku")
}

func TestTypeEntryEnumValuesFoldBaseAndExtensions(t *testing.T) {
	doc := mustParseSchema(t, `
		extend enum Status { PENDING }
		enum Status { ACTIVE }
		extend enum ExtensionOnly { FIRST }
		extend enum Status { ARCHIVED ACTIVE }
		extend enum ExtensionOnly { SECOND }
	`)

	idx := schemaindex.New(doc)
	status := idx.LookupType("Status")
	if status == nil {
		t.Fatal(`LookupType("Status") = nil`)
	}
	assertEnumValueNames(t, status.EnumValues(), "ACTIVE", "PENDING", "ARCHIVED", "ACTIVE")

	extensionOnly := idx.LookupType("ExtensionOnly")
	if extensionOnly == nil {
		t.Fatal(`LookupType("ExtensionOnly") = nil`)
	}
	assertTypeNames(t, idx.TypeNames(), "Status", "ExtensionOnly")
	assertEnumValueNames(t, extensionOnly.EnumValues(), "FIRST", "SECOND")
}

func TestTypeEntryUnionMembersFoldBaseAndExtensions(t *testing.T) {
	doc := mustParseSchema(t, `
		extend union Search = Review
		union Search = User
		extend union SearchOnly = Product
		extend union Search = Product | Missing | User
		extend union SearchOnly = Missing
	`)

	idx := schemaindex.New(doc)
	search := idx.LookupType("Search")
	if search == nil {
		t.Fatal(`LookupType("Search") = nil`)
	}
	assertNamedTypeNames(t, search.UnionMembers(), "User", "Review", "Product", "Missing", "User")

	searchOnly := idx.LookupType("SearchOnly")
	if searchOnly == nil {
		t.Fatal(`LookupType("SearchOnly") = nil`)
	}
	assertTypeNames(t, idx.TypeNames(), "Search", "SearchOnly")
	assertNamedTypeNames(t, searchOnly.UnionMembers(), "Product", "Missing")
}

func TestTypeEntryScalarDirectivesFoldBaseAndExtensions(t *testing.T) {
	doc := mustParseSchema(t, `
		directive @tag(name: String) on SCALAR
		directive @specifiedBy(url: String!) on SCALAR

		scalar URL @specifiedBy(url: "https://example.com/url")
		extend scalar URL @tag(name: "first")
		scalar DateTime
		extend scalar DateTime @specifiedBy(url: "https://example.com/date-time")
		extend scalar URL @tag(name: "second") @tag(name: "second-duplicate")
		extend scalar ExtensionOnly @tag(name: "only")
	`)

	idx := schemaindex.New(doc)
	url := idx.LookupType("URL")
	if url == nil {
		t.Fatal(`LookupType("URL") = nil`)
	}
	assertDirectiveNames(t, url.ScalarDirectives(), "specifiedBy", "tag", "tag", "tag")

	dateTime := idx.LookupType("DateTime")
	if dateTime == nil {
		t.Fatal(`LookupType("DateTime") = nil`)
	}
	assertDirectiveNames(t, dateTime.ScalarDirectives(), "specifiedBy")

	extensionOnly := idx.LookupType("ExtensionOnly")
	if extensionOnly == nil {
		t.Fatal(`LookupType("ExtensionOnly") = nil`)
	}
	assertTypeNames(t, idx.TypeNames(), "URL", "DateTime", "ExtensionOnly")
	assertDirectiveNames(t, extensionOnly.ScalarDirectives(), "tag")
}

func TestTypeEntryDirectiveOnlyExtensionsPreserveMetadataWithoutMembers(t *testing.T) {
	doc := mustParseSchema(t, `
		directive @tag on OBJECT | INTERFACE | INPUT_OBJECT

		type Query { base: String }
		extend type Query @tag
		extend type Product @tag

		interface Node { id: ID! }
		extend interface Node @tag
		extend interface Entity @tag

		input UserInput { id: ID }
		extend input UserInput @tag
		extend input ProductInput @tag
	`)

	idx := schemaindex.New(doc)
	query := idx.LookupType("Query")
	if query == nil {
		t.Fatal(`LookupType("Query") = nil`)
	}
	assertFieldNames(t, query.ObjectFields(), "base")
	if got := len(query.Extensions()); got != 1 {
		t.Fatalf("Query Extensions() length = %d; want 1", got)
	}

	product := idx.LookupType("Product")
	if product == nil {
		t.Fatal(`LookupType("Product") = nil`)
	}
	assertFieldNames(t, product.ObjectFields())
	if got := len(product.Extensions()); got != 1 {
		t.Fatalf("Product Extensions() length = %d; want 1", got)
	}

	node := idx.LookupType("Node")
	if node == nil {
		t.Fatal(`LookupType("Node") = nil`)
	}
	assertFieldNames(t, node.InterfaceFields(), "id")
	assertNamedTypeNames(t, node.InterfaceInterfaces())
	if got := len(node.Extensions()); got != 1 {
		t.Fatalf("Node Extensions() length = %d; want 1", got)
	}

	entity := idx.LookupType("Entity")
	if entity == nil {
		t.Fatal(`LookupType("Entity") = nil`)
	}
	assertFieldNames(t, entity.InterfaceFields())
	assertNamedTypeNames(t, entity.InterfaceInterfaces())
	if got := len(entity.Extensions()); got != 1 {
		t.Fatalf("Entity Extensions() length = %d; want 1", got)
	}

	userInput := idx.LookupType("UserInput")
	if userInput == nil {
		t.Fatal(`LookupType("UserInput") = nil`)
	}
	assertInputValueNames(t, userInput.InputFields(), "id")
	if got := len(userInput.Extensions()); got != 1 {
		t.Fatalf("UserInput Extensions() length = %d; want 1", got)
	}

	productInput := idx.LookupType("ProductInput")
	if productInput == nil {
		t.Fatal(`LookupType("ProductInput") = nil`)
	}
	assertInputValueNames(t, productInput.InputFields())
	if got := len(productInput.Extensions()); got != 1 {
		t.Fatalf("ProductInput Extensions() length = %d; want 1", got)
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
	assertTypeNames(t, idx.TypeNames(), "Query")
	for _, name := range []string{"Root", "tag", "Get", "Fields"} {
		if got := idx.LookupType(name); got != nil {
			t.Fatalf("LookupType(%q) = %#v; want nil", name, got)
		}
	}
}

func TestIndexTypeNamesAreFirstSeenCopies(t *testing.T) {
	doc, err := parser.Parse(`
		schema { query: Root }
		directive @tag on SCALAR
		query Get { viewer }
		fragment Fields on Query { viewer }
		extend enum Status { PENDING }
		enum Status { ACTIVE }
		type Query { viewer: String }
		extend type Query { other: String }
		scalar URL
		extend scalar URL @tag
	`)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	idx := schemaindex.New(doc)
	names := idx.TypeNames()
	assertTypeNames(t, names, "Status", "Query", "URL")

	names[0] = "Mutated"
	assertTypeNames(t, idx.TypeNames(), "Status", "Query", "URL")
}

func TestNewNilDocumentIsEmpty(t *testing.T) {
	idx := schemaindex.New(nil)
	if got := idx.LookupType("Query"); got != nil {
		t.Fatalf("LookupType(%q) = %#v; want nil", "Query", got)
	}
	assertTypeNames(t, idx.TypeNames())
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

func TestTypeEntryFoldedMemberSlicesAreCopies(t *testing.T) {
	doc := mustParseSchema(t, `
		type User implements Node { id: ID }
		extend type User implements Resource { name: String }
		interface Entity implements Node { id: ID }
		extend interface Entity implements Resource { name: String }
		input UserInput { id: ID }
		extend input UserInput { name: String }
		enum Status { ACTIVE }
		extend enum Status { ARCHIVED }
		union Search = User
		extend union Search = Product
		scalar URL @specifiedBy(url: "https://example.com/url")
		extend scalar URL @tag
	`)

	idx := schemaindex.New(doc)
	user := idx.LookupType("User")
	if user == nil {
		t.Fatal(`LookupType("User") = nil`)
	}
	fields := user.ObjectFields()
	fields[0] = nil
	assertFieldNames(t, user.ObjectFields(), "id", "name")

	interfaces := user.ObjectInterfaces()
	interfaces[0] = nil
	assertNamedTypeNames(t, user.ObjectInterfaces(), "Node", "Resource")

	entity := idx.LookupType("Entity")
	if entity == nil {
		t.Fatal(`LookupType("Entity") = nil`)
	}
	interfaceFields := entity.InterfaceFields()
	interfaceFields[0] = nil
	assertFieldNames(t, entity.InterfaceFields(), "id", "name")

	interfaceInterfaces := entity.InterfaceInterfaces()
	interfaceInterfaces[0] = nil
	assertNamedTypeNames(t, entity.InterfaceInterfaces(), "Node", "Resource")

	userInput := idx.LookupType("UserInput")
	if userInput == nil {
		t.Fatal(`LookupType("UserInput") = nil`)
	}
	inputFields := userInput.InputFields()
	inputFields[0] = nil
	assertInputValueNames(t, userInput.InputFields(), "id", "name")

	status := idx.LookupType("Status")
	if status == nil {
		t.Fatal(`LookupType("Status") = nil`)
	}
	enumValues := status.EnumValues()
	enumValues[0] = nil
	assertEnumValueNames(t, status.EnumValues(), "ACTIVE", "ARCHIVED")

	search := idx.LookupType("Search")
	if search == nil {
		t.Fatal(`LookupType("Search") = nil`)
	}
	unionMembers := search.UnionMembers()
	unionMembers[0] = nil
	assertNamedTypeNames(t, search.UnionMembers(), "User", "Product")

	url := idx.LookupType("URL")
	if url == nil {
		t.Fatal(`LookupType("URL") = nil`)
	}
	scalarDirectives := url.ScalarDirectives()
	scalarDirectives[0] = nil
	assertDirectiveNames(t, url.ScalarDirectives(), "specifiedBy", "tag")
}

func TestTypeEntryFoldedAccessorsIgnoreMixedKindEntries(t *testing.T) {
	doc := mustParseSchema(t, `
		scalar Shared @scalarBase
		type Shared { objectBase: String }
		interface Shared { interfaceBase: String }
		union Shared = BaseMember
		enum Shared { ENUM_BASE }
		input Shared { inputBase: String }
		extend scalar Shared @scalarExtension
		extend type Shared implements Node { objectExtension: String }
		extend interface Shared implements Resource { interfaceExtension: String }
		extend union Shared = ExtensionMember
		extend enum Shared { ENUM_EXTENSION }
		extend input Shared { inputExtension: String }
	`)

	entry := schemaindex.New(doc).LookupType("Shared")
	if entry == nil {
		t.Fatal(`LookupType("Shared") = nil`)
	}

	assertFieldNames(t, entry.ObjectFields(), "objectBase", "objectExtension")
	assertNamedTypeNames(t, entry.ObjectInterfaces(), "Node")
	assertFieldNames(t, entry.InterfaceFields(), "interfaceBase", "interfaceExtension")
	assertNamedTypeNames(t, entry.InterfaceInterfaces(), "Resource")
	assertInputValueNames(t, entry.InputFields(), "inputBase", "inputExtension")
	assertEnumValueNames(t, entry.EnumValues(), "ENUM_BASE", "ENUM_EXTENSION")
	assertNamedTypeNames(t, entry.UnionMembers(), "BaseMember", "ExtensionMember")
	assertDirectiveNames(t, entry.ScalarDirectives(), "scalarBase", "scalarExtension")
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

func assertFieldNames(t *testing.T, fields ast.FieldDefinitionList, names ...string) {
	t.Helper()
	if got := len(fields); got != len(names) {
		t.Fatalf("field count = %d; want %d", got, len(names))
	}
	for i, name := range names {
		if fields[i].Name != name {
			t.Fatalf("field[%d].Name = %q; want %q", i, fields[i].Name, name)
		}
	}
}

func assertNamedTypeNames(t *testing.T, types []*ast.NamedType, names ...string) {
	t.Helper()
	if got := len(types); got != len(names) {
		t.Fatalf("named type count = %d; want %d", got, len(names))
	}
	for i, name := range names {
		if types[i].Name != name {
			t.Fatalf("namedType[%d].Name = %q; want %q", i, types[i].Name, name)
		}
	}
}

func assertInputValueNames(t *testing.T, values ast.InputValueList, names ...string) {
	t.Helper()
	if got := len(values); got != len(names) {
		t.Fatalf("input value count = %d; want %d", got, len(names))
	}
	for i, name := range names {
		if values[i].Name != name {
			t.Fatalf("inputValue[%d].Name = %q; want %q", i, values[i].Name, name)
		}
	}
}

func assertEnumValueNames(t *testing.T, values ast.EnumValueList, names ...string) {
	t.Helper()
	if got := len(values); got != len(names) {
		t.Fatalf("enum value count = %d; want %d", got, len(names))
	}
	for i, name := range names {
		if values[i].Name != name {
			t.Fatalf("enumValue[%d].Name = %q; want %q", i, values[i].Name, name)
		}
	}
}

func assertDirectiveNames(t *testing.T, directives ast.DirectiveList, names ...string) {
	t.Helper()
	if got := len(directives); got != len(names) {
		t.Fatalf("directive count = %d; want %d", got, len(names))
	}
	for i, name := range names {
		if directives[i].Name != name {
			t.Fatalf("directive[%d].Name = %q; want %q", i, directives[i].Name, name)
		}
	}
}

func assertTypeNames(t *testing.T, names []string, want ...string) {
	t.Helper()
	if got := len(names); got != len(want) {
		t.Fatalf("type name count = %d; want %d", got, len(want))
	}
	for i, name := range want {
		if names[i] != name {
			t.Fatalf("typeNames[%d] = %q; want %q", i, names[i], name)
		}
	}
}

func mustParseSchema(t *testing.T, body string) *ast.Document {
	t.Helper()
	doc, err := parser.ParseSchema(body)
	if err != nil {
		t.Fatalf("ParseSchema returned error: %v", err)
	}
	return doc
}
