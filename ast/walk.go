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
	if walkBuiltInChildren(v, node) {
		return
	}
	for _, child := range node.Children() {
		Walk(v, child)
	}
}

func walkBuiltInChildren(v Visitor, node Node) bool {
	switch n := node.(type) {
	case *Document:
		n.walkChildren(v)
	case *OperationDefinition:
		n.walkChildren(v)
	case *FragmentDefinition:
		n.walkChildren(v)
	case *VariableDefinition:
		n.walkChildren(v)
	case *SelectionSet:
		n.walkChildren(v)
	case *Field:
		n.walkChildren(v)
	case *FragmentSpread:
		n.walkChildren(v)
	case *InlineFragment:
		n.walkChildren(v)
	case *Argument:
		n.walkChildren(v)
	case *Directive:
		n.walkChildren(v)
	case *ListValue:
		n.walkChildren(v)
	case *ObjectValue:
		n.walkChildren(v)
	case *ObjectField:
		n.walkChildren(v)
	case *ListType:
		n.walkChildren(v)
	case *NonNullType:
		n.walkChildren(v)
	case *SchemaDefinition:
		n.walkChildren(v)
	case *SchemaExtension:
		n.walkChildren(v)
	case *OperationTypeDefinition:
		n.walkChildren(v)
	case *ScalarTypeDefinition:
		n.walkChildren(v)
	case *ScalarTypeExtension:
		n.walkChildren(v)
	case *ObjectTypeDefinition:
		n.walkChildren(v)
	case *ObjectTypeExtension:
		n.walkChildren(v)
	case *InterfaceTypeDefinition:
		n.walkChildren(v)
	case *InterfaceTypeExtension:
		n.walkChildren(v)
	case *UnionTypeDefinition:
		n.walkChildren(v)
	case *UnionTypeExtension:
		n.walkChildren(v)
	case *EnumTypeDefinition:
		n.walkChildren(v)
	case *EnumTypeExtension:
		n.walkChildren(v)
	case *InputObjectTypeDefinition:
		n.walkChildren(v)
	case *InputObjectTypeExtension:
		n.walkChildren(v)
	case *FieldDefinition:
		n.walkChildren(v)
	case *InputValueDefinition:
		n.walkChildren(v)
	case *EnumValueDefinition:
		n.walkChildren(v)
	case *DirectiveDefinition:
		n.walkChildren(v)
	default:
		return false
	}
	return true
}

func walkNode(v Visitor, child Node) {
	if child != nil {
		Walk(v, child)
	}
}

func walkDefinitions(v Visitor, children DefinitionList) {
	for _, child := range children {
		walkNode(v, child)
	}
}

func walkVariableDefinitions(v Visitor, children VariableDefinitionList) {
	for _, child := range children {
		if child != nil {
			Walk(v, child)
		}
	}
}

func walkSelections(v Visitor, children []Selection) {
	for _, child := range children {
		walkNode(v, child)
	}
}

func walkArguments(v Visitor, children ArgumentList) {
	for _, child := range children {
		if child != nil {
			Walk(v, child)
		}
	}
}

func walkDirectives(v Visitor, children DirectiveList) {
	for _, child := range children {
		if child != nil {
			Walk(v, child)
		}
	}
}

func walkValues(v Visitor, children []Value) {
	for _, child := range children {
		walkNode(v, child)
	}
}

func walkObjectFields(v Visitor, children []*ObjectField) {
	for _, child := range children {
		if child != nil {
			Walk(v, child)
		}
	}
}

func walkNamedTypes(v Visitor, children []*NamedType) {
	for _, child := range children {
		if child != nil {
			Walk(v, child)
		}
	}
}

func walkOperationTypes(v Visitor, children []*OperationTypeDefinition) {
	for _, child := range children {
		if child != nil {
			Walk(v, child)
		}
	}
}

func walkFieldDefinitions(v Visitor, children FieldDefinitionList) {
	for _, child := range children {
		if child != nil {
			Walk(v, child)
		}
	}
}

func walkInputValues(v Visitor, children InputValueList) {
	for _, child := range children {
		if child != nil {
			Walk(v, child)
		}
	}
}

func walkEnumValues(v Visitor, children EnumValueList) {
	for _, child := range children {
		if child != nil {
			Walk(v, child)
		}
	}
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
