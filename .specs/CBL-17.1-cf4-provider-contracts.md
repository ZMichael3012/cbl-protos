# CBL-17.1. CF-4: provider integration contracts

| Field | Value |
| :-- | :-- |
| Epic | E17 (task CBL-17.1) |
| Repository | `cbl-protos` |
| Size | M |
| Wave and parallel group | W5-PROVIDER, stage A |
| Depends on | `cbl-protos@cf-3.10`, the provider integration plan (section 4) |
| Contract version | `cbl-protos@cf-4` (freeze) |
| Status | done |

## 1. Context and goal

KI-01 is closed by the multi-provider decision (NOWPayments, CoinGate, BitPay, BTCPay
Server, 16.10.2026). Epic E17 (`CBL-004`) starts with contract work: the adapters, the
connector core, payment rules and reporting all need additive contract fields before
parallel development may start. This spec freezes CF-4, the contract point of the whole
provider wave; the descriptive contract (CBL-003, sections 4 - 7) is synchronized with
the umbrella in the same task.

Everything here is additive: no field is renumbered or removed, deleted semantics are
kept with deprecation markers, existing consumers keep working against `cf-3.10` until
they migrate in their own E17 tasks.

## 2. Requirements

FR-1. `billing/connector/v1` (connector.proto):

- `Capability` becomes provider-qualified (additive fields): `provider = 5`,
  `sandbox = 6`, `max_amount = 7`, `decimals = 8`, `supports_payouts = 9`,
  `supports_refunds = 10`. An empty `provider` means a legacy single-provider row;
  `max_amount` zero means no bound; `decimals` is the asset scale for minor-unit
  conversion (3.5 of the plan).
- `GetCapabilitiesRequest` gains `provider = 1` (filter; empty means all providers).
- `CreateInvoiceResponse` gains `checkout_url = 5`: hosted providers (CoinGate, BitPay,
  BTCPay, NOWPayments) return a payment page URL instead of an on-chain address;
  address-based providers leave it empty. The payment service stores exactly one of
  `address`/`checkout_url` and the BFF returns it (17.10, 17.14).
- `CanonicalPaymentStatus` gains `raw_status = 10`: the raw provider status string is
  preserved for reconciliation, reporting and support; the canonical status is the
  mapping result (3.4).
- `CanonicalPaymentStatusValue` gains `PENDING = 7`: a fetched invoice before any
  deposit is `pending` (for example `/v1/payment/` of NOWPayments); the previous enum
  could not express it, and fetch-to-verify (4.4) needs a stable value.
- `ConnectorAdminService` gains `EnableProvider(provider, reason) -> already_enabled`:
  the mirror of `DisableProvider` for the admin contour (17.15 closes KI-connector-09);
  adding an RPC to a service is non-breaking.

FR-2. `billing/events/v1` (gateway_events.proto):

- `GatewayWebhookRaw` gains `WebhookSignatureState signature_state = 7`
  (`UNSPECIFIED`, `VERIFIED`, `HMAC_FAILED`, `ECHO_MISMATCH`, `ABSENT`) and
  `bool fetch_required = 8`. The old `signature_verified = 6` is deprecated but stays
  populated during the transition as `signature_state == VERIFIED`, because the
  ingress and connector switch in 17.4/17.2; deleting the field would break the
  BACKWARD_TRANSITIVE schema contract of `gateway.webhooks.raw`.
- `CanonicalPaymentEvent` gains `raw_status = 10`.
- `GatewayPayoutEvent` gains `provider = 7`: without it `gateway.payout.events`
  consumers cannot attribute the external result (the companion of KI-payout-11,
  filled by 17.11).

FR-3. `billing/events/v1` (payment_events.proto): `PaymentEvent` gains (additive):

- `provider = 12` and `raw_status = 13`: reporting `fact_*` and reconciliation need the
  provider and its raw status (17.13, 17.15); `payment.id` remains the join key.
- fee split `gateway_fee = 14`, `network_fee = 15`, `fx_spread = 16` (COMMON-5):
  `fee = 6` stays the total charged fee; a zero component means the provider did not
  report it separately (for example an estimated fee before settlement).

FR-4. `billing/payment/v1` (payment.proto):

