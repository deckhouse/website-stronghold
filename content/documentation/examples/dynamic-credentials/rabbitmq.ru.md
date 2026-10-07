---
title: "Динамические пользователи RabbitMQ для сервиса"
linkTitle: "Динамические пользователи RabbitMQ"
description: "Выдача сервису временных пользователей RabbitMQ с правами только на свой виртуальный хост и очереди через механизм секретов rabbitmq: подключение, аренда, роль, политика и проверка через rabbitmqctl."
weight: 40
params:
  relatedLinks:
    - title: "Механизм секретов RabbitMQ"
      url: ../../../user/secrets-engines/rabbitmq/
    - title: "Аренда, продление и отзыв"
      url: ../../../concepts/lease/
    - title: "Доставка секретов в поды Kubernetes"
      url: ../../delivery/kubernetes-workloads/
    - title: "API механизмов секретов"
      url: ../../../reference/api/secrets/
---

Сервис получает от Stronghold отдельного пользователя RabbitMQ с правами только на свой виртуальный хост и на ресурсы с заданным префиксом имени. Stronghold создаёт пользователя через RabbitMQ management HTTP API и удаляет его при отзыве или по истечении аренды.

## Цель

Настроить выдачу динамических учётных данных RabbitMQ сервису `orders`: пользователь получает на виртуальном хосте `orders` права `configure`, `write` и `read` только на очереди и обмены с именами, начинающимися с `orders.`, живёт 1 час с возможностью продления до 24 часов.

## Предварительные требования

- Stronghold и токен с правами на настройку механизмов секретов и политик.
- RabbitMQ с включённым плагином `rabbitmq_management`, management HTTP API доступен из Stronghold.
- Созданный виртуальный хост `orders`, например командой `rabbitmqctl add_vhost orders`.
- Доступ к `rabbitmqctl` на узле RabbitMQ для проверки.

## Шаг 1. Подготовьте учётную запись для Stronghold

Создайте в RabbitMQ отдельного пользователя с тегом `administrator` — от его имени Stronghold будет создавать и удалять пользователей:

```bash
rabbitmqctl add_user stronghold-admin '<пароль>'
rabbitmqctl set_user_tags stronghold-admin administrator
```

Не используйте эту учётную запись для других целей.

## Шаг 2. Настройте механизм секретов

1. Включите механизм секретов RabbitMQ:

   ```bash
   d8 stronghold secrets enable rabbitmq
   ```

1. Настройте подключение к management HTTP API:

   ```bash
   d8 stronghold write rabbitmq/config/connection \
     connection_uri="https://rabbitmq.example.com:15672" \
     username="stronghold-admin" \
     password="<пароль>"
   ```

1. Задайте параметры аренды в секундах. Они действуют для всех ролей этого механизма секретов:

   ```bash
   d8 stronghold write rabbitmq/config/lease \
     ttl=3600 \
     max_ttl=86400
   ```

   Чтобы задать для разных сервисов разные сроки аренды, подключите отдельные экземпляры механизма секретов с аргументом `-path`.

## Шаг 3. Создайте роль

```bash
d8 stronghold write rabbitmq/roles/orders \
  vhosts='{"orders":{"configure":"^orders\\..*","write":"^orders\\..*","read":"^orders\\..*"}}'
```

- `vhosts` — JSON-карта виртуальных хостов и прав. Значения `configure`, `write` и `read` — регулярные выражения RabbitMQ для имён очередей и обменов;
- `tags` не задан — пользователь не получает доступ к веб-интерфейсу management и HTTP API.

Если сервис публикует сообщения в обмен типа `topic`, ограничьте допустимые ключи маршрутизации параметром `vhost_topics`:

```bash
d8 stronghold write rabbitmq/roles/orders \
  vhosts='{"orders":{"configure":"^orders\\..*","write":"^orders\\..*","read":"^orders\\..*"}}' \
  vhost_topics='{"orders":{"orders.events":{"write":"^order\\.(created|paid)$","read":".*"}}}'
```

## Шаг 4. Создайте политику

