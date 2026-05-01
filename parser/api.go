package parser

import "github.com/brian-bell/graphql-parser/ast"

// Parse parses a GraphQL source document and returns its [ast.Document].
// The source is wrapped in an ast.Source named "GraphQL" for error messages;
// use [ParseSource] to supply your own filename and location offset.
func Parse(body string, opts ...Option) (*ast.Document, error) {
	return ParseSource(&ast.Source{Body: body, Name: "GraphQL"}, opts...)
}

// ParseSource parses a GraphQL document from src.
//
// In the default fail-fast mode the first syntax error aborts and the
// returned *ast.Document is nil. With [WithRecovery], the parser collects
// every syntax error it encounters and returns a partial document plus a
// non-nil [ParseErrors]; the document may contain Bad{Definition,Field,Value,
// Type} placeholder nodes where the parser had to resynchronize.
func ParseSource(src *ast.Source, opts ...Option) (*ast.Document, error) {
	p := newParser(src, opts)
	doc, err := p.parseDocument()
	if err != nil {
		if p.cfg.recovery {
			_ = p.recordError(err)
		} else {
			return nil, err
		}
	}
	if err := p.expectEOF(); err != nil {
		if !p.cfg.recovery {
			return nil, err
		}
		_ = p.recordError(err)
	}
	if p.cfg.recovery && len(p.errors) > 0 {
		return doc, ParseErrors(p.errors)
	}
	return doc, nil
}

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
