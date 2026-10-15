# CBL-2.10. Контрактные поля Ledger для BFF-выписок

| Поле | Значение |
| :-- | :-- |
| Эпик | E2 (Ledger), контрактная часть задачи 2.10 |
| Репозиторий | `cbl-protos` |
| Размер | S (аддитивные поля и RPC) |
| Волна | W3-BILLING (предпосылка 7.4, FR-8 спеки 7.1) |
| Зависит от | CF-3, тег `cbl-protos@cf-3.1` |
| Версия контрактов | `cbl-protos@cf-3.2` (этот тег) |
| Статус | in progress |

## 1. Контекст и цель

Ревью 7.1 показало: `/balances` и `/ledger/entries` (7.4) не реализуемы на
текущем Ledger gRPC - нет списка счетов мерчанта с балансами и ownership, нет
фильтра по мерчанту и периоду в выписке. Задача 2.10 (cbl-ledger) реализует
серверную часть; этот коммит фиксирует аддитивные поля контракта под тегом
`cbl-protos@cf-3.2`. Полная спека - `cbl-ledger/.specs/CBL-2.10-ledger-bff-extension.md`.

## 2. Состав

| Артефакт | Изменение |
| :-- | :-- |
| `ListEntriesRequest` | аддитивные `merchant_ref = 4`, `period_from = 5`, `period_to = 6` |
| `ListMerchantAccountsRequest/Response`, `MerchantAccount` | новые сообщения |
| `LedgerService.ListMerchantAccounts` | новый RPC |

## 3. Требования

FR-1. `ListEntries`: заполняется `account_id` или `merchant_ref` (хотя бы одно);
`merchant_ref` фильтрует счета типа `merchant`/`payout` мерчанта, `account_id`
вне набора - ненайденный (fail-closed, без раскрытия существования); период
фильтрует `created_at` транзакции (`from` включительно, `to` исключительно).

FR-2. `ListMerchantAccounts`: обязательный `merchant_ref`; возвращает счета
мерчанта (merchant и payout) с балансами, счета без проводок - нулевые, порядок
по `(asset, type)`.

FR-3. Совместимость: `buf lint`, `buf breaking` против `cbl-protos@cf-3.1`,
`buf generate`, `go build ./...` - зеленые; существующие потребители не
затрагиваются (аддитивность).

## 4. Проверки

| Проверка | Команда | Результат |
| :-- | :-- | :-- |
| lint/breaking | `buf lint`; `buf breaking --against ".git#tag=cbl-protos@cf-3.1"` | clean |
| generate/build | `buf generate`; `go build ./...` | clean (дифф стабов: новые поля и RPC) |
| эм-деш | `scripts/check-no-emdash.ps1` | clean |

## 5. План и коммит

Коммит `CBL-2.10: аддитивные поля Ledger для BFF-выписок`, аннотированный тег
`cbl-protos@cf-3.2`; затем реализация в `cbl-ledger`.
