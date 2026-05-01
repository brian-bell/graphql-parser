// Package lexer tokenizes a GraphQL source document.
//
// Tokens hold byte offsets into the source rather than copied strings; for
// NAME, INT, and FLOAT tokens the Value field is a sub-slice of the source
// body (zero allocation). STRING tokens own their decoded value.
package lexer

// Kind identifies a token's lexical category. The zero value is reserved for
// "uninitialized" so that an empty Token can be distinguished from a real one.
type Kind uint8

const (
	invalid Kind = iota
	EOF
	BANG     // !
	DOLLAR   // $
	AMP      // &
	LPAREN   // (
	RPAREN   // )
	SPREAD   // ...
	COLON    // :
	EQUALS   // =
	AT       // @
	LBRACKET // [
	RBRACKET // ]
	LBRACE   // {
	PIPE     // |
	RBRACE   // }
	NAME
	INT
	FLOAT
	STRING
	COMMENT
)

var kindNames = [...]string{
	invalid:  "<invalid>",
	EOF:      "<EOF>",
	BANG:     "!",
	DOLLAR:   "$",
	AMP:      "&",
	LPAREN:   "(",
	RPAREN:   ")",
	SPREAD:   "...",
	COLON:    ":",
	EQUALS:   "=",
	AT:       "@",
	LBRACKET: "[",
	RBRACKET: "]",
	LBRACE:   "{",
	PIPE:     "|",
	RBRACE:   "}",
	NAME:     "Name",
	INT:      "Int",
	FLOAT:    "Float",
	STRING:   "String",
	COMMENT:  "Comment",
}

// String returns a human-readable name for the token kind, suitable for use
// in error messages.
func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "<unknown>"
}

// Token is a single lexical unit.
//
// Start and End are byte offsets into the source body; End is exclusive
// (half-open [Start, End)).
//
// Value is populated for NAME, INT, FLOAT, STRING, and COMMENT tokens.
// For NAME/INT/FLOAT it is a sub-slice of the source body.
// For STRING it is the escape-decoded (or block-string-dedented) value.
// For COMMENT it is the comment text (everything after the '#', up to but
// not including the line terminator).
//
// Block is set on STRING tokens that were written using the """...""" syntax.
type Token struct {
	Kind  Kind
	Start int
	End   int
	Value string
	Block bool
}
