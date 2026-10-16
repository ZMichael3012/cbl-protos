# CBL-10.6. Refund contracts: refund_invoice_id

| Field | Value |
| :-- | :-- |
| Repository | `cbl-protos` |
| Plan task | CBL-10.6 (part: the contract input of refunds) |
| Status | done |

## Changes (additive, BACKWARD_TRANSITIVE)

A refund is a special case of a payout (`CBL-000`, section 5; `CBL-003`, section 2:
`POST /v1/refunds` with `invoice_id`, `amount`, `destination_address`). To
distinguish a refund from a regular payout, optional fields were added:

- `billing.payout.v1.CreatePayoutRequest.refund_invoice_id = 7` - the source invoice
  of the refund; a non-empty value creates a refund through the same saga path;
- `billing.payout.v1.Payout.refund_invoice_id = 12` - the refund marker in the responses
  CreatePayout/GetPayout;
- `billing.events.v1.PayoutEvent.refund_invoice_id = 11` - the marker for
  Notification (8.x publishes `refund.*` instead of `payout.*`) and ClickHouse.

Verification of the refund amount (not more than the refundable account balance) is the
responsibility of Billing API (7.x): Payout does not own account data (database-per-service).

## Checks

- `buf lint` - green; `buf breaking --against cbl-protos@cf-2` - green
  (adding fields does not break compatibility);
- generation `buf generate` without a diff outside payout files; `go build ./...` -
  green.

## Consumption

`cbl-payout` (the `refund_invoice_id` migration, UUID validation, emission to
`payout.events`) - task CBL-10.6; `cbl-billing-api` 7.x - REST `/v1/refunds`
and the balance check; `cbl-notification` 8.x - the `refund.completed`/
`refund.failed` types by the marker.
