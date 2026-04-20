---
title: "Базы данных"
linkTitle: "Обзор"
description: "Обзор механизма секретов баз данных в Deckhouse Stronghold"
weight: 10
---

Механизм секретов баз данных в Deckhouse Stronghold позволяет выдавать учётные данные для поддерживаемых СУБД на основе заранее настроенных ролей. В зависимости от плагина можно использовать динамические роли, статические роли и ротацию root-учётной записи.

Этот раздел помогает понять, как устроен механизм секретов баз данных и когда его использовать.

## Когда использовать

Используйте механизм секретов баз данных, если нужно:

- выдавать приложениям временные учётные данные для подключения к базе данных;
- ограничивать срок действия учётных данных;
- отказаться от постоянных паролей, которые хранятся в конфигурации приложений;
- централизовать управление доступом к PostgreSQL и MySQL;
- автоматически создавать учётные записи по роли в Stronghold.

## Как это работает

Базовый сценарий выглядит так:

1. Администратор включает механизм секретов `database`.
2. Администратор настраивает подключение Stronghold к конкретной СУБД через нужный плагин.
3. Администратор создаёт роль Stronghold и задаёт SQL-запросы, по которым будут создаваться учётные записи.
4. Пользователь или приложение запрашивает учётные данные по роли.
5. Stronghold создаёт учётную запись в базе данных и возвращает логин и пароль с ограниченным TTL.

Такой подход позволяет не хранить постоянные учётные данные в приложениях и выдавать доступ только на нужное время.

## Что входит в раздел

Сейчас в документации есть страницы для таких СУБД:

- [PostgreSQL](./postgresql/) — динамические и статические роли, настройка подключения и выпуск учётных данных;
- [MySQL](./mysql/) — настройка плагинов MySQL, выпуск учётных данных, а также дополнительные сценарии, включая x509 Client-side Certificate Authentication.

## Поддерживаемые возможности

По старым черновикам в документации подтверждены такие возможности:

- **PostgreSQL**
  - изменение root-учётной записи;
  - динамические роли;
  - статические роли;
  - кастомизация имени пользователя;

- **MySQL**
  - изменение root-учётной записи;
  - динамические роли;
  - статические роли;
  - кастомизация имени пользователя.

## Общий сценарий настройки

Ниже приведён типовой порядок настройки механизма секретов баз данных.

### Шаг 1. Включите механизм секретов базы данных

```shell-session
$ d8 stronghold secrets enable database
Success! Enabled the database secrets engine at: database/
```

По умолчанию механизм секретов монтируется по имени `database/`. Если нужен другой путь, используйте аргумент `-path`.

### Шаг 2. Настройте подключение к базе данных

Stronghold нужно настроить для работы с конкретной СУБД через соответствующий плагин.

Для PostgreSQL пример выглядит так:

```shell-session
$ d8 stronghold write database/config/my-postgresql-database \
  plugin_name="postgresql-database-plugin" \
  allowed_roles="my-role" \
  connection_url="postgresql://{{username}}:{{password}}@localhost:5432/database-name" \
  username="strongholduser" \
  password="strongholdpass" \
  password_authentication="scram-sha-256"
```

Для MySQL пример выглядит так:

```shell-session
$ d8 stronghold write database/config/my-mysql-database \
    plugin_name=mysql-database-plugin \
    connection_url="{{username}}:{{password}}@tcp(127.0.0.1:3306)/" \
    allowed_roles="my-role" \
    username="strongholduser" \
    password="strongholdpass"
```

На этом этапе Stronghold получает параметры подключения и список ролей, которые разрешено использовать с этим подключением.

### Шаг 3. Создайте роль Stronghold

Роль связывает имя роли в Stronghold с SQL-запросами, которые будут выполняться в базе данных для создания учётной записи.

Пример для PostgreSQL:

```shell-session
$ d8 stronghold write database/roles/my-role \
  db_name="my-postgresql-database" \
  creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
      GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
  default_ttl="1h" \
  max_ttl="24h"
Success! Data written to: database/roles/my-role
```

Пример для MySQL:

```shell-session
$ d8 stronghold write database/roles/my-role \
    db_name=my-mysql-database \
    creation_statements="CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';GRANT SELECT ON *.* TO '{{name}}'@'%';" \
    default_ttl="1h" \
    max_ttl="24h"
Success! Data written to: database/roles/my-role
```

Здесь:

- `db_name` — имя настроенного подключения к базе данных;
- `creation_statements` — SQL-запросы для создания учётной записи;
- `default_ttl` — срок действия учётных данных по умолчанию;
- `max_ttl` — максимальный срок действия учётных данных.

### Шаг 4. Получите учётные данные

После настройки роли пользователь или приложение могут запросить временные учётные данные.

Пример для PostgreSQL:

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

Пример для MySQL:

```shell-session
$ d8 stronghold read database/creds/my-role
Key                Value
---                -----
lease_id           database/creds/my-role/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
lease_duration     1h
lease_renewable    true
password           yY-57n3X5UQhxnmFRP3f
username           v_strongholduser_my-role_crBWVqVh2Hc1
```

Stronghold вернёт имя пользователя, пароль и параметры аренды. Эти данные можно использовать для подключения приложения к базе данных.

## Практические рекомендации

Чтобы клиенту было проще использовать механизм секретов баз данных:

- создавайте отдельные роли Stronghold для разных приложений и сценариев доступа;
- задавайте короткий TTL для временных учётных данных;
- проверяйте SQL-запросы в `creation_statements` на тестовой базе данных перед использованием в production;
- ограничивайте права создаваемых пользователей только нужными действиями;
- используйте отдельное подключение Stronghold к каждой базе данных или логической группе баз данных.

## Что дальше

- Если вы работаете с PostgreSQL, откройте [PostgreSQL](./postgresql/).
- Если вы работаете с MySQL, откройте [MySQL](./mysql/).
