package ast

// NamedType is a bare type name like "User" or "Int".
type NamedType struct {
	Name string
	Loc  *Loc
}

func (t *NamedType) GetLoc() *Loc   { return t.Loc }
func (*NamedType) Children() []Node { return nil }
func (*NamedType) isType()          {}

// ListType is a [T] type. OfType is the inner type, which may itself be any
// Type (NamedType, ListType, or NonNullType).
type ListType struct {
	OfType Type
	Loc    *Loc
}

func (t *ListType) GetLoc() *Loc { return t.Loc }
func (t *ListType) Children() []Node {
	if t.OfType == nil {
		return nil
	}
	return []Node{t.OfType}
}
func (*ListType) isType() {}

// NonNullType is a T! type. Per spec, OfType must be a NamedType or ListType,
// not another NonNullType — the parser enforces this.
type NonNullType struct {
	OfType Type
	Loc    *Loc
}

func (t *NonNullType) GetLoc() *Loc { return t.Loc }
func (t *NonNullType) Children() []Node {
	if t.OfType == nil {
		return nil
	}
	return []Node{t.OfType}
}
func (*NonNullType) isType() {}

// TypeString returns the canonical type-reference spelling for t, or "" when
// t is nil, a BadType, or a malformed type reference.
func TypeString(t Type) string {
	switch t := t.(type) {
	case *NamedType:
		if t == nil {
			return ""
		}
		return t.Name
	case *ListType:
		if t == nil {
			return ""
		}
		inner := TypeString(t.OfType)
		if inner == "" {
			return ""
		}
		return "[" + inner + "]"
	case *NonNullType:
		if t == nil {
			return ""
		}
		if _, ok := t.OfType.(*NonNullType); ok {
			return ""
		}
		inner := TypeString(t.OfType)
		if inner == "" {
			return ""
		}
		return inner + "!"
	default:
		return ""
	}
}

// NamedTypeName returns the innermost named type name for t, or "" when t is
// nil, a BadType, or a malformed type reference.
func NamedTypeName(t Type) string {
	switch t := t.(type) {
	case *NamedType:
		if t == nil {
			return ""
		}
		return t.Name
	case *ListType:
		if t == nil {
			return ""
		}
		return NamedTypeName(t.OfType)
	case *NonNullType:
		if t == nil {
			return ""
		}
		if _, ok := t.OfType.(*NonNullType); ok {
			return ""
		}
		return NamedTypeName(t.OfType)
	default:
		return ""
	}
}
