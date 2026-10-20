# CBL-17.21 - Asset and network catalog artifact (cbl-protos)

Task: CBL-17.21 in `CBL-004` (E17, W5-SCALE), depends on 17.1 (done).
Repositories: `cbl-protos` (this spec: the versioned artifact, the JSON Schema and the
contract-side validation) and `cbl-kit` (the runtime parser, validation and lookups, spec in
that repository). Status: in progress.

## 1. Problem and scope

The platform needs one contract artifact for crypto assets and their networks: canonical
asset codes and decimals (ledger money), canonical network codes (routing and capability
matching) and the provider codes used on the wire (CoinGate currency plus platform,
NOWPayments currency code, BitPay currency, BTCPay asset and payment method). In scope: the
JSON artifact, the JSON Schema, the contract-side validation test and the release tag. Out
of scope: runtime parsing (cbl-kit), adapter wiring (consumer tasks), fiat currencies.

## 2. Artifact design

`catalog/assets.json`:

```json
{
  "schema_version": 1,
  "catalog_version": 1,
  "updated_at": "2026-10-20",
  "assets": [
    {
      "code": "USDT",
      "name": "Tether USD",
      "decimals": 6,
      "status": "active",
      "networks": [
        {
          "code": "usdt_trc20",
          "name": "Tether on TRON (TRC20)",
          "kind": "token",
          "chain": "tron",
          "status": "active",
          "provider_codes": {
            "coingate": {"asset": "USDT_TRC20"},
            "nowpayments": {"asset": "usdttrc20"},
            "bitpay": {"asset": "USDT"},
            "btcpay": {"asset": "USDT"}
          }
        }
      ]
    }
  ]
}
```

Initial content (catalog version 1, 20.10.2026): BTC (bitcoin, lightning), ETH (ethereum),
USDT (usdt_trc20, usdt_erc20), USDC (usdc_erc20). Provider codes come from the adapter
tests already merged in 17.6 - 17.9: CoinGate and NOWPayments use one combined wire code
(`USDT_TRC20`, `usdttrc20`), BitPay and BTCPay take a chain-agnostic currency code
(`USDT`), so the same wire code covers TRON and Ethereum. Only mappings exercised by the
adapters are recorded; the catalog is a contract, not a wish list.

Fields:

| Field | Meaning |
| :-- | :-- |
| `schema_version` | artifact schema, starts at 1; a breaking shape change bumps it |
| `catalog_version` | data revision, monotonically increasing |
| `updated_at` | date of the revision |
| `assets[].code` | canonical uppercase asset code (ledger and reporting) |
| `assets[].decimals` | canonical minor-unit scale (money invariant) |
| `assets[].status` | `active` or `deprecated`; deprecated stays resolvable |
| `networks[].code` | canonical lowercase network code (routing, capabilities) |
| `networks[].kind` | `native`, `token` or `layer` |
| `networks[].chain` | chain identifier shared by native and token networks |
| `provider_codes` | provider id to the provider wire codes (`asset` required, `network` optional) |

## 3. Schema and validation

`catalog/assets.schema.json` is JSON Schema draft 2020-12, `additionalProperties: false` at
every level, with enums and patterns mirroring the rules below. It is documentation and a
tooling input; `catalog/catalog_test.go` enforces the cross-field rules that JSON Schema
cannot express (uniqueness, provider coverage, active network coverage).

Contract-side validation rules (mirrored by the kit parser, the runtime gate):

C1. `schema_version` = 1, `catalog_version` >= 1, `updated_at` is a date.
C2. Asset codes `^[A-Z0-9]{2,10}$`, unique, decimals 0..18, non-empty names, status enum.
C3. Network codes `^[a-z0-9_]{2,32}$`, globally unique, kind/chain/status valid.
C4. Known providers only; provider codes match `^[A-Za-z0-9_.:-]{1,32}$` (network optional);
    one provider asset code maps to exactly one canonical asset; a non-empty provider
    network code is unique per asset code; an active network has a provider code.
C5. Every known provider appears at least once; at least one active asset.
C6. No surrounding whitespace in codes and names.
C7. Strict JSON: unknown fields fail.

## 4. Consumption and release

Services depend on `cbl-protos` already; the artifact is copied or mounted from this
repository at build/deploy time and parsed by `cbl-kit/catalog`. The release is tagged
`cbl-protos@catalog-1` after the commit. A catalog data change is additive when it adds
assets, networks or provider codes; removing or changing a code or decimals is a breaking
change and requires a new `catalog_version` plus a migration note in `CBL-003`.

## 5. Test plan

| ID | Case | Expectation |
| :-- | :-- | :-- |
| P1 | artifact is valid strict JSON and matches the schema's top-level required fields | pass |
| P2 | schema file is valid JSON, draft 2020-12, closes additional properties | pass |
| P3 | uniqueness and pattern rules (C2, C3) over the artifact | pass |
| P4 | provider rules and coverage (C4, C5) over the artifact | pass |
| P5 | every active network has at least one provider code and every active asset an active network | pass |
| P6 | no em dash and no tabs in the artifact | pass |

Integration and e2e: not applicable here; the artifact is consumed by the kit parser test
K19 in `cbl-kit` and by service stands only after adapter wiring. The artifact validation
test runs in `go test ./...` in this repository, which is the CI gate for contract changes.

## 6. Red evidence

Recorded 20.10.2026: `go test ./catalog` failed with `read the schema: open
assets.schema.json: The system cannot find the file specified` and `read the artifact:
open assets.json: The system cannot find the file specified`. Both files were added after
the red run.

## 7. Results and review closure (20.10.2026)

- Artifact v1: 4 assets, 6 networks, provider wire codes for all four providers; the
  validation matrix P1 - P6 green; `go build ./...`, `go test ./...`, `go vet ./...`,
  golangci-lint 0 issues, em dash check green; tag `cbl-protos@catalog-1`.
- Review pass 1 (no-context): the first draft listed CoinGate `USDT` + `TRON` and BTCPay
  `USDT` + `TRON` provider codes, but the adapters send one combined CoinGate pay
  currency (`USDT_TRC20`) and no platform for BitPay/BTCPay. Closed: the artifact and
  the spec now use the adapter wire codes; USDT on TRON and Ethereum share `USDT` for
  the chain-agnostic providers.
- Review pass 2 (architecture, blocking): the triple uniqueness rule contradicted that
  chain-agnostic reuse. Closed: the rule checks that one provider asset code maps to one
  canonical asset, and that a non-empty provider network code is unique; the kit mirrors
  it (K13/K22/K24).
- Review pass 3 (with-context): the catalog is a separate contract line, not part of
  CF-1 - CF-4; the release tag and the AGENTS/README references are recorded; the
  consumers parse it with `cbl-kit/catalog`.
- Review pass 4 (security): the file is a trusted checked-in artifact, no secrets; the
  runtime parser owns the 1 MiB input bound; the contract test rejects padded or
  duplicate codes, so a bad mapping fails before it reaches a service.
- NIT: the test duplicates the kit invariants by design (the contracts repository stays
  dependency-free); the kit cross-check parses this artifact.
