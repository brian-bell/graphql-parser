# AGENTS.md

This file provides guidance to AI coding agents working in this repository.

## Commands

```sh
go test ./...                              # run all tests
go test -race ./...                        # race-checked (default for CI)
go test ./parser/ -run TestParse_Foo       # one test
go test -v ./parser/ -run TestCorpus       # full subtest output for one suite
go test -fuzz=FuzzParse$ -fuzztime=30s -run=- ./parser/   # fuzz one target (regex must anchor)
go test -bench=. -benchmem -run=- ./parser/                # benchmarks
gofmt -l .                                 # must be empty before commit
go vet ./...
```

Module path: `github.com/brian-bell/graphql-parser`. Go 1.26 floor.

## Architecture

Four packages with a strict dependency DAG:

```text
lexer -> ast
parser -> ast, lexer
schemaindex -> ast
```

`ast/` is intentionally importable on its own. A future printer, formatter, or
LSP can depend on AST types without dragging in the parser.

### Package responsibilities

- `ast/` - `Node`, `Definition`, `Selection`, `Value`, `Type` interfaces; concrete pointer-receiver structs for every spec production; `Source`, `Position`, `Loc` with a lazy line-start index; `BlockStringValue`, `TypeString`, `NamedTypeName`, and `DirectiveStringArg` helpers; `Walk`/`Inspect`; `SyntaxError` with the graphql-js-style snippet renderer.
- `lexer/` - synchronous, single-token-lookahead, hand-written tokenizer. `Token` carries byte offsets plus a `Value` that is a sub-slice of source for `NAME`/`INT`/`FLOAT`. `STRING` tokens own their decoded value. `Lexer.PreserveComments` is the internal switch the parser flips for `WithComments`.
- `parser/` - recursive descent. Public API is `Parse`, `ParseSource`, `ParseSchema`, `ParseSchemaSource`, `ParseValue`, `ParseConstValue`, `ParseType`, plus `WithRecovery()` and `WithComments()`. Everything else is unexported.
- `schemaindex/` - small public SDL index over a parsed `*ast.Document`. `New` records the six named type definitions and six matching extension forms by name, keeps base definitions and extensions separate in source order, ignores non-type definitions, and does not validate schema semantics. Object, interface, and input helper accessors expose base members followed by matching extension members without deduplicating or merging raw definitions.

### Key invariants

- Every `*ast.Loc` covers the full extent of its node (`{Start, End}` byte offsets, half-open), not just the first token. When adding a grammar production, build the `Loc` with `p.loc(start)` where `start` is the first token's `Start` and `p.lastEnd` is the end.
- `ast.SyntaxError` is the single error type produced by both the lexer and the parser. `parser.ParseError = ast.SyntaxError` is a type alias. The rich snippet and caret rendering lives in `ast/error.go`.
- `Source.LocationOffset` shifts reported line/column for fragments parsed from a larger file. Zero value means no offset, treated as `{1,1}`. The first-line column offset applies only on line 1; subsequent lines reset.
- `Source` must not be copied after first use because it lazily builds a line-start index under `sync.Once`. Always pass `*Source`.
- Default mode is fail-fast and corpus-conformant. `Bad*` nodes never appear unless `WithRecovery()` is set. Comments fields are nil unless `WithComments()` is set. Both options must be byte-additive; `comments_test.go` includes a comments-off invariance test.
- `ParseSchema` and `ParseSchemaSource` reuse the normal document parser, then reject top-level `*ast.OperationDefinition` and `*ast.FragmentDefinition`. Do not fork the grammar for schema-only parsing.

### Recovery (`WithRecovery`)

Two synchronization points:

- Top-level Definition: on error, `parser/recover.go:skipToDefinitionStart` advances until the next definition keyword, leading description, or `{` shorthand; a `BadDefinition` is appended.
- Selection inside a SelectionSet: `skipToSelectionStart` advances until `NAME` / `...` / `}` / EOF; a `BadField` is appended.

Value- and type-level recovery (`BadValue`, `BadType`) are wired through the AST
and `ParseErrors` plumbing but not yet auto-injected. Adding them means wrapping
the relevant parse function with the same error-record-and-resync pattern.

Schema-only policy errors are appended to the same `ParseErrors` aggregate in
`WithRecovery` mode.

### Comments (`WithComments`)

`parser.peek` and `parser.advance` drain `COMMENT` tokens into
`parser.pendingLeading`. A definition's leading comments must be captured
before `parseDefinition` is called, otherwise the first inner
`FieldDefinition` or similar node can take them. See `parseDocument` for the
pattern:

```go
defLeading := p.pendingLeading
p.pendingLeading = nil
def, err := p.parseDefinition()
...
attachLeadingComments(def, defLeading)
```

`commentGroupOf(node)` is a type-switch returning `**ast.CommentGroup` for
nodes that have a `Comments` field. Extend this when adding new node types that
should carry trivia.

### Type-system parser dispatch

`parser.parseTypeSystemDefinitionOrExtension` returns `(ast.Definition, ok bool,
error)`. The `ok` boolean lets `parseDefinition` distinguish "this token is not
a type-system production, fall through to executable parsing" from a real
syntax error. When adding a new top-level keyword, branch in this function and
return `(node, true, nil)`.

### Performance discipline

`docs/perf.md` is load-bearing: no `regexp`, no `reflect`, no `fmt.Sprintf` on
the hot path. Tokens carry byte offsets, not copied strings, for
`NAME`/`INT`/`FLOAT`. `Walk` is type-switch dispatch, not reflection.
`regexp`/`reflect`/`Sprintf` greps are part of the verification checklist.

## Testing layout

- `*_test.go` next to each file - unit tests
- `parser/corpus_test.go` - graphql-js parser/lexer/schema-parser conformance, table-driven
- `parser/schema_api_test.go` - schema-only `ParseSchema*` public API behavior
- `parser/recover_test.go` / `parser/comments_test.go` - option-specific behavior
- `schemaindex/index_test.go` / `schemaindex/example_test.go` - public schema index behavior and usage
- `parser/fuzz_test.go` - `FuzzParse`, `FuzzParseWithRecovery`, `FuzzParseValue`, `FuzzParseType` with shared seed corpus
- `parser/benchmark_test.go` - tiny query, mid-size schema, large schema benchmarks

## Workflow

- Use TDD for coding changes unless there is a good reason not to.
- Branch off `main`; never commit or push directly to `main`.
- Keep changes scoped. Do not ship unrelated work on an existing PR.
- Before committing, run the relevant focused tests, `go test ./...`, `go vet ./...`, and `gofmt -l .`.
- For API-shape changes, check the design rationale in `~/.claude/plans/on-creating-a-graphql-witty-key.md` when it is available.
