package ast

// SchemaDefinition declares the root operation types of a schema.
type SchemaDefinition struct {
	Description    *StringValue
	Directives     DirectiveList
	OperationTypes []*OperationTypeDefinition
	Loc            *Loc
	Comments       *CommentGroup
}

func (d *SchemaDefinition) GetLoc() *Loc { return d.Loc }
func (*SchemaDefinition) isDefinition()  {}

// SchemaExtension extends a schema with additional directives and/or root
// operation types.
type SchemaExtension struct {
	Directives     DirectiveList
	OperationTypes []*OperationTypeDefinition
	Loc            *Loc
	Comments       *CommentGroup
}

func (d *SchemaExtension) GetLoc() *Loc { return d.Loc }
func (*SchemaExtension) isDefinition()  {}

// OperationTypeDefinition is one "query: Query" entry inside a SchemaDefinition.
type OperationTypeDefinition struct {
	Operation OperationType
	Type      *NamedType
	Loc       *Loc
	Comments  *CommentGroup
}

func (d *OperationTypeDefinition) GetLoc() *Loc { return d.Loc }

// ScalarTypeDefinition declares a custom scalar type.
type ScalarTypeDefinition struct {
	Description *StringValue
	Name        string
	Directives  DirectiveList
	Loc         *Loc
	Comments    *CommentGroup
}

func (d *ScalarTypeDefinition) GetLoc() *Loc { return d.Loc }
func (*ScalarTypeDefinition) isDefinition()  {}

// ScalarTypeExtension adds directives to an existing scalar type.
type ScalarTypeExtension struct {
	Name       string
	Directives DirectiveList
	Loc        *Loc
	Comments   *CommentGroup
}

func (d *ScalarTypeExtension) GetLoc() *Loc { return d.Loc }
func (*ScalarTypeExtension) isDefinition()  {}

// ObjectTypeDefinition declares a GraphQL object type.
type ObjectTypeDefinition struct {
	Description *StringValue
	Name        string
	Interfaces  []*NamedType
	Directives  DirectiveList
	Fields      FieldDefinitionList
	Loc         *Loc
	Comments    *CommentGroup
}

func (d *ObjectTypeDefinition) GetLoc() *Loc { return d.Loc }
func (*ObjectTypeDefinition) isDefinition()  {}

// ObjectTypeExtension extends an existing object type.
type ObjectTypeExtension struct {
	Name       string
	Interfaces []*NamedType
	Directives DirectiveList
	Fields     FieldDefinitionList
	Loc        *Loc
	Comments   *CommentGroup
}

func (d *ObjectTypeExtension) GetLoc() *Loc { return d.Loc }
func (*ObjectTypeExtension) isDefinition()  {}

// InterfaceTypeDefinition declares an interface type.
type InterfaceTypeDefinition struct {
	Description *StringValue
	Name        string
	Interfaces  []*NamedType
	Directives  DirectiveList
	Fields      FieldDefinitionList
	Loc         *Loc
	Comments    *CommentGroup
}

func (d *InterfaceTypeDefinition) GetLoc() *Loc { return d.Loc }
func (*InterfaceTypeDefinition) isDefinition()  {}

// InterfaceTypeExtension extends an interface type.
type InterfaceTypeExtension struct {
	Name       string
	Interfaces []*NamedType
	Directives DirectiveList
	Fields     FieldDefinitionList
	Loc        *Loc
	Comments   *CommentGroup
}

func (d *InterfaceTypeExtension) GetLoc() *Loc { return d.Loc }
func (*InterfaceTypeExtension) isDefinition()  {}

// UnionTypeDefinition declares a union type: "union U = A | B | C".
type UnionTypeDefinition struct {
	Description *StringValue
	Name        string
	Directives  DirectiveList
	Members     []*NamedType
	Loc         *Loc
	Comments    *CommentGroup
}

func (d *UnionTypeDefinition) GetLoc() *Loc { return d.Loc }
func (*UnionTypeDefinition) isDefinition()  {}

// UnionTypeExtension extends a union type.
type UnionTypeExtension struct {
	Name       string
	Directives DirectiveList
	Members    []*NamedType
	Loc        *Loc
	Comments   *CommentGroup
}

func (d *UnionTypeExtension) GetLoc() *Loc { return d.Loc }
func (*UnionTypeExtension) isDefinition()  {}

