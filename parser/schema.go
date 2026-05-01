package parser

import (
	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/lexer"
)

// directiveLocations is the set of valid directive-location names per the
// October 2021 spec.
var directiveLocations = map[string]struct{}{
	// Executable
	"QUERY": {}, "MUTATION": {}, "SUBSCRIPTION": {}, "FIELD": {},
	"FRAGMENT_DEFINITION": {}, "FRAGMENT_SPREAD": {}, "INLINE_FRAGMENT": {},
	"VARIABLE_DEFINITION": {},
	// Type system
	"SCHEMA": {}, "SCALAR": {}, "OBJECT": {}, "FIELD_DEFINITION": {},
	"ARGUMENT_DEFINITION": {}, "INTERFACE": {}, "UNION": {}, "ENUM": {},
	"ENUM_VALUE": {}, "INPUT_OBJECT": {}, "INPUT_FIELD_DEFINITION": {},
}

// parseTypeSystemDefinitionOrExtension parses any of the type-system grammar
// productions (schema/scalar/type/interface/union/enum/input/directive
// definitions, plus their "extend" variants). It returns ok=false when the
// next token does not begin one of these productions, in which case the
// caller's dispatch should report an error.
func (p *parser) parseTypeSystemDefinitionOrExtension() (ast.Definition, bool, error) {
	// Optional leading description.
	var desc *ast.StringValue
	tok, err := p.peek()
	if err != nil {
		return nil, false, err
	}
	descStart := tok.Start
	if tok.Kind == lexer.STRING {
		strTok, err := p.advance()
		if err != nil {
			return nil, false, err
		}
		desc = &ast.StringValue{
			Value: strTok.Value,
			Block: strTok.Block,
			Loc:   p.loc(strTok.Start),
		}
		tok, err = p.peek()
		if err != nil {
			return nil, false, err
		}
	}

	if tok.Kind != lexer.NAME {
		if desc != nil {
			return nil, false, p.errAtTok(tok, "Expected type-system definition keyword after description, found "+describeToken(tok)+".")
		}
		return nil, false, nil
	}

	switch tok.Value {
	case "schema":
		return p.parseSchemaDefinition(desc, descStart)
	case "scalar":
		return p.parseScalarTypeDefinition(desc, descStart)
	case "type":
		return p.parseObjectTypeDefinition(desc, descStart)
	case "interface":
		return p.parseInterfaceTypeDefinition(desc, descStart)
	case "union":
		return p.parseUnionTypeDefinition(desc, descStart)
	case "enum":
		return p.parseEnumTypeDefinition(desc, descStart)
	case "input":
		return p.parseInputObjectTypeDefinition(desc, descStart)
	case "directive":
		return p.parseDirectiveDefinition(desc, descStart)
	case "extend":
		if desc != nil {
			return nil, false, p.errAtTok(tok, "Type extensions cannot have a description.")
		}
		return p.parseTypeSystemExtension()
	}

	if desc != nil {
		return nil, false, p.errAtTok(tok, "Unexpected Name "+describeToken(tok)+".")
	}
	return nil, false, nil
}

// definitionStart returns whichever of the two starts is non-zero, preferring
// the description start when present.
func definitionStart(descStart, kwStart int, hasDesc bool) int {
	if hasDesc {
		return descStart
	}
	return kwStart
}

func (p *parser) parseSchemaDefinition(desc *ast.StringValue, descStart int) (ast.Definition, bool, error) {
	kw, err := p.expectKeyword("schema")
	if err != nil {
		return nil, false, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, false, err
	}
	ots, err := p.parseOperationTypeDefinitions()
	if err != nil {
		return nil, false, err
	}
	return &ast.SchemaDefinition{
		Description:    desc,
		Directives:     dirs,
		OperationTypes: ots,
		Loc:            p.loc(definitionStart(descStart, kw.Start, desc != nil)),
	}, true, nil
}

