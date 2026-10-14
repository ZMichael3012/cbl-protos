# CBL-CF2. Заморозка схем событий v1 и OpenAPI мерчанта

| Поле | Значение |
| :-- | :-- |
| Эпик | Контрактная точка CF-2 (CBL-004, раздел 3) |
| Репозиторий | `cbl-protos` |
| Размер | S (заморозка, схемы созданы в 1.7) |
| Волна | конец W1, до старта W2 |
| Зависит от | 1.7 (CF-1), реализаций W1 (2.7, 3.5, 5.1, 6.3) |
| Версия контрактов | `cbl-protos@cf-2` (этот тег) |
| Статус | done, ревью закрыто |

## 1. Контекст и цель

CF-1 заморозил gRPC v1 и события gateway (`gateway.webhooks.raw`,
`gateway.events.canonical`). Задачи W1 реализовали продюсеров и потребителей и выявили
фактические схемы. CF-2 фиксирует схемы событий фазы 2 (`payment.events`, `ledger.events`,
`payout.commands`, `gateway.payout.events`) и черновик OpenAPI мерчанта v1, после чего
волна W2 (Payment, Payout, Reconciler) стартует без изменений парных контрактов.

## 2. Границы

In scope: верификация и заморозка схем событий v1, черновика OpenAPI мерчанта,
обновление системных спек и тега версии.
Out of scope: финализация OpenAPI мерчанта (7.1) и админ-API (12.1), контракты CF-3
(Notification, Reporting, `recon.alerts`, Admin API), инфраструктура Schema Registry (0.5).

## 3. Состав заморозки

| Схема | Файл | Топик | Потребители |
| :-- | :-- | :-- | :-- |
| `PaymentEvent` (типы created...failed) | `proto/billing/events/v1/payment_events.proto` | `payment.events` | Ledger (2.7), Notification (8.x), ClickHouse (11.x) |
| `LedgerEvent` + `LedgerEntry` (transaction.posted, reserve.*) | `proto/billing/events/v1/ledger_events.proto` | `ledger.events` | Payment (4.x), Payout (10.x), ClickHouse |
| `PayoutCommand` | `proto/billing/events/v1/payout_events.proto` | `payout.commands` | Connector (10.3) |
| `GatewayPayoutEvent` (submitted/confirmed/failed) | `proto/billing/events/v1/gateway_events.proto` | `gateway.payout.events` | Payout (10.x), Reconciler (9.x) |
| `PayoutEvent` (requested...compensated) | `proto/billing/events/v1/payout_events.proto` | `payout.events` | Notification, ClickHouse |
| `EventMeta` (конверт) | `proto/billing/events/v1/envelope.proto` | все топики | все |
| OpenAPI мерчанта черновик | `openapi/billing.v1.yaml` (1.0.0-draft) | - | Billing API (7.x), API Gateway |

`EventMeta` заморожен еще в CF-1 и не менялся; `payout.events` и OpenAPI-черновик входят
в тег cf-2 для волны W2 (CBL-004, CF-2 дополнен `payout.events`). Проводные схемы
(номера полей, типы) с cf-1 не менялись: добавлены только комментарии семантики
(reserve-события, `occurred_at`, `state_version`), генерация обновляет стабы
без изменения wire-формата.

## 4. Функциональные требования

FR-1. Верификация схем против реализаций W1: `PaymentEvent` соответствует контракту
потребителя 2.7 (completed -> Dr clearing / Cr merchant / Cr fee; consumer читает
payment_id, type, merchant_id, received_amount, fee; остальные поля доступны
потребителям W2/W3); `LedgerEvent` соответствует продюсеру 2.7 (transaction.posted,
reserve.created/committed/released; transaction_id, ref_type, ref_id, entries[];
семантика reserve-событий зафиксирована комментариями proto); `GatewayWebhookRaw`
и `CanonicalPaymentEvent` закреплены использованием в 5.1 и 3.5; `GatewayPayoutEvent`
и `PayoutCommand` готовы для W2. Семантика `occurred_at` (тело - авторитетно, заголовок
конверта зеркалит) и `state_version` (монотонный счетчик переходов) зафиксирована
комментариями proto.

FR-2. Совместимость: `buf lint` (STANDARD) и `buf breaking` против тега
`cbl-protos@cf-1` (FILE) - без замечаний; генерация идемпотентна (`buf generate`
меняет `gen/go` только вслед за комментариями proto), `go build ./...` зеленый.

FR-3. Документация: `CBL-003` раздел 6 приведен в соответствие фактическим схемам
(пример `PaymentCompleted` заменен на `PaymentEvent`, каталог уточнен), статус точек
заморозки обновлен (CBL-003 раздел 8, CBL-004 раздел 3). Решение по ключу
`ledger.events` (открытый вопрос T-LED-09): ключ `account_id` первой проводки
оставлен, пересмотр - только при появлении требования строгого порядка событий счета.

FR-4. Фиксация: аннотированный тег `cbl-protos@cf-2` на коммите заморозки; правило
после заморозки - только обратно совместимые изменения (BACKWARD_TRANSITIVE),
ломающие - новая мажорная версия пакета и переход всех потребителей одной волной.
Контрактные заморозки не имеют номера задачи плана: коммит `CBL-CF2`, тег
`cbl-protos@cf-2` (оговорка зафиксирована в AGENTS).

