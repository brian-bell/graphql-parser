package parser_test

import (
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/parser"
)

func TestParseValue_Int(t *testing.T) {
	v, err := parser.ParseValue("42")
	if err != nil {
		t.Fatal(err)
	}
	iv, ok := v.(*ast.IntValue)
	if !ok {
		t.Fatalf("got %T; want *ast.IntValue", v)
	}
	if iv.Value != "42" {
		t.Errorf("Value = %q; want 42", iv.Value)
	}
}

func TestParseValue_Float(t *testing.T) {
	v, err := parser.ParseValue("3.14")
	if err != nil {
		t.Fatal(err)
	}
	fv, ok := v.(*ast.FloatValue)
	if !ok {
		t.Fatalf("got %T; want *ast.FloatValue", v)
	}
	if fv.Value != "3.14" {
		t.Errorf("Value = %q; want 3.14", fv.Value)
	}
}

func TestParseValue_String(t *testing.T) {
	v, err := parser.ParseValue(`"hello"`)
	if err != nil {
		t.Fatal(err)
	}
	sv, ok := v.(*ast.StringValue)
	if !ok {
		t.Fatalf("got %T; want *ast.StringValue", v)
	}
	if sv.Value != "hello" || sv.Block {
		t.Errorf("got Value=%q Block=%v; want hello/false", sv.Value, sv.Block)
	}
}

func TestParseValue_BlockString(t *testing.T) {
	v, err := parser.ParseValue(`"""hi"""`)
	if err != nil {
		t.Fatal(err)
	}
	sv, ok := v.(*ast.StringValue)
	if !ok {
		t.Fatalf("got %T", v)
	}
	if !sv.Block || sv.Value != "hi" {
		t.Errorf("got Value=%q Block=%v; want hi/true", sv.Value, sv.Block)
	}
}

func TestParseValue_Boolean(t *testing.T) {
	for _, c := range []struct {
		body string
		want bool
	}{{"true", true}, {"false", false}} {
		v, err := parser.ParseValue(c.body)
		if err != nil {
			t.Fatal(err)
		}
		bv, ok := v.(*ast.BooleanValue)
		if !ok {
			t.Fatalf("body %q: got %T", c.body, v)
		}
		if bv.Value != c.want {
			t.Errorf("body %q: got %v; want %v", c.body, bv.Value, c.want)
		}
	}
}

func TestParseValue_Null(t *testing.T) {
	v, err := parser.ParseValue("null")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := v.(*ast.NullValue); !ok {
		t.Fatalf("got %T; want *ast.NullValue", v)
	}
}

func TestParseValue_Enum(t *testing.T) {
	v, err := parser.ParseValue("RED")
	if err != nil {
		t.Fatal(err)
	}
	ev, ok := v.(*ast.EnumValue)
	if !ok {
		t.Fatalf("got %T; want *ast.EnumValue", v)
	}
	if ev.Value != "RED" {
		t.Errorf("Value = %q; want RED", ev.Value)
	}
}

func TestParseValue_EnumKeywordsExcluded(t *testing.T) {
	// true, false, null are NOT enum values.
	for _, body := range []string{"true", "false", "null"} {
		v, err := parser.ParseValue(body)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := v.(*ast.EnumValue); ok {
			t.Errorf("body %q parsed as EnumValue; should be Boolean/Null", body)
		}
	}
}

func TestParseValue_List(t *testing.T) {
	v, err := parser.ParseValue("[1, 2, 3]")
	if err != nil {
		t.Fatal(err)
	}
	lv, ok := v.(*ast.ListValue)
	if !ok {
		t.Fatalf("got %T; want *ast.ListValue", v)
	}
	if len(lv.Values) != 3 {
		t.Fatalf("len = %d; want 3", len(lv.Values))
	}
	for i, want := range []string{"1", "2", "3"} {
		iv, ok := lv.Values[i].(*ast.IntValue)
		if !ok || iv.Value != want {
			t.Errorf("element %d: got %T %v; want IntValue %q", i, lv.Values[i], lv.Values[i], want)
		}
	}
}

func TestParseValue_ListEmpty(t *testing.T) {
	v, err := parser.ParseValue("[]")
	if err != nil {
		t.Fatal(err)
	}
	lv, ok := v.(*ast.ListValue)
	if !ok {
		t.Fatalf("got %T", v)
	}
	if len(lv.Values) != 0 {
		t.Errorf("len = %d; want 0", len(lv.Values))
	}
}

func TestParseValue_ListNested(t *testing.T) {
	v, err := parser.ParseValue("[[1], [2, 3]]")
	if err != nil {
		t.Fatal(err)
	}
	lv := v.(*ast.ListValue)
	if len(lv.Values) != 2 {
		t.Fatalf("outer len = %d; want 2", len(lv.Values))
	}
	inner := lv.Values[1].(*ast.ListValue)
	if len(inner.Values) != 2 {
		t.Errorf("inner len = %d; want 2", len(inner.Values))
	}
}

func TestParseValue_Object(t *testing.T) {
	v, err := parser.ParseValue(`{a: 1, b: "hi", c: null}`)
	if err != nil {
		t.Fatal(err)
	}
	ov, ok := v.(*ast.ObjectValue)
	if !ok {
		t.Fatalf("got %T; want *ast.ObjectValue", v)
	}
	if len(ov.Fields) != 3 {
		t.Fatalf("len = %d; want 3", len(ov.Fields))
	}
	for i, want := range []string{"a", "b", "c"} {
		if ov.Fields[i].Name != want {
			t.Errorf("field %d: name = %q; want %q", i, ov.Fields[i].Name, want)
		}
	}
}

