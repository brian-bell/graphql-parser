package parser

import (
	"fmt"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/lexer"
)

// parseValueLiteral parses a Value (or ConstValue when isConst is true).
//
//	Value: Variable | IntValue | FloatValue | StringValue | BooleanValue
//	     | NullValue | EnumValue | ListValue | ObjectValue
//
// In a const context, Variable is rejected; otherwise it is allowed.
func (p *parser) parseValueLiteral(isConst bool) (ast.Value, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	switch tok.Kind {
	case lexer.LBRACKET:
		return p.parseListValue(isConst)
	case lexer.LBRACE:
		return p.parseObjectValue(isConst)
	case lexer.INT:
		scope := p.scopeAt(tok.Start)
		if _, err := p.advance(); err != nil {
			return nil, err
		}
		return &ast.IntValue{Value: tok.Value, Loc: scope.close()}, nil
	case lexer.FLOAT:
		scope := p.scopeAt(tok.Start)
		if _, err := p.advance(); err != nil {
			return nil, err
		}
		return &ast.FloatValue{Value: tok.Value, Loc: scope.close()}, nil
	case lexer.STRING:
		scope := p.scopeAt(tok.Start)
		if _, err := p.advance(); err != nil {
			return nil, err
		}
		return &ast.StringValue{Value: tok.Value, Block: tok.Block, Loc: scope.close()}, nil
	case lexer.NAME:
		switch tok.Value {
		case "true", "false":
			scope := p.scopeAt(tok.Start)
			if _, err := p.advance(); err != nil {
				return nil, err
			}
			return &ast.BooleanValue{Value: tok.Value == "true", Loc: scope.close()}, nil
		case "null":
			scope := p.scopeAt(tok.Start)
			if _, err := p.advance(); err != nil {
				return nil, err
			}
			return &ast.NullValue{Loc: scope.close()}, nil
		default:
			scope := p.scopeAt(tok.Start)
			if _, err := p.advance(); err != nil {
				return nil, err
			}
			return &ast.EnumValue{Value: tok.Value, Loc: scope.close()}, nil
		}
	case lexer.DOLLAR:
		if isConst {
			return nil, p.errAtTok(tok, "Unexpected variable in const context.")
		}
		return p.parseVariable()
	}
	return nil, p.errAtTok(tok, fmt.Sprintf("Unexpected %s.", describeToken(tok)))
}

func (p *parser) parseVariable() (*ast.Variable, error) {
	dollar, err := p.expect(lexer.DOLLAR)
	if err != nil {
		return nil, err
	}
	scope := p.scopeAt(dollar.Start)
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, err
	}
	return &ast.Variable{Name: name.Value, Loc: scope.close()}, nil
}

func (p *parser) parseListValue(isConst bool) (*ast.ListValue, error) {
	scope, err := p.enter()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.LBRACKET); err != nil {
		return nil, err
	}
	var values []ast.Value
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind == lexer.RBRACKET {
			break
		}
		if tok.Kind == lexer.EOF {
			return nil, p.errAtTok(tok, "Expected ']', found <EOF>.")
		}
		v, err := p.parseValueLiteral(isConst)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	if _, err := p.expect(lexer.RBRACKET); err != nil {
		return nil, err
	}
	return &ast.ListValue{Values: values, Loc: scope.close()}, nil
}

func (p *parser) parseObjectValue(isConst bool) (*ast.ObjectValue, error) {
	scope, err := p.enter()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.LBRACE); err != nil {
		return nil, err
	}
	var fields []*ast.ObjectField
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
		f, err := p.parseObjectField(isConst)
		if err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}
	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}
	return &ast.ObjectValue{Fields: fields, Loc: scope.close()}, nil
}

func (p *parser) parseObjectField(isConst bool) (*ast.ObjectField, error) {
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
	return &ast.ObjectField{Name: name.Value, Value: v, Loc: scope.close()}, nil
}
