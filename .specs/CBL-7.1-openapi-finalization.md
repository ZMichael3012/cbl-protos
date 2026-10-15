# CBL-7.1. Финализация OpenAPI v1 мерчанта

| Поле | Значение |
| :-- | :-- |
| Эпик | E7 (Billing API и Gateway) |
| Репозиторий | `cbl-protos` |
| Размер | M (финализация форматов и кодов, маршруты заморожены CF-2/CF-3) |
| Волна | W3-BILLING (первая задача) |
| Зависит от | CF-3 (тег `cbl-protos@cf-3`) |
| Версия контракта | `billing.v1.yaml` 1.0.0, тег `cbl-protos@cf-3.1` (аддитивная финализация поверх cf-3) |
| Статус | done, ревью закрыто |

## 1. Контекст и цель

CF-2 заморозил состав маршрутов черновика мерчант-API (`1.0.0-draft`), CF-3
доработал webhook-части (`events[]` enum, https/SSRF, секрет). 7.1 финализирует
контракт: полные наборы ответов по кодам CBL-003, единые Problem-компоненты с
`Retry-After`, семантика идемпотентности (`Idempotent-Replayed`, 409
`idempotency_key_reuse`), обязательные поля схем, правила безопасности. После
7.1 контракт закрыт для W3-BILLING (7.2 - 7.6) и документации 15.4.

## 2. Границы

In scope: `openapi/billing.v1.yaml` (форматы, коды, заголовки, версия 1.0.0),
синхронизация `CBL-003` раздела 2, AGENTS и журнал. Out of scope: реализация
Billing API и Gateway (7.2 - 7.6), checkout (7.6), `webhooks.v1.yaml` и
`admin.v1.yaml` (заморожены CF-3), протобаф-контракты.

## 3. Функциональные требования

FR-1. Состав маршрутов не меняется (CBL-003, раздел 2): `POST/GET /invoices`,
`GET /invoices/{id}`, `POST /invoices/{id}/cancel`, `POST /payouts`,
`GET /payouts/{id}`, `POST /refunds`, `GET /balances`, `GET /ledger/entries`,
`GET /reports/{report_type}`, `POST /webhook-endpoints`.

FR-2. Единая модель ошибок: `application/problem+json` (RFC 9457,
`type/title/status/code/request_id` обязательны); на всех маршрутах - `401`
(с `WWW-Authenticate: Bearer`) и `403`; `404` - ресурс не найден или не
принадлежит мерчанту (чужие идентификаторы не раскрываются; ownership-проверка
предшествует бизнес-валидации - требование к 7.4); `409` - конфликт состояния
или ключа идемпотентности; `422` - бизнес-валидация; `429` и `503` несут
`Retry-After` и объявлены на всех проксирующих маршрутах (503: `code`
`provider_unavailable` для шлюза, `service_unavailable` для внутренних
сервисов). `code` - стабильный машиночитаемый код: `idempotency_key_reuse`,
`invoice_not_cancellable`, `webhook_endpoint_exists`, `unknown_report_type`,
`insufficient_funds`, `validation_failed`, `provider_unavailable`,
`service_unavailable`, `rate_limited` (перечень открытый, аддитивен);
`detail` не содержит секретов и персональных данных.

FR-3. Идемпотентность: все `POST` требуют `Idempotency-Key` (заморожено CF-2);
ключ скоупится мерчантом из токена (уникальный индекс PostgreSQL
`(merchant_id, idempotency_key)` - 7.3, T-02); отпечаток - метод, путь и хэш
сырого тела; повтор отпечатка возвращает исходный успешный ответ плюс заголовок
`Idempotent-Replayed: true` (на первом успехе заголовка нет); тот же ключ с
другим отпечатком - `409 idempotency_key_reuse`; кэшируются только успешные
ответы (ошибки, включая 503, ключ не занимают); окно хранения 24 часа
(CBL-003, раздел 7); для регистрации вебхука реплей возвращает исходный ответ
вместе с `signing_secret` (иначе секрет теряется безвозвратно).

FR-4. Обязательные поля схем ответов и ответных заголовков: `Invoice` -
`id/status`; `Payout`/`Refund` - `id/status`; `Quote` - `rate/locked_until`;
`TimelineItem` - `state/occurred_at`; списки - `items`/`rows`; балансовые
строки - `currency/available` (структура: `available`, `reserved`, `pending`).

FR-5. Запросы закрыты для mass-assignment (`additionalProperties: false`):
`InvoiceCreateRequest`, `PayoutCreateRequest` (дополнительно к CF-3-схемам
refund/webhook); `merchant_ref` - форма `^[A-Za-z0-9_-]{1,128}$` (бизнес-ключ
плана счетов, CBL-003 раздел 9); `expires_in_sec` 60 - 86400 (не задан -
900 секунд); `amount` больше нуля (серверная валидация); `metadata` -
32 свойства, ключи до 64, значения до 512 символов.