// EnumTypeDefinition declares an enum type.
type EnumTypeDefinition struct {
	Description *StringValue
	Name        string
	Directives  DirectiveList
	Values      EnumValueList
	Loc         *Loc
	Comments    *CommentGroup
}

func (d *EnumTypeDefinition) GetLoc() *Loc { return d.Loc }
func (*EnumTypeDefinition) isDefinition()  {}

// EnumTypeExtension extends an enum type.
type EnumTypeExtension struct {
	Name       string
	Directives DirectiveList
	Values     EnumValueList
	Loc        *Loc
	Comments   *CommentGroup
}

func (d *EnumTypeExtension) GetLoc() *Loc { return d.Loc }
func (*EnumTypeExtension) isDefinition()  {}

// InputObjectTypeDefinition declares an input object type.
type InputObjectTypeDefinition struct {
	Description *StringValue
	Name        string
	Directives  DirectiveList
	Fields      InputValueList
	Loc         *Loc
	Comments    *CommentGroup
}

func (d *InputObjectTypeDefinition) GetLoc() *Loc { return d.Loc }
func (*InputObjectTypeDefinition) isDefinition()  {}

// InputObjectTypeExtension extends an input object type.
type InputObjectTypeExtension struct {
	Name       string
	Directives DirectiveList
	Fields     InputValueList
	Loc        *Loc
	Comments   *CommentGroup
}

func (d *InputObjectTypeExtension) GetLoc() *Loc { return d.Loc }
func (*InputObjectTypeExtension) isDefinition()  {}

// FieldDefinition is one entry inside an Object or Interface type, e.g.
// "name(args): Type @directive...".
type FieldDefinition struct {
	Description *StringValue
	Name        string
	Arguments   InputValueList
	Type        Type
	Directives  DirectiveList
	Loc         *Loc
	Comments    *CommentGroup
}

func (d *FieldDefinition) GetLoc() *Loc { return d.Loc }

// InputValueDefinition is used for both argument definitions (on
// FieldDefinition / DirectiveDefinition) and input-object fields. It
// optionally carries a const default value.
type InputValueDefinition struct {
	Description  *StringValue
	Name         string
	Type         Type
	DefaultValue Value
	Directives   DirectiveList
	Loc          *Loc
	Comments     *CommentGroup
}

func (d *InputValueDefinition) GetLoc() *Loc { return d.Loc }

// EnumValueDefinition is one entry inside an EnumTypeDefinition. The Name
// must not be true, false, or null per spec.
type EnumValueDefinition struct {
	Description *StringValue
	Name        string
	Directives  DirectiveList
	Loc         *Loc
	Comments    *CommentGroup
}

func (d *EnumValueDefinition) GetLoc() *Loc { return d.Loc }

// DirectiveDefinition declares a directive: "directive @name(args) on LOCATIONS".
// Repeatable is true if the definition included the "repeatable" keyword.
type DirectiveDefinition struct {
	Description *StringValue
	Name        string
	Arguments   InputValueList
	Repeatable  bool
	Locations   []string
	Loc         *Loc
	Comments    *CommentGroup
}

func (d *DirectiveDefinition) GetLoc() *Loc { return d.Loc }
func (*DirectiveDefinition) isDefinition()  {}

// FieldDefinitionList is a slice of *FieldDefinition with helper methods.
type FieldDefinitionList []*FieldDefinition

// ForName returns the first field definition matching name, or nil.
func (fl FieldDefinitionList) ForName(name string) *FieldDefinition {
	for _, f := range fl {
		if f.Name == name {
			return f
		}
	}
	return nil
}

// InputValueList is a slice of *InputValueDefinition.
type InputValueList []*InputValueDefinition

// ForName returns the first input value matching name, or nil.
func (il InputValueList) ForName(name string) *InputValueDefinition {
	for _, v := range il {
		if v.Name == name {
			return v
		}
	}
	return nil
}

// EnumValueList is a slice of *EnumValueDefinition.
type EnumValueList []*EnumValueDefinition

// ForName returns the first enum value matching name, or nil.
func (el EnumValueList) ForName(name string) *EnumValueDefinition {
	for _, v := range el {
		if v.Name == name {
			return v
		}
	}
	return nil
}
