---
title: "Серверные журналы"
description: "Уровни и формат серверных журналов Stronghold, где их найти в Linux и DKP и как изменить уровень логирования без перезапуска."
weight: 20
---

Серверные журналы описывают работу самого процесса Stronghold: запуск, распечатывание, выборы лидера Raft, ошибки плагинов и подключений. Они не заменяют [журналы аудита](../../audit/overview/), в которых фиксируются запросы клиентов.

## Уровни логирования

Уровень задаётся параметром `log_level` в конфигурационном файле или переменной окружения `VAULT_LOG_LEVEL` (см. [«Настройка»](../../../install/standalone/configuration/)). Поддерживаемые значения в порядке возрастания подробности:

| Уровень | Когда использовать |
| --- | --- |
| `error` | Только ошибки |
| `warn` | Ошибки и предупреждения |
| `info` | Значение по умолчанию, рекомендуется для production |
| `debug` | Диагностика проблем, кратковременно |
| `trace` | Детальная диагностика, кратковременно. Создаёт большой объём записей |

В DP параметр `log_level` в конфигурации модуля не задан, поэтому используется уровень по умолчанию (`info`). Чтобы временно повысить его, используйте эндпоинт [`/sys/loggers`](#через-api-без-перезапуска).

{{< alert level="warning" >}}
Не оставляйте уровни `debug` и `trace` включёнными надолго: они увеличивают объём журналов и нагрузку на диск.
{{< /alert >}}

## Формат журналов

Параметр `log_format` принимает значения `standard` (по умолчанию) и `json`. Для отправки журналов в систему централизованного логирования используйте формат `json`:

```hcl
log_level  = "info"
log_format = "json"
```

Пример записи в формате JSON:

```json
{"@level":"info","@message":"attempting to join possible raft leader node","@module":"core","@timestamp":"2025-10-20T10:54:02.578963Z","leader_addr":"https://stronghold-0.stronghold-internal:8300"}
```

В DP в конфигурации модуля зашито значение `log_format = "json"`.

Чтобы писать журналы в файл, задайте параметр `log_file` и параметры ротации `log_rotate_duration`, `log_rotate_bytes`, `log_rotate_max_files`.

## Просмотр журналов

{{< tabs name="stronghold_logs_view" >}}
{{% tab name="Stronghold в Linux" %}}

При запуске через systemd-unit `stronghold.service` журналы попадают в journald:

```shell
journalctl -u stronghold.service -f
```

Журналы за определённый период:

```shell
journalctl -u stronghold.service --since "1 hour ago"
```

{{% /tab %}}
{{% tab name="Stronghold в DKP" %}}

Stronghold работает в пространстве имён `d8-stronghold`. Список подов:

```shell
d8 k -n d8-stronghold get pods -o wide
```

В поде `stronghold-N` есть контейнер `stronghold`, sidecar `kube-rbac-proxy` (метрики) и init-контейнер `plugin-fetcher` (загрузка плагинов). Основной контейнер указывается через `-c`:

```shell
d8 k -n d8-stronghold logs stronghold-0 -c stronghold -f
```

Журналы предыдущего экземпляра контейнера после перезапуска:

```shell
d8 k -n d8-stronghold logs stronghold-0 -c stronghold --previous
```

Чтобы получить журналы активного узла (например, при диагностике алертов автоматических снимков), используйте сервис `stronghold-active`:

```shell
d8 k -n d8-stronghold logs svc/stronghold-active -c stronghold
```

{{% /tab %}}
{{< /tabs >}}

## Изменение уровня логирования

### Через конфигурационный файл

1. Измените значение `log_level` в конфигурационном файле.
1. Отправьте процессу сигнал `SIGHUP`:

   ```shell
   systemctl reload stronghold
   ```

   Команда использует `ExecReload=/bin/kill -HUP $MAINPID` из systemd-unit. При получении `SIGHUP` уровень логирования обновляется, флаги CLI и переменные окружения игнорируются. Не все подсистемы (например, плагины) поддерживают динамическое изменение уровня.

### Через API без перезапуска

Эндпоинт [`/sys/loggers`](../../../reference/api/system/#post-sysloggers) меняет уровень логирования на работающем узле. Запрос применяется к узлу, на который он отправлен.

Повысить уровень для всех подсистем:

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  --request POST \
  --data '{"level": "debug"}' \
  "${STRONGHOLD_ADDR}/v1/sys/loggers"
```

Повысить уровень для одной подсистемы (например, `core`):

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  --request POST \
  --data '{"level": "trace"}' \
  "${STRONGHOLD_ADDR}/v1/sys/loggers/core"
```

Вернуть уровень из конфигурации:

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  --request DELETE \
  "${STRONGHOLD_ADDR}/v1/sys/loggers"
```

Имена подсистем совпадают с именами логгеров в строках журнала (например, `core`, `expiration`, `raft`, `audit`). Актуальный список для узла возвращает `GET /v1/sys/loggers`.

### Потоковый просмотр журналов

Команда `d8 stronghold monitor` подключается к эндпоинту [`/sys/monitor`](../../../reference/api/system/#get-sysmonitor) и выводит журналы узла в реальном времени с заданным уровнем, не меняя конфигурацию:

```shell
d8 stronghold monitor -log-level=debug
```
