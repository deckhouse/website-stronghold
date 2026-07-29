---
title: "Performance standby"
linkTitle: "Performance standby"
weight: 40
params:
  edition: ee
description: "Обслуживание запросов на чтение неактивными HA-узлами внутри одного кластера Stronghold."
---

Performance standby позволяет неактивным HA-узлам **внутри одного кластера**
обслуживать запросы на чтение локально, а не перенаправлять каждый запрос на
активный узел. Это распределяет нагрузку чтения между всеми узлами кластера и
снижает задержки, при этом записи и другие операции, изменяющие состояние,
по-прежнему перенаправляются на активный узел.

В отличие от [репликации Performance](../performance/) и
[Disaster recovery](../disaster-recovery/), которые работают между отдельными
кластерами, performance standby работает внутри одного кластера и не требует
межкластерной настройки.

## Как это работает

- В HA-кластере один узел активный, остальные — standby-узлы.
- В Stronghold EE каждый standby-узел становится performance standby: он
  обрабатывает запросы на чтение по своему локальному состоянию.
- Запросы, изменяющие состояние (записи, удаления), перенаправляются на активный
  узел и применяются на нём.

Performance standby включён по умолчанию в Enterprise Edition. Дополнительная
настройка не требуется — достаточно запустить HA-кластер на integrated Raft
storage.

## Проверка performance standby

Отправьте запрос на чтение напрямую на standby-узел с заголовком
`X-Vault-No-Request-Forwarding`.
Performance standby всё равно вернёт `200` на чтение, потому что обслуживает его
локально; запрос,
которому нужен активный узел, при подавлении перенаправления вернёт статус не из
диапазона 2xx.

```shell
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  --header "X-Vault-No-Request-Forwarding: true" \
  "${STANDBY_ADDR}/v1/secret/data/hello"
```

Роль узла также можно проверить в выводе health и статуса — performance standby
сообщает о себе как о standby-узле, обслуживающем чтения.

## Отключение performance standby

Чтобы отключить performance standby, не затрагивая репликацию, задайте в
конфигурации сервера параметр:

```hcl
disable_performance_standby = true
```

С этим параметром standby-узлы перенаправляют все запросы на активный узел, как
это было бы без Enterprise-возможности standby.

{{< alert level="info" >}}
Параметр `disable_performance_standby` применяется в конфигурации сервера
Standalone-установки Stronghold. Отключение всего пайплайна репликации через
`disable_wal_replication = true` также отключает performance standby.
{{< /alert >}}