func (p *parser) parseOperationTypeDefinitions() ([]*ast.OperationTypeDefinition, error) {
	if _, err := p.expect(lexer.LBRACE); err != nil {
		return nil, err
	}
	var out []*ast.OperationTypeDefinition
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind == lexer.RBRACE {
			break
		}
		ot, err := p.parseOperationTypeDefinition()
		if err != nil {
			return nil, err
		}
		out = append(out, ot)
	}
	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, p.errAtTok(lexer.Token{Start: p.lastEnd}, "Expected operation type definition.")
	}
	return out, nil
}

func (p *parser) parseOperationTypeDefinition() (*ast.OperationTypeDefinition, error) {
	op, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, err
	}
	switch op.Value {
	case "query", "mutation", "subscription":
	default:
		return nil, p.errAtTok(op, "Expected query, mutation, or subscription, found Name "+op.Value+".")
	}
	if _, err := p.expect(lexer.COLON); err != nil {
		return nil, err
	}
	t, err := p.parseNamedType()
	if err != nil {
		return nil, err
	}
	return &ast.OperationTypeDefinition{
		Operation: ast.OperationType(op.Value),
		Type:      t,
		Loc:       p.loc(op.Start),
	}, nil
}

func (p *parser) parseScalarTypeDefinition(desc *ast.StringValue, descStart int) (ast.Definition, bool, error) {
	kw, err := p.expectKeyword("scalar")
	if err != nil {
		return nil, false, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, false, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, false, err
	}
	return &ast.ScalarTypeDefinition{
		Description: desc,
		Name:        name.Value,
		Directives:  dirs,
		Loc:         p.loc(definitionStart(descStart, kw.Start, desc != nil)),
	}, true, nil
}

func (p *parser) parseObjectTypeDefinition(desc *ast.StringValue, descStart int) (ast.Definition, bool, error) {
	kw, err := p.expectKeyword("type")
	if err != nil {
		return nil, false, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, false, err
	}
	ifaces, err := p.parseImplementsInterfaces()
	if err != nil {
		return nil, false, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, false, err
	}
	fields, err := p.parseFieldsDefinition()
	if err != nil {
		return nil, false, err
	}
	return &ast.ObjectTypeDefinition{
		Description: desc,
		Name:        name.Value,
		Interfaces:  ifaces,
		Directives:  dirs,
		Fields:      fields,
		Loc:         p.loc(definitionStart(descStart, kw.Start, desc != nil)),
	}, true, nil
}

func (p *parser) parseInterfaceTypeDefinition(desc *ast.StringValue, descStart int) (ast.Definition, bool, error) {
	kw, err := p.expectKeyword("interface")
	if err != nil {
		return nil, false, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, false, err
	}
	ifaces, err := p.parseImplementsInterfaces()
	if err != nil {
		return nil, false, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, false, err
	}
	fields, err := p.parseFieldsDefinition()
	if err != nil {
		return nil, false, err
	}
	return &ast.InterfaceTypeDefinition{
		Description: desc,
		Name:        name.Value,
		Interfaces:  ifaces,
		Directives:  dirs,
		Fields:      fields,
		Loc:         p.loc(definitionStart(descStart, kw.Start, desc != nil)),
	}, true, nil
}

func (p *parser) parseImplementsInterfaces() ([]*ast.NamedType, error) {
	ok, err := p.optionalKeyword("implements")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	// Optional leading "&" per spec.
	if _, _, err := p.optional(lexer.AMP); err != nil {
		return nil, err
	}
	var out []*ast.NamedType
	for {
		t, err := p.parseNamedType()
		if err != nil {
			return nil, err
		}
		out = append(out, t)
		_, ok, err := p.optional(lexer.AMP)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
	}
	return out, nil
}

