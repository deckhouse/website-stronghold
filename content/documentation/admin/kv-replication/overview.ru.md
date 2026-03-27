---
title: "Репликация KV1/KV2"
linkTitle: "Введение"
weight: 10
description: "Руководство администратора по репликации KV1/KV2 в Stronghold."
---

## Описание механизма репликации KV1/KV2 в Stronghold

Под механизмом репликации понимается автоматическое копирование секретов между несколькими экземплярами Stronghold в режиме `master-slave` с использованием pull-модели.

Репликация поддерживается только для хранилищ `KV1` и `KV2`.

Синхронизация данных выполняется периодически по расписанию или в соответствии с индивидуальными настройками конкретного хранилища `KV1/KV2`.

Для работы механизма репликации необходимо:

- обеспечить сетевую связанность с удалённым кластером Stronghold;
- настроить корректное TLS-соединение, если оно используется;
- получить токен для доступа к удалённому кластеру;
- выдать этому токену права `list` и `read` для хранилищ `KV1/KV2` на удалённом кластере.

Для включения репликации необходимо задать параметры репликации при монтировании нового хранилища `KV1/KV2`.

Важно учитывать:

- названия удалённого и локального `mount path` могут не совпадать;
- репликация может быть настроена между разными пространствами имён на локальном и удалённом кластерах;
- можно настроить репликацию нескольких локальных хранилищ с разными именами на одно удалённое хранилище;
- если для локального хранилища настроена репликация, оно работает только в режиме `read-only`.

Запись, изменение и удаление секретов в локальном реплицируемом хранилище невозможны. Все изменения должны выполняться в исходном `master`-хранилище.

После следующего запуска синхронизации изменения из удалённого хранилища будут применены к локальному.

При отключении репликации статус `read-only` снимается, и операции добавления, изменения и удаления секретов становятся доступны локально.

При повторном включении репликации все локальные изменения будут удалены или перезаписаны данными из исходного хранилища.

## Настройка репликации KV1/KV2

Настройка выполняется на стороне потребителя, то есть на `slave`-кластере Stronghold, при монтировании нового хранилища `KV1/KV2`.

Настройки включают следующие параметры:

- адрес удалённого кластера Stronghold;
- токен для доступа к удалённому кластеру;
- `wrapping token` для безопасной передачи токена доступа;
- TLS-сертификат или путь к сертификату для подключения к удалённому кластеру;
- `namespace path`, в котором находится удалённое хранилище `KV1/KV2`;
- `mount path` удалённого хранилища `KV1/KV2`;
- список `secret path` для репликации;
- период запуска репликации;
- включение или отключение репликации;
- версию KV-хранилища.

{{< alert level="warning" >}}
Версия локального и удалённого KV-хранилища должна совпадать. Нельзя настроить репликацию `kv1` в `kv2` или `kv2` в `kv1`.
{{< /alert >}}

## Как создать токен для репликации

Токен для доступа к удалённому кластеру должен иметь права `list` и `read` для реплицируемых секретов.

Если выданный токен поддерживает самопродление, Stronghold автоматически продлевает его на 30 дней, когда оставшийся TTL становится меньше 7 дней и не превышен параметр `maxTTL`.

Ниже приведён пример создания политики и токена для репликации из `mount` `dev-secrets`, расположенного в пространстве имён `ns_path_1`:

```shell
d8 stronghold policy write -namespace=ns_path_1 replicate-dev-secrets - <<'EOF'
# Allow token to list/read secrets from dev-secrets
path "dev-secrets/*" {
  capabilities = ["read", "list"]
}

# Allow token to read info about dev-secrets
path "sys/mounts/dev-secrets" {
  capabilities = ["read"]
}

# Allow token to look up own properties
path "auth/token/lookup-self" {
  capabilities = ["read"]
}

# Allow token to renew self
path "auth/token/renew-self" {
  capabilities = ["update"]
}
EOF

d8 stronghold token create -namespace=ns_path_1 -policy=replicate-dev-secrets -orphan=true -period=30d
```

## Создание wrapping token

Настройку репликации рекомендуется выполнять через `wrapping token`.

`Wrapping token` — это одноразовый токен с ограниченным TTL, внутри которого передаётся реальный токен доступа. При настройке репликации на кластер-потребитель передаётся именно wrapping token, после чего кластер-потребитель выполняет `unwrap` на кластере-источнике и получает реальный токен для репликации.

Пример создания wrapping token на исходном кластере:

```shell
d8 stronghold token create \
  -namespace=ns_path_1 \
  -policy=replicate-dev-secrets \
  -orphan=true \
  -period=30d \
  -wrap-ttl=5m \
  -field=wrapping_token
```

Полученный `wrapping token` нужно передать при настройке репликации на кластере-потребителе.

## Настройка репликации через CLI

### Без TLS

```shell
d8 stronghold secrets enable \
  -path=<local_mount_path_name> \
  -src-address=<address_of_source_cluster> \
  -src-wrapping-token=<wrapping_token_from_source_cluster> \
  -src-namespace=<namespace_path_in_source_cluster> \
  -src-mount-path=<mount_path_in_source_cluster> \
  -sync-period-min=<interval_in_minutes> \
  -version=<1|2> \
  -namespace=<namespace_path_in_local_cluster> \
  kv
```

### С TLS

```shell
d8 stronghold secrets enable \
  -path=<local_mount_path_name> \
  -src-address=<address_of_source_cluster> \
  -src-wrapping-token=<wrapping_token_from_source_cluster> \
  -src-namespace=<namespace_path_in_source_cluster> \
  -src-mount-path=<mount_path_in_source_cluster> \
  -src-ca-cert=@<path_to_file_with_certificate> \
  -sync-period-min=<interval_in_minutes> \
  -version=<1|2> \
  -namespace=<namespace_path_in_local_cluster> \
  kv
```

