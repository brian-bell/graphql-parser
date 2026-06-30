// Package parser converts GraphQL source text into an [ast.Document] (or, via
// partial entry points, a single [ast.Value] or [ast.Type]). Errors are
// returned as [ParseError]; in fail-fast mode the first syntax error aborts
// the parse. Recovery mode, comments, and experimental-feature flags are
// configured via [Option] values.
package parser

import (
	"fmt"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/lexer"
)

// parser is the internal recursive-descent driver. It is not safe for
// concurrent use.
type parser struct {
	source         *ast.Source
	lex            *lexer.Lexer
	cfg            *config
	lastEnd        int // byte offset just past the last consumed token
	errors         []*ast.SyntaxError
	pendingLeading []*ast.Comment
}

type prodScope struct {
	p     *parser
	start int
}

func newParser(src *ast.Source, opts []Option) *parser {
	cfg := newConfig(opts)
	var lopts []lexer.Option
	if cfg.preserveComments {
		lopts = append(lopts, lexer.WithComments())
	}
	l := lexer.New(src, lopts...)
	return &parser{source: src, lex: l, cfg: cfg}
}

// peek returns the next non-comment token without consuming it. When
// comment preservation is on, any COMMENT tokens at the current position
// are drained into p.pendingLeading first.
func (p *parser) peek() (lexer.Token, error) {
	p.drainComments()
	return p.lex.Peek()
}

// advance consumes and returns the next non-comment token, updating lastEnd.
func (p *parser) advance() (lexer.Token, error) {
	p.drainComments()
	tok, err := p.lex.Next()
	if err != nil {
		return tok, err
	}
	p.lastEnd = tok.End
	return tok, nil
}

func (p *parser) enter() (prodScope, error) {
	tok, err := p.peek()
	if err != nil {
		return prodScope{}, err
	}
	return p.scopeAt(tok.Start), nil
}

func (p *parser) scopeAt(start int) prodScope {
	return prodScope{p: p, start: start}
}

func (s prodScope) close() *ast.Loc {
	return &ast.Loc{Start: s.start, End: s.p.lastEnd, Source: s.p.source}
}

// expect consumes the next token, returning an error if its kind is not k.
func (p *parser) expect(k lexer.Kind) (lexer.Token, error) {
	tok, err := p.peek()
	if err != nil {
		return tok, err
	}
	if tok.Kind != k {
		return tok, p.errAtTok(tok, fmt.Sprintf("Expected %s, found %s.", k, describeToken(tok)))
	}
	return p.advance()
}

// optional consumes the next token if it has kind k and returns whether it did.
func (p *parser) optional(k lexer.Kind) (lexer.Token, bool, error) {
	tok, err := p.peek()
	if err != nil {
		return tok, false, err
	}
	if tok.Kind != k {
		return tok, false, nil
	}
	got, err := p.advance()
	return got, err == nil, err
}

// expectKeyword consumes the next token and verifies it is a NAME token whose
// value equals kw. This is how the grammar dispatches on identifiers like
// "query", "type", "on", etc.
func (p *parser) expectKeyword(kw string) (lexer.Token, error) {
	tok, err := p.peek()
	if err != nil {
		return tok, err
	}
	if tok.Kind != lexer.NAME || tok.Value != kw {
		return tok, p.errAtTok(tok, fmt.Sprintf("Expected %q, found %s.", kw, describeToken(tok)))
	}
	return p.advance()
}

// optionalKeyword consumes a NAME token equal to kw and returns whether it did.
func (p *parser) optionalKeyword(kw string) (bool, error) {
	tok, err := p.peek()
	if err != nil {
		return false, err
	}
	if tok.Kind != lexer.NAME || tok.Value != kw {
		return false, nil
	}
	if _, err := p.advance(); err != nil {
		return false, err
	}
	return true, nil
}

// expectEOF errors if the next token is not EOF — used to validate that a
// partial-parse entry point (ParseValue, ParseType) consumed the entire input.
func (p *parser) expectEOF() error {
	tok, err := p.peek()
	if err != nil {
		return err
	}
	if tok.Kind != lexer.EOF {
		return p.errAtTok(tok, fmt.Sprintf("Expected <EOF>, found %s.", describeToken(tok)))
	}
	return nil
}

// errAtTok builds a SyntaxError pointing at a specific token.
func (p *parser) errAtTok(tok lexer.Token, msg string) *ast.SyntaxError {
	return &ast.SyntaxError{Source: p.source, Offset: tok.Start, Message: msg}
}

// describeToken renders a token for error messages.
func describeToken(tok lexer.Token) string {
	switch tok.Kind {
	case lexer.EOF:
		return "<EOF>"
	case lexer.NAME:
		return fmt.Sprintf("Name %q", tok.Value)
	case lexer.INT:
		return fmt.Sprintf("Int %q", tok.Value)
	case lexer.FLOAT:
		return fmt.Sprintf("Float %q", tok.Value)
	case lexer.STRING:
		if tok.Block {
			return "BlockString"
		}
		return "String"
	case lexer.COMMENT:
		return "Comment"
	default:
		return tok.Kind.String()
	}
}
