package ast_test

import (
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
)

var concreteNodes = []ast.Node{
	(*ast.Document)(nil),
	(*ast.OperationDefinition)(nil),
	(*ast.FragmentDefinition)(nil),
	(*ast.VariableDefinition)(nil),
	(*ast.SelectionSet)(nil),
	(*ast.Field)(nil),
	(*ast.FragmentSpread)(nil),
	(*ast.InlineFragment)(nil),
	(*ast.Argument)(nil),
	(*ast.Directive)(nil),
	(*ast.SchemaDefinition)(nil),
	(*ast.SchemaExtension)(nil),
	(*ast.OperationTypeDefinition)(nil),
	(*ast.ScalarTypeDefinition)(nil),
	(*ast.ScalarTypeExtension)(nil),
	(*ast.ObjectTypeDefinition)(nil),
	(*ast.ObjectTypeExtension)(nil),
	(*ast.InterfaceTypeDefinition)(nil),
	(*ast.InterfaceTypeExtension)(nil),
	(*ast.UnionTypeDefinition)(nil),
	(*ast.UnionTypeExtension)(nil),
	(*ast.EnumTypeDefinition)(nil),
	(*ast.EnumTypeExtension)(nil),
	(*ast.InputObjectTypeDefinition)(nil),
	(*ast.InputObjectTypeExtension)(nil),
	(*ast.FieldDefinition)(nil),
	(*ast.InputValueDefinition)(nil),
	(*ast.EnumValueDefinition)(nil),
	(*ast.DirectiveDefinition)(nil),
	(*ast.IntValue)(nil),
	(*ast.FloatValue)(nil),
	(*ast.StringValue)(nil),
	(*ast.BooleanValue)(nil),
	(*ast.NullValue)(nil),
	(*ast.EnumValue)(nil),
	(*ast.ListValue)(nil),
	(*ast.ObjectValue)(nil),
	(*ast.ObjectField)(nil),
	(*ast.Variable)(nil),
	(*ast.NamedType)(nil),
	(*ast.ListType)(nil),
	(*ast.NonNullType)(nil),
	(*ast.BadValue)(nil),
	(*ast.BadType)(nil),
	(*ast.BadField)(nil),
	(*ast.BadDefinition)(nil),
	(*ast.Comment)(nil),
}

const concreteNodeCount = 47

func TestNode_RegistryExhaustive(t *testing.T) {
	expected := structsWithGetLocMethod(t)

	actual := map[string]bool{}
	for _, n := range concreteNodes {
		actual[reflect.TypeOf(n).Elem().Name()] = true
	}

	if len(expected) != concreteNodeCount {
		t.Fatalf("source declares %d node structs; want %d (got %s)",
			len(expected), concreteNodeCount, sortedKeys(expected))
	}
	if len(concreteNodes) != concreteNodeCount {
		t.Fatalf("concreteNodes registry has %d entries; want %d", len(concreteNodes), concreteNodeCount)
	}

	for name := range expected {
		if !actual[name] {
			t.Errorf("%s has GetLoc but is missing from concreteNodes", name)
		}
	}
	for name := range actual {
		if !expected[name] {
			t.Errorf("%s is registered in concreteNodes but has no GetLoc method", name)
		}
	}
}

