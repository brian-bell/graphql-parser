package ast

// Comment is a single # ... line of comment text. Text is everything after
// the # up to (but not including) the line terminator.
type Comment struct {
	Text string
	Loc  *Loc
}

// GetLoc returns the comment's location.
func (c *Comment) GetLoc() *Loc { return c.Loc }

// CommentGroup attaches a node's leading and trailing comments. It is set on
// AST nodes only when the parser is run with the WithComments option;
// otherwise the field on each node is nil.
type CommentGroup struct {
	Leading  []*Comment
	Trailing []*Comment
}
