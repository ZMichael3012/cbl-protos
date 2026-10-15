# CBL-CF3. Заморозка контрактов Notification, Reporting, recon.alerts и Admin API

| Поле | Значение |
| :-- | :-- |
| Эпик | Контрактная точка CF-3 (CBL-004, раздел 3) |
| Репозиторий | `cbl-protos` |
| Размер | S (заморозка; Reporting/recon.alerts созданы в 1.7, webhooks, Admin API и gRPC Notification дорабатываются до заморозки) |
| Волна | конец W2, до старта W3 |
| Зависит от | CF-2 (тег `cbl-protos@cf-2.1`), реализаций W2 (4.x, 10.x - для сверки семантики) |
| Версия контрактов | `cbl-protos@cf-3` (этот тег) |
| Статус | in progress |

## 1. Контекст и цель

CF-2 заморозил схемы событий фазы 2 и черновик OpenAPI мерчанта. К концу W2
реализованы Payment (4.x) и Payout (10.x), определены фактические состояния и
события, от которых зависит доставка вебхуков мерчанту (Notification, 8.x), а
также контрактные составы отчетов (Reporting, 11.x) и расхождений (Reconciler,
9.x). CF-3 фиксирует контракты волны W3: webhooks мерчанта (типы и подпись),
Reporting `Query` v1, `recon.alerts` v1, черновик Admin API v1 и управление
эндпоинтами Notification, после чего волны W3 и W4 стартуют без изменений
парных контрактов.

## 2. Границы

In scope: новый артефакт `openapi/webhooks.v1.yaml` (конверт, каталог типов,
подпись, повторы, схемы данных), заморозка черновика `openapi/admin.v1.yaml`,
новый пакет `proto/billing/notification/v1` (управление эндпоинтами), верификация
`proto/billing/reporting/v1/reporting.proto` и
`proto/billing/events/v1/recon_events.proto` (аддитивное `merchant_id`),
аддитивный тип `rejected` в `payout.events`, обновление `CBL-003`/`CBL-004`,
AGENTS, тег `cbl-protos@cf-3`.

Out of scope: финализация OpenAPI мерчанта (7.1) и Admin API (12.1), реализация
сервисов 8.x/9.x/11.x/12.x, инфраструктура Schema Registry (0.5), исполнение
ручных корректировок расхождений (дизайн 12.3 после 9.2).

## 3. Состав заморозки

| Схема | Файл | Потребители |
| :-- | :-- | :-- |
| Webhooks мерчанта: конверт, каталог типов, подпись, повторы, схемы данных | `openapi/webhooks.v1.yaml` (новый, 1.0.0-cf3) | `cbl-notification` (8.1 - 8.3), мерчанты (15.4), Billing API (7.1) |
| Admin API v1 (черновик, финализация 12.1) | `openapi/admin.v1.yaml` (1.0.0-draft) | `cbl-backoffice` (12.x), API Gateway |
| Управление эндпоинтами Notification | `proto/billing/notification/v1/notification.proto` (новый) | `cbl-notification` (8.1, сервер), `cbl-billing-api` (7.x, клиент) |
| Reporting `Query` v1 | `proto/billing/reporting/v1/reporting.proto` | `cbl-reporting` (11.4), `cbl-billing-api` |
| `recon.alerts` v1 (`ReconAlert` + `merchant_id`) | `proto/billing/events/v1/recon_events.proto` | Reconciler (9.2), back-office (12.3), Notification, ClickHouse (11.x) |
| `payout.events`: аддитивный `PAYOUT_EVENT_TYPE_REJECTED` | `proto/billing/events/v1/payout_events.proto` | Notification (8.x: `payout.failed`/`refund.failed` со статусом `rejected`), ClickHouse |
| `EventMeta` (конверт) | `proto/billing/events/v1/envelope.proto` | без изменений (заморожен CF-1) |

## 4. Функциональные требования

FR-1. Notification (типы, данные и подпись). `openapi/webhooks.v1.yaml`
фиксирует:

