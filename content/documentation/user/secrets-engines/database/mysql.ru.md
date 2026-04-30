---
title: "MySQL"
linkTitle: "MySQL"
description: "Работа с MySQL в механизме секретов баз данных Deckhouse Stronghold"
weight: 30
---

`MySQL` — один из поддерживаемых плагинов механизма секретов баз данных в Deckhouse Stronghold. Он позволяет динамически выдавать учётные данные для MySQL на основе ролей Stronghold и поддерживает статические роли.

В Stronghold доступно несколько вариантов MySQL-плагина. Они отличаются в первую очередь ограничениями на длину генерируемых имён пользователей для разных версий MySQL:

- `mysql-database-plugin`
- `mysql-aurora-database-plugin`
- `mysql-rds-database-plugin`
- `mysql-legacy-database-plugin`

## Когда использовать

Используйте MySQL-плагин, если нужно:

- выдавать приложениям временные учётные данные для MySQL;
- ограничивать срок действия доступа через TTL;
- отказаться от постоянных логинов и паролей в конфигурации приложений;
- централизованно управлять доступом к MySQL через роли Stronghold;
- использовать статические роли;
- при необходимости подключаться к MySQL с x509 Client-side Certificate Authentication.

## Что поддерживается

MySQL-плагин поддерживает:

- изменение root-учётной записи;
- динамические роли;
- статические роли;
- кастомизацию имени пользователя.

## Как это работает

Базовый сценарий выглядит так:

1. Администратор включает механизм секретов `database`.
2. Администратор настраивает подключение Stronghold к MySQL через один из MySQL-плагинов.
3. Администратор создаёт роль Stronghold и задаёт SQL-запросы для создания учётных записей.
4. Пользователь или приложение запрашивает учётные данные по роли.
5. Stronghold создаёт нового пользователя в MySQL и возвращает логин, пароль и параметры аренды.

## Настройка

### Шаг 1. Включите механизм секретов базы данных

Если механизм секретов `database` ещё не включён, включите его:

```shell-session
$ d8 stronghold secrets enable database
Success! Enabled the database secrets engine at: database/
```

По умолчанию механизм секретов монтируется по имени `database/`. Если нужен другой путь, используйте аргумент `-path`.

### Шаг 2. Настройте подключение к MySQL

Настройте Stronghold с помощью подходящего MySQL-плагина и параметров подключения:

```shell-session
$ d8 stronghold write database/config/my-mysql-database \
    plugin_name=mysql-database-plugin \
    connection_url="{{username}}:{{password}}@tcp(127.0.0.1:3306)/" \
    allowed_roles="my-role" \
    username="strongholduser" \
    password="strongholdpass"
```

Здесь:

- `plugin_name` — имя MySQL-плагина;
- `connection_url` — строка подключения к MySQL;
- `allowed_roles` — список ролей Stronghold, которые можно использовать с этим подключением;
- `username` и `password` — учётные данные пользователя, от имени которого Stronghold будет управлять доступом.

### Шаг 3. Создайте роль Stronghold

Роль Stronghold определяет, как создавать пользователя в MySQL и на какой срок выдавать доступ.

Пример:

```shell-session
$ d8 stronghold write database/roles/my-role \
    db_name=my-mysql-database \
    creation_statements="CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';GRANT SELECT ON *.* TO '{{name}}'@'%';" \
    default_ttl="1h" \
    max_ttl="24h"
Success! Data written to: database/roles/my-role
```

Здесь:

- `db_name` — имя подключения, которое вы создали на предыдущем шаге;
- `creation_statements` — SQL-запросы для создания учётной записи;
- `default_ttl` — срок действия учётных данных по умолчанию;
- `max_ttl` — максимальный срок действия учётных данных.

В этом примере Stronghold:

