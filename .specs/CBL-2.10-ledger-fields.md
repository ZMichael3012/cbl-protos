# CBL-2.10. Ledger contract fields for BFF statements

| Field | Value |
| :-- | :-- |
| Epic | E2 (Ledger), the contract part of task 2.10 |
| Repository | `cbl-protos` |
| Size | S (additive fields and RPC) |
| Wave | W3-BILLING (prerequisite of 7.4, FR-8 of spec 7.1) |
| Depends on | CF-3, tag `cbl-protos@cf-3.1` |
| Contract version | `cbl-protos@cf-3.2` (this tag) |
| Status | done (implementation - `cbl-ledger` fbe0c45) |

## 1. Context and goal

The 7.1 review showed: `/balances` and `/ledger/entries` (7.4) are not implementable on
the current Ledger gRPC - there is no list of merchant accounts with balances and
ownership and no filter by merchant and period in the statement. Task 2.10 (cbl-ledger)
implements the server part; this commit fixes the additive contract fields under the tag
`cbl-protos@cf-3.2`. The full spec is
`cbl-ledger/.specs/CBL-2.10-ledger-bff-extension.md`.

## 2. Composition

| Artifact | Change |
| :-- | :-- |
| `ListEntriesRequest` | additive `merchant_ref = 4`, `period_from = 5`, `period_to = 6` |
| `ListMerchantAccountsRequest/Response`, `MerchantAccount` | new messages |
| `LedgerService.ListMerchantAccounts` | new RPC |

## 3. Requirements

FR-1. `ListEntries`: `account_id` or `merchant_ref` is provided (at least one);
`merchant_ref` filters the accounts of type `merchant`/`payout` of the merchant; an
`account_id` outside the set is not found (fail-closed, without revealing its existence);
the period filters the transaction `created_at` (`from` inclusive, `to` exclusive).

FR-2. `ListMerchantAccounts`: mandatory `merchant_ref`; returns the merchant's accounts
(merchant and payout) with balances; accounts without postings are zero; order
by `(asset, type)`.

FR-3. Compatibility: `buf lint`, `buf breaking` against `cbl-protos@cf-3.1`,
`buf generate`, `go build ./...` are green; existing consumers are unaffected
(additivity).

## 4. Checks

| Check | Command | Result |
| :-- | :-- | :-- |
| lint/breaking | `buf lint`; `buf breaking --against ".git#tag=cbl-protos@cf-3.1"` | clean |
| generate/build | `buf generate`; `go build ./...` | clean (stub diff: new fields and RPC) |
| em dash | `scripts/check-no-emdash.ps1` | clean |

## 5. Plan and commit

Commit `CBL-2.10: additive Ledger fields for BFF statements`, annotated tag
`cbl-protos@cf-3.2`; then the implementation in `cbl-ledger`.