func structsWithGetLocMethod(t *testing.T) map[string]bool {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(astDir(t), "*.go"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	found := map[string]bool{}
	fset := token.NewFileSet()
	for _, path := range matches {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := goparser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		goast.Inspect(file, func(n goast.Node) bool {
			decl, ok := n.(*goast.FuncDecl)
			if !ok || decl.Name.Name != "GetLoc" || decl.Recv == nil || len(decl.Recv.List) != 1 {
				return true
			}
			star, ok := decl.Recv.List[0].Type.(*goast.StarExpr)
			if !ok {
				return true
			}
			ident, ok := star.X.(*goast.Ident)
			if ok {
				found[ident.Name] = true
			}
			return true
		})
	}
	return found
}

func TestChildren_OrderForNonLeafNodes(t *testing.T) {
	desc := &ast.StringValue{Value: "desc"}
	variable := &ast.Variable{Name: "v"}
	named := &ast.NamedType{Name: "T"}
	defaultValue := &ast.IntValue{Value: "1"}
	arg := &ast.Argument{Name: "arg", Value: defaultValue}
	dir := &ast.Directive{Name: "dir", Arguments: ast.ArgumentList{arg}}
	selectionSet := &ast.SelectionSet{}
	field := &ast.Field{Name: "field"}
	fragmentSpread := &ast.FragmentSpread{Name: "frag"}
	inlineFragment := &ast.InlineFragment{}
	variableDefinition := &ast.VariableDefinition{
		Variable:     variable,
		Type:         named,
		DefaultValue: defaultValue,
		Directives:   ast.DirectiveList{dir},
	}
	operationDefinition := &ast.OperationDefinition{
		VariableDefinitions: ast.VariableDefinitionList{variableDefinition},
		Directives:          ast.DirectiveList{dir},
		SelectionSet:        selectionSet,
	}
	fragmentDefinition := &ast.FragmentDefinition{
		TypeCondition: named,
		Directives:    ast.DirectiveList{dir},
		SelectionSet:  selectionSet,
	}
	objectField := &ast.ObjectField{Name: "field", Value: defaultValue}
	fieldDefinition := &ast.FieldDefinition{
		Description: desc,
		Arguments:   ast.InputValueList{},
		Type:        named,
		Directives:  ast.DirectiveList{dir},
	}
	inputValue := &ast.InputValueDefinition{
		Description:  desc,
		Type:         named,
		DefaultValue: defaultValue,
		Directives:   ast.DirectiveList{dir},
	}
	enumValue := &ast.EnumValueDefinition{Description: desc, Directives: ast.DirectiveList{dir}}
	operationType := &ast.OperationTypeDefinition{Type: named}

	tests := []struct {
		name string
		node ast.Node
		want []ast.Node
	}{
		{
			name: "Document",
			node: &ast.Document{Definitions: ast.DefinitionList{operationDefinition, fragmentDefinition}},
			want: []ast.Node{operationDefinition, fragmentDefinition},
		},
		{
			name: "OperationDefinition",
			node: operationDefinition,
			want: []ast.Node{variableDefinition, dir, selectionSet},
		},
		{
			name: "FragmentDefinition",
			node: fragmentDefinition,
			want: []ast.Node{named, dir, selectionSet},
		},
		{
			name: "VariableDefinition",
			node: variableDefinition,
			want: []ast.Node{variable, named, defaultValue, dir},
		},
		{
			name: "SelectionSet",
			node: &ast.SelectionSet{Selections: []ast.Selection{field, fragmentSpread, inlineFragment}},
			want: []ast.Node{field, fragmentSpread, inlineFragment},
		},
		{
			name: "Field",
			node: &ast.Field{Arguments: ast.ArgumentList{arg}, Directives: ast.DirectiveList{dir}, SelectionSet: selectionSet},
			want: []ast.Node{arg, dir, selectionSet},
		},
		{
			name: "FragmentSpread",
			node: &ast.FragmentSpread{Directives: ast.DirectiveList{dir}},
			want: []ast.Node{dir},
		},
		{
			name: "InlineFragment",
			node: &ast.InlineFragment{TypeCondition: named, Directives: ast.DirectiveList{dir}, SelectionSet: selectionSet},
			want: []ast.Node{named, dir, selectionSet},
		},
		{
			name: "Argument",
			node: arg,
			want: []ast.Node{defaultValue},
		},
		{
			name: "Directive",
			node: dir,
			want: []ast.Node{arg},
		},
		{
			name: "ListValue",
			node: &ast.ListValue{Values: []ast.Value{defaultValue, desc}},
			want: []ast.Node{defaultValue, desc},
		},
		{
			name: "ObjectValue",
			node: &ast.ObjectValue{Fields: []*ast.ObjectField{objectField}},
			want: []ast.Node{objectField},
		},
		{
			name: "ObjectField",
			node: objectField,
			want: []ast.Node{defaultValue},
		},
		{
			name: "ListType",
			node: &ast.ListType{OfType: named},
			want: []ast.Node{named},
		},
		{
			name: "NonNullType",
			node: &ast.NonNullType{OfType: named},
			want: []ast.Node{named},
		},
		{
			name: "SchemaDefinition",
			node: &ast.SchemaDefinition{Description: desc, Directives: ast.DirectiveList{dir}, OperationTypes: []*ast.OperationTypeDefinition{operationType}},
			want: []ast.Node{desc, dir, operationType},
		},
		{
			name: "SchemaExtension",
			node: &ast.SchemaExtension{Directives: ast.DirectiveList{dir}, OperationTypes: []*ast.OperationTypeDefinition{operationType}},
			want: []ast.Node{dir, operationType},
		},
		{
			name: "OperationTypeDefinition",
			node: operationType,
			want: []ast.Node{named},
		},
		{
			name: "ScalarTypeDefinition",
			node: &ast.ScalarTypeDefinition{Description: desc, Directives: ast.DirectiveList{dir}},
			want: []ast.Node{desc, dir},
		},
		{
			name: "ScalarTypeExtension",
			node: &ast.ScalarTypeExtension{Directives: ast.DirectiveList{dir}},
			want: []ast.Node{dir},
		},
		{
			name: "ObjectTypeDefinition",
			node: &ast.ObjectTypeDefinition{Description: desc, Interfaces: []*ast.NamedType{named}, Directives: ast.DirectiveList{dir}, Fields: ast.FieldDefinitionList{fieldDefinition}},
			want: []ast.Node{desc, named, dir, fieldDefinition},
		},
		{
			name: "ObjectTypeExtension",
			node: &ast.ObjectTypeExtension{Interfaces: []*ast.NamedType{named}, Directives: ast.DirectiveList{dir}, Fields: ast.FieldDefinitionList{fieldDefinition}},
			want: []ast.Node{named, dir, fieldDefinition},
		},
		{
			name: "InterfaceTypeDefinition",
			node: &ast.InterfaceTypeDefinition{Description: desc, Interfaces: []*ast.NamedType{named}, Directives: ast.DirectiveList{dir}, Fields: ast.FieldDefinitionList{fieldDefinition}},
			want: []ast.Node{desc, named, dir, fieldDefinition},
		},
		{
			name: "InterfaceTypeExtension",
			node: &ast.InterfaceTypeExtension{Interfaces: []*ast.NamedType{named}, Directives: ast.DirectiveList{dir}, Fields: ast.FieldDefinitionList{fieldDefinition}},
			want: []ast.Node{named, dir, fieldDefinition},
		},
		{
			name: "UnionTypeDefinition",
			node: &ast.UnionTypeDefinition{Description: desc, Directives: ast.DirectiveList{dir}, Members: []*ast.NamedType{named}},
			want: []ast.Node{desc, dir, named},
		},
		{
			name: "UnionTypeExtension",
			node: &ast.UnionTypeExtension{Directives: ast.DirectiveList{dir}, Members: []*ast.NamedType{named}},
			want: []ast.Node{dir, named},
		},
		{
			name: "EnumTypeDefinition",
			node: &ast.EnumTypeDefinition{Description: desc, Directives: ast.DirectiveList{dir}, Values: ast.EnumValueList{enumValue}},
			want: []ast.Node{desc, dir, enumValue},
		},
		{
			name: "EnumTypeExtension",
			node: &ast.EnumTypeExtension{Directives: ast.DirectiveList{dir}, Values: ast.EnumValueList{enumValue}},
			want: []ast.Node{dir, enumValue},
		},
		{
			name: "InputObjectTypeDefinition",
			node: &ast.InputObjectTypeDefinition{Description: desc, Directives: ast.DirectiveList{dir}, Fields: ast.InputValueList{inputValue}},
			want: []ast.Node{desc, dir, inputValue},
		},
		{
			name: "InputObjectTypeExtension",
			node: &ast.InputObjectTypeExtension{Directives: ast.DirectiveList{dir}, Fields: ast.InputValueList{inputValue}},
			want: []ast.Node{dir, inputValue},
		},
		{
			name: "FieldDefinition",
			node: fieldDefinition,
			want: []ast.Node{desc, named, dir},
		},
		{
			name: "InputValueDefinition",
			node: inputValue,
			want: []ast.Node{desc, named, defaultValue, dir},
		},
		{
			name: "EnumValueDefinition",
			node: enumValue,
			want: []ast.Node{desc, dir},
		},
		{
			name: "DirectiveDefinition",
			node: &ast.DirectiveDefinition{Description: desc, Arguments: ast.InputValueList{inputValue}, Locations: []string{"FIELD"}},
			want: []ast.Node{desc, inputValue},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertChildren(t, tt.node, tt.want...)
		})
	}
}