- конверт доставки: заголовки `X-Billing-Signature` (схема
  `t=<unix>,v1=<hex HMAC-SHA256(secret, t + "." + body)>`, hex строчными
  буквами), `X-Billing-Event-Id` (равен `id` тела и неизменен при повторах,
  меняется только `t`), `Content-Type: application/json`; тело
  `{"id":"evt_<uuidv7>","type":"<type>","created_at":"<RFC 3339 UTC>","data":{...}}`;
- каталог типов и маппинг источников: `payment.events`
  (`detected/confirming/completed/expired/underpaid/overpaid/failed`) ->
  `invoice.*`; `payout.events` без маркера возврата (`completed` ->
  `payout.completed`; терминальные `failed/compensated/rejected` ->
  `payout.failed` с точным состоянием в `data.payout.status`); `payout.events`
  с непустым `refund_invoice_id` -> `refund.completed`/`refund.failed`
  (`data.refund.status`); `payment.created` и `payout.requested/approved`
  вебхуков не порождают;
- терминальные состояния доставляются полностью: `invoice.failed` закрывает
  отказ платежа (состояние `failed` есть в Payment 4.1 и эмитится 4.7),
  `rejected` - вето оператора: на CF-3 в `payout.events` добавлен аддитивный
  тип `PAYOUT_EVENT_TYPE_REJECTED`, а `cbl-payout` эмитит его в транзакции
  перехода (сопроводительный коммит `CBL-10.5 (fix)`); без этого обещание
  статуса `rejected` не имело источника;
- схемы данных самодостаточны (не ссылаются на черновик `billing.v1.yaml`):
  `data.invoice` несет `id`, `status`, `amount`, `paid_amount`, `fee`, `rate`,
  `txid`, `occurred_at`; `data.payout` и `data.refund` - `id` (для возврата
  `invoice_id` - исходный счет возврата), `status`, `amount`, `fee`, `txid`,
  `failure_reason`, `occurred_at`; значения полей берутся из события-источника,
  отсутствующие (адрес, сроки) мерчант получает через REST GET; каталог
  расширяется аддитивно, неизвестный тип мерчант принимает (2xx) и игнорирует;
- подпись проверяется по сырому телу (постоянное по времени сравнение), окно
  `t` +/- 5 минут; секрет - не менее 32 случайных байт CSPRNG, показывается
  один раз и не логируется; мерчант дедуплицирует по `id` тела, сверяя его с
  `X-Billing-Event-Id`;
- доставка: успех - только `2xx` (3xx не успех), таймаут попытки 10 с,
  повторы 1 с, 5 с, 30 с, 2 мин, 15 мин, 1 ч; после исчерпания - `failed` и
  ручной retry (8.3); порядок доставки не гарантируется;
- регистрация (`billing.v1.yaml`, `POST /v1/webhook-endpoints`): `events` -
  enum каталога CF-3, `minItems: 1`, `uniqueItems`, неизвестный тип - 422;
  доставка фильтруется по списку; новые типы каталога не включаются в подписку
  автоматически (нужна повторная регистрация); `url` - только `https` без
  userinfo и IP-литералов, SSRF-защита платформы нормативно (блокировка
  private/loopback/link-local/metadata после DNS-резолвинга, без redirect-ов).

FR-2. Reporting `Query` v1: комментарии proto закрепляют семантику без изменения
проводного формата: `merchant_id` заполняет Billing API из токена (11.4), прямые
вызовы gRPC - внутренние; `limit` (0 - дефолт 20 как `Limit` в billing.v1.yaml,
максимум 100); `cursor` непрозрачен; денежные значения и курсы в
`ReportRow.values` - десятичные строки канонического вида (без экспоненты, float
запрещен); каталог `report_type` фиксируется 11.4/15.4; дедлайн RPC 5 с
(CBL-003, раздел 5).

FR-3. `recon.alerts` v1: в `ReconAlert` включено поле `merchant_id = 8`
(комментарий протокола: UUID мерчанта, заполняется продюсером 9.2 только при
привязке к мерчанту; пусто - запись только для back-office и ClickHouse,
доставки мерчанту нет); `alert_id` равен `event_id` конверта (единый ключ
дедупликации inbox и ClickHouse); `expected`/`actual` - строки (десятичные
суммы или имена состояний), ключ топика `ref_id`, `created_at` в теле
авторитетно. Мерчантские уведомления о расхождениях в каталог типов CF-3 не
входят (решение о типе - 8.x/12.3, аддитивно через cf-3.x).

