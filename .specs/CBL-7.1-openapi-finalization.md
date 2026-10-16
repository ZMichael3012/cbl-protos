# CBL-7.1. Merchant OpenAPI v1 finalization

| Field | Value |
| :-- | :-- |
| Epic | E7 (Billing API and Gateway) |
| Repository | `cbl-protos` |
| Size | M (finalization of formats and codes, routes frozen by CF-2/CF-3) |
| Wave | W3-BILLING (first task) |
| Depends on | CF-3 (tag `cbl-protos@cf-3`) |
| Contract version | `billing.v1.yaml` 1.0.0, tag `cbl-protos@cf-3.1` (additive finalization on top of cf-3) |
| Status | done, review closed |

## 1. Context and goal

CF-2 froze the route composition of the merchant API draft (`1.0.0-draft`), and CF-3
finalized the webhook parts (`events[]` enum, https/SSRF, secret). 7.1 finalizes the
contract: complete response sets by CBL-003 codes, unified Problem components with
`Retry-After`, idempotency semantics (`Idempotent-Replayed`, 409
`idempotency_key_reuse`), required schema fields, security rules. After 7.1 the
contract is closed for W3-BILLING (7.2 - 7.6) and documentation 15.4.

## 2. Boundaries

In scope: `openapi/billing.v1.yaml` (formats, codes, headers, version 1.0.0),
synchronization of `CBL-003` section 2, AGENTS and the journal. Out of scope:
implementation of Billing API and Gateway (7.2 - 7.6), checkout (7.6),
`webhooks.v1.yaml` and `admin.v1.yaml` (frozen by CF-3), protobuf contracts.

## 3. Functional requirements

FR-1. The route composition does not change (CBL-003, section 2): `POST/GET /invoices`,
`GET /invoices/{id}`, `POST /invoices/{id}/cancel`, `POST /payouts`,
`GET /payouts/{id}`, `POST /refunds`, `GET /balances`, `GET /ledger/entries`,
`GET /reports/{report_type}`, `POST /webhook-endpoints`.

FR-2. Unified error model: `application/problem+json` (RFC 9457,
`type/title/status/code/request_id` are required); on all routes there are `401`
(with `WWW-Authenticate: Bearer`) and `403`; `404` means the resource is not found or
does not belong to the merchant (foreign identifiers are not disclosed; the ownership
check precedes business validation, a requirement for 7.4); `409` is a state or
idempotency key conflict; `422` is business validation; `429` and `503` carry
`Retry-After` and are declared on all proxying routes (503: `code`
`provider_unavailable` for the gateway, `service_unavailable` for internal services).
`code` is a stable machine-readable code: `idempotency_key_reuse`,
`invoice_not_cancellable`, `webhook_endpoint_exists`, `unknown_report_type`,
`insufficient_funds`, `validation_failed`, `provider_unavailable`,
`service_unavailable`, `rate_limited` (the list is open and additive); `detail`
contains no secrets or personal data.

FR-3. Idempotency: all `POST` requests require `Idempotency-Key` (frozen by CF-2);
the key is scoped to the merchant from the token (PostgreSQL unique index
`(merchant_id, idempotency_key)` - 7.3, T-02); the fingerprint is the method, path and
hash of the raw body; a replay of the fingerprint returns the original successful
response plus the header `Idempotent-Replayed: true` (on the first success there is no
header); the same key with a different fingerprint gives `409 idempotency_key_reuse`;
only successful responses are cached (errors, including 503, do not occupy the key);
the retention window is 24 hours (CBL-003, section 7); for webhook registration a
replay returns the original response together with the `signing_secret` (otherwise
the secret is lost irrecoverably).

FR-4. Required fields of response schemas and response headers: `Invoice` -
`id/status`; `Payout`/`Refund` - `id/status`; `Quote` - `rate/locked_until`;
`TimelineItem` - `state/occurred_at`; lists - `items`/`rows`; balance rows -
`currency/available` (structure: `available`, `reserved`, `pending`).

FR-5. Requests are closed against mass-assignment (`additionalProperties: false`):
`InvoiceCreateRequest`, `PayoutCreateRequest` (in addition to the CF-3 refund/webhook
schemas); `merchant_ref` has the pattern `^[A-Za-z0-9_-]{1,128}$` (business key of the
chart of accounts, CBL-003 section 9); `expires_in_sec` 60 - 86400 (if not set, 900
seconds); `amount` greater than zero (server-side validation); `metadata` - 32
properties, keys up to 64 characters, values up to 512 characters.

