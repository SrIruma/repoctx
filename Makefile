# repoctx build orchestration.
#
# Usage:
#   make build                 # bin/repoctx (local dev build)
#   make test                  # go test ./...
#   make install               # go install -> $GOBIN/repoctx
#   make release VERSION=v0.1.0  # dist/ binaries + SHA256SUMS.txt
#   make clean

GO      ?= go
VERSION ?= dev

# Version is injected at build time; see internal/cli/root.go.
LD_FLAGS = -X github.com/SrIruma/repoctx/internal/cli.version=$(VERSION)

# Release targets. Override with PLATFORMS="linux/amd64" to narrow a build.
PLATFORMS ?= linux/amd64 linux/arm64 windows/amd64

.PHONY: all build test install release clean

all: build

build: ## Build a local dev binary.
	@mkdir -p bin
	$(GO) build -o bin/repoctx ./cmd/repoctx
	@echo "build: bin/repoctx"

test: ## Run the test suite.
	$(GO) test ./...

install: ## Install the tool globally via go install.
	$(GO) install -ldflags="$(LD_FLAGS)" ./cmd/repoctx
	@echo "installed repoctx $(VERSION)"

# Output file for a platform tuple. Windows binaries get a .exe suffix.
release: ## Cross-compile a versioned release into dist/.
	@test "$(VERSION)" != "dev" || (echo "release requires VERSION (e.g. VERSION=v0.1.0)" && exit 1)
	@mkdir -p dist
	@rm -f dist/repoctx_* dist/SHA256SUMS.txt
	@for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; \
		if [ "$$os" = "windows" ]; then name="repoctx_$${os}_$${arch}.exe"; else name="repoctx_$${os}_$${arch}"; fi; \
		echo "building repoctx for $$os/$$arch"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 $(GO) build \
			-ldflags="-s -w $(LD_FLAGS)" \
			-o dist/$$name \
			./cmd/repoctx; \
	done
	@cd dist && sha256sum repoctx_* > SHA256SUMS.txt
	@echo "release: dist/ ($(VERSION))"
	@cd dist && sha256sum -c SHA256SUMS.txt

clean: ## Remove local build artifacts.
	rm -rf bin dist