Пояснение параметров:

- `-path` — имя `mount path` локального KV-хранилища, куда будут копироваться данные;
- `-src-address` — адрес удалённого кластера Stronghold;
- `-src-token` — токен для доступа к удалённому кластеру; обязателен, если не указан `-src-wrapping-token`;
- `-src-wrapping-token` — wrapping token, который будет развёрнут на удалённом кластере; обязателен, если не указан `-src-token`;
- `-src-namespace` — `namespace path` удалённого кластера; по умолчанию `root`;
- `-src-mount-path` — имя `mount path` удалённого KV-хранилища;
- `-src-secret-path` — список `secret path` для репликации;
- `-src-ca-cert` — сертификат CA для TLS-соединения; если сертификат в файле, используйте форму `@ca-cert.pem`;
- `-sync-period-min` — период запуска синхронизации в минутах; по умолчанию `1`;
- `-version` — версия KV-хранилища (`1` или `2`);
- `-namespace` — `namespace path` локального кластера; по умолчанию `root`.

## Изменение настроек репликации через CLI

Для редактирования доступны:

- токен доступа к удалённому кластеру;
- wrapping token для обновления токена;
- TLS-сертификат;
- список `secret path`;
- период запуска репликации;
- включение и отключение репликации.

{{< alert level="warning" >}}
При изменении `secret path` старый путь в локальном кластере остаётся неизменным, а новый добавляется. Если старые и новые пути пересекаются, новые данные могут частично перезаписать существующие.
{{< /alert >}}

Пример изменения настроек:

```shell
d8 stronghold secrets tune \
  -src-wrapping-token=<wrapping_token_from_source_cluster> \
  -src-secret-path=<list_of_secret_paths_in_source_cluster> \
  -src-ca-cert=@<path_to_file_with_certificate> \
  -sync-enable=true \
  -sync-period-min=<interval_in_minutes> \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

Для отключения репликации:

```shell
d8 stronghold secrets tune \
  -sync-enable=false \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

Для повторного включения:

```shell
d8 stronghold secrets tune \
  -sync-enable=true \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

Для чтения текущих настроек:

```shell
d8 stronghold read \
  -namespace=<namespace_path_in_local_cluster> \
  sys/mounts/<mount_path>/tune
```

## Настройка через API

Для создания нового `mount` с включённой репликацией используйте API создания `mount` и передайте конфигурацию в `replication_config`:

```shell
curl \
  --header "X-Vault-Token: <token_for_local_cluster>" \
  --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  --request POST \
  --data '{
    "type": "<kv-v1|kv-v2>",
    "config": {
      "replication_config": {
        "src_address": "<address_of_source_cluster>",
        "src_wrapping_token": "<wrapping_token_from_source_cluster>",
        "src_ca_cert": "<tls_cert_for_source_cluster>",
        "src_namespace": "<namespace_path_in_source_cluster>",
        "src_mount_path": "<mount_path_in_source_cluster>",
        "src_secret_path": ["<list_of_secret_paths_in_source_cluster>"],
        "sync_period_min": <interval_in_minutes_for_synchronization_period>
      }
    }
  }' \
  <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>
```

Если удалённый кластер не использует TLS, параметр `src_ca_cert` можно не передавать.

По умолчанию `src_secret_path="*"` и реплицируются все `secret path`.

Основные параметры:

- `local_stronghold_address` — адрес локального кластера Stronghold;
- `token_for_local_cluster` — токен, который даёт право создавать `mount`;
- `namespace_path_in_local_cluster` — `namespace path` локального кластера;
- `local_mount_path_name` — имя локального `mount path`;
- `src_address` — адрес удалённого кластера;
- `src_wrapping_token` — wrapping token для получения реального токена репликации;
- `src_token` — токен доступа к удалённому кластеру;
- `src_namespace` — `namespace path` удалённого кластера;
- `src_mount_path` — `mount path` удалённого KV-хранилища;
- `src_secret_path` — список `secret path` для репликации;
- `src_ca_cert` — CA-сертификат для TLS;
- `sync_period_min` — период запуска синхронизации;
- `type` — тип KV-хранилища.

## Изменение настроек через API

Для редактирования используйте endpoint настройки `mount`:

```shell
curl \
  --header "X-Vault-Token: <token_for_local_cluster>" \
  --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  --request POST \
  --data '{
    "replication_config": {
      "src_wrapping_token": "<wrapping_token_from_source_cluster>",
      "src_ca_cert": "<tls_cert_for_source_cluster>",
      "src_secret_path": ["<list_of_secret_paths_in_source_cluster>"],
      "sync_period_min": <interval_in_minutes_for_synchronization_period>,
      "sync_enable": true
    }
  }' \
  <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```

Для отключения репликации:

```shell
curl \
  --header "X-Vault-Token: <token_for_local_cluster>" \
  --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  --request POST \
  --data '{
    "replication_config": {
      "sync_enable": false
    }
  }' \
  <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```

Для включения репликации:

```shell
curl \
  --header "X-Vault-Token: <token_for_local_cluster>" \
  --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  --request POST \
  --data '{
    "replication_config": {
      "sync_enable": true
    }
  }' \
  <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```

Для чтения текущих настроек:

```shell
curl -X GET \
  -H "X-Vault-Token: <token_for_local_cluster>" \
  -H "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```
