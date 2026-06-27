# graphql-parser

A small, dependency-free Go library that parses GraphQL source text into an
AST. Supports both grammars from the
[October 2021 GraphQL spec](https://spec.graphql.org/October2021/) — executable
documents and the type-system / SDL — in three small packages:

```
github.com/brian-bell/graphql-parser/ast    // node types, Source, Position, Loc, Walk
github.com/brian-bell/graphql-parser/lexer  // tokenizer
github.com/brian-bell/graphql-parser/parser // Parse, ParseSource, ParseSchema, ParseSchemaSource, ParseValue, ParseConstValue, ParseType
```

Validation, execution, and printing are out of scope.

## Install

```sh
go get github.com/brian-bell/graphql-parser
```

Requires Go 1.26+.

## Usage

### Parse a document

```go
import "github.com/brian-bell/graphql-parser/parser"

doc, err := parser.Parse(`
    query GetUser($id: ID!) {
        user(id: $id) { id name email }
    }
`)
if err != nil { panic(err) }

for _, def := range doc.Definitions {
    // type-switch on def: *ast.OperationDefinition, *ast.FragmentDefinition,
    // *ast.ObjectTypeDefinition, etc.
}
```

### Parse SDL only

Use `ParseSchema` when callers expect a schema / SDL document and want to
reject executable definitions (operations and fragments) after syntax parsing.
It accepts the same options as `Parse`; use `ParseSchemaSource` to provide a
custom `*ast.Source` name or `LocationOffset`.

```go
schemaDoc, err := parser.ParseSchema(`
    type Query {
        user(id: ID!): User
    }

    extend type Query {
        viewer: User
    }
`)
```

### Parse a single value or type literal

```go
v, err := parser.ParseValue(`{ id: 1, tags: ["a", "b"] }`)
t, err := parser.ParseType(`[String!]!`)
```

`ParseConstValue` is the const-context variant; it rejects `$variables` at
any nesting depth. Use it for default-value parsing.

### Position-rich errors

Errors include a graphql-js-style source-line snippet with a caret pointer:

```
Syntax Error: Expected Name, found <EOF>.

example.graphql:3:1
2 |   field(
3 | }
  | ^
```

Reported line/column numbers honor `Source.LocationOffset`, so errors in a
fragment parsed from a larger file (e.g. an embedded `gql` template tag)
report the original file's coordinates.

### Walking the AST

`ast.Walk` and `ast.Inspect` traverse a node depth-first in source order:

```go
import "github.com/brian-bell/graphql-parser/ast"

ast.Inspect(doc, func(n ast.Node) bool {
    if f, ok := n.(*ast.Field); ok {
        fmt.Println(f.Name)
    }
    return true // continue descending
})
```

## Options

Every `Parse*` entry point accepts variadic options.

### `WithRecovery`

Collect multiple syntax errors instead of aborting on the first.

```go
doc, err := parser.Parse(brokenSource, parser.WithRecovery())
if err != nil {
    var es parser.ParseErrors
    if errors.As(err, &es) {
        for _, pe := range es {
            fmt.Println(pe)
        }
    }
}
```

The returned document is still populated with whatever could be parsed; the
parser inserts `BadDefinition` and `BadField` placeholder nodes where it
had to resynchronize. This is what an LSP wants when the user is mid-typing.

### `WithComments`

Attach `# ...` line comments as `Leading` trivia on AST nodes (top-level
Definitions, FieldDefinitions, EnumValueDefinitions, InputValueDefinitions).

```go
doc, _ := parser.Parse(source, parser.WithComments())
def := doc.Definitions[0].(*ast.ObjectTypeDefinition)
for _, c := range def.Comments.Leading {
    fmt.Println(c.Text)
}
```

When this option is not set, the parser silently skips comments and the
AST is byte-identical to the no-comments configuration.

## Why this library?

There are existing Go GraphQL parsers ([gqlparser](https://github.com/vektah/gqlparser),
[graphql-go/graphql](https://github.com/graphql-go/graphql/tree/master/language)),
and they're fine for most use cases. This library targets the gap they leave
for tooling consumers:

- **Full-extent positions.** Every `*ast.Loc` carries `{Start, End}` byte
  offsets covering the entire node, not just its first token. Formatters,
  LSPs, linters, and renamers all want this.
- **Idiomatic Go AST.** Interfaces at spec union points (`Definition`,
  `Selection`, `Value`, `Type`); concrete pointer-receiver structs
  everywhere else. No `kind` discriminator, no nil-load-bearing fields.
- **Recovery + comments built-in.** Both are opt-in (off by default to
  preserve fail-fast and corpus conformance) but there's no extra package
  to import or library to wrap.

Conformance is held to a simple bar: in default fail-fast mode, the
parser passes the ported `graphql-js` parser test corpus byte-for-byte on
error messages.

## License

MIT — see [LICENSE](./LICENSE). The ported test corpus under
`parser/testdata/graphql-js/` carries the upstream graphql-js MIT license;
see [THIRD_PARTY_LICENSES.md](./THIRD_PARTY_LICENSES.md).
