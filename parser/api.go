package parser

import "github.com/bellbm/graphql-parser/ast"

// ParseValue parses a single GraphQL value literal. Variables ($name) are
// allowed; for the const-context grammar use [ParseConstValue].
func ParseValue(body string, opts ...Option) (ast.Value, error) {
	return parseValueEntry(body, false, opts)
}

// ParseConstValue parses a single GraphQL const value literal. Variables
// ($name) are rejected at any nesting depth.
func ParseConstValue(body string, opts ...Option) (ast.Value, error) {
	return parseValueEntry(body, true, opts)
}

// ParseType parses a single GraphQL type reference (NamedType, ListType, or
// NonNullType).
func ParseType(body string, opts ...Option) (ast.Type, error) {
	src := &ast.Source{Body: body, Name: "GraphQL"}
	p := newParser(src, opts)
	t, err := p.parseTypeReference()
	if err != nil {
		return nil, err
	}
	if err := p.expectEOF(); err != nil {
		return nil, err
	}
	return t, nil
}

func parseValueEntry(body string, isConst bool, opts []Option) (ast.Value, error) {
	src := &ast.Source{Body: body, Name: "GraphQL"}
	p := newParser(src, opts)
	v, err := p.parseValueLiteral(isConst)
	if err != nil {
		return nil, err
	}
	if err := p.expectEOF(); err != nil {
		return nil, err
	}
	return v, nil
}
