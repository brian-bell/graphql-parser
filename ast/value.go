package ast

// IntValue is an integer literal. Value is the raw lexeme (e.g. "123",
// "-1") so that callers can choose how to convert it (strconv.ParseInt,
// math/big, etc.) — the spec leaves the integer range up to the consumer.
type IntValue struct {
	Value    string
	Loc      *Loc
	Comments *CommentGroup
}

func (v *IntValue) GetLoc() *Loc { return v.Loc }
func (*IntValue) isValue()       {}

// FloatValue is a floating-point literal. Value is the raw lexeme.
type FloatValue struct {
	Value    string
	Loc      *Loc
	Comments *CommentGroup
}

func (v *FloatValue) GetLoc() *Loc { return v.Loc }
func (*FloatValue) isValue()       {}

// StringValue is a string literal. Value is the post-decode value: escape
// sequences are processed for "..." and the BlockStringValue dedent is
// applied for """...""". The raw delimited bytes can be recovered via
// Loc.Source.Body[Loc.Start:Loc.End].
type StringValue struct {
	Value    string
	Block    bool // true if written as """..."""
	Loc      *Loc
	Comments *CommentGroup
}

func (v *StringValue) GetLoc() *Loc { return v.Loc }
func (*StringValue) isValue()       {}

// BooleanValue is one of the keywords true or false.
type BooleanValue struct {
	Value    bool
	Loc      *Loc
	Comments *CommentGroup
}

func (v *BooleanValue) GetLoc() *Loc { return v.Loc }
func (*BooleanValue) isValue()       {}

// NullValue is the null keyword.
type NullValue struct {
	Loc      *Loc
	Comments *CommentGroup
}

func (v *NullValue) GetLoc() *Loc { return v.Loc }
func (*NullValue) isValue()       {}

// EnumValue is an enum literal — a Name that is not true, false, or null.
type EnumValue struct {
	Value    string
	Loc      *Loc
	Comments *CommentGroup
}

func (v *EnumValue) GetLoc() *Loc { return v.Loc }
func (*EnumValue) isValue()       {}

// ListValue is a [...] list literal.
type ListValue struct {
	Values   []Value
	Loc      *Loc
	Comments *CommentGroup
}

func (v *ListValue) GetLoc() *Loc { return v.Loc }
func (*ListValue) isValue()       {}

// ObjectValue is a {field: value, ...} object literal.
type ObjectValue struct {
	Fields   []*ObjectField
	Loc      *Loc
	Comments *CommentGroup
}

func (v *ObjectValue) GetLoc() *Loc { return v.Loc }
func (*ObjectValue) isValue()       {}

// ObjectField is one Name: Value entry inside an ObjectValue.
type ObjectField struct {
	Name     string
	Value    Value
	Loc      *Loc
	Comments *CommentGroup
}

func (f *ObjectField) GetLoc() *Loc { return f.Loc }

// Variable is a $name reference. It implements Value but the parser refuses
// to produce one inside a const-value context (default values, directive
// arguments on type definitions); see parser.ParseConstValue.
type Variable struct {
	Name     string
	Loc      *Loc
	Comments *CommentGroup
}

func (v *Variable) GetLoc() *Loc { return v.Loc }
func (*Variable) isValue()       {}