FR-6. Security: the merchant bearer API key (issuance/rotation/revocation are outside
the merchant API: 7.2, 12.x) is checked at the perimeter and in the BFF; `merchant_id`
is not accepted in requests, all resources are scoped by the token subject; ownership
of `account_id` in `/ledger/entries` gives 404 for a foreign one (the precondition is
2.10); `Idempotency-Key` is limited by character class and length (protection against
log injection); webhook registration follows the CF-3 https/SSRF norm, tightened to a
check on every delivery attempt (DNS rebinding, IPv4/IPv6 ranges, no redirect), the
secret is hex of at least 64 characters; the Problem `detail` carries no secrets.

FR-7. Version and rules: `version: 1.0.0`; `info.description` records the 7.1
finalization and the change rule (backward-compatible additive edits only; breaking
ones mean `/v2` and a new wave); extensible lists (`InvoiceStatus`, `PayoutState`,
`direction`) are marked `x-extensible-enum` with the client obligation to tolerate
unknown values. `CBL-003` sections 2 and 8 are extended (the `Idempotent-Replayed`
header, codes, the final version, the `cf-N.M` tag rule); AGENTS and the `cbl-protos`
journal are updated; `CBL-004` records the 7.1 status and the 7.4 contract
preconditions (tasks 2.10, 4.8); `CBL-000` records the Billing API -> Notification
link.

FR-8. Preconditions for 7.4 (recorded by the 7.1 review, implemented by separate tasks
before 7.4 starts): 2.10 - additive extension of Ledger gRPC (`ListMerchantAccounts`
with accounts, balances and ownership; the period filter of `ListEntries`) for
`/balances` and `/ledger/entries`; 4.8 - additive extension of Payment gRPC
(cf-3.3/cf-3.4: `rate`, `locked_until`, `txid`/`confirmations`, `expected_amount` of
creation (the source of the 201 response `pay_amount`), the `ListPayments` period) for
invoice completeness; both are cf-3.x packages. `pending` in balances is a Payment
aggregation (detected/confirming states), not Ledger data; the checkout contract is
created by task 7.6 as a separate artifact, and 7.1 records only `checkout_url`.

## 4. Checks

| Check | Command | Result |
| :-- | :-- | :-- |
| YAML parse | `python -c "import yaml; yaml.safe_load(...)"` | clean (10 routes, 11 operations, version 1.0.0) |
| route verification | against `CBL-003` section 2 (10 routes) | matches |
| code verification | responses of all operations against FR-2 | matches |
| buf | `buf lint`, `buf breaking --against ".git#tag=cbl-protos@cf-3"`, `go build ./...` | clean |
| em dash | `scripts/check-no-emdash.ps1` | clean |

## 5. Non-functional requirements

Routes and webhook parts do not change (frozen by CF-2/CF-3); additive-only after
1.0.0; the OpenAPI validator is not connected (KI-P1), so verification is manual.

## 6. Edge cases

A replay of an idempotent request with the same body after the resource state changes
returns the original response (the replay takes precedence over the current state);
different bodies under one key give 409 regardless of state; errors are not cached
(503 does not "poison" the key); foreign ids give 404 (ownership precedes business
validation); an unknown `report_type` gives 404 `unknown_report_type` (the type
catalog is 11.4, additive); invoice creation may return `creating`, and the client
polls GET; `Retry-After` is required on 429/503; `Idempotent-Replayed` only on
replays; a webhook registration replay returns the `signing_secret`; new states and
types are tolerated by the client (`x-extensible-enum`).

## 7. Review (CBL-001, step 6)

