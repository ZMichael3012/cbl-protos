# AGENTS.md - cbl-protos

Repository of contracts for the CBL platform: Protobuf schemas for gRPC and Kafka events,
OpenAPI drafts, compatibility rules (buf). Mandatory global rules: `../../AGENTS.md`,
`../../.specs/CBL-001-development-process.md`, `../../.specs/CBL-003-contracts.md`.
The em dash (U+2014) is forbidden; only the hyphen is used.

## 1. Purpose and boundaries

In scope: `.proto` files, OpenAPI drafts, `buf` configuration, generated Go stubs
(`gen/`), compatibility rules.

Out of scope: service business logic, Schema Registry infrastructure (task 0.5),
Billing API implementation (7.1 finalized by this repository, code - in `cbl-billing-api`)
and admin (12.1).

Contract freeze points: CF-1 (Payment, Quote, Ledger, Connector, gateway events),
CF-2 (`payment.events`, `ledger.events`, `payout.commands`, `gateway.payout.events`),
CF-3 (the merchant webhook contract, Reporting `Query` v1, `recon.alerts` v1, Admin API v1).
See `CBL-004`.
Status: CF-1 is frozen by the tag `cbl-protos@cf-1` (13.10.2026), CF-2 - by the tag
`cbl-protos@cf-2` (15.10.2026, spec `.specs/CBL-CF2-freeze.md`), CF-3 - by the tag
`cbl-protos@cf-3` (15.10.2026, spec `.specs/CBL-CF3-freeze.md`); after a freeze only
backward-compatible additive fields are released (the tag letter `cf-2.1`,
CBL-10.6: `refund_invoice_id` in payout contracts).

## 2. Contracts and consumers

| Artifact | Consumers |
| :-- | :-- |
| `billing/payment/v1` | `cbl-payment`, `cbl-billing-api`, `cbl-reconciler` |
| `billing/quote/v1` | `cbl-quote`, `cbl-payment`, `cbl-billing-api` |
| `billing/ledger/v1` | `cbl-ledger`, `cbl-billing-api`, `cbl-payout`, `cbl-reconciler` |
| `billing/connector/v1` | `cbl-connector`, `cbl-payment`, `cbl-payout`, `cbl-reconciler` |
| `billing/payout/v1` | `cbl-payout`, `cbl-billing-api`, `cbl-backoffice` |
| `billing/reporting/v1` | `cbl-reporting`, `cbl-billing-api` |
| `billing/notification/v1` | `cbl-notification`, `cbl-billing-api` |
| `billing/events/v1` | all Kafka producers and consumers (`CBL-003`, section 6) |
| `openapi/billing.v1.yaml` | `cbl-billing-api`, `cbl-api-gateway` |
| `openapi/webhooks.v1.yaml` | `cbl-notification`, merchants, `cbl-billing-api` |
| `openapi/admin.v1.yaml` | `cbl-backoffice`, `cbl-api-gateway` |

## 3. Architecture

```
cbl-protos/
├── buf.yaml                # v2, lint STANDARD, breaking FILE
├── buf.gen.yaml            # protoc-gen-go, protoc-gen-go-grpc -> gen/go
├── proto/billing/...       # v1 domain packages
├── openapi/                # OpenAPI v1 drafts
├── gen/go/                 # generated stubs (committed)
└── Makefile
```

Module: `github.com/cbl/cbl-protos`; `go_package`:
`github.com/cbl/cbl-protos/gen/go/billing/<domain>/v1;<domain>v1`.

Evolution rules: field numbers are never reused; deleted fields are marked `reserved`;
breaking changes only via a new major package version and migration of all consumers
in a single wave (`CBL-003`, section 8).

## 4. Commands

```powershell
# lint (Docker)
docker run --rm -v "%CD%":/w -w /w bufbuild/buf:1.47.2 lint

# breaking against the latest freeze (see Makefile)
docker run --rm -v "%CD%":/w -w /w bufbuild/buf:1.47.2 breaking --against ".git#tag=cbl-protos@cf-3"

# generation (plugins required in PATH: %USERPROFILE%\go\bin)
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
go run github.com/bufbuild/buf/cmd/buf@v1.47.2 generate

# stub build
go mod tidy
go build ./...
```

This repository is a contracts library: it has no service stand and no e2e scenarios; validation runs are `make lint`, `make breaking` and `make build`, plus regeneration with `make generate` after .proto comment or schema edits.

## 5. Development rules

- Every change: a task spec in `.specs/`, then the proto edit, `buf lint`, `buf breaking`
  against the previous tag, generation, `go build ./...`.
- The contract version is pinned by the tag `cbl-protos@<version>`; the version number lands
  in the specs of dependent services. Incremental additive changes and wave artifact
  finalizations after `cf-N` are released with the tag `cf-N.M` and recorded in `CBL-003`
  section 8 (examples: `cf-2.1` CBL-10.6, `cf-3.1` CBL-7.1).
- Commit: `CBL-<epic>.<number>: <short message>`, the spec in the same commit. Contract
  freezes have no global plan task number: commit `CBL-CF<n>`, tag `cbl-protos@cf-<n>`
  (recorded by this spec and CBL-004).
- Comments and documentation are written in English: proto comments, specs, AGENTS.md, README; wire-level identifiers and protocol field names stay unchanged.
- Comments explain why, not what; errors via `%w`; panics are forbidden.
- External proto dependencies are forbidden except well-known types.

