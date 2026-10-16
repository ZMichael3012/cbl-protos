# CBL-CF3. Freeze of Notification, Reporting, recon.alerts and Admin API contracts

| Field | Value |
| :-- | :-- |
| Epic | Contract freeze point CF-3 (CBL-004, section 3) |
| Repository | `cbl-protos` |
| Size | S (freeze; Reporting/recon.alerts were created in 1.7, webhooks, Admin API and Notification gRPC are finalized before the freeze) |
| Wave | end of W2, before the W3 start |
| Depends on | CF-2 (tag `cbl-protos@cf-2.1`), W2 implementations (4.x, 10.x - to verify semantics) |
| Contract version | `cbl-protos@cf-3` (this tag) |
| Status | in progress |

## 1. Context and goal

CF-2 froze the phase 2 event schemas and the merchant OpenAPI draft. By the end of W2,
Payment (4.x) and Payout (10.x) are implemented, the actual states and events that
merchant webhook delivery depends on (Notification, 8.x) are determined, as are the
contractual compositions of reports (Reporting, 11.x) and discrepancies (Reconciler,
9.x). CF-3 fixes the W3 wave contracts: merchant webhooks (types and signature),
Reporting `Query` v1, `recon.alerts` v1, the Admin API v1 draft and Notification
endpoint management, after which waves W3 and W4 start without changes to the paired
contracts.

## 2. Boundaries

In scope: the new artifact `openapi/webhooks.v1.yaml` (envelope, type catalog,
signature, retries, data schemas), freeze of the `openapi/admin.v1.yaml` draft,
the new package `proto/billing/notification/v1` (endpoint management), verification
of `proto/billing/reporting/v1/reporting.proto` and
`proto/billing/events/v1/recon_events.proto` (additive `merchant_id`),
the additive `rejected` type in `payout.events`, updates to `CBL-003`/`CBL-004`,
AGENTS, the tag `cbl-protos@cf-3`.

Out of scope: finalization of the merchant OpenAPI (7.1) and Admin API (12.1),
implementation of services 8.x/9.x/11.x/12.x, Schema Registry infrastructure (0.5),
execution of manual discrepancy adjustments (design 12.3 after 9.2).

## 3. Freeze composition

| Schema | File | Consumers |
| :-- | :-- | :-- |
| Merchant webhooks: envelope, type catalog, signature, retries, data schemas | `openapi/webhooks.v1.yaml` (new, 1.0.0-cf3) | `cbl-notification` (8.1 - 8.3), merchants (15.4), Billing API (7.1) |
| Admin API v1 (draft, finalization 12.1) | `openapi/admin.v1.yaml` (1.0.0-draft) | `cbl-backoffice` (12.x), API Gateway |
| Notification endpoint management | `proto/billing/notification/v1/notification.proto` (new) | `cbl-notification` (8.1, server), `cbl-billing-api` (7.x, client) |
| Reporting `Query` v1 | `proto/billing/reporting/v1/reporting.proto` | `cbl-reporting` (11.4), `cbl-billing-api` |
| `recon.alerts` v1 (`ReconAlert` + `merchant_id`) | `proto/billing/events/v1/recon_events.proto` | Reconciler (9.2), back-office (12.3), Notification, ClickHouse (11.x) |
| `payout.events`: additive `PAYOUT_EVENT_TYPE_REJECTED` | `proto/billing/events/v1/payout_events.proto` | Notification (8.x: `payout.failed`/`refund.failed` with status `rejected`), ClickHouse |
| `EventMeta` (envelope) | `proto/billing/events/v1/envelope.proto` | unchanged (frozen by CF-1) |

## 4. Functional requirements

FR-1. Notification (types, data and signature). `openapi/webhooks.v1.yaml`
fixes:

- delivery envelope: headers `X-Billing-Signature` (scheme
  `t=<unix>,v1=<hex HMAC-SHA256(secret, t + "." + body)>`, hex in lowercase
  letters), `X-Billing-Event-Id` (equals the body `id` and does not change on
  retries, only `t` changes), `Content-Type: application/json`; body
  `{"id":"evt_<uuidv7>","type":"<type>","created_at":"<RFC 3339 UTC>","data":{...}}`;
- type catalog and source mapping: `payment.events`
  (`detected/confirming/completed/expired/underpaid/overpaid/failed`) ->
  `invoice.*`; `payout.events` without a refund marker (`completed` ->
  `payout.completed`; terminal `failed/compensated/rejected` ->
  `payout.failed` with the exact state in `data.payout.status`); `payout.events`
  with a non-empty `refund_invoice_id` -> `refund.completed`/`refund.failed`
  (`data.refund.status`); `payment.created` and `payout.requested/approved`
  produce no webhooks;
