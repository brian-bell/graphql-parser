package parser

import (
	"github.com/bellbm/graphql-parser/ast"
	"github.com/bellbm/graphql-parser/lexer"
)

// parseSelectionSet parses "{ Selection+ }". An empty selection set is a
// syntax error per spec.
func (p *parser) parseSelectionSet() (*ast.SelectionSet, error) {
	open, err := p.expect(lexer.LBRACE)
	if err != nil {
		return nil, err
	}
	var sels []ast.Selection
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind == lexer.RBRACE {
			break
		}
		if tok.Kind == lexer.EOF {
			return nil, p.errAtTok(tok, "Expected '}', found <EOF>.")
		}
		s, err := p.parseSelection()
		if err != nil {
			return nil, err
		}
		sels = append(sels, s)
	}
	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}
	if len(sels) == 0 {
		return nil, p.errAtTok(lexer.Token{Start: p.lastEnd}, "Expected at least one selection.")
	}
	return &ast.SelectionSet{Selections: sels, Loc: p.loc(open.Start)}, nil
}

func (p *parser) parseSelection() (ast.Selection, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if tok.Kind == lexer.SPREAD {
		return p.parseFragment()
	}
	return p.parseField()
}

// parseField parses "[Alias :] Name (Arguments)? @Directive...? SelectionSet?".
func (p *parser) parseField() (*ast.Field, error) {
	first, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, err
	}
	field := &ast.Field{}

	// If the next token is a colon, the first NAME was an alias.
	if next, err := p.peek(); err != nil {
		return nil, err
	} else if next.Kind == lexer.COLON {
		if _, err := p.advance(); err != nil {
			return nil, err
		}
		nameTok, err := p.expect(lexer.NAME)
		if err != nil {
			return nil, err
		}
		field.Alias = first.Value
		field.Name = nameTok.Value
	} else {
		field.Name = first.Value
	}

	args, err := p.parseArguments(false)
	if err != nil {
		return nil, err
	}
	field.Arguments = args

	dirs, err := p.parseDirectives(false)
	if err != nil {
		return nil, err
	}
	field.Directives = dirs

	if next, err := p.peek(); err != nil {
		return nil, err
	} else if next.Kind == lexer.LBRACE {
		set, err := p.parseSelectionSet()
		if err != nil {
			return nil, err
		}
		field.SelectionSet = set
	}

	field.Loc = p.loc(first.Start)
	return field, nil
}

// parseFragment dispatches between a FragmentSpread and an InlineFragment
// based on the token following "...".
//
//	"..." Name [not "on"] Directives?         → FragmentSpread
//	"..." "on" NamedType Directives? Set      → InlineFragment with TypeCondition
//	"..." Directives? SelectionSet            → InlineFragment without TypeCondition
func (p *parser) parseFragment() (ast.Selection, error) {
	spread, err := p.expect(lexer.SPREAD)
	if err != nil {
		return nil, err
	}
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	// "...on" is always an InlineFragment with TypeCondition.
	if tok.Kind == lexer.NAME && tok.Value == "on" {
		if _, err := p.advance(); err != nil {
			return nil, err
		}
		cond, err := p.parseNamedType()
		if err != nil {
			return nil, err
		}
		dirs, err := p.parseDirectives(false)
		if err != nil {
			return nil, err
		}
		set, err := p.parseSelectionSet()
		if err != nil {
			return nil, err
		}
		return &ast.InlineFragment{
			TypeCondition: cond,
			Directives:    dirs,
			SelectionSet:  set,
			Loc:           p.loc(spread.Start),
		}, nil
	}
	// "...Name" is a FragmentSpread (Name is not "on").
	if tok.Kind == lexer.NAME {
		name, err := p.advance()
		if err != nil {
			return nil, err
		}
		dirs, err := p.parseDirectives(false)
		if err != nil {
			return nil, err
		}
		return &ast.FragmentSpread{
			Name:       name.Value,
			Directives: dirs,
			Loc:        p.loc(spread.Start),
		}, nil
	}
	// "...@dir { ... }" or "...{ ... }" is an InlineFragment without TypeCondition.
	if tok.Kind == lexer.AT || tok.Kind == lexer.LBRACE {
		dirs, err := p.parseDirectives(false)
		if err != nil {
			return nil, err
		}
		set, err := p.parseSelectionSet()
		if err != nil {
			return nil, err
		}
		return &ast.InlineFragment{
			Directives:   dirs,
			SelectionSet: set,
			Loc:          p.loc(spread.Start),
		}, nil
	}
	return nil, p.errAtTok(tok, "Expected fragment spread or inline fragment, found "+describeToken(tok)+".")
}
