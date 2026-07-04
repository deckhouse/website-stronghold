# CLAUDE.md — website-stronghold

This repository contains Hugo-based documentation for the Deckhouse Kubernetes Stronghold.

## Run the docs site locally

```bash
make up      # start via Docker Compose
make down    # stop and remove containers
```

- EN: http://localhost/products/stronghold/documentation/
- RU: http://ru.localhost/products/stronghold/documentation/

All targets:

| Target | Purpose |
|--------|---------|
| `make up` | Start docs via Docker Compose |
| `make down` | Stop and remove containers |
| `make serve` | Start Hugo dev server locally without Docker |
| `make build` | Build the site to `./public` |
| `make lint-markdown` | Lint Markdown files |
| `make lint-markdown-fix` | Lint and auto-fix Markdown files |
| `make mod` | Clean up Hugo modules (`hugo mod tidy`) |

## Editorial policy

The normative style guide and glossary are at **https://pp.flant.ru/llms.txt**.  
Fetch it before writing or reviewing documentation.

## Documentation style

- Write concise technical text with clear user value.
- No first-level headings (`#`) in documentation files. Minimum level is `##`.
- Prefer short paragraphs and structured lists.
- For emphasis, use **bold**; avoid unnecessary italic.
- Keep instructions actionable and testable.
- In command examples use `d8 k` instead of `kubectl`.
- YAML snippets must be syntactically valid.
- Markdown ordered lists: use `1.` for every item (not `1.`, `2.`, `3.`).

### Inline code

Use inline code for: parameters, module names, commands, file paths, HTTP codes.  
Do not overuse it for resource type names in narrative text.  
Do not wrap values inside YAML files in backticks — YAML values are already code context.

### Links

- Use meaningful link anchors (avoid "here" or "тут").
- Links to project pages must be relative.
- Links to Deckhouse Kubernetes Platform docs must be absolute without domain (`/products/kubernetes-platform/documentation/v1/...`).

### Code blocks

Every fenced code block must have an explicit language tag from the Hugo/Chroma list.  
If the needed language is absent, use `text` or `plain`.

Common tags: `bash`, `yaml`, `json`, `go`, `go-html-template`, `text`, `plain`.  
Full list: see `.cursor/rules/docs/hugo-supported-codeblock-languages.mdc`.

### Hugo shortcodes

```go-html-template
{{< alert level="warning" >}}
Message text.
{{< /alert >}}
```

Alert levels: `info`, `warning`, `danger`.

```go-html-template
{{< tabs name="uniq_name" >}}
{{% tab name="Tab 1" %}}Content{{% /tab %}}
{{% tab name="Tab 2" %}}Content{{% /tab %}}
{{< /tabs >}}
```

```go-html-template
{{% details "Summary..." %}}
Markdown content.
{{% /details %}}
```

### Release notes (`content/documentation/release-notes/`)

- Under a heading, if the only body content is a single bullet — write plain prose, not a list.
- Keep lists only when there are two or more top-level bullets under the same heading.

## Terminology

Full glossary: https://pp.flant.ru/llms.txt (see "Glossary" section).

Key rules:
- `K8s` (uppercase K)
- Use "узел", not "нода"
- Use "веб-интерфейс"
- Use `IP-адрес`
- Use "файлы cookie"
- Avoid "платформа Deckhouse"; use explicit product names (e.g. Deckhouse Kubernetes Platform)
- Module name must match `module.yaml.name` exactly (lowercase kebab-case): `stronghold`, not `Stronghold`
- Do not translate product names and abbreviations that the glossary keeps in EN (e.g. `RBAC`)

Mixed EN/RU compound terms use a hyphen: `S3-бакет`, `managed-сервис`, `master-узел`, `HTTP-протокол`.

## Russian text style

In Russian instructional text, always use the imperative вы-form:
- «создайте», not «создаем»
- «нажмите», not «нажимаем»
- «перейдите», not «переходим»

This applies to all step-by-step instructions, quickstart guides, and procedural text.

## Front matter

Required fields: `title`, `description` (concise, unique, not a copy of title).

- Related links: use `params.relatedLinks` in front matter; do not add manual "Related links" sections in the Markdown body.

## EN/RU parity

- Russian files use the `.ru.md` suffix only (not `.RU.md`, `_RU.md`).
- For each change relevant to both languages, update both files.
- Do not leave EN/RU pairs semantically diverged unless the change is intentionally language-specific.
- Localized media: `image1.jpg` (EN) / `image1.ru.jpg` (RU).

## Media

Store images, PDFs, and media only in `docs/` or its subdirectories.  
Use relative links for files within the same module repository.
