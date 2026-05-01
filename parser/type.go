package parser

import (
	"github.com/bellbm/graphql-parser/ast"
	"github.com/bellbm/graphql-parser/lexer"
)

// parseTypeReference parses a Type:
//
//	Type: NamedType | ListType | NonNullType
//	NamedType: Name
//	ListType: [ Type ]
//	NonNullType: NamedType ! | ListType !
//
// NonNull is a postfix operator that may apply to either of the inner forms,
// but never to itself (Int!! is rejected).
func (p *parser) parseTypeReference() (ast.Type, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	var inner ast.Type
	switch tok.Kind {
	case lexer.LBRACKET:
		open, err := p.advance()
		if err != nil {
			return nil, err
		}
		of, err := p.parseTypeReference()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.RBRACKET); err != nil {
			return nil, err
		}
		inner = &ast.ListType{OfType: of, Loc: p.loc(open.Start)}
	case lexer.NAME:
		name, err := p.advance()
		if err != nil {
			return nil, err
		}
		inner = &ast.NamedType{Name: name.Value, Loc: p.loc(name.Start)}
	default:
		return nil, p.errAtTok(tok, "Expected type, found "+describeToken(tok)+".")
	}
	// Optional trailing "!".
	_, ok, err := p.optional(lexer.BANG)
	if err != nil {
		return nil, err
	}
	if ok {
		return &ast.NonNullType{OfType: inner, Loc: p.loc(inner.GetLoc().Start)}, nil
	}
	return inner, nil
}
