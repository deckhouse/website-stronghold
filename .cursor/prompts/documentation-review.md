Review documentation changes for compliance with this repository's rules.

Context:
- This repository contains Hugo-based product documentation.
- Review the changed documentation files, and inspect nearby related files when needed.
- If the reviewed page describes product or module behavior, verify the statements against the relevant source code, configuration, schemas, examples, or adjacent documentation. Do not guess functionality.
- If a rule is not applicable to the changed files, explicitly skip it instead of forcing a finding.

Primary rule sources:
- `.cursor/rules/docs/global-style.mdc`
- `.cursor/rules/docs/terminology.mdc`
- `.cursor/rules/docs/refs-editorial-policy-full.mdc`
- `.cursor/rules/docs/refs-glossary-full.mdc`
- `.cursor/rules/docs/hugo-supported-codeblock-languages.mdc`

Apply additional rules when relevant:
- `.cursor/rules/docs/ru-en-parity.mdc` for EN/RU pairs, `.ru.md` naming, and localized media.
- `.cursor/rules/docs/hugo-shortcodes.mdc` for `alert`, `tabs`, `details`, and shortcode usage.
- `.cursor/rules/docs/frontmatter-links-media.mdc` for front matter, related links, and media placement.
- `.cursor/rules/docs/crd-translation-files.mdc` for `crds/doc-ru-*.yaml` translation files.
- `.cursor/rules/docs/openapi-x-doc.mdc` for `openapi/**/*.yaml` and `x-doc-*` fields.

Review checklist:

1. Accuracy and consistency
- Check that the documentation matches the actual behavior described by the code, configuration, schema, or examples.
- Flag claims that are unsupported, outdated, misleading, or contradicted by the implementation.
- Flag missing prerequisites, caveats, limits, or version-sensitive behavior when they are necessary for safe use.

2. Editorial and style compliance
- Check compliance with `.cursor/rules/docs/refs-editorial-policy-full.mdc` and `.cursor/rules/docs/global-style.mdc`.
- Check for concise technical wording, meaningful link anchors, correct emphasis, and actionable instructions.
- For Russian instructional text, require imperative "вы"-form.
- For docs pages, do not allow first-level headings (`#`). Minimum heading level is `##`.

3. Terminology compliance
- Check compliance with `.cursor/rules/docs/terminology.mdc` and `.cursor/rules/docs/refs-glossary-full.mdc`.
- Flag forbidden or inconsistent terms.
- Require exact approved product naming.
- Require `d8 k` instead of `kubectl` in command examples unless the repository rule clearly does not apply to that case.

4. Markdown, Hugo, and code examples
- Every fenced code block must have an explicit language tag.
- The language tag must be present in `.cursor/rules/docs/hugo-supported-codeblock-languages.mdc`.
- Markdown ordered lists must use `1.` for every item, not `1.`, `2.`, `3.`.
- Check Hugo shortcode usage when present.
- Check that YAML, JSON, shell, and other examples are syntactically plausible and internally consistent.
- Check that commands, flags, paths, field names, and HTTP codes are formatted as inline code where appropriate.

5. Localization and parity
- RU files must use the `.ru.md` suffix only.
- If a change is relevant to both EN and RU pages, check whether both sides were updated consistently.
- Flag semantic divergence between paired EN/RU pages when it appears unintentional.
- Check localized media naming when media is language-specific.

6. Front matter, links, and structure
- Check required front matter fields when present and applicable.
- Check that links are concise, meaningful, and not broken by obvious path mistakes.
- Flag manual "related links" sections when front matter rules require structured related links instead.

7. Schema-specific rules
- For CRD translation files, check path naming, translation scope, and YAML structure.
- For OpenAPI docs, check `x-doc-*` usage, example structure, and description conventions.

Severity model:
- `critical`: incorrect or dangerous guidance, a behavioral contradiction, or a change that can cause user harm or a broken workflow.
- `major`: a strong rules violation, missing important context, broken parity, or a likely user-facing documentation defect.
- `minor`: wording, formatting, consistency, or low-risk style issues.

Output requirements:
- Report findings first, ordered by severity: `critical`, `major`, `minor`.
- Only report real issues. Do not pad the review.
- For each finding, provide:
  - severity;
  - file path;
  - exact issue;
  - why it matters;
  - specific proposed fix.
- If there are no findings, say: `No issues found.`
- After findings, add a short section named `Residual risk` with any unverified assumptions or areas you could not validate from the available sources.
