---
title: "Автоматические снимки"
weight: 40
description: "Настройка автоматического резервного копирования встроенного хранилища Stronghold."
---

Автоматические снимки позволяют Stronghold по расписанию создавать резервные копии встроенного Raft-хранилища и сохранять их на локальный диск или в S3-совместимое объектное хранилище.

{{< alert level="warning" >}}
Автоматические снимки доступны только при использовании интегрированного Raft-хранилища. Для `etcd`, `postgresql` и других внешних backend-ов нужно настраивать резервное копирование средствами самого хранилища.
{{< /alert >}}

## Когда использовать

Автоматические снимки подходят для регулярного резервного копирования без ручного запуска команд. Они особенно полезны для production-кластеров, где важно иметь повторяемую политику хранения и выносить резервные копии за пределы самого кластера.

## Как это работает

- Можно создать несколько именованных конфигураций снимков.
- Каждая конфигурация определяет интервал запуска, политику хранения и тип хранилища.
- Поддерживаются типы хранилища `local` и `aws-s3`.
- Для production-сред локальное хранилище обычно менее предпочтительно, чем внешнее объектное хранилище: активный узел кластера может меняться, а резервные копии лучше хранить отдельно от защищаемой системы.

## Создание или обновление конфигурации

| Метод | Путь |
|-------|------|
| POST  | `/sys/storage/raft/snapshot-auto/config/:name` |

Для работы с endpoint требуются права `sudo`.

### Основные параметры

<div class="table__styling--container"></div>

| Параметр | Тип | Обязательный | Значение по умолчанию | Описание |
|----------|-----|--------------|-----------------------|----------|
| `name` | Строка | Да | — | Имя конфигурации, которую нужно создать или обновить. |
| `interval` | Целое число или строка | Да | — | Интервал между созданием резервных копий. Можно указывать в секундах или в формате Go duration, например `24h`. |
| `retain` | Целое число | Нет | `3` | Сколько резервных копий хранить. При превышении лимита самые старые копии удаляются. |
| `storage_type` | Неизменяемая строка | Да | — | Тип хранилища: `local` или `aws-s3`. |
| `path_prefix` | Неизменяемая строка | Да | — | Для `local` это директория хранения снимков. Для `aws-s3` это префикс объекта в бакете. |
| `file_prefix` | Неизменяемая строка | Нет | `stronghold-snapshot` | Префикс имени файла или объекта. |

### Дополнительные параметры для `local`

<div class="table__styling--container"></div>

| Параметр | Тип | Обязательный | Значение по умолчанию | Описание |
|----------|-----|--------------|-----------------------|----------|
| `local_max_space` | Целое число | Нет | `0` | Максимальный объём в байтах, который можно занять резервными копиями с указанным `file_prefix` в директории `path_prefix`. Значение `0` отключает проверку. |

### Дополнительные параметры для `aws-s3`

<div class="table__styling--container"></div>

| Параметр | Тип | Обязательный | Значение по умолчанию | Описание |
|----------|-----|--------------|-----------------------|----------|
| `aws_s3_bucket` | Строка | Да | — | Имя бакета для хранения резервных копий. |
| `aws_s3_region` | Строка | Нет | — | Регион бакета. |
| `aws_access_key_id` | Строка | Нет | — | Идентификатор ключа доступа к бакету. |
| `aws_secret_access_key` | Строка | Нет | — | Секретный ключ доступа к бакету. |
| `aws_s3_endpoint` | Строка | Нет | — | Endpoint S3-сервиса. |
| `aws_s3_disable_tls` | Булевый | Нет | — | Отключает TLS для S3-endpoint. Используйте только для тестирования. |
| `aws_s3_ca_certificate` | Строка | Нет | — | CA-сертификат для S3-endpoint в PEM-формате. |

## Примеры конфигурации

### Локальный диск