func (p *parser) parseFieldsDefinition() (ast.FieldDefinitionList, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if tok.Kind != lexer.LBRACE {
		return nil, nil
	}
	if _, err := p.advance(); err != nil {
		return nil, err
	}
	var out ast.FieldDefinitionList
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind == lexer.RBRACE {
			break
		}
		f, err := p.parseFieldDefinition()
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, p.errAtTok(lexer.Token{Start: p.lastEnd}, "Expected field definition.")
	}
	return out, nil
}

func (p *parser) parseFieldDefinition() (*ast.FieldDefinition, error) {
	leading := p.pendingLeading
	p.pendingLeading = nil
	desc, descStart, err := p.parseOptionalDescription()
	if err != nil {
		return nil, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, err
	}
	args, err := p.parseArgumentsDefinition()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.COLON); err != nil {
		return nil, err
	}
	t, err := p.parseTypeReference()
	if err != nil {
		return nil, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, err
	}
	start := descStart
	if desc == nil {
		start = name.Start
	}
	fd := &ast.FieldDefinition{
		Description: desc,
		Name:        name.Value,
		Arguments:   args,
		Type:        t,
		Directives:  dirs,
		Loc:         p.loc(start),
	}
	if len(leading) > 0 {
		fd.Comments = &ast.CommentGroup{Leading: leading}
	}
	return fd, nil
}

func (p *parser) parseArgumentsDefinition() (ast.InputValueList, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if tok.Kind != lexer.LPAREN {
		return nil, nil
	}
	if _, err := p.advance(); err != nil {
		return nil, err
	}
	var out ast.InputValueList
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind == lexer.RPAREN {
			break
		}
		v, err := p.parseInputValueDefinition()
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if _, err := p.expect(lexer.RPAREN); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, p.errAtTok(lexer.Token{Start: p.lastEnd}, "Expected argument definition.")
	}
	return out, nil
}

func (p *parser) parseInputValueDefinition() (*ast.InputValueDefinition, error) {
	leading := p.pendingLeading
	p.pendingLeading = nil
	desc, descStart, err := p.parseOptionalDescription()
	if err != nil {
		return nil, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.COLON); err != nil {
		return nil, err
	}
	t, err := p.parseTypeReference()
	if err != nil {
		return nil, err
	}
	var def ast.Value
	if _, ok, err := p.optional(lexer.EQUALS); err != nil {
		return nil, err
	} else if ok {
		v, err := p.parseValueLiteral(true)
		if err != nil {
			return nil, err
		}
		def = v
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, err
	}
	start := descStart
	if desc == nil {
		start = name.Start
	}
	iv := &ast.InputValueDefinition{
		Description:  desc,
		Name:         name.Value,
		Type:         t,
		DefaultValue: def,
		Directives:   dirs,
		Loc:          p.loc(start),
	}
	if len(leading) > 0 {
		iv.Comments = &ast.CommentGroup{Leading: leading}
	}
	return iv, nil
}

func (p *parser) parseUnionTypeDefinition(desc *ast.StringValue, descStart int) (ast.Definition, bool, error) {
	kw, err := p.expectKeyword("union")
	if err != nil {
		return nil, false, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, false, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, false, err
	}
	members, err := p.parseUnionMemberTypes()
	if err != nil {
		return nil, false, err
	}
	return &ast.UnionTypeDefinition{
		Description: desc,
		Name:        name.Value,
		Directives:  dirs,
		Members:     members,
		Loc:         p.loc(definitionStart(descStart, kw.Start, desc != nil)),
	}, true, nil
}

func (p *parser) parseUnionMemberTypes() ([]*ast.NamedType, error) {
	if _, ok, err := p.optional(lexer.EQUALS); err != nil {
		return nil, err
	} else if !ok {
		return nil, nil
	}
	if _, _, err := p.optional(lexer.PIPE); err != nil {
		return nil, err
	}
	var out []*ast.NamedType
	for {
		t, err := p.parseNamedType()
		if err != nil {
			return nil, err
		}
		out = append(out, t)
		_, ok, err := p.optional(lexer.PIPE)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
	}
	return out, nil
}

