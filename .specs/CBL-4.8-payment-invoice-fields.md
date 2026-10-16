# CBL-4.8. Payment contract fields for the merchant invoice

| Field | Value |
| :-- | :-- |
| Epic | E4 (Payment), the contract part of task 4.8 |
| Repository | `cbl-protos` |
| Size | S (additive fields) |
| Wave | W3-BILLING (prerequisite of 7.4, FR-8 of spec CBL-7.1) |
| Depends on | CF-3, tag `cbl-protos@cf-3.2` |
| Contract version | `cbl-protos@cf-3.3` (base) and `cbl-protos@cf-3.4` (review fields) |
| Status | done (implementation - `cbl-payment`, see `.specs/CBL-4.8-invoice-read-api.md`) |

## 1. Context and goal

The 7.1 review (an architecture BLOCKER) showed: `Payment`/`CreatePaymentResponse` did
not carry the locked rate, its lock expiry and network data (`txid`,
`confirmations`), and `ListPayments` had no period support - the merchant REST invoice
(`quote{rate, locked_until}`, `txid`, `confirmations`, list filters) had no
source, and the creation 201 lacked `pay_amount`. Task 4.8 (cbl-payment) implements the
read; the contract was released in two additive steps: `cf-3.3` (rate/locked_until,
period) and `cf-3.4` (txid/confirmations, `expected_amount` in the creation response) -
after four review passes. The full spec is
`cbl-payment/.specs/CBL-4.8-invoice-read-api.md`.

## 2. Composition

| Artifact | Change |
| :-- | :-- |
| `Payment` | additive `rate = 13`, `locked_until = 14` (cf-3.3), `txid = 15`, `confirmations = 16` (cf-3.4) |
| `CreatePaymentResponse` | additive `rate = 6`, `locked_until = 7` (cf-3.3), `expected_amount = 8` (cf-3.4) |
| `ListPaymentsRequest` | additive `created_from = 5`, `created_to = 6`; cursor - the opaque `next_cursor` (cf-3.3) |

## 3. Requirements

FR-1. `rate` - a decimal string of the locked rate; `locked_until` - the lock
expiry; `txid`/`confirmations` - network data (filled by consumer 4.3);
the creation `expected_amount` - the assembled amount and pay asset (the source
of `pay_amount`/`pay_asset` in the 201 response); `pay_amount`/`pay_asset` of reads are
derived from `expected_amount`; `fee` is not exposed (KI-payment-06).

FR-2. `created_from`/`created_to` - a period by `created_at` (from inclusive,
to exclusive); `merchant_id` is mandatory; the cursor - the opaque value
`next_cursor` of the previous response (order `created_at DESC, id DESC`).

FR-3. Compatibility: `buf lint`, `buf breaking` against the previous tags,
`buf generate`, `go build ./...` are green; existing consumers are
unaffected (additivity).

## 4. Checks

| Check | Command | Result |
| :-- | :-- | :-- |
| lint/breaking | `buf lint`; `buf breaking --against ".git#tag=cbl-protos@cf-3.3"` | clean |
| generate/build | `buf generate`; `go build ./...` | clean (stub diff: txid/confirmations/expected_amount) |
| em dash | `scripts/check-no-emdash.ps1` | clean |

## 5. Plan and commit

Commits `CBL-4.8` (cf-3.3, 96002bb) and `CBL-4.8` (cf-3.4, the review fields);
implementation - `cbl-payment` `CBL-4.8`.
