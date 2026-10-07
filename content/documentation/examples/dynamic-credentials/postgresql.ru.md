---
title: "Динамические учётные данные PostgreSQL для приложения в DKP"
linkTitle: "Динамические учётные данные PostgreSQL"
description: "Выдача приложению в Deckhouse Kubernetes Platform временных учётных данных PostgreSQL через механизм секретов database и метод аутентификации Kubernetes."
weight: 10
params:
  relatedLinks:
    - title: "Механизм секретов баз данных"
      url: ../../../user/secrets-engines/databases/overview/
    - title: "PostgreSQL"
      url: ../../../user/secrets-engines/databases/postgresql/
    - title: "Метод Kubernetes"
      url: ../../../user/auth/kubernetes/
    - title: "Доставка секретов в поды Kubernetes"
      url: ../../delivery/kubernetes-workloads/
    - title: "Аренда, продление и отзыв"
      url: ../../../concepts/lease/
---

Приложение получает от Stronghold уникальную пару «логин — пароль» к PostgreSQL с ограниченным сроком действия. Stronghold создаёт пользователя в базе данных при запросе и удаляет его по истечении аренды. Пароль не хранится в манифестах и в объектах Secret.

## Цель

Настроить выдачу динамических учётных данных PostgreSQL приложению `myapp`, которое работает в пространстве имён `myapp` кластера Deckhouse Platform, и доставить их в под с автоматическим продлением.

## Предварительные требования

- Stronghold, развёрнутый в DP, и токен с правами на настройку механизмов секретов, политик и методов аутентификации.
- Сервер PostgreSQL, доступный из подов Stronghold, и учётная запись с правом создавать роли (`CREATEROLE`), например `stronghold`.
- Доступ к кластеру через `d8 k`.

В DP метод аутентификации Kubernetes для текущего кластера создаётся автоматически по пути `kubernetes_local`. Если вы используете другой путь, замените `kubernetes_local` в командах ниже.

## Шаг 1. Настройте механизм секретов database

1. Включите механизм секретов:

   ```bash
   d8 stronghold secrets enable database
   ```

1. Настройте подключение к PostgreSQL:

   ```bash
   d8 stronghold write database/config/myapp-postgres \
     plugin_name="postgresql-database-plugin" \
     allowed_roles="myapp" \
     connection_url="postgresql://{{username}}:{{password}}@postgres.db.svc:5432/myapp?sslmode=require" \
     username="stronghold" \
     password="<initial_password>" \
     password_authentication="scram-sha-256"
   ```

1. Смените пароль служебной учётной записи, чтобы он был известен только Stronghold:

   ```bash
   d8 stronghold write -force database/rotate-root/myapp-postgres
   ```

## Шаг 2. Создайте роль базы данных

Роль задаёт SQL-запросы для создания пользователя и срок жизни учётных данных:

```bash
d8 stronghold write database/roles/myapp \
  db_name="myapp-postgres" \
  creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
  revocation_statements="REVOKE ALL ON ALL TABLES IN SCHEMA public FROM \"{{name}}\"; DROP ROLE IF EXISTS \"{{name}}\";" \
  default_ttl="1h" \
  max_ttl="24h"
```

- `default_ttl` — срок аренды при выдаче и при каждом продлении;
- `max_ttl` — максимальный срок жизни учётных данных; после него приложению нужны новые учётные данные.

## Шаг 3. Создайте политику

```bash
d8 stronghold policy write myapp-db - <<'POLICY'
path "database/creds/myapp" {
  capabilities = ["read"]
}
POLICY
```

Продление и отзыв собственных аренд и токена разрешены встроенной политикой `default`.

## Шаг 4. Создайте роль метода Kubernetes

1. Создайте пространство имён и учётную запись сервиса:

   ```bash
   d8 k create namespace myapp
   d8 k -n myapp create serviceaccount myapp
   ```

1. Свяжите учётную запись сервиса с политикой:

   ```bash
   d8 stronghold write auth/kubernetes_local/role/myapp \
     bound_service_account_names=myapp \
     bound_service_account_namespaces=myapp \
     policies=myapp-db \
     ttl=1h
   ```

## Шаг 5. Доставьте учётные данные в под

Для динамических учётных данных используйте [Stronghold Agent](../../../user/agent/overview/) в режиме sidecar: он продлевает аренду и перерисовывает файл, когда учётные данные меняются. Способы, которые подставляют значения только при старте пода (env-injector, init-контейнер), подходят, только если под перезапускается чаще, чем истекает `max_ttl`.

Конфигурация Agent для этого сценария:

```hcl
stronghold {
  address = "https://stronghold.example.com"
}

auto_auth {
  method "kubernetes" {
    mount_path = "auth/kubernetes_local"
    config = {
      role = "myapp"
    }
  }
}

template {
  destination = "/secrets/db.env"
  contents    = <<-EOT
  {{ with secret "database/creds/myapp" }}
  DB_USER={{ .Data.username }}
  DB_PASSWORD={{ .Data.password }}
  {{ end }}
  EOT
}
```

Манифест пода с Agent в режиме sidecar, томом `emptyDir` в памяти и ConfigMap с конфигурацией приведён в разделе [«Доставка секретов в поды Kubernetes»](../../delivery/kubernetes-workloads/#stronghold-agent).

## Шаг 6. Настройте реакцию приложения на смену учётных данных

Agent продлевает аренду каждые `default_ttl`, пока не будет достигнут `max_ttl`. После этого Agent запрашивает новые учётные данные и перезаписывает файл `/secrets/db.env`. Старый пользователь удаляется при отзыве аренды.

Приложение должно:

- перечитывать файл при изменении и переподключаться к базе данных;
- либо завершаться при ошибке аутентификации, чтобы Kubernetes перезапустил контейнер с новыми учётными данными.

Выберите `max_ttl` так, чтобы смена учётных данных происходила не чаще, чем приложение способно корректно переподключиться.

## Проверка

1. Запросите учётные данные вручную с токеном администратора:

   ```bash
   d8 stronghold read database/creds/myapp
   ```

   В ответе должны быть поля `lease_id`, `lease_duration`, `username` и `password`.

1. Проверьте, что пользователь создан в PostgreSQL:

   ```sql
   SELECT rolname, rolvaliduntil FROM pg_roles WHERE rolname LIKE 'v-%';
   ```

1. Убедитесь, что в поде появился файл с учётными данными:

   ```bash
   d8 k -n myapp exec deploy/myapp -c app -- cat /secrets/db.env
   ```

1. Просмотрите активные аренды роли:

   ```bash
   d8 stronghold list sys/leases/lookup/database/creds/myapp/
   ```

## Очистка

1. Удалите тестовые ресурсы в кластере: `d8 k delete namespace myapp`.
1. Отзовите все выданные учётные данные роли:

   ```bash
   d8 stronghold lease revoke -prefix database/creds/myapp/
   ```

1. Удалите роль, политику и конфигурацию:

   ```bash
   d8 stronghold delete auth/kubernetes_local/role/myapp
   d8 stronghold policy delete myapp-db
   d8 stronghold delete database/roles/myapp
   d8 stronghold delete database/config/myapp-postgres
   ```
