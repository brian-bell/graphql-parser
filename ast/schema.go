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
func (d *SchemaDefinition) Children() []Node {
	children := make([]Node, 0, len(d.Directives)+len(d.OperationTypes)+1)
	if d.Description != nil {
		children = append(children, d.Description)
	}
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	for _, op := range d.OperationTypes {
		if op != nil {
			children = append(children, op)
		}
	}
	return children
}
func (*SchemaDefinition) isDefinition()                 {}
func (d *SchemaDefinition) CommentSlot() **CommentGroup { return &d.Comments }

// SchemaExtension extends a schema with additional directives and/or root
// operation types.
type SchemaExtension struct {
	Directives     DirectiveList
	OperationTypes []*OperationTypeDefinition
	Loc            *Loc
	Comments       *CommentGroup
}

func (d *SchemaExtension) GetLoc() *Loc { return d.Loc }
func (d *SchemaExtension) Children() []Node {
	children := make([]Node, 0, len(d.Directives)+len(d.OperationTypes))
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	for _, op := range d.OperationTypes {
		if op != nil {
			children = append(children, op)
		}
	}
	return children
}
func (*SchemaExtension) isDefinition()                 {}
func (d *SchemaExtension) CommentSlot() **CommentGroup { return &d.Comments }

// OperationTypeDefinition is one "query: Query" entry inside a SchemaDefinition.
type OperationTypeDefinition struct {
	Operation OperationType
	Type      *NamedType
	Loc       *Loc
	Comments  *CommentGroup
}

func (d *OperationTypeDefinition) GetLoc() *Loc { return d.Loc }
func (d *OperationTypeDefinition) Children() []Node {
	if d.Type == nil {
		return nil
	}
	return []Node{d.Type}
}
func (d *OperationTypeDefinition) CommentSlot() **CommentGroup { return &d.Comments }

// ScalarTypeDefinition declares a custom scalar type.
type ScalarTypeDefinition struct {
	Description *StringValue
	Name        string
	Directives  DirectiveList
	Loc         *Loc
	Comments    *CommentGroup
}

func (d *ScalarTypeDefinition) GetLoc() *Loc { return d.Loc }
func (d *ScalarTypeDefinition) Children() []Node {
	children := make([]Node, 0, len(d.Directives)+1)
	if d.Description != nil {
		children = append(children, d.Description)
	}
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	return children
}
func (*ScalarTypeDefinition) isDefinition()                 {}
func (d *ScalarTypeDefinition) CommentSlot() **CommentGroup { return &d.Comments }

// ScalarTypeExtension adds directives to an existing scalar type.
type ScalarTypeExtension struct {
	Name       string
	Directives DirectiveList
	Loc        *Loc
	Comments   *CommentGroup
}

func (d *ScalarTypeExtension) GetLoc() *Loc { return d.Loc }
func (d *ScalarTypeExtension) Children() []Node {
	children := make([]Node, 0, len(d.Directives))
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	return children
}
func (*ScalarTypeExtension) isDefinition()                 {}
func (d *ScalarTypeExtension) CommentSlot() **CommentGroup { return &d.Comments }

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
func (d *ObjectTypeDefinition) Children() []Node {
	children := make([]Node, 0, len(d.Interfaces)+len(d.Directives)+len(d.Fields)+1)
	if d.Description != nil {
		children = append(children, d.Description)
	}
	for _, iface := range d.Interfaces {
		if iface != nil {
			children = append(children, iface)
		}
	}
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	for _, field := range d.Fields {
		if field != nil {
			children = append(children, field)
		}
	}
	return children
}
func (*ObjectTypeDefinition) isDefinition()                 {}
func (d *ObjectTypeDefinition) CommentSlot() **CommentGroup { return &d.Comments }

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
func (d *ObjectTypeExtension) Children() []Node {
	children := make([]Node, 0, len(d.Interfaces)+len(d.Directives)+len(d.Fields))
	for _, iface := range d.Interfaces {
		if iface != nil {
			children = append(children, iface)
		}
	}
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	for _, field := range d.Fields {
		if field != nil {
			children = append(children, field)
		}
	}
	return children
}
func (*ObjectTypeExtension) isDefinition()                 {}
func (d *ObjectTypeExtension) CommentSlot() **CommentGroup { return &d.Comments }

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
func (d *InterfaceTypeDefinition) Children() []Node {
	children := make([]Node, 0, len(d.Interfaces)+len(d.Directives)+len(d.Fields)+1)
	if d.Description != nil {
		children = append(children, d.Description)
	}
	for _, iface := range d.Interfaces {
		if iface != nil {
			children = append(children, iface)
		}
	}
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	for _, field := range d.Fields {
		if field != nil {
			children = append(children, field)
		}
	}
	return children
}
func (*InterfaceTypeDefinition) isDefinition()                 {}
func (d *InterfaceTypeDefinition) CommentSlot() **CommentGroup { return &d.Comments }

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
func (d *InterfaceTypeExtension) Children() []Node {
	children := make([]Node, 0, len(d.Interfaces)+len(d.Directives)+len(d.Fields))
	for _, iface := range d.Interfaces {
		if iface != nil {
			children = append(children, iface)
		}
	}
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	for _, field := range d.Fields {
		if field != nil {
			children = append(children, field)
		}
	}
	return children
}
func (*InterfaceTypeExtension) isDefinition()                 {}
func (d *InterfaceTypeExtension) CommentSlot() **CommentGroup { return &d.Comments }

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
func (d *UnionTypeDefinition) Children() []Node {
	children := make([]Node, 0, len(d.Directives)+len(d.Members)+1)
	if d.Description != nil {
		children = append(children, d.Description)
	}
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	for _, member := range d.Members {
		if member != nil {
			children = append(children, member)
		}
	}
	return children
}
func (*UnionTypeDefinition) isDefinition()                 {}
func (d *UnionTypeDefinition) CommentSlot() **CommentGroup { return &d.Comments }

