# CBL-CF2. Freeze of v1 event schemas and merchant OpenAPI

| Field | Value |
| :-- | :-- |
| Epic | Contract freeze point CF-2 (CBL-004, section 3) |
| Repository | `cbl-protos` |
| Size | S (freeze, schemas created in 1.7) |
| Wave | end of W1, before the W2 start |
| Depends on | 1.7 (CF-1), W1 implementations (2.7, 3.5, 5.1, 6.3) |
| Contract version | `cbl-protos@cf-2` (this tag) |
| Status | done, review closed |

## 1. Context and goal

CF-1 froze gRPC v1 and the gateway events (`gateway.webhooks.raw`,
`gateway.events.canonical`). W1 tasks implemented producers and consumers and revealed
the actual schemas. CF-2 fixes the phase 2 event schemas (`payment.events`, `ledger.events`,
`payout.commands`, `gateway.payout.events`) and the v1 merchant OpenAPI draft, after which
wave W2 (Payment, Payout, Reconciler) starts without changes to the paired contracts.

## 2. Boundaries

In scope: verification and freeze of the v1 event schemas, the v1 merchant OpenAPI draft,
updates of the system specs and the version tag.
Out of scope: finalization of the merchant OpenAPI (7.1) and the admin API (12.1), CF-3
contracts (Notification, Reporting, `recon.alerts`, Admin API), Schema Registry
infrastructure (0.5).

## 3. Freeze composition

| Schema | File | Topic | Consumers |
| :-- | :-- | :-- | :-- |
| `PaymentEvent` (types created...failed) | `proto/billing/events/v1/payment_events.proto` | `payment.events` | Ledger (2.7), Notification (8.x), ClickHouse (11.x) |
| `LedgerEvent` + `LedgerEntry` (transaction.posted, reserve.*) | `proto/billing/events/v1/ledger_events.proto` | `ledger.events` | Payment (4.x), Payout (10.x), ClickHouse |
| `PayoutCommand` | `proto/billing/events/v1/payout_events.proto` | `payout.commands` | Connector (10.3) |
| `GatewayPayoutEvent` (submitted/confirmed/failed) | `proto/billing/events/v1/gateway_events.proto` | `gateway.payout.events` | Payout (10.x), Reconciler (9.x) |
| `PayoutEvent` (requested...compensated) | `proto/billing/events/v1/payout_events.proto` | `payout.events` | Notification, ClickHouse |
| `EventMeta` (envelope) | `proto/billing/events/v1/envelope.proto` | all topics | all |
| Merchant OpenAPI draft | `openapi/billing.v1.yaml` (1.0.0-draft) | - | Billing API (7.x), API Gateway |

`EventMeta` was frozen back in CF-1 and has not changed; `payout.events` and the OpenAPI
draft are part of the cf-2 tag for wave W2 (CBL-004, CF-2 was extended with `payout.events`).
Wire schemas (field numbers, types) have not changed since cf-1: only semantic comments
were added (reserve events, `occurred_at`, `state_version`); generation updates the stubs
without changing the wire format.

## 4. Functional requirements

FR-1. Verification of schemas against W1 implementations: `PaymentEvent` matches the 2.7
consumer contract (completed -> Dr clearing / Cr merchant / Cr fee; the consumer reads
payment_id, type, merchant_id, received_amount, fee; the remaining fields are available to
W2/W3 consumers); `LedgerEvent` matches the 2.7 producer (transaction.posted,
reserve.created/committed/released; transaction_id, ref_type, ref_id, entries[]; the
semantics of reserve events is fixed by proto comments); `GatewayWebhookRaw`
and `CanonicalPaymentEvent` are pinned by their use in 5.1 and 3.5; `GatewayPayoutEvent`
and `PayoutCommand` are ready for W2. The semantics of `occurred_at` (the body is
authoritative, the envelope header mirrors it) and `state_version` (a monotonic transition
counter) is fixed by proto comments.

FR-2. Compatibility: `buf lint` (STANDARD) and `buf breaking` against the tag
`cbl-protos@cf-1` (FILE) - no findings; generation is idempotent (`buf generate`
changes `gen/go` only following proto comments), `go build ./...` is green.

FR-3. Documentation: `CBL-003` section 6 is aligned with the actual schemas
(the `PaymentCompleted` example replaced with `PaymentEvent`, the catalog refined), the
freeze point statuses are updated (CBL-003 section 8, CBL-004 section 3). The decision on
the `ledger.events` key (open question T-LED-09): the `account_id` key of the first entry
is kept; a revision is due only if a requirement for strict ordering of account events
appears.

FR-4. Pinning: annotated tag `cbl-protos@cf-2` on the freeze commit; the rule after a
freeze is backward compatible changes only (BACKWARD_TRANSITIVE); breaking changes require
a new major package version and migration of all consumers in a single wave. Contract
freezes have no global plan task number: commit `CBL-CF2`, tag
`cbl-protos@cf-2` (the caveat is recorded in AGENTS).

