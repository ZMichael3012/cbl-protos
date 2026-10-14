# AGENTS.md - cbl-protos

Репозиторий контрактов платформы CBL: Protobuf-схемы gRPC и Kafka-событий, OpenAPI-черновики,
правила совместимости (buf). Обязательные глобальные правила: `../../AGENTS.md`,
`../../.specs/CBL-001-development-process.md`, `../../.specs/CBL-003-contracts.md`.
Длинное тире (U+2014) запрещено, используется только дефис.

## 1. Назначение и границы

In scope: `.proto`-файлы, OpenAPI-черновики, `buf`-конфигурация, сгенерированные Go-стабы
(`gen/`), правила совместимости.

Out of scope: бизнес-логика сервисов, инфраструктура Schema Registry (задача 0.5), финализация
OpenAPI мерчанта (7.1) и admin (12.1).

Точки заморозки контрактов: CF-1 (Payment, Quote, Ledger, Connector, события gateway),
CF-2 (`payment.events`, `ledger.events`, `payout.commands`, `gateway.payout.events`),
CF-3 (Notification, Reporting, `recon.alerts`, Admin API). См. `CBL-004`.

## 2. Контракты и потребители

| Артефакт | Потребители |
| :-- | :-- |
| `billing/payment/v1` | `cbl-payment`, `cbl-billing-api`, `cbl-reconciler` |
| `billing/quote/v1` | `cbl-quote`, `cbl-payment`, `cbl-billing-api` |
| `billing/ledger/v1` | `cbl-ledger`, `cbl-billing-api`, `cbl-payout`, `cbl-reconciler` |
| `billing/connector/v1` | `cbl-connector`, `cbl-payment`, `cbl-payout`, `cbl-reconciler` |
| `billing/payout/v1` | `cbl-payout`, `cbl-billing-api`, `cbl-backoffice` |
| `billing/reporting/v1` | `cbl-reporting`, `cbl-billing-api` |
| `billing/events/v1` | все продюсеры и консьюмеры Kafka (`CBL-003`, раздел 6) |
| `openapi/billing.v1.yaml` | `cbl-billing-api`, `cbl-api-gateway` |
| `openapi/admin.v1.yaml` | `cbl-backoffice`, `cbl-api-gateway` |

## 3. Архитектура

```
cbl-protos/
├── buf.yaml                # v2, lint STANDARD, breaking FILE
├── buf.gen.yaml            # protoc-gen-go, protoc-gen-go-grpc -> gen/go
├── proto/billing/...       # доменные пакеты v1
├── openapi/                # черновики OpenAPI v1
├── gen/go/                 # сгенерированные стабы (коммитятся)
└── Makefile
```

Модуль: `github.com/cbl/cbl-protos`; `go_package`:
`github.com/cbl/cbl-protos/gen/go/billing/<domain>/v1;<domain>v1`.

Правила эволюции: номера полей не переиспользуются; удаленные поля помечаются `reserved`;
ломающие изменения только через новую мажорную версию пакета и переход всех потребителей
одной волной (`CBL-003`, раздел 8).

## 4. Команды

```powershell
# lint (Docker)
docker run --rm -v "%CD%":/w -w /w bufbuild/buf:1.47.2 lint

# генерация (нужны плагины в PATH: %USERPROFILE%\go\bin)
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
go run github.com/bufbuild/buf/cmd/buf@v1.47.2 generate

# сборка стабов
go mod tidy
go build ./...
```

## 5. Правила разработки

- Каждое изменение: спека задачи в `.specs/`, затем правка proto, `buf lint`, `buf breaking`
  против предыдущего тега, генерация, `go build ./...`.
- Версия контракта фиксируется тегом `cbl-protos@<version>`; номер версии попадает в спеки
  зависимых сервисов.
- Коммит: `CBL-<эпик>.<номер>: <краткое сообщение>`, спека в том же коммите.
- Docstring на русском в стиле godoc для рукописного Go; сгенерированный код исключен.
- Комментарии объясняют почему, а не что; ошибки через `%w`; панки запрещены.
- Запрещены внешние proto-зависимости, кроме well-known types.

## 6. Trade-offs

| ID | Решение | Плюсы | Минусы и компенсация |
| :-- | :-- | :-- | :-- |
| T-P1 | buf v2 как единственный инструмент lint/breaking/generate | единые правила, одна утилита | привязка к версии buf; версия фиксируется в Makefile и CI |
| T-P2 | сгенерированные стабы коммитятся в `gen/` | потребители собираются без кодогенерации | диффы растут; регенерация обязательна в каждом изменении proto |

## 7. Known issues

| ID | Проблема | Влияние | План | Статус |
| :-- | :-- | :-- | :-- | :-- |
| KI-P1 | Валидатор OpenAPI (spectral/redocly) не подключен | опечатки в схемах до появления CI | задача 0.2 | open |
| KI-P2 | Schema Registry подключается отдельно | события проверяются после интеграции | задача 0.5 | open |

## 8. Журнал задач

| CBL | Дата | Статус | Спека | Коммит |
| :-- | :-- | :-- | :-- | :-- |
| CBL-1.7 | 2026-10-13 | done (buf lint, generate, go build, emdash - зеленые) | `.specs/CBL-1.7-protos-and-contract-freeze.md` | ожидает коммита/тега |

## 9. Runbook

Выпуск новой версии контракта:

1. PR с правкой proto и обновлением спеки.
2. `buf lint` и `buf breaking` против предыдущего тега, регенерация стабов.
3. Merge, тег `cbl-protos@<version>`, обновление зависимостей сервисов-потребителей.
4. Обновление `CBL-003` при изменении описательной части, уведомление агентов волны.
