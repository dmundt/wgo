# wgo development targets. Mirrors the CI pipeline.

GO        ?= go
GOLANGCI  ?= golangci-lint
GOLANGCI_VERSION ?= v2.12.2
FUZZTIME  ?= 15s

.PHONY: all fmt vet build test race lint bench fuzz golden parity tidy coverage clean

all: fmt vet test

fmt:
	gofmt -l cmd pkg

vet:
	$(GO) vet ./...

build:
	$(GO) build ./...

tidy:
	$(GO) mod tidy

test:
	$(GO) test ./...

# Race detector + full golden suite.
race:
	$(GO) test -race ./...

lint:
	$(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_VERSION) run

coverage:
	$(GO) test -coverpkg=./... -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

bench:
	$(GO) test -bench=. -benchmem ./...

# Run the parser fuzzers briefly. Add FUZZTIME=0 to run forever.
fuzz:
	$(GO) test ./pkg/parser -run '^$$' -fuzz=FuzzLoadYAML -fuzztime=$(FUZZTIME)
	$(GO) test ./pkg/parser -run '^$$' -fuzz=FuzzRun -fuzztime=$(FUZZTIME)

# Golden outputs (svg/png/html) require Graphviz `dot` on PATH.
golden:
	$(GO) test ./pkg/parser/ -run TestGoldenOutputs -count=1

# Compare against the reference WireViz 0.4.1 (requires python + pip).
parity:
	$(GO) build -o wgo-bin ./cmd/wgo
	pip install -q wireviz==0.4.1
	@set -eu; for yml in testdata/*.yml; do \
	  name="$$(basename "$$yml" .yml)"; \
	  ref="$$(mktemp -d)"; got="$$(mktemp -d)"; \
	  wireviz -f ghtsp -O "$$name" -o "$$ref" "$$yml" >/dev/null 2>&1; \
	  ./wgo-bin -f ghtsp -O "$$name" -o "$$got" "$$yml" >/dev/null 2>&1; \
	  for ext in gv bom.tsv svg png html; do \
	    if ! cmp -s "$$ref/$$name.$$ext" "$$got/$$name.$$ext"; then \
	      echo "MISMATCH: $$name.$$ext"; diff "$$ref/$$name.$$ext" "$$got/$$name.$$ext" | head -40 || true; \
	      rm -rf "$$ref" "$$got"; rm -f wgo-bin; exit 1; \
	    fi; \
	  done; \
	  rm -rf "$$ref" "$$got"; \
	done; echo "All outputs match reference WireViz 0.4.1."; rm -f wgo-bin

clean:
	rm -rf coverage.out coverage.out.tmp wgo-bin
