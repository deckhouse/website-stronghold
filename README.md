# Deckhouse Stronghold documentation

This is the source for the Deckhouse Stronghold documentation website.

The project uses [Hugo](gohugo.io) SSG and the [hugo-web-product-module](https://github.com/deckhouse/hugo-web-product-module/) module for a theme (see [README.md](https://github.com/deckhouse/hugo-web-product-module/blob/main/README.md) for details about content markup).

Read [`hugo-web-product-module` README.md](https://github.com/deckhouse/hugo-web-product-module/blob/main/README.md) for information about content markup and other details.
  
## How to run the documentation site locally

To run locally:
1. Install werf and docker.
1. Run:

   ```bash
   make up
   ```

1. Open `http://localhost/products/stronghold/documentation/` in your browser (for the english version) or `http://ru.localhost/products/stronghold/documentation/` (for the russian version).

## Generating PDF/DOCX exports

Run `make pdf` — the werf `print-artifacts` image is built and the resulting files are
extracted into `public/{en,ru}/documentation/downloads/print/stronghold.{pdf,docx}`.
On deploy the same image is imported into `web`, so the site serves them under the same URL.

This project enables PDF/DOCX through:

- `params.pdf: true` in `config/_default/hugo.yaml`;
- `outputs: [HTML, search, print]` in the front matter of `content/documentation/_index.{md,ru.md}`;
- `{{< downloads >}}` shortcode on the documentation landing page (in-content buttons); the
  module theme also renders sidebar download links automatically.

For the full description of the pipeline (werf stages, requirements, how to enable/disable
for a new product website) see the [PDF/DOCX exports section in the module README](https://github.com/deckhouse/hugo-web-product-module/blob/main/README.md#pdfdocx-exports).
