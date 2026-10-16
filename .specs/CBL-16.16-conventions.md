# CBL-16.16. cbl-protos: English contract comments and documentation

| Field | Value |
| :-- | :-- |
| Epic | E16, `Convention alignments` |
| Repository | `cbl-protos` |
| Size | M |
| Wave and parallel group | W0-CORR |
| Depends on | `CBL-16.1` (convention), `CBL-16.2` (pilot pattern) |
| Contract version | comments only; the freeze tags `cbl-protos@cf-1..cf-3.4` are unaffected |
| Status | done, review closed |

## 1. Context and goal

Customer correction of 16.10.2026: comments and documentation are written in English in
all files. cbl-protos is the contracts repository: .proto comments feed the generated Go
documentation, OpenAPI descriptions feed consumer tooling, so the translation covers
contracts, generated code (through regeneration), specs and docs.

## 2. Scope

In scope: English translation of all Russian comments in `proto/**` (14 files), all
Russian descriptions in `openapi/*.yaml` (billing, admin, webhooks), all eight specs
(CF-2/CF-3 freeze, 1.7, 2.10, 4.8, 7.1, 10.6), AGENTS.md and README.md; regeneration of
`gen/go` with the pinned plugin versions (protoc-gen-go v1.34.2, protoc-gen-go-grpc
v1.5.1) so generated comments match the sources; the `make generate` target fix
(`@v$(BUF_VERSION)`, the Go module version requires the v prefix).

Out of scope: any wire-visible change (packages, messages, fields, enums, field numbers,
options, file names), the freeze tags, other repositories.

## 3. Requirements

FR-1. Language: no Russian text remains anywhere in the repository (including
`gen/` after regeneration and the Makefile comments).

FR-2. Contracts safety: `buf lint` (STANDARD) green; `buf breaking` against the last
freeze tag `cbl-protos@cf-3.4` green (comments cannot affect it, verified); `go build ./...`
green; the generated diff contains only comment lines and the plugin version headers are
unchanged.

FR-3. AGENTS.md records the English-documentation rule for contracts and notes that the
repository is a contracts library with no service stand and no e2e scenarios (validation
is `make lint`, `make breaking`, `make build`, plus `make generate` after .proto edits).

## 4. Plan

Parallel translation (.proto, OpenAPI, specs, AGENTS/README) -> plugin install and
regeneration -> lint/breaking/build -> spec and journal -> commit
`CBL-16.16: English contract comments and docs`.

## 5. Reviews and runs

Verified: buf lint 0, buf breaking green against `cbl-protos@cf-3.4`, go build green,
generated diff comment-only with the same plugin headers, Cyrillic scan over the whole
repository clean, em-dash check clean. Incident during the task: the first `generate`
run cleaned `gen/` and failed because the pinned plugins were not in PATH; `gen/` was
restored from git and the run repeated with the plugins installed at the exact pinned
versions.