FR-4. Admin API v1: черновик `openapi/admin.v1.yaml` замораживается: маршруты
CBL-003 раздела 3 (`/payouts/{id}/approve`, `/reject`, `/routing-rules` GET/PUT,
`/reconciliation/discrepancies` GET с фильтром `merchant_id` и `/{id}/resolve`,
`/providers/{id}/disable`), bearer-токен оператора, `application/problem+json`,
ответы `400/401/403/404/409/422`, двухэтапный апрув по политике порогов (по
умолчанию два оператора; повторный голос - 409; вето одного достаточно),
`operator_id` тела обязан совпадать с субъектом токена, иначе 403 (аудит - по
субъекту токена); замещение правил маршрутизации - с ревизией (`revision` из
GET, несовпадение - 409), отдельной RBAC-ролью и оговоркой о задержке
применения (30 - 60 с, кеш Connector); повторное отключение провайдера
идемпотентно (204); аудит-запись (кто, что, до/после) на каждое действие;
финализация - 12.1. Отдельный `Idempotency-Key` на мутациях Admin API не
вводится: идемпотентность голосов/решений обеспечена доменом (append-only,
409 при повторе); пересмотр - 12.1 (trade-off).

FR-5. Notification gRPC: новый пакет `billing/notification/v1` с
`NotificationService.RegisterWebhookEndpoint` (`idempotency_key`, `merchant_id`,
`url`, `events[]` -> `endpoint_id`, `url`, `events[]`, `signing_secret`) - канал
Billing API (7.x) и Back-office (12.x) к Notification (8.1); `merchant_id`
берется из токена мерчанта, события валидируются по каталогу CF-3; дедлайн 2 с
(CBL-003, раздел 5).

FR-6. Совместимость: `buf lint` (STANDARD) без замечаний; `buf breaking` против
тега `cbl-protos@cf-2.1` (FILE) чистый (комментарии и аддитивные поля/типы
допустимы); `buf generate` обновляет `gen/go` (комментарии и новое поле/тип
попадают в стабы), `go build ./...` зеленый.

FR-7. Фиксация: `CBL-003` разделы 3/4/5/6 приведены к контрактам (каталог с
`invoice.failed`, равенство заголовка и `id`, стабильность id при повторах,
`rejected`, `merchant_id`, строка Notification), `CBL-000` раздел 2
(back-office потребляет `recon.alerts`), AGENTS `cbl-protos` - артефакты,
потребители, журнал; аннотированный тег `cbl-protos@cf-3`. После заморозки
изменения только BACKWARD-совместимые (инкрементальные теги `cf-3.x` по правилу
`CBL-003` раздела 8).

## 5. Проверки

| Проверка | Команда | Результат |
| :-- | :-- | :-- |
| lint | `buf lint` (bufbuild/buf:1.47.2) | clean |
| breaking | `buf breaking --against ".git#tag=cbl-protos@cf-2.1"` | clean |
| generate | `buf generate` против закоммиченного `gen/go` | дифф комментариев, `MerchantId` и типа `PAYOUT_EVENT_TYPE_REJECTED` в стабах |
| build | `go build ./...` | clean |
| сверка | webhooks-каталог против `CBL-003` раздела 4 и `payment_events.proto`/`payout_events.proto`; Admin API против `CBL-003` раздела 3; Notification RPC против задачи 8.1 | соответствует |
| payout | эмиссия `payout.rejected` при вето (`cbl-payout`, integration-тест) | integration green |
| эм-деш | `scripts/check-no-emdash.ps1` | clean |

## 6. Нефункциональные требования

Проводные схемы Reporting и `recon.alerts` меняются только аддитивно
(`merchant_id`; `PayoutEventType.REJECTED` добавляется значением в конец enum);
OpenAPI-артефакты - обязательный контракт для 8.x/12.x после заморозки;
валидатор OpenAPI не подключен (KI-P1), проверка вручную; режим
`BACKWARD_TRANSITIVE` держит `buf breaking` до подключения Schema Registry
(KI-P2).