```bash
d8 stronghold policy write orders-rabbitmq - <<'POLICY'
path "rabbitmq/creds/orders" {
  capabilities = ["read"]
}
POLICY
```

Назначьте политику роли метода аутентификации, через который входит сервис, например роли метода Kubernetes, как в руководстве [«Динамические учётные данные PostgreSQL»](../postgresql/#шаг-4-создайте-роль-метода-kubernetes). Продление и отзыв собственных аренд разрешены встроенной политикой `default`.

## Шаг 5. Получите учётные данные

1. Запросите учётные данные:

   ```bash
   d8 stronghold read rabbitmq/creds/orders
   ```

   Пример вывода:

   ```text
   Key                Value
   ---                -----
   lease_id           rabbitmq/creds/orders/I39Hu8XXOombof4wiK5bKMn9
   lease_duration     1h
   lease_renewable    true
   password           3yNDBikgQvrkx2VA2zhq5IdSM7IWk1RyMYJr
   username           token-39669250-3894-8032-c420-3d58483ebfc4
   ```

1. Продлите аренду до истечения `lease_duration`:

   ```bash
   d8 stronghold lease renew rabbitmq/creds/orders/I39Hu8XXOombof4wiK5bKMn9
   ```

1. Когда учётные данные больше не нужны, отзовите аренду. Stronghold удалит пользователя в RabbitMQ:

   ```bash
   d8 stronghold lease revoke rabbitmq/creds/orders/I39Hu8XXOombof4wiK5bKMn9
   ```

## Шаг 6. Доставьте учётные данные в сервис

Stronghold Agent продлевает аренду и перерисовывает файл при получении новых учётных данных. Пример шаблона, который формирует URI подключения по AMQP:

```text
{{ with secret "rabbitmq/creds/orders" }}
AMQP_URL=amqps://{{ .Data.username }}:{{ .Data.password }}@rabbitmq.example.com:5671/orders
{{ end }}
```

Конфигурация Agent в режиме sidecar приведена в разделе [«Доставка секретов в поды Kubernetes»](../../delivery/kubernetes-workloads/#stronghold-agent). После достижения `max_ttl` сервис получает нового пользователя, а старый удаляется: сервис должен переподключаться к брокеру с новыми учётными данными.

## Проверка

1. Запросите учётные данные и сохраните имя пользователя и пароль:

   ```bash
   creds="$(d8 stronghold read -format=json rabbitmq/creds/orders)"
   MQ_USER="$(echo "$creds" | jq -r '.data.username')"
   MQ_PASSWORD="$(echo "$creds" | jq -r '.data.password')"
   LEASE_ID="$(echo "$creds" | jq -r '.lease_id')"
   ```

1. На узле RabbitMQ проверьте, что пользователь создан и аутентифицируется:

   ```bash
   rabbitmqctl list_users
   rabbitmqctl authenticate_user "$MQ_USER" "$MQ_PASSWORD"
   ```

1. Проверьте права пользователя на виртуальных хостах:

   ```bash
   rabbitmqctl list_user_permissions "$MQ_USER"
   ```

   В выводе должен быть только виртуальный хост `orders` с регулярными выражениями `^orders\..*`.

1. Отзовите аренду и убедитесь, что пользователь удалён:

   ```bash
   d8 stronghold lease revoke "$LEASE_ID"
   rabbitmqctl list_users
   ```

Если RabbitMQ работает в кластере DP, выполняйте `rabbitmqctl` в поде брокера, например `d8 k -n rabbitmq exec rabbitmq-0 -- rabbitmqctl list_users`.

## Очистка

1. Отзовите все выданные учётные данные роли:

   ```bash
   d8 stronghold lease revoke -prefix rabbitmq/creds/orders/
   ```

1. Удалите роль и политику, а при необходимости отключите механизм секретов:

   ```bash
   d8 stronghold delete rabbitmq/roles/orders
   d8 stronghold policy delete orders-rabbitmq
   d8 stronghold secrets disable rabbitmq
   ```

1. Удалите учётную запись Stronghold в RabbitMQ: `rabbitmqctl delete_user stronghold-admin`.
