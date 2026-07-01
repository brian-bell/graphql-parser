package ast

// Visitor is the callback interface accepted by [Walk]. The Visit method is
// called for each node in source order. Returning nil stops the traversal
// from descending into the node's children; returning a non-nil Visitor
// (typically the same value) continues with that visitor for the subtree.
type Visitor interface {
	Visit(node Node) Visitor
}

type childWalker interface {
	walkChildren(Visitor)
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
	if walker, ok := node.(childWalker); ok {
		walker.walkChildren(v)
		return
	}
	for _, child := range node.Children() {
		Walk(v, child)
	}
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
