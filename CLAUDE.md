# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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

Three packages with a strict dependency DAG:

```
lexer  →  ast
parser →  ast, lexer
```

`ast/` is intentionally importable on its own — a future printer / formatter / LSP can depend on AST types without dragging in the parser.

### Package responsibilities

- `ast/` — `Node`, `Definition`, `Selection`, `Value`, `Type` interfaces; concrete pointer-receiver structs for every spec production; `Source`, `Position`, `Loc` (with lazy line-start index); `BlockStringValue` helper; `Walk`/`Inspect`; `SyntaxError` with the graphql-js-style snippet renderer.
- `lexer/` — synchronous, single-token-lookahead, hand-written. `Token` carries byte offsets + a `Value` that's a sub-slice of source for `NAME`/`INT`/`FLOAT`. `STRING` tokens own their decoded value (escape-decoded and block-string-dedented at lex time, not parse time). `Lexer.PreserveComments` is the internal switch the parser flips for `WithComments`.
- `parser/` — recursive descent. Public API is `Parse`, `ParseSource`, `ParseValue`, `ParseConstValue`, `ParseType`, plus `WithRecovery()` and `WithComments()`. Everything else (the `parser` struct, grammar productions, error helpers) is unexported.

### Key invariants

- **Every `*ast.Loc` covers the FULL EXTENT of its node** (`{Start, End}` byte offsets, half-open). Not just the first token. This diverges from `gqlparser` deliberately and is what makes formatters / LSPs / range-format possible. When adding a grammar production, build the `Loc` with `p.loc(start)` where `start` is the first token's `Start` and `p.lastEnd` (updated by `advance()`) is the End.
- **`ast.SyntaxError` is the single error type** produced by both the lexer and the parser. `parser.ParseError = ast.SyntaxError` is a type alias. The rich snippet+caret rendering lives in `ast/error.go`.
- **`Source.LocationOffset`** shifts reported line/column for fragments parsed from a larger file (e.g. `gql` template tags). Zero value means "no offset" (treated as `{1,1}`). The first-line column offset only applies on line 1; subsequent lines reset.
- **`Source` must not be copied after first use** — it lazily builds a line-start index under `sync.Once`. Always pass `*Source`.
- **Default mode is fail-fast and corpus-conformant.** `Bad*` nodes never appear unless `WithRecovery()` is set. Comments fields are nil unless `WithComments()` is set. Both options must be byte-additive; in particular, `comments_test.go` includes a "comments-off invariance" test asserting the AST is identical with `WithComments` off.

### Recovery (`WithRecovery`)

Two synchronization points:
- **Top-level Definition**: on error, `parser/recover.go:skipToDefinitionStart` advances until the next definition-keyword, leading description, or `{` shorthand; a `BadDefinition` is appended.
- **Selection inside a SelectionSet**: `skipToSelectionStart` advances until `NAME` / `...` / `}` / EOF; a `BadField` is appended.

Value- and type-level recovery (`BadValue`, `BadType`) are wired through the AST and `ParseErrors` plumbing but not yet auto-injected. Adding them = wrap the relevant parse function with the same error-record-and-resync pattern.

### Comments (`WithComments`)

Tricky part: `parser.peek` and `parser.advance` drain `COMMENT` tokens into `parser.pendingLeading`. **A definition's leading comments must be captured BEFORE `parseDefinition` is called**, otherwise the first inner `FieldDefinition` (or similar) will steal them. See `parseDocument` for the pattern:

```go
defLeading := p.pendingLeading
p.pendingLeading = nil
def, err := p.parseDefinition()
...
attachLeadingComments(def, defLeading)
```

`commentGroupOf(node)` is a type-switch returning `**ast.CommentGroup` for nodes that have a `Comments` field — extend this when adding new node types that should carry trivia.

### Type-system parser dispatch

`parser.parseTypeSystemDefinitionOrExtension` returns `(ast.Definition, ok bool, error)`. The `ok` boolean lets `parseDefinition` distinguish "this token isn't a type-system production, fall through to the executable path" from a real syntax error. When adding a new top-level keyword, branch in this function and return `(node, true, nil)`.

### Performance discipline

`docs/perf.md` is load-bearing: no `regexp`, no `reflect`, no `fmt.Sprintf` on the hot path (only in `ast/error.go` snippet rendering and `parser/parser.go:describeToken`). Tokens carry byte offsets, not copied strings, for `NAME`/`INT`/`FLOAT`. `Walk` is type-switch dispatch, not reflection. `regexp`/`reflect`/`Sprintf` greps are part of the verification checklist.

## Testing layout

- `*_test.go` next to each file — unit tests
- `parser/corpus_test.go` — graphql-js parser/lexer/schema-parser conformance, table-driven (extend by adding rows)
- `parser/recover_test.go` / `parser/comments_test.go` — option-specific behavior
- `parser/fuzz_test.go` — `FuzzParse`, `FuzzParseWithRecovery`, `FuzzParseValue`, `FuzzParseType` with shared seed corpus
- `parser/benchmark_test.go` — tiny query, mid-size schema, large schema (50× synthesized)

## Workflow

- TDD per global instructions: failing test first, then implementation.
- Branch off main; never commit or push directly to main. Current development branch is `feat/ast-loc` (which became the foundation branch despite the name; keep using it or branch from it).
- Plan file with full design rationale: `~/.claude/plans/on-creating-a-graphql-witty-key.md` (17 locked decisions; consult before making API-shape changes).
