package lexer

import (
	"fmt"
	"unicode/utf8"

	"github.com/bellbm/graphql-parser/ast"
)

// Lexer is a synchronous, single-token-lookahead tokenizer for a [ast.Source].
// It is not safe for concurrent use; callers should construct a fresh Lexer
// per source.
type Lexer struct {
	source *ast.Source
	body   string // == source.Body, cached for the hot loop
	pos    int    // current byte offset into body

	// PreserveComments controls whether COMMENT tokens reach the caller.
	// When false (the default), the lexer skips comments silently as required
	// by the GraphQL grammar (they are "ignored tokens"). When true, each
	// comment is emitted as a COMMENT token whose Value is the text after
	// '#' up to (but not including) the line terminator.
	PreserveComments bool

	peeked    bool
	peekedTok Token
	peekedErr error
}

// New constructs a Lexer for the given source. The source must outlive the
// Lexer; tokens hold byte offsets into Source.Body.
func New(src *ast.Source) *Lexer {
	return &Lexer{source: src, body: src.Body}
}

// Source returns the underlying source.
func (l *Lexer) Source() *ast.Source { return l.source }

// Next consumes and returns the next token. At end of input it returns an
// EOF token; subsequent calls continue to return EOF.
func (l *Lexer) Next() (Token, error) {
	if l.peeked {
		tok, err := l.peekedTok, l.peekedErr
		l.peeked = false
		l.peekedTok = Token{}
		l.peekedErr = nil
		return tok, err
	}
	return l.lex()
}

// Peek returns the next token without consuming it. Repeated calls without
// an intervening Next return the same token.
func (l *Lexer) Peek() (Token, error) {
	if l.peeked {
		return l.peekedTok, l.peekedErr
	}
	tok, err := l.lex()
	l.peeked = true
	l.peekedTok = tok
	l.peekedErr = err
	return tok, err
}

func (l *Lexer) lex() (Token, error) {
	for {
		l.skipIgnored()
		if l.pos >= len(l.body) {
			return Token{Kind: EOF, Start: l.pos, End: l.pos}, nil
		}
		c := l.body[l.pos]
		switch c {
		case '!':
			return l.singleByte(BANG), nil
		case '$':
			return l.singleByte(DOLLAR), nil
		case '&':
			return l.singleByte(AMP), nil
		case '(':
			return l.singleByte(LPAREN), nil
		case ')':
			return l.singleByte(RPAREN), nil
		case ':':
			return l.singleByte(COLON), nil
		case '=':
			return l.singleByte(EQUALS), nil
		case '@':
			return l.singleByte(AT), nil
		case '[':
			return l.singleByte(LBRACKET), nil
		case ']':
			return l.singleByte(RBRACKET), nil
		case '{':
			return l.singleByte(LBRACE), nil
		case '|':
			return l.singleByte(PIPE), nil
		case '}':
			return l.singleByte(RBRACE), nil
		case '.':
			if l.pos+2 < len(l.body) && l.body[l.pos+1] == '.' && l.body[l.pos+2] == '.' {
				tok := Token{Kind: SPREAD, Start: l.pos, End: l.pos + 3}
				l.pos += 3
				return tok, nil
			}
			return Token{}, l.errAt(l.pos, fmt.Sprintf("Cannot parse the unexpected character %s.", quoteChar(c)))
		case '#':
			tok, ok, err := l.lexComment()
			if err != nil {
				return Token{}, err
			}
			if ok {
				return tok, nil
			}
			continue
		case '"':
			return l.lexString()
		}
		if c == '-' || isDigit(c) {
			return l.lexNumber()
		}
		if isNameStart(c) {
			return l.lexName(), nil
		}
		return Token{}, l.errAt(l.pos, fmt.Sprintf("Cannot parse the unexpected character %s.", quoteChar(c)))
	}
}

func (l *Lexer) singleByte(k Kind) Token {
	tok := Token{Kind: k, Start: l.pos, End: l.pos + 1}
	l.pos++
	return tok
}

// skipIgnored advances past ignored tokens: BOM, whitespace, line terminators,
// commas, and (when PreserveComments is false) comments. Comments are always
// lexed for their byte range; this method just skips them when retention is off.
func (l *Lexer) skipIgnored() {
	for l.pos < len(l.body) {
		c := l.body[l.pos]
		switch c {
		case ' ', '\t', ',', '\n', '\r':
			l.pos++
		case 0xEF:
			// UTF-8 BOM: EF BB BF
			if l.pos+2 < len(l.body) && l.body[l.pos+1] == 0xBB && l.body[l.pos+2] == 0xBF {
				l.pos += 3
				continue
			}
			return
		case '#':
			if l.PreserveComments {
				return
			}
			// Skip to end of line; do not consume the terminator (the loop
			// above will absorb it as whitespace).
			for l.pos < len(l.body) && l.body[l.pos] != '\n' && l.body[l.pos] != '\r' {
				l.pos++
			}
		default:
			return
		}
	}
}