func (p *parser) parseEnumTypeDefinition(desc *ast.StringValue, descStart int) (ast.Definition, bool, error) {
	kw, err := p.expectKeyword("enum")
	if err != nil {
		return nil, false, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, false, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, false, err
	}
	values, err := p.parseEnumValuesDefinition()
	if err != nil {
		return nil, false, err
	}
	return &ast.EnumTypeDefinition{
		Description: desc,
		Name:        name.Value,
		Directives:  dirs,
		Values:      values,
		Loc:         p.loc(definitionStart(descStart, kw.Start, desc != nil)),
	}, true, nil
}

func (p *parser) parseEnumValuesDefinition() (ast.EnumValueList, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if tok.Kind != lexer.LBRACE {
		return nil, nil
	}
	if _, err := p.advance(); err != nil {
		return nil, err
	}
	var out ast.EnumValueList
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind == lexer.RBRACE {
			break
		}
		v, err := p.parseEnumValueDefinition()
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, p.errAtTok(lexer.Token{Start: p.lastEnd}, "Expected enum value definition.")
	}
	return out, nil
}

func (p *parser) parseEnumValueDefinition() (*ast.EnumValueDefinition, error) {
	leading := p.pendingLeading
	p.pendingLeading = nil
	desc, descStart, err := p.parseOptionalDescription()
	if err != nil {
		return nil, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, err
	}
	switch name.Value {
	case "true", "false", "null":
		return nil, p.errAtTok(name, "Enum value cannot be true, false, or null.")
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, err
	}
	start := descStart
	if desc == nil {
		start = name.Start
	}
	ev := &ast.EnumValueDefinition{
		Description: desc,
		Name:        name.Value,
		Directives:  dirs,
		Loc:         p.loc(start),
	}
	if len(leading) > 0 {
		ev.Comments = &ast.CommentGroup{Leading: leading}
	}
	return ev, nil
}

func (p *parser) parseInputObjectTypeDefinition(desc *ast.StringValue, descStart int) (ast.Definition, bool, error) {
	kw, err := p.expectKeyword("input")
	if err != nil {
		return nil, false, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, false, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, false, err
	}
	fields, err := p.parseInputFieldsDefinition()
	if err != nil {
		return nil, false, err
	}
	return &ast.InputObjectTypeDefinition{
		Description: desc,
		Name:        name.Value,
		Directives:  dirs,
		Fields:      fields,
		Loc:         p.loc(definitionStart(descStart, kw.Start, desc != nil)),
	}, true, nil
}

func (p *parser) parseInputFieldsDefinition() (ast.InputValueList, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if tok.Kind != lexer.LBRACE {
		return nil, nil
	}
	if _, err := p.advance(); err != nil {
		return nil, err
	}
	var out ast.InputValueList
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.Kind == lexer.RBRACE {
			break
		}
		v, err := p.parseInputValueDefinition()
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, p.errAtTok(lexer.Token{Start: p.lastEnd}, "Expected input field definition.")
	}
	return out, nil
}

func (p *parser) parseDirectiveDefinition(desc *ast.StringValue, descStart int) (ast.Definition, bool, error) {
	kw, err := p.expectKeyword("directive")
	if err != nil {
		return nil, false, err
	}
	if _, err := p.expect(lexer.AT); err != nil {
		return nil, false, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, false, err
	}
	args, err := p.parseArgumentsDefinition()
	if err != nil {
		return nil, false, err
	}
	repeatable, err := p.optionalKeyword("repeatable")
	if err != nil {
		return nil, false, err
	}
	if _, err := p.expectKeyword("on"); err != nil {
		return nil, false, err
	}
	locs, err := p.parseDirectiveLocations()
	if err != nil {
		return nil, false, err
	}
	return &ast.DirectiveDefinition{
		Description: desc,
		Name:        name.Value,
		Arguments:   args,
		Repeatable:  repeatable,
		Locations:   locs,
		Loc:         p.loc(definitionStart(descStart, kw.Start, desc != nil)),
	}, true, nil
}