## 7. Edge cases

Изменение после заморозки (только BACKWARD-совместимое; иначе мажор и повторный
цикл); новый тип вебхука - аддитивное значение каталога и явная повторная
регистрация (автовключение в подписки не гарантируется); `data.refund` для
частичных возвратов несет сумму конкретной выплаты-возврата, агрегация - Billing
API 7.x; мерчант без зарегистрированных endpoints - доставок нет (8.1);
таймаут попытки 10 с, ответ 3xx не успех, неизвестный тип игнорируется с 2xx;
потеря ответа платформой после успешной обработки мерчантом - повтор с тем же
`id` и новым `t`, дедупликация на стороне мерчанта; ротация секрета подписи и
двойная подпись при ротации - аддитивное расширение 7.1/8.1 (13.2), в cf-3 не
входит; расхождение `CBL-003` и артефактов - правится этой спекой до тега.

## 8. Ревью (CBL-001, шаг 6)

Четыре независимых прохода (15.10.2026). Блокеров после правок нет. BLOCKER
закрыты: обещание статуса `rejected` без источника (аддитивный
`PAYOUT_EVENT_TYPE_REJECTED` + эмиссия в `cbl-payout`, integration-тест);
подмена авторства оператора (субъект RBAC/аудита - только из bearer-токена,
`operator_id` тела обязан совпадать, иначе 403). MAJOR закрыты: `data.*` больше
не ссылаются на изменяемый черновик `billing.v1.yaml` (самодостаточные схемы
данных); отсутствовавший gRPC-контракт управления эндпоинтами Notification
(новый `billing/notification/v1`); `events[]` регистрации получил enum каталога
и `minItems`; `merchant_id` добавлен в `Discrepancy` и фильтр списка; задокумен-
тированы источник полей `data` (событие + REST GET); `operator_id`/ревизия
routing/идемпотентность Admin API зафиксированы; противоречие back-office
(consumes `recon.alerts`) устранено в `CBL-000` и AGENTS; формулировки
`resolved`-повтора и идемпотентности поправлены; `X-Billing-Event-Id` связан с
`id` тела и стабилен при повторах; https/SSRF и требования к секрету
нормативны. MINOR/NIT закрыты: паттерн UUIDv7 для id, lowercase hex подписи,
`additionalProperties: false` на запросах, `PayoutAdmin.amount` - `Money`,
идемпотентный повтор disable (204), комментарии proto (Reporting: дефолт 20,
канонический десятичный вид; recon: producer-валидация, alert_id = event_id),
`recon.alert` в каталоге `event_type`, порядок статусов документов (тег в том же
коммите), комментарий Makefile про инкрементальные теги. Принято без изменений:
отсутствие `Idempotency-Key` на мутациях Admin API (доменная идемпотентность,
пересмотр - 12.1); мерчантский тип уведомлений о расхождениях (8.x/12.3,
аддитивно); ротация секрета подписи (7.1/8.1, 13.2); исполнение
`manual_adjustment` (12.3 после 9.2).

