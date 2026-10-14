# CBL-1.7. Репозиторий cbl-protos и заморозка CF-1

| Поле | Значение |
| :-- | :-- |
| Эпик | E1, общая библиотека |
| Репозиторий | `cbl-protos` |
| Размер | M |
| Волна и параллельная группа | W0-PROTO |
| Зависит от | нет |
| Версия контрактов | CF-1 (эта задача и есть CF-1) |
| Статус | done (локально, ожидает коммита) |

## 1. Контекст и цель

Контракты в `CBL-003` описательные (KI-02). Задача создает артефакты: Protobuf-схемы gRPC и
Kafka-событий, черновик OpenAPI v1. После зеленых `buf lint` и `buf generate` контракт
объявляется замороженным (CF-1), и волна W1 стартует параллельно в четырех репозиториях.

## 2. Границы

In scope: `buf.yaml`, `buf.gen.yaml`, proto-файлы, OpenAPI-черновики, сборка stubs.
Out of scope: реализация сервисов, Schema Registry (инфраструктура 0.5), финализация OpenAPI
мерчанта (7.1), admin OpenAPI (12.1).

## 3. Функциональные требования

FR-1. Структура и go_package: модуль `github.com/cbl/cbl-protos`, пакеты
`github.com/cbl/cbl-protos/gen/go/billing/<domain>/v1;<domain>v1`.

FR-2. Файлы proto (поля - строго по `CBL-003`):

| Файл | Содержание |
| :-- | :-- |
| `proto/billing/common/v1/money.proto` | `Money{minor_units,asset,decimals}`, `Direction{DEBIT,CREDIT}` |
| `proto/billing/payment/v1/payment.proto` | `PaymentState` (CREATED, CREATING, PENDING, DETECTED, CONFIRMING, COMPLETED, EXPIRED, UNDERPAID, OVERPAID, FAILED, CANCELLED), `Payment`, `TimelineItem`, RPC `CreatePayment`, `GetPayment`, `ListPayments`, `CancelPayment` |
| `proto/billing/quote/v1/quote.proto` | RPC `LockQuote`, `GetRate`; `LockQuoteRequest{idempotency_key? нет - по CBL-003: from_asset,to_asset,amount,ttl}`, ответ `{quote_id,rate,locked_until}` |
| `proto/billing/ledger/v1/ledger.proto` | `EntryInput`, `Entry`, `AccountBalance`; RPC `PostTransaction`, `Reserve`, `CommitReserve`, `ReleaseReserve`, `GetBalance`, `ListEntries` |
| `proto/billing/connector/v1/connector.proto` | `CanonicalPaymentStatus`, `Capability`; RPC `CreateInvoice`, `GetGatewayPayment`, `ListGatewayPayments`, `GetCapabilities` |
| `proto/billing/payout/v1/payout.proto` | `PayoutState`, `Payout`; RPC `CreatePayout`, `ApprovePayout`, `GetPayout` |
| `proto/billing/reporting/v1/reporting.proto` | RPC `Query{report_type,merchant_id,period,group_by,cursor,limit}` -> строки отчета `map<string,string>` |
| `proto/billing/events/v1/envelope.proto` | `EventMeta{event_id,event_type,occurred_at,producer,traceparent,schema_version}` (заголовки Kafka) |
| `proto/billing/events/v1/gateway_events.proto` | `GatewayWebhookRaw`, `CanonicalPaymentEvent` (типы DETECTED, CONFIRMED, FAILED, EXPIRED, UNDERPAID, OVERPAID), `GatewayPayoutEvent` (SUBMITTED, CONFIRMED, FAILED) |
| `proto/billing/events/v1/payment_events.proto` | `PaymentEvent` (типы CREATED, DETECTED, CONFIRMING, COMPLETED, EXPIRED, UNDERPAID, OVERPAID, FAILED; поля merchant_id, expected_amount, received_amount, fee, quote_id, rate, txid, state_version) |
| `proto/billing/events/v1/ledger_events.proto` | `LedgerEvent` (типы TRANSACTION_POSTED, RESERVE_CREATED, RESERVE_COMMITTED, RESERVE_RELEASED; entries[], ref_type, ref_id) |
| `proto/billing/events/v1/payout_events.proto` | `PayoutCommand` (SendPayout), `PayoutEvent` (REQUESTED, APPROVED, COMPLETED, FAILED, COMPENSATED) |
| `proto/billing/events/v1/recon_events.proto` | `ReconAlert` (типы MISSING_IN_LEDGER, AMOUNT_MISMATCH, STATUS_MISMATCH; severity INFO, WARNING, CRITICAL) |

FR-3. Все мутирующие RPC принимают `idempotency_key` (по `CBL-003`, раздел 5); денежные поля -
только `billing.common.v1.Money`; время - `google.protobuf.Timestamp`, идентификаторы - string
(UUIDv7 на уровне сервисов).

FR-4. `buf.yaml` (v2): модуль `proto`, lint `STANDARD`, breaking `FILE`; `buf.gen.yaml` (v2):
`protoc-gen-go` и `protoc-gen-go-grpc`, вывод в `gen/go`, `paths=source_relative`.

FR-5. OpenAPI-черновики: `openapi/billing.v1.yaml` (пути из `CBL-003`, раздел 2, с
`Idempotency-Key`, схемой `Problem` RFC 9457 и bearer-аутентификацией) и
`openapi/admin.v1.yaml` (пути из `CBL-003`, раздел 3).

FR-6. Корневой `go.mod` модуля `github.com/cbl/cbl-protos`; `go build ./...` компилирует
сгенерированные стабы.

## 4. Контракты

Задача создает контракты; после приемки фиксируется CF-1. Изменения после заморозки - только
обратно совместимые (`BACKWARD_TRANSITIVE`), см. `CBL-003`, раздел 8.

## 5. Модель данных и миграции

Не применимо.

## 6. Конечные автоматы

Enums состояний из FR-2 соответствуют автоматам `CBL-000` (раздел 5): платеж
`created -> pending -> detected -> confirming -> completed / expired / underpaid / overpaid`;
выплата `requested -> reserved -> sent_to_gateway -> completed / failed -> compensated`.

## 7. Edge cases

- Конфликт имен пакетов: домены `events` и `common` отделены от сервисных пакетов.
- `buf` CLI не установлен: lint запускается через Docker (`docker run --rm -v <repo>:/w -w /w bufbuild/buf:1.47.2 lint`) или через `go run github.com/bufbuild/buf/cmd/buf@v1.47.2`.
- Валидатор OpenAPI недоступен локально: файлы проверяются вручную, финализация в 7.1 (KI).

## 8. Нефункциональные требования

`buf lint` без замечаний; генерация стабов воспроизводима; никаких внешних proto-зависимостей,
кроме well-known types.

## 9. План тестов

| Уровень | Сценарии |
| :-- | :-- |
| Smoke | `go run buf lint`; `go run buf generate`; `go build ./...` |
| Unit | не применимо (генерируемый код) |
| Integration | не применимо |
| E2E | не применимо |
| Load | не применимо |

## 10. План работ и коммит

Коммит `CBL-1.7: proto-контракты и заморозка CF-1` после приемки; в текущей сессии без коммитов.

## 11. Открытые вопросы

- Валидация OpenAPI (spectral/redocly) будет добавлена в CI задачей 0.2.