func (p *parser) parseDirectiveLocations() ([]string, error) {
	if _, _, err := p.optional(lexer.PIPE); err != nil {
		return nil, err
	}
	var out []string
	for {
		name, err := p.expect(lexer.NAME)
		if err != nil {
			return nil, err
		}
		if _, ok := directiveLocations[name.Value]; !ok {
			return nil, p.errAtTok(name, "Unexpected directive location "+name.Value+".")
		}
		out = append(out, name.Value)
		_, ok, err := p.optional(lexer.PIPE)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
	}
	return out, nil
}

// parseOptionalDescription consumes a leading StringValue (used as a
// description) if present. Returns (nil, 0, nil) when the next token is not
// a STRING.
func (p *parser) parseOptionalDescription() (*ast.StringValue, int, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, 0, err
	}
	if tok.Kind != lexer.STRING {
		return nil, 0, nil
	}
	str, err := p.advance()
	if err != nil {
		return nil, 0, err
	}
	return &ast.StringValue{
		Value: str.Value,
		Block: str.Block,
		Loc:   p.loc(str.Start),
	}, str.Start, nil
}

// parseTypeSystemExtension dispatches "extend ..." to the appropriate
// extension-parser. The "extend" keyword has already been peeked.
func (p *parser) parseTypeSystemExtension() (ast.Definition, bool, error) {
	kw, err := p.expectKeyword("extend")
	if err != nil {
		return nil, false, err
	}
	tok, err := p.peek()
	if err != nil {
		return nil, false, err
	}
	if tok.Kind != lexer.NAME {
		return nil, false, p.errAtTok(tok, "Expected type-system extension keyword, found "+describeToken(tok)+".")
	}
	switch tok.Value {
	case "schema":
		return p.parseSchemaExtension(kw.Start)
	case "scalar":
		return p.parseScalarTypeExtension(kw.Start)
	case "type":
		return p.parseObjectTypeExtension(kw.Start)
	case "interface":
		return p.parseInterfaceTypeExtension(kw.Start)
	case "union":
		return p.parseUnionTypeExtension(kw.Start)
	case "enum":
		return p.parseEnumTypeExtension(kw.Start)
	case "input":
		return p.parseInputObjectTypeExtension(kw.Start)
	}
	return nil, false, p.errAtTok(tok, "Unexpected extension target Name "+tok.Value+".")
}

func (p *parser) parseSchemaExtension(start int) (ast.Definition, bool, error) {
	if _, err := p.expectKeyword("schema"); err != nil {
		return nil, false, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, false, err
	}
	var ots []*ast.OperationTypeDefinition
	if tok, err := p.peek(); err != nil {
		return nil, false, err
	} else if tok.Kind == lexer.LBRACE {
		ots, err = p.parseOperationTypeDefinitions()
		if err != nil {
			return nil, false, err
		}
	}
	if len(dirs) == 0 && len(ots) == 0 {
		return nil, false, p.errAtTok(lexer.Token{Start: p.lastEnd}, "Schema extension must define at least one directive or operation type.")
	}
	return &ast.SchemaExtension{
		Directives:     dirs,
		OperationTypes: ots,
		Loc:            p.loc(start),
	}, true, nil
}

func (p *parser) parseScalarTypeExtension(start int) (ast.Definition, bool, error) {
	if _, err := p.expectKeyword("scalar"); err != nil {
		return nil, false, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, false, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, false, err
	}
	if len(dirs) == 0 {
		return nil, false, p.errAtTok(lexer.Token{Start: p.lastEnd}, "Scalar extension must define at least one directive.")
	}
	return &ast.ScalarTypeExtension{
		Name:       name.Value,
		Directives: dirs,
		Loc:        p.loc(start),
	}, true, nil
}

