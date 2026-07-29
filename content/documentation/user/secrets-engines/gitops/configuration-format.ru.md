---
title: "Формат конфигурации"
weight: 20
description: "Декларативный YAML-формат ресурсов Stronghold для механизма секретов GitOps."
---

Механизм секретов GitOps применяет ресурсы, описанные в YAML. Каждый документ — один вызов create или update API. В одном файле может быть несколько документов, разделённых `---`. Механизм рекурсивно загружает все файлы `.yaml` и `.yml` из репозитория.

Поддерживаются только операции create и update с телом запроса (не list и не get), кроме случаев, когда явно задано `method: GET`.

## Поля ресурса

```yaml
# Обязательно
path: <путь с подставленными path-параметрами>
data: <объект — тело запроса для операции API>

# Опционально
namespace: /           # неймспейс Stronghold; если не указан, заголовок не передаётся
name: ""               # человекочитаемое имя; должно быть уникальным среди ресурсов
revision: 0            # неотрицательное целое; увеличьте, чтобы принудительно переприменить ресурс
dependencies: []       # имена ресурсов, от которых зависит данный
ignore_failures: false # при true ошибка этого ресурса не прерывает применение
method: POST           # GET или POST (по умолчанию POST); для GET тело не отправляется
```

| Поле | Обязательное | По умолчанию | Описание |
|------|--------------|--------------|----------|
| `path` | да | — | Путь API с уже подставленными path-параметрами, без префикса `/v1/` |
| `data` | да | — | Тело запроса для соответствующей операции API |
| `namespace` | нет | не передавать | Неймспейс для запроса (`X-Vault-Namespace`); например `ns1/` или `ns1/team/` |
| `name` | нет | `""` | Имя ресурса — ключ в состоянии; если не задано, ключ — `namespace` + `path` |
| `revision` | нет | `0` | Участвует в дайджесте; увеличьте, чтобы принудительно переприменить ресурс при неизменном `data` |
| `dependencies` | нет | `[]` | Имена ресурсов, которые должны быть применены раньше |
| `ignore_failures` | нет | `false` | Продолжать применение при ошибке этого ресурса |
| `method` | нет | `POST` | `GET` или `POST`; для `GET` тело не отправляется |

Минимум для одного ресурса: `path` и `data`.

### Путь и удаление

Используйте пути, в которых идентификатор ресурса входит в `path`, чтобы при удалении ресурса из YAML корректно вызывался DELETE. Предпочитайте `identity/group/name/my-group` вместо `identity/group`, а также `identity/entity/name/my-entity`, `sys/policies/acl/mypolicy`, `sys/namespaces/ns1`, `auth/ldap/groups/dev-group`.

Некоторые ресурсы API нельзя удалить по тому же пути, по которому они созданы (например `pki/root/generate/internal`). Если DELETE возвращает `405` или `404`, запись удаляется из состояния без ошибки применения.

### Имя и ключ в состоянии

Ключ в состоянии — всегда уникальное имя:

- если задан `name`, он и есть ключ;
- иначе ключ — нормализованный неймспейс (со слешем в конце) плюс `path`, например `ns1/kv-v2/secret`, а в root — только `kv-v2/secret`.

Имена должны быть уникальными. Если изменился только `name`, а `data` нет, запрос в API не отправляется — обновляется только состояние.

### revision

При неизменном `data` ресурс повторно не применяется. Увеличьте `revision` (например с `0` до `1`), чтобы изменить дайджест и принудительно выполнить применение — это нужно после ручного удаления или при перегенерации сертификата или пароля.

### dependencies

Порядок применения и удаления выводится из графа зависимостей (топологическая сортировка). Автоматического упорядочивания нет: чтобы неймспейс создался раньше ресурсов в нём, укажите ресурс неймспейса в `dependencies`.

### Шаблоны

В полях `data` можно подставлять значения из ответов API уже применённых ресурсов. Строка вида `<name:key>` при применении заменяется на значение из состояния:

- `name` — имя ресурса-источника (`name` из конфигурации или ключ по умолчанию namespace+path);
- `key` — путь по JSON сохранённого ответа (поля через точку, элементы массива по индексу), например `client_token`, `keys.0`, `id`.

Шаблон распознаётся только если строка целиком заключена в угловые скобки и содержит ровно две части, разделённые `:`.

Ресурс с шаблоном должен указать ресурс-источник в `dependencies`.

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

При `ignore_failures: true` ошибка этого ресурса фиксируется, но остальные ресурсы продолжают применяться. Используйте для опциональных или зависящих от окружения ресурсов, например подключения к базе данных, когда база недоступна.

## Неймспейсы

Создавайте неймспейсы по пути `sys/namespaces/{path}`:

```yaml
---
path: sys/namespaces/ns1
data: {}
---
path: sys/namespaces/ns1/team
data: {}
```

Для ресурсов внутри неймспейса задайте `namespace` (например `ns1/`). Запросы create, update и delete выполняются с тем же заголовком неймспейса. Укажите ресурс неймспейса в `dependencies`, чтобы он создался первым.

## Примеры

### ACL-политика и механизмы секретов

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

### Метод аутентификации с зависимостями и шаблонами

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

### Механизм секретов баз данных

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