- terminal states are delivered in full: `invoice.failed` closes a payment
  failure (the `failed` state exists in Payment 4.1 and is emitted by 4.7),
  `rejected` is an operator veto: on CF-3 the additive type
  `PAYOUT_EVENT_TYPE_REJECTED` was added to `payout.events`, and `cbl-payout`
  emits it in the transition transaction (accompanying commit `CBL-10.5 (fix)`);
  without this, the promise of the `rejected` status had no source;
- data schemas are self-contained (they do not reference the `billing.v1.yaml`
  draft): `data.invoice` carries `id`, `status`, `amount`, `paid_amount`, `fee`,
  `rate`, `txid`, `occurred_at`; `data.payout` and `data.refund` carry `id` (for
  a refund, `invoice_id` is the original invoice being refunded), `status`,
  `amount`, `fee`, `txid`, `failure_reason`, `occurred_at`; field values are
  taken from the source event, and the merchant obtains the missing ones
  (address, deadlines) via REST GET; the catalog is extended additively, and an
  unknown type is accepted by the merchant (2xx) and ignored;
- the signature is verified over the raw body (constant-time comparison), the
  `t` window is +/- 5 minutes; the secret is at least 32 random CSPRNG bytes,
  is shown once and is never logged; the merchant deduplicates by the body
  `id`, checking it against `X-Billing-Event-Id`;
- delivery: success is `2xx` only (3xx is not success), the attempt timeout is
  10 s, retries after 1 s, 5 s, 30 s, 2 min, 15 min, 1 h; after exhaustion
  `failed` and manual retry (8.3); delivery order is not guaranteed;
- registration (`billing.v1.yaml`, `POST /v1/webhook-endpoints`): `events` is
  an enum of the CF-3 catalog, `minItems: 1`, `uniqueItems`, an unknown type
  yields 422; delivery is filtered by the list; new catalog types are not added
  to a subscription automatically (re-registration is required); `url` is
  `https` only, without userinfo and IP literals, platform SSRF protection is
  normative (blocking of private/loopback/link-local/metadata after DNS
  resolution, no redirects).

FR-2. Reporting `Query` v1: proto comments pin down the semantics without
changing the wire format: `merchant_id` is filled in by Billing API from the
token (11.4), direct gRPC calls are internal; `limit` (0 means the default of
20, like `Limit` in billing.v1.yaml, maximum 100); `cursor` is opaque; monetary
values and rates in `ReportRow.values` are decimal strings in canonical form
(no exponent, float is forbidden); the `report_type` catalog is fixed by
11.4/15.4; the RPC deadline is 5 s (CBL-003, section 5).

FR-3. `recon.alerts` v1: `ReconAlert` includes the field `merchant_id = 8`
(protocol comment: the merchant UUID, filled in by producer 9.2 only when bound
to a merchant; when empty, the record is for back-office and ClickHouse only,
with no delivery to the merchant); `alert_id` equals the envelope `event_id`
(a single deduplication key for inbox and ClickHouse); `expected`/`actual` are
strings (decimal amounts or state names), the topic key is `ref_id`, and
`created_at` in the body is authoritative. Merchant notifications about
discrepancies are not part of the CF-3 type catalog (the type decision is
8.x/12.3, additively via cf-3.x).

FR-4. Admin API v1: the `openapi/admin.v1.yaml` draft is frozen: the CBL-003
section 3 routes (`/payouts/{id}/approve`, `/reject`, `/routing-rules` GET/PUT,
`/reconciliation/discrepancies` GET with a `merchant_id` filter and
`/{id}/resolve`, `/providers/{id}/disable`), operator bearer token,
`application/problem+json`, responses `400/401/403/404/409/422`, two-stage
approval under the threshold policy (two operators by default; a repeated vote
gives 409; a veto by one operator is sufficient), the body `operator_id` must
match the token subject, otherwise 403 (the audit is by token subject);
replacement of routing rules is versioned (`revision` from GET, a mismatch
gives 409), requires a separate RBAC role and carries a note about the
application delay (30 - 60 s, Connector cache); repeated provider disable is
idempotent (204); an audit record (who, what, before/after) for every action;
finalization is 12.1. No separate `Idempotency-Key` is introduced on Admin API
mutations: idempotency of votes/decisions is provided by the domain
(append-only, 409 on repeat); revisit is 12.1 (trade-off).

FR-5. Notification gRPC: the new package `billing/notification/v1` with
`NotificationService.RegisterWebhookEndpoint` (`idempotency_key`, `merchant_id`,
`url`, `events[]` -> `endpoint_id`, `url`, `events[]`, `signing_secret`) is the
channel from Billing API (7.x) and Back-office (12.x) to Notification (8.1);
`merchant_id` is taken from the merchant token, events are validated against
the CF-3 catalog; the deadline is 2 s (CBL-003, section 5).

