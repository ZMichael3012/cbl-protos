# CBL-12.2. ConnectorAdminService: routing configuration and provider disable

| Field | Value |
| :-- | :-- |
| Epic | E12 (task CBL-12.2, the `cbl-protos` part) |
| Repository | `cbl-protos` |
| Size | S |
| Wave and parallel group | W4-ADMIN |
| Depends on | CF-3 (Admin API v1 composition: routing-rules revision, RBAC subject from the token) |
| Contract version | `cbl-protos@cf-3.7` (incremental tag over `cf-3.6`) |
| Status | done |

## 1. Context and goal

KI-connector-08: the admin path for routing rules (`GET/PUT /admin/v1/routing-rules`,
`POST /admin/v1/providers/{id}/disable`) needs a gRPC contract in the Connector (the
source of truth of routing rules is PostgreSQL connector, 3.4 ADR). This spec records
the contract; the implementations land in `cbl-connector` and `cbl-backoffice` with the
same task number.

## 2. Requirements

FR-1. `billing/connector/v1` gains the additive `ConnectorAdminService` (a separate
service, so the existing `ConnectorService` server interface and its fakes stay
source-compatible):

- `GetRoutingRules` -> the current `revision` and all rules (`RoutingRuleAdmin`:
  provider, operation, networks, assets, priority, weight, min/max amount in minor
  units, enabled);
- `PutRoutingRules(revision, items)` -> full replacement with optimistic concurrency:
  a revision mismatch is `FailedPrecondition`; invalid rules are `InvalidArgument`;
  the response carries the new revision and the stored rules;
- `DisableProvider(provider, reason)` -> manual disable; an unknown provider is
  `NotFound`, a repeated disable returns `already_disabled = true` without a state
  change; the router excludes disabled providers from selection.

FR-2. `openapi/admin.v1.yaml` 1.0.0: the `/payouts/{id}/reject` responses gain the
missing `'422'` (a missing/empty reject reason is business validation), matching the
field description that already says 422; the `RoutingRule` schema is finalized to the
form of the source of truth (`RoutingRuleAdmin` in the connector): `operation`,
`networks[]`/`assets[]` arrays (empty means any), `weight` minimum 1, and
`min_amount`/`max_amount` as integer minor units of the asset (0 means no bound) instead
of decimal `Money` strings, which were not convertible without a decimals source.

FR-3. Additive only in the proto (`ConnectorAdminService` is a new service; no field
renumbering, no change to existing messages); the `RoutingRule` schema revision in the
OpenAPI is an in-place finalization recorded by this spec (the back-office is the only
implementer; the gateway proxies opaque bodies). buf lint/breaking against `cf-3.6`
green; generation with the pinned plugins; the tags are `cbl-protos@cf-3.7` (proto) and
`cbl-protos@cf-3.8` (the RoutingRule finalization).

FR-4. `billing/payout/v1`: the additive `Payout.approved_by` (field 13, repeated
string, vote order) - the back-office payout decision response (PayoutAdmin.approved_by
of the frozen admin API) needs the approval progress; the vote journal itself stays
internal to `cbl-payout`. buf lint/breaking against `cf-3.8` green; the tag is
`cbl-protos@cf-3.9`.

## 3. Work plan

Proto + OpenAPI edits -> spec -> buf lint/breaking, generate, build -> commit
`CBL-12.2: connector admin service contract` (tag `cbl-protos@cf-3.7`), the RoutingRule finalization (tag cf-3.8), the additive Payout.approved_by (tag cf-3.9) and the review follow-up adding the 503 ServiceUnavailable responses to the business routes (tag cf-3.10).
Consumers: `cbl-connector` (server), `cbl-backoffice` (client); `CBL-003` section 3/7
notes are updated with the umbrella on instruction.
