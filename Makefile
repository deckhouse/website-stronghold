# Makefile for the Hugo website

PLATFORM_NAME := $(shell uname -p)
ifneq ($(filter arm%,$(PLATFORM_NAME)),)
	export WERF_PLATFORM=linux/amd64
endif

.DEFAULT_GOAL := help

# Tools / variables (can be overridden on the command line)
HUGO ?= hugo
BIND ?= 0.0.0.0
SERVE_FLAGS ?= --cleanDestinationDir --bind=$(BIND)
HUGOFLAGS ?= --minify
MARKDOWNLINT_VERSION ?= v0.45.0
WERF ?= werf
WERF_PLATFORM ?= linux/amd64

CURRENT_UID ?= $(shell id -u)
CURRENT_GID ?= $(shell id -g)
PORTS_TO_FREE ?= 80 1313 1314

PRODUCT_CODE ?= $(shell awk '/^  productCode:/ {print tolower($$2); exit}' config/_default/hugo.yaml)

##@ Main

up: ## Start documentation (available at http://localhost and http://ru.localhost)
	@$(MAKE) down
	@$(MAKE) free-ports
	@UID=$(CURRENT_UID) GID=$(CURRENT_GID) docker compose up

serve: ## Start Hugo dev server (hugo serve --cleanDestinationDir)
	$(HUGO) serve $(SERVE_FLAGS)

build: ## Build the site to ./public
	@echo "Building site to ./public..."
	$(HUGO) $(HUGOFLAGS)

pdf: ## Build the site and generate PDF+DOCX exports via werf
	##~ Output: public/{en,ru}/documentation/downloads/print/<productCode>.{pdf,docx}
	##~ Need external registry (e.g. export WERF_REPO=localhost:4999/docs) to run.
	@echo "Building print-artifacts via werf..."
	@$(WERF) build print-artifacts
	@echo "Extracting PDF/DOCX ($(PRODUCT_CODE)) to ./public/{en,ru}/documentation/downloads/print/..."
	@IMG=$$($(WERF) stage image print-artifacts) && \
	  CID=$$(docker create $$IMG) && \
	  trap "docker rm $$CID >/dev/null" EXIT && \
	  mkdir -p ./public/en/documentation/downloads/print ./public/ru/documentation/downloads/print && \
	  docker cp $$CID:/out/en/documentation/downloads/print/. ./public/en/documentation/downloads/print/ && \
	  docker cp $$CID:/out/ru/documentation/downloads/print/. ./public/ru/documentation/downloads/print/
	@echo "Done. Files: public/{en,ru}/documentation/downloads/print/$(PRODUCT_CODE).{pdf,docx}"

down: ## Stop and remove documentation containers
	docker compose down --remove-orphans

##@ Linters

lint-markdown: ## Lint markdown files
	@echo "Linting markdown files..."
	@docker run --rm -v "$(PWD):/workdir" -w /workdir ghcr.io/igorshubovych/markdownlint-cli:$(MARKDOWNLINT_VERSION) "**/*.md" -c markdownlint.yaml

lint-markdown-fix: ## Lint and auto-fix markdown files
	@echo "Fixing markdown files..."
	@docker run --rm -v "$(PWD):/workdir" -w /workdir ghcr.io/igorshubovych/markdownlint-cli:$(MARKDOWNLINT_VERSION) "**/*.md" -c markdownlint.yaml --fix

##@ Helpers

mod: ## Clean up Hugo modules (hugo mod tidy)
	@echo "Cleaning up Hugo modules..."
	$(HUGO) mod tidy

free-ports: ## Stop containers using known dev ports ($(PORTS_TO_FREE))
	@containers="$$(for port in $(PORTS_TO_FREE); do docker ps -q --filter "publish=$$port"; done | sort -u)"; \
	if [ -n "$$containers" ]; then \
		echo "Stopping containers using ports $(PORTS_TO_FREE): $$containers"; \
		docker stop $$containers; \
	fi

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {\
		FS = ":.*?## "; \
	} \
	/^##@/                    { printf "\n%s\n", substr($$0, 5) } \
	/^[a-zA-Z0-9_-]+:.*?## / { printf "  %-20s %s\n", $$1, $$2 } \
	/^.?.?##~/              { printf "  %-20s %s\n", "", substr($$1, 6) }' $(MAKEFILE_LIST)

.PHONY: help up serve build pdf down lint-markdown lint-markdown-fix mod free-ports
