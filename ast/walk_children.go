package ast

func (d *Document) walkChildren(v Visitor) {
	walkDefinitions(v, d.Definitions)
}

func (d *OperationDefinition) walkChildren(v Visitor) {
	walkVariableDefinitions(v, d.VariableDefinitions)
	walkDirectives(v, d.Directives)
	if d.SelectionSet != nil {
		Walk(v, d.SelectionSet)
	}
}

func (d *FragmentDefinition) walkChildren(v Visitor) {
	if d.TypeCondition != nil {
		Walk(v, d.TypeCondition)
	}
	walkDirectives(v, d.Directives)
	if d.SelectionSet != nil {
		Walk(v, d.SelectionSet)
	}
}

func (vdef *VariableDefinition) walkChildren(v Visitor) {
	if vdef.Variable != nil {
		Walk(v, vdef.Variable)
	}
	walkNode(v, vdef.Type)
	walkNode(v, vdef.DefaultValue)
	walkDirectives(v, vdef.Directives)
}

func (s *SelectionSet) walkChildren(v Visitor) {
	walkSelections(v, s.Selections)
}

func (f *Field) walkChildren(v Visitor) {
	walkArguments(v, f.Arguments)
	walkDirectives(v, f.Directives)
	if f.SelectionSet != nil {
		Walk(v, f.SelectionSet)
	}
}

func (s *FragmentSpread) walkChildren(v Visitor) {
	walkDirectives(v, s.Directives)
}

func (f *InlineFragment) walkChildren(v Visitor) {
	if f.TypeCondition != nil {
		Walk(v, f.TypeCondition)
	}
	walkDirectives(v, f.Directives)
	if f.SelectionSet != nil {
		Walk(v, f.SelectionSet)
	}
}

func (a *Argument) walkChildren(v Visitor) {
	walkNode(v, a.Value)
}

func (d *Directive) walkChildren(v Visitor) {
	walkArguments(v, d.Arguments)
}

func (vlist *ListValue) walkChildren(v Visitor) {
	walkValues(v, vlist.Values)
}

func (vobj *ObjectValue) walkChildren(v Visitor) {
	walkObjectFields(v, vobj.Fields)
}

func (f *ObjectField) walkChildren(v Visitor) {
	walkNode(v, f.Value)
}

func (t *ListType) walkChildren(v Visitor) {
	if t.OfType != nil {
		Walk(v, t.OfType)
	}
}

func (t *NonNullType) walkChildren(v Visitor) {
	if t.OfType != nil {
		Walk(v, t.OfType)
	}
}

func (d *SchemaDefinition) walkChildren(v Visitor) {
	if d.Description != nil {
		Walk(v, d.Description)
	}
	walkDirectives(v, d.Directives)
	walkOperationTypes(v, d.OperationTypes)
}

func (d *SchemaExtension) walkChildren(v Visitor) {
	walkDirectives(v, d.Directives)
	walkOperationTypes(v, d.OperationTypes)
}

func (d *OperationTypeDefinition) walkChildren(v Visitor) {
	if d.Type != nil {
		Walk(v, d.Type)
	}
}

func (d *ScalarTypeDefinition) walkChildren(v Visitor) {
	if d.Description != nil {
		Walk(v, d.Description)
	}
	walkDirectives(v, d.Directives)
}

func (d *ScalarTypeExtension) walkChildren(v Visitor) {
	walkDirectives(v, d.Directives)
}

func (d *ObjectTypeDefinition) walkChildren(v Visitor) {
	if d.Description != nil {
		Walk(v, d.Description)
	}
	walkNamedTypes(v, d.Interfaces)
	walkDirectives(v, d.Directives)
	walkFieldDefinitions(v, d.Fields)
}

func (d *ObjectTypeExtension) walkChildren(v Visitor) {
	walkNamedTypes(v, d.Interfaces)
	walkDirectives(v, d.Directives)
	walkFieldDefinitions(v, d.Fields)
}

func (d *InterfaceTypeDefinition) walkChildren(v Visitor) {
	if d.Description != nil {
		Walk(v, d.Description)
	}
	walkNamedTypes(v, d.Interfaces)
	walkDirectives(v, d.Directives)
	walkFieldDefinitions(v, d.Fields)
}

func (d *InterfaceTypeExtension) walkChildren(v Visitor) {
	walkNamedTypes(v, d.Interfaces)
	walkDirectives(v, d.Directives)
	walkFieldDefinitions(v, d.Fields)
}

func (d *UnionTypeDefinition) walkChildren(v Visitor) {
	if d.Description != nil {
		Walk(v, d.Description)
	}
	walkDirectives(v, d.Directives)
	walkNamedTypes(v, d.Members)
}

func (d *UnionTypeExtension) walkChildren(v Visitor) {
	walkDirectives(v, d.Directives)
	walkNamedTypes(v, d.Members)
}

func (d *EnumTypeDefinition) walkChildren(v Visitor) {
	if d.Description != nil {
		Walk(v, d.Description)
	}
	walkDirectives(v, d.Directives)
	walkEnumValues(v, d.Values)
}

func (d *EnumTypeExtension) walkChildren(v Visitor) {
	walkDirectives(v, d.Directives)
	walkEnumValues(v, d.Values)
}

func (d *InputObjectTypeDefinition) walkChildren(v Visitor) {
	if d.Description != nil {
		Walk(v, d.Description)
	}
	walkDirectives(v, d.Directives)
	walkInputValues(v, d.Fields)
}

func (d *InputObjectTypeExtension) walkChildren(v Visitor) {
	walkDirectives(v, d.Directives)
	walkInputValues(v, d.Fields)
}

func (d *FieldDefinition) walkChildren(v Visitor) {
	if d.Description != nil {
		Walk(v, d.Description)
	}
	walkInputValues(v, d.Arguments)
	walkNode(v, d.Type)
	walkDirectives(v, d.Directives)
}

func (d *InputValueDefinition) walkChildren(v Visitor) {
	if d.Description != nil {
		Walk(v, d.Description)
	}
	walkNode(v, d.Type)
	walkNode(v, d.DefaultValue)
	walkDirectives(v, d.Directives)
}

func (d *EnumValueDefinition) walkChildren(v Visitor) {
	if d.Description != nil {
		Walk(v, d.Description)
	}
	walkDirectives(v, d.Directives)
}

func (d *DirectiveDefinition) walkChildren(v Visitor) {
	if d.Description != nil {
		Walk(v, d.Description)
	}
	walkInputValues(v, d.Arguments)
}
