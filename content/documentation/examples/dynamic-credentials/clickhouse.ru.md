---
title: "Динамические учётные данные ClickHouse для приложения"
linkTitle: "Динамические учётные данные ClickHouse"
description: "Выдача приложению временных пользователей ClickHouse с заранее подготовленной ролью через механизм секретов database: подключение, роль, политика, продление, отзыв и проверка клиентом clickhouse-client."
weight: 30
params:
  relatedLinks:
    - title: "Механизм секретов баз данных"
      url: ../../../user/secrets-engines/databases/overview/
    - title: "ClickHouse"
      url: ../../../user/secrets-engines/databases/clickhouse/
    - title: "Динамические учётные данные PostgreSQL"
      url: ../postgresql/
    - title: "Доставка секретов в поды Kubernetes"
      url: ../../delivery/kubernetes-workloads/
    - title: "Аренда, продление и отзыв"
      url: ../../../concepts/lease/
---

Сервис аналитики получает от Stronghold отдельного пользователя ClickHouse. Права пользователя задаются заранее подготовленной ролью ClickHouse, поэтому SQL-запросы в Stronghold остаются короткими, а набор прав меняется в одном месте — в ClickHouse.

## Цель

Настроить выдачу динамических учётных данных ClickHouse сервису `reports`: пользователь получает роль ClickHouse `analytics_reader` с правом чтения базы данных `analytics`, живёт 1 час с возможностью продления до 24 часов и удаляется после отзыва аренды.

## Предварительные требования

- Stronghold и токен с правами на настройку механизмов секретов и политик.
- Сервер ClickHouse с включённым управлением доступом через SQL (SQL-driven access control), доступный из Stronghold по нативному протоколу.
- База данных `analytics`.
- Клиент `clickhouse-client` и утилита `jq` на рабочей станции для проверки.

## Шаг 1. Подготовьте роль и служебную учётную запись в ClickHouse

Выполните в ClickHouse от имени администратора:

```sql
CREATE ROLE analytics_reader;
GRANT SELECT ON analytics.* TO analytics_reader;

CREATE USER stronghold IDENTIFIED BY '<initial_password>';
GRANT CREATE USER, ALTER USER, DROP USER ON *.* TO stronghold;
GRANT analytics_reader TO stronghold WITH ADMIN OPTION;
```

`WITH ADMIN OPTION` позволяет учётной записи `stronghold` назначать роль `analytics_reader` создаваемым пользователям.

Для работы плагина учётной записи `stronghold` нужны права на создание, изменение и удаление пользователей (`CREATE USER`, `ALTER USER`, `DROP USER`): они используются при выдаче учётных данных, при `rotate-root` и при отзыве аренды по умолчанию. Для выдачи роли нужно право назначать её, например `WITH ADMIN OPTION`, как в примере выше.

В кластере ClickHouse добавьте к запросам `ON CLUSTER '<имя_кластера>'`, как в примере из раздела [«ClickHouse»](../../../user/secrets-engines/databases/clickhouse/).

## Шаг 2. Настройте подключение

1. Включите механизм секретов, если он ещё не включён:

   ```bash
   d8 stronghold secrets enable database
   ```

1. Настройте подключение к ClickHouse:

   ```bash
   d8 stronghold write database/config/reports-clickhouse \
     plugin_name="clickhouse-database-plugin" \
     allowed_roles="reports" \
     connection_url="clickhouse://clickhouse.example.com:9440?username={{username}}&password={{password}}&secure=true" \
     username="stronghold" \
     password="<initial_password>"
   ```

   Параметр `secure=true` включает TLS. Порт `9440` — стандартный порт нативного протокола ClickHouse с TLS.

   Формат `connection_url`: `clickhouse://<хост>:<порт>?username=...&password=...`. Порт обязателен; параметры `secure` и `skip_verify` управляют TLS.

1. Смените пароль служебной учётной записи, чтобы он был известен только Stronghold:

   ```bash
   d8 stronghold write -force database/rotate-root/reports-clickhouse
   ```

## Шаг 3. Создайте роль