FR-6. Безопасность: bearer API-ключ мерчанта (выдача/ротация/отзыв - вне
мерчант-API: 7.2, 12.x) проверяется на периметре и в BFF; `merchant_id` в
запросах не принимается - все ресурсы скоупятся субъектом токена; владение
`account_id` в `/ledger/entries` - 404 при чужом (предпосылка - 2.10);
`Idempotency-Key` ограничен символьным классом и длиной (защита от log
injection); webhook-регистрация - https/SSRF-норматив CF-3, ужесточенный до
проверки на каждой попытке доставки (DNS-rebinding, IPv4/IPv6-диапазоны, без
redirect), секрет - hex не менее 64 символов; `detail` Problem без секретов.

FR-7. Версия и правила: `version: 1.0.0`; в `info.description` зафиксированы
финал 7.1 и правило изменений (только назад-совместимые аддитивные правки;
ломающие - `/v2` и новая волна); расширяемые перечни (`InvoiceStatus`,
`PayoutState`, `direction`) помечены `x-extensible-enum` с обязанностью
клиента терпеть неизвестные значения. `CBL-003` разделы 2 и 8 дополнены
(заголовок `Idempotent-Replayed`, коды, финальная версия, правило тегов
`cf-N.M`); AGENTS и журнал `cbl-protos` обновлены; `CBL-004` - статус 7.1 и
контрактные предпосылки 7.4 (задачи 2.10, 4.8); `CBL-000` - связь
Billing API -> Notification.

FR-8. Предпосылки 7.4 (зафиксированы ревью 7.1, реализуются отдельными
задачами до старта 7.4): 2.10 - аддитивное расширение Ledger gRPC
(`ListMerchantAccounts` со счетами, балансами и ownership; period-фильтр
`ListEntries`) для `/balances` и `/ledger/entries`; 4.8 - аддитивное
расширение Payment gRPC (cf-3.3/cf-3.4: `rate`, `locked_until`,
`txid`/`confirmations`, `expected_amount` создания (источник `pay_amount`
ответа 201), период `ListPayments`) для полноты инвойса; оба - пакеты cf-3.x.
`pending` в балансах - агрегация Payment (состояния detected/confirming), не
данные Ledger; checkout-контракт создается задачей 7.6 отдельным артефактом,
7.1 фиксирует только `checkout_url`.

## 4. Проверки

| Проверка | Команда | Результат |
| :-- | :-- | :-- |
| YAML parse | `python -c "import yaml; yaml.safe_load(...)"` | clean (10 маршрутов, 11 операций, version 1.0.0) |
| сверка маршрутов | против `CBL-003` раздела 2 (10 маршрутов) | соответствует |
| сверка кодов | ответы всех операций против FR-2 | соответствует |
| buf | `buf lint`, `buf breaking --against ".git#tag=cbl-protos@cf-3"`, `go build ./...` | clean |
| эм-деш | `scripts/check-no-emdash.ps1` | clean |

## 5. Нефункциональные требования

Маршруты и webhook-части не меняются (заморожены CF-2/CF-3); additive-only
после 1.0.0; валидатор OpenAPI не подключен (KI-P1) - проверка вручную.

## 6. Edge cases

Повтор идемпотентного запроса с тем же телом после смены состояния ресурса -
исходный ответ (replay важнее текущего состояния); разные тела при одном ключе
- 409 независимо от состояния; ошибки не кэшируются (503 не «отравляет» ключ);
чужие id - 404 (ownership предшествует бизнес-валидации); неизвестный
`report_type` - 404 `unknown_report_type` (каталог типов - 11.4, аддитивен);
создание счета может вернуть `creating` - клиент опрашивает GET;
`Retry-After` обязателен на 429/503; `Idempotent-Replayed` только на реплеях;
реплей регистрации вебхука возвращает `signing_secret`; новые состояния и типы
переживаются клиентом (`x-extensible-enum`).

## 7. Ревью (CBL-001, шаг 6)

