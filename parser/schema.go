package parser

import "github.com/bellbm/graphql-parser/ast"

// parseTypeSystemDefinitionOrExtension is a stub placeholder for phase 7,
// which will implement the schema/SDL grammar. For now it simply reports
// "no match" so the executable-document parser can fall through to its
// own error path.
//
// The boolean return is true when the parser recognized a type-system
// definition or extension and produced a Definition; false means the
// caller's dispatch should report a "definition not recognized" error.
func (p *parser) parseTypeSystemDefinitionOrExtension() (ast.Definition, bool, error) {
	return nil, false, nil
}