Следующий файл `local-snapshot.json` создаёт конфигурацию, которая сохраняет снимок каждые 5 минут в каталог `/stronghold/data/backups`, хранит 4 копии и использует префикс `main_stronghold`:

```json
{
  "interval": "5m",
  "path_prefix": "/stronghold/data/backups",
  "file_prefix": "main_stronghold",
  "retain": "4",
  "storage_type": "local"
}
```

```shell
d8 stronghold write sys/storage/raft/snapshot-auto/config/my-local-snapshots @local-snapshot.json
```

{{< alert level="info" >}}
Перед применением конфигурации убедитесь, что каталог из `path_prefix` существует и доступен для записи. Ошибка `failed to create snapshot directory at destination` обычно означает, что каталог отсутствует или недоступен.
{{< /alert >}}

### S3-совместимое хранилище

Следующий файл `minio-snapshot.json` сохраняет снимки в S3-совместимое хранилище:

```json
{
  "interval": "3m",
  "path_prefix": "snapshots",
  "file_prefix": "stronghold_backup",
  "retain": "15",
  "storage_type": "aws-s3",
  "aws_s3_bucket": "my_bucket",
  "aws_s3_endpoint": "minio.domain.ru",
  "aws_access_key_id": "<ACCESS_KEY>",
  "aws_secret_access_key": "<SECRET_ACCESS_KEY>"
}
```

```shell
d8 stronghold write sys/storage/raft/snapshot-auto/config/my-remote-snapshots @minio-snapshot.json
```

{{< alert level="info" >}}
Перед применением конфигурации убедитесь, что бакет уже существует, а указанные учётные данные имеют права на чтение и запись.
{{< /alert >}}

### Обновление существующей конфигурации

Чтобы изменить только часть параметров, достаточно передать неполный JSON:

```json
{
  "interval": "3m",
  "retain": "10"
}
```

```shell
d8 stronghold write sys/storage/raft/snapshot-auto/config/my-local-snapshots @local-snapshot-update.json
```

## Просмотр списка конфигураций

| Метод | Путь |
|-------|------|
| LIST  | `/sys/storage/raft/snapshot-auto/config` |

```shell
d8 stronghold list sys/storage/raft/snapshot-auto/config
```

## Получение параметров конфигурации

| Метод | Путь |
|-------|------|
| GET   | `/sys/storage/raft/snapshot-auto/config/:name` |

```shell
d8 stronghold read sys/storage/raft/snapshot-auto/config/my-remote-snapshots
```

Для `aws-s3` значения `aws_access_key_id` и `aws_secret_access_key` в ответе не отображаются.

## Удаление конфигурации

| Метод | Путь |
|-------|------|
| DELETE | `/sys/storage/raft/snapshot-auto/config/:name` |

```shell
d8 stronghold delete sys/storage/raft/snapshot-auto/config/my-remote-snapshots
```

{{< alert level="info" >}}
Удаление конфигурации автоматического резервного копирования не удаляет уже созданные файлы снимков из локального или объектного хранилища.
{{< /alert >}}

## Получение статуса

| Метод | Путь |
|-------|------|
| GET   | `/sys/storage/raft/snapshot-auto/status/:name` |

```shell
d8 stronghold read sys/storage/raft/snapshot-auto/status/my-remote-snapshots
```

Ключевые поля статуса:

- `consecutive_errors` — количество подряд идущих ошибок резервного копирования;
- `last_snapshot_end` — время завершения последнего успешного снимка;
- `last_snapshot_error` — текст последней ошибки;
- `last_snapshot_start` — время начала последнего завершённого резервного копирования;
- `last_snapshot_url` — адрес последнего успешно созданного снимка;
- `next_snapshot_start` — время следующего запуска;
- `snapshot_start` — время начала текущего резервного копирования;
- `snapshot_url` — адрес текущего создаваемого снимка.

## См. также

- [Создание снимка](./save/)
- [Восстановление из снимка](./restore/)
