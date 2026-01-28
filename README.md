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