FR-6. Compatibility: `buf lint` (STANDARD) with no findings; `buf breaking`
against the tag `cbl-protos@cf-2.1` (FILE) is clean (comments and additive
fields/types are allowed); `buf generate` updates `gen/go` (comments and the
new field/type land in the stubs), `go build ./...` is green.

FR-7. Recording: `CBL-003` sections 3/4/5/6 are brought in line with the
contracts (the catalog with `invoice.failed`, the header equals the `id`, id
stability on retries, `rejected`, `merchant_id`, the Notification line),
`CBL-000` section 2 (back-office consumes `recon.alerts`), AGENTS of
`cbl-protos` (artifacts, consumers, journal); annotated tag `cbl-protos@cf-3`.
After the freeze, changes must be BACKWARD-compatible only (incremental tags
`cf-3.x` per the `CBL-003` section 8 rule).

## 5. Checks

| Check | Command | Result |
| :-- | :-- | :-- |
| lint | `buf lint` (bufbuild/buf:1.47.2) | clean |
| breaking | `buf breaking --against ".git#tag=cbl-protos@cf-2.1"` | clean |
| generate | `buf generate` against the committed `gen/go` | diff of comments, `MerchantId` and the `PAYOUT_EVENT_TYPE_REJECTED` type in the stubs |
| build | `go build ./...` | clean |
| verification | webhooks catalog against `CBL-003` section 4 and `payment_events.proto`/`payout_events.proto`; Admin API against `CBL-003` section 3; Notification RPC against task 8.1 | matches |
| payout | emission of `payout.rejected` on veto (`cbl-payout`, integration test) | integration green |
| em dash | `scripts/check-no-emdash.ps1` | clean |

## 6. Non-functional requirements

The Reporting and `recon.alerts` wire schemas change additively only
(`merchant_id`; `PayoutEventType.REJECTED` is added as a value at the end of the
enum); the OpenAPI artifacts are a mandatory contract for 8.x/12.x after the
freeze; the OpenAPI validator is not connected (KI-P1), verification is manual;
the `BACKWARD_TRANSITIVE` mode is upheld by `buf breaking` until Schema Registry
is connected (KI-P2).

## 7. Edge cases

A change after the freeze (BACKWARD-compatible only; otherwise a major version
and a repeated cycle); a new webhook type is an additive catalog value plus
explicit re-registration (automatic inclusion in subscriptions is not
guaranteed); `data.refund` for partial refunds carries the amount of the
specific payout-refund, aggregation is Billing API 7.x; a merchant without
registered endpoints receives no deliveries (8.1); the attempt timeout is 10 s,
a 3xx response is not success, an unknown type is ignored with 2xx; loss of the
platform response after successful processing by the merchant leads to a retry
with the same `id` and a new `t`, with deduplication on the merchant side;
signing secret rotation and double signing during rotation are an additive
extension of 7.1/8.1 (13.2) and are not part of cf-3; a divergence between
`CBL-003` and the artifacts is fixed by this spec before the tag.

## 8. Review (CBL-001, step 6)

Four independent passes (15.10.2026). There are no blockers after the fixes.
BLOCKER items closed: the promise of the `rejected` status without a source
(additive `PAYOUT_EVENT_TYPE_REJECTED` + emission in `cbl-payout`, integration
test); impersonation of operator authorship (the RBAC/audit subject comes only
from the bearer token, the body `operator_id` must match, otherwise 403).
MAJOR items closed: `data.*` no longer references the mutable `billing.v1.yaml`
draft (self-contained data schemas); the missing gRPC contract for Notification
endpoint management (the new `billing/notification/v1`); the registration
`events[]` received the catalog enum and `minItems`; `merchant_id` was added to
`Discrepancy` and to the list filter; the source of the `data` fields is
documented (event + REST GET); `operator_id`/routing revision/Admin API
idempotency are recorded; the back-office contradiction (consumes
`recon.alerts`) was removed in `CBL-000` and AGENTS; the wording of the
`resolved` retry and of idempotency was corrected; `X-Billing-Event-Id` is tied
to the body `id` and is stable on retries; https/SSRF and the secret
requirements are normative. MINOR/NIT items closed: the UUIDv7 pattern for ids,
lowercase hex in the signature, `additionalProperties: false` on requests,
`PayoutAdmin.amount` as `Money`, idempotent repeated disable (204), proto
comments (Reporting: default 20, canonical decimal form; recon: producer
validation, alert_id = event_id), `recon.alert` in the `event_type` catalog,
the order of document statuses (the tag in the same commit), the Makefile
comment about incremental tags. Accepted without changes: the absence of
`Idempotency-Key` on Admin API mutations (domain idempotency, revisit is 12.1);
the merchant-facing discrepancy notification type (8.x/12.3, additively);
signing secret rotation (7.1/8.1, 13.2); execution of `manual_adjustment`
(12.3 after 9.2).

