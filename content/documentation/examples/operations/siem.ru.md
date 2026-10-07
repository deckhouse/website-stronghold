---
title: "Отправка журналов аудита в SIEM"
linkTitle: "SIEM"
description: "Передача журналов аудита Stronghold в SIEM-системы: аудит-устройства, сбор через log-shipper, Vector и Fluent Bit, сопоставление полей для ELK, OpenSearch, Splunk, MaxPatrol SIEM и KUMA."
weight: 70
---

Журналы аудита Stronghold — основной источник событий безопасности: кто, когда, с какого адреса и к какому пути обращался и с каким результатом. Для централизованного анализа передавайте их в SIEM-систему.

{{< alert level="info" >}}
Команда `audit enable` и аудит-устройства доступны только в Stronghold EE.
{{< /alert >}}

Общая схема:

1. Включите одно или несколько [аудит-устройств](../../../admin/audit/overview/) (`file`, `syslog`, `socket`).
1. Соберите записи агентом доставки логов (`log-shipper` в Deckhouse Platform, Vector, Fluent Bit, rsyslog).
1. Разберите JSON-записи и сопоставьте поля со схемой SIEM-системы.

## Выбор аудит-устройства

| Устройство | Способ сбора | Особенности |
| --- | --- | --- |
| `file` | Агент читает файл и отправляет записи в SIEM | Наиболее предсказуемый вариант. Нужна ротация файла и контроль свободного места |
| `syslog` | Локальный syslog-агент пересылает записи в SIEM | Записи бывают большими: используйте надёжный транспорт и дублируйте устройством `file` |
| `socket` | Stronghold отправляет записи напрямую в TCP-, UDP- или UNIX-сокет коллектора | При недоступности коллектора запросы к Stronghold могут блокироваться, по UDP возможна потеря записей |

Включайте как минимум два аудит-устройства: если запись в аудит не удаётся, Stronghold может перестать обслуживать запросы. Подробнее — в разделе [«Аудит в Stronghold»](../../../admin/audit/overview/).

Пример: файл для доставки в SIEM и syslog как резервный канал:

```bash
d8 stronghold audit enable file file_path=/var/log/stronghold_audit.log
d8 stronghold audit enable syslog tag="stronghold" facility="AUTH"
```

## Сбор в Deckhouse Platform

В DP для доставки логов используется модуль [`log-shipper`](/products/kubernetes-platform/documentation/v1/modules/log-shipper/). Он читает логи подов и отправляет их в Elasticsearch, OpenSearch, Splunk, Logstash, Kafka, Loki, Vector или по протоколу syslog.

Чтобы `log-shipper` мог собрать аудит как логи подов, записи должны выводиться в стандартный поток вывода. В DP (Stronghold EE, `management.mode: Automatic`) аудит включается параметром `enableAuditLog` в ModuleConfig модуля `stronghold`:

```yaml
spec:
  settings:
    enableAuditLog: true
```

Модуль сам включает аудит-устройство `stronghold_audit_stdout` (тип `file`, `file_path=stdout`). В EE, если параметр не задан, модуль отключает аудит. Параметр нельзя задать при `management.mode: Manual`; в этом случае, а также в standalone-установке, включите устройство вручную:

```bash
d8 stronghold audit enable -path=file-stdout file file_path=stdout
```

Пример конфигурации `log-shipper`, которая отправляет логи подов Stronghold в Elasticsearch:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ClusterLoggingConfig
metadata:
  name: stronghold-audit
spec:
  type: KubernetesPods
  kubernetesPods:
    namespaceSelector:
      labelSelector:
        matchLabels:
          kubernetes.io/metadata.name: d8-stronghold
  destinationRefs:
    - siem-elasticsearch
---
apiVersion: deckhouse.io/v1alpha1
kind: ClusterLogDestination
metadata:
  name: siem-elasticsearch
