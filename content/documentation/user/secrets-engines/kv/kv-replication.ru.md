---
title: "Репликация KV"
linkTitle: "Репликация KV"
weight: 40
description: "Настройка репликации KV v1 и KV v2 в Deckhouse Stronghold."
---

Под механизмом репликации понимается автоматическое копирование секретов между несколькими экземплярами **Deckhouse Stronghold** в режиме `master-slave` с использованием pull-модели. Репликация поддерживается только для хранилищ `KV v1` и `KV v2`.

Синхронизация данных выполняется периодически по расписанию или в соответствии с индивидуальными настройками конкретного хранилища `KV`.

## Как работает репликация

Для работы механизма репликации необходимо:

- обеспечить сетевую связанность с удалённым кластером Stronghold;
- настроить корректное TLS-соединение, если оно используется;
- получить токен для доступа к удалённому кластеру;
- выдать этому токену права `list` и `read` для хранилищ `KV v1` или `KV v2` на удалённом кластере.

Репликация настраивается на стороне потребителя (`slave`), то есть локальный `mount` создаётся и синхронизируется из удалённого `master`.

## Что важно учитывать

- названия удалённого и локального `mount path` могут не совпадать;
- репликация может быть настроена между разными пространствами имён;
- можно настроить репликацию нескольких локальных хранилищ с разными именами на одно удалённое хранилище;
- локальное реплицируемое хранилище работает только в режиме `read-only`;
- все изменения должны выполняться в исходном `master`-хранилище;
- при повторном включении репликации локальные изменения будут удалены или перезаписаны данными из источника.

> Предупреждение  
> Версия локального и удалённого `KV`-хранилища должна совпадать. Нельзя настроить репликацию `kv1` в `kv2` или `kv2` в `kv1`.

## Параметры репликации

Настройки репликации включают:

- адрес удалённого кластера Stronghold;
- токен для доступа к удалённому кластеру;
- TLS-сертификат или путь к сертификату;
- `namespace path` удалённого кластера;
- `mount path` удалённого `KV`-хранилища;
- список `secret path` для репликации;
- период запуска синхронизации;
- включение или отключение репликации;
- версию `KV`-хранилища.

## Как создать токен для репликации

Токен для доступа к удалённому кластеру должен иметь права `list` и `read` для реплицируемых секретов.

Если токен поддерживает самопродление, Stronghold может автоматически продлевать его, когда оставшийся TTL становится меньше 7 дней и не превышен `maxTTL`.

Пример политики и токена для репликации:

```bash
d8 stronghold policy write -namespace=ns_path_1 replicate-dev-secrets - <<EOF
path "dev-secrets/*" {
  capabilities = ["read", "list"]
}

path "sys/mounts/dev-secrets" {
  capabilities = ["read"]
}

path "auth/token/lookup-self" {
  capabilities = ["read"]
}

path "auth/token/renew-self" {
  capabilities = ["update"]
}
EOF

d8 stronghold token create -namespace=ns_path_1 -policy=replicate-dev-secrets -orphan=true -period=30d
```

## Настройка репликации через CLI

### Без TLS

```bash
d8 stronghold secrets enable \
 -path=<local_mount_path_name> \
 -src-address=<address_of_source_cluster> \
 -src-token=<token_of_source_cluster> \
 -src-namespace=<namespace_path_in_source_cluster> \
 -src-mount-path=<mount_path_in_source_cluster> \
 -sync-period-min=3 \
 -version=<1/2> \
 -namespace=<namespace_path_in_local_cluster> \
 kv
```

### С TLS

```bash
d8 stronghold secrets enable \
 -path=<local_mount_path_name> \
 -src-address=<address_of_source_cluster> \
 -src-token=<token_of_source_cluster> \
 -src-namespace=<namespace_path_in_source_cluster> \
 -src-mount-path=<mount_path_in_source_cluster> \
 -src-ca-cert=@<path_to_file_with_certificate> \
 -sync-period-min=3 \
 -version=<1/2> \
 -namespace=<namespace_path_in_local_cluster> \
 kv
```

### Пояснение параметров

