---
name: run-website
description: Runs the documentation website locally via `Makefile` and werf. Use when the user wants to run the docs, preview the documentation site, start the docs server, or open the documentation website locally.
---

# Run Documentation Website

## Quick start

1. **Run in the repository root**:
   ```bash
   make up
   ```

3. **Open** in a browser:
   - for the English version — http://localhost/products/kubernetes-platform/documentation/v1/
   - for the Russian version — http://ru.localhost/products/kubernetes-platform/documentation/v1/

4. **Stop** when done:
   ```bash
   make down
   ```

## Other targets (from docs/site/)

| Target | Use case                                                                    |
|--------|-----------------------------------------------------------------------------|
| `make up` | Start docs in watch mode that rebuilds on commit                            |
| `make lint-markdown-fix` | Run markdown linter and fix some problems automatically                    |
| `make down` | Stop containers, remove networks, and stop the local registry               |
