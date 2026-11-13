GO ?= go

# The go fix bundled with Go lacks some analyzers that the standalone
# modernize command has. `make fix` and `make fix-check` run go fix plus
# modernize with only the analyzers in MODERNIZE_EXTRA that go fix doesn't
# list. Once `go tool fix help` lists all of them, the modernize step is
# skipped and MODERNIZE_EXTRA can be emptied.
MODERNIZE_VERSION ?= v0.52.0
MODERNIZE_EXTRA   ?= importcomment reflecttypeassert slicesclip
MODERNIZE         := golang.org/x/tools/go/analysis/passes/modernize/cmd/modernize@$(MODERNIZE_VERSION)

# Shell snippet: set $$flags to "-a -b ..." for each MODERNIZE_EXTRA
# analyzer that go fix doesn't list.
define modernize_flags
available=$$($(GO) tool fix help | awk '/^Registered analyzers:/{f=1;next} /^By default/{f=0} f && NF{print $$1}') || exit 1; \
flags=""; \
for a in $(MODERNIZE_EXTRA); do \
	echo "$$available" | grep -qx "$$a" || flags="$$flags -$$a"; \
done
endef

default: help

fix: ## Apply go fix and modernize
	@echo "make: Applying go fix..."
	@$(modernize_flags); \
	$(GO) fix ./... || exit 1; \
	if [ -n "$$flags" ]; then \
		echo "make: Applying modernize$$flags..."; \
		$(GO) run $(MODERNIZE) -fix $$flags ./... || exit 1; \
	fi

fix-check: ## [CI] Go Fix Check / go fix and modernize
	@echo "make: Checking go fix..."
	@$(modernize_flags); \
	status=0; \
	$(GO) fix -diff ./... || status=1; \
	if [ -n "$$flags" ]; then \
		echo "make: Checking modernize$$flags..."; \
		$(GO) run $(MODERNIZE) $$flags ./... || status=1; \
	else \
		echo "make: go fix now includes all of: $(MODERNIZE_EXTRA). Empty MODERNIZE_EXTRA in GNUmakefile."; \
	fi; \
	if [ $$status -ne 0 ]; then echo "make: To fix, run: make fix"; fi; \
	exit $$status

help: ## Display this help
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}'

.PHONY: default fix fix-check help