func TestParseValue_ObjectEmpty(t *testing.T) {
	v, err := parser.ParseValue("{}")
	if err != nil {
		t.Fatal(err)
	}
	ov := v.(*ast.ObjectValue)
	if len(ov.Fields) != 0 {
		t.Errorf("len = %d; want 0", len(ov.Fields))
	}
}

func TestParseValue_Variable(t *testing.T) {
	v, err := parser.ParseValue("$x")
	if err != nil {
		t.Fatal(err)
	}
	vv, ok := v.(*ast.Variable)
	if !ok {
		t.Fatalf("got %T; want *ast.Variable", v)
	}
	if vv.Name != "x" {
		t.Errorf("Name = %q; want x", vv.Name)
	}
}

func TestParseConstValue_RejectsVariable(t *testing.T) {
	if _, err := parser.ParseConstValue("$x"); err == nil {
		t.Error("expected error for variable in const context")
	}
}

func TestParseConstValue_RejectsNestedVariable(t *testing.T) {
	cases := []string{"[$x]", "{a: $x}", "[1, $x, 3]"}
	for _, body := range cases {
		t.Run(body, func(t *testing.T) {
			if _, err := parser.ParseConstValue(body); err == nil {
				t.Errorf("body %q: expected error", body)
			}
		})
	}
}

func TestParseValue_NegativeNumber(t *testing.T) {
	v, err := parser.ParseValue("-42")
	if err != nil {
		t.Fatal(err)
	}
	if iv, ok := v.(*ast.IntValue); !ok || iv.Value != "-42" {
		t.Errorf("got %T %v; want IntValue -42", v, v)
	}
}

func TestParseValue_LeftoverInputIsError(t *testing.T) {
	if _, err := parser.ParseValue("1 2"); err == nil {
		t.Error("expected error for trailing input")
	}
}

func TestParseValue_EmptyInputIsError(t *testing.T) {
	if _, err := parser.ParseValue(""); err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseValue_LocationCoversWholeValue(t *testing.T) {
	v, err := parser.ParseValue("[1, 2, 3]")
	if err != nil {
		t.Fatal(err)
	}
	loc := v.GetLoc()
	if loc == nil {
		t.Fatal("Loc is nil")
	}
	if loc.Start != 0 || loc.End != 9 {
		t.Errorf("Loc = [%d, %d); want [0, 9)", loc.Start, loc.End)
	}
}

func TestParseType_Named(t *testing.T) {
	tt, err := parser.ParseType("User")
	if err != nil {
		t.Fatal(err)
	}
	nt, ok := tt.(*ast.NamedType)
	if !ok {
		t.Fatalf("got %T; want *ast.NamedType", tt)
	}
	if nt.Name != "User" {
		t.Errorf("Name = %q; want User", nt.Name)
	}
}

func TestParseType_List(t *testing.T) {
	tt, err := parser.ParseType("[String]")
	if err != nil {
		t.Fatal(err)
	}
	lt, ok := tt.(*ast.ListType)
	if !ok {
		t.Fatalf("got %T; want *ast.ListType", tt)
	}
	if nt, ok := lt.OfType.(*ast.NamedType); !ok || nt.Name != "String" {
		t.Errorf("inner: got %T; want NamedType String", lt.OfType)
	}
}

func TestParseType_NonNull(t *testing.T) {
	tt, err := parser.ParseType("Int!")
	if err != nil {
		t.Fatal(err)
	}
	nn, ok := tt.(*ast.NonNullType)
	if !ok {
		t.Fatalf("got %T; want *ast.NonNullType", tt)
	}
	if nt, ok := nn.OfType.(*ast.NamedType); !ok || nt.Name != "Int" {
		t.Errorf("inner: got %T; want NamedType Int", nn.OfType)
	}
}

func TestParseType_ListOfNonNull(t *testing.T) {
	// [Int!] — list of non-null int
	tt, err := parser.ParseType("[Int!]")
	if err != nil {
		t.Fatal(err)
	}
	lt := tt.(*ast.ListType)
	nn := lt.OfType.(*ast.NonNullType)
	if nt := nn.OfType.(*ast.NamedType); nt.Name != "Int" {
		t.Errorf("inner: %v", nt)
	}
}

func TestParseType_NonNullList(t *testing.T) {
	// [Int]! — non-null list of nullable int
	tt, err := parser.ParseType("[Int]!")
	if err != nil {
		t.Fatal(err)
	}
	nn := tt.(*ast.NonNullType)
	lt := nn.OfType.(*ast.ListType)
	if nt := lt.OfType.(*ast.NamedType); nt.Name != "Int" {
		t.Errorf("inner: %v", nt)
	}
}

func TestParseType_DoubleNonNullRejected(t *testing.T) {
	// Int!! is not a valid type per spec.
	if _, err := parser.ParseType("Int!!"); err == nil {
		t.Error("expected error for Int!!")
	}
}

func TestParseType_UnclosedListRejected(t *testing.T) {
	for _, body := range []string{"[", "[Int", "[Int]]"} {
		t.Run(body, func(t *testing.T) {
			if _, err := parser.ParseType(body); err == nil {
				t.Errorf("body %q: expected error", body)
			}
		})
	}
}

func TestParseType_LocationCoversWholeType(t *testing.T) {
	tt, err := parser.ParseType("[Int!]!")
	if err != nil {
		t.Fatal(err)
	}
	loc := tt.GetLoc()
	if loc == nil {
		t.Fatal("nil Loc")
	}
	if loc.Start != 0 || loc.End != 7 {
		t.Errorf("Loc = [%d, %d); want [0, 7)", loc.Start, loc.End)
	}
}

func TestParseType_EmptyInputIsError(t *testing.T) {
	if _, err := parser.ParseType(""); err == nil {
		t.Error("expected error for empty input")
	}
}
