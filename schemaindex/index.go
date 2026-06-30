// Package schemaindex indexes parsed GraphQL SDL type definitions.
//
// The index records top-level named type definitions and matching extensions.
// Raw base definitions and extensions stay separate, while helper accessors
// expose base object, interface, input, enum, union, and scalar metadata
// followed by matching extension metadata. It does not validate schema
// semantics, merge duplicate definitions, or deduplicate folded members.
package schemaindex

import "github.com/brian-bell/graphql-parser/ast"

// Index provides name-based lookup for parsed SDL type definitions.
type Index struct {
	types map[string]*TypeEntry
	names []string
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
		idx.names = append(idx.names, name)
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

// TypeNames returns indexed type names in first-seen document order.
func (idx *Index) TypeNames() []string {
	return append([]string(nil), idx.names...)
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

// ObjectFields returns object fields from base definitions followed by matching
// object extensions, each in source order.
func (e *TypeEntry) ObjectFields() ast.FieldDefinitionList {
	var fields ast.FieldDefinitionList
	for _, def := range e.baseDefinitions {
		if objectDef, ok := def.(*ast.ObjectTypeDefinition); ok {
			fields = append(fields, objectDef.Fields...)
		}
	}
	for _, def := range e.extensions {
		if objectExt, ok := def.(*ast.ObjectTypeExtension); ok {
			fields = append(fields, objectExt.Fields...)
		}
	}
	return fields
}

// ObjectInterfaces returns object implemented interfaces from base definitions
// followed by matching object extensions, each in source order.
func (e *TypeEntry) ObjectInterfaces() []*ast.NamedType {
	var interfaces []*ast.NamedType
	for _, def := range e.baseDefinitions {
		if objectDef, ok := def.(*ast.ObjectTypeDefinition); ok {
			interfaces = append(interfaces, objectDef.Interfaces...)
		}
	}
	for _, def := range e.extensions {
		if objectExt, ok := def.(*ast.ObjectTypeExtension); ok {
			interfaces = append(interfaces, objectExt.Interfaces...)
		}
	}
	return interfaces
}

// InterfaceInterfaces returns interface implemented interfaces from base
// definitions followed by matching interface extensions, each in source order.
func (e *TypeEntry) InterfaceInterfaces() []*ast.NamedType {
	var interfaces []*ast.NamedType
	for _, def := range e.baseDefinitions {
		if interfaceDef, ok := def.(*ast.InterfaceTypeDefinition); ok {
			interfaces = append(interfaces, interfaceDef.Interfaces...)
		}
	}
	for _, def := range e.extensions {
		if interfaceExt, ok := def.(*ast.InterfaceTypeExtension); ok {
			interfaces = append(interfaces, interfaceExt.Interfaces...)
		}
	}
	return interfaces
}

// InterfaceFields returns interface fields from base definitions followed by
// matching interface extensions, each in source order.
func (e *TypeEntry) InterfaceFields() ast.FieldDefinitionList {
	var fields ast.FieldDefinitionList
	for _, def := range e.baseDefinitions {
		if interfaceDef, ok := def.(*ast.InterfaceTypeDefinition); ok {
			fields = append(fields, interfaceDef.Fields...)
		}
	}
	for _, def := range e.extensions {
		if interfaceExt, ok := def.(*ast.InterfaceTypeExtension); ok {
			fields = append(fields, interfaceExt.Fields...)
		}
	}
	return fields
}

// InputFields returns input object fields from base definitions followed by
// matching input object extensions, each in source order.
func (e *TypeEntry) InputFields() ast.InputValueList {
	var fields ast.InputValueList
	for _, def := range e.baseDefinitions {
		if inputDef, ok := def.(*ast.InputObjectTypeDefinition); ok {
			fields = append(fields, inputDef.Fields...)
		}
	}
	for _, def := range e.extensions {
		if inputExt, ok := def.(*ast.InputObjectTypeExtension); ok {
			fields = append(fields, inputExt.Fields...)
		}
	}
	return fields
}

// EnumValues returns enum values from base definitions followed by matching
// enum extensions, each in source order.
func (e *TypeEntry) EnumValues() ast.EnumValueList {
	var values ast.EnumValueList
	for _, def := range e.baseDefinitions {
		if enumDef, ok := def.(*ast.EnumTypeDefinition); ok {
			values = append(values, enumDef.Values...)
		}
	}
	for _, def := range e.extensions {
		if enumExt, ok := def.(*ast.EnumTypeExtension); ok {
			values = append(values, enumExt.Values...)
		}
	}
	return values
}

// UnionMembers returns union members from base definitions followed by matching
// union extensions, each in source order.
func (e *TypeEntry) UnionMembers() []*ast.NamedType {
	var members []*ast.NamedType
	for _, def := range e.baseDefinitions {
		if unionDef, ok := def.(*ast.UnionTypeDefinition); ok {
			members = append(members, unionDef.Members...)
		}
	}
	for _, def := range e.extensions {
		if unionExt, ok := def.(*ast.UnionTypeExtension); ok {
			members = append(members, unionExt.Members...)
		}
	}
	return members
}

// ScalarDirectives returns scalar directives from base definitions followed by
// matching scalar extensions, each in source order.
func (e *TypeEntry) ScalarDirectives() ast.DirectiveList {
	var directives ast.DirectiveList
	for _, def := range e.baseDefinitions {
		if scalarDef, ok := def.(*ast.ScalarTypeDefinition); ok {
			directives = append(directives, scalarDef.Directives...)
		}
	}
	for _, def := range e.extensions {
		if scalarExt, ok := def.(*ast.ScalarTypeExtension); ok {
			directives = append(directives, scalarExt.Directives...)
		}
	}
	return directives
}

func copyDefinitions(defs []ast.Definition) []ast.Definition {
	return append([]ast.Definition(nil), defs...)
}
