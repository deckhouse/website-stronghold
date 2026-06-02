---
title: "Репликация KV1/KV2"
weight: 40
---

Репликация KV1/KV2 — это механизм автоматического копирования секретов между экземплярами Stronghold в режиме master-slave с использованием pull-модели.
Репликация поддерживается только для хранилищ KV1/KV2.
Синхронизация данных выполняется периодически по расписанию или в соответствии с настройками конкретного хранилища KV1/KV2.

Для работы репликации обеспечьте сетевую связанность с удалённым кластером Stronghold, настройте TLS-соединение и получите токен доступа.
Токен должен иметь права `list` и `read` для хранилища KV1/KV2 на удалённом кластере.

Репликация настраивается при монтировании нового хранилища KV1/KV2.
Имена удалённого и локального mount path могут не совпадать.
Репликацию также можно настроить между разными неймспейсами на локальном и удалённом кластерах.
Допускается репликация нескольких локальных хранилищ с разными именами из одного удалённого хранилища.

Если для локального хранилища KV1/KV2 настроена репликация, оно доступно только для чтения.
Запись, изменение и удаление секретов в таком хранилище недоступны.
Все изменения необходимо выполнять в исходном хранилище.
После очередного запуска репликации данные будут перенесены в локальное хранилище.

Если отключить репликацию, режим только для чтения будет снят.
После этого станут доступны операции добавления, изменения и удаления секретов.
Если затем снова включить репликацию, локальные изменения будут удалены или перезаписаны данными из исходного хранилища.

## Настройка репликации

Настройка репликации выполняется на стороне потребителя — в slave-кластере Stronghold.
Для этого задайте параметры репликации при монтировании нового хранилища KV1/KV2.

Поддерживаются следующие параметры:

- адрес удалённого кластера Stronghold;
- токен доступа к удалённому кластеру Stronghold;
- сертификат TLS или путь к сертификату TLS для подключения к удалённому кластеру Stronghold;
- имя namespace path, в котором находится хранилище KV1/KV2 на удалённом кластере Stronghold. По умолчанию используется `root`;
- имя mount path хранилища KV1/KV2 на удалённом кластере Stronghold;
- список secret path для репликации. По умолчанию реплицируются все секреты;
- период запуска репликации данных. По умолчанию — 1 минута;
- включение и отключение репликации. Для нового хранилища KV1/KV2 репликация включена по умолчанию;
- версия KV-хранилища для монтирования и репликации.

{{< alert level="warning" >}}
Версия локального и удалённого KV-хранилища должна совпадать.
Нельзя настроить репликацию `kv1` в `kv2` или `kv2` в `kv1`.
{{< /alert >}}

## Создание токена для репликации

Токен доступа к удалённому кластеру должен иметь права `list` и `read` для реплицируемых секретов.
Если токен поддерживает самопродление, Stronghold будет автоматически продлевать его на 30 дней, когда оставшийся TTL станет меньше 7 дней и не будет превышен параметр `maxTTL`.

Ниже приведён пример создания политики и токена для репликации из mount path `dev-secrets`, который находится в неймспейсе `ns_path_1`.
Выполните эти команды на исходном сервере:

```shell
d8 stronghold policy write -namespace=ns_path_1 replicate-dev-secrets - <<EOF
# Allow token to list/read secrets from dev-secrets.
path "dev-secrets/*" {
  capabilities = ["read", "list"]
}

# Allow token to read info about dev-secrets.
path "sys/mounts/dev-secrets" {
  capabilities = ["read"]
}

# Allow token to look up own properties.
path "auth/token/lookup-self" {
  capabilities = ["read"]
}

# Allow token to renew self.
path "auth/token/renew-self" {
  capabilities = ["update"]
}
EOF

d8 stronghold token create \
  -namespace=ns_path_1 \
  -policy=replicate-dev-secrets \
  -orphan=true \
  -period=30d
```

## Настройка репликации через CLI Stronghold

Для настройки репликации через CLI Stronghold выполните одну из следующих команд.

Без использования TLS-соединения:

```shell
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

С использованием TLS-соединения:

```shell
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

Используются следующие параметры:

