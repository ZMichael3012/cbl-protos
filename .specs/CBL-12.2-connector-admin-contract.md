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
field description that already says 422.

FR-3. Additive only: new messages/RPCs/service, no field renumbering, no change to
existing messages; buf lint/breaking against `cf-3.6` green; generation with the pinned
plugins; the tag is `cbl-protos@cf-3.7`.

## 3. Work plan

Proto + OpenAPI edits -> spec -> buf lint/breaking, generate, build -> commit
`CBL-12.2: connector admin service contract` -> tag `cbl-protos@cf-3.7`.
Consumers: `cbl-connector` (server), `cbl-backoffice` (client); `CBL-003` section 3/7
notes are updated with the umbrella on instruction.
