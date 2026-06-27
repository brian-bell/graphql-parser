package ast

// Document is the top-level AST node — a sequence of one or more Definitions.
//
// Per spec, a Document may contain a mix of executable definitions
// (operations, fragments) and type-system definitions/extensions; the parser
// does not enforce a "queries only" or "schema only" partitioning.
type Document struct {
	Definitions DefinitionList
	Loc         *Loc
	Comments    *CommentGroup
}

// GetLoc returns the document's location.
func (d *Document) GetLoc() *Loc { return d.Loc }

// DefinitionList is a slice of Definition with helper methods.
type DefinitionList []Definition

// ArgumentList is a slice of *Argument.
type ArgumentList []*Argument

// ForName returns the first argument named name, or nil if none exists.
func (al ArgumentList) ForName(name string) *Argument {
	for _, a := range al {
		if a.Name == name {
			return a
		}
	}
	return nil
}

// DirectiveList is a slice of *Directive.
type DirectiveList []*Directive

// ForName returns the first directive named name, or nil if none exists.
func (dl DirectiveList) ForName(name string) *Directive {
	for _, d := range dl {
		if d.Name == name {
			return d
		}
	}
	return nil
}

// DirectiveStringArg returns the decoded string value for the first matching
// directive argument. It returns ("", false) when the directive or argument is
// absent, nil, or not a StringValue. An explicit empty string returns ("",
// true).
func DirectiveStringArg(dirs DirectiveList, directiveName, argumentName string) (string, bool) {
	for _, dir := range dirs {
		if dir == nil || dir.Name != directiveName {
			continue
		}
		for _, arg := range dir.Arguments {
			if arg == nil || arg.Name != argumentName {
				continue
			}
			value, ok := arg.Value.(*StringValue)
			if !ok || value == nil {
				return "", false
			}
			return value.Value, true
		}
		return "", false
	}
	return "", false
}

// VariableDefinitionList is a slice of *VariableDefinition.
type VariableDefinitionList []*VariableDefinition

// ForName returns the first variable definition matching name (with no
// leading '$'), or nil if none exists.
func (vl VariableDefinitionList) ForName(name string) *VariableDefinition {
	for _, v := range vl {
		if v.Variable != nil && v.Variable.Name == name {
			return v
		}
	}
	return nil
}
