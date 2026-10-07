---
title: "AppRole для приложения и скриптов"
linkTitle: "AppRole"
description: "Аутентификация приложения без пользователя: метод AppRole, роль с ограничением по сети и числу использований secret_id, получение токена и чтение секрета."
weight: 90
params:
  relatedLinks:
    - title: "Метод аутентификации AppRole"
      url: ../../../user/auth/approle/
    - title: "Политики"
      url: ../../../concepts/policy/
    - title: "Приложение на виртуальной машине со Stronghold Agent"
      url: ../legacy-app-on-vm/
---

AppRole подходит для приложений и скриптов, которым нужен доступ к секретам без участия человека. Приложение предъявляет пару идентификаторов: `role_id` (почти не меняется, не секрет) и `secret_id` (выдаётся на ограниченное время и число использований), и получает токен с политиками роли.

## Цель

Настроить роль для приложения `myapp`, получить для неё `role_id` и `secret_id`, войти по ним и прочитать секрет, ограничив срок жизни токена, число использований `secret_id` и сети, из которых разрешён вход.

## Предварительные требования

- Токен Stronghold с правами на включение методов аутентификации, запись политик и ролей.
- Механизм секретов KV версии 2, смонтированный в `secret/` (в режиме разработки он включён по умолчанию).

## Шаг 1. Создайте секрет и политику

1. Запишите секрет, который будет читать приложение:

   ```bash
   d8 stronghold kv put -mount=secret myapp/config db_password=s3cr3t
   ```

1. Создайте политику, которая разрешает только чтение этого пути:

   ```bash
   d8 stronghold policy write myapp - <<'POLICY'
   path "secret/data/myapp/*" {
     capabilities = ["read"]
   }
   POLICY
   ```

## Шаг 2. Включите метод и создайте роль

```bash
d8 stronghold auth enable approle

d8 stronghold write auth/approle/role/myapp \
  token_policies=myapp \
  token_ttl=15m \
  token_max_ttl=1h \
  secret_id_ttl=24h \
  secret_id_num_uses=1 \
  token_bound_cidrs="127.0.0.1/32,10.0.0.0/8"
```

Параметры роли:

| Параметр | Назначение |
| --- | --- |
| `token_policies` | Политики токена, который получит приложение |
| `token_ttl`, `token_max_ttl` | Срок жизни токена и предел его продления |
| `secret_id_ttl` | Срок, в течение которого можно использовать выданный `secret_id` |
| `secret_id_num_uses` | Сколько раз можно войти по одному `secret_id`. `1` означает одноразовый `secret_id` |
| `token_bound_cidrs` | Сети, из которых можно использовать выданный токен |
| `secret_id_bound_cidrs` | Сети, из которых можно войти по `secret_id` |

Укажите в `token_bound_cidrs` реальные сети серверов приложения. В примере добавлен `127.0.0.1/32`, чтобы его можно было проверить на одной машине.

## Шаг 3. Получите идентификаторы

```bash
ROLE_ID=$(d8 stronghold read -field=role_id auth/approle/role/myapp/role-id)
SECRET_ID=$(d8 stronghold write -f -field=secret_id auth/approle/role/myapp/secret-id)
```

`role_id` можно положить в конфигурацию приложения. `secret_id` передайте по защищённому каналу, лучше в [обёрнутом виде](../../operations/secret-sharing-wrapping/): обёрнутый `secret_id` можно получить только один раз.

## Шаг 4. Войдите и прочитайте секрет

1. Войдите по `role_id` и `secret_id`:

   ```bash
   APP_TOKEN=$(d8 stronghold write -field=token auth/approle/login \
     role_id="$ROLE_ID" secret_id="$SECRET_ID")
   ```

1. Прочитайте секрет токеном приложения:

   ```bash
   STRONGHOLD_TOKEN="$APP_TOKEN" d8 stronghold kv get -mount=secret myapp/config
   ```

## Проверка

1. Убедитесь, что токен получил только политики роли:

   ```bash
   STRONGHOLD_TOKEN="$APP_TOKEN" d8 stronghold token lookup
   ```

   В поле `policies` должны быть `default` и `myapp`.

1. Убедитесь, что `secret_id` одноразовый: повторный вход должен завершиться ошибкой:

   ```bash
   # повторный вход по тому же secret_id
   d8 stronghold write auth/approle/login role_id="$ROLE_ID" secret_id="$SECRET_ID"
   ```

1. Убедитесь, что приложение не читает чужие секреты:

   ```bash
   # чтение пути вне политики
   STRONGHOLD_TOKEN="$APP_TOKEN" d8 stronghold kv get -mount=secret other/config
   ```

   Запрос должен завершиться ошибкой `permission denied`.

## Очистка

```bash
d8 stronghold token revoke "$APP_TOKEN"
d8 stronghold delete auth/approle/role/myapp
d8 stronghold policy delete myapp
d8 stronghold kv metadata delete -mount=secret myapp/config
```
