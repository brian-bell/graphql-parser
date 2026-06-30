package schemaindex_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/parser"
	"github.com/brian-bell/graphql-parser/schemaindex"
)

func TestRealWorldSDLFixturesParseSchema(t *testing.T) {
	fixtures := []string{
		"github.graphql",
		"saleor.graphql",
		"apollo-supergraph-demo.graphql",
		"swapi.graphql",
		"directive-only-extension.graphql",
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			src := readRealWorldFixture(t, fixture)
			if _, err := parser.ParseSchema(src); err != nil {
				t.Fatalf("ParseSchema(%s) error = %v", fixture, err)
			}
		})
	}
}

func readRealWorldFixture(t *testing.T, name string) string {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("testdata", "realworld-sdl", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(src)
}

func TestApolloSupergraphDemoFixtureIndexesExtensionsInSourceOrder(t *testing.T) {
	idx := realWorldFixtureIndex(t, "apollo-supergraph-demo.graphql")

	user := requireRealWorldEntry(t, idx, "User")
	assertFieldNames(t, user.ObjectFields(), "email", "name", "totalProductsCreated", "email", "totalProductsCreated")
	userExtensions := user.Extensions()
	if got := len(userExtensions); got != 1 {
		t.Fatalf("User Extensions() length = %d; want 1", got)
	}
	userExt := requireObjectTypeExtension(t, userExtensions[0])
	assertDirectiveNames(t, userExt.Directives, "key")

	product := requireRealWorldEntry(t, idx, "Product")
	assertFieldNames(t, product.ObjectFields(), "id", "sku", "package", "variation", "dimensions", "createdBy")

	query := requireRealWorldEntry(t, idx, "Query")
	if got := len(query.BaseDefinitions()); got != 0 {
		t.Fatalf("Query BaseDefinitions() length = %d; want 0", got)
	}
	queryExtensions := query.Extensions()
	if got := len(queryExtensions); got != 1 {
		t.Fatalf("Query Extensions() length = %d; want 1", got)
	}
	requireObjectTypeExtension(t, queryExtensions[0])
	assertFieldNames(t, query.ObjectFields(), "allProducts", "product")
}

func TestSaleorFixtureIndexesExplicitRootsAndHelpers(t *testing.T) {
	doc := realWorldFixtureDocument(t, "saleor.graphql")
	idx := schemaindex.New(doc)

	assertRootName(t, "LookupQueryRoot", idx.LookupQueryRoot(), "Query")
	assertRootName(t, "LookupMutationRoot", idx.LookupMutationRoot(), "Mutation")
	assertRootName(t, "LookupSubscriptionRoot", idx.LookupSubscriptionRoot(), "Subscription")

	webhookEventsInfo := requireDirectiveDefinition(t, doc, "webhookEventsInfo")
	asyncEvents := webhookEventsInfo.Arguments.ForName("asyncEvents")
	if asyncEvents == nil {
		t.Fatal(`webhookEventsInfo argument "asyncEvents" = nil`)
	}
	if got, want := ast.TypeString(asyncEvents.Type), "[WebhookEventTypeAsyncEnum!]!"; got != want {
		t.Fatalf("TypeString(asyncEvents.Type) = %q; want %q", got, want)
	}
	if got, want := ast.NamedTypeName(asyncEvents.Type), "WebhookEventTypeAsyncEnum"; got != want {
		t.Fatalf("NamedTypeName(asyncEvents.Type) = %q; want %q", got, want)
	}

	webhook := requireRealWorldEntry(t, idx, "Webhook")
	events := webhook.ObjectFields().ForName("events")
	if events == nil {
		t.Fatal(`Webhook field "events" = nil`)
	}
	if got, ok := ast.DirectiveStringArg(events.Directives, "deprecated", "reason"); !ok {
		t.Fatal(`DirectiveStringArg(events.Directives, "deprecated", "reason") ok = false`)
	} else if want := "Use `asyncEvents` or `syncEvents` instead."; got != want {
		t.Fatalf("DirectiveStringArg(events.Directives, deprecated, reason) = %q; want %q", got, want)
	}

	webhookCreateInput := requireRealWorldEntry(t, idx, "WebhookCreateInput")
	assertInputValueNames(t, webhookCreateInput.InputFields(), "name", "targetUrl", "events", "asyncEvents", "syncEvents", "app", "isActive", "secretKey", "query", "customHeaders")

	asyncEnum := requireRealWorldEntry(t, idx, "WebhookEventTypeAsyncEnum")
	assertEnumValueNames(t, asyncEnum.EnumValues(), "ANY_EVENTS", "ACCOUNT_CONFIRMATION_REQUESTED", "ACCOUNT_CHANGE_EMAIL_REQUESTED", "ACCOUNT_EMAIL_CHANGED")
}

func TestSWAPIFixtureIndexesExplicitRootAndRelayFields(t *testing.T) {
	idx := realWorldFixtureIndex(t, "swapi.graphql")

	root := idx.LookupQueryRoot()
	assertRootName(t, "LookupQueryRoot", root, "Root")
	assertFieldNames(t, root.ObjectFields(), "allFilms", "film", "node")

	film := requireRealWorldEntry(t, idx, "Film")
	assertNamedTypeNames(t, film.ObjectInterfaces(), "Node")

	filmsConnection := requireRealWorldEntry(t, idx, "FilmsConnection")
	assertFieldNames(t, filmsConnection.ObjectFields(), "pageInfo", "edges", "totalCount", "films")
}

func TestGitHubFixtureExercisesDeprecatedDirectiveAndUnionMembers(t *testing.T) {
	idx := realWorldFixtureIndex(t, "github.graphql")

	assignable := requireRealWorldEntry(t, idx, "Assignable")
	assertFieldNames(t, assignable.InterfaceFields(), "assignees")

	assignedEvent := requireRealWorldEntry(t, idx, "AssignedEvent")
	assertNamedTypeNames(t, assignedEvent.ObjectInterfaces(), "Node")
	user := assignedEvent.ObjectFields().ForName("user")
	if user == nil {
		t.Fatal(`AssignedEvent field "user" = nil`)
	}
	if got, ok := ast.DirectiveStringArg(user.Directives, "deprecated", "reason"); !ok {
		t.Fatal(`DirectiveStringArg(user.Directives, "deprecated", "reason") ok = false`)
	} else if want := "Assignees can now be mannequins. Use the `assignee` field instead. Removal on 2020-01-01 UTC."; got != want {
		t.Fatalf("DirectiveStringArg(user.Directives, deprecated, reason) = %q; want %q", got, want)
	}

	assignee := requireRealWorldEntry(t, idx, "Assignee")
	assertNamedTypeNames(t, assignee.UnionMembers(), "Bot", "Mannequin", "Organization", "User")
}

func TestDirectiveOnlyExtensionFixturePreservesRawDirectiveMetadata(t *testing.T) {
	idx := realWorldFixtureIndex(t, "directive-only-extension.graphql")
	catalogItem := requireRealWorldEntry(t, idx, "CatalogItem")

	assertFieldNames(t, catalogItem.ObjectFields(), "id")
	extensions := catalogItem.Extensions()
	if got := len(extensions); got != 1 {
		t.Fatalf("CatalogItem Extensions() length = %d; want 1", got)
	}
	ext := requireObjectTypeExtension(t, extensions[0])
	if got := len(ext.Fields); got != 0 {
		t.Fatalf("CatalogItem directive-only extension field count = %d; want 0", got)
	}
	assertDirectiveNames(t, ext.Directives, "tag")
	if got, ok := ast.DirectiveStringArg(ext.Directives, "tag", "name"); !ok {
		t.Fatal(`DirectiveStringArg(ext.Directives, "tag", "name") ok = false`)
	} else if want := "locally-constructed-directive-only-extension"; got != want {
		t.Fatalf("DirectiveStringArg(ext.Directives, tag, name) = %q; want %q", got, want)
	}
}

func realWorldFixtureDocument(t *testing.T, name string) *ast.Document {
	t.Helper()
	doc, err := parser.ParseSchema(readRealWorldFixture(t, name))
	if err != nil {
		t.Fatalf("ParseSchema(%s) error = %v", name, err)
	}
	return doc
}

func realWorldFixtureIndex(t *testing.T, name string) *schemaindex.Index {
	t.Helper()
	return schemaindex.New(realWorldFixtureDocument(t, name))
}

func requireRealWorldEntry(t *testing.T, idx *schemaindex.Index, name string) *schemaindex.TypeEntry {
	t.Helper()
	entry := idx.LookupType(name)
	if entry == nil {
		t.Fatalf("LookupType(%q) = nil", name)
	}
	return entry
}

func assertRootName(t *testing.T, lookup string, entry *schemaindex.TypeEntry, name string) {
	t.Helper()
	if entry == nil {
		t.Fatalf("%s() = nil", lookup)
	}
	if entry.Name() != name {
		t.Fatalf("%s().Name() = %q; want %q", lookup, entry.Name(), name)
	}
}

func requireDirectiveDefinition(t *testing.T, doc *ast.Document, name string) *ast.DirectiveDefinition {
	t.Helper()
	for _, def := range doc.Definitions {
		directiveDef, ok := def.(*ast.DirectiveDefinition)
		if ok && directiveDef.Name == name {
			return directiveDef
		}
	}
	t.Fatalf("directive definition %q = nil", name)
	return nil
}
