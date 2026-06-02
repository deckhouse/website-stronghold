---
title: "ClickHouse"
hidden: true
weight: 40
---

ClickHouse — один из поддерживаемых плагинов механизма секретов баз данных.
Плагин динамически генерирует учётные данные для базы данных на основе настроенных ролей.
Также он поддерживает статические роли.

## Возможности

Поддерживаются следующие возможности:

| Имя плагина | Изменение root-учётной записи | Динамические роли | Статические роли | Кастомизация имени пользователя |
| --- | --- | --- | --- | --- |
| `clickhouse-database-plugin` | Да | Да | Да | Да |

## Настройка подключения

Чтобы настроить плагин ClickHouse, выполните следующие шаги:

1. Включите механизм секретов баз данных, если он ещё не включён.

   ```shell
   d8 stronghold secrets enable database
   ```

1. Настройте подключение к ClickHouse.

   ```shell
   d8 stronghold write database/config/my-clickhouse-database \
     plugin_name="clickhouse-database-plugin" \
     allowed_roles="my-role" \
     connection_url="clickhouse://clickhouse-server.my:9000??username={{username}}&password={{password}}&secure=true&skip_verify=true" \
     username="strongholduser" \
     password="strongholdpass"
   ```

1. Создайте роль Stronghold.

   В примере предполагается, что в кластере баз данных `my_cluster` уже создана роль `readonly`.

   ```shell
   d8 stronghold write database/roles/my-role \
     db_name="my-clickhouse-database" \
     creation_statements="CREATE USER '{{name}}' IDENTIFIED BY '{{password}}' ON CLUSTER 'my_cluster'; \
       GRANT readonly TO '{{name}}' ON CLUSTER 'my_cluster'; \
       SET DEFAULT ROLE readonly TO '{{name}}';" \
     default_ttl="1h" \
     max_ttl="24h"
   ```

## Получение учётных данных

Чтобы сгенерировать новую учётную запись, выполните команду:

```shell
d8 stronghold read database/creds/my-role
```

Пример вывода:

```console
Key                Value
---                -----
lease_id           database/creds/my-role/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
```
