package lexer

import (
	"fmt"
	"unicode/utf8"

	"github.com/bellbm/graphql-parser/ast"
)

// lexString reads a string token starting at l.pos (which must point at '"').
// It dispatches to lexBlockString for """...""" syntax.
func (l *Lexer) lexString() (Token, error) {
	if l.pos+2 < len(l.body) && l.body[l.pos+1] == '"' && l.body[l.pos+2] == '"' {
		return l.lexBlockString()
	}
	return l.lexSimpleString()
}

func (l *Lexer) lexSimpleString() (Token, error) {
	start := l.pos
	l.pos++ // consume opening "

	var buf []byte
	runStart := l.pos

	for l.pos < len(l.body) {
		c := l.body[l.pos]
		switch {
		case c == '"':
			value := l.body[runStart:l.pos]
			if buf != nil {
				buf = append(buf, value...)
				value = string(buf)
			}
			l.pos++
			return Token{Kind: STRING, Start: start, End: l.pos, Value: value}, nil

		case c == '\\':
			if buf == nil {
				buf = make([]byte, 0, l.pos-start)
			}
			buf = append(buf, l.body[runStart:l.pos]...)
			l.pos++ // consume backslash
			if err := l.decodeEscape(&buf); err != nil {
				return Token{}, err
			}
			runStart = l.pos

		case c == '\n' || c == '\r':
			return Token{}, l.errAt(l.pos, "Unterminated string.")

		case c < 0x20 && c != '\t':
			return Token{}, l.errAt(l.pos, fmt.Sprintf("Invalid character within String: %s.", quoteChar(c)))

		default:
			l.pos++
		}
	}
	return Token{}, l.errAt(l.pos, "Unterminated string.")
}

// decodeEscape consumes a backslash-escape (the backslash itself has already
// been consumed by the caller) and appends the decoded bytes to dst.
func (l *Lexer) decodeEscape(dst *[]byte) error {
	if l.pos >= len(l.body) {
		return l.errAt(l.pos, "Unterminated string.")
	}
	c := l.body[l.pos]
	switch c {
	case '"':
		l.pos++
		*dst = append(*dst, '"')
	case '\\':
		l.pos++
		*dst = append(*dst, '\\')
	case '/':
		l.pos++
		*dst = append(*dst, '/')
	case 'b':
		l.pos++
		*dst = append(*dst, '\b')
	case 'f':
		l.pos++
		*dst = append(*dst, '\f')
	case 'n':
		l.pos++
		*dst = append(*dst, '\n')
	case 'r':
		l.pos++
		*dst = append(*dst, '\r')
	case 't':
		l.pos++
		*dst = append(*dst, '\t')
	case 'u':
		l.pos++
		return l.decodeUnicodeEscape(dst)
	default:
		// Caller has already consumed the backslash; report the error at it.
		return l.errAt(l.pos-1, fmt.Sprintf("Invalid character escape sequence: \"\\%s\".", string(c)))
	}
	return nil
}

func (l *Lexer) decodeUnicodeEscape(dst *[]byte) error {
	if l.pos < len(l.body) && l.body[l.pos] == '{' {
		return l.decodeVariableUnicodeEscape(dst)
	}
	return l.decodeFixedUnicodeEscape(dst)
}

func (l *Lexer) decodeFixedUnicodeEscape(dst *[]byte) error {
	// l.pos is just past 'u'. Need exactly 4 hex digits.
	escapeStart := l.pos - 2 // points at '\\'
	if l.pos+4 > len(l.body) {
		return l.errAt(escapeStart, "Invalid Unicode escape sequence: too short.")
	}
	cp, ok := parseHex(l.body[l.pos : l.pos+4])
	if !ok {
		return l.errAt(escapeStart, fmt.Sprintf("Invalid Unicode escape sequence: \"\\u%s\".", l.body[l.pos:l.pos+4]))
	}
	l.pos += 4

	// Handle a high surrogate followed by a fixed low-surrogate escape.
	if cp >= 0xD800 && cp <= 0xDBFF {
		if l.pos+6 <= len(l.body) &&
			l.body[l.pos] == '\\' && l.body[l.pos+1] == 'u' && l.body[l.pos+2] != '{' {
			lo, ok2 := parseHex(l.body[l.pos+2 : l.pos+6])
			if ok2 && lo >= 0xDC00 && lo <= 0xDFFF {
				cp = 0x10000 + ((cp - 0xD800) << 10) + (lo - 0xDC00)
				l.pos += 6
				return appendRune(dst, rune(cp))
			}
		}
		return l.errAt(escapeStart, "Invalid Unicode escape sequence: lone high surrogate.")
	}
	if cp >= 0xDC00 && cp <= 0xDFFF {
		return l.errAt(escapeStart, "Invalid Unicode escape sequence: lone low surrogate.")
	}
	return appendRune(dst, rune(cp))
}