- `-path` — имя mount path локального хранилища KV1/KV2, в которое будут скопированы данные из источника. Обязательный параметр. Пример: `my-mount-kv2`;
- `-src-address` — адрес удалённого кластера Stronghold. Обязательный параметр. Примеры: `127.0.0.1:8200`, `vault.mycompany.tld:8200`, `stronghold.mycompany.tld:443`;
- `-src-token` — токен доступа к удалённому кластеру Stronghold. Обязательный параметр. Пример: `z6VXjAi6F3vjaclHu99FLOcr`;
- `-src-namespace` — имя namespace path, в котором находится хранилище KV1/KV2 на удалённом кластере Stronghold. Необязательный параметр. По умолчанию используется `root`;
- `-src-mount-path` — имя mount path хранилища KV1/KV2 на удалённом кластере Stronghold. Обязательный параметр. Пример: `remote-mount-kv2`;
- `-src-secret-path` — список secret path для репликации. Необязательный параметр;
- `-src-ca-cert` — сертификат CA для установки TLS-соединения. Если сертификат находится в файле, используйте формат `-src-ca-cert=@ca-cert.pem`. Необязательный параметр;
- `-sync-period-min` — интервал в минутах, через который выполняется репликация хранилища. Необязательный параметр. По умолчанию — `60`;
- `-version` — версия KV-хранилища для монтирования и репликации. Обязательный параметр;
- `-namespace` — имя namespace path, в котором создаётся локальное хранилище KV1/KV2. Необязательный параметр. По умолчанию используется `root`.

{{< alert level="warning" >}}
Версия локального и удалённого KV-хранилища должна совпадать.
{{< /alert >}}

## Изменение настроек репликации через CLI Stronghold

Для редактирования доступны следующие параметры:

- токен доступа к удалённому кластеру Stronghold;
- сертификат TLS или путь к сертификату TLS для подключения к удалённому кластеру Stronghold;
- список secret path для репликации. Сейчас параметр не используется, поэтому по умолчанию реплицируются все секреты;
- период запуска репликации;
- включение и отключение репликации.

{{< alert level="warning" >}}
При изменении `secret path` в конфигурации репликации старый путь в локальном кластере останется без изменений, а новый будет добавлен.
Если старый и новый пути пересекаются, часть данных может быть перезаписана.

Например, до изменения было указано `-src-secret-path=[first-secret/one, second-sercet/two]`,
а после изменения — `-src-secret-path=[first-secret/two, second-sercet/two]`.
В этом случае данные в `first-secret/one` останутся в прежнем состоянии и больше не будут обновляться.
{{< /alert >}}

Для изменения настроек репликации через CLI Stronghold выполните команду:

```shell
d8 stronghold secrets tune \
  -src-token=<token_of_source_cluster> \
  -src-secret-path=<list_of_secret_paths_in_source_cluster> \
  -src-ca-cert=@<path_to_file_with_certificate> \
  -sync-enable=true \
  -sync-period-min=3 \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

Используются следующие параметры:

- `-src-token` — токен доступа к удалённому кластеру Stronghold. Обязательный параметр. Пример: `z6VXjAi6F3vjaclHu99FLOcr`;
- `-src-secret-path` — список secret path для репликации. Необязательный параметр;
- `-src-ca-cert` — сертификат CA для установки TLS-соединения. Если сертификат находится в файле, используйте формат `-src-ca-cert=@ca-cert.pem`. Необязательный параметр;
- `-sync-enable` — включает или отключает репликацию для локального mount path. Обязательный параметр;
- `-sync-period-min` — интервал в минутах, через который выполняется репликация хранилища. Необязательный параметр;
- `-namespace` — имя namespace path, в котором создано локальное хранилище KV1/KV2. Необязательный параметр. По умолчанию используется `root`.

Чтобы отключить репликацию, выполните команду:

```shell
d8 stronghold secrets tune \
  -sync-enable=false \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

Если передать остальные параметры репликации вместе с `-sync-enable=false`, они будут проигнорированы.

Чтобы включить репликацию, выполните команду:

```shell
d8 stronghold secrets tune \
  -sync-enable=true \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

В этом случае можно дополнительно передать и остальные параметры репликации.
Они будут учтены.

Чтобы прочитать настройки репликации, выполните команду:

```shell
d8 stronghold read \
  -namespace=<namespace_path_in_local_cluster> \
  sys/mounts/<mount_path>/tune
