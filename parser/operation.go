package parser

import (
	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/lexer"
)

// parseShorthandOperation parses a bare "{ ... }" as an anonymous query.
func (p *parser) parseShorthandOperation() (*ast.OperationDefinition, error) {
	scope, err := p.enter()
	if err != nil {
		return nil, err
	}
	set, err := p.parseSelectionSet()
	if err != nil {
		return nil, err
	}
	return &ast.OperationDefinition{
		Operation:    ast.OperationQuery,
		SelectionSet: set,
		Loc:          scope.close(),
	}, nil
}

// parseOperationDefinition parses "query|mutation|subscription Name?
// VariableDefinitions? Directives? SelectionSet".
func (p *parser) parseOperationDefinition() (*ast.OperationDefinition, error) {
	keyword, err := p.advance()
	if err != nil {
		return nil, err
	}
	scope := p.scopeAt(keyword.Start)
	op := ast.OperationType(keyword.Value)

	def := &ast.OperationDefinition{Operation: op}

	// Optional Name.
	if next, err := p.peek(); err != nil {
		return nil, err
	} else if next.Kind == lexer.NAME {
		name, err := p.advance()
		if err != nil {
			return nil, err
		}
		def.Name = name.Value
	}

	// Optional VariableDefinitions.
	if next, err := p.peek(); err != nil {
		return nil, err
	} else if next.Kind == lexer.LPAREN {
		vars, err := p.parseVariableDefinitions()
		if err != nil {
			return nil, err
		}
		def.VariableDefinitions = vars
	}

	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, err
	}
	def.Directives = dirs

	set, err := p.parseSelectionSet()
	if err != nil {
		return nil, err
	}
	def.SelectionSet = set
	def.Loc = scope.close()
	return def, nil
}

// parseFragmentDefinition parses "fragment Name on NamedType Directives?
// SelectionSet".
func (p *parser) parseFragmentDefinition() (*ast.FragmentDefinition, error) {
	keyword, err := p.expectKeyword("fragment")
	if err != nil {
		return nil, err
	}
	scope := p.scopeAt(keyword.Start)
	name, err := p.parseFragmentName()
	if err != nil {
		return nil, err
	}
	if _, err := p.expectKeyword("on"); err != nil {
		return nil, err
	}
	cond, err := p.parseNamedType()
	if err != nil {
		return nil, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, err
	}
	set, err := p.parseSelectionSet()
	if err != nil {
		return nil, err
	}
	return &ast.FragmentDefinition{
		Name:          name,
		TypeCondition: cond,
		Directives:    dirs,
		SelectionSet:  set,
		Loc:           scope.close(),
	}, nil
}

// parseFragmentName accepts a Name that is not the reserved keyword "on".
func (p *parser) parseFragmentName() (string, error) {
	tok, err := p.peek()
	if err != nil {
		return "", err
	}
	if tok.Kind != lexer.NAME {
		return "", p.errAtTok(tok, "Expected fragment name, found "+describeToken(tok)+".")
	}
	if tok.Value == "on" {
		return "", p.errAtTok(tok, "Unexpected Name \"on\".")
	}
	if _, err := p.advance(); err != nil {
		return "", err
	}
	return tok.Value, nil
}

func (p *parser) parseNamedType() (*ast.NamedType, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	scope := p.scopeAt(tok.Start)
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, err
	}
	return &ast.NamedType{Name: name.Value, Loc: scope.close()}, nil
}

func (p *parser) parseVariableDefinitions() (ast.VariableDefinitionList, error) {
	if _, err := p.expect(lexer.LPAREN); err != nil {
		return nil, err
	}
	var out ast.VariableDefinitionList
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind == lexer.RPAREN {
			break
		}
		v, err := p.parseVariableDefinition()
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if _, err := p.expect(lexer.RPAREN); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		// Spec: VariableDefinitions = ( VariableDefinition+ ) — at least one.
		// Match graphql-js's error pointing at the closing paren.
		return nil, p.errAtTok(lexer.Token{Start: p.lastEnd}, "Expected variable definition.")
	}
	return out, nil
}

func (p *parser) parseVariableDefinition() (*ast.VariableDefinition, error) {
	dollar, err := p.peek()
	if err != nil {
		return nil, err
	}
	scope := p.scopeAt(dollar.Start)
	v, err := p.parseVariable()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.COLON); err != nil {
		return nil, err
	}
	t, err := p.parseTypeReference()
	if err != nil {
		return nil, err
	}
	def := &ast.VariableDefinition{Variable: v, Type: t}
	if _, ok, err := p.optional(lexer.EQUALS); err != nil {
		return nil, err
	} else if ok {
		dv, err := p.parseValueLiteral(true)
		if err != nil {
			return nil, err
		}
		def.DefaultValue = dv
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, err
	}
	def.Directives = dirs
	def.Loc = scope.close()
	return def, nil
}

// parseDirectives parses zero or more directives. isConst constrains directive
// arguments to const values.
func (p *parser) parseDirectives(isConst bool) (ast.DirectiveList, error) {
	var out ast.DirectiveList
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind != lexer.AT {
			break
		}
		d, err := p.parseDirective(isConst)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func (p *parser) parseDirective(isConst bool) (*ast.Directive, error) {
	at, err := p.expect(lexer.AT)
	if err != nil {
		return nil, err
	}
	scope := p.scopeAt(at.Start)
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, err
	}
	args, err := p.parseArguments(isConst)
	if err != nil {
		return nil, err
	}
	return &ast.Directive{
		Name:      name.Value,
		Arguments: args,
		Loc:       scope.close(),
	}, nil
}

func (p *parser) parseArguments(isConst bool) (ast.ArgumentList, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if tok.Kind != lexer.LPAREN {
		return nil, nil
	}
	if _, err := p.advance(); err != nil {
		return nil, err
	}
	var out ast.ArgumentList
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind == lexer.RPAREN {
			break
		}
		a, err := p.parseArgument(isConst)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	if _, err := p.expect(lexer.RPAREN); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, p.errAtTok(lexer.Token{Start: p.lastEnd}, "Expected argument.")
	}
	return out, nil
}

func (p *parser) parseArgument(isConst bool) (*ast.Argument, error) {
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, err
	}
	scope := p.scopeAt(name.Start)
	if _, err := p.expect(lexer.COLON); err != nil {
		return nil, err
	}
	v, err := p.parseValueLiteral(isConst)
	if err != nil {
		return nil, err
	}
	return &ast.Argument{
		Name:  name.Value,
		Value: v,
		Loc:   scope.close(),
	}, nil
}
