package parser

import "github.com/bellbm/graphql-parser/ast"

// ParseError is the syntax-error type produced by all Parse* entry points.
// It is currently an alias for ast.SyntaxError; phase 8 enhances the shared
// Error() format with a graphql-js-style source-line snippet and caret pointer.
type ParseError = ast.SyntaxError