- создаёт нового пользователя MySQL;
- задаёт пароль;
- выдаёт права `SELECT`.

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
password           yY-57n3X5UQhxnmFRP3f
username           v_strongholduser_my-role_crBWVqVh2Hc1
```

Stronghold вернёт:

- `username` — имя созданного пользователя MySQL;
- `password` — пароль;
- `lease_duration` — срок действия учётных данных;
- `lease_renewable` — признак того, что аренду можно продлить.

Эти данные можно сразу использовать в приложении для подключения к MySQL.

## Аутентификация клиента по сертификату x509

Этот плагин поддерживает MySQL x509 Client-side Certificate Authentication.

Чтобы использовать этот способ аутентификации, настройте подключение так:

```shell-session
$ d8 stronghold write database/config/my-mysql-database \
    plugin_name=mysql-database-plugin \
    allowed_roles="my-role" \
    connection_url="user:password@tcp(localhost:3306)/test" \
    tls_certificate_key=@/path/to/client.pem \
    tls_ca=@/path/to/client.ca
```

Параметры:

- `tls_certificate_key` соответствует настройке `ssl-cert` вместе с `ssl-key` в MySQL;
- `tls_ca` соответствует настройке `ssl-ca`.

В Stronghold эти параметры передаются как содержимое файлов, а не как пути к файлам. Эти параметры не зависят друг от друга.

## Примеры

### Использование шаблонов в `grant statements`

MySQL поддерживает шаблоны в `grant statements`. Это полезно, если приложению нужен доступ к большому количеству баз данных по шаблону.

Например, если нужно выдать доступ ко всем базам данных, имена которых начинаются с `fooapp_`, используйте такой `creation_statements`:

```text
CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}'; GRANT SELECT ON `fooapp\_%`.* TO '{{name}}'@'%';
```

Если вы передаёте такой запрос через CLI, shell может неправильно интерпретировать содержимое в кавычках. В этом случае старый черновик рекомендует закодировать `creation_statements` в Base64 и передать его в Stronghold.

Пример:

```shell-session
$ d8 stronghold write database/roles/my-role \
    db_name=mysql \
    creation_statements="Q1JFQVRFIFVTRVIgJ3t7bmFtZX19J0AnJScgSURFTlRJRklFRCBCWSAne3twYXNzd29yZH19JzsgR1JBTlQgU0VMRUNUIE9OIGBmb29hcHBcXyVgLiogVE8gJ3t7bmFtZX19J0AnJSc7" \
    default_ttl="1h" \
    max_ttl="24h"
```

### Изменение root-учётных данных в MySQL 5.6

По умолчанию для MySQL используется синтаксис `ALTER USER`, который есть в MySQL 5.7 и новее.

Если вы используете MySQL 5.6, настройте `root_rotation_statements` с использованием старого синтаксиса `SET PASSWORD`.

Пример:

```shell-session
$ d8 stronghold write database/config/my-mysql-database \
    plugin_name=mysql-database-plugin \
    connection_url="{{username}}:{{password}}@tcp(127.0.0.1:3306)/" \
    root_rotation_statements="SET PASSWORD = PASSWORD('{{password}}')" \
    allowed_roles="my-role" \
    username="root" \
    password="mysql"
```

## Практические рекомендации

Чтобы клиенту было проще использовать MySQL через Stronghold:

- выбирайте подходящий MySQL-плагин с учётом ограничений на длину имени пользователя;
- создавайте отдельные роли Stronghold для разных приложений;
- задавайте минимально необходимые права в `creation_statements`;
- используйте короткий TTL для временных учётных данных;
- проверяйте SQL-запросы в тестовой базе данных перед использованием в production;
- если используете сложные `grant statements`, заранее проверьте, как shell обрабатывает кавычки и спецсимволы;
- для MySQL 5.6 отдельно настройте `root_rotation_statements`.

## Что дальше

- Если вам нужен обзор всего раздела, откройте [Обзор](./overview/).
- Если вы работаете с PostgreSQL, откройте [PostgreSQL](./postgresql/).
- Если вы работаете с ClickHouse, откройте [ClickHouse](./clickhouse/).
