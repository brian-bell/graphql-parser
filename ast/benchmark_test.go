package ast_test

import (
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/parser"
)

const walkExecutable = `
query GetUser($id: ID!, $includeFriends: Boolean = true) @trace {
	user(id: $id) {
		id
		name
		friends @include(if: $includeFriends) {
			nodes { id name }
		}
		...UserFields
		... on Admin { permissions }
	}
}

fragment UserFields on User {
	email
	profile { avatar }
}
`

const walkSchema = `
"""User account."""
type User implements Node @auth {
	"""Stable id."""
	id: ID!
	name(format: NameFormat = LONG): String!
	friends(first: Int, after: String): UserConnection!
}

interface Node { id: ID! }
union SearchResult = User | Organization
enum NameFormat { SHORT LONG }
input UserFilter { q: String active: Boolean = true }
scalar DateTime
directive @auth(role: Role = ADMIN) repeatable on FIELD_DEFINITION | OBJECT
schema { query: Query }
extend type User { createdAt: DateTime }
`

type countVisitor int

func (c *countVisitor) Visit(ast.Node) ast.Visitor {
	*c++
	return c
}

func benchmarkWalk(b *testing.B, node ast.Node) {
	b.ReportAllocs()
	for b.Loop() {
		var count countVisitor
		ast.Walk(&count, node)
		if count == 0 {
			b.Fatal("walk visited no nodes")
		}
	}
}

func BenchmarkWalkExecutable(b *testing.B) {
	doc, err := parser.Parse(walkExecutable)
	if err != nil {
		b.Fatal(err)
	}
	benchmarkWalk(b, doc)
}

func BenchmarkWalkSchema(b *testing.B) {
	doc, err := parser.Parse(walkSchema)
	if err != nil {
		b.Fatal(err)
	}
	benchmarkWalk(b, doc)
}

func BenchmarkWalkValue(b *testing.B) {
	value, err := parser.ParseValue(`{a: [1, "hi", $v, ENUM, true, null], b: {nested: 1}}`)
	if err != nil {
		b.Fatal(err)
	}
	benchmarkWalk(b, value)
}

func BenchmarkWalkType(b *testing.B) {
	typ, err := parser.ParseType("[[Int!]!]!")
	if err != nil {
		b.Fatal(err)
	}
	benchmarkWalk(b, typ)
}