// lexComment reads a # comment, starting at the '#'. The returned token's
// Value is the comment text (after '#', before the line terminator).
// It returns ok=false if PreserveComments is false; in that case the caller
// should resume looking for the next non-ignored token.
func (l *Lexer) lexComment() (Token, bool, error) {
	if !l.PreserveComments {
		// skipIgnored should have absorbed it; defensive.
		for l.pos < len(l.body) && l.body[l.pos] != '\n' && l.body[l.pos] != '\r' {
			l.pos++
		}
		return Token{}, false, nil
	}
	start := l.pos
	l.pos++ // consume '#'
	textStart := l.pos
	for l.pos < len(l.body) && l.body[l.pos] != '\n' && l.body[l.pos] != '\r' {
		l.pos++
	}
	return Token{
		Kind:  COMMENT,
		Start: start,
		End:   l.pos,
		Value: l.body[textStart:l.pos],
	}, true, nil
}

func (l *Lexer) lexName() Token {
	start := l.pos
	l.pos++ // first char already passed isNameStart
	for l.pos < len(l.body) && isNameContinue(l.body[l.pos]) {
		l.pos++
	}
	return Token{
		Kind:  NAME,
		Start: start,
		End:   l.pos,
		Value: l.body[start:l.pos],
	}
}

func (l *Lexer) lexNumber() (Token, error) {
	start := l.pos
	if l.body[l.pos] == '-' {
		l.pos++
		if l.pos >= len(l.body) {
			return Token{}, l.errAt(l.pos, "Invalid number, expected digit but got: <EOF>.")
		}
	}
	c := l.body[l.pos]
	if c == '0' {
		l.pos++
		// Reject leading-zero integers like "01", "012".
		if l.pos < len(l.body) && isDigit(l.body[l.pos]) {
			return Token{}, l.errAt(l.pos, fmt.Sprintf("Invalid number, unexpected digit after 0: %s.", quoteChar(l.body[l.pos])))
		}
	} else if isDigit(c) {
		for l.pos < len(l.body) && isDigit(l.body[l.pos]) {
			l.pos++
		}
	} else {
		return Token{}, l.errAt(l.pos, fmt.Sprintf("Invalid number, expected digit but got: %s.", quoteChar(c)))
	}

	isFloat := false

	if l.pos < len(l.body) && l.body[l.pos] == '.' {
		isFloat = true
		l.pos++
		if l.pos >= len(l.body) || !isDigit(l.body[l.pos]) {
			return Token{}, l.errAt(l.pos, fmt.Sprintf("Invalid number, expected digit but got: %s.", l.describeAt(l.pos)))
		}
		for l.pos < len(l.body) && isDigit(l.body[l.pos]) {
			l.pos++
		}
	}

	if l.pos < len(l.body) && (l.body[l.pos] == 'e' || l.body[l.pos] == 'E') {
		isFloat = true
		l.pos++
		if l.pos < len(l.body) && (l.body[l.pos] == '+' || l.body[l.pos] == '-') {
			l.pos++
		}
		if l.pos >= len(l.body) || !isDigit(l.body[l.pos]) {
			return Token{}, l.errAt(l.pos, fmt.Sprintf("Invalid number, expected digit but got: %s.", l.describeAt(l.pos)))
		}
		for l.pos < len(l.body) && isDigit(l.body[l.pos]) {
			l.pos++
		}
	}

	// A number must not be immediately followed by a NameStart character or
	// another '.', which would indicate a malformed token like "123abc" or
	// "1.0.0".
	if l.pos < len(l.body) {
		c := l.body[l.pos]
		if c == '.' || isNameStart(c) {
			return Token{}, l.errAt(l.pos, fmt.Sprintf("Invalid number, expected digit but got: %s.", quoteChar(c)))
		}
	}

	kind := INT
	if isFloat {
		kind = FLOAT
	}
	return Token{
		Kind:  kind,
		Start: start,
		End:   l.pos,
		Value: l.body[start:l.pos],
	}, nil
}

// describeAt returns a quoted character at offset, or "<EOF>" if past the end.
func (l *Lexer) describeAt(offset int) string {
	if offset >= len(l.body) {
		return "<EOF>"
	}
	return quoteChar(l.body[offset])
}

func (l *Lexer) errAt(offset int, msg string) *ast.SyntaxError {
	return &ast.SyntaxError{Source: l.source, Offset: offset, Message: msg}
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isNameStart(c byte) bool {
	return c == '_' || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isNameContinue(c byte) bool {
	return isNameStart(c) || isDigit(c)
}

// quoteChar formats a byte for inclusion in error messages, escaping
// non-printables. It accepts a byte but renders it as a Unicode-escape if it
// is the leading byte of a multi-byte sequence, since the lexer typically
// errors on the first byte of an invalid character.
func quoteChar(c byte) string {
	if c >= utf8.RuneSelf {
		return fmt.Sprintf("U+%04X", c)
	}
	if c < 0x20 || c == 0x7F {
		return fmt.Sprintf("U+%04X", c)
	}
	return fmt.Sprintf("%q", string(c))
}
