---
title: "Механизм секретов trdl"
linkTitle: "trdl"
description: "Встроенный плагин trdl в Stronghold EE: безопасная сборка, подпись и публикация релизов из Git-репозитория с проверкой кворума подписей коммитов."
weight: 115
---

Механизм секретов `trdl` — встроенный плагин Stronghold EE для безопасной доставки релизов ПО по модели [trdl](https://trdl.dev/). Плагин собирает артефакты релиза из Git-тега, подписывает их и публикует в S3-хранилище, а также публикует каналы обновлений. Сборка и публикация выполняются только для коммитов, у которых есть требуемое число проверенных PGP-подписей (кворум).

Основные свойства:

- исходный код и конфигурация релиза (`trdl.yaml`) хранятся в Git-репозитории;
- конфигурация каналов обновлений (`trdl_channels.yaml`) может храниться в отдельной ветке;
- доверенные открытые PGP-ключи и требуемое число подписей задаются в конфигурации плагина;
- артефакты релиза подписываются PGP-ключом, которым управляет плагин;
- поддерживается подпись macOS-сборок и ELF-бинарников через Delivery Kit;
- сборки выполняются асинхронно в виде задач, статус и журнал которых доступны через API.

<!-- TODO(verify): уточнить связь с werf/trdl-клиентом, формат trdl.yaml и trdl_channels.yaml и способ хранения подписей коммитов; при необходимости дать ссылки на документацию trdl. -->

## Настройка

1. Включите механизм секретов:

   ```bash
   d8 stronghold secrets enable -path=trdl-myapp trdl
   ```

1. Настройте плагин: Git-репозиторий, S3-хранилище для публикации и требуемое число проверенных подписей коммита:

   ```bash
   d8 stronghold write trdl-myapp/configure \
     git_repo_url="https://git.example.com/org/myapp.git" \
     required_number_of_verified_signatures_on_commit=2 \
     s3_endpoint="https://s3.example.com" \
     s3_region="ru-central1" \
     s3_bucket_name="myapp-tuf" \
     s3_access_key_id="<идентификатор ключа>" \
     s3_secret_access_key="<секретный ключ>"
   ```

   Обязательные параметры:

   | Параметр | Описание |
   |----------|----------|
   | `git_repo_url` | URL Git-репозитория. |
   | `required_number_of_verified_signatures_on_commit` | Требуемое число проверенных подписей коммита. |
   | `s3_endpoint`, `s3_region`, `s3_bucket_name` | Эндпоинт, регион и имя бакета S3-хранилища. |
   | `s3_access_key_id`, `s3_secret_access_key` | Учётные данные доступа к S3-хранилищу. |

   Дополнительные параметры:

   | Параметр | Описание |
   |----------|----------|
   | `git_trdl_path` | Путь к файлу конфигурации релиза в репозитории. По умолчанию `trdl.yaml`. |
   | `git_trdl_channels_path` | Путь к файлу конфигурации каналов. По умолчанию `trdl_channels.yaml`. |
   | `git_trdl_channels_branch` | Отдельная ветка Git для файла конфигурации каналов. |
   | `initial_last_published_git_commit` | Начальный коммит последней успешной публикации. |
   | `buildx_driver`, `buildx_driver_opts` | Драйвер buildx для сборки (`docker-container` по умолчанию или `kubernetes`) и его опции. |
   | `buildkitd_address` | Адрес запущенного buildkitd (`unix://`, `tcp://`, `docker-container://` или `kube-pod://`). |
   | `buildkitd_driver`, `buildkitd_driver_opts` | Эфемерный buildkitd для каждой сборки (например, `kubernetes`) и его опции. |

   {{< alert level="warning" >}}
   Секреты сборки передаются демону buildkitd. Канал `tcp://` не шифруется и не аутентифицируется, поэтому защитите канал и изолируйте демон самостоятельно.
   {{< /alert >}}

1. Добавьте доверенные открытые PGP-ключи, подписи которых учитываются при проверке кворума:

   ```bash
   d8 stronghold write trdl-myapp/configure/trusted_pgp_public_key \
     name="developer-1" \
     public_key=@developer-1.asc
   ```

   Список добавленных ключей:

   ```bash
   d8 stronghold list trdl-myapp/configure/trusted_pgp_public_key
   ```

1. Если репозиторий закрытый, настройте учётные данные Git:

   ```bash
   d8 stronghold write trdl-myapp/configure/git_credential \
     username="trdl-bot" \
     password="<токен доступа>"
   ```

1. При необходимости добавьте секреты сборки, которые будут доступны при сборке артефактов:

   ```bash
   d8 stronghold write trdl-myapp/configure/build/secrets \
     id="npm-token" \
     data="<значение>"
   ```

1. При необходимости настройте менеджер задач:

   ```bash
   d8 stronghold write trdl-myapp/task/configure \
     task_timeout=30m \
     task_history_limit=10
   ```

## Подпись артефактов

- Открытую часть PGP-ключа, которым плагин подписывает артефакты релиза, можно прочитать по пути `configure/pgp_signing_key`:

  ```bash
  d8 stronghold read trdl-myapp/configure/pgp_signing_key
  ```

  <!-- TODO(verify): генерируется ли ключ подписи автоматически и что именно возвращает GET configure/pgp_signing_key. -->

- Для подписи macOS-сборок задайте сертификат и параметры Notary в `configure/build/mac_signing_identity` (`data`, `password`, `notary_issuer`, `notary_key`, `notary_key_id`).
- Для подписи ELF-бинарников через Delivery Kit задайте сертификат и закрытый ключ в `configure/delivery_kit_elf_signing`. Ключ можно передать в base64 или ссылкой `hashivault://<key>` на ключ механизма секретов Transit; в этом случае задайте параметры `vault_*`.

## Использование

1. Выполните релиз для подписанного Git-тега:

   ```bash
   d8 stronghold write trdl-myapp/release git_tag="v1.2.3"
   ```

   Плагин проверяет подписи, собирает и публикует артефакты. Операция выполняется асинхронно: в ответе возвращается идентификатор задачи.

   <!-- TODO(verify): формат ответа (поле с UUID задачи). -->

1. Опубликуйте каналы обновлений из файла `trdl_channels.yaml`:

   ```bash
   d8 stronghold write -force trdl-myapp/publish
   ```

1. Отслеживайте задачи:

   ```bash
   d8 stronghold read trdl-myapp/task
   d8 stronghold read trdl-myapp/task/<uuid>
   d8 stronghold read trdl-myapp/task/<uuid>/log
   ```

   Чтобы отменить выполняющуюся задачу:

   ```bash
   d8 stronghold write -force trdl-myapp/task/<uuid>/cancel
   ```

Последний опубликованный коммит хранится по пути `configure/last_published_git_commit`; его можно прочитать или удалить.

Полный список эндпоинтов и параметров приведён в [справочнике API механизмов секретов](../../../reference/api/secrets/).
