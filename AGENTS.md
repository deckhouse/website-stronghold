# AGENTS.md for Deckhouse Stronghold

This repository contains Hugo-based documentation for the Deckhouse Stronghold.

## Run the docs site locally

```bash
make up      # start via Docker Compose
make down    # stop and remove containers
```

- EN: <http://localhost/products/stronghold/documentation/>
- RU: <http://ru.localhost/products/stronghold/documentation/>

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

The normative style guide and glossary are at **<https://pp.flant.ru/llms.txt>**.  
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
{{< tab name="Tab 1" >}}Content{{< /tab >}}
{{< tab name="Tab 2" >}}Content{{< /tab >}}
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

Full glossary: <https://pp.flant.ru/llms.txt> (see "Glossary" section).

Key rules:
- `K8s` (uppercase K)
- Use "узел", not "нода"
- Use "веб-интерфейс"
- Use `IP-адрес`
- Use "файлы cookie"
- Avoid "платформа Deckhouse"; use explicit product names (e.g. Deckhouse Platform)
- Module name must match `module.yaml.name` exactly (lowercase kebab-case): `stronghold`, not `Stronghold`
- Do not translate product names and abbreviations that the glossary keeps in EN (e.g. `RBAC`)

Mixed EN/RU compound terms use a hyphen: `S3-бакет`, `managed-сервис`, `master-узел`, `HTTP-протокол`.

### Stronghold product and edition names

- The base product is **Stronghold** (corresponds to Vault CE); the enterprise
  build is **Stronghold EE**.
- Do not use bare `CE`/`EE` as product names. Write `Stronghold` or
  `Stronghold EE`. When you need the base edition explicitly, write
  «базовый Stronghold» / "base Stronghold".
- `Vault CE` / `Vault Enterprise` are allowed only when referring to upstream
  HashiCorp Vault (for example, an external KV store for KV1/KV2 replication).
- Stronghold CLI examples use `d8 stronghold ...`.

### Replication terminology (EN kept)

Keep these established terms in English (both languages), do not translate:
`primary`, `secondary`, `standby`, `active node`, `promote`, `demote`, `WAL`,
`Raft`, `seal`, `seal wrap`, `mount`, `token store`, `master`/`slave`,
`performance standby`. Diagram labels on images (`Data WAL Streaming`,
`Forward Writes`, `R/W`, and similar) are kept in English to match the pictures.

### Russian translations of common terms

- `entity` → «сущность»; `identity` (the subsystem) →
  «идентификационные данные (Identity)».
- `identity` in the sense of a cluster's identity → «идентификатор кластера».
- Do not call standby nodes «неактивные»: use «standby-узлы» or
  «неведущие узлы HA», because they still serve reads under performance standby.
- «seal material» → «данные, зашифрованные через Seal»; the wrapper block on
  diagrams is «внешний блок Sealwrapper», not «блок сбоку».

## Writing rules (digest of ai-tools)

Source: the `deckhouse-writing` and `deckhouse-product-site` plugins of the `ai-tools` repository.
If a rule below conflicts with `GLOSSARY.ru.md`, this section wins. The glossary still covers terms that this section does not mention.

### Sources and precedence

1. The normative style guide and glossary: <https://pp.flant.ru/llms.txt>. Fetch it before a review and when a question is not covered here.
1. This file.
1. The theme contract of `hugo-web-product-module` (`AUTHORING.md`) for shortcodes, page parameters and render hooks.
   Use only the shortcodes it lists, with their exact parameters. Never use Liquid tags.

### Content

- Write only what changes what the reader does. Leave out internal design, development process, history and workarounds.
- Describe current behaviour only. Do not describe features that do not exist yet or the absence of a planned feature.
- Verify every statement about product behaviour against the product source. If the code and the text disagree, the text follows the code; report the disagreement.
- Interface element names are the strings the product shows: English captions on `.md` pages, Russian captions on `.ru.md` pages.
- Make every section self-contained and describe each fact once. Add a cross-link only where the task cannot be completed without it.
- Do not link to the product source code.

### Wording

- No verdicts: «безопасно», «рекомендуется», «удобно», «просто», "simply", "easy", "just". State the behaviour and the action.
- No jargon or anglicisms outside the glossary: «гейт», «скоуп», «маппинг», «нативный», «из коробки», «обвязка».
- Do not put a count in front of an enumeration («состоит из двух компонентов: X и Y»). Name the things.
- Write the affirmative form. Avoid «не X, а Y», «не только X, но и Y», double negation.
- No meta sentences («на этой странице описано…», "this section describes").
- Put the main idea of a paragraph in its first sentence. Split long sentences.
- Expand an abbreviation at its first occurrence on a page: «Deckhouse Platform (DP)».
- State requirements in full sentences, not «нужно», «надо», «не забудьте».
- Explain an example or leave it out. Name an example after the task it solves, start with the minimal working case, and describe the expected result.

### Russian specifics

- Imperative вы-form in instructions; no «мы», «наш», «пожалуйста»; neutral technical register («развернуть», not «поднять»; «вручную», not «ручками»).
- Use the letter «ё» and «ёлочки» quotes only. Put section and interface element names in quotes: раздел «Обзор», кнопка «Сохранить».
- Em dash (`—`) only where a predicate is omitted, in «термин — определение» and before a generalising word; `–` without spaces in numeric ranges; `→` for navigation in the web interface, not `->`.
- Put a colon before a list, table or code block, but not instead of a conjunction.
- Do not write «в т. ч.», «т. н.», «т. е.», «т. к.», «см.», «кол-во» (the last one is allowed in tables).
- List items that continue the lead-in start with a lowercase letter and end with `;` (the last one with `.`); items that are sentences start with a capital letter and end with `.`.
- In a «term — explanation» item, put an em dash after the term and start the explanation with a lowercase letter.
- Include prepositions and dependent words in link text: «[в руководстве по установке](...)».

