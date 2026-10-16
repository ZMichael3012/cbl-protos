# CBL-1.7. The cbl-protos repository and CF-1 freeze

| Field | Value |
| :-- | :-- |
| Epic | E1, shared library |
| Repository | `cbl-protos` |
| Size | M |
| Wave and parallel group | W0-PROTO |
| Depends on | none |
| Contract version | CF-1 (this task is CF-1) |
| Status | done (locally, awaiting commit) |

## 1. Context and goal

The contracts in `CBL-003` are descriptive (KI-02). The task creates the artifacts:
Protobuf schemas for gRPC and Kafka events, and the OpenAPI v1 draft. After `buf lint`
and `buf generate` are green, the contract is declared frozen (CF-1), and wave W1
starts in parallel in four repositories.

## 2. Boundaries

In scope: `buf.yaml`, `buf.gen.yaml`, proto files, OpenAPI drafts, stub build.
Out of scope: service implementation, Schema Registry (infrastructure 0.5), merchant
OpenAPI finalization (7.1), admin OpenAPI (12.1).

## 3. Functional requirements

FR-1. Structure and go_package: the module `github.com/cbl/cbl-protos`, the packages
`github.com/cbl/cbl-protos/gen/go/billing/<domain>/v1;<domain>v1`.

FR-2. Proto files (fields strictly per `CBL-003`):

| File | Content |
| :-- | :-- |
| `proto/billing/common/v1/money.proto` | `Money{minor_units,asset,decimals}`, `Direction{DEBIT,CREDIT}` |
| `proto/billing/payment/v1/payment.proto` | `PaymentState` (CREATED, CREATING, PENDING, DETECTED, CONFIRMING, COMPLETED, EXPIRED, UNDERPAID, OVERPAID, FAILED, CANCELLED), `Payment`, `TimelineItem`, RPC `CreatePayment`, `GetPayment`, `ListPayments`, `CancelPayment` |
| `proto/billing/quote/v1/quote.proto` | RPC `LockQuote`, `GetRate`; `LockQuoteRequest{idempotency_key? no - per CBL-003: from_asset,to_asset,amount,ttl}`, response `{quote_id,rate,locked_until}` |
| `proto/billing/ledger/v1/ledger.proto` | `EntryInput`, `Entry`, `AccountBalance`; RPC `PostTransaction`, `Reserve`, `CommitReserve`, `ReleaseReserve`, `GetBalance`, `ListEntries` |
| `proto/billing/connector/v1/connector.proto` | `CanonicalPaymentStatus`, `Capability`; RPC `CreateInvoice`, `GetGatewayPayment`, `ListGatewayPayments`, `GetCapabilities` |
| `proto/billing/payout/v1/payout.proto` | `PayoutState`, `Payout`; RPC `CreatePayout`, `ApprovePayout`, `GetPayout` |
| `proto/billing/reporting/v1/reporting.proto` | RPC `Query{report_type,merchant_id,period,group_by,cursor,limit}` -> report rows `map<string,string>` |
| `proto/billing/events/v1/envelope.proto` | `EventMeta{event_id,event_type,occurred_at,producer,traceparent,schema_version}` (Kafka headers) |
| `proto/billing/events/v1/gateway_events.proto` | `GatewayWebhookRaw`, `CanonicalPaymentEvent` (types DETECTED, CONFIRMED, FAILED, EXPIRED, UNDERPAID, OVERPAID), `GatewayPayoutEvent` (SUBMITTED, CONFIRMED, FAILED) |
| `proto/billing/events/v1/payment_events.proto` | `PaymentEvent` (types CREATED, DETECTED, CONFIRMING, COMPLETED, EXPIRED, UNDERPAID, OVERPAID, FAILED; fields merchant_id, expected_amount, received_amount, fee, quote_id, rate, txid, state_version) |
| `proto/billing/events/v1/ledger_events.proto` | `LedgerEvent` (types TRANSACTION_POSTED, RESERVE_CREATED, RESERVE_COMMITTED, RESERVE_RELEASED; entries[], ref_type, ref_id) |
| `proto/billing/events/v1/payout_events.proto` | `PayoutCommand` (SendPayout), `PayoutEvent` (REQUESTED, APPROVED, COMPLETED, FAILED, COMPENSATED) |
| `proto/billing/events/v1/recon_events.proto` | `ReconAlert` (types MISSING_IN_LEDGER, AMOUNT_MISMATCH, STATUS_MISMATCH; severity INFO, WARNING, CRITICAL) |

FR-3. All mutating RPCs accept `idempotency_key` (per `CBL-003`, section 5); monetary
fields are only `billing.common.v1.Money`; time is `google.protobuf.Timestamp`,
identifiers are string (UUIDv7 at the service level).

FR-4. `buf.yaml` (v2): module `proto`, lint `STANDARD`, breaking `FILE`;
`buf.gen.yaml` (v2): `protoc-gen-go` and `protoc-gen-go-grpc`, output to `gen/go`,
`paths=source_relative`.

FR-5. OpenAPI drafts: `openapi/billing.v1.yaml` (paths from `CBL-003`, section 2, with
`Idempotency-Key`, the `Problem` scheme of RFC 9457 and bearer authentication) and
`openapi/admin.v1.yaml` (paths from `CBL-003`, section 3).

FR-6. Root `go.mod` of the module `github.com/cbl/cbl-protos`; `go build ./...`
compiles the generated stubs.

## 4. Contracts

The task creates the contracts; CF-1 is recorded after acceptance. After the freeze,
changes must be backward compatible only (`BACKWARD_TRANSITIVE`), see `CBL-003`,
section 8.

## 5. Data model and migrations

Not applicable.

## 6. State machines

The state enums from FR-2 correspond to the `CBL-000` state machines (section 5):
payment `created -> pending -> detected -> confirming -> completed / expired /
underpaid / overpaid`; payout `requested -> reserved -> sent_to_gateway ->
completed / failed -> compensated`.

## 7. Edge cases

- Package name conflicts: the `events` and `common` domains are separated from
  service packages.
- `buf` CLI is not installed: lint is run via Docker
  (`docker run --rm -v <repo>:/w -w /w bufbuild/buf:1.47.2 lint`) or via
  `go run github.com/bufbuild/buf/cmd/buf@v1.47.2`.
- The OpenAPI validator is unavailable locally: the files are checked manually, with
  finalization in 7.1 (KI).

## 8. Non-functional requirements

`buf lint` with no findings; stub generation is reproducible; no external proto
dependencies except well-known types.

## 9. Test plan

| Level | Scenarios |
| :-- | :-- |
| Smoke | `go run buf lint`; `go run buf generate`; `go build ./...` |
| Unit | not applicable (generated code) |
| Integration | not applicable |
| E2E | not applicable |
| Load | not applicable |

## 10. Work plan and commit

The commit `CBL-1.7: proto contracts and CF-1 freeze` after acceptance; no commits in
the current session.

## 11. Open questions

- OpenAPI validation (spectral/redocly) will be added to CI by task 0.2.
