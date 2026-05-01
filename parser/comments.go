package parser

import (
	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/lexer"
)

// drainComments consumes any pending COMMENT tokens at the current lexer
// position into p.pendingLeading. Called from peek/advance whenever
// preserveComments is on; a no-op otherwise (the lexer doesn't emit
// COMMENT tokens unless PreserveComments is set).
func (p *parser) drainComments() {
	if !p.cfg.preserveComments {
		return
	}
	for {
		tok, err := p.lex.Peek()
		if err != nil || tok.Kind != lexer.COMMENT {
			return
		}
		consumed, err := p.lex.Next()
		if err != nil {
			return
		}
		p.pendingLeading = append(p.pendingLeading, &ast.Comment{
			Text: consumed.Value,
			Loc: &ast.Loc{
				Start:  consumed.Start,
				End:    consumed.End,
				Source: p.source,
			},
		})
	}
}

// attachLeadingComments writes the given comment slice onto a node's
// Comments.Leading field. No-op when leading is empty or the node has no
// Comments field (e.g. a Type reference).
func attachLeadingComments(n ast.Node, leading []*ast.Comment) {
	if len(leading) == 0 || n == nil {
		return
	}
	cg := commentGroupOf(n)
	if cg == nil {
		return
	}
	if *cg == nil {
		*cg = &ast.CommentGroup{}
	}
	(*cg).Leading = append((*cg).Leading, leading...)
}

// commentGroupOf returns a pointer to the Comments field of n, or nil if n
// has no such field.
func commentGroupOf(n ast.Node) **ast.CommentGroup {
	switch v := n.(type) {
	case *ast.Document:
		return &v.Comments
	case *ast.OperationDefinition:
		return &v.Comments
	case *ast.FragmentDefinition:
		return &v.Comments
	case *ast.SchemaDefinition:
		return &v.Comments
	case *ast.SchemaExtension:
		return &v.Comments
	case *ast.ScalarTypeDefinition:
		return &v.Comments
	case *ast.ScalarTypeExtension:
		return &v.Comments
	case *ast.ObjectTypeDefinition:
		return &v.Comments
	case *ast.ObjectTypeExtension:
		return &v.Comments
	case *ast.InterfaceTypeDefinition:
		return &v.Comments
	case *ast.InterfaceTypeExtension:
		return &v.Comments
	case *ast.UnionTypeDefinition:
		return &v.Comments
	case *ast.UnionTypeExtension:
		return &v.Comments
	case *ast.EnumTypeDefinition:
		return &v.Comments
	case *ast.EnumTypeExtension:
		return &v.Comments
	case *ast.InputObjectTypeDefinition:
		return &v.Comments
	case *ast.InputObjectTypeExtension:
		return &v.Comments
	case *ast.DirectiveDefinition:
		return &v.Comments
	case *ast.FieldDefinition:
		return &v.Comments
	case *ast.InputValueDefinition:
		return &v.Comments
	case *ast.EnumValueDefinition:
		return &v.Comments
	}
	return nil
}