| Role | Key findings | Closure |
| :-- | :-- | :-- |
| No context | BLOCKER: the `rejected` status without an event type; MAJOR: `data` from the draft, field sources, merchant_id filter/field in Admin API, discrepancy notifications without a type; MINOR: file version vs spec, breaking tag before release, wording of the resolved retry, patterns, edge cases, trade-offs; NIT: UUIDv7 pattern, terminology, event_type `recon.alert`, tags | REJECTED in cf-3 + emission; self-contained schemas; field sources recorded; merchant_id/filter; type decision explicitly deferred; version agreed; the order review-fix-tag; wording; patterns/edge cases; trade-offs; comments; terminology; catalog; Makefile |
| Architecture | BLOCKER: rejected without emission; MAJOR: no gRPC for Notification endpoint management, back-office/Reconciler channel, `data` not derivable from the topics, `events[]` without a catalog, admin mutations without idempotency, operator_id vs token, routing without revision, ClickHouse schema, merchant_id filter, 12.2 surface, drift toward the draft, 8.3 contract; MINOR/NIT: statuses/tag, trade-offs, default limit, Refund overload, secret rotation | REJECTED + emission; `billing/notification/v1`; back-office consumes `recon.alerts` + design 12.3; self-contained schemas; `events[]` enum; domain idempotency + 12.1; token subject; revision; alert_id = event_id; merchant_id/filter; explicitly deferred to 12.2; self-contained schemas; 8.3 is manual retry in 8.x + statuses; tag in the commit; trade-offs section 9; default 20; terminology; deferred to 7.1/8.1 |
| With context | No BLOCKER; MAJOR: rejected, admin file version, operator_id; MINOR: CBL-003 header, non-producing events, recon.alerts/back-office consumers, resolved retry, statuses ahead of the tag; NIT: merchant_id additivity, default limit, AGENTS/Makefile comments, UUIDv7 pattern | REJECTED + emission; version agreed; token subject; header updated; caveat added; back-office/`recon.alert`; wording; single commit and tag; wording in proto; default 20; corrected; pattern strengthened |
| Security | BLOCKER: operator_id from the body (four-eyes bypass); MAJOR: `X-Billing-Event-Id` outside the signature and retry stability, https/SSRF not normative, secret requirements, routing without concurrency control, merchant_id semantics; MINOR/NIT: header pattern, `events` without a catalog, conditional reason, mass-assignment, disable repeat, statuses/tag, `rejected` status, unknown type, Money/string format | token subject + 403; id = header, stability, merchant-side check; https pattern + SSRF norm; secret CSPRNG/once/logs; revision + 409 + RBAC role; producer validation and empty value; pattern; enum/minItems/422; 422 for reason; additionalProperties false; 204 repeat; single commit; closed by the REJECTED type; ignore rule; canonical format |

## 9. Open questions

- The OpenAPI validator is not connected (KI-P1, 0.2): the artifacts were
  checked manually for composition and compliance with `CBL-003`.
- Until Schema Registry is connected (0.5), the `BACKWARD_TRANSITIVE` mode is
  not enforced by the infrastructure: compatibility is upheld by `buf breaking`
  (KI-P2).
- KI-01 does not affect CF-3: the contracts are provider-neutral (string
  requisites); Reconciler 9.x remains blocked by task 3.2.
- Populating `merchant_id` in `recon.alerts` is a requirement for 9.2 when
  publishing merchant-bound discrepancies; delivery of discrepancy
  notifications to merchants is not part of the CF-3 catalog (the type is a
  decision of 8.x/12.3, additively).
- Signing secret rotation and double signing during rotation are an additive
  extension of 7.1/8.1 (13.2); the absence of `Idempotency-Key` on Admin API
  mutations is a trade-off of FR-4, revisit in 12.1.
- `ListWebhookEndpoints`/`Delete`/secret reissue in `billing/notification/v1`
  are not part of the minimal 8.1 contract and are added additively (cf-3.x)
  when the requirement appears.

## 10. Work plan and commit

Verification of and fixes to the proto (comments, `merchant_id`, `REJECTED`,
`notification/v1`), regeneration, OpenAPI fixes (`webhooks.v1.yaml`,
`admin.v1.yaml`, `billing.v1.yaml`), four review passes, fixes to
`CBL-003`/`CBL-004`/`CBL-000`/AGENTS, the accompanying commit `CBL-10.5 (fix)`
in `cbl-payout` (emission of `rejected`), the commit `CBL-CF3: freeze of
Notification, Reporting, recon.alerts and Admin API contracts`, annotated tag
`cbl-protos@cf-3`.
