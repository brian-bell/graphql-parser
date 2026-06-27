package ast_test

import (
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/parser"
)

func TestDirectiveStringArg_ReturnsStringValue(t *testing.T) {
	doc, err := parser.ParseSchema(`
		enum Color {
			BLUE @deprecated(reason: "use INDIGO")
			EMPTY @deprecated(reason: "")
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	enumDef := doc.Definitions[0].(*ast.EnumTypeDefinition)

	blue := enumDef.Values.ForName("BLUE")
	got, ok := ast.DirectiveStringArg(blue.Directives, "deprecated", "reason")
	if got != "use INDIGO" || !ok {
		t.Fatalf("DirectiveStringArg(BLUE) = %q, %v; want %q, true", got, ok, "use INDIGO")
	}

	empty := enumDef.Values.ForName("EMPTY")
	got, ok = ast.DirectiveStringArg(empty.Directives, "deprecated", "reason")
	if got != "" || !ok {
		t.Fatalf("DirectiveStringArg(EMPTY) = %q, %v; want empty string, true", got, ok)
	}
}

func TestDirectiveStringArg_ReturnsFalseForMissingNilOrNonString(t *testing.T) {
	doc, err := parser.ParseSchema(`
		enum Color {
			RED @deprecated
			GREEN
			BLUE @deprecated(reason: 123)
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	enumDef := doc.Definitions[0].(*ast.EnumTypeDefinition)

	cases := []struct {
		name string
		dirs ast.DirectiveList
		arg  string
	}{
		{name: "missing argument", dirs: enumDef.Values.ForName("RED").Directives, arg: "reason"},
		{name: "missing directive", dirs: enumDef.Values.ForName("GREEN").Directives, arg: "reason"},
		{name: "non-string argument", dirs: enumDef.Values.ForName("BLUE").Directives, arg: "reason"},
		{name: "wrong argument name", dirs: enumDef.Values.ForName("BLUE").Directives, arg: "missing"},
		{
			name: "nil entries",
			dirs: ast.DirectiveList{
				nil,
				&ast.Directive{
					Name: "deprecated",
					Arguments: ast.ArgumentList{
						nil,
						&ast.Argument{Name: "reason", Value: &ast.IntValue{Value: "123"}},
					},
				},
			},
			arg: "reason",
		},
		{
			name: "nil argument value",
			dirs: ast.DirectiveList{
				&ast.Directive{
					Name: "deprecated",
					Arguments: ast.ArgumentList{
						&ast.Argument{Name: "reason"},
					},
				},
			},
			arg: "reason",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ast.DirectiveStringArg(tc.dirs, "deprecated", tc.arg)
			if got != "" || ok {
				t.Fatalf("DirectiveStringArg() = %q, %v; want empty string, false", got, ok)
			}
		})
	}
}