FR-5. Черновик OpenAPI мерчанта (1.0.0-draft) покрывает маршруты `CBL-003` раздела 2
(внешний REST мерчанта): `/invoices`, `/invoices/{id}`, `/invoices/{id}/cancel`,
`/payouts`, `/payouts/{id}`, `/refunds`, `/balances`, `/ledger/entries`,
`/reports/{report_type}`, `/webhook-endpoints`; финализация форматов - 7.1,
заморозка черновика фиксирует состав маршрутов и правки ревью безопасности
(проверка владельца `account_id` в `/ledger/entries`, `Idempotency-Key`
на `/webhook-endpoints`, неотрицательные `Money`).

## 5. Проверки

| Проверка | Команда | Результат |
| :-- | :-- | :-- |
| lint | `buf lint` (bufbuild/buf:1.47.2) | clean |
| breaking | `buf breaking --against ".git#tag=cbl-protos@cf-1"` | clean (gate проверен санити-мутацией поля) |
| верификация W1 | сверка схем с `cbl-ledger/internal/usecase/{ledger_events.go,payment_consumer.go}`, `cbl-connector/internal/usecase/normalizer.go`, `cbl-webhook-ingress/internal/usecase/handler.go` | соответствует FR-1 |
| generate | `buf generate` (против закоммиченного `gen/go`) | дифф только от комментариев proto |
| build | `go build ./...` | clean |
| эм-деш | `scripts/check-no-emdash.ps1` | clean |

## 6. Нефункциональные требования

Проводные схемы при заморозке не меняются (допустимы только wire-совместимые
уточнения комментариев); subject в Schema Registry - по имени топика, режим
`BACKWARD_TRANSITIVE` (0.5; до подключения SR совместимость держит `buf breaking`);
стабы в `gen/go` закоммичены и воспроизводимы.

## 7. Edge cases

Изменение схемы после заморозки (только BACKWARD-совместимое; иначе мажор); расхождение
`CBL-003` и proto (системная спека обновлена этим коммитом); новые поля потребители
обязаны игнорировать (protobuf); enum-значения добавляются только в конец.

## 8. Ревью (CBL-001, шаг 6)

Четыре независимых прохода (15.10.2026). Блокеров нет. MAJOR: строка `payout.commands`
в CBL-003 не соответствовала замороженной схеме (исправлена), статусы документов
опережали факт заморозки (закрыто коммитом и тегом одновременно с этой фиксацией),
OpenAPI-черновик закреплял IDOR-риск `/ledger/entries` и мутацию без `Idempotency-Key`
(исправлено до заморозки). MINOR/NIT закрыты: состав заморозки согласован в CBL-003/
CBL-004, ссылка FR-5 на раздел 2, метка черновика CF-2, шапка CBL-003, семантика
reserve-событий и occurred_at/state_version в комментариях proto, ключ ledger.events
оставлен (T-LED-09), цель `make breaking`, оговорка про `CBL-CFn`.

| Роль | Ключевые замечания | Закрытие |
| :-- | :-- | :-- |
| Без контекста | MAJOR: строка payout.commands (SendPayout/asset) против схемы; состав в спеке шире CBL-004; ссылка FR-5 на раздел 3; статусы опережают тег; метка «Черновик CF-1»; FR-1 без артефактов проверки; устаревшая шапка CBL-003; пути без префикса proto/ | CBL-003 исправлен; состав согласован (payout.events + EventMeta с пояснением); FR-5 -> раздел 2; финализация синхронно с тегом; метка CF-2; таблица проверок расширена; шапка обновлена; пути полные |
| Архитектуры | статусы опережают заморозку; состав расходится; FR-5 не тот раздел; метка CF-1; шапка CBL-003; потребитель ledger.events в CBL-000; reserve-семантика не зафиксирована; occurred_at/state_version; KI-01 формулировка | заморозка+тег в этом коммите; состав согласован; раздел 2; метка CF-2; шапка; CBL-000 §2 упоминает canonical (потребителя ledger.events определит 4.x, зафиксировано в CF-3 при необходимости); комментарии proto; KI-01 обновлен; W2-гейт уточнен (9.1 ждет 3.2) |
| С контекстом | состав заморозки; формат коммита CBL-CF2; статусы; метка CF-2; ключ ledger.events не пересмотрен; нет make breaking; шапка спеки; CBL-003 payout.commands; FR-1 переоценивает Reads 2.7 | все закрыто: оговорка в AGENTS §5; решение по ключу в FR-3; `make breaking`; шапка; строка исправлена; FR-1 уточнен |
| Безопасности | MAJOR: IDOR в `/ledger/entries`; `POST /webhook-endpoints` без Idempotency-Key; NIT: SR не enforced (KI-P2), reserved-задел, Money без валидации в схеме, occurred_at дубль | описания исправлены до заморозки (владелец из токена, Idempotency-Key, Money неотрицателен); SR - 0.5 (KI-P2); reserved - правило CBL-003 §6; occurred_at - комментарий proto (тело авторитетно) |

## 9. Открытые вопросы

- Валидатор OpenAPI не подключен (KI-P1, 0.2): черновик проверен вручную на состав
  маршрутов и правки ревью безопасности.
- KI-01 (провайдер) не влияет на CF-2: схемы событий нейтральны к провайдеру
  (`provider` - строка). При этом задача 9.1 (Reconciler) зависит от 3.2 и не может
  стартовать до выбора провайдера: CF-2 разблокирует контрактный вход W2, не сами
  задачи, зависящие от KI-01.
- До подключения Schema Registry (0.5) режим `BACKWARD_TRANSITIVE` не enforced
  инфраструктурно: совместимость держит только `buf breaking` (KI-P2).

## 10. План работ и коммит

Верификация (FR-1, FR-2), правки `CBL-003`, обновление AGENTS и журнала, коммит
`CBL-CF2: заморозка схем событий v1 и OpenAPI мерчанта`, аннотированный тег
`cbl-protos@cf-2`, обновление `CBL-004`.