spec:
  type: Elasticsearch
  elasticsearch:
    endpoint: https://elasticsearch.example.com:9200
    index: stronghold-audit-%F
    auth:
      strategy: Basic
      user: stronghold
      password: <пароль в base64>
```

В логах подов вместе с аудитом присутствуют и операционные журналы Stronghold. Отделяйте записи аудита по наличию полей `type` (`request` или `response`) и `request.id`.

## Сбор вне Kubernetes

Для Stronghold на Linux используйте любой агент, который умеет читать файл с JSON-строками. Пример конфигурации Vector:

```yaml
sources:
  stronghold_audit:
    type: file
    include:
      - /var/log/stronghold_audit.log

transforms:
  parse:
    type: remap
    inputs: [stronghold_audit]
    source: |
      . = parse_json!(.message)

sinks:
  siem:
    type: elasticsearch
    inputs: [parse]
    endpoints: ["https://elasticsearch.example.com:9200"]
    bulk:
      index: "stronghold-audit-%Y.%m.%d"
```

Пример для Fluent Bit:

```ini
[INPUT]
    Name   tail
    Path   /var/log/stronghold_audit.log
    Parser json
    Tag    stronghold.audit

[OUTPUT]
    Name   forward
    Match  stronghold.audit
    Host   siem-collector.example.com
    Port   24224
```

Настройте ротацию файла аудита (например, через `logrotate`) и после ротации отправьте Stronghold сигнал `SIGHUP`, чтобы он переоткрыл файл.

## Подключение к SIEM-системам

| SIEM-система | Рекомендуемый способ приёма |
| --- | --- |
| ELK / OpenSearch | Прямая запись в индекс через `log-shipper`, Vector или Logstash; разбор JSON на стороне агента |
| Splunk | HTTP Event Collector (HEC) с `sourcetype` типа `_json` |
| MaxPatrol SIEM | Приём JSON-событий по syslog или из файла через агент сбора; требуется правило нормализации для формата Stronghold <!-- TODO(verify): наличие готового пакета нормализации для Stronghold/Vault в MaxPatrol SIEM --> |
| KUMA | Коллектор с JSON-нормализатором, приём по TCP/syslog <!-- TODO(verify): наличие готового нормализатора для Stronghold/Vault в KUMA --> |

## Сопоставление полей

Структура записей описана в разделе [«Схема записей журнала аудита»](../../../admin/audit/log-format/). Типовое сопоставление с полями SIEM:

| Поле Stronghold | Значение в SIEM |
| --- | --- |
| `time` | Время события |
| `type` | Тип события: `request` или `response` |
| `request.id` | Идентификатор для связи запроса и ответа |
| `request.operation` | Действие: `create`, `read`, `update`, `delete`, `list` |
| `request.path` | Объект доступа |
| `request.mount_type`, `request.mount_point` | Тип и путь механизма секретов или метода аутентификации |
| `request.namespace.path` | Пространство имён Stronghold |
| `request.remote_address` | IP-адрес источника |
| `auth.display_name` | Имя субъекта |
| `auth.entity_id` | Идентификатор сущности |
| `auth.policies`, `auth.token_policies` | Применённые политики |
| `error` | Результат: при наличии поля — ошибка или отказ в доступе |

Практические замечания:

- Чувствительные строки (токены, значения секретов) в записях хешируются HMAC-SHA256 с солью аудит-устройства. Для сравнения значения с хешем используйте механизм audit hash того же устройства.
- Для корреляции используйте записи `response`: в них есть и данные запроса, и результат.
- Для снижения объёма используйте [фильтрацию](../../../admin/audit/filtering/) и [исключение полей](../../../admin/audit/exclusion/) на отдельном аудит-устройстве, направленном в SIEM. Сохраните полный журнал на другом устройстве.
- Примеры событий для правил корреляции: множественные ошибки аутентификации (`auth/*/login` с `error`), операции с `sys/audit`, `sys/policies`, `sys/auth`, использование root-токена (`auth.policies` содержит `root`).