```

## Настройка репликации через API Stronghold

Для настройки репликации через API Stronghold вызовите API создания mount и добавьте в тело запроса конфигурацию репликации:

```shell
curl --header "X-Vault-Token: <token_for_local_cluster>" \
  --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  --request POST \
  --data '{
    "type": "<kv-v1>/<kv-v2>",
    "config": {
      "replication_config": {
        "src_address": "<address_of_source_cluster>",
        "src_token": "<token_of_source_cluster>",
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

Если удалённый кластер-источник не поддерживает TLS, параметр `"src_ca_cert"` передавать не нужно.
По умолчанию параметр `"src_secret_path"` имеет значение `"*"`, то есть реплицируются все secret path.

Используются следующие параметры:

- `local_stronghold_address` — адрес локального кластера Stronghold, на котором настраивается репликация;
- `token_for_local_cluster` — токен локального кластера, который имеет доступ к созданию mount;
- `namespace_path_in_local_cluster` — имя namespace path, в котором создаётся локальное хранилище KV1/KV2. Необязательный параметр. По умолчанию используется `root`;
- `local_mount_path_name` — имя mount path локального хранилища KV1/KV2, в которое будут скопированы данные из источника. Обязательный параметр. Пример: `my-mount-kv2`;
- `src_address` — адрес удалённого кластера Stronghold. Обязательный параметр. Примеры: `127.0.0.1:8200`, `vault.mycompany.tld:8200`, `stronghold.mycompany.tld:443`;
- `src_token` — токен доступа к удалённому кластеру Stronghold. Обязательный параметр. Пример: `z6VXjAi6F3vjaclHu99FLOcr`;
- `src_namespace` — имя namespace path, в котором находится хранилище KV1/KV2 на удалённом кластере Stronghold. Необязательный параметр. По умолчанию используется `root`;
- `src_mount_path` — имя mount path хранилища KV1/KV2 на удалённом кластере Stronghold. Обязательный параметр. Пример: `remote-mount-kv2`;
- `src_secret_path` — список secret path для репликации. Необязательный параметр;
- `src_ca_cert` — сертификат CA для установки TLS-соединения. Необязательный параметр;
- `sync_period_min` — интервал в минутах, через который выполняется репликация хранилища. Необязательный параметр. По умолчанию — `1`;
- `type` — версия KV-хранилища для монтирования и репликации. Обязательный параметр.

{{< alert level="warning" >}}
Версия локального и удалённого KV-хранилища должна совпадать.
{{< /alert >}}

## Изменение настроек репликации через API Stronghold

Для редактирования доступны следующие параметры:

- токен доступа к удалённому кластеру Stronghold;
- сертификат TLS или путь к сертификату TLS для подключения к удалённому кластеру Stronghold;
- список secret path для репликации. По умолчанию реплицируются все секреты;
- период запуска репликации;
- включение и отключение репликации.

{{< alert level="warning" >}}
При изменении `secret path` в конфигурации репликации старый путь в локальном кластере останется без изменений, а новый будет добавлен.
Если старый и новый пути пересекаются, часть данных может быть перезаписана.

Например, до изменения было указано `"src_secret_path"=["first-secret/one", "second-sercet/two"]`,
а после изменения — `"src_secret_path"=["first-secret/two", "second-sercet/two"]`.
В этом случае данные в `"first-secret/one"` останутся в прежнем состоянии и больше не будут обновляться.
{{< /alert >}}

Для изменения настроек репликации через API Stronghold вызовите API изменения mount и передайте новую конфигурацию репликации:

```shell
curl --header "X-Vault-Token: <token_for_local_cluster>" \
  --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  --request POST \
  --data '{
    "replication_config": {
      "src_token": "<token_of_source_cluster>",
      "src_ca_cert": "<tls_cert_for_source_cluster>",
      "src_secret_path": ["<list_of_secret_paths_in_source_cluster>"],
      "sync_period_min": <interval_in_minutes_for_synchronization_period>,
      "sync_enable": true
    }
  }' \
  <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```

Используются следующие параметры:

- `local_stronghold_address` — адрес локального кластера Stronghold, на котором настраивается репликация;
- `token_for_local_cluster` — токен локального кластера, который имеет доступ к изменению mount;
- `namespace_path_in_local_cluster` — имя namespace path, в котором создано локальное хранилище KV1/KV2. Необязательный параметр. По умолчанию используется `root`;
- `local_mount_path_name` — имя mount path локального хранилища KV1/KV2, в которое копируются данные из источника. Обязательный параметр. Пример: `my-mount-kv2`;
- `src_token` — токен доступа к удалённому кластеру Stronghold. Обязательный параметр. Пример: `z6VXjAi6F3vjaclHu99FLOcr`;
- `src_ca_cert` — сертификат CA для установки TLS-соединения. Необязательный параметр;
- `sync_period_min` — интервал в минутах, через который выполняется репликация хранилища. Необязательный параметр;
- `sync_enable` — включает или отключает репликацию для локального mount path. Обязательный параметр;
- `src_secret_path` — список secret path для репликации. Необязательный параметр.

Чтобы отключить репликацию, выполните запрос:

```shell
curl --header "X-Vault-Token: <token_for_local_cluster>" \
  --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  --request POST \
  --data '{
    "replication_config": {
      "sync_enable": false
    }
  }' \
  <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```

Если передать остальные параметры репликации вместе с `"sync_enable": false`, они будут проигнорированы.

Чтобы включить репликацию, выполните запрос:

```shell
curl --header "X-Vault-Token: <token_for_local_cluster>" \
  --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  --request POST \
  --data '{
    "replication_config": {
      "sync_enable": true
    }
  }' \
  <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```

В этом случае можно дополнительно передать и остальные параметры репликации.
Они будут учтены.

Чтобы прочитать настройки репликации, выполните запрос:

```shell
curl -X GET \
  -H "X-Vault-Token: <token_for_local_cluster>" \
  -H "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```
