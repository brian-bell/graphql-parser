package ast

// Node is the super-interface satisfied by every AST node. The location is
// nullable: synthetic nodes constructed by tools (codegen, transforms) may
// have a nil Loc. Children returns the node's AST children in source order,
// excluding nil children and comments.
type Node interface {
	GetLoc() *Loc
	Children() []Node
}

// Definition is the union of all top-level definitions in a Document:
// executable definitions (operation, fragment) and type-system definitions
// (schema, type, directive, plus their extensions).
type Definition interface {
	Node
	isDefinition()
}

// Selection is the union of items inside a SelectionSet: Field,
// FragmentSpread, and InlineFragment, plus BadField when recovery is on.
type Selection interface {
	Node
	isSelection()
}

// Value is the union of literal value nodes that may appear in arguments,
// default values, list/object construction, and so on.
type Value interface {
	Node
	isValue()
}

// Type is the union of type-reference nodes that appear in operation variable
// declarations and type-system definitions: NamedType, ListType, NonNullType.
type Type interface {
	Node
	isType()
}
