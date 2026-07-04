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

PLAYWRIGHT_VERSION ?= 1.61.1
PLAYWRIGHT_IMAGE ?= mcr.microsoft.com/playwright:v$(PLAYWRIGHT_VERSION)-noble
HTTP_PORT ?= 8088
PRODUCT_CODE ?= $(shell awk '/^  productCode:/ {print tolower($$2); exit}' config/_default/hugo.yaml)
MODULE_DIR ?= $(PWD)/../hugo-web-product-module

.PHONY: help serve build down lint-markdown lint-markdown-fix mod free-ports pdf

help:
	@echo "Usage: make [target]"
	@echo
	@echo "Common targets:"
	@echo "  up               Start documentation (available at http://localhost and http://ru.localhost)"
	@echo "  serve            Start Hugo dev server (hugo serve --cleanDestinationDir)"
	@echo "  build            Build the site to ./public"
	@echo "  pdf              Build the site and generate PDF+DOCX exports into ./public/{en,ru}/documentation/downloads/print/"
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

pdf: build
	@echo "Generating PDF and DOCX ($(PRODUCT_CODE), EN + RU) via Playwright + Pandoc..."
	@if [ ! -d "$(MODULE_DIR)/.github/scripts" ]; then \
		echo "ERROR: hugo-web-product-module scripts not found at $(MODULE_DIR)/.github/scripts"; \
		exit 1; \
	fi
	@docker run --rm --network host \
		-e PRODUCT_CODE="$(PRODUCT_CODE)" \
		-e NODE_PATH=/deps/node_modules \
		-v "$(PWD):/workdir" \
		-v "$(MODULE_DIR)/.github/scripts:/scripts:ro" \
		-w /workdir \
		$(PLAYWRIGHT_IMAGE) \
		bash -c '\
			set -e; \
			apt-get update && apt-get install -y pandoc ; \
			mkdir -p /deps && cd /deps && npm init -y >/dev/null && npm install --silent --no-audit --no-fund playwright@$(PLAYWRIGHT_VERSION) http-server >/dev/null; \
			cd /workdir; \
			(cd public && /deps/node_modules/.bin/http-server -p $(HTTP_PORT) -s >/tmp/http.log 2>&1 &) ; \
			for i in $$(seq 1 30); do curl -sf http://localhost:$(HTTP_PORT)/ > /dev/null && break; sleep 1; done; \
			node /scripts/print-export.js en http://localhost:$(HTTP_PORT); \
			node /scripts/print-export.js ru http://localhost:$(HTTP_PORT) \
		'
	@echo "Done. Files: public/{en,ru}/documentation/downloads/print/$(PRODUCT_CODE).{pdf,docx}"
