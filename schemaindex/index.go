// Package schemaindex indexes parsed GraphQL SDL type definitions.
//
// The index records top-level named type definitions and matching extensions.
// It does not validate schema semantics, merge duplicate definitions, or fold
// extensions into their base definitions.
package schemaindex

import "github.com/brian-bell/graphql-parser/ast"

// Index provides name-based lookup for parsed SDL type definitions.
type Index struct {
	types map[string]*TypeEntry
}

// New builds an index from doc. A nil document returns an empty index.
func New(doc *ast.Document) *Index {
	idx := &Index{types: make(map[string]*TypeEntry)}
	if doc == nil {
		return idx
	}
	for _, def := range doc.Definitions {
		if name, ok := baseTypeName(def); ok {
			entry := idx.entryFor(name)
			entry.baseDefinitions = append(entry.baseDefinitions, def)
			continue
		}

		if name, ok := extensionTypeName(def); ok {
			entry := idx.entryFor(name)
			entry.extensions = append(entry.extensions, def)
		}
	}
	return idx
}

func (idx *Index) entryFor(name string) *TypeEntry {
	entry := idx.types[name]
	if entry == nil {
		entry = &TypeEntry{name: name}
		idx.types[name] = entry
	}
	return entry
}

func baseTypeName(def ast.Definition) (string, bool) {
	switch d := def.(type) {
	case *ast.ScalarTypeDefinition:
		return d.Name, true
	case *ast.ObjectTypeDefinition:
		return d.Name, true
	case *ast.InterfaceTypeDefinition:
		return d.Name, true
	case *ast.UnionTypeDefinition:
		return d.Name, true
	case *ast.EnumTypeDefinition:
		return d.Name, true
	case *ast.InputObjectTypeDefinition:
		return d.Name, true
	default:
		return "", false
	}
}

func extensionTypeName(def ast.Definition) (string, bool) {
	switch d := def.(type) {
	case *ast.ScalarTypeExtension:
		return d.Name, true
	case *ast.ObjectTypeExtension:
		return d.Name, true
	case *ast.InterfaceTypeExtension:
		return d.Name, true
	case *ast.UnionTypeExtension:
		return d.Name, true
	case *ast.EnumTypeExtension:
		return d.Name, true
	case *ast.InputObjectTypeExtension:
		return d.Name, true
	default:
		return "", false
	}
}

// LookupType returns the indexed type entry for name, or nil when absent.
func (idx *Index) LookupType(name string) *TypeEntry {
	return idx.types[name]
}

// TypeEntry contains the parsed base definitions and extensions for one type name.
type TypeEntry struct {
	name            string
	baseDefinitions []ast.Definition
	extensions      []ast.Definition
}

// Name returns the type name for this entry.
func (e *TypeEntry) Name() string {
	return e.name
}

// BaseDefinitions returns a shallow copy of the parsed base type definitions
// for this entry. The AST nodes inside the returned slice are the original
// parsed nodes.
func (e *TypeEntry) BaseDefinitions() []ast.Definition {
	return copyDefinitions(e.baseDefinitions)
}

// Extensions returns a shallow copy of the parsed type extensions for this
// entry. The AST nodes inside the returned slice are the original parsed nodes.
func (e *TypeEntry) Extensions() []ast.Definition {
	return copyDefinitions(e.extensions)
}

func copyDefinitions(defs []ast.Definition) []ast.Definition {
	return append([]ast.Definition(nil), defs...)
}
