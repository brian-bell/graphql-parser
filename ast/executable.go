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
func (d *OperationDefinition) Children() []Node {
	children := make([]Node, 0, len(d.VariableDefinitions)+len(d.Directives)+1)
	for _, vd := range d.VariableDefinitions {
		if vd != nil {
			children = append(children, vd)
		}
	}
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	if d.SelectionSet != nil {
		children = append(children, d.SelectionSet)
	}
	return children
}
func (*OperationDefinition) isDefinition()                 {}
func (d *OperationDefinition) CommentSlot() **CommentGroup { return &d.Comments }

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
func (d *FragmentDefinition) Children() []Node {
	children := make([]Node, 0, len(d.Directives)+2)
	if d.TypeCondition != nil {
		children = append(children, d.TypeCondition)
	}
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	if d.SelectionSet != nil {
		children = append(children, d.SelectionSet)
	}
	return children
}
func (*FragmentDefinition) isDefinition()                 {}
func (d *FragmentDefinition) CommentSlot() **CommentGroup { return &d.Comments }

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
func (v *VariableDefinition) Children() []Node {
	children := make([]Node, 0, len(v.Directives)+3)
	if v.Variable != nil {
		children = append(children, v.Variable)
	}
	if v.Type != nil {
		children = append(children, v.Type)
	}
	if v.DefaultValue != nil {
		children = append(children, v.DefaultValue)
	}
	for _, dir := range v.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	return children
}
func (v *VariableDefinition) CommentSlot() **CommentGroup { return &v.Comments }

// SelectionSet is a "{ ... }" block of one or more Selections.
type SelectionSet struct {
	Selections []Selection
	Loc        *Loc
	Comments   *CommentGroup
}

// GetLoc returns the location covering the selection set including its braces.
func (s *SelectionSet) GetLoc() *Loc { return s.Loc }
func (s *SelectionSet) Children() []Node {
	children := make([]Node, 0, len(s.Selections))
	for _, sel := range s.Selections {
		if sel != nil {
			children = append(children, sel)
		}
	}
	return children
}
func (s *SelectionSet) CommentSlot() **CommentGroup { return &s.Comments }

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
func (f *Field) Children() []Node {
	children := make([]Node, 0, len(f.Arguments)+len(f.Directives)+1)
	for _, arg := range f.Arguments {
		if arg != nil {
			children = append(children, arg)
		}
	}
	for _, dir := range f.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	if f.SelectionSet != nil {
		children = append(children, f.SelectionSet)
	}
	return children
}
func (*Field) isSelection()                  {}
func (f *Field) CommentSlot() **CommentGroup { return &f.Comments }

// FragmentSpread is "...FragmentName Directives?". The name must not be "on";
// the parser disambiguates against InlineFragment.
type FragmentSpread struct {
	Name       string
	Directives DirectiveList
	Loc        *Loc
	Comments   *CommentGroup
}

func (s *FragmentSpread) GetLoc() *Loc { return s.Loc }
func (s *FragmentSpread) Children() []Node {
	children := make([]Node, 0, len(s.Directives))
	for _, dir := range s.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	return children
}
func (*FragmentSpread) isSelection()                  {}
func (s *FragmentSpread) CommentSlot() **CommentGroup { return &s.Comments }

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
func (f *InlineFragment) Children() []Node {
	children := make([]Node, 0, len(f.Directives)+2)
	if f.TypeCondition != nil {
		children = append(children, f.TypeCondition)
	}
	for _, dir := range f.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	if f.SelectionSet != nil {
		children = append(children, f.SelectionSet)
	}
	return children
}
func (*InlineFragment) isSelection()                  {}
func (f *InlineFragment) CommentSlot() **CommentGroup { return &f.Comments }

// Argument is one "name: value" entry in a function-style argument list.
type Argument struct {
	Name     string
	Value    Value
	Loc      *Loc
	Comments *CommentGroup
}

func (a *Argument) GetLoc() *Loc { return a.Loc }
func (a *Argument) Children() []Node {
	if a.Value == nil {
		return nil
	}
	return []Node{a.Value}
}
func (a *Argument) CommentSlot() **CommentGroup { return &a.Comments }

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
func (d *Directive) Children() []Node {
	children := make([]Node, 0, len(d.Arguments))
	for _, arg := range d.Arguments {
		if arg != nil {
			children = append(children, arg)
		}
	}
	return children
}
func (d *Directive) CommentSlot() **CommentGroup { return &d.Comments }
