# Generated Go SDK types

Auto-generated from `mashgate/contracts/proto/v1/*.proto` через:
- `protoc + protoc-gen-openapi` → `openapi.yaml`
- `oapi-codegen` → `types.gen.go` + `client.gen.go` (package `mashgatev1`)

Клиент покрывает КАЖДУЮ операцию спецификации — покрытие равно контракту по
построению, а не тому, успел ли кто-то написать обёртку руками.

**Не редактируй файлы вручную.** Регенерация: `make sdk-gen-go` в mashgate.

Pipeline: `mashgate/scripts/sdk-gen-go.sh`.
