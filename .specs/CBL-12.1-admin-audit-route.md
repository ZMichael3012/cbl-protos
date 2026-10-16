# CBL-12.1. admin.v1.yaml finalization: audit-log route

| Field | Value |
| :-- | :-- |
| Epic | E12 (task CBL-12.1, the `cbl-backoffice` part of the task) |
| Repository | `cbl-protos` |
| Size | S |
| Wave and parallel group | W4-ADMIN |
| Depends on | CF-3 (`openapi/admin.v1.yaml` draft frozen by route composition) |
| Contract version | `cbl-protos@cf-3.5` (incremental tag over `cf-3.4`) |
| Status | done |

## 1. Context and goal

CF-3 froze the Admin API v1 route composition; the OpenAPI draft states that the
document "is finalized by task 12.1". Task 12.1 (cbl-backoffice: Admin API, RBAC,
action audit log) needs one additional read route over the audit journal
(`GET /admin/v1/audit-log`) so operators and auditors can inspect the append-only
action log. This spec records that additive finalization; the main 12.1 spec lives in
`cbl-backoffice` (`.specs/CBL-12.1-admin-api-rbac-audit.md`).

## 2. Requirements

FR-1. `openapi/admin.v1.yaml` 1.0.0 (draft suffix removed): an added `GET /audit-log`
route (tag `audit`) with filters `operator_id`, `action`, `period_from`/`period_to`
(RFC 3339), shared `Limit`/`Cursor` parameters, `items`/`next_cursor` response, 400/401/403
problem responses; the new `AuditRecord` schema (`id`, `occurred_at`, `operator_id`,
`action`, `resource_type`, `resource_id`, `request_id`, nullable `before`/`after`).

FR-2. Additive only: no existing path, operationId, schema field or response semantic
changes; the description stops saying "finalized by task 12.1".

FR-3. No proto changes: `gen/` and the tags `cf-3.1..cf-3.4` stay untouched; the new
incremental tag is `cbl-protos@cf-3.5`.

## 3. Checks

YAML parses (python yaml), the diff touches only the version, the info description,
the new tag, the new path and the new schema; `buf lint`/`buf breaking`/`go build` are
unaffected (OpenAPI is not covered by buf; the OpenAPI validator is KI-P1).

## 4. Work plan

Edit the YAML -> spec -> commit `CBL-12.1: admin.v1.yaml finalization: audit-log route`
-> tag `cbl-protos@cf-3.5`; `CBL-003` section 3 gets the audit-log bullet (umbrella,
committed with the umbrella on instruction).
