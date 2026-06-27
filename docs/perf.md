# Performance discipline

This library does not advertise a performance bar in its README. Benchmarks
exist as a regression detector, not as a feature gate. Large regressions
flagged by CI are surfaced on PRs but do not block merges. At v1.0, the
README publishes absolute benchmark numbers (with machine spec); we do not
publish comparisons to other libraries.

What we *do* commit to is the implementation discipline below. These are
the rules that keep the parser idiomatic, allocation-aware, and fast
enough for the use cases the library targets without making perf a
load-bearing feature of the project.

## Rules

1. **No `regexp`** in `lexer/`, `parser/`, or `schemaindex/`.
   The grammar and SDL index are small enough to hand-write. Regexp's setup
   cost on a per-parse basis swamps the parsing itself for small inputs.

2. **No `reflect`** in `lexer/`, `parser/`, `schemaindex/`, or `ast/walk.go`.
   The AST has ~40 concrete node types. Type switches are cheaper, easier
   to read, and produce useful compile-time errors when a node kind is
   missed.

3. **No `fmt.Sprintf` on the hot path.**
   Acceptable in error formatting (`ast/error.go` snippet rendering). Not
   acceptable in `lexer/lexer.go`, `parser/parser.go`, or any function
   called once per token / per node on a successful parse.

4. **Tokens carry byte offsets**, not copied substrings.
   `Token.Value` for `NAME`, `INT`, `FLOAT` is a sub-slice of
   `Source.Body` (zero allocation). `STRING` and `COMMENT` tokens own
   their decoded values because escape-decoding requires allocation
   anyway.

5. **`append`-grown lists**. No preallocation tricks unless a benchmark
   proves they matter on a representative input. The Go runtime's slice
   growth policy is good enough at the sizes GraphQL documents reach.

6. **Memory linear in input size**. No hidden quadratic loops; no
   unbounded-buffer accumulation.

## Verification

- `gofmt -l .` is empty.
- `go vet ./...` is clean.
- `grep -rn 'regexp\|reflect\|fmt.Sprintf' lexer/ parser/ ast/ schemaindex/`
  returns only allowed sites: test files, `ast/error.go`, `ast/loc.go` (column
  formatting in `Position.String`), lexer/parser error-message paths, and
  `parser/parser.go`'s `describeToken`.
- Benchmarks in `parser/benchmark_test.go` run cleanly on every PR with
  `go test -bench=. -benchmem ./parser/`.

## Reference numbers

Captured at v1.0 on a known machine. Until then, treat
`go test -bench=. ./parser/` as the source of truth on your hardware.
