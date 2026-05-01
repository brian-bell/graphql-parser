package parser

import (
	"github.com/bellbm/graphql-parser/ast"
	"github.com/bellbm/graphql-parser/lexer"
)

// parseDocument consumes the full source and returns a Document containing
// one or more Definitions. Empty input is an error per spec.
func (p *parser) parseDocument() (*ast.Document, error) {
	first, err := p.peek()
	if err != nil {
		return nil, err
	}
	start := first.Start
	var defs ast.DefinitionList
	for {
		tok, err := p.peek()
		if err != nil {
			if p.cfg.recovery {
				_ = p.recordError(err)
				p.skipToDefinitionStart()
				continue
			}
			return nil, err
		}
		if tok.Kind == lexer.EOF {
			break
		}
		defStart := tok.Start
		def, err := p.parseDefinition()
		if err != nil {
			if !p.recordError(err) {
				return nil, err
			}
			se, _ := err.(*ast.SyntaxError)
			p.skipToDefinitionStart()
			defs = append(defs, &ast.BadDefinition{Err: se, Loc: p.loc(defStart)})
			continue
		}
		defs = append(defs, def)
	}
	if len(defs) == 0 {
		return nil, p.errAtTok(first, "Expected at least one definition.")
	}
	return &ast.Document{
		Definitions: defs,
		Loc:         &ast.Loc{Start: start, End: p.lastEnd, Source: p.source},
	}, nil
}

// parseDefinition dispatches on the first token of a definition. Phase 6
// handles executable definitions (operations and fragments) and the bare
// "{ ... }" shorthand operation; phase 7 extends this to type-system
// definitions and extensions.
func (p *parser) parseDefinition() (ast.Definition, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	switch tok.Kind {
	case lexer.LBRACE:
		return p.parseShorthandOperation()
	case lexer.NAME:
		switch tok.Value {
		case "query", "mutation", "subscription":
			return p.parseOperationDefinition()
		case "fragment":
			return p.parseFragmentDefinition()
		}
		// Type-system keywords are handled in phase 7. For now, fall through.
		def, ok, err := p.parseTypeSystemDefinitionOrExtension()
		if err != nil {
			return nil, err
		}
		if ok {
			return def, nil
		}
	case lexer.STRING:
		// A leading string is a description for a type-system definition; phase 7.
		def, ok, err := p.parseTypeSystemDefinitionOrExtension()
		if err != nil {
			return nil, err
		}
		if ok {
			return def, nil
		}
	}
	return nil, p.errAtTok(tok, "Unexpected "+describeToken(tok)+".")
}
