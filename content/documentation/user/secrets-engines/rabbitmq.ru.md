---
title: "Механизм секретов RabbitMQ"
linkTitle: "RabbitMQ"
description: "Динамическая выдача учётных данных RabbitMQ: подключение к management HTTP API, роли с правами на виртуальные хосты и теги, получение учётных данных и настройка аренды."
weight: 85
---

Механизм секретов RabbitMQ генерирует учётные данные пользователей RabbitMQ динамически на основе настроенных ролей. Для каждого запроса Stronghold создаёт в RabbitMQ отдельного пользователя с заданными правами на виртуальные хосты (vhosts) и тегами, а по истечении [аренды](../../../concepts/lease/) удаляет его.

Stronghold управляет пользователями через RabbitMQ management HTTP API, поэтому в RabbitMQ должен быть включён плагин `rabbitmq_management`.

Механизм секретов RabbitMQ доступен во всех редакциях Stronghold.

## Возможности

| Возможность | Поддержка |
|-------------|-----------|
| Динамические учётные данные | Да |
| Права на виртуальные хосты (`vhosts`) | Да |
| Права на обмены тем (`vhost_topics`) | Да |
| Теги пользователя (`tags`) | Да |
| Кастомизация имени пользователя (`username_template`) | Да |
| Политика паролей (`password_policy`) | Да |

## Настройка

Шаги настройки обычно выполняет администратор Stronghold или средство управления конфигурацией.

1. Включите механизм секретов RabbitMQ:

   ```bash
   d8 stronghold secrets enable rabbitmq
   ```

   По умолчанию механизм подключается по пути `rabbitmq/`. Чтобы подключить его по другому пути, используйте аргумент `-path`.

1. Настройте подключение к RabbitMQ management HTTP API. Укажите пользователя с правами администратора RabbitMQ (тег `administrator`): от его имени Stronghold будет создавать и удалять пользователей.

   ```bash
   d8 stronghold write rabbitmq/config/connection \
     connection_uri="https://rabbitmq.example.com:15672" \
     username="stronghold-admin" \
     password="<пароль>"
   ```

   Параметры эндпоинта `config/connection`:

   | Параметр | Описание |
   |----------|----------|
   | `connection_uri` | URI RabbitMQ management HTTP API. |
   | `username` | Имя пользователя-администратора RabbitMQ. |
   | `password` | Пароль пользователя-администратора RabbitMQ. |
   | `verify_connection` | Проверять ли `connection_uri` реальным подключением к management API. По умолчанию `true`. |
   | `username_template` | Шаблон, по которому генерируются имена динамических пользователей. |
   | `password_policy` | Имя [политики паролей](../../../concepts/password-policy/) для генерации паролей динамических пользователей. |

   {{< alert level="warning" >}}
   Создайте в RabbitMQ отдельного пользователя для Stronghold и не используйте его для других целей.
   {{< /alert >}}

1. Настройте параметры аренды выдаваемых учётных данных:

   ```bash
   d8 stronghold write rabbitmq/config/lease \
     ttl=1800 \
     max_ttl=3600
   ```

   - `ttl` — срок, в течение которого учётные данные действуют без продления, в секундах;
   - `max_ttl` — максимальный срок, после которого учётные данные не продлеваются, в секундах.

   Значение `0` (по умолчанию) означает использование системных значений TTL.

   Текущие параметры аренды можно прочитать командой:

   ```bash
   d8 stronghold read rabbitmq/config/lease
   ```

1. Создайте роль, которая описывает права создаваемых пользователей:

   ```bash
   d8 stronghold write rabbitmq/roles/my-role \
     vhosts='{"/":{"configure":".*","write":".*","read":".*"}}' \
     tags="management"
   ```

   Параметры роли:

   | Параметр | Описание |
   |----------|----------|
   | `vhosts` | JSON-карта виртуальных хостов и прав `configure`, `write` и `read` (регулярные выражения RabbitMQ) для каждого из них. |
   | `vhost_topics` | Вложенная JSON-карта: виртуальный хост → обмен (exchange) → права `write` и `read` на темы. |
   | `tags` | Список тегов пользователя RabbitMQ через запятую, например `management` или `monitoring`. |

   Пример роли с правами на темы в обмене `amq.topic`:

   ```bash
   d8 stronghold write rabbitmq/roles/topic-role \
     vhosts='{"/":{"configure":"","write":".*","read":".*"}}' \
     vhost_topics='{"/":{"amq.topic":{"write":"^orders\\..*","read":".*"}}}'
   ```

## Использование

После настройки механизма секретов пользователь или приложение с токеном Stronghold и соответствующей политикой может получать учётные данные.

1. Запросите учётные данные для роли:

   ```bash
   d8 stronghold read rabbitmq/creds/my-role
   ```

   Пример вывода:

   ```text
   Key                Value
   ---                -----
   lease_id           rabbitmq/creds/my-role/I39Hu8XXOombof4wiK5bKMn9
   lease_duration     30m
   lease_renewable    true
   password           3yNDBikgQvrkx2VA2zhq5IdSM7IWk1RyMYJr
   username           root-39669250-3894-8032-c420-3d58483ebfc4
   ```

   Stronghold создаёт в RabbitMQ пользователя с правами из роли и возвращает его имя и пароль вместе с идентификатором аренды.

1. При необходимости продлите аренду до истечения `lease_duration`:

   ```bash
   d8 stronghold lease renew rabbitmq/creds/my-role/I39Hu8XXOombof4wiK5bKMn9
   ```

1. Когда учётные данные больше не нужны, отзовите аренду. Stronghold удалит пользователя в RabbitMQ:

   ```bash
   d8 stronghold lease revoke rabbitmq/creds/my-role/I39Hu8XXOombof4wiK5bKMn9
   ```

## Управление ролями

- Список ролей:

  ```bash
  d8 stronghold list rabbitmq/roles
  ```

- Чтение роли:

  ```bash
  d8 stronghold read rabbitmq/roles/my-role
  ```

- Удаление роли:

  ```bash
  d8 stronghold delete rabbitmq/roles/my-role
  ```

## Пример политики

Политика, которая разрешает только получение учётных данных для роли `my-role`:

```hcl
path "rabbitmq/creds/my-role" {
  capabilities = ["read"]
}
```

Полный список эндпоинтов и параметров приведён в [справочнике API механизмов секретов](../../../reference/api/secrets/).

## Примеры использования

Готовые примеры с этим механизмом:

- [Динамические пользователи RabbitMQ для сервиса](../../../examples/dynamic-credentials/rabbitmq/)

Все примеры собраны в разделе [«Примеры использования»](../../../examples/).