Четыре независимых прохода (15.10.2026). BLOCKER: скоупинг `Idempotency-Key` -
ключ привязан к мерчанту (уникальный индекс `(merchant_id, idempotency_key)`),
отпечаток метод+путь+сырое тело, реплей только успешных ответов - закрыт.
MAJOR: нехватка gRPC-источников для 7.4 (`/balances`, `/ledger/entries`,
полнота 201 счета) зафиксирована предпосылками отдельными задачами 2.10 и 4.8
(cf-3.x) - FR-8; версия финализируется тегом `cbl-protos@cf-3.1` (правило
`cf-N.M` внесено в CBL-003 §8 и AGENTS §5); идемпотентность дополнена TTL 24 ч,
некэшированием ошибок и канонизацией отпечатка; жизненный цикл webhook-эндпоинта
(ротация/удаление) - trade-off: вне 1.0.0, компенсация через перевыпуск
(8.1/13.2); `creating`-семантика 201 описана; 503 выровнен на всех операциях;
секрет вебхука - hex и replay-семантика; SSRF-норматив расширен (per-attempt
resolver, IPv6/IPv4-диапазоны, DNS-rebinding); Admin API разрешения сужены.
MINOR/NIT закрыты: `x-extensible-enum` для состояний и direction, закрыты
запросы и лимиты (`metadata`, `merchant_order_id`, `group_by`, `cursor`),
паттерны UUIDv7 для id/invoice_id, `Money.maxLength` и `Quote.rate`-паттерн,
`request_id` обязателен, `detail` без секретов, `WWW-Authenticate` на 401,
`Retry-After` - компонент, единый `ResourceID`, `Payout.invoice_id` для
возвратов, опциональные `currency/network` в возврате, `pending`-пояснение,
Billing API -> Notification в CBL-000. Принято без изменений: `group_by`
свободной строкой (валидация 11.4), `unknown_report_type` через маппинг
INVALID_ARGUMENT (11.4/7.4), вторая реализация «секрет один раз» для чтений.

| Роль | Ключевые замечания | Закрытие |
| :-- | :-- | :-- |
| Без контекста | MAJOR: 422 у отчета без сценария; scope Idempotency-Key (метод+путь+тело); MINOR/NIT: 503-описания, CBL-004/AGENTS-статусы, `Asset` closed, rate-паттерн, реплей секрета, 503 у чтений и cancel, дубли id-параметров, формулировки | 422 убран; отпечаток и скоупинг; описания и коды выровнены; статусы в коммите; `Asset` открыт; паттерн; реплей-семантика; 503 добавлен; `ResourceID`; поправлено |
| Архитектуры | BLOCKER: `/balances`/`/ledger/entries` нереализуемы на текущем Ledger (owner, списки, pending, period); 201 счета без источника `pay_amount/quote`; MAJOR: проекция `merchant_order_id/metadata/фильтры`, детали идемпотентности (TTL/не-2xx/канонизация/скоуп), закрытые enum, webhook-жизненный цикл, тег, checkout, creating, 503; MINOR: `invoice_id` возврата, валюта возврата, 404-mapping, 409-webhook, `request_id`, дефолт срока; NIT: TTL | задачи 2.10/4.8 (FR-8); FR-8; проекция - 7.4/4.8; FR-3; `x-extensible-enum`; trade-off 8.1/13.2; тег cf-3.1; 7.6 отдельно; создание описано; 503 выровнен; `Payout.invoice_id`; `currency/network`; 11.4/7.4; код добавлен; required; 900 с; ссылка |
| С контекстом | MAJOR: тег версии и статусы; MINOR: FR-7 частично, `Asset` closed, 503-неоднородность; NIT: inline-дубли, стиль описаний, §1 AGENTS | тег cf-3.1 и правило; статусы в коммите; `Asset` открыт; 503 выровнен; единый `ResourceID`/RetryAfter; описания добавлены; §1 обновлен |
| Безопасности | BLOCKER: скоупинг ключа и состав PG-индекса (кросс-мерчантный реплей с секретом); MAJOR: SSRF per-attempt/IPv6, ротация секрета (trade-off); MINOR: ownership-first, символы Idempotency-Key, maxLength денег/rate, `group_by`-whitelist, паттерны id, url-пробелы, кодировка секрета, `detail`, лимиты metadata; NIT: 401-заголовок, 503, описания | `(merchant_id, idempotency_key)`, отпечаток; SSRF-норматив; trade-off зафиксирован; 404-first в 7.4; pattern ключа; maxLength/паттерны; whitelist; паттерны; url-паттерн; hex; `detail`; лимиты; `WWW-Authenticate`; 503; выровнено |

## 8. Открытые вопросы

- Валидатор OpenAPI не подключен (KI-P1, 0.2): проверка вручную (YAML, ссылки,
  маршруты).
- Каталог `report_type` и форма `rows` финализируются 11.4 (аддитивно);
  `unknown_report_type` требует явного маппинга INVALID_ARGUMENT - 11.4/7.4.
- Коды ошибок аддитивны: 7.x добавляет коды без ломающих изменений.
- Ротация/удаление webhook-эндпоинта и секрета - аддитивное расширение 8.1/13.2:
  до его появления компрометация секрета лечится регистрацией нового эндпоинта
  и ops-процедурой (trade-off 1.0.0).
- Предпосылки 7.4 (2.10, 4.8) выполняются до старта 7.4; пакеты cf-3.x
  выпускаются по мере готовности.

## 9. План работ и коммит

Финализация `billing.v1.yaml`, правки `CBL-003`/`CBL-004`/`CBL-000`/AGENTS,
четыре ревью-прохода, коммит `CBL-7.1: финализация OpenAPI v1 мерчанта`,
аннотированный тег `cbl-protos@cf-3.1`.
