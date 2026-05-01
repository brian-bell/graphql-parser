package ast

// OperationType identifies which root operation this is (query, mutation, or
// subscription). A bare "{ ... }" shorthand operation is parsed as an
// OperationDefinition with Operation == OperationQuery.
type OperationType string

// Operation type constants.
const (
	OperationQuery        OperationType = "query"
	OperationMutation     OperationType = "mutation"
	OperationSubscription OperationType = "subscription"
)

// OperationDefinition is a query, mutation, or subscription. Anonymous
// operations (the "{ ... }" shorthand) have an empty Name and no
// VariableDefinitions or Directives.
type OperationDefinition struct {
	Operation           OperationType
	Name                string
	VariableDefinitions VariableDefinitionList
	Directives          DirectiveList
	SelectionSet        *SelectionSet
	Loc                 *Loc
	Comments            *CommentGroup
}

func (d *OperationDefinition) GetLoc() *Loc { return d.Loc }
func (*OperationDefinition) isDefinition()  {}

// FragmentDefinition is a named fragment: "fragment Name on Type { ... }".
type FragmentDefinition struct {
	Name          string
	TypeCondition *NamedType
	Directives    DirectiveList
	SelectionSet  *SelectionSet
	Loc           *Loc
	Comments      *CommentGroup
}

func (d *FragmentDefinition) GetLoc() *Loc { return d.Loc }
func (*FragmentDefinition) isDefinition()  {}

// VariableDefinition declares an operation variable: "$name: Type = default
// @directive...".
type VariableDefinition struct {
	Variable     *Variable
	Type         Type
	DefaultValue Value // const value; nil if no default
	Directives   DirectiveList
	Loc          *Loc
	Comments     *CommentGroup
}

func (v *VariableDefinition) GetLoc() *Loc { return v.Loc }

// SelectionSet is a "{ ... }" block of one or more Selections.
type SelectionSet struct {
	Selections []Selection
	Loc        *Loc
	Comments   *CommentGroup
}

// GetLoc returns the location covering the selection set including its braces.
func (s *SelectionSet) GetLoc() *Loc { return s.Loc }

// Field is a leaf or non-leaf selection: "[alias:] name(args)? @dir...?
// SelectionSet?".
type Field struct {
	Alias        string // empty when no alias was written
	Name         string
	Arguments    ArgumentList
	Directives   DirectiveList
	SelectionSet *SelectionSet // nil for leaf fields
	Loc          *Loc
	Comments     *CommentGroup
}

func (f *Field) GetLoc() *Loc { return f.Loc }
func (*Field) isSelection()   {}

// FragmentSpread is "...FragmentName Directives?". The name must not be "on";
// the parser disambiguates against InlineFragment.
type FragmentSpread struct {
	Name       string
	Directives DirectiveList
	Loc        *Loc
	Comments   *CommentGroup
}

func (s *FragmentSpread) GetLoc() *Loc { return s.Loc }
func (*FragmentSpread) isSelection()   {}

// InlineFragment is "... TypeCondition? Directives? SelectionSet". The
// type condition is optional; when absent, the fragment inherits the
// parent's type.
type InlineFragment struct {
	TypeCondition *NamedType // nil when omitted
	Directives    DirectiveList
	SelectionSet  *SelectionSet
	Loc           *Loc
	Comments      *CommentGroup
}

func (f *InlineFragment) GetLoc() *Loc { return f.Loc }
func (*InlineFragment) isSelection()   {}

// Argument is one "name: value" entry in a function-style argument list.
type Argument struct {
	Name     string
	Value    Value
	Loc      *Loc
	Comments *CommentGroup
}

func (a *Argument) GetLoc() *Loc { return a.Loc }

// Directive is "@name(args)?". Arguments must be const values when the
// directive appears on a type-system definition; the parser enforces this
// based on call site, not on the AST type.
type Directive struct {
	Name      string
	Arguments ArgumentList
	Loc       *Loc
	Comments  *CommentGroup
}

func (d *Directive) GetLoc() *Loc { return d.Loc }
