package parser_test

import (
	"strings"
	"testing"

	"github.com/brian-bell/graphql-parser/parser"
)

const tinyQuery = `query GetUser($id: ID!) { user(id: $id) { id name email } }`

const midSchema = `
"""User account."""
type User implements Node {
	id: ID!
	name(format: NameFormat = LONG): String!
	email: String
	friends(first: Int, after: String): UserConnection!
	posts(filter: PostFilter): [Post!]!
}

type Post {
	id: ID!
	title: String!
	body: String!
	author: User!
	tags: [String!]!
	createdAt: DateTime!
}

type UserConnection {
	edges: [UserEdge!]!
	pageInfo: PageInfo!
}

type UserEdge {
	cursor: String!
	node: User!
}

type PageInfo {
	hasNextPage: Boolean!
	endCursor: String
}

input PostFilter {
	tag: String
	author: ID
}

enum NameFormat { SHORT LONG }

interface Node {
	id: ID!
}

union SearchResult = User | Post

scalar DateTime

directive @auth(role: Role!) on FIELD_DEFINITION | OBJECT

type Query {
	user(id: ID!): User
	posts(filter: PostFilter): [Post!]!
	search(q: String!): [SearchResult!]!
}

schema { query: Query }
`

func largeSchema() string {
	// Synthesize a large schema by repeating the mid schema with renamed types.
	var sb strings.Builder
	for i := range 50 {
		s := strings.ReplaceAll(midSchema, "User", "U"+itoa(i))
		s = strings.ReplaceAll(s, "Post", "P"+itoa(i))
		sb.WriteString(s)
		sb.WriteString("\n")
	}
	return sb.String()
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [10]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}

func BenchmarkParseTinyQuery(b *testing.B) {
	body := tinyQuery
	b.ReportAllocs()
	b.SetBytes(int64(len(body)))
	for b.Loop() {
		_, err := parser.Parse(body)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseMidSchema(b *testing.B) {
	body := midSchema
	b.ReportAllocs()
	b.SetBytes(int64(len(body)))
	for b.Loop() {
		_, err := parser.Parse(body)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseLargeSchema(b *testing.B) {
	body := largeSchema()
	b.ReportAllocs()
	b.SetBytes(int64(len(body)))
	for b.Loop() {
		_, err := parser.Parse(body)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseTinyQuery_WithComments(b *testing.B) {
	body := "# leading\n" + tinyQuery
	b.ReportAllocs()
	b.SetBytes(int64(len(body)))
	for b.Loop() {
		_, err := parser.Parse(body, parser.WithComments())
		if err != nil {
			b.Fatal(err)
		}
	}
}