| Роль | Ключевые замечания | Закрытие |
| :-- | :-- | :-- |
| Без контекста | BLOCKER: статус `rejected` без типа события; MAJOR: `data` из черновика, источник полей, фильтр/поле merchant_id в Admin API, уведомления о расхождениях без типа; MINOR: версия файла vs спеки, breaking-тег до выпуска, формулировка resolved-повтора, паттерны, edge cases, trade-offs; NIT: UUIDv7-паттерн, терминология, event_type `recon.alert`, теги | REJECTED в cf-3 + эмиссия; самодостаточные схемы; источник полей зафиксирован; merchant_id/фильтр; решение о типе отложено явно; версия согласована; порядок ревью-правки-тег; формулировки; паттерны/крайние случаи; trade-offs; комментарии; терминология; каталог; Makefile |
| Архитектуры | BLOCKER: rejected без эмиссии; MAJOR: нет gRPC управления эндпоинтами Notification, канал back-office/Reconciler, `data` невыводима из топиков, `events[]` без каталога, админ-мутации без идемпотентности, operator_id vs токен, routing без ревизии, схема ClickHouse, фильтр merchant_id, поверхность 12.2, дрейф на черновик, контракт 8.3; MINOR/NIT: статусы/тег, trade-offs, дефолт limit, перегрузка Refund, ротация секрета | REJECTED + эмиссия; `billing/notification/v1`; back-office consumes `recon.alerts` + дизайн 12.3; самодостаточные схемы; enum `events[]`; доменная идемпотентность + 12.1; токен-субъект; ревизия; alert_id = event_id; merchant_id/фильтр; отложено 12.2 явно; самодостаточные схемы; 8.3 - ручной retry в 8.x + статусы; тег в коммите; trade-offs §9; дефолт 20; терминология; отсрочено 7.1/8.1 |
| С контекстом | BLOCKER нет; MAJOR: rejected, версия admin-файла, operator_id; MINOR: шапка CBL-003, не-порождающие события, потребители recon.alerts/back-office, resolved-повтор, статусы опережают тег; NIT: аддитивность merchant_id, дефолт limit, комментарии AGENTS/Makefile, UUIDv7-паттерн | REJECTED + эмиссия; версия согласована; токен-субъект; шапка обновлена; оговорка добавлена; back-office/`recon.alert`; формулировка; единый коммит и тег; формулировка в proto; дефолт 20; поправлено; паттерн усилен |
| Безопасности | BLOCKER: operator_id из тела (обход four-eyes); MAJOR: `X-Billing-Event-Id` вне подписи и стабильность при повторах, https/SSRF ненормативны, требования к секрету, routing без конкурентного контроля, семантика merchant_id; MINOR/NIT: паттерн заголовка, `events` без каталога, условный reason, mass-assignment, disable-повтор, статусы/тег, статус `rejected`, неизвестный тип, формат Money/строк | токен-субъект + 403; id = header, стабильность, сверка мерчантом; https-паттерн + норматив SSRF; секрет CSPRNG/один раз/логи; revision + 409 + RBAC-роль; producer-валидация и пустое значение; паттерн; enum/minItems/422; 422 для reason; additionalProperties false; 204-повтор; единый коммит; закрыт типом REJECTED; правило игнорирования; канонический формат |

## 9. Открытые вопросы

- Валидатор OpenAPI не подключен (KI-P1, 0.2): артефакты проверены вручную на
  состав и соответствие `CBL-003`.
- До подключения Schema Registry (0.5) режим `BACKWARD_TRANSITIVE` не enforced
  инфраструктурно: совместимость держит `buf breaking` (KI-P2).
- KI-01 не влияет на CF-3: контракты нейтральны к провайдеру (строковые
  реквизиты); Reconciler 9.x остается заблокированным задачей 3.2.
- Заполнение `merchant_id` в `recon.alerts` - требование к 9.2 при публикации
  расхождений, привязанных к мерчанту; доставка уведомлений о расхождениях
  мерчантам не входит в каталог CF-3 (тип - решение 8.x/12.3, аддитивно).
- Ротация секрета подписи и двойная подпись при ротации - аддитивное расширение
  7.1/8.1 (13.2); отсутствие `Idempotency-Key` на мутациях Admin API - trade-off
  FR-4, пересмотр 12.1.
- `ListWebhookEndpoints`/`Delete`/перевыпуск секрета в `billing/notification/v1`
  не входят в минимальный контракт 8.1 и добавляются аддитивно (cf-3.x) при
  появлении требования.

## 10. План работ и коммит

Верификация и правки proto (комментарии, `merchant_id`, `REJECTED`,
`notification/v1`), регенерация, правки OpenAPI (`webhooks.v1.yaml`,
`admin.v1.yaml`, `billing.v1.yaml`), четыре ревью-прохода, правки `CBL-003`/
`CBL-004`/`CBL-000`/AGENTS, сопроводительный коммит `CBL-10.5 (fix)` в
`cbl-payout` (эмиссия `rejected`), коммит `CBL-CF3: заморозка контрактов
Notification, Reporting, recon.alerts и Admin API`, аннотированный тег
`cbl-protos@cf-3`.
