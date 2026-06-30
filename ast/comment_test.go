package ast_test

import (
	"go/ast"
	goparser "go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"testing"

	gast "github.com/brian-bell/graphql-parser/ast"
)

// commentedNodes registers every struct that carries a Comments *CommentGroup
// field. The slice's element type, gast.CommentedNode, is the compile-time
// proof that each listed type implements the interface: if a struct lacks the
// CommentSlot method, its line here fails to compile. The parity test below
// proves this list is exhaustive against the source.
var commentedNodes = []gast.CommentedNode{
	(*gast.Document)(nil),
	(*gast.OperationDefinition)(nil),
	(*gast.FragmentDefinition)(nil),
	(*gast.VariableDefinition)(nil),
	(*gast.SelectionSet)(nil),
	(*gast.Field)(nil),
	(*gast.FragmentSpread)(nil),
	(*gast.InlineFragment)(nil),
	(*gast.Argument)(nil),
	(*gast.Directive)(nil),
	(*gast.SchemaDefinition)(nil),
	(*gast.SchemaExtension)(nil),
	(*gast.OperationTypeDefinition)(nil),
	(*gast.ScalarTypeDefinition)(nil),
	(*gast.ScalarTypeExtension)(nil),
	(*gast.ObjectTypeDefinition)(nil),
	(*gast.ObjectTypeExtension)(nil),
	(*gast.InterfaceTypeDefinition)(nil),
	(*gast.InterfaceTypeExtension)(nil),
	(*gast.UnionTypeDefinition)(nil),
	(*gast.UnionTypeExtension)(nil),
	(*gast.EnumTypeDefinition)(nil),
	(*gast.EnumTypeExtension)(nil),
	(*gast.InputObjectTypeDefinition)(nil),
	(*gast.InputObjectTypeExtension)(nil),
	(*gast.FieldDefinition)(nil),
	(*gast.InputValueDefinition)(nil),
	(*gast.EnumValueDefinition)(nil),
	(*gast.DirectiveDefinition)(nil),
	(*gast.IntValue)(nil),
	(*gast.FloatValue)(nil),
	(*gast.StringValue)(nil),
	(*gast.BooleanValue)(nil),
	(*gast.NullValue)(nil),
	(*gast.EnumValue)(nil),
	(*gast.ListValue)(nil),
	(*gast.ObjectValue)(nil),
	(*gast.ObjectField)(nil),
	(*gast.Variable)(nil),
}

// commentedNodeCount pins the expected number of Comments-bearing structs. Both
// the source-derived "expected" set and the registry "actual" set must hit this
// count before the set comparison runs, so the test fails loudly rather than
// degrading into two empty sets that compare equal if the glob under-collects.
const commentedNodeCount = 39

// astDir resolves the directory holding the production ast sources relative to
// this test file, independent of the process working directory.
func astDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(thisFile)
}

// structsWithCommentsField parses the non-test ast sources and returns the set
// of struct type names that declare a Comments *CommentGroup field.
func structsWithCommentsField(t *testing.T) map[string]bool {
	t.Helper()
	dir := astDir(t)
	matches, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	found := map[string]bool{}
	fset := token.NewFileSet()
	for _, path := range matches {
		if filepath.Base(path) != "" && len(path) >= len("_test.go") &&
			path[len(path)-len("_test.go"):] == "_test.go" {
			continue
		}
		file, err := goparser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				return true
			}
			for _, field := range st.Fields.List {
				if !isCommentGroupPtr(field.Type) {
					continue
				}
				for _, name := range field.Names {
					if name.Name == "Comments" {
						found[ts.Name.Name] = true
					}
				}
			}
			return true
		})
	}
	return found
}

// isCommentGroupPtr reports whether expr is *CommentGroup.
func isCommentGroupPtr(expr ast.Expr) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}
	ident, ok := star.X.(*ast.Ident)
	return ok && ident.Name == "CommentGroup"
}

func TestCommentedNode_RegistryExhaustive(t *testing.T) {
	expected := structsWithCommentsField(t)

	actual := map[string]bool{}
	for _, n := range commentedNodes {
		actual[reflect.TypeOf(n).Elem().Name()] = true
	}

	if len(expected) != commentedNodeCount {
		t.Fatalf("source declares %d Comments-bearing structs; want %d (glob under/over-collected: %s)",
			len(expected), commentedNodeCount, sortedKeys(expected))
	}
	if len(commentedNodes) != commentedNodeCount {
		t.Fatalf("commentedNodes registry has %d entries; want %d", len(commentedNodes), commentedNodeCount)
	}

	for name := range expected {
		if !actual[name] {
			t.Errorf("%s has a Comments field but is missing from the commentedNodes registry", name)
		}
	}
	for name := range actual {
		if !expected[name] {
			t.Errorf("%s is registered in commentedNodes but has no Comments field in source", name)
		}
	}
}

func TestCommentedNode_SlotIdentity(t *testing.T) {
	for _, n := range commentedNodes {
		typ := reflect.TypeOf(n).Elem()
		fresh := reflect.New(typ) // *T, addressable
		cn, ok := fresh.Interface().(gast.CommentedNode)
		if !ok {
			t.Fatalf("%s does not implement CommentedNode", typ.Name())
		}
		slot := cn.CommentSlot()
		group := &gast.CommentGroup{}
		*slot = group

		field := fresh.Elem().FieldByName("Comments")
		if !field.IsValid() {
			t.Fatalf("%s has no Comments field", typ.Name())
		}
		if field.IsNil() {
			t.Errorf("%s.CommentSlot did not write through to the Comments field", typ.Name())
		}
		if field.Interface().(*gast.CommentGroup) != group {
			t.Errorf("%s.CommentSlot returned a pointer to a different field than Comments", typ.Name())
		}
	}
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
