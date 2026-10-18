# CBL-17.12. PayoutCommand refund marker (cf-4.2)

| Field | Value |
| :-- | :-- |
| Epic | E17 (real provider integration) |
| Repository | `cbl-protos` |
| Size | S |
| Wave and parallel group | W5-PROVIDER |
| Depends on | cf-4.1 (17.1), 17.12 |
| Contract version | additive `cf-4.2` (new tag) |
| Status | done |

## 1. Context and goal

Provider refunds (17.12) go back through the provider that accepted the original
payment and use the provider refunds API (CoinGate refunds, BitPay refunds,
BTCPay refund pull payment), not the payout API. The payout saga already models a
refund as a payout with `refund_invoice_id`, but the `payout.commands` command
carried no refund marker, so the Connector could not tell a refund from a payout.
This task adds the marker additively and freezes the contract as `cf-4.2`.

## 2. Scope

In scope: `PayoutCommand.refund_invoice_id = 6` (additive, optional), the
regenerated Go stubs, `buf lint` and `buf breaking` against `cf-4.1`.

Out of scope: consumer behaviour (the `cbl-payout` and `cbl-connector`
increments of 17.12), other contract changes.

## 3. Contracts

`billing/events/v1/payout_events.proto`: `PayoutCommand` gains
`string refund_invoice_id = 6;` (empty means a normal payout; a non-empty value
is the source payment id of a refund). The field is optional and
BACKWARD_TRANSITIVE compatible: old producers omit it, old consumers ignore it.

## 4. Test plan

| ID | Case | Expected |
| :-- | :-- | :-- |
| C-01 | `buf lint` | no findings |
| C-02 | `buf breaking --against .git#tag=cbl-protos@cf-4.1` | no breaking changes |
| C-03 | `go build ./...` over the regenerated stubs | green |

## 5. Evidence

`buf lint` and `buf breaking` (against `cf-4.1`) are green (docker
`bufbuild/buf:1.47.2`); the stubs are regenerated with the pinned plugins
(`protoc-gen-go v1.34.2`, `protoc-gen-go-grpc v1.5.1`); `go build ./...` is
green; the diff touches only `payout_events.proto` and `payout_events.pb.go`.
The tag `cbl-protos@cf-4.2` freezes the addition.
