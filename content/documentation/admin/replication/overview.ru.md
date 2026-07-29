---
title: "Обзор"
linkTitle: "Обзор"
weight: 10
description: "Обзор нативной репликации Performance и Disaster Recovery между кластерами Stronghold."
---

Нативная репликация переносит данные одного кластера Stronghold на другие
кластеры по сети. В отличие от
[репликации KV1/KV2](../../kv-replication/overview/), которая копирует отдельные
секреты между хранилищами, нативная репликация переносит состояние кластера
целиком: подключённые хранилища, политики, identity и остальные данные.

Поддерживаются два независимых режима.

| Режим | Назначение | Что переносит | Обслуживает клиентов |
| --- | --- | --- | --- |
| **Performance** | Масштабирование и распределение чтений | Общие данные кластера (хранилища, политики, identity), кроме локальных для узла | Да: читает локально, записи перенаправляет на primary, ведёт собственный token store |
| **Disaster Recovery (DR)** | Горячий резерв и аварийное переключение | Все данные кластера, включая локальные (токены, аренды) | Нет: ждёт promote |

Отдельно есть **performance standby** — это не другой кластер, а неактивные
HA-узлы внутри одного кластера, которые обслуживают чтения локально. Смотрите
[Performance standby](../performance-standby/).

## Требования

- **Enterprise Edition.** Нативная репликация есть только в Stronghold EE.
- **Integrated Raft storage.** Репликация ведёт журнал изменений, который
  пишется только на integrated Raft storage. На других бэкендах она не
  запускается, а её эндпоинты возвращают
  `replication not supported by this physical backend`.
- **Доступ к кластерному порту.** Secondary подключается к primary по
  кластерному порту (`cluster_address`, TLS с ALPN), а не по API-порту. Этот
  порт должен быть доступен между кластерами, а их TLS-сертификаты —
  согласованы.

{{< alert level="warning" >}}
Нативная репликация Performance и DR доступна только в Standalone-установке
Stronghold. В составе модуля DKP межкластерная репликация не поддерживается.
{{< /alert >}}

## Включение и отключение

В Stronghold EE репликация включена по умолчанию: эндпоинты `sys/replication/*`
доступны сразу. Включение конкретного режима — отдельная операция, она описана
на страницах Performance и DR.

Чтобы полностью отключить репликацию, задайте в конфигурации сервера параметр:

```hcl
disable_wal_replication = true
```

Тогда узел загружается как обычный: репликация, performance standby и эндпоинты
`sys/replication/*` недоступны. Чтобы отключить только performance standby, не
трогая репликацию, используйте `disable_performance_standby = true`.

{{< alert level="info" >}}
Параметры `disable_wal_replication` и `disable_performance_standby` задаются в
конфигурации сервера Standalone-установки Stronghold.
{{< /alert >}}

Если репликация отключена, `GET sys/replication/status` возвращает `404`.

## Проверка состояния

```shell
d8 stronghold read sys/replication/status
```

Статус режима показывает `mode`, `cluster_id`, `state` (`running`,
`stream-wals`, `idle`), `connection_state`, `last_wal`, `last_remote_wal` и
списки известных primary и secondary. Secondary догнал primary, когда его
`last_wal` дошёл до `last_remote_wal`.

{{< alert level="info" >}}
Если репликацию включают на кластере, где уже есть данные (например, после
обновления со сборки без репликации), или после периода с отключённой
репликацией, узел при первом запуске сам заново индексирует хранилище. Вызывать
`sys/replication/reindex` вручную не нужно.
{{< /alert >}}

## Доступные страницы

- [Топологии и сценарии](../topologies/) — схемы и типовые кейсы для всех трёх
  режимов.
- [Репликация Performance](../performance/) — масштабирование чтений, настройка
  secondary, фильтры путей.
- [Disaster recovery](../disaster-recovery/) — горячий резерв, аварийное
  переключение и церемония promote.
- [Performance standby](../performance-standby/) — чтения на неактивных HA-узлах
  внутри кластера.