### English specifics

- Imperative in instructions; no "please", "we", "our", exclamation marks.
- Sentence case in headings; prefer the "-ing" form for actions ("Configuring access").
- Straight quotes; product names without quotes. In a "term: explanation" item, put a colon after the term and start the explanation with a capital letter.
- Omit the article at the start of a parameter description and in link text.

### Terminology

- Product names: **Deckhouse Platform (DP)** is the current name; Deckhouse Kubernetes Platform (DKP) is the former one. Use DP in prose; keep DKP in URL paths, identifiers, file names, command output, code and published release notes.
- Do not write «платформа Deckhouse» or bare «Deckhouse» for the product.
- Name an instance by its kind: «создайте NodeGroup», not «создайте ресурс NodeGroup». Use «объект» for instances as a noun and «ресурс» only for the kind («кастомный ресурс», «Ingress-ресурс»).
- Terms that are often written wrong:

  | English | Russian | Incorrect forms |
  | --- | --- | --- |
  | node | узел | нода |
  | pod | под | — |
  | namespace | неймспейс | пространство имён |
  | cache | кеш | кэш |
  | hash | хеш | хэш |
  | timeout | тайм-аут | таймаут |
  | snapshot | снимок | снапшот, снэпшот |
  | directory | директория | папка |
  | endpoint | эндпоинт | эндпойнт, конечная точка |
  | label | лейбл | тег, метка |
  | tag | тег | тэг |
  | dashboard | дашборд | дэшборд |
  | job | задача | джоба |
  | ID | идентификатор | id |
  | symbolic link | символическая ссылка | симлинк, symlink |
  | self-signed | самоподписанный | самоподписной |
  | offline | офлайн | оффлайн |
  | reverse proxy | обратный прокси | reverse proxy |
  | production cluster | production-кластер | продуктивный кластер |
  | clusters, servers | кластеры, серверы | кластера, сервера |

- Control plane, data plane, bare metal, taints, tolerations, RBAC, CRD and ServiceAccount stay in English.
- Use one form of a term on every page. Check headings, alerts, tables, examples and code comments, not only body text.
- Use `d8 k` instead of `kubectl`; keep `kubectl` only when you name the utility itself.

### Markdown

- Headings: H2–H4 only; no inline code in headings; no full stop; a question mark only in FAQ headings.
- Follow every heading, list, table and code block with an introductory sentence; do not start a section with a list, table, code block or nested heading.
- Use a list for three or more items; write two items as a sentence. Introduce every list with a lead-in phrase and a blank line.
- Use a numbered list only for steps, with `1.` for every item.
- Inline code: module names, parameters and values, states, commands, file paths, ports, HTTP codes and headers, endpoints. Not for Kubernetes kinds (NodeGroup, ModuleConfig) and not for values inside YAML.
- Code blocks: a language tag from the Chroma list; `shell` or `bash` for commands, `console` or `text` for output; one-line commands without `\` continuations.
  Placeholders: `<CLUSTER_NAME>`, explained after the block. Comments: sentences on a separate line before the element.
- Alerts: only for information the reader must not miss; at most two in a row; levels `info`, `warning`, `danger`.
- Tabs: write tab sets as `{{< tabs name="…" >}}` with `{{< tab name="…" >}}` in new or changed markup; every tab set has a name that is unique within the page.
- Accordions: for optional content, with a self-explanatory title (not «Подробнее» or «Пример»).
- Links: text says where the link leads, five words or fewer; no «здесь», «тут», «по ссылке», «см.».
  To enable a module, link to the `enable` section of its configuration page instead of writing your own steps.
- Tables: headings in the nominative singular; no full stop at the end of a cell; `—` for an empty value.
- Images: alt text with a capital letter, no full stop, in the language of the page.
- Break lines longer than 500 characters by meaning (after a sentence or a clause).

### Page structure

- Typical sections have the same names across products: «Обзор» / "Overview", «Ограничения» / "Limitations", «Доступность в редакциях» / "Availability in editions", «Дополнительные материалы» / "Additional materials".
- Overview page: what the topic is and why it is needed, its place in the product, the main principles, navigation. Do not list every object and scenario.
- Guide: quick start first, then the main resource, groups of parameters, the result, troubleshooting (exact error messages, cause, fix). Collect limits in one table.
- Examples: on a separate page, one H2 and one task per example, from simple to complex, long manifests in an accordion.
- Release notes: sections per version from newest to oldest, 2–5 key changes first, then groups «Новые возможности», «Улучшения», «Исправления», «Безопасность», «Несовместимые изменения», «Рекомендации по обновлению», «Известные ограничения», «Документация», «Зависимости». Do not edit notes of released versions.

### Review

Review order: accuracy and content, wording, terminology, formatting, links, engine and structure, EN/RU parity.
Severity: `critical` (incorrect or dangerous guidance, contradiction with the code, broken build), `major` (strong rule violation, missing context, broken EN/RU parity), `minor` (wording and formatting).
Before you say the work is done: run `make lint-markdown`, check EN/RU heading parity, and check every relative link you added in the rendered page.

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
