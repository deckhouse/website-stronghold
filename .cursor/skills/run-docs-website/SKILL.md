---
name: run-website
description: Runs the documentation website locally via `Makefile`. Use when the user wants to run the docs, preview the documentation site, start the docs server, or open the documentation website locally.
---

# Run Documentation Website

## Quick start

1. **Run in the repository root**:
   ```bash
   make up
   ```

3. **Open** in a browser:
   - EN: http://localhost/products/stronghold/documentation/
   - RU: http://ru.localhost/products/stronghold/documentation/

1. **Stop** when done:
   ```bash
   make down
   ```

## All targets

| Target | Use case |
|--------|---------|
| `make up` | Start docs via Docker Compose |
| `make down` | Stop and remove containers (`docker compose down --remove-orphans`) |
| `make serve` | Start Hugo dev server locally without Docker (`hugo serve`) |
| `make build` | Build the site to `./public` |
| `make lint-markdown` | Lint Markdown files |
| `make lint-markdown-fix` | Lint and auto-fix Markdown files |
| `make mod` | Clean up Hugo modules (`hugo mod tidy`) |
