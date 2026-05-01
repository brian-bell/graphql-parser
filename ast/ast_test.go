package ast_test

import (
	"testing"

	"github.com/bellbm/graphql-parser/ast"
)

// Compile-time interface conformance for every concrete node type.
var (
	_ ast.Node = (*ast.Comment)(nil)

	_ ast.Value = (*ast.IntValue)(nil)
	_ ast.Value = (*ast.FloatValue)(nil)
	_ ast.Value = (*ast.StringValue)(nil)
	_ ast.Value = (*ast.BooleanValue)(nil)
	_ ast.Value = (*ast.NullValue)(nil)
	_ ast.Value = (*ast.EnumValue)(nil)
	_ ast.Value = (*ast.ListValue)(nil)
	_ ast.Value = (*ast.ObjectValue)(nil)
	_ ast.Node  = (*ast.ObjectField)(nil)
	_ ast.Value = (*ast.Variable)(nil)
	_ ast.Value = (*ast.BadValue)(nil)

	_ ast.Type = (*ast.NamedType)(nil)
	_ ast.Type = (*ast.ListType)(nil)
	_ ast.Type = (*ast.NonNullType)(nil)
	_ ast.Type = (*ast.BadType)(nil)

	_ ast.Selection  = (*ast.BadField)(nil)
	_ ast.Definition = (*ast.BadDefinition)(nil)
)

func TestNode_GetLocReturnsField(t *testing.T) {
	loc := &ast.Loc{Start: 1, End: 4}
	cases := []ast.Node{
		&ast.IntValue{Value: "1", Loc: loc},
		&ast.FloatValue{Value: "1.0", Loc: loc},
		&ast.StringValue{Value: "x", Loc: loc},
		&ast.BooleanValue{Value: true, Loc: loc},
		&ast.NullValue{Loc: loc},
		&ast.EnumValue{Value: "X", Loc: loc},
		&ast.ListValue{Loc: loc},
		&ast.ObjectValue{Loc: loc},
		&ast.ObjectField{Name: "a", Loc: loc},
		&ast.Variable{Name: "v", Loc: loc},
		&ast.NamedType{Name: "T", Loc: loc},
		&ast.ListType{Loc: loc},
		&ast.NonNullType{Loc: loc},
		&ast.BadValue{Loc: loc},
		&ast.BadType{Loc: loc},
		&ast.BadField{Loc: loc},
		&ast.BadDefinition{Loc: loc},
		&ast.Comment{Text: "x", Loc: loc},
	}
	for i, n := range cases {
		if got := n.GetLoc(); got != loc {
			t.Errorf("node %d (%T).GetLoc() != stored Loc", i, n)
		}
	}
}

func TestNode_NilLoc(t *testing.T) {
	// Synthetic nodes with nil Loc are legal.
	v := &ast.IntValue{Value: "0"}
	if v.GetLoc() != nil {
		t.Errorf("expected nil Loc, got %v", v.GetLoc())
	}
}

func TestCommentGroup_ZeroValue(t *testing.T) {
	var g ast.CommentGroup
	if len(g.Leading) != 0 || len(g.Trailing) != 0 {
		t.Error("zero-value CommentGroup should have empty lists")
	}
}
