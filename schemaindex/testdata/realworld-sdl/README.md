# Real-world SDL fixtures

These fixtures are pinned local snapshots used by `schemaindex` integration
tests. Tests must not download upstream SDL at runtime.

## github.graphql

- Source: https://github.com/octokit/graphql-schema
- Upstream path: `schema.graphql`
- Commit: `cdadcbc340d1e54cb1741dda696c62d2ed7f2261`
- License: MIT; see `THIRD_PARTY_LICENSES.md`.
- Copied ranges: lines 3, 135-176, and 1540-1607.
- Trimming: unrelated definitions were omitted. Copied definitions keep their
  original field and directive spelling and may reference omitted types.
- Coverage: custom directive definition, input object, interface, object
  implementing an interface, non-null/list-style field arguments, union, and
  deprecated field reasons.

## saleor.graphql

- Source: https://github.com/saleor/saleor
- Upstream path: `saleor/graphql/schema.graphql`
- Commit: `84a3d2924bed6b505c6fded639258a0b24dfa781`
- License: BSD-3-Clause; see `THIRD_PARTY_LICENSES.md`.
- Copied ranges: lines 1-20, 22-36, 1677-1692, 1744-1764, 2292-2310,
  2357-2363, 2366-2377, 17574-17606, 22420-22452, and 33996-34012.
- Trimming: large object and enum definitions were shortened to retained
  members and then closed locally. Unrelated definitions were omitted. Copied
  members keep their original field, type, directive, and deprecation spelling.
- Coverage: explicit query/mutation/subscription roots, custom directive
  definitions and usages, deprecated fields and input fields, enum values, input
  objects, and a real wrapped type argument `[WebhookEventTypeAsyncEnum!]!`.

## apollo-supergraph-demo.graphql

- Source: https://github.com/apollographql/supergraph-demo
- Upstream paths:
  - `subgraphs/users/users.graphql`
  - `subgraphs/products/products.graphql`
- Commit: `c385e0b79d988921b88ca3280141a06076c83a56`
- License: MIT; see `THIRD_PARTY_LICENSES.md`.
- Copied ranges: all of `subgraphs/users/users.graphql` lines 1-4 followed by
  all of `subgraphs/products/products.graphql` lines 1-30.
- Composition: the users base type is placed before the products subgraph
  extensions to exercise base-plus-extension folding. No Apollo base `type
  Query` is added, preserving extension-only `Query` coverage.
- Coverage: federation directive usages, repeatable directive definition,
  `extend type Query`, `extend type User`, extension-only index entries, and
  duplicate folded fields from base plus extension order.

## swapi.graphql

- Source: https://github.com/graphql/swapi-graphql
- Upstream path: `schema.graphql`
- Commit: `8c0bf5868f11fe84942548180751fceeb36bd606`
- License: BSD-3-Clause from pinned `package.json`; no root license file exists
  at this commit. See `THIRD_PARTY_LICENSES.md`.
- Copied ranges: lines 1-38, 112-146, 257-275, and 643-662.
- Trimming: unrelated connection and object types were omitted. The `Root`
  definition was shortened to retained fields and then closed locally.
- Coverage: explicit non-`Query` schema root, Relay-style connection and edge
  types, object interface implementation, and root field lookup.

## directive-only-extension.graphql

- Source: locally constructed for this coverage issue.
- License: project license.
- Coverage: directive-only object extension where folded member accessors return
  only base members while the raw extension still preserves directive metadata.
