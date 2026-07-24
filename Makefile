.PHONY: build test fmt fmt-check vet lint govulncheck check ci

# Build all packages (core + nested deck module).
build:
	go build ./...
	$(MAKE) -C deck build

# Tooling lives in ./tools (separate module) so consumers don't inherit linter deps.
GO_TOOL = go tool -modfile=tools/go.mod

# Format all Go source in place (gofmt + goimports via golangci-lint).
fmt:
	$(GO_TOOL) golangci-lint fmt

# Verify all Go source is formatted; fail (showing the diff) if not.
fmt-check:
	$(GO_TOOL) golangci-lint fmt --diff

# go vet: the standard toolchain analyzers.
vet:
	go vet ./...

# golangci-lint: aggregate static analysis (govet, staticcheck, errcheck, ...).
lint:
	$(GO_TOOL) golangci-lint run

# govulncheck: report known vulnerabilities in dependencies and reachable code.
govulncheck:
	$(GO_TOOL) govulncheck ./...

# Run the test suite (core + nested deck module).
test:
	go test ./...
	$(MAKE) -C deck test

# Full gate: build, format check, lint, vulncheck, test.
check: build fmt-check lint govulncheck test

# CI entrypoint: identical to the full gate.
ci: check
