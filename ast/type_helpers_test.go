package ast_test

import (
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/parser"
)

func TestTypeString_RendersWrappedTypes(t *testing.T) {
	cases := []string{
		"Book",
		"Book!",
		"[Book]",
		"[Book!]!",
		"[[Book!]!]",
	}
	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			typ, err := parser.ParseType(input)
			if err != nil {
				t.Fatalf("ParseType(%q): %v", input, err)
			}
			if got := ast.TypeString(typ); got != input {
				t.Fatalf("TypeString(%q) = %q, want %q", input, got, input)
			}
		})
	}
}

func TestNamedTypeName_UnwrapsWrappedTypes(t *testing.T) {
	cases := []string{
		"Book",
		"Book!",
		"[Book]",
		"[Book!]!",
		"[[Book!]!]",
	}
	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			typ, err := parser.ParseType(input)
			if err != nil {
				t.Fatalf("ParseType(%q): %v", input, err)
			}
			if got := ast.NamedTypeName(typ); got != "Book" {
				t.Fatalf("NamedTypeName(%q) = %q, want %q", input, got, "Book")
			}
		})
	}
}

func TestTypeHelpers_ReturnEmptyForInvalidTypes(t *testing.T) {
	var nilNamed *ast.NamedType
	var nilList *ast.ListType
	var nilNonNull *ast.NonNullType

	cases := []struct {
		name string
		typ  ast.Type
	}{
		{name: "nil"},
		{name: "typed nil named", typ: nilNamed},
		{name: "typed nil list", typ: nilList},
		{name: "typed nil non-null", typ: nilNonNull},
		{name: "bad type", typ: &ast.BadType{}},
		{name: "list with nil inner", typ: &ast.ListType{}},
		{name: "non-null with nil inner", typ: &ast.NonNullType{}},
		{
			name: "double non-null",
			typ:  &ast.NonNullType{OfType: &ast.NonNullType{OfType: &ast.NamedType{Name: "Book"}}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ast.TypeString(tc.typ); got != "" {
				t.Fatalf("TypeString() = %q, want empty string", got)
			}
			if got := ast.NamedTypeName(tc.typ); got != "" {
				t.Fatalf("NamedTypeName() = %q, want empty string", got)
			}
		})
	}
}