Four independent passes (15.10.2026). BLOCKER: `Idempotency-Key` scoping - the key is
bound to the merchant (unique index `(merchant_id, idempotency_key)`), the fingerprint
is method+path+raw body, replay of successful responses only - closed.
MAJOR: the lack of gRPC sources for 7.4 (`/balances`, `/ledger/entries`, the 201
invoice completeness) is recorded as preconditions via separate tasks 2.10 and 4.8
(cf-3.x) - FR-8; the version is finalized with the tag `cbl-protos@cf-3.1` (the
`cf-N.M` rule was added to CBL-003 section 8 and AGENTS section 5); idempotency was
extended with a 24 h TTL, non-caching of errors and fingerprint canonicalization; the
webhook endpoint lifecycle (rotation/deletion) is a trade-off: outside 1.0.0,
compensated via reissue (8.1/13.2); the `creating` semantics of the 201 response is
described; 503 is aligned across all operations; the webhook secret is hex and the
replay semantics are defined; the SSRF norm is extended (per-attempt resolver,
IPv6/IPv4 ranges, DNS rebinding); Admin API permissions are narrowed.
MINOR/NIT items closed: `x-extensible-enum` for states and direction, requests and
limits closed (`metadata`, `merchant_order_id`, `group_by`, `cursor`), UUIDv7
patterns for id/invoice_id, `Money.maxLength` and the `Quote.rate` pattern,
`request_id` required, `detail` without secrets, `WWW-Authenticate` on 401,
`Retry-After` as a component, the unified `ResourceID`, `Payout.invoice_id` for
refunds, optional `currency/network` in a refund, the `pending` explanation, Billing
API -> Notification in CBL-000. Accepted without changes: `group_by` as a free-form
string (validation in 11.4), `unknown_report_type` via INVALID_ARGUMENT mapping
(11.4/7.4), the second implementation of "secret shown once" for reads.

| Role | Key findings | Closure |
| :-- | :-- | :-- |
| No context | MAJOR: 422 for a report without a scenario; Idempotency-Key scope (method+path+body); MINOR/NIT: 503 descriptions, CBL-004/AGENTS statuses, `Asset` closed, rate pattern, secret replay, 503 on reads and cancel, duplicate id parameters, wording | 422 removed; fingerprint and scoping; descriptions and codes aligned; statuses in the commit; `Asset` opened; pattern; replay semantics; 503 added; `ResourceID`; corrected |
| Architecture | BLOCKER: `/balances`/`/ledger/entries` are not implementable on the current Ledger (owner, lists, pending, period); the invoice 201 without a source of `pay_amount/quote`; MAJOR: projection of `merchant_order_id/metadata/filters`, idempotency details (TTL/non-2xx/canonicalization/scope), closed enums, webhook lifecycle, tag, checkout, creating, 503; MINOR: refund `invoice_id`, refund currency, 404 mapping, 409 webhook, `request_id`, default expiry; NIT: TTL | tasks 2.10/4.8 (FR-8); FR-8; projection is 7.4/4.8; FR-3; `x-extensible-enum`; trade-off 8.1/13.2; tag cf-3.1; 7.6 separately; creation described; 503 aligned; `Payout.invoice_id`; `currency/network`; 11.4/7.4; code added; required; 900 s; reference |
| With context | MAJOR: version tag and statuses; MINOR: FR-7 partially, `Asset` closed, 503 inconsistency; NIT: inline duplicates, description style, AGENTS section 1 | tag cf-3.1 and the rule; statuses in the commit; `Asset` opened; 503 aligned; unified `ResourceID`/RetryAfter; descriptions added; section 1 updated |
| Security | BLOCKER: key scoping and the PG index composition (cross-merchant replay with the secret); MAJOR: SSRF per-attempt/IPv6, secret rotation (trade-off); MINOR: ownership-first, Idempotency-Key characters, maxLength of money/rate, `group_by` whitelist, id patterns, url whitespace, secret encoding, `detail`, metadata limits; NIT: 401 header, 503, descriptions | `(merchant_id, idempotency_key)`, fingerprint; SSRF norm; trade-off recorded; 404-first in 7.4; key pattern; maxLength/patterns; whitelist; patterns; url pattern; hex; `detail`; limits; `WWW-Authenticate`; 503; aligned |

## 8. Open questions

- The OpenAPI validator is not connected (KI-P1, 0.2): verification is manual (YAML,
  references, routes).
- The `report_type` catalog and the `rows` shape are finalized by 11.4 (additively);
  `unknown_report_type` requires an explicit INVALID_ARGUMENT mapping - 11.4/7.4.
- Error codes are additive: 7.x adds codes without breaking changes.
- Rotation/deletion of a webhook endpoint and secret is an additive extension of
  8.1/13.2: until it appears, secret compromise is remedied by registering a new
  endpoint and an ops procedure (trade-off of 1.0.0).
- The 7.4 preconditions (2.10, 4.8) are completed before 7.4 starts; cf-3.x packages
  are released as they become ready.

## 9. Work plan and commit

Finalization of `billing.v1.yaml`, fixes to `CBL-003`/`CBL-004`/`CBL-000`/AGENTS,
four review passes, the commit `CBL-7.1: merchant OpenAPI v1 finalization`,
annotated tag `cbl-protos@cf-3.1`.
