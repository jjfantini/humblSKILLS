.PHONY: registry registry-check sync-adapters adapters-check test vet tidy build

ROOT := $(CURDIR)
ADAPTERS_SRC := $(ROOT)/adapters
ADAPTERS_DST := $(ROOT)/cli/internal/platform/builtin

# Regenerate registry.json at the repo root from skills/ + adapters/.
registry:
	go -C cli run ./cmd/build-registry \
		--skills-dir=$(ROOT)/skills \
		--adapters-dir=$(ROOT)/adapters \
		--out=$(ROOT)/registry.json

# Fail if registry.json is out of sync with skills/ + adapters/.
registry-check:
	go -C cli run ./cmd/build-registry \
		--skills-dir=$(ROOT)/skills \
		--adapters-dir=$(ROOT)/adapters \
		--out=$(ROOT)/registry.json \
		--check

# Mirror adapters/ into the CLI's embed directory so the binary stays self-
# contained. Canonical source lives at ./adapters; the CI drift check below
# ensures the two copies never diverge.
sync-adapters:
	@rm -rf $(ADAPTERS_DST)
	@mkdir -p $(ADAPTERS_DST)
	@cp $(ADAPTERS_SRC)/*.yaml $(ADAPTERS_DST)/
	@echo "synced $(ADAPTERS_SRC) -> $(ADAPTERS_DST)"

# Fail if the embedded adapters have drifted from ./adapters.
adapters-check:
	@diff -r $(ADAPTERS_SRC) $(ADAPTERS_DST) >/dev/null 2>&1 || { \
		echo "embedded adapters drift — run 'make sync-adapters'"; \
		diff -r $(ADAPTERS_SRC) $(ADAPTERS_DST) || true; \
		exit 1; \
	}

build:
	go -C cli build -o $(ROOT)/bin/humblskills ./cmd/humblskills

test:
	go -C cli test ./...

vet:
	go -C cli vet ./...

tidy:
	go -C cli mod tidy