func (l *Lexer) decodeVariableUnicodeEscape(dst *[]byte) error {
	escapeStart := l.pos - 2 // points at '\\'
	l.pos++                  // consume '{'
	hexStart := l.pos
	for l.pos < len(l.body) && l.body[l.pos] != '}' {
		l.pos++
	}
	if l.pos >= len(l.body) {
		return l.errAt(escapeStart, "Invalid Unicode escape sequence: unterminated.")
	}
	hex := l.body[hexStart:l.pos]
	l.pos++ // consume '}'
	if len(hex) == 0 || len(hex) > 6 {
		return l.errAt(escapeStart, fmt.Sprintf("Invalid Unicode escape sequence: \"\\u{%s}\".", hex))
	}
	cp, ok := parseHex(hex)
	if !ok || cp > 0x10FFFF {
		return l.errAt(escapeStart, fmt.Sprintf("Invalid Unicode escape sequence: \"\\u{%s}\".", hex))
	}
	if cp >= 0xD800 && cp <= 0xDFFF {
		return l.errAt(escapeStart, fmt.Sprintf("Invalid Unicode escape sequence: \"\\u{%s}\" (surrogate).", hex))
	}
	return appendRune(dst, rune(cp))
}

// parseHex parses up to 6 hex digits as a uint32. Returns false if any byte
// is not a hex digit.
func parseHex(s string) (uint32, bool) {
	var v uint32
	for i := 0; i < len(s); i++ {
		c := s[i]
		var d uint32
		switch {
		case c >= '0' && c <= '9':
			d = uint32(c - '0')
		case c >= 'a' && c <= 'f':
			d = uint32(c-'a') + 10
		case c >= 'A' && c <= 'F':
			d = uint32(c-'A') + 10
		default:
			return 0, false
		}
		v = v<<4 | d
	}
	return v, true
}

func appendRune(dst *[]byte, r rune) error {
	if !utf8.ValidRune(r) {
		return fmt.Errorf("invalid rune U+%04X after escape decoding", r)
	}
	var buf [4]byte
	n := utf8.EncodeRune(buf[:], r)
	*dst = append(*dst, buf[:n]...)
	return nil
}

// lexBlockString reads a """...""" string. The only escape inside a block
// string is \""" (which becomes literal """); all other characters are taken
// literally, including line terminators and backslashes.
func (l *Lexer) lexBlockString() (Token, error) {
	start := l.pos
	l.pos += 3 // consume opening """

	var buf []byte
	runStart := l.pos

	for l.pos < len(l.body) {
		// End of block string: """
		if l.body[l.pos] == '"' &&
			l.pos+2 < len(l.body) &&
			l.body[l.pos+1] == '"' &&
			l.body[l.pos+2] == '"' {
			raw := l.body[runStart:l.pos]
			if buf != nil {
				buf = append(buf, raw...)
				raw = string(buf)
			}
			l.pos += 3
			return Token{
				Kind:  STRING,
				Block: true,
				Start: start,
				End:   l.pos,
				Value: ast.BlockStringValue(raw),
			}, nil
		}
		// Escaped triple quote: \""" → """
		if l.body[l.pos] == '\\' &&
			l.pos+3 < len(l.body) &&
			l.body[l.pos+1] == '"' &&
			l.body[l.pos+2] == '"' &&
			l.body[l.pos+3] == '"' {
			if buf == nil {
				buf = make([]byte, 0, l.pos-start)
			}
			buf = append(buf, l.body[runStart:l.pos]...)
			buf = append(buf, '"', '"', '"')
			l.pos += 4
			runStart = l.pos
			continue
		}
		l.pos++
	}
	return Token{}, l.errAt(l.pos, "Unterminated block string.")
}