```bash
d8 stronghold write database/roles/reports \
  db_name="reports-clickhouse" \
  creation_statements="CREATE USER '{{name}}' IDENTIFIED BY '{{password}}'; \
    GRANT analytics_reader TO '{{name}}'; \
    SET DEFAULT ROLE analytics_reader TO '{{name}}';" \
  revocation_statements="DROP USER IF EXISTS '{{name}}';" \
  default_ttl="1h" \
  max_ttl="24h"
```

- `creation_statements` — создают пользователя, назначают ему роль `analytics_reader` и делают её ролью по умолчанию, чтобы права действовали сразу после входа;
- `revocation_statements` — удаляют пользователя при отзыве аренды;
- `default_ttl` и `max_ttl` — срок аренды при выдаче и продлении и максимальный срок жизни учётных данных.

Если в роли заданы `revocation_statements`, плагин выполняет их при отзыве аренды. Если они не заданы, плагин выполняет `DROP USER IF EXISTS`.

## Шаг 4. Создайте политику

```bash
d8 stronghold policy write reports-clickhouse - <<'POLICY'
path "database/creds/reports" {
  capabilities = ["read"]
}
POLICY
```

Назначьте политику роли метода аутентификации, через который входит сервис, например роли метода Kubernetes, как в руководстве [«Динамические учётные данные PostgreSQL»](../postgresql/#шаг-4-создайте-роль-метода-kubernetes).

## Шаг 5. Получите учётные данные

1. Запросите учётные данные:

   ```bash
   d8 stronghold read database/creds/reports
   ```

   Пример вывода:

   ```text
   Key                Value
   ---                -----
   lease_id           database/creds/reports/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
   lease_duration     1h
   lease_renewable    true
   password           SsnoaA-8Tv4t34f41baD
   username           v-token-reports-x
   ```

1. Продлите аренду до истечения `lease_duration`:

   ```bash
   d8 stronghold lease renew database/creds/reports/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
   ```

1. Когда учётные данные больше не нужны, отзовите аренду:

   ```bash
   d8 stronghold lease revoke database/creds/reports/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
   ```

Для доставки учётных данных в сервис и их автоматического продления используйте Stronghold Agent с шаблоном `{{ with secret "database/creds/reports" }}`. Конфигурация Agent приведена в разделе [«Доставка секретов в поды Kubernetes»](../../delivery/kubernetes-workloads/#stronghold-agent).

## Проверка

1. Запросите учётные данные и сохраните их в переменные:

   ```bash
   creds="$(d8 stronghold read -format=json database/creds/reports)"
   CH_USER="$(echo "$creds" | jq -r '.data.username')"
   CH_PASSWORD="$(echo "$creds" | jq -r '.data.password')"
   LEASE_ID="$(echo "$creds" | jq -r '.lease_id')"
   ```

1. Подключитесь клиентом `clickhouse-client` и проверьте текущего пользователя и его права:

   ```bash
   clickhouse-client --host clickhouse.example.com --port 9440 --secure \
     --user "$CH_USER" --password "$CH_PASSWORD" \
     --query "SELECT currentUser(); SHOW GRANTS;" --multiquery
   ```

   В выводе должна быть роль `analytics_reader`.

1. Убедитесь, что запись запрещена:

   ```bash
   clickhouse-client --host clickhouse.example.com --port 9440 --secure \
     --user "$CH_USER" --password "$CH_PASSWORD" \
     --query "CREATE TABLE analytics.t (x UInt8) ENGINE = Memory"
   ```

   Команда должна завершиться ошибкой `ACCESS_DENIED`.

1. Отзовите аренду и проверьте, что пользователь удалён:

   ```bash
   d8 stronghold lease revoke "$LEASE_ID"
   clickhouse-client --host clickhouse.example.com --port 9440 --secure \
     --user "$CH_USER" --password "$CH_PASSWORD" --query "SELECT 1"
   ```

   Подключение должно завершиться ошибкой аутентификации.

## Очистка

1. Отзовите все выданные учётные данные роли:

   ```bash
   d8 stronghold lease revoke -prefix database/creds/reports/
   ```

1. Удалите роль, политику и подключение:

   ```bash
   d8 stronghold policy delete reports-clickhouse
   d8 stronghold delete database/roles/reports
   d8 stronghold delete database/config/reports-clickhouse
   ```

1. При необходимости удалите в ClickHouse служебную учётную запись и роль: `DROP USER stronghold; DROP ROLE analytics_reader;`.