- `Payment` gains `checkout_url = 17` (hosted providers), `provider = 18`,
  `raw_status = 19`.
- `CreatePaymentResponse` gains `checkout_url = 9`. In the 7.4 invoice response
  exactly one of `pay_address` and `checkout_url` is filled (17.14).

FR-5. `openapi/admin.v1.yaml`: additive `POST /providers/{id}/enable` mirroring
`/disable` (204, idempotent, optional reason, 400/401/403/404/503); the operation
closes the admin symmetry for the provider control (17.15).

FR-6. Freeze procedure: `buf lint` green; `buf breaking` against `cbl-protos@cf-3.10`
green; generation with the pinned plugins (`protoc-gen-go` 1.34.2,
`protoc-gen-go-grpc` 1.5.1); `go build ./...` green; the full freeze tag is
`cbl-protos@cf-4` annotated (the practice of CF-1 - CF-3); task journal updated. CBL-003
sections 4 - 8 updated by the umbrella: webhook verification strategies and response
policies (4), the gRPC table rows for the Connector (5), the topic catalog and the
`PaymentEvent` example (6), the per-provider rate limit key (7), the CF-4 record (8);
section 3 route list and the CBL-000/CBL-004/README references follow in the same sync.

FR-7. Review follow-up (additive, tag `cbl-protos@cf-4.1`):

- Fee path: `CanonicalPaymentStatus` and `CanonicalPaymentEvent` gain the same fee split
  (`gateway_fee = 11`, `network_fee = 12`, `fx_spread = 13`): the connector is the only
  component that sees the provider fee data (COMMON-5), so the frozen `payment.events`
  split had no producer path without it; payment validates and re-publishes the values
  (17.10).
- CBL-003 section 4 hardening: the CoinGate `token` is removed from the raw body before
  it is persisted or published (a verification secret must never land in the outbox,
  Kafka or logs); token comparison is constant-time; the response policy applies after
  the allowlist and intake limits (off-allowlist `403` and `429` keep their codes and
  raise an alert); `fetch_required` is mandatory for CoinGate and BitPay regardless of
  the flag, a fetch failure is retried and never published unchecked; unverified
  deliveries are published with a random `event_id` (they must not poison the
  deterministic dedup key of a genuine delivery) and never produce canonical events
  (fail closed, only `VERIFIED` is trusted; `UNSPECIFIED` falls back to
  `signature_verified` until the 17.4 cutover); mismatch alerts are aggregated.
- CBL-003 section 6 and the proto comments: fee split validation rule (non-negative,
  same asset and decimals as the payment amount; the producer validates; an inconsistent
  split raises a recon alert and does not reach the ledger).
- Comments: `max_amount` unset semantics, `decimals` authority versus the `Money` scale,
  `EnableProvider` wording with the breaker/cache caveat, `checkout_url` wording (hosted
  flow versus address flow for NOWPayments), `fetch_required` and `signature_state`
  transition semantics.

Review closure (4 reviewers, 16.10.2026): no-context - approved; architecture - approved
after the fee path fix (FR-7); with-context - approved (tag annotation fixed by
re-issuing `cf-4` annotated; doc sync items closed in the umbrella follow-up);
security - approved after the token redaction and fee validation rules (FR-7).

Out of scope (separate additive tags of later E17 tasks if needed): `quote.proto` audit
fields (17.10), `recon.alerts` provider (17.13), notification contract changes (17.15),
`REFUNDED` enum values (17.12), payout provider state `awaiting_provider` (17.11).

## 3. Work plan

Spec -> proto and OpenAPI edits -> `buf lint`, `buf breaking` against cf-3.10,
`make generate`, `go build ./...` -> commit `CBL-17.1: CF-4 provider integration
contracts` with the spec and the tag `cbl-protos@cf-4` -> umbrella: CBL-003 sections
4 - 7, CBL-004 E17 status, journal.

Consumers of the new fields (implementation tasks, same wave): `cbl-connector` (17.2,
17.11), `cbl-webhook-ingress` (17.4), `cbl-payment` (17.10), `cbl-quote` (17.10),
`cbl-billing-api` (17.14), `cbl-reporting` (17.15), `cbl-reconciler` (17.13),
`cbl-backoffice` (17.15).
