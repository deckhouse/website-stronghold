# Makefile for the Hugo website

# Tools / variables (can be overridden on the command line)
HUGO ?= hugo
BIND ?= 0.0.0.0
SERVE_FLAGS ?= --cleanDestinationDir --bind=$(BIND)
HUGOFLAGS ?= --minify
MARKDOWNLINT_VERSION ?= v0.45.0
WERF_PLATFORM ?= linux/amd64

CURRENT_UID ?= $(shell id -u)
CURRENT_GID ?= $(shell id -g)
PORTS_TO_FREE ?= 80 1313 1314

.PHONY: help serve build down lint-markdown lint-markdown-fix mod free-ports

help:
	@echo "Usage: make [target]"
	@echo
	@echo "Common targets:"
	@echo "  up               Start documentation (available at http://localhost and http://ru.localhost)"
	@echo "  serve            Start Hugo dev server (hugo serve --cleanDestinationDir)"
	@echo "  build            Build the site to ./public"
	@echo "  down             Stop and remove documentation containers"
	@echo "  lint-markdown    Lint markdown files"
	@echo "  lint-markdown-fix Fix markdown files automatically"
	@echo "  mod              Clean up Hugo modules (hugo mod tidy)"
	@echo "  help             Show this help"
	@echo
	@echo "Variables (can be overridden):"
	@echo "  HUGO=$(HUGO)"
	@echo "  PORT=$(PORT)"
	@echo "  BIND=$(BIND)"
	@echo "  BASEURL=$(BASEURL)"
	@echo "  MARKDOWNLINT_VERSION=$(MARKDOWNLINT_VERSION)"

up:
	@$(MAKE) down
	@$(MAKE) free-ports
	@UID=$(CURRENT_UID) GID=$(CURRENT_GID) docker compose up

free-ports:
	@containers="$$(for port in $(PORTS_TO_FREE); do docker ps -q --filter "publish=$$port"; done | sort -u)"; \
	if [ -n "$$containers" ]; then \
		echo "Stopping containers using ports $(PORTS_TO_FREE): $$containers"; \
		docker stop $$containers; \
	fi

down:
	docker compose down --remove-orphans

serve:
	$(HUGO) serve $(SERVE_FLAGS)

build:
	@echo "Building site to ./public..."
	$(HUGO) $(HUGOFLAGS)

lint-markdown:
	@echo "Linting markdown files..."
	@docker run --rm -v "$(PWD):/workdir" -w /workdir ghcr.io/igorshubovych/markdownlint-cli:$(MARKDOWNLINT_VERSION) "**/*.md" -c markdownlint.yaml

lint-markdown-fix:
	@echo "Fixing markdown files..."
	@docker run --rm -v "$(PWD):/workdir" -w /workdir ghcr.io/igorshubovych/markdownlint-cli:$(MARKDOWNLINT_VERSION) "**/*.md" -c markdownlint.yaml --fix

mod:
	@echo "Cleaning up Hugo modules..."
	$(HUGO) mod tidy
