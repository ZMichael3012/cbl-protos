# cbl-protos

Контракты платформы крипто-биллинга CBL: Protobuf-схемы gRPC и Kafka-событий,
черновики OpenAPI v1, правила совместимости (buf).

Правила и конвенции: `AGENTS.md`, глобальные спеки `../../AGENTS.md` и
`../../.specs/CBL-003-contracts.md`.

## Команды

```powershell
# buf lint через Docker
docker run --rm -v "%CD%":/w -w /w bufbuild/buf:1.47.2 lint

# генерация стабов (плагины в PATH: %USERPROFILE%\go\bin)
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
go run github.com/bufbuild/buf/cmd/buf@v1.47.2 generate

# сборка
go mod tidy
go build ./...
```

## Совместимость

Изменения только обратно совместимые (BACKWARD_TRANSITIVE). Номера полей не переиспользуются,
удаленные поля помечаются `reserved`. Ломающие изменения - новая мажорная версия пакета
и синхронный переход всех потребителей (точки CF-1 - CF-3 в `../../.specs/CBL-004-global-plan.md`).
