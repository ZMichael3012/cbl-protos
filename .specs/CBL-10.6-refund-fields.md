# CBL-10.6. Контракты возвратов: refund_invoice_id

| Поле | Значение |
| :-- | :-- |
| Репозиторий | `cbl-protos` |
| Задача плана | CBL-10.6 (часть: контрактный вход возвратов) |
| Статус | done |

## Изменения (аддитивные, BACKWARD_TRANSITIVE)

Возврат - частный случай выплаты (`CBL-000`, раздел 5; `CBL-003`, раздел 2:
`POST /v1/refunds` с `invoice_id`, `amount`, `destination_address`). Для
отличия возврата от обычной выплаты добавлены необязательные поля:

- `billing.payout.v1.CreatePayoutRequest.refund_invoice_id = 7` - исходный счет
  возврата; непустое значение создает возврат тем же путем саги;
- `billing.payout.v1.Payout.refund_invoice_id = 12` - маркер возврата в ответах
  CreatePayout/GetPayout;
- `billing.events.v1.PayoutEvent.refund_invoice_id = 11` - маркер для
  Notification (8.x публикует `refund.*` вместо `payout.*`) и ClickHouse.

Проверка суммы возврата (не больше возвращаемого остатка счета) - обязанность
Billing API (7.x): Payout не владеет данными о счетах (database-per-service).

## Проверки

- `buf lint` - зеленый; `buf breaking --against cbl-protos@cf-2` - зеленый
  (добавление полей не ломает совместимость);
- генерация `buf generate` без диффа вне payout-файлов; `go build ./...` -
  зеленый.

## Потребление

`cbl-payout` (миграция `refund_invoice_id`, валидация UUID, эмиссия в
`payout.events`) - задача CBL-10.6; `cbl-billing-api` 7.x - REST `/v1/refunds`
и проверка остатка; `cbl-notification` 8.x - типы `refund.completed`/
`refund.failed` по маркеру.
