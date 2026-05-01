package ast

// NamedType is a bare type name like "User" or "Int".
type NamedType struct {
	Name string
	Loc  *Loc
}

func (t *NamedType) GetLoc() *Loc { return t.Loc }
func (*NamedType) isType()        {}

// ListType is a [T] type. OfType is the inner type, which may itself be any
// Type (NamedType, ListType, or NonNullType).
type ListType struct {
	OfType Type
	Loc    *Loc
}

func (t *ListType) GetLoc() *Loc { return t.Loc }
func (*ListType) isType()        {}

// NonNullType is a T! type. Per spec, OfType must be a NamedType or ListType,
// not another NonNullType — the parser enforces this.
type NonNullType struct {
	OfType Type
	Loc    *Loc
}

func (t *NonNullType) GetLoc() *Loc { return t.Loc }
func (*NonNullType) isType()        {}
