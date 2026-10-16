# cbl-protos

Contracts of the CBL crypto-billing platform: Protobuf schemas for gRPC and Kafka events,
OpenAPI v1 drafts, compatibility rules (buf).

Rules and conventions: `AGENTS.md`, the global specs `../../AGENTS.md` and
`../../.specs/CBL-003-contracts.md`.

## Commands

```powershell
# buf lint via Docker
docker run --rm -v "%CD%":/w -w /w bufbuild/buf:1.47.2 lint

# stub generation (plugins in PATH: %USERPROFILE%\go\bin)
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
go run github.com/bufbuild/buf/cmd/buf@v1.47.2 generate

# build
go mod tidy
go build ./...
```

## Compatibility

Changes are backward compatible only (BACKWARD_TRANSITIVE). Field numbers are never reused,
deleted fields are marked `reserved`. Breaking changes - a new major package version
and a synchronous migration of all consumers (freeze points CF-1 - CF-3 in `../../.specs/CBL-004-global-plan.md`).
