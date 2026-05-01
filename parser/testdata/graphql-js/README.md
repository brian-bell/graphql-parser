# graphql-js parser-conformance corpus

Test fixtures ported from [graphql/graphql-js](https://github.com/graphql/graphql-js)
under its MIT license (see `THIRD_PARTY_LICENSES.md` at the repo root). The
upstream files this corpus draws from:

- `src/language/__tests__/lexer-test.ts`
- `src/language/__tests__/parser-test.ts`
- `src/language/__tests__/schema-parser-test.ts`
- `src/language/__tests__/printer-test.ts` (printer-independent input cases)

This directory holds the Go-side test driver in `corpus_test.go` plus
optional `*.graphql` and `*.error` fixtures for cases where embedding the
input as a string literal is awkward.

## Adding a case

Most cases live as table entries in `corpus_test.go`. To add one:

1. Open `corpus_test.go`.
2. Find the table that matches the kind of test (`parserCases`,
   `lexerCases`, `schemaCases`, or `errorCases`).
3. Add an entry with the upstream test name, input, and expectation.

For inputs that contain awkward escape sequences or large blocks, drop a
`.graphql` file in this directory and reference it by filename in the
table.

## Conformance bar

In default fail-fast mode, the parser must:

- Accept every input the upstream `graphql-js` parser accepts.
- Reject every input the upstream rejects, with a `Syntax Error` whose
  position matches upstream byte-for-byte.

Error-message text is matched by substring on the relevant phrase; the
graphql-js-style snippet rendering produces the same multi-line format
but small wording differences are tolerated as long as the position is
correct.
