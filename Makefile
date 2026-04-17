.PHONY: registry registry-check test vet tidy

ROOT := $(CURDIR)

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

test:
	go -C cli test ./...

vet:
	go -C cli vet ./...

tidy:
	go -C cli mod tidy
