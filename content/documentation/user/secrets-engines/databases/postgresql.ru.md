---
title: "PostgreSQL"
weight: 20
---

PostgreSQL — один из поддерживаемых плагинов механизма секретов баз данных.
Плагин динамически генерирует учётные данные базы данных на основе настроенных ролей PostgreSQL.
Также поддерживаются статические роли.

## Возможности

Поддерживаются следующие возможности:

| Имя плагина | Изменение root-учётной записи | Динамические роли | Статические роли | Кастомизация имени пользователя |
| --- | --- | --- | --- | --- |
| `postgresql-database-plugin` | Да | Да | Да | Да |

## Настройка подключения

Чтобы настроить плагин PostgreSQL, выполните следующие шаги:

1. Включите механизм секретов баз данных, если он ещё не включён.

   ```shell
   d8 stronghold secrets enable database
   ```

1. Настройте подключение к PostgreSQL.

   ```shell
   d8 stronghold write database/config/my-postgresql-database \
     plugin_name="postgresql-database-plugin" \
     allowed_roles="my-role" \
     connection_url="postgresql://{{username}}:{{password}}@localhost:5432/database-name" \
     username="strongholduser" \
     password="strongholdpass" \
     password_authentication="scram-sha-256"
   ```

1. Создайте роль Stronghold, которая сопоставляет имя роли с SQL-инструкциями для создания учётной записи в PostgreSQL.

   ```shell
   d8 stronghold write database/roles/my-role \
     db_name="my-postgresql-database" \
     creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
       GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
     default_ttl="1h" \
     max_ttl="24h"
   ```

## Получение учётных данных

Чтобы сгенерировать новую учётную запись, выполните команду:

```shell
d8 stronghold read database/creds/my-role
```
