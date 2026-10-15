# CBL-4.8. Контрактные поля Payment для счета мерчанта

| Поле | Значение |
| :-- | :-- |
| Эпик | E4 (Payment), контрактная часть задачи 4.8 |
| Репозиторий | `cbl-protos` |
| Размер | S (аддитивные поля) |
| Волна | W3-BILLING (предпосылка 7.4, FR-8 спеки CBL-7.1) |
| Зависит от | CF-3, тег `cbl-protos@cf-3.2` |
| Версия контрактов | `cbl-protos@cf-3.3` (этот тег) |
| Статус | in progress |

## 1. Контекст и цель

Ревью 7.1 (BLOCKER архитектуры) показало: `Payment`/`CreatePaymentResponse` не
несут зафиксированный курс и срок его фиксации, а `ListPayments` не умеет
период, - REST-инвойс мерчанта (`quote{rate, locked_until}`, фильтры списка) не
имел источника. Задача 4.8 (cbl-payment) реализует чтение; этот коммит
фиксирует аддитивные поля под тегом `cbl-protos@cf-3.3`. Полная спека -
`cbl-payment/.specs/CBL-4.8-invoice-read-api.md`.

## 2. Состав

| Артефакт | Изменение |
| :-- | :-- |
| `Payment` | аддитивные `rate = 13`, `locked_until = 14` |
| `CreatePaymentResponse` | аддитивные `rate = 6`, `locked_until = 7` |
| `ListPaymentsRequest` | аддитивные `created_from = 5`, `created_to = 6` |

## 3. Требования

FR-1. `rate` - десятичная строка зафиксированного курса; `locked_until` - срок
фиксации котировки; заполняются при создании платежа (4.2, KI-01);
`pay_amount`/`pay_asset` REST-инвойса выводятся из `expected_amount` (Money с
активом), новых полей не вводят.

FR-2. `created_from`/`created_to` - период по `created_at` (from включительно,
to исключительно); `merchant_id` обязателен; курсор - `payment_id` последней
записи (порядок `created_at DESC, id DESC`, совместим с существующим курсором
2.x).

FR-3. Совместимость: `buf lint`, `buf breaking` против `cbl-protos@cf-3.2`,
`buf generate`, `go build ./...` - зеленые; существующие потребители не
затронуты (аддитивность).

## 4. Проверки

| Проверка | Команда | Результат |
| :-- | :-- | :-- |
| lint/breaking | `buf lint`; `buf breaking --against ".git#tag=cbl-protos@cf-3.2"` | clean |
| generate/build | `buf generate`; `go build ./...` | clean (дифф стабов: новые поля) |
| эм-деш | `scripts/check-no-emdash.ps1` | clean |

## 5. План и коммит

Коммит `CBL-4.8: аддитивные поля Payment для счета мерчанта`, аннотированный
тег `cbl-protos@cf-3.3`; затем реализация в `cbl-payment`.
