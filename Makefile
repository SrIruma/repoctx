# repoctx build orchestration.
#
# Usage:
#   make build                 # bin/repoctx (local dev build)
#   make test                  # go test ./...
#   make install               # go install -> $GOBIN/repoctx
#   make release VERSION=v0.1.0  # dist/repoctx_{linux_amd64,linux_arm64}
#   make clean

GO      ?= go
VERSION ?= dev

# Version is injected at build time; see internal/cli/root.go.
LD_FLAGS = -X github.com/SrIruma/repoctx/internal/cli.version=$(VERSION)

PLATFORMS = linux/amd64 linux/arm64

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

release: ## Cross-compile a versioned release into dist/.
	@test "$(VERSION)" != "dev" || (echo "release requires VERSION (e.g. VERSION=v0.1.0)" && exit 1)
	@mkdir -p dist
	@for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; \
		echo "building repoctx for $$os/$$arch"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 $(GO) build \
			-ldflags="-s -w $(LD_FLAGS)" \
			-o dist/repoctx_$${os}_$${arch} \
			./cmd/repoctx; \
	done
	@echo "release: dist/ ($(VERSION))"

clean: ## Remove local build artifacts.
	rm -rf bin dist
