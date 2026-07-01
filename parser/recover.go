package parser

import (
	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/lexer"
)

// definitionStartKeywords names every keyword that may begin a top-level
// Definition (executable + type-system, plus extensions).
var definitionStartKeywords = map[string]struct{}{
	"query": {}, "mutation": {}, "subscription": {}, "fragment": {},
	"schema": {}, "scalar": {}, "type": {}, "interface": {},
	"union": {}, "enum": {}, "input": {}, "directive": {}, "extend": {},
}

// isDefinitionStart reports whether tok could begin a Definition.
func isDefinitionStart(tok lexer.Token) bool {
	switch tok.Kind {
	case lexer.LBRACE: // shorthand operation
		return true
	case lexer.STRING: // leading description
		return true
	case lexer.NAME:
		_, ok := definitionStartKeywords[tok.Value]
		return ok
	}
	return false
}

// recordError accumulates err into p.errors. It returns true when recovery
// is enabled (caller may continue with a Bad* placeholder); false otherwise.
func (p *parser) recordError(err error) bool {
	if !p.cfg.recovery {
		return false
	}
	se, ok := err.(*ast.SyntaxError)
	if !ok {
		return false
	}
	p.errors = append(p.errors, se)
	return true
}

// skipToDefinitionStart advances past tokens until the next one could begin
// a Definition (or EOF). Used by recovery to resync after a malformed
// top-level definition.
func (p *parser) skipToDefinitionStart() {
	for {
		tok, err := p.peek()
		if err != nil {
			// Lexer error: try to consume one byte by attempting to advance.
			// If advance still errors, accumulate and stop.
			if !p.recordError(err) {
				return
			}
			if _, err2 := p.advance(); err2 != nil {
				_ = p.recordError(err2)
				return
			}
			continue
		}
		if tok.Kind == lexer.EOF {
			return
		}
		if isDefinitionStart(tok) {
			return
		}
		if _, err := p.advance(); err != nil {
			_ = p.recordError(err)
			return
		}
	}
}

// skipToSelectionStart advances past tokens until the next one could begin
// a Selection (or "}" / EOF). Used by recovery inside a SelectionSet.
func (p *parser) skipToSelectionStart() {
	for {
		tok, err := p.peek()
		if err != nil {
			if !p.recordError(err) {
				return
			}
			if _, err2 := p.advance(); err2 != nil {
				_ = p.recordError(err2)
				return
			}
			continue
		}
		if tok.Kind == lexer.EOF || tok.Kind == lexer.RBRACE ||
			tok.Kind == lexer.NAME || tok.Kind == lexer.SPREAD {
			return
		}
		if _, err := p.advance(); err != nil {
			_ = p.recordError(err)
			return
		}
	}
}

// skipToEOF advances to the end of a partial entry point after a root-level
// value/type failure. It intentionally does not try to recover nested grammar.
func (p *parser) skipToEOF() {
	for {
		tok, err := p.peek()
		if err != nil {
			// The caller already recorded the root failure that put the
			// partial entry into recovery. Clear a cached lexer error without
			// recording the same error again, then treat the rest of the entry
			// as covered by the root BadValue/BadType placeholder.
			_, _ = p.advance()
			p.lastEnd = len(p.source.Body)
			return
		}
		if tok.Kind == lexer.EOF {
			return
		}
		if _, err := p.advance(); err != nil {
			_ = p.recordError(err)
			return
		}
	}
}
