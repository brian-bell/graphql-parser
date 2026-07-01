package ast

// Comment is a single # ... line of comment text. Text is everything after
// the # up to (but not including) the line terminator.
type Comment struct {
	Text string
	Loc  *Loc
}

// GetLoc returns the comment's location.
func (c *Comment) GetLoc() *Loc { return c.Loc }

// Children returns nil because comments are traversal leaves.
func (*Comment) Children() []Node { return nil }

// CommentGroup attaches a node's leading and trailing comments. It is set on
// AST nodes only when the parser is run with the WithComments option;
// otherwise the field on each node is nil.
type CommentGroup struct {
	Leading  []*Comment
	Trailing []*Comment
}

// CommentedNode is implemented by every AST node that carries a CommentGroup.
// CommentSlot returns a pointer to the node's own Comments field, letting a
// binder attach trivia without a type switch. The interface is the single
// source of truth for "which nodes can hold comments"; the registry and parity
// test in comment_test.go pin its implementors to the structs that actually
// declare a Comments field. Implementing CommentedNode does not, by itself,
// cause comments to be attached — only the parser's bind call sites do that.
type CommentedNode interface {
	Node
	CommentSlot() **CommentGroup
}
