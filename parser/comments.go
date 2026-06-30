package parser

import (
	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/lexer"
)

// drainComments consumes any pending COMMENT tokens at the current lexer
// position into p.pendingLeading. Called from peek/advance whenever
// preserveComments is on; a no-op otherwise (the lexer doesn't emit
// COMMENT tokens unless built with lexer.WithComments).
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

// takeLeading returns and clears the pending leading-comment buffer. Call it at
// the entry of a production whose node owns those comments, before parsing any
// children (which would otherwise drain their own comments into the same
// buffer and steal attribution). It runs unconditionally — do NOT add a
// preserveComments guard here. The buffer must be cleared even when
// preservation is off so timing matches the historical capture-at-entry code;
// when off, drainComments never fills it, so the clear is a harmless no-op.
func (p *parser) takeLeading() []*ast.Comment {
	leading := p.pendingLeading
	p.pendingLeading = nil
	return leading
}

// bindLeading attaches captured leading comments to n through the
// ast.CommentedNode interface. It is a no-op when preservation is off, the
// slice is empty, or n is nil/does not carry comments. The assertion succeeds
// for any node that has a Comments field, so binding scope is controlled by
// where this is called, not by the helper rejecting node types: the parser
// only calls it at the four sites that historically bound leading comments.
func (p *parser) bindLeading(n ast.Node, leading []*ast.Comment) {
	if !p.cfg.preserveComments || len(leading) == 0 || n == nil {
		return
	}
	cn, ok := n.(ast.CommentedNode)
	if !ok {
		return
	}
	slot := cn.CommentSlot()
	if *slot == nil {
		*slot = &ast.CommentGroup{}
	}
	(*slot).Leading = append((*slot).Leading, leading...)
}