- `-path` — имя локального `mount path`, куда будут копироваться данные;
- `-src-address` — адрес удалённого кластера Stronghold;
- `-src-token` — токен для доступа к удалённому кластеру;
- `-src-namespace` — `namespace path` удалённого кластера;
- `-src-mount-path` — имя `mount path` удалённого `KV`-хранилища;
- `-src-secret-path` — список `secret path` для репликации;
- `-src-ca-cert` — CA-сертификат для TLS;
- `-sync-period-min` — период синхронизации в минутах;
- `-version` — версия `KV` (`1` или `2`);
- `-namespace` — `namespace path` локального кластера.

## Изменение настроек через CLI

Для редактирования доступны:

- токен;
- TLS-сертификат;
- список `secret path`;
- период синхронизации;
- включение и отключение репликации.

Пример изменения:

```bash
d8 stronghold secrets tune \
 -src-token=<token_of_source_cluster> \
 -src-secret-path=<list_of_secret_paths_in_source_cluster> \
 -src-ca-cert=@<path_to_file_with_certificate> \
 -sync-enable=true \
 -sync-period-min=3 \
 -namespace=<namespace_path_in_local_cluster> \
 <local_mount_path_name>
```

### Отключение репликации

```bash
d8 stronghold secrets tune -sync-enable=false -namespace=<namespace_path_in_local_cluster> <local_mount_path_name>
```

### Повторное включение

```bash
d8 stronghold secrets tune -sync-enable=true -namespace=<namespace_path_in_local_cluster> <local_mount_path_name>
```

### Чтение текущих настроек

```bash
d8 stronghold read -namespace=<namespace_path_in_local_cluster> sys/mounts/<mount_path>/tune
```

> Предупреждение  
> При изменении `secret path` старые пути в локальном кластере сохраняются, а новые добавляются. Если пути пересекаются, данные могут частично перезаписаться.

## Настройка через API

Для настройки репликации через API используйте создание `mount` с передачей `replication_config`:

```bash
curl --header "X-Vault-Token: <token_for_local_cluster>" \
     --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
     --request POST \
     --data '{
  "type" : "<kv-v1>/<kv-v2>",
  "config" : {
    "replication_config" : {
      "src_address" : "<address_of_source_cluster>",
      "src_token" : "<token_of_source_cluster>",
      "src_ca_cert" : "<tls_cert_for_source_cluster>",
      "src_namespace" : "<namespace_path_in_source_cluster>",
      "src_mount_path" : "<mount_path_in_source_cluster>",
      "src_secret_path" : [ "<list_of_secret_paths_in_source_cluster>" ],
      "sync_period_min" : <interval_in_minutes_for_synchronization_period>
    }
  }
}' <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>
```

Если источник не использует TLS, параметр `src_ca_cert` можно не передавать.

## Изменение настроек через API

Для изменения настроек используйте endpoint `.../tune`:

```bash
curl --header "X-Vault-Token: <token_for_local_cluster>" \
     --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
     --request POST \
     --data '{
        "replication_config" : {
          "src_token" : "<token_of_source_cluster>",
          "src_ca_cert" : "<tls_cert_for_source_cluster>",
          "src_secret_path" : [ "<list_of_secret_paths_in_source_cluster>" ],
          "sync_period_min" : <interval_in_minutes_for_synchronization_period>,
          "sync_enable" : true
        }
    }' \
    <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```

### Отключение через API

```bash
curl --header "X-Vault-Token: <token_for_local_cluster>" \
     --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
     --request POST \
     --data '{
        "replication_config" : {
          "sync_enable" : false
        }
    }' \
    <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```

### Включение через API

```bash
curl --header "X-Vault-Token: <token_for_local_cluster>" \
     --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
     --request POST \
     --data '{
        "replication_config" : {
          "sync_enable" : true
        }
    }' \
    <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```

### Чтение настроек через API

```bash
curl -X GET \
     -H "X-Vault-Token: <token_for_local_cluster>" \
     -H "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
     <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```

## Практические рекомендации

- Используйте отдельный токен только для репликации.
- Давайте этому токену только `read` и `list`.
- Всегда проверяйте совпадение версии `KV` на источнике и приёмнике.
- Используйте TLS, если кластеры взаимодействуют через сеть вне доверенного локального контура.
- Учитывайте, что локальное реплицируемое хранилище работает только в режиме чтения.
- Не рассматривайте репликацию как замену резервному копированию.

## Что дальше

Если вам нужен базовый обзор механизма `KV`, используйте раздел [Обзор](../kv/).  
Если нужна работа с локальным хранилищем без репликации, перейдите в [KV v1](./kv-v1/) или [KV v2](./kv-v2/).
