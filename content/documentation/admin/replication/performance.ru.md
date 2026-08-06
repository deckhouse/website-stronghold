---
title: "Репликация Performance"
linkTitle: "Репликация Performance"
weight: 20
description: "Настройка репликации Performance между кластерами Stronghold, её проверка и управление фильтрами путей."
---

Репликация Performance копирует хранилище primary-кластера (кроме локальных
путей) на один или несколько performance-secondary. Secondary обслуживает чтения
локально, а записи перенаправляет на primary. Так нагрузку чтения можно
масштабировать и размещать ближе к потребителям.

## Перед началом

- Убедитесь, что оба кластера работают на Stronghold EE с integrated Raft
  storage и что репликация включена (смотрите [Обзор](../overview/)).
- Убедитесь, что кластерный порт primary доступен с secondary и что у вас есть
  CA-сертификат primary для TLS.
- Подготовьте токен с правами на `sys/replication/*` на обоих кластерах.

В примерах ниже `${PRIMARY_ADDR}` и `${SECONDARY_ADDR}` — API-адреса кластеров,
а `${VAULT_TOKEN}` — токен с правами на `sys/replication/*`.

{{< alert level="info" >}}
Если репликацию включают на кластере, где уже есть данные (например, после
обновления со сборки без репликации), или после периода с отключённой
репликацией, узел при первом запуске сам заново индексирует хранилище. Вызывать
`sys/replication/reindex` вручную не нужно.
{{< /alert >}}

## Шаг 1. Включите primary

```shell
d8 stronghold write -force sys/replication/performance/primary/enable
```

{{< alert level="warning" >}}
Включение primary перезапускает ядро, и на несколько секунд узел становится
недоступен. Прежде чем продолжать, дождитесь, пока он снова начнёт отвечать и
`sys/replication/performance/status` покажет `mode: primary`.
{{< /alert >}}

## Шаг 2. Создайте activation-токен для secondary

Сгенерируйте wrapping-токен активации для конкретного secondary, заданного через
`id`:

```shell
d8 stronghold write sys/replication/performance/primary/secondary-token \
  id=sec-1 ttl=24h
```

Параметр `id` обязателен, `ttl` по умолчанию — `24h`. Команда возвращает
одноразовый wrapping-токен в поле `wrap_info.token` — его и передайте на
secondary.

## Шаг 3. Включите secondary

Передайте wrapping-токен из предыдущего шага. Для окружений с самоподписанными
сертификатами параметр `ca_cert` (CA primary в формате PEM) обязателен, иначе
TLS-соединение с primary не пройдёт проверку.

```shell
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  --request POST \
  --data @secondary-enable.json \
  "${SECONDARY_ADDR}/v1/sys/replication/performance/secondary/enable"
```

Пример `secondary-enable.json`:

```json
{
  "token": "<wrapping_token_from_step_2>",
  "primary_api_addr": "<primary_api_address>",
  "ca_cert": "<primary_ca_certificate_in_pem>"
}
```

Поля `secondary/enable`:

- `token` — обязательный; wrapping-токен из шага 2.
- `primary_api_addr` — адрес primary для разворачивания токена; обязателен, если
  в токене нет claim `addr`.
- `ca_cert` — inline PEM CA primary для TLS при unwrap.
- `ca_file` / `ca_path` — путь к PEM-файлу или директории PEM-файлов, читается
  на сервере.
- `client_cert_pem` / `client_key_pem` — опциональные клиентский сертификат и
  ключ для TLS-соединения при unwrap.

{{< alert level="warning" >}}
После bootstrap secondary обнуляет собственный token store, поэтому исходный
root-токен secondary перестаёт работать. Выполняйте вход через
**реплицированный** метод аутентификации (например, `userpass`), который был
настроен на primary **до** включения репликации.
{{< /alert >}}

## Шаг 4. Проверьте статус

```shell
d8 stronghold read -address="${SECONDARY_ADDR}" sys/replication/performance/status
```

Когда secondary подключён и тянет WAL, поле `state` равно `stream-wals`, а
`connection_state` — `ready`. Secondary догнал primary, когда его `last_wal`
дошёл до `last_remote_wal`.

## Шаг 5. Проверьте репликацию данных

Команды к secondary выполняйте с токеном, полученным при входе через
реплицированный метод аутентификации (шаг 3): root-токен primary на secondary
недействителен.

```shell
# Запись на primary.
d8 stronghold kv put -address="${PRIMARY_ADDR}" secret/hello value=world

# Чтение на secondary после схождения.
d8 stronghold kv get -address="${SECONDARY_ADDR}" secret/hello

# Запись на secondary перенаправляется на primary и возвращается через WAL.
d8 stronghold kv put -address="${SECONDARY_ADDR}" secret/fromsec value=1
```

## Фильтры путей

Primary может выборочно не отдавать конкретные хранилища или пространства имён
конкретному secondary. Фильтры настраиваются для каждого secondary по `id`:

- `deny` — реплицировать всё, **кроме** перечисленных путей.
- `allow` — реплицировать **только** перечисленные пути.

Режим по умолчанию — `deny`.

```shell
# Запретить хранилище blocked/ для sec-1.
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  --request POST \
  --data '{"mode":"deny","paths":["blocked/"]}' \
  "${PRIMARY_ADDR}/v1/sys/replication/performance/primary/paths-filter/sec-1"

# Прочитать фильтр.
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  "${PRIMARY_ADDR}/v1/sys/replication/performance/primary/paths-filter/sec-1"

# Прочитать развёрнутый фильтр (реальные storage-префиксы, которые будут применены).
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  "${PRIMARY_ADDR}/v1/sys/replication/performance/primary/paths-filter/sec-1/dynamic"

# Снять фильтр.
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  --request DELETE \
  "${PRIMARY_ADDR}/v1/sys/replication/performance/primary/paths-filter/sec-1"
```

Изменение фильтра применяется вживую: при добавлении пути в `deny` (или удалении
из `allow`) primary стримит инвалидацию и secondary удаляет ранее разрешённые
данные; при обратном изменении данные реплицируются заново.

{{< alert level="warning" >}}
Performance-secondary нельзя повысить (promote), пока на нём настроены фильтры
путей. Сначала снимите фильтры.
{{< /alert >}}

Когда движок секретов или аутентификации изменят или
отключают, его префикс автоматически убирается из всех фильтров. Так
устаревшая запись фильтра не влияет на новый движок, смонтированный по тому же
пути.

## Операции управления

| Действие | Эндпоинт |
| --- | --- |
| Статус | `d8 stronghold read sys/replication/performance/status` |
| Отозвать secondary | `d8 stronghold write sys/replication/performance/primary/revoke-secondary id=sec-1` |
| Отключить на primary | `d8 stronghold write -force sys/replication/performance/primary/disable` |
| Отключить на secondary | `d8 stronghold write -force sys/replication/performance/secondary/disable` |
| Понизить primary (demote) | `d8 stronghold write -force sys/replication/performance/primary/demote` |
| Повысить secondary (promote) | `d8 stronghold write -force sys/replication/performance/secondary/promote` |
| Перенаправить secondary | `d8 stronghold write sys/replication/performance/secondary/update-primary token=<activation_token>` |

Примечания:

- `revoke-secondary` также удаляет фильтр этого secondary.
- Для `update-primary`, когда у нового primary **новый** идентификатор кластера
  (например, это повышенный бывший secondary), используйте token-метод со свежим
  activation-токеном нового primary. Address-метод (`primary_cluster_addr`)
  применим только если primary сохранил прежний идентификатор.
