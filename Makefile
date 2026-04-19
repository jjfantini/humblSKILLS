.PHONY: registry registry-check test vet tidy build

ROOT := $(CURDIR)

# Regenerate registry.json at the repo root from skills/ + the adapters
# embedded in the build-registry binary.
registry:
	go -C cli run ./cmd/build-registry \
		--skills-dir=$(ROOT)/skills \
		--out=$(ROOT)/registry.json

# Fail if registry.json is out of sync with skills/ + embedded adapters.
registry-check:
	go -C cli run ./cmd/build-registry \
		--skills-dir=$(ROOT)/skills \
		--out=$(ROOT)/registry.json \
		--check

build:
	go -C cli build -o $(ROOT)/bin/humblskills ./cmd/humblskills

test:
	go -C cli test ./...

vet:
	go -C cli vet ./...

tidy:
	go -C cli mod tidy
