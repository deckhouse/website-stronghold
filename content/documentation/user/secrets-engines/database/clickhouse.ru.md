---
title: "ClickHouse"
linkTitle: "ClickHouse"
description: "Работа с ClickHouse в механизме секретов баз данных Deckhouse Stronghold"
weight: 40
---

`ClickHouse` — один из поддерживаемых плагинов механизма секретов баз данных в Deckhouse Stronghold. Этот плагин позволяет динамически выдавать учётные данные для ClickHouse на основе ролей Stronghold и поддерживает статические роли.

## Когда использовать

Используйте плагин `clickhouse-database-plugin`, если нужно:

- выдавать приложениям временные учётные данные для ClickHouse;
- ограничивать срок действия доступа через TTL;
- отказаться от постоянных логинов и паролей в конфигурации приложений;
- централизованно управлять доступом к ClickHouse через роли Stronghold;
- использовать статические роли для заранее созданных учётных записей.

## Что поддерживается

Плагин `clickhouse-database-plugin` поддерживает:

- изменение root-учётной записи;
- динамические роли;
- статические роли;
- кастомизацию имени пользователя.

## Как это работает

Базовый сценарий выглядит так:

1. Администратор включает механизм секретов `database`.
2. Администратор настраивает подключение Stronghold к ClickHouse через `clickhouse-database-plugin`.
3. Администратор создаёт роль Stronghold и задаёт SQL-запросы для создания учётных записей.
4. Пользователь или приложение запрашивает учётные данные по роли.
5. Stronghold создаёт нового пользователя в ClickHouse и возвращает логин, пароль и параметры аренды.

## Настройка

### Шаг 1. Включите механизм секретов базы данных

Если механизм секретов `database` ещё не включён, включите его:

```shell-session
$ d8 stronghold secrets enable database
Success! Enabled the database secrets engine at: database/
```

По умолчанию механизм секретов монтируется по имени `database/`. Если нужен другой путь, используйте аргумент `-path`.

### Шаг 2. Настройте подключение к ClickHouse

Настройте Stronghold с помощью плагина `clickhouse-database-plugin` и параметров подключения:

```shell-session
$ d8 stronghold write database/config/my-clickhouse-database \
    plugin_name="clickhouse-database-plugin" \
    allowed_roles="my-role" \
    connection_url="clickhouse://clickhouse-server.my:9000??username={{username}}&password={{password}}&secure=true&skip_verify=true" \
    username="strongholduser" \
    password="strongholdpass"
```

Здесь:

- `plugin_name` — имя плагина ClickHouse;
- `allowed_roles` — список ролей Stronghold, которые можно использовать с этим подключением;
- `connection_url` — строка подключения к ClickHouse;
- `username` и `password` — учётные данные пользователя, от имени которого Stronghold будет управлять доступом.

### Шаг 3. Создайте роль Stronghold

Роль Stronghold определяет, как именно создавать пользователя в ClickHouse и на какой срок выдавать доступ.

Пример:

```shell-session
$ d8 stronghold write database/roles/my-role \
    db_name="my-clickhouse-database" \
    creation_statements="CREATE USER '{{name}}' IDENTIFIED BY '{{password}}' ON CLUSTER 'my_cluster'; \
        GRANT readonly TO '{{name}}' ON CLUSTER 'my_cluster'; \
        SET DEFAULT ROLE readonly TO '{{name}}';" \
    default_ttl="1h" \
    max_ttl="24h"
Success! Data written to: database/roles/my-role
```

Здесь:

- `db_name` — имя подключения, которое вы создали на предыдущем шаге;
- `creation_statements` — SQL-запросы, которые Stronghold выполнит для создания учётной записи;
- `default_ttl` — срок действия учётных данных по умолчанию;
- `max_ttl` — максимальный срок действия учётных данных.

В этом примере предполагается, что в кластере баз данных `my_cluster` уже создана роль `readonly`.

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

- `username` — имя созданного пользователя ClickHouse;
- `password` — пароль;
- `lease_duration` — срок действия учётных данных;
- `lease_renewable` — признак того, что аренду можно продлить.

Эти данные можно сразу использовать в приложении для подключения к ClickHouse.

## Практические рекомендации

Чтобы клиенту было проще использовать ClickHouse через Stronghold:

- создавайте отдельные роли Stronghold для разных приложений;
- задавайте минимально необходимые права в `creation_statements`;
- используйте короткий TTL для временных учётных данных;
- проверяйте SQL-запросы в тестовом окружении перед использованием в production;
- заранее убедитесь, что роль `readonly` или другая требуемая роль уже создана в кластере ClickHouse, если вы ссылаетесь на неё в `creation_statements`.

## Что дальше

- Если вам нужен обзор всего раздела, откройте [Обзор](./overview/).
- Если вы работаете с PostgreSQL, откройте [PostgreSQL](./postgresql/).
- Если вы работаете с MySQL, откройте [MySQL](./mysql/).
