# Repository Guidelines

## Project Structure & Module Organization

SeaweedFS is a Go module (`github.com/seaweedfs/seaweedfs`). The main executable and most production code live under `weed/`: commands are in `weed/command`, servers in `weed/server`, storage logic in `weed/storage`, and protocol definitions/generated bindings in `weed/pb`. Unit tests sit beside their packages as `*_test.go`. Broader integration and S3 scenarios are under `test/`. Deployment resources live in `docker/` and `k8s/charts/seaweedfs/`; diagrams and documentation assets are in `note/`. Auxiliary examples and clients are collected in `other/`.

## Build, Test, and Development Commands

- `make install` builds and installs the `weed` binary from `weed/`.
- `make full_install` enables all supported optional backends through Go build tags.
- `make test` runs the full tagged Go test suite used by CI.
- `go test ./weed/storage/...` runs a focused package subtree during development.
- `go test ./weed/storage -run TestName` runs one test by name.
- `make server` installs and starts a local all-in-one server with filer and S3 endpoints; it expects the configuration in `docker/compose/s3.json`.

Use Go 1.24 or the toolchain declared in `go.mod`. Some full-tag builds require native libraries or external services; use focused, untagged tests when those integrations are unrelated.

## Coding Style & Naming Conventions

Format every changed Go file with `gofmt` (tabs are canonical) and organize imports using standard Go conventions. Keep package names short and lowercase, exported identifiers in `PascalCase`, and internal identifiers in `camelCase`. Follow surrounding naming and error-handling patterns; avoid broad refactors in feature or bug-fix changes. Treat generated protobuf files as generated artifacts—update their source definitions and regeneration workflow instead of hand-editing bindings.

## Testing Guidelines

Use Go's `testing` package. Name files `*_test.go`, tests `TestXxx`, and benchmarks `BenchmarkXxx`; table-driven cases are preferred for multiple inputs. Add tests beside the affected package and cover regressions whenever practical. Before submitting, run focused tests, then `make test` when the required optional dependencies are available. Chart changes should pass `ct lint` and `ct install` against `k8s/charts`.

## Commit & Pull Request Guidelines

Recent commits favor concise, imperative summaries, often with an issue/PR reference, for example `Fix sftp performance (#6792)`; dependency automation uses `chore(deps): ...`. Keep each commit scoped to one logical change. Pull requests must explain the problem, the solution, and how it was tested, following `.github/pull_request_template.md`. Link relevant issues, add unit tests when possible, and include screenshots or rendered output for UI, dashboard, or Helm-facing changes.
