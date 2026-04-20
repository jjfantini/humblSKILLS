.PHONY: registry registry-check test vet tidy build eval-mock eval-showcase

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

# Run the eval harness against the bundled use-smart-skill scenarios using
# the mock runner. Emits artifacts to ./.eval-workspace/ so the run is easy
# to inspect without touching the user's persistent workspace.
eval-mock: build
	@rm -rf .eval-workspace
	@HUMBLSKILLS_EVAL_WORKSPACE=$(ROOT)/.eval-workspace \
	  $(ROOT)/bin/humblskills eval run use-smart-skill --runner mock --yes
	@echo "--- artifacts ---"
	@ls $(ROOT)/.eval-workspace/use-smart-skill/iteration-1/

# Open the canonical demo against use-smart-skill using whichever runner
# the environment offers (first CLI or API runner in the registry).
eval-showcase: build
	$(ROOT)/bin/humblskills eval showcase
