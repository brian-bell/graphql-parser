package parser

import "github.com/brian-bell/graphql-parser/ast"

// Parser is a reusable parse entry point configured by [Option] values.
//
// A Parser stores only immutable parse configuration. Each parse method creates
// a fresh internal lexer/parser, so one Parser can be reused across calls
// without leaking source, comment, recovery, or error state.
type Parser struct {
	cfg config
}

// New returns a reusable Parser configured by opts.
func New(opts ...Option) *Parser {
	return &Parser{cfg: *newConfig(opts)}
}

// Parse parses a GraphQL source document and returns its [ast.Document].
// The source is wrapped in an ast.Source named "GraphQL" for error messages;
// use [ParseSource] to supply your own filename and location offset.
func Parse(body string, opts ...Option) (*ast.Document, error) {
	return New(opts...).Parse(&ast.Source{Body: body, Name: "GraphQL"})
}

// ParseSchema parses a GraphQL schema document and returns its [ast.Document].
// The source is wrapped in an ast.Source named "GraphQL" for error messages;
// use [ParseSchemaSource] to supply your own filename and location offset.
func ParseSchema(body string, opts ...Option) (*ast.Document, error) {
	return New(opts...).ParseSchema(&ast.Source{Body: body, Name: "GraphQL"})
}

// ParseSource parses a GraphQL document from src.
//
// In the default fail-fast mode the first syntax error aborts and the
// returned *ast.Document is nil. With [WithRecovery], the parser collects
// every syntax error it encounters and returns a partial document plus a
// non-nil [ParseErrors]; the document may contain Bad{Definition,Field,Value,
// Type} placeholder nodes where the parser had to resynchronize.
func ParseSource(src *ast.Source, opts ...Option) (*ast.Document, error) {
	return New(opts...).Parse(src)
}

// Parse parses a GraphQL document from src.
//
// In the default fail-fast mode the first syntax error aborts and the
// returned *ast.Document is nil. With [WithRecovery], the parser collects
// every syntax error it encounters and returns a partial document plus a
// non-nil [ParseErrors].
func (p *Parser) Parse(src *ast.Source) (*ast.Document, error) {
	doc, _, err := p.parseSource(src)
	return doc, err
}

func (p *Parser) parseSource(src *ast.Source) (*ast.Document, *parser, error) {
	internal := newParserWithConfig(src, p.cfg)
	doc, err := internal.parseDocument()
	if err != nil {
		if internal.cfg.recovery {
			_ = internal.recordError(err)
			if doc == nil {
				internal.skipToEOF()
			}
		} else {
			return nil, internal, err
		}
	}
	if err := internal.expectEOF(); err != nil {
		if !internal.cfg.recovery {
			return nil, internal, err
		}
		_ = internal.recordError(err)
	}
	if internal.cfg.recovery && len(internal.errors) > 0 {
		return doc, internal, ParseErrors(internal.errors)
	}
	return doc, internal, nil
}

// ParseSchemaSource parses a GraphQL schema document from src.
func ParseSchemaSource(src *ast.Source, opts ...Option) (*ast.Document, error) {
	return New(opts...).ParseSchema(src)
}

// ParseSchema parses a GraphQL schema document from src.
func (p *Parser) ParseSchema(src *ast.Source) (*ast.Document, error) {
	doc, internal, err := p.parseSource(src)
	if err != nil && !internal.cfg.recovery {
		return nil, err
	}
	if doc == nil {
		return nil, err
	}

	schemaErrs := schemaOnlyErrors(doc)
	if len(schemaErrs) == 0 {
		return doc, err
	}
	if !internal.cfg.recovery {
		return nil, schemaErrs[0]
	}
	internal.errors = append(internal.errors, schemaErrs...)
	return doc, ParseErrors(internal.errors)
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
	return New(opts...).ParseValue(&ast.Source{Body: body, Name: "GraphQL"})
}

// ParseConstValue parses a single GraphQL const value literal. Variables
// ($name) are rejected at any nesting depth.
func ParseConstValue(body string, opts ...Option) (ast.Value, error) {
	return New(opts...).ParseConstValue(&ast.Source{Body: body, Name: "GraphQL"})
}

// ParseType parses a single GraphQL type reference (NamedType, ListType, or
// NonNullType).
func ParseType(body string, opts ...Option) (ast.Type, error) {
	return New(opts...).ParseType(&ast.Source{Body: body, Name: "GraphQL"})
}

// ParseValue parses a single GraphQL value literal from src. Variables
// ($name) are allowed; for the const-context grammar use [Parser.ParseConstValue].
func (p *Parser) ParseValue(src *ast.Source) (ast.Value, error) {
	return p.parseValue(src, false)
}

// ParseConstValue parses a single GraphQL const value literal from src.
// Variables ($name) are rejected at any nesting depth.
func (p *Parser) ParseConstValue(src *ast.Source) (ast.Value, error) {
	return p.parseValue(src, true)
}

// ParseType parses a single GraphQL type reference (NamedType, ListType, or
// NonNullType) from src.
func (p *Parser) ParseType(src *ast.Source) (ast.Type, error) {
	internal := newParserWithConfig(src, p.cfg)
	start := partialEntryStart(internal)
	t, err := internal.parseTypeReference()
	if err != nil {
		if internal.recordError(err) {
			se, _ := err.(*ast.SyntaxError)
			internal.skipToEOF()
			return &ast.BadType{Err: se, Loc: internal.partialEntryLoc(start)}, ParseErrors(internal.errors)
		}
		return nil, err
	}
	if err := internal.expectEOF(); err != nil {
		if internal.recordError(err) {
			se, _ := err.(*ast.SyntaxError)
			internal.skipToEOF()
			return &ast.BadType{Err: se, Loc: internal.partialEntryLoc(start)}, ParseErrors(internal.errors)
		}
		return nil, err
	}
	return t, nil
}

func (p *Parser) parseValue(src *ast.Source, isConst bool) (ast.Value, error) {
	internal := newParserWithConfig(src, p.cfg)
	start := partialEntryStart(internal)
	v, err := internal.parseValueLiteral(isConst)
	if err != nil {
		if internal.recordError(err) {
			se, _ := err.(*ast.SyntaxError)
			internal.skipToEOF()
			return &ast.BadValue{Err: se, Loc: internal.partialEntryLoc(start)}, ParseErrors(internal.errors)
		}
		return nil, err
	}
	if err := internal.expectEOF(); err != nil {
		if internal.recordError(err) {
			se, _ := err.(*ast.SyntaxError)
			internal.skipToEOF()
			return &ast.BadValue{Err: se, Loc: internal.partialEntryLoc(start)}, ParseErrors(internal.errors)
		}
		return nil, err
	}
	return v, nil
}

func partialEntryStart(p *parser) int {
	tok, err := p.peek()
	if err != nil {
		return 0
	}
	return tok.Start
}

func (p *parser) partialEntryLoc(start int) *ast.Loc {
	end := p.lastEnd
	if end < start {
		end = start
	}
	return &ast.Loc{Start: start, End: end, Source: p.source}
}
