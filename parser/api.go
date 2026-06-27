package parser

import "github.com/brian-bell/graphql-parser/ast"

// Parse parses a GraphQL source document and returns its [ast.Document].
// The source is wrapped in an ast.Source named "GraphQL" for error messages;
// use [ParseSource] to supply your own filename and location offset.
func Parse(body string, opts ...Option) (*ast.Document, error) {
	return ParseSource(&ast.Source{Body: body, Name: "GraphQL"}, opts...)
}

// ParseSchema parses a GraphQL schema document and returns its [ast.Document].
// The source is wrapped in an ast.Source named "GraphQL" for error messages;
// use [ParseSchemaSource] to supply your own filename and location offset.
func ParseSchema(body string, opts ...Option) (*ast.Document, error) {
	return ParseSchemaSource(&ast.Source{Body: body, Name: "GraphQL"}, opts...)
}

// ParseSource parses a GraphQL document from src.
//
// In the default fail-fast mode the first syntax error aborts and the
// returned *ast.Document is nil. With [WithRecovery], the parser collects
// every syntax error it encounters and returns a partial document plus a
// non-nil [ParseErrors]; the document may contain Bad{Definition,Field,Value,
// Type} placeholder nodes where the parser had to resynchronize.
func ParseSource(src *ast.Source, opts ...Option) (*ast.Document, error) {
	doc, _, err := parseSource(src, opts)
	return doc, err
}

func parseSource(src *ast.Source, opts []Option) (*ast.Document, *parser, error) {
	p := newParser(src, opts)
	doc, err := p.parseDocument()
	if err != nil {
		if p.cfg.recovery {
			_ = p.recordError(err)
		} else {
			return nil, p, err
		}
	}
	if err := p.expectEOF(); err != nil {
		if !p.cfg.recovery {
			return nil, p, err
		}
		_ = p.recordError(err)
	}
	if p.cfg.recovery && len(p.errors) > 0 {
		return doc, p, ParseErrors(p.errors)
	}
	return doc, p, nil
}

// ParseSchemaSource parses a GraphQL schema document from src.
func ParseSchemaSource(src *ast.Source, opts ...Option) (*ast.Document, error) {
	doc, p, err := parseSource(src, opts)
	if err != nil && !p.cfg.recovery {
		return nil, err
	}
	if doc == nil {
		return nil, err
	}

	schemaErrs := schemaOnlyErrors(doc)
	if len(schemaErrs) == 0 {
		return doc, err
	}
	if !p.cfg.recovery {
		return nil, schemaErrs[0]
	}
	p.errors = append(p.errors, schemaErrs...)
	return doc, ParseErrors(p.errors)
}

func schemaOnlyErrors(doc *ast.Document) []*ast.SyntaxError {
	var errs []*ast.SyntaxError
	for _, def := range doc.Definitions {
		switch def.(type) {
		case *ast.OperationDefinition, *ast.FragmentDefinition:
			loc := def.GetLoc()
			var src *ast.Source
			var offset int
			if loc != nil {
				src = loc.Source
				offset = loc.Start
			}
			errs = append(errs, &ast.SyntaxError{
				Source:  src,
				Offset:  offset,
				Message: "Executable definitions are not allowed in schema documents.",
			})
		}
	}
	return errs
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
