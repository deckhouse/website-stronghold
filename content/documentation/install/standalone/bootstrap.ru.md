---
title: "Быстрый старт"
description: "Генерация артефактов развёртывания Stronghold командой stronghold bootstrap: systemd, Helm-чарт и Docker-образ."
weight: 10
---

Группа команд `stronghold bootstrap` генерирует артефакты развёртывания Stronghold. Доступна начиная с версии `v1.19`.

Подкоманды:

| Подкоманда | Назначение |
| --- | --- |
| `service` | Сгенерировать shell-скрипт подготовки одноузлового Linux-сервера (каталоги, конфигурация, TLS, systemd, пользователь, бинарный файл) |
| `helm` | Записать на диск архив Helm-чарта Stronghold, встроенный в бинарный файл |
| `docker` | Собрать OCI-совместимый tar-архив Docker-образа с бинарным файлом Stronghold |

```shell
stronghold bootstrap -help
```

## Подготовка Linux-сервера

Команда `stronghold bootstrap service` генерирует shell-скрипт, который подготавливает одноузловое окружение Stronghold: каталоги в `/opt/stronghold`, файл конфигурации с хранилищем Raft, при необходимости самоподписанные TLS-сертификаты, systemd-unit, системного пользователя и установку бинарного файла.

{{< alert level="warning" >}}
Самоподписанные TLS-сертификаты подходят только для тестирования.
Для production подготовьте собственные сертификаты и используйте параметр `-no-tls`.
{{< /alert >}}

По умолчанию скрипт выводится в stdout — его можно просмотреть перед применением:

```shell
stronghold bootstrap service | less
```

Примените скрипт:

```shell
stronghold bootstrap service | sudo sh
```

или выполните установку напрямую:

```shell
sudo stronghold bootstrap service -apply
```

После запуска сервиса инициализируйте и распечатайте Stronghold, как описано в разделе [Установка](../installation/#быстрая-установка).

Для развёртывания кластера в режиме HA базовое окружение на каждом узле можно подготовить с параметром `-no-tls`, а затем разместить собственные сертификаты и конфигурацию. См. [Развёртывание кластера в режиме HA](../installation/#развёртывание-кластера-в-режиме-ha).

### Параметры service

Полный список параметров доступен в справке:

```shell
stronghold bootstrap service -help
```

Часто используемые параметры:

| Параметр | Назначение |
| --- | --- |
| `-apply` | Выполнить сгенерированный скрипт вместо вывода в stdout |
| `-base-dir` | Базовый каталог установки (по умолчанию `/opt/stronghold`) |
| `-config` | Путь к файлу конфигурации Stronghold |
| `-data-dir` | Путь к каталогу данных Raft |
| `-tls-dir` | Путь к каталогу TLS-сертификатов |
| `-bin-path` | Путь назначения для бинарного файла Stronghold |
| `-source-binary` | Исходный путь бинарного файла Stronghold для установки (по умолчанию — текущий исполняемый файл) |
| `-systemd-unit` | Путь к файлу systemd-unit (по умолчанию `/etc/systemd/system/stronghold.service`) |
| `-user` | Системный пользователь для запуска сервиса (по умолчанию `stronghold`) |
| `-group` | Системная группа для запуска сервиса (по умолчанию `stronghold`) |
| `-storage` | Тип хранилища (поддерживается только `raft`) |
| `-node-id` | Идентификатор узла Raft (по умолчанию — имя хоста) |
| `-listener-address` | Адрес TCP-listener в формате `host:port` (по умолчанию `0.0.0.0:8200`) |
| `-api-addr` | API-адрес, который Stronghold сообщает клиентам (определяется автоматически, если не задан) |
| `-cluster-addr` | Адрес кластера для Raft (определяется автоматически, если не задан) |
| `-no-user-create` | Не создавать системного пользователя |
| `-no-systemd` | Не создавать и не включать systemd-unit |
| `-no-tls` | Не генерировать TLS-сертификаты |
| `-tls-disable` | Отключить TLS в конфигурации listener (только для разработки) |
| `-force` | Перезаписать существующие файлы вместо ошибки |

## Извлечение Helm-чарта

Команда `stronghold bootstrap helm` записывает на диск архив Helm-чарта Stronghold, встроенный в бинарный файл. Используйте архив с `helm install` или `helm template`.

Предварительные требования: Helm 3 и доступ к кластеру Kubernetes.

Запишите архив чарта в текущий каталог (имя файла по умолчанию соответствует версии встроенного чарта):

```shell
stronghold bootstrap helm
```

или укажите путь явно:

```shell
stronghold bootstrap helm -output stronghold.tgz
```

Сформируйте манифесты без установки:

```shell
helm template stronghold stronghold.tgz > stronghold.yaml
```

Установите чарт:

```shell
helm install stronghold stronghold.tgz
```

### Параметры helm

| Параметр | Назначение |
| --- | --- |
| `-output` | Путь к выходному архиву (по умолчанию — имя файла встроенного чарта) |

```shell
stronghold bootstrap helm -help
```

## Сборка Docker-образа

Команда `stronghold bootstrap docker` собирает минимальный OCI-совместимый tar-архив образа на диске. По умолчанию в образ встраивается текущий бинарный файл Stronghold. Импортируйте архив командой `docker load`.

Соберите tar-архив образа:

```shell
stronghold bootstrap docker
```

По умолчанию файл называется `stronghold-v<version>.tar` в текущем каталоге, а тег образа — `stronghold:<version>`.

Загрузите и запустите образ:

```shell
docker load -i stronghold-v1.19.0.tar
docker run --rm -p 8200:8200 stronghold:1.19.0
```

{{< alert level="info" >}}
Команда контейнера по умолчанию — `stronghold server -dev`.
Для запуска, близкого к production, передайте собственные аргументы сервера и смонтируйте файл конфигурации.
{{< /alert >}}

Для динамически слинкованного бинарного файла встройте дополнительные разделяемые библиотеки или плагины параметром `-extra-file` в формате `source=destination`:

```shell
stronghold bootstrap docker \
  -extra-file /lib64/ld-linux-x86-64.so.2=/lib64/ld-linux-x86-64.so.2 \
  -extra-file /lib/x86_64-linux-gnu/libc.so.6=/lib/x86_64-linux-gnu/libc.so.6 \
  -extra-file /opt/aktivco/rutokenecp/amd64/librtpkcs11ecp.so=/lib/librtpkcs11ecp.so
```

### Параметры docker

| Параметр | Назначение |
| --- | --- |
| `-output` | Путь к выходному tar-архиву |
| `-tag` | Тег Docker-образа в формате `repo:ref` |
| `-source-binary` | Исходный путь бинарного файла Stronghold для встраивания (по умолчанию — текущий исполняемый файл) |
| `-ca-certs` | Путь к бандлу CA-сертификатов для встраивания (по умолчанию `/etc/ssl/certs/ca-certificates.crt`) |
| `-extra-file` | Дополнительный файл для встраивания в формате `source=destination` (можно указать несколько раз) |

```shell
stronghold bootstrap docker -help
```
