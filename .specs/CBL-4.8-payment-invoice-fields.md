# CBL-4.8. Контрактные поля Payment для счета мерчанта

| Поле | Значение |
| :-- | :-- |
| Эпик | E4 (Payment), контрактная часть задачи 4.8 |
| Репозиторий | `cbl-protos` |
| Размер | S (аддитивные поля) |
| Волна | W3-BILLING (предпосылка 7.4, FR-8 спеки CBL-7.1) |
| Зависит от | CF-3, тег `cbl-protos@cf-3.2` |
| Версия контрактов | `cbl-protos@cf-3.3` (базовый) и `cbl-protos@cf-3.4` (поля по ревью) |
| Статус | done (реализация - `cbl-payment`, см. `.specs/CBL-4.8-invoice-read-api.md`) |

## 1. Контекст и цель

Ревью 7.1 (BLOCKER архитектуры) показало: `Payment`/`CreatePaymentResponse` не
несли зафиксированный курс, срок его фиксации и данные сети (`txid`,
`confirmations`), а `ListPayments` не умел период, - REST-инвойс мерчанта
(`quote{rate, locked_until}`, `txid`, `confirmations`, фильтры списка) не имел
источника, а 201 создания - `pay_amount`. Задача 4.8 (cbl-payment) реализует
чтение; контракт выпущен в два аддитивных шага: `cf-3.3` (rate/locked_until,
период) и `cf-3.4` (txid/confirmations, `expected_amount` в ответе создания) -
после четырех ревью-проходов. Полная спека -
`cbl-payment/.specs/CBL-4.8-invoice-read-api.md`.

## 2. Состав

| Артефакт | Изменение |
| :-- | :-- |
| `Payment` | аддитивные `rate = 13`, `locked_until = 14` (cf-3.3), `txid = 15`, `confirmations = 16` (cf-3.4) |
| `CreatePaymentResponse` | аддитивные `rate = 6`, `locked_until = 7` (cf-3.3), `expected_amount = 8` (cf-3.4) |
| `ListPaymentsRequest` | аддитивные `created_from = 5`, `created_to = 6`; курсор - непрозрачный `next_cursor` (cf-3.3) |

## 3. Требования

FR-1. `rate` - десятичная строка зафиксированного курса; `locked_until` - срок
фиксации; `txid`/`confirmations` - данные сети (заполняет consumer 4.3);
`expected_amount` создания - смонтированные сумма и актив оплаты (источник
`pay_amount`/`pay_asset` ответа 201); `pay_amount`/`pay_asset` чтений выводятся
из `expected_amount`; `fee` не выносится (KI-payment-06).

FR-2. `created_from`/`created_to` - период по `created_at` (from включительно,
to исключительно); `merchant_id` обязателен; курсор - непрозрачное значение
`next_cursor` предыдущего ответа (порядок `created_at DESC, id DESC`).

FR-3. Совместимость: `buf lint`, `buf breaking` против предыдущих тегов,
`buf generate`, `go build ./...` - зеленые; существующие потребители не
затронуты (аддитивность).

## 4. Проверки

| Проверка | Команда | Результат |
| :-- | :-- | :-- |
| lint/breaking | `buf lint`; `buf breaking --against ".git#tag=cbl-protos@cf-3.3"` | clean |
| generate/build | `buf generate`; `go build ./...` | clean (дифф стабов: txid/confirmations/expected_amount) |
| эм-деш | `scripts/check-no-emdash.ps1` | clean |

## 5. План и коммит

Коммиты `CBL-4.8` (cf-3.3, 96002bb) и `CBL-4.8` (cf-3.4, поля по ревью);
реализация - `cbl-payment` `CBL-4.8`.