func (p *parser) parseObjectTypeExtension(start int) (ast.Definition, bool, error) {
	if _, err := p.expectKeyword("type"); err != nil {
		return nil, false, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, false, err
	}
	ifaces, err := p.parseImplementsInterfaces()
	if err != nil {
		return nil, false, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, false, err
	}
	fields, err := p.parseFieldsDefinition()
	if err != nil {
		return nil, false, err
	}
	if len(ifaces) == 0 && len(dirs) == 0 && len(fields) == 0 {
		return nil, false, p.errAtTok(lexer.Token{Start: p.lastEnd}, "Object extension must define at least one of: implements interfaces, directives, or fields.")
	}
	return &ast.ObjectTypeExtension{
		Name:       name.Value,
		Interfaces: ifaces,
		Directives: dirs,
		Fields:     fields,
		Loc:        p.loc(start),
	}, true, nil
}

func (p *parser) parseInterfaceTypeExtension(start int) (ast.Definition, bool, error) {
	if _, err := p.expectKeyword("interface"); err != nil {
		return nil, false, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, false, err
	}
	ifaces, err := p.parseImplementsInterfaces()
	if err != nil {
		return nil, false, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, false, err
	}
	fields, err := p.parseFieldsDefinition()
	if err != nil {
		return nil, false, err
	}
	if len(ifaces) == 0 && len(dirs) == 0 && len(fields) == 0 {
		return nil, false, p.errAtTok(lexer.Token{Start: p.lastEnd}, "Interface extension must define at least one of: implements interfaces, directives, or fields.")
	}
	return &ast.InterfaceTypeExtension{
		Name:       name.Value,
		Interfaces: ifaces,
		Directives: dirs,
		Fields:     fields,
		Loc:        p.loc(start),
	}, true, nil
}

func (p *parser) parseUnionTypeExtension(start int) (ast.Definition, bool, error) {
	if _, err := p.expectKeyword("union"); err != nil {
		return nil, false, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, false, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, false, err
	}
	members, err := p.parseUnionMemberTypes()
	if err != nil {
		return nil, false, err
	}
	if len(dirs) == 0 && len(members) == 0 {
		return nil, false, p.errAtTok(lexer.Token{Start: p.lastEnd}, "Union extension must define at least one directive or member type.")
	}
	return &ast.UnionTypeExtension{
		Name:       name.Value,
		Directives: dirs,
		Members:    members,
		Loc:        p.loc(start),
	}, true, nil
}

func (p *parser) parseEnumTypeExtension(start int) (ast.Definition, bool, error) {
	if _, err := p.expectKeyword("enum"); err != nil {
		return nil, false, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, false, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, false, err
	}
	values, err := p.parseEnumValuesDefinition()
	if err != nil {
		return nil, false, err
	}
	if len(dirs) == 0 && len(values) == 0 {
		return nil, false, p.errAtTok(lexer.Token{Start: p.lastEnd}, "Enum extension must define at least one directive or value.")
	}
	return &ast.EnumTypeExtension{
		Name:       name.Value,
		Directives: dirs,
		Values:     values,
		Loc:        p.loc(start),
	}, true, nil
}

func (p *parser) parseInputObjectTypeExtension(start int) (ast.Definition, bool, error) {
	if _, err := p.expectKeyword("input"); err != nil {
		return nil, false, err
	}
	name, err := p.expect(lexer.NAME)
	if err != nil {
		return nil, false, err
	}
	dirs, err := p.parseDirectives(true)
	if err != nil {
		return nil, false, err
	}
	fields, err := p.parseInputFieldsDefinition()
	if err != nil {
		return nil, false, err
	}
	if len(dirs) == 0 && len(fields) == 0 {
		return nil, false, p.errAtTok(lexer.Token{Start: p.lastEnd}, "Input object extension must define at least one directive or field.")
	}
	return &ast.InputObjectTypeExtension{
		Name:       name.Value,
		Directives: dirs,
		Fields:     fields,
		Loc:        p.loc(start),
	}, true, nil
}
