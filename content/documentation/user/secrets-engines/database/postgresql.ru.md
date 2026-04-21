---
title: "PostgreSQL"
linkTitle: "PostgreSQL"
description: "Работа с PostgreSQL в механизме секретов баз данных Deckhouse Stronghold"
weight: 20
---

`PostgreSQL` — один из поддерживаемых плагинов механизма секретов баз данных в Deckhouse Stronghold. Этот плагин позволяет динамически выдавать учётные данные для PostgreSQL на основе ролей Stronghold. Также он поддерживает статические роли.

## Когда использовать

Используйте плагин `postgresql-database-plugin`, если нужно:

- выдавать приложениям временные учётные данные для PostgreSQL;
- ограничивать срок действия доступа через TTL;
- отказаться от постоянных логинов и паролей в конфигурации приложений;
- централизованно управлять доступом к PostgreSQL через роли Stronghold;
- использовать статические роли для заранее созданных учётных записей.

## Что поддерживается

Плагин `postgresql-database-plugin` поддерживает:

- изменение root-учётной записи;
- динамические роли;
- статические роли;
- кастомизацию имени пользователя.

## Как это работает

Базовый сценарий выглядит так:

1. Администратор включает механизм секретов `database`.
2. Администратор настраивает подключение Stronghold к PostgreSQL через `postgresql-database-plugin`.
3. Администратор создаёт роль Stronghold и задаёт SQL-запросы для создания учётных записей.
4. Пользователь или приложение запрашивает учётные данные по роли.
5. Stronghold создаёт нового пользователя в PostgreSQL и возвращает логин, пароль и параметры аренды.

## Настройка

### Шаг 1. Включите механизм секретов базы данных

Если механизм секретов `database` ещё не включён, включите его:

```shell-session
$ d8 stronghold secrets enable database
Success! Enabled the database secrets engine at: database/
```

По умолчанию механизм секретов монтируется по имени `database/`. Если нужен другой путь, используйте аргумент `-path`.

### Шаг 2. Настройте подключение к PostgreSQL

Настройте Stronghold с помощью плагина `postgresql-database-plugin` и параметров подключения:

```shell-session
$ d8 stronghold write database/config/my-postgresql-database \
  plugin_name="postgresql-database-plugin" \
  allowed_roles="my-role" \
  connection_url="postgresql://{{username}}:{{password}}@localhost:5432/database-name" \
  username="strongholduser" \
  password="strongholdpass" \
  password_authentication="scram-sha-256"
```

Здесь:

- `plugin_name` — имя плагина PostgreSQL;
- `allowed_roles` — список ролей Stronghold, которые можно использовать с этим подключением;
- `connection_url` — строка подключения к PostgreSQL;
- `username` и `password` — учётные данные пользователя, от имени которого Stronghold будет управлять доступом;
- `password_authentication` — способ аутентификации пароля.

### Шаг 3. Создайте роль Stronghold

Роль Stronghold определяет, как именно создавать пользователя в PostgreSQL и на какой срок выдавать доступ.

Пример:

```shell-session
$ d8 stronghold write database/roles/my-role \
  db_name="my-postgresql-database" \
  creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
      GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
  default_ttl="1h" \
  max_ttl="24h"
Success! Data written to: database/roles/my-role
```

Здесь:

- `db_name` — имя подключения, которое вы создали на предыдущем шаге;
- `creation_statements` — SQL-запросы, которые Stronghold выполнит для создания пользователя;
- `default_ttl` — срок действия учётных данных по умолчанию;
- `max_ttl` — максимальный срок действия учётных данных.

В этом примере Stronghold:

- создаёт роль PostgreSQL с логином и паролем;
- задаёт срок действия учётной записи через `VALID UNTIL`;
- выдаёт права `SELECT` на все таблицы схемы `public`.

## Получение учётных данных

После настройки механизма секретов и получения токена Stronghold с нужными правами можно запросить временные учётные данные.

Пример:

```shell-session
$ d8 stronghold read database/creds/my-role
Key                Value
---                -----
lease_id           database/creds/my-role/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
lease_duration     1h
lease_renewable    true
password           SsnoaA-8Tv4t34f41baD
username           v-strongholduse-my-role-x
```

Stronghold вернёт:

- `username` — имя созданного пользователя PostgreSQL;
- `password` — пароль;
- `lease_duration` — срок действия учётных данных;
- `lease_renewable` — признак того, что аренду можно продлить.

Эти данные можно сразу использовать в приложении для подключения к PostgreSQL.

## Практические рекомендации

Чтобы клиенту было проще использовать PostgreSQL через Stronghold:

- создавайте отдельные роли Stronghold для разных приложений;
- задавайте минимально необходимые права в `creation_statements`;
- используйте короткий TTL для временных учётных данных;
- проверяйте SQL-запросы в тестовой базе данных перед использованием в production;
- используйте отдельную учётную запись PostgreSQL для Stronghold, а не персональную учётную запись администратора.

## Что дальше

- Если вам нужен обзор всего раздела, откройте [Обзор](./overview/).
- Если вы работаете с MySQL, откройте [MySQL](./mysql/).
- Если вы работаете с ClickHouse, откройте [ClickHouse](./clickhouse/).