func TestChildren_SkipsNilChildren(t *testing.T) {
	value := &ast.IntValue{Value: "1"}
	node := &ast.ListValue{Values: []ast.Value{nil, value, nil}}
	assertChildren(t, node, value)
}

func TestChildren_LeavesReturnNil(t *testing.T) {
	leaves := []ast.Node{
		&ast.IntValue{},
		&ast.FloatValue{},
		&ast.StringValue{},
		&ast.BooleanValue{},
		&ast.NullValue{},
		&ast.EnumValue{},
		&ast.Variable{},
		&ast.NamedType{},
		&ast.Comment{},
		&ast.BadValue{},
		&ast.BadType{},
		&ast.BadField{},
		&ast.BadDefinition{},
	}

	for _, node := range leaves {
		if children := node.Children(); len(children) != 0 {
			t.Errorf("%T.Children() = %v; want no children", node, children)
		}
	}
}

func assertChildren(t *testing.T, node ast.Node, want ...ast.Node) {
	t.Helper()
	got := node.Children()
	if len(got) != len(want) {
		t.Fatalf("%T.Children() length = %d; want %d (%#v)", node, len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%T.Children()[%d] = %T(%p); want %T(%p)",
				node, i, got[i], got[i], want[i], want[i])
		}
	}
}
