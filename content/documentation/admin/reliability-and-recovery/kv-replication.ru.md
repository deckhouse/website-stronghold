---
title: "Репликация KV"
linkTitle: "Репликация KV"
description: "Руководство администратора по репликации KV v1 и KV v2 в Deckhouse Stronghold"
weight: 50
---

Под механизмом репликации понимается автоматическое копирование секретов между несколькими экземплярами **Deckhouse Stronghold** в режиме `master-slave` с использованием pull-модели.

Репликация поддерживается только для хранилищ:

- `KV v1`;
- `KV v2`.

Синхронизация данных выполняется периодически по расписанию или в соответствии с индивидуальными настройками конкретного хранилища `KV`.

> Примечание  
> Возможность репликации `KV` зависит от редакции Deckhouse Stronghold. Актуальная информация о доступности функции приведена в разделе [Редакции](../../about/editions/).

## Для чего нужна репликация KV

Репликация `KV` полезна в следующих сценариях:

- распространение секретов между несколькими кластерами Stronghold;
- доставка одинаковых секретов в разные контуры эксплуатации;
- построение управляемой схемы чтения секретов из выделенного кластера-потребителя;
- разделение точки записи и точек чтения;
- снижение числа прикладных систем, напрямую обращающихся к исходному кластеру.

Репликация `KV` не заменяет:

- резервное копирование;
- восстановление после аварии;
- кворум `Raft`;
- полную межкластерную репликацию всего инстанса Stronghold.

Это отдельный механизм, предназначенный именно для копирования данных `KV`.

## Общая модель работы

Репликация `KV` работает по модели **pull**:

- исходный кластер (`master`) хранит первичные данные;
- кластер-получатель (`slave`) периодически запрашивает данные из исходного `KV`-хранилища;
- локальное хранилище на стороне получателя становится реплицируемым и работает в режиме `read-only`.

Для работы механизма репликации необходимо:

- обеспечить сетевую связанность с удалённым кластером Stronghold;
- настроить корректное TLS-соединение, если оно используется;
- получить токен для доступа к удалённому кластеру;
- выдать этому токену права `list` и `read` для хранилищ `KV v1` или `KV v2` на удалённом кластере.

## Что важно учитывать

При использовании репликации `KV` учитывайте следующее:

- названия удалённого и локального `mount path` могут не совпадать;
- репликация может быть настроена между разными пространствами имён на локальном и удалённом кластерах;
- можно настроить репликацию нескольких локальных хранилищ с разными именами на одно удалённое хранилище;
- если для локального хранилища настроена репликация, оно работает только в режиме `read-only`;
- запись, изменение и удаление секретов в локальном реплицируемом хранилище невозможны;
- все изменения должны выполняться в исходном `master`-хранилище;
- после следующего запуска синхронизации изменения из удалённого хранилища будут применены к локальному;
- при отключении репликации статус `read-only` снимается;
- при повторном включении репликации все локальные изменения будут удалены или перезаписаны данными из исходного хранилища.

> Предупреждение  
> Репликация `KV v1` в `KV v2` и `KV v2` в `KV v1` не поддерживается. Версия локального и удалённого `KV`-хранилища должна совпадать.

## KV v1 и KV v2 в контексте репликации

Stronghold поддерживает репликацию как для `KV v1`, так и для `KV v2`, однако при проектировании нужно учитывать версию механизма секретов.

### KV v1

Для `KV v1` реплицируются данные в их базовой форме без встроенной версионности.

Такой вариант проще по структуре, но не поддерживает те возможности версионирования, которые есть в `KV v2`.

### KV v2

Для `KV v2` реплицируются данные с учётом модели `KV v2`, включая особенности хранения и обращения к данным через соответствующие пути.

При использовании `KV v2` особенно важно:

- не перепутать mount-тип на источнике и приёмнике;
- учитывать, что структура данных и API отличаются от `KV v1`;
- при диагностике проблем обращать внимание на конкретный путь и тип хранилища.

## Где настраивается репликация

Настройка выполняется на стороне потребителя, то есть на `slave`-кластере Stronghold, при монтировании нового хранилища `KV v1` или `KV v2`.

В конфигурацию входят следующие параметры:

- адрес удалённого кластера Stronghold;
- токен для доступа к удалённому кластеру;
- `wrapping token` для безопасной передачи токена доступа;
- TLS-сертификат или путь к сертификату для подключения к удалённому кластеру;
- `namespace path`, в котором находится удалённое хранилище;
- `mount path` удалённого хранилища;
- список `secret path` для репликации;
- период запуска репликации;
- включение или отключение репликации;
- версия `KV`-хранилища.

## Как создать токен для репликации

Токен для доступа к удалённому кластеру должен иметь права `list` и `read` для реплицируемых секретов.

Если выданный токен поддерживает самопродление, Stronghold автоматически продлевает его на 30 дней, когда оставшийся TTL становится меньше 7 дней и не превышен параметр `maxTTL`.

Ниже приведён пример создания политики и токена для репликации из `mount` `dev-secrets`, расположенного в пространстве имён `ns_path_1`:

```bash
d8 stronghold policy write -namespace=ns_path_1 replicate-dev-secrets - <<'EOF'
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

## Использование wrapping token

Настройку репликации рекомендуется выполнять через `wrapping token`.

`Wrapping token` — это одноразовый токен с ограниченным TTL, внутри которого передаётся реальный токен доступа. При настройке репликации на кластер-получатель передаётся именно wrapping token, после чего кластер-получатель выполняет `unwrap` на кластере-источнике и получает реальный токен для репликации.

Пример создания wrapping token на исходном кластере:

```bash
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

```bash
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

```bash
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

### Пояснение параметров CLI

- `-path` — имя `mount path` локального `KV`-хранилища, куда будут копироваться данные;
- `-src-address` — адрес удалённого кластера Stronghold;
- `-src-token` — токен для доступа к удалённому кластеру; обязателен, если не указан `-src-wrapping-token`;
- `-src-wrapping-token` — wrapping token, который будет развёрнут на удалённом кластере; обязателен, если не указан `-src-token`;
- `-src-namespace` — `namespace path` удалённого кластера; по умолчанию `root`;
- `-src-mount-path` — имя `mount path` удалённого `KV`-хранилища;
- `-src-secret-path` — список `secret path` для репликации;
- `-src-ca-cert` — сертификат CA для TLS-соединения; если сертификат в файле, используйте форму `@ca-cert.pem`;
- `-sync-period-min` — период запуска синхронизации в минутах; по умолчанию `1`;
- `-version` — версия `KV`-хранилища (`1` или `2`);
- `-namespace` — `namespace path` локального кластера; по умолчанию `root`.

## Изменение настроек репликации через CLI

Для редактирования доступны:

- токен доступа к удалённому кластеру;
- wrapping token для обновления токена;
- TLS-сертификат;
- список `secret path`;
- период запуска репликации;
- включение и отключение репликации.

> Предупреждение  
> При изменении `secret path` старый путь в локальном кластере остаётся неизменным, а новый добавляется. Если старые и новые пути пересекаются, новые данные могут частично перезаписать существующие.

Пример изменения настроек:

```bash
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

```bash
d8 stronghold secrets tune \
  -sync-enable=false \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

Для повторного включения:

```bash
d8 stronghold secrets tune \
  -sync-enable=true \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

Для чтения текущих настроек:

```bash
d8 stronghold read \
  -namespace=<namespace_path_in_local_cluster> \
  sys/mounts/<mount_path>/tune
```

## Настройка репликации через API

Для создания нового `mount` с включённой репликацией используйте API создания `mount` и передайте конфигурацию в `replication_config`:

```bash
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

### Основные параметры API

- `local_stronghold_address` — адрес локального кластера Stronghold;
- `token_for_local_cluster` — токен, который даёт право создавать `mount`;
- `namespace_path_in_local_cluster` — `namespace path` локального кластера;
- `local_mount_path_name` — имя локального `mount path`;
- `src_address` — адрес удалённого кластера;
- `src_wrapping_token` — wrapping token для получения реального токена репликации;
- `src_token` — токен доступа к удалённому кластеру;
- `src_namespace` — `namespace path` удалённого кластера;
- `src_mount_path` — `mount path` удалённого `KV`-хранилища;
- `src_secret_path` — список `secret path` для репликации;
- `src_ca_cert` — CA-сертификат для TLS;
- `sync_period_min` — период запуска синхронизации;
- `type` — тип `KV`-хранилища.

## Изменение настроек через API

Для редактирования используйте endpoint настройки `mount`:

```bash
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

```bash
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

```bash
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

```bash
curl -X GET \
  -H "X-Vault-Token: <token_for_local_cluster>" \
  -H "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```

## Что проверять после включения репликации

После настройки репликации рекомендуется проверить:

- что локальный `mount` создан корректно;
- что версия локального и удалённого `KV` совпадает;
- что синхронизация выполняется без ошибок;
- что локальное хранилище действительно находится в режиме `read-only`;
- что секреты появляются на стороне получателя в ожидаемом виде;
- что TLS и токен доступа к исходному кластеру работают корректно;
- что replication token можно продлить, если используется сценарий с self-renewal.

## Риски и эксплуатационные замечания

При эксплуатации репликации `KV` учитывайте следующие риски:

- потеря сетевой связности с исходным кластером приведёт к остановке обновления данных;
- ошибки TLS или истечение срока действия токена нарушат синхронизацию;
- неправильный выбор `secret path` может привести к репликации избыточного или нежелательного набора данных;
- локальные изменения в реплицируемом `mount` будут потеряны при повторном включении репликации;
- при пересечении старых и новых путей после перенастройки возможна частичная перезапись данных.

## Практические рекомендации

- используйте отдельный токен только для репликации;
- выдавайте этому токену минимально необходимые права;
- предпочитайте передачу токена через `wrapping token`;
- документируйте соответствие между локальными и удалёнными `mount path`;
- перед включением репликации проверяйте совпадение версии `KV`;
- для production-сценариев обязательно используйте TLS;
- заранее определите политику контроля и ротации replication token;
- не рассматривайте репликацию `KV` как замену резервному копированию.

## Что дальше

Для работы с изоляцией данных и разграничением административных областей перейдите в раздел [Пространства имён](../basic-exploitation/namespaces/), а для резервного копирования данных Stronghold используйте раздел [Резервное копирование](../reliability-and-recovery/backups/).
