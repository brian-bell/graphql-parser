package ast

// Visitor is the callback interface accepted by [Walk]. The Visit method is
// called for each node in source order. Returning nil stops the traversal
// from descending into the node's children; returning a non-nil Visitor
// (typically the same value) continues with that visitor for the subtree.
type Visitor interface {
	Visit(node Node) Visitor
}

// Walk performs a depth-first, source-order traversal of node and its
// children. It does not visit Comments or descend into BadValue/BadType/
// BadField/BadDefinition placeholder nodes — they are leaves for traversal
// purposes.
func Walk(v Visitor, node Node) {
	if node == nil {
		return
	}
	v = v.Visit(node)
	if v == nil {
		return
	}
	switch n := node.(type) {
	// Document and executable definitions
	case *Document:
		for _, def := range n.Definitions {
			Walk(v, def)
		}
	case *OperationDefinition:
		for _, vd := range n.VariableDefinitions {
			Walk(v, vd)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
		if n.SelectionSet != nil {
			Walk(v, n.SelectionSet)
		}
	case *FragmentDefinition:
		if n.TypeCondition != nil {
			Walk(v, n.TypeCondition)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
		if n.SelectionSet != nil {
			Walk(v, n.SelectionSet)
		}
	case *VariableDefinition:
		if n.Variable != nil {
			Walk(v, n.Variable)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}
		if n.DefaultValue != nil {
			Walk(v, n.DefaultValue)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
	case *SelectionSet:
		for _, s := range n.Selections {
			Walk(v, s)
		}
	case *Field:
		for _, a := range n.Arguments {
			Walk(v, a)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
		if n.SelectionSet != nil {
			Walk(v, n.SelectionSet)
		}
	case *FragmentSpread:
		for _, d := range n.Directives {
			Walk(v, d)
		}
	case *InlineFragment:
		if n.TypeCondition != nil {
			Walk(v, n.TypeCondition)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
		if n.SelectionSet != nil {
			Walk(v, n.SelectionSet)
		}
	case *Argument:
		if n.Value != nil {
			Walk(v, n.Value)
		}
	case *Directive:
		for _, a := range n.Arguments {
			Walk(v, a)
		}

	// Values
	case *ListValue:
		for _, val := range n.Values {
			Walk(v, val)
		}
	case *ObjectValue:
		for _, f := range n.Fields {
			Walk(v, f)
		}
	case *ObjectField:
		if n.Value != nil {
			Walk(v, n.Value)
		}

	// Types
	case *ListType:
		if n.OfType != nil {
			Walk(v, n.OfType)
		}
	case *NonNullType:
		if n.OfType != nil {
			Walk(v, n.OfType)
		}

	// Type-system definitions
	case *SchemaDefinition:
		if n.Description != nil {
			Walk(v, n.Description)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
		for _, ot := range n.OperationTypes {
			Walk(v, ot)
		}
	case *SchemaExtension:
		for _, d := range n.Directives {
			Walk(v, d)
		}
		for _, ot := range n.OperationTypes {
			Walk(v, ot)
		}
	case *OperationTypeDefinition:
		if n.Type != nil {
			Walk(v, n.Type)
		}

	case *ScalarTypeDefinition:
		if n.Description != nil {
			Walk(v, n.Description)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
	case *ScalarTypeExtension:
		for _, d := range n.Directives {
			Walk(v, d)
		}

	case *ObjectTypeDefinition:
		if n.Description != nil {
			Walk(v, n.Description)
		}
		for _, i := range n.Interfaces {
			Walk(v, i)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
		for _, f := range n.Fields {
			Walk(v, f)
		}
	case *ObjectTypeExtension:
		for _, i := range n.Interfaces {
			Walk(v, i)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
		for _, f := range n.Fields {
			Walk(v, f)
		}

	case *InterfaceTypeDefinition:
		if n.Description != nil {
			Walk(v, n.Description)
		}
		for _, i := range n.Interfaces {
			Walk(v, i)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
		for _, f := range n.Fields {
			Walk(v, f)
		}
	case *InterfaceTypeExtension:
		for _, i := range n.Interfaces {
			Walk(v, i)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
		for _, f := range n.Fields {
			Walk(v, f)
		}

	case *UnionTypeDefinition:
		if n.Description != nil {
			Walk(v, n.Description)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
		for _, m := range n.Members {
			Walk(v, m)
		}
	case *UnionTypeExtension:
		for _, d := range n.Directives {
			Walk(v, d)
		}
		for _, m := range n.Members {
			Walk(v, m)
		}

	case *EnumTypeDefinition:
		if n.Description != nil {
			Walk(v, n.Description)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
		for _, ev := range n.Values {
			Walk(v, ev)
		}
	case *EnumTypeExtension:
		for _, d := range n.Directives {
			Walk(v, d)
		}
		for _, ev := range n.Values {
			Walk(v, ev)
		}

	case *InputObjectTypeDefinition:
		if n.Description != nil {
			Walk(v, n.Description)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
		for _, f := range n.Fields {
			Walk(v, f)
		}
	case *InputObjectTypeExtension:
		for _, d := range n.Directives {
			Walk(v, d)
		}
		for _, f := range n.Fields {
			Walk(v, f)
		}

	case *FieldDefinition:
		if n.Description != nil {
			Walk(v, n.Description)
		}
		for _, a := range n.Arguments {
			Walk(v, a)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
	case *InputValueDefinition:
		if n.Description != nil {
			Walk(v, n.Description)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}
		if n.DefaultValue != nil {
			Walk(v, n.DefaultValue)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
	case *EnumValueDefinition:
		if n.Description != nil {
			Walk(v, n.Description)
		}
		for _, d := range n.Directives {
			Walk(v, d)
		}
	case *DirectiveDefinition:
		if n.Description != nil {
			Walk(v, n.Description)
		}
		for _, a := range n.Arguments {
			Walk(v, a)
		}
	}
	// Falling through with no case match is a no-op: leaf nodes
	// (IntValue, FloatValue, StringValue, BooleanValue, NullValue,
	// EnumValue, Variable, NamedType, Comment, Bad*) have no children,
	// so simply visiting them and returning is correct. Any future node
	// kind added without a case here defaults to the same behavior.
}

// inspector adapts a func(Node) bool to the Visitor interface for Inspect.
type inspector func(Node) bool

func (f inspector) Visit(n Node) Visitor {
	if f(n) {
		return f
	}
	return nil
}

// Inspect walks node depth-first, calling fn for each visited node. If fn
// returns false, descent into that node's children is skipped (sibling
// traversal continues).
func Inspect(node Node, fn func(Node) bool) {
	Walk(inspector(fn), node)
}