FR-5. The merchant OpenAPI draft (1.0.0-draft) covers the routes of `CBL-003` section 2
(the external merchant REST): `/invoices`, `/invoices/{id}`, `/invoices/{id}/cancel`,
`/payouts`, `/payouts/{id}`, `/refunds`, `/balances`, `/ledger/entries`,
`/reports/{report_type}`, `/webhook-endpoints`; format finalization - 7.1;
the draft freeze fixes the route composition and the security review edits
(the `account_id` owner check in `/ledger/entries`, `Idempotency-Key`
on `/webhook-endpoints`, non-negative `Money`).

## 5. Checks

| Check | Command | Result |
| :-- | :-- | :-- |
| lint | `buf lint` (bufbuild/buf:1.47.2) | clean |
| breaking | `buf breaking --against ".git#tag=cbl-protos@cf-1"` | clean (the gate was verified by a field sanity mutation) |
| W1 verification | cross-check of schemas against `cbl-ledger/internal/usecase/{ledger_events.go,payment_consumer.go}`, `cbl-connector/internal/usecase/normalizer.go`, `cbl-webhook-ingress/internal/usecase/handler.go` | matches FR-1 |
| generate | `buf generate` (against the committed `gen/go`) | diff only from proto comments |
| build | `go build ./...` | clean |
| em dash | `scripts/check-no-emdash.ps1` | clean |

## 6. Non-functional requirements

Wire schemas do not change at freeze time (only wire-compatible comment clarifications are
allowed); the Schema Registry subject follows the topic name, mode
`BACKWARD_TRANSITIVE` (0.5; until SR is connected, compatibility is held by `buf breaking`);
stubs in `gen/go` are committed and reproducible.

## 7. Edge cases

Schema change after the freeze (BACKWARD-compatible only; otherwise major); divergence
between `CBL-003` and the proto (the system spec was updated by this commit); consumers are
obliged to ignore new fields (protobuf); enum values are appended only at the end.

## 8. Review (CBL-001, step 6)

Four independent passes (15.10.2026). No blockers. MAJOR: the `payout.commands` row
in CBL-003 did not match the frozen schema (fixed), document statuses anticipated the freeze
(closed by the commit and tag simultaneously with this fix), the OpenAPI draft pinned an
IDOR risk in `/ledger/entries` and a mutation without `Idempotency-Key` (fixed before the
freeze). MINOR/NIT closed: the freeze composition is agreed in CBL-003/CBL-004, the FR-5
reference to section 2, the CF-2 draft label, the CBL-003 header, the semantics of
reserve events and occurred_at/state_version in proto comments, the ledger.events key kept
(T-LED-09), the `make breaking` target, the caveat about `CBL-CFn`.

| Role | Key findings | Closure |
| :-- | :-- | :-- |
| No context | MAJOR: the payout.commands row (SendPayout/asset) against the schema; the composition in the spec is wider than CBL-004; the FR-5 reference to section 3; statuses anticipate the tag; the "Draft CF-1" label; FR-1 without verification artifacts; the outdated CBL-003 header; paths without the proto/ prefix | CBL-003 fixed; the composition agreed (payout.events + EventMeta with an explanation); FR-5 -> section 2; finalization synchronous with the tag; label CF-2; the checks table extended; the header updated; paths complete |
| Architecture | statuses anticipate the freeze; the composition diverges; FR-5 is the wrong section; label CF-1; the CBL-003 header; the ledger.events consumer in CBL-000; reserve semantics not fixed; occurred_at/state_version; KI-01 wording | freeze+tag in this commit; the composition agreed; section 2; label CF-2; the header; CBL-000 §2 mentions canonical (the ledger.events consumer will be determined by 4.x, recorded in CF-3 if needed); proto comments; KI-01 updated; the W2 gate refined (9.1 waits for 3.2) |
| With context | the freeze composition; the CBL-CF2 commit format; statuses; label CF-2; the ledger.events key not revised; no make breaking; the spec header; CBL-003 payout.commands; FR-1 overestimates Reads 2.7 | all closed: the caveat in AGENTS §5; the key decision in FR-3; `make breaking`; the header; the row fixed; FR-1 refined |
| Security | MAJOR: IDOR in `/ledger/entries`; `POST /webhook-endpoints` without Idempotency-Key; NIT: SR not enforced (KI-P2), the reserved placeholder, Money without schema validation, the occurred_at duplicate | descriptions fixed before the freeze (owner from the token, Idempotency-Key, Money non-negative); SR - 0.5 (KI-P2); reserved - the CBL-003 §6 rule; occurred_at - a proto comment (the body is authoritative) |

## 9. Open questions

- The OpenAPI validator is not connected (KI-P1, 0.2): the draft was verified manually for
  the route composition and the security review edits.
- KI-01 (provider) does not affect CF-2: event schemas are provider-neutral
  (`provider` is a string). At the same time, task 9.1 (Reconciler) depends on 3.2 and cannot
  start before the provider is chosen: CF-2 unblocks the W2 contract input, not the tasks
  themselves that depend on KI-01.
- Until Schema Registry is connected (0.5), the `BACKWARD_TRANSITIVE` mode is not enforced
  infrastructurally: compatibility is held only by `buf breaking` (KI-P2).

## 10. Work plan and commit

Verification (FR-1, FR-2), `CBL-003` edits, AGENTS and journal updates, commit
`CBL-CF2: freeze of v1 event schemas and merchant OpenAPI`, annotated tag
`cbl-protos@cf-2`, `CBL-004` update.
