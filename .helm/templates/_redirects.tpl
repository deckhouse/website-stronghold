{{- define "redirects" -}}
rewrite ^/products/stronghold/documentation/user/concepts/policy\.html$ /products/stronghold/documentation/user/concepts/policy/ redirect;
rewrite ^/products/stronghold/documentation/user/secrets-engines/kv/overview\.html$ /products/stronghold/documentation/user/secrets-engines/kv/overview/ redirect;
rewrite ^/products/stronghold/documentation/user/secrets-engines/databases/overview\.html$ /products/stronghold/documentation/user/secrets-engines/databases/overview/ redirect;
rewrite ^/products/stronghold/documentation/user/auto-snapshot\.html$ /products/stronghold/documentation/user/auto-snapshot/ redirect;
rewrite ^/products/stronghold/documentation/admin/standalone/hsm\.html$ /products/stronghold/documentation/admin/standalone/hsm/ redirect;
rewrite ^/products/stronghold/documentation/about/editions\.html$ /products/stronghold/documentation/about/editions/ redirect;
rewrite ^/products/stronghold/documentation/admin/platform-management/node-management/advanced/examples\.html$ /products/stronghold/documentation/admin/platform-management/node-management/advanced/ redirect;
rewrite ^/products/stronghold/documentation/user/auth/kubernetes\.html$ /products/stronghold/documentation/user/auth/kubernetes/ redirect;
rewrite ^/products/stronghold/documentation/user/concepts/response-wrapping\.html$ /products/stronghold/documentation/user/concepts/response-wrapping/ redirect;
rewrite ^/products/stronghold/documentation/admin/overview\.html$ /products/stronghold/documentation/admin/overview/ redirect;
rewrite ^/products/stronghold/documentation/user/secrets-engines/admin_guide/configuration\.html$ /products/stronghold/documentation/user/secrets-engines/admin_guide/configuration/ redirect;
{{- end -}}