## 6. Trade-offs

| ID | Decision | Pros | Cons and compensation |
| :-- | :-- | :-- | :-- |
| T-P1 | buf v2 as the single lint/breaking/generate tool | uniform rules, one utility | coupling to the buf version; the version is pinned in the Makefile and CI |
| T-P2 | generated stubs are committed to `gen/` | consumers build without code generation | diffs grow; regeneration is mandatory on every proto change |

## 7. Known issues

| ID | Problem | Impact | Plan | Status |
| :-- | :-- | :-- | :-- | :-- |
| KI-P1 | OpenAPI validator (spectral/redocly) is not connected | typos in schemas until CI appears | task 0.2 | open |
| KI-P2 | Schema Registry is connected separately | events are validated after integration | task 0.5 | open |

## 8. Task journal

| CBL | Date | Status | Spec | Commit |
| :-- | :-- | :-- | :-- | :-- |
| CBL-1.7 | 2026-10-13 | done (buf lint, generate, go build, emdash - green; CF-1 frozen) | `.specs/CBL-1.7-protos-and-contract-freeze.md` | 5c63b8b, tag `cbl-protos@cf-1` |
| CBL-CF2 | 2026-10-15 | done (schema verification against W1, buf lint/breaking against cf-1, generate without a diff; CF-2 freeze) | `.specs/CBL-CF2-freeze.md` | tag `cbl-protos@cf-2` |
| CBL-10.6 | 2026-10-15 | done (additive refund_invoice_id: CreatePayoutRequest=7, Payout=12, PayoutEvent=11; buf lint/breaking against cf-2 green; tag cf-2.1) | `.specs/CBL-10.6-refund-fields.md` | 5b3d1e0 |
| CBL-CF3 | 2026-10-15 | done (webhooks.v1.yaml - a catalog with invoice.failed, self-contained data schemas, signature, retries; new billing/notification/v1; Admin API v1 - routing revision, subject from the token; Reporting Query v1 - comments; recon.alerts v1 + merchant_id; additive PAYOUT_EVENT_TYPE_REJECTED; buf lint/breaking against cf-2.1, generate, build green; CF-3 freeze) | `.specs/CBL-CF3-freeze.md` | tag `cbl-protos@cf-3` |
| CBL-7.1 | 2026-10-15 | done, review closed (billing.v1.yaml 1.0.0: unified Problem components and Retry-After, Idempotent-Replayed, mandatory schema fields, closed requests, bearer-scoping; CBL-003 section 2 synchronized) | `.specs/CBL-7.1-openapi-finalization.md` | e9283e1, tag `cbl-protos@cf-3.1` |
| CBL-2.10 | 2026-10-15 | done (additive ListMerchantAccounts and merchant_ref/period ListEntries for BFF statements 7.4; buf lint/breaking against cf-3.1, generate, build green; tag cf-3.2; implementation - cbl-ledger fbe0c45) | `.specs/CBL-2.10-ledger-fields.md` | c0d2c54 |
| CBL-4.8 | 2026-10-15 | done (Payment: rate/locked_until (cf-3.3), txid/confirmations, creation expected_amount (cf-3.4), ListPayments period; buf lint/breaking, generate, build green; implementation - cbl-payment bc903bd) | `.specs/CBL-4.8-payment-invoice-fields.md` | 96002bb, 6d1f442; tags `cbl-protos@cf-3.3`/`cf-3.4` |
| CBL-16.16 | 2026-10-16 | done (English contract comments and documentation: all .proto comments, OpenAPI descriptions, eight specs, AGENTS.md, README; gen/ regenerated with the pinned plugins (diff comment-only); buf lint/breaking/build green; make generate target fixed (v prefix)) | `.specs/CBL-16.16-conventions.md` | this commit |
| CBL-12.1 | 2026-10-16 | done (admin.v1.yaml finalization for the back-office: version 1.0.0, additive GET /audit-log with the AuditRecord schema and filters; review follow-up: the 503 ServiceUnavailable response with Retry-After; no proto changes; tags cf-3.5/cf-3.6) | `.specs/CBL-12.1-admin-audit-route.md` | this commit, tags `cbl-protos@cf-3.5`/`cf-3.6` |
| CBL-12.2 | 2026-10-16 | done (additive ConnectorAdminService: GetRoutingRules/PutRoutingRules with the revision-based optimistic concurrency, DisableProvider with the idempotent repeated disable; RoutingRuleAdmin; admin.v1.yaml gains the 422 response on payout reject and the RoutingRule finalization to the source-of-truth form (networks/assets arrays, minor-unit bounds); buf lint/breaking against cf-3.6, generate, build green; additive Payout.approved_by (cf-3.9) for the admin decision response; tags cf-3.7/cf-3.8/cf-3.9/cf-3.10) | `.specs/CBL-12.2-connector-admin-contract.md` | b8df893, 8af92fe, 4062b4e, f6d733d; tags `cbl-protos@cf-3.7`/`cf-3.8`/`cf-3.9`/`cf-3.10` |

## 9. Runbook

Releasing a new contract version:

1. PR with the proto edit and the spec update.
2. `buf lint` and `buf breaking` against the previous tag, stub regeneration.
3. Merge, tag `cbl-protos@<version>`, update of consumer service dependencies.
4. Update of `CBL-003` when the descriptive part changes, notification of the wave agents.
