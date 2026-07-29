---
title: "Configuration format"
weight: 20
description: "Declarative YAML format for Stronghold resources managed by the GitOps secrets engine."
---

The GitOps secrets engine applies resources described in YAML. Each document is one create or update API call. A file may contain several documents separated by `---`. The engine recursively loads all `.yaml` and `.yml` files from the repository.

Only create and update operations with a request body are supported (not list or get), except where `method: GET` is set explicitly.

## Resource fields

```yaml
# Required
path: <path with path parameters filled in>
data: <object — request body for the API operation>

# Optional
namespace: /           # Stronghold namespace; if omitted, the header is not sent
name: ""               # human-readable name; must be unique across resources
revision: 0            # non-negative integer; increase to force re-apply
dependencies: []       # names of resources this resource depends on
ignore_failures: false # if true, an error for this resource does not abort apply
method: POST           # GET or POST (default POST); GET sends no body
```

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `path` | yes | — | API path with path parameters already substituted, without the `/v1/` prefix |
| `data` | yes | — | Request body for the matching API operation |
| `namespace` | no | not sent | Namespace for the request (`X-Vault-Namespace`); for example `ns1/` or `ns1/team/` |
| `name` | no | `""` | Resource name used as the state key; if unset, the key is `namespace` + `path` |
| `revision` | no | `0` | Included in the digest; increase to force re-apply when `data` is unchanged |
| `dependencies` | no | `[]` | Resource names that must be applied before this resource |
| `ignore_failures` | no | `false` | Continue apply when this resource fails |
| `method` | no | `POST` | `GET` or `POST`; `GET` sends no body |

Minimum for one resource: `path` and `data`.

### Path and deletion

Use paths where the resource identifier is part of the path, so DELETE works when a resource is removed from YAML. Prefer `identity/group/name/my-group` over `identity/group`, and similarly `identity/entity/name/my-entity`, `sys/policies/acl/mypolicy`, `sys/namespaces/ns1`, `auth/ldap/groups/dev-group`.

Some API resources cannot be deleted on the same path they were created on (for example `pki/root/generate/internal`). If DELETE returns `405` or `404`, the entry is removed from state without failing the apply.

### Name and state key

The state key is always a unique name:

- if `name` is set, that value is the key;
- otherwise the key is the normalized namespace (with a trailing `/`) plus `path` — for example `ns1/kv-v2/secret`, or just `kv-v2/secret` in the root namespace.

Names must be unique. If only `name` changes and `data` does not, no API request is sent — only state is updated.

### Revision

Unchanged `data` is not re-applied. Increase `revision` (for example from `0` to `1`) to change the digest and force a new apply — useful after manual deletion or when regenerating a certificate or password.

### Dependencies

Apply and delete order come from the dependency graph (topological sort). There is no automatic ordering: to create a namespace before resources inside it, list the namespace resource in `dependencies`.

### Templates

In `data` fields you can substitute values from API responses of already applied resources. A string of the form `<name:key>` is replaced at apply time:

- `name` — source resource name (`name` from config or the default namespace+path key);
- `key` — JSON path into the saved response (dot-separated fields, array indexes), for example `client_token`, `keys.0`, `id`.

The template is recognized only when the whole string is wrapped in angle brackets and contains exactly two parts separated by `:`.

A resource that uses a template must list the source resource in `dependencies`.

```yaml
---
name: token-create
path: auth/token/create
data:
  policies: ["default"]
  ttl: 1h
---
name: save-token-to-kv
path: kv1/mysecret
dependencies:
  - token-create
data:
  token: <token-create:client_token>
```

### ignore_failures

With `ignore_failures: true`, an error for this resource is recorded but other resources continue. Use this for optional or environment-dependent resources (for example a database connection when the database is unavailable).

## Namespaces

Create namespaces with path `sys/namespaces/{path}`:

```yaml
---
path: sys/namespaces/ns1
data: {}
---
path: sys/namespaces/ns1/team
data: {}
```

For resources inside a namespace, set `namespace` (for example `ns1/`). Create, update, and delete requests use the same namespace header. List the namespace resource in `dependencies` so it is created first.

## Examples

### ACL policy and secrets engines

```yaml
---
path: sys/policies/acl/mypolicy
data:
  policy: |
    path "*" { capabilities = ["read","list"] }
---
path: sys/mounts/kv-v2
data:
  type: kv
  description: KV secrets engine version 2
  options:
    version: "2"
---
path: sys/mounts/transit
data:
  type: transit
  description: Transit encryption
---
path: transit/keys/my-aes-key
dependencies:
  - sys/mounts/transit
data:
  type: aes256-gcm96
```

### Auth method with dependencies and templates

```yaml
---
path: sys/auth/approle
data:
  type: approle
  description: Approle auth method
---
path: auth/approle/role/app
dependencies:
  - sys/auth/approle
data:
  policies:
    - mypolicy
  secret_id_ttl: 1h
  secret_id_num_uses: 10
---
path: auth/approle/role/app/secret-id
dependencies:
  - auth/approle/role/app
data: {}
---
path: sys/mounts/my-secrets
data:
  type: kv
  description: KV secrets engine version 2
  options:
    version: "2"
---
path: my-secrets/data/my-data
dependencies:
  - sys/mounts/my-secrets
  - auth/approle/role/app/secret-id
data:
  data:
    secret_id: <auth/approle/role/app/secret-id:secret_id>
```

### Database secrets engine

```yaml
---
path: sys/mounts/database
data:
  type: database
  description: Database secrets engine
---
path: database/config/postgres
dependencies:
  - sys/mounts/database
data:
  plugin_name: postgresql-database-plugin
  connection_url: postgresql://{{username}}:{{password}}@postgres.example.com:5432/postgres?sslmode=disable
  username: vault
  password: secret
  allowed_roles:
    - app
  verify_connection: false
---
path: database/roles/app
dependencies:
  - database/config/postgres
data:
  db_name: postgres
  creation_statements:
    - CREATE ROLE "{{name}}" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';
    - GRANT SELECT ON ALL TABLES IN SCHEMA public TO "{{name}}";
  revocation_statements:
    - REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM "{{name}}";
    - DROP ROLE IF EXISTS "{{name}}";
  default_ttl: 1h
  max_ttl: 1h
```

### Identity

```yaml
---
path: identity/entity/name/my-entity
data:
  policies:
    - mypolicy
---
path: identity/group/name/my-group
data:
  type: internal
  policies:
    - mypolicy
```
