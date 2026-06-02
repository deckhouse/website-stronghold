---
title: "MySQL"
weight: 30
---

MySQL — один из поддерживаемых плагинов механизма секретов баз данных в Stronghold.

Плагин динамически генерирует учётные данные базы данных на основе настроенных ролей MySQL.
Также поддерживаются статические роли.

В Stronghold доступно несколько вариантов этого плагина.
Каждый вариант предназначен для разных драйверов MySQL.
Основное различие между ними заключается в допустимой длине имён пользователей, поскольку разные версии MySQL поддерживают разную длину имени пользователя.

Доступны следующие плагины:

- `mysql-database-plugin`;
- `mysql-aurora-database-plugin`;
- `mysql-rds-database-plugin`;
- `mysql-legacy-database-plugin`.

## Возможности

Поддерживаются следующие возможности:

| Имя плагина | Изменение root-учётной записи | Динамические роли | Статические роли | Настройка имени пользователя |
| --- | --- | --- | --- | --- |
| `mysql-database-plugin` | Да | Да | Да | Да |
| `mysql-aurora-database-plugin` | Да | Да | Да | Да |
| `mysql-rds-database-plugin` | Да | Да | Да | Да |
| `mysql-legacy-database-plugin` | Да | Да | Да | Да |

## Настройка подключения

Чтобы настроить MySQL-плагин, выполните следующие шаги:

1. Включите механизм секретов баз данных, если он ещё не включён.

   ```shell
   d8 stronghold secrets enable database
   ```

1. Настройте Stronghold, указав нужный плагин и параметры подключения.

   ```shell
   d8 stronghold write database/config/my-mysql-database \
     plugin_name=mysql-database-plugin \
     connection_url="{{username}}:{{password}}@tcp(127.0.0.1:3306)/" \
     allowed_roles="my-role" \
     username="strongholduser" \
     password="strongholdpass"
   ```

1. Создайте роль Stronghold, которая сопоставляет имя роли с SQL-запросом для создания учётной записи в MySQL.

   ```shell
   d8 stronghold write database/roles/my-role \
     db_name=my-mysql-database \
     creation_statements="CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';GRANT SELECT ON *.* TO '{{name}}'@'%';" \
     default_ttl="1h" \
     max_ttl="24h"
   ```

## Получение учётных данных

Чтобы создать новую учётную запись, используйте эндпоинт `database/creds/<имя-роли>`:

```shell
d8 stronghold read database/creds/my-role
```
