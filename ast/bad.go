package ast

// BadValue is a placeholder Value produced by the parser in WithRecovery mode
// when a value expression failed to parse. Err captures why; Loc covers the
// span of source the parser consumed before resynchronizing.
type BadValue struct {
	Err *SyntaxError
	Loc *Loc
}

func (v *BadValue) GetLoc() *Loc   { return v.Loc }
func (*BadValue) Children() []Node { return nil }
func (*BadValue) isValue()         {}

// BadType is a placeholder Type for recovery.
type BadType struct {
	Err *SyntaxError
	Loc *Loc
}

func (t *BadType) GetLoc() *Loc   { return t.Loc }
func (*BadType) Children() []Node { return nil }
func (*BadType) isType()          {}

// BadField is a placeholder Selection for recovery — used when a selection
// inside a SelectionSet failed to parse.
type BadField struct {
	Err *SyntaxError
	Loc *Loc
}

func (f *BadField) GetLoc() *Loc   { return f.Loc }
func (*BadField) Children() []Node { return nil }
func (*BadField) isSelection()     {}

// BadDefinition is a placeholder top-level Definition for recovery.
type BadDefinition struct {
	Err *SyntaxError
	Loc *Loc
}

func (d *BadDefinition) GetLoc() *Loc   { return d.Loc }
func (*BadDefinition) Children() []Node { return nil }
func (*BadDefinition) isDefinition()    {}