// UnionTypeExtension extends a union type.
type UnionTypeExtension struct {
	Name       string
	Directives DirectiveList
	Members    []*NamedType
	Loc        *Loc
	Comments   *CommentGroup
}

func (d *UnionTypeExtension) GetLoc() *Loc { return d.Loc }
func (d *UnionTypeExtension) Children() []Node {
	children := make([]Node, 0, len(d.Directives)+len(d.Members))
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	for _, member := range d.Members {
		if member != nil {
			children = append(children, member)
		}
	}
	return children
}
func (*UnionTypeExtension) isDefinition()                 {}
func (d *UnionTypeExtension) CommentSlot() **CommentGroup { return &d.Comments }

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
func (d *EnumTypeDefinition) Children() []Node {
	children := make([]Node, 0, len(d.Directives)+len(d.Values)+1)
	if d.Description != nil {
		children = append(children, d.Description)
	}
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	for _, value := range d.Values {
		if value != nil {
			children = append(children, value)
		}
	}
	return children
}
func (*EnumTypeDefinition) isDefinition()                 {}
func (d *EnumTypeDefinition) CommentSlot() **CommentGroup { return &d.Comments }

// EnumTypeExtension extends an enum type.
type EnumTypeExtension struct {
	Name       string
	Directives DirectiveList
	Values     EnumValueList
	Loc        *Loc
	Comments   *CommentGroup
}

func (d *EnumTypeExtension) GetLoc() *Loc { return d.Loc }
func (d *EnumTypeExtension) Children() []Node {
	children := make([]Node, 0, len(d.Directives)+len(d.Values))
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	for _, value := range d.Values {
		if value != nil {
			children = append(children, value)
		}
	}
	return children
}
func (*EnumTypeExtension) isDefinition()                 {}
func (d *EnumTypeExtension) CommentSlot() **CommentGroup { return &d.Comments }

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
func (d *InputObjectTypeDefinition) Children() []Node {
	children := make([]Node, 0, len(d.Directives)+len(d.Fields)+1)
	if d.Description != nil {
		children = append(children, d.Description)
	}
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	for _, field := range d.Fields {
		if field != nil {
			children = append(children, field)
		}
	}
	return children
}
func (*InputObjectTypeDefinition) isDefinition()                 {}
func (d *InputObjectTypeDefinition) CommentSlot() **CommentGroup { return &d.Comments }

// InputObjectTypeExtension extends an input object type.
type InputObjectTypeExtension struct {
	Name       string
	Directives DirectiveList
	Fields     InputValueList
	Loc        *Loc
	Comments   *CommentGroup
}

func (d *InputObjectTypeExtension) GetLoc() *Loc { return d.Loc }
func (d *InputObjectTypeExtension) Children() []Node {
	children := make([]Node, 0, len(d.Directives)+len(d.Fields))
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	for _, field := range d.Fields {
		if field != nil {
			children = append(children, field)
		}
	}
	return children
}
func (*InputObjectTypeExtension) isDefinition()                 {}
func (d *InputObjectTypeExtension) CommentSlot() **CommentGroup { return &d.Comments }

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
func (d *FieldDefinition) Children() []Node {
	children := make([]Node, 0, len(d.Arguments)+len(d.Directives)+2)
	if d.Description != nil {
		children = append(children, d.Description)
	}
	for _, arg := range d.Arguments {
		if arg != nil {
			children = append(children, arg)
		}
	}
	if d.Type != nil {
		children = append(children, d.Type)
	}
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	return children
}
func (d *FieldDefinition) CommentSlot() **CommentGroup { return &d.Comments }

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
func (d *InputValueDefinition) Children() []Node {
	children := make([]Node, 0, len(d.Directives)+3)
	if d.Description != nil {
		children = append(children, d.Description)
	}
	if d.Type != nil {
		children = append(children, d.Type)
	}
	if d.DefaultValue != nil {
		children = append(children, d.DefaultValue)
	}
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	return children
}
func (d *InputValueDefinition) CommentSlot() **CommentGroup { return &d.Comments }

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
func (d *EnumValueDefinition) Children() []Node {
	children := make([]Node, 0, len(d.Directives)+1)
	if d.Description != nil {
		children = append(children, d.Description)
	}
	for _, dir := range d.Directives {
		if dir != nil {
			children = append(children, dir)
		}
	}
	return children
}
func (d *EnumValueDefinition) CommentSlot() **CommentGroup { return &d.Comments }

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
func (d *DirectiveDefinition) Children() []Node {
	children := make([]Node, 0, len(d.Arguments)+1)
	if d.Description != nil {
		children = append(children, d.Description)
	}
	for _, arg := range d.Arguments {
		if arg != nil {
			children = append(children, arg)
		}
	}
	return children
}
func (*DirectiveDefinition) isDefinition()                 {}
func (d *DirectiveDefinition) CommentSlot() **CommentGroup { return &d.Comments }

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
