package parser

import (
	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/lexer"
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
	var scope prodScope
	switch tok.Kind {
	case lexer.LBRACKET:
		scope = p.scopeAt(tok.Start)
		if _, err := p.advance(); err != nil {
			return nil, err
		}
		of, err := p.parseTypeReference()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.RBRACKET); err != nil {
			return nil, err
		}
		inner = &ast.ListType{OfType: of, Loc: scope.close()}
	case lexer.NAME:
		scope = p.scopeAt(tok.Start)
		name, err := p.advance()
		if err != nil {
			return nil, err
		}
		inner = &ast.NamedType{Name: name.Value, Loc: scope.close()}
	default:
		return nil, p.errAtTok(tok, "Expected type, found "+describeToken(tok)+".")
	}
	// Optional trailing "!".
	_, ok, err := p.optional(lexer.BANG)
	if err != nil {
		return nil, err
	}
	if ok {
		return &ast.NonNullType{OfType: inner, Loc: scope.close()}, nil
	}
	return inner, nil
}
