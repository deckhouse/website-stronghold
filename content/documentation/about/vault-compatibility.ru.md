---
title: "Совместимость с HashiCorp Vault"
linkTitle: "Совместимость с Vault"
description: "Совместимость Stronghold с API, CLI и инструментами экосистемы HashiCorp Vault, а также отличия Stronghold от upstream Vault."
weight: 40
---

Stronghold разработан на основе HashiCorp Vault: первая версия Stronghold (v1.0, февраль 2024) основана на Vault v1.14.x (см. [историю изменений](../../release-notes/)). Поэтому Stronghold сохраняет модель данных, HTTP API, формат конфигурации и CLI upstream Vault и дополняет их собственными возможностями.

Кодовая база upstream — Vault OSS 1.14.8 (последняя версия под лицензией MPL 2.0). Нумерация версий Stronghold (1.19) не совпадает с нумерацией Vault: возможности более новых версий Vault в Stronghold не переносятся автоматически.

Базовый Stronghold соответствует Vault CE. Часть возможностей, которые в upstream Vault доступны только в Vault Enterprise, в Stronghold входит в редакцию Stronghold EE (см. [«Редакции»](../editions/)).

## Совместимость API

HTTP API Stronghold совместим с API Vault:

- запросы выполняются к тем же путям с префиксом `/v1/` (`/v1/sys/...`, `/v1/auth/...`, `/v1/<mount>/...`);
- токен передаётся в заголовке `X-Vault-Token`;
- форматы запросов и ответов механизмов секретов и методов аутентификации совпадают с Vault;
- шифртекст механизма `Transit` имеет формат `vault:v<версия ключа>:...`, поэтому данные, зашифрованные в Vault, остаются в привычном формате;
- API хранилищ `KV1/KV2` совместим с Vault настолько, что Stronghold может [реплицировать KV-хранилища](../../admin/replication/kv-replication/) из Vault CE, Vault Enterprise или любого хранилища с совместимым Vault KV/KV2 API.

Пример запроса к API:

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  "${STRONGHOLD_ADDR}/v1/sys/health"
```

Полное описание эндпоинтов приведено в [справочнике API](../../reference/api/).

## CLI

CLI Stronghold повторяет команды и флаги CLI `vault`. Отличается только способ вызова:

| Окружение | Команда | Пример |
| --- | --- | --- |
| HashiCorp Vault | `vault <команда>` | `vault kv get -mount=secret app` |
| Stronghold в DP (через Deckhouse CLI) | `d8 stronghold <команда>` | `d8 stronghold kv get -mount=secret app` |
| Stronghold в Linux (standalone) | `stronghold <команда>` | `stronghold kv get -mount=secret app` |

Подкоманды, флаги и их поведение описаны в разделе [«d8 stronghold»](../../reference/console-tools/d8-stronghold/).

### Переменные окружения

Клиентские переменные окружения в документации Stronghold используют префикс `STRONGHOLD_`:

| Stronghold | Аналог в Vault | Назначение |
| --- | --- | --- |
| `STRONGHOLD_ADDR` | `VAULT_ADDR` | Адрес сервера |
| `STRONGHOLD_TOKEN` | `VAULT_TOKEN` | Токен клиента |
| `STRONGHOLD_CACERT` | `VAULT_CACERT` | Путь к CA-сертификату для проверки TLS-сертификата сервера |

CLI принимает и переменные с префиксом `VAULT_` (например, `VAULT_ADDR`, `VAULT_TOKEN`). Если заданы обе переменные, приоритет у `STRONGHOLD_*`.

Переменные окружения, которые переопределяют параметры конфигурации сервера, сохраняют префикс `VAULT_`: например, `VAULT_API_ADDR`, `VAULT_CLUSTER_ADDR`, `VAULT_RAFT_PATH`, `VAULT_RAFT_NODE_ID` и `VAULT_ENABLE_FILE_PERMISSIONS_CHECK` (см. [«Настройка»](../../install/standalone/configuration/)).

### Конфигурация сервера и агента

- Конфигурационный файл сервера использует формат HCL или JSON и те же секции, что и Vault: `listener`, `storage`, `seal`, `telemetry`, `ui` и другие (см. [«Настройка»](../../install/standalone/configuration/)).
- [Stronghold Agent](../../user/agent/overview/) использует секцию `stronghold` для подключения к серверу. Для подключения к HashiCorp Vault используйте секцию `vault` (см. [«Настройки агента»](../../user/agent/settings/)).

## Инструменты экосистемы

Поскольку API совместим, большинство инструментов экосистемы Vault можно направить на Stronghold, указав его адрес. Статус совместимости:

| Инструмент | Способ подключения | Статус |
| --- | --- | --- |
| Модуль DP [`secrets-store-integration`](/modules/secrets-store-integration/) (CSI-драйвер, env-injector) | Нативная интеграция | Поддерживается |
| [Stronghold Agent](../../user/agent/overview/) | Нативная интеграция | Поддерживается |
| Ansible, Terraform | Провайдеры и модули для Vault | Поддерживается (см. [«Редакции»](../editions/)) <!-- TODO(verify): протестированные версии провайдера hashicorp/vault для Terraform и коллекции community.hashi_vault для Ansible --> |
| [External Secrets Operator](../../examples/delivery/kubernetes-workloads/#external-secrets-operator) | Провайдер `vault` | Не подтверждено тестами <!-- TODO(verify): протестированная версия ESO и API external-secrets.io --> |
| cert-manager | Issuer типа `vault` | Не подтверждено тестами <!-- TODO(verify): совместимость Vault Issuer cert-manager с механизмом PKI Stronghold --> |
| Jenkins (HashiCorp Vault Plugin), GitLab CI, GitHub Actions | Плагины и действия для Vault, JWT-аутентификация | См. [«Интеграции»](../../examples/) <!-- TODO(verify): статус тестирования для каждого инструмента --> |
| Клиентские SDK (Go `github.com/hashicorp/vault/api`, Python `hvac` и другие) | Клиентские библиотеки Vault | Не подтверждено тестами <!-- TODO(verify): список протестированных SDK и версий --> |

Способы доставки секретов в Kubernetes сравниваются на странице [«Доставка секретов в поды Kubernetes»](../../examples/delivery/kubernetes-workloads/).

## Возможности, которых нет в upstream Vault

| Возможность | Редакция | Описание |
| --- | --- | --- |
| [Репликация KV1/KV2](../../admin/replication/kv-replication/) | Stronghold EE | Pull-репликация отдельных KV-хранилищ по API, в том числе из Vault CE и Vault Enterprise |
| [Механизм секретов GitOps](../../user/secrets-engines/gitops/overview/) | Stronghold EE | Управление конфигурацией Stronghold через кворум подписей Git-коммитов |
| [Встроенный плагин `trdl`](../../user/secrets-engines/trdl/) | Stronghold EE | Управление сборкой и подписью артефактов через кворум подписей Git-коммитов |
| [Seal `inner-cluster`](../../concepts/seal/) | Stronghold EE, Stronghold CSE | Встроенное автоматическое распечатывание: распечатанные узлы кластера распечатывают другие узлы без внешнего KMS |
| [Seal `yandexcloudkms`](../../admin/kms-hsm/yandexcloudkms/) | Stronghold EE | Автоматическое распечатывание и защита root-ключа с помощью Yandex Cloud KMS (только standalone) |
| [ГОСТ-криптография](../../admin/cryptography/overview/) | Stronghold, Stronghold EE | `TLS 1.3` с ГОСТ-шифрованием `Магма` и `Кузнечик`, сертификаты `ГОСТ 34.10` в `PKI`, ГОСТ-алгоритмы в `Transit` |
| `CryptoPro seal wrapper` | Сборка с поддержкой HSM (Linux) | Seal wrapper для сценариев с российской криптографией |
| [Managed Keys](../../user/managed-keys/overview/) с Yandex KMS | Stronghold EE | Работа с ключевым материалом в Yandex KMS и PKCS #11-устройствах из `Transit`, `PKI` и `SSH` |
| [Блокировка пространств имён](../../admin/namespaces/overview/) через веб-интерфейс | Stronghold EE | Блокировка и разблокировка Namespace API |
| [Команды `stronghold bootstrap`](../../install/standalone/bootstrap/) | Stronghold EE | Генерация скрипта установки сервиса Linux, Helm-чарта и Docker-образа |
| Веб-интерфейс на русском языке, управление ролями и политиками в веб-интерфейсе | Управление ролями — Stronghold EE | Локализованный веб-интерфейс с расширенными возможностями администрирования |
| Развёртывание модулем DP | Все редакции | Автоматическая инициализация, распечатывание и интеграция с Dex (см. [«Настройка Stronghold»](../../install/dkp/configuration/)) |

Возможности Vault Enterprise, которые доступны в Stronghold EE: [пространства имён](../../admin/namespaces/overview/), [межкластерная репликация (PR) и репликация для аварийного восстановления (DR)](../../admin/replication/overview/), [performance standby](../../admin/replication/performance-standby/), [автоматические снимки](../../admin/backups/automated-snapshots/), [seal wrap](../../admin/kms-hsm/sealwrap/), [HSM (PKCS #11)](../../admin/kms-hsm/hsm/), Managed Keys и [аутентификация SAML](../../user/auth/saml/). При миграции с Vault Enterprise конфигурация автоматических снимков сохраняется.

## Отличия и ограничения

- Из внешних облачных KMS для `seal` поддерживается только Yandex Cloud KMS. Конфигурации `awskms`, `gcpckms`, `azurekeyvault`, `ocikms` и `alicloudkms` не поддерживаются; `transit` seal поддерживается.
- Работа с HSM (`seal "pkcs11"`) и `seal "yandexcloudkms"` в текущей версии поддерживается только при standalone-установке.
- В DP модуль `stronghold` поддерживает режимы `Automatic` (по умолчанию) и `Manual` и инлеты `Ingress` (по умолчанию), `GatewayAPI`, `LoadBalancer`, `NodePort` и `None`; в режиме `Automatic` инициализация и распечатывание выполняются модулем (см. [«Настройка Stronghold»](../../install/dkp/configuration/)).
- В Stronghold `seal wrap` можно отключить для всех данных, кроме root key, параметром `disable_sealwrap = true`.
- <!-- TODO(verify): статус в Stronghold функций Vault Enterprise, не упомянутых в документации: Sentinel, Control Groups, KMIP, Transform, Key Management secrets engine, MFA на уровне логина (Login MFA) в upstream-реализации. -->

## Миграция с HashiCorp Vault

Порядок переноса данных и конфигурации из HashiCorp Vault в Stronghold описан в руководстве [«Миграция с HashiCorp Vault»](../../examples/operations/migration-from-vault/).

## Примеры использования

Готовые примеры с этим механизмом:

- [Миграция с HashiCorp Vault](../../examples/operations/migration-from-vault/)

Все примеры собраны в разделе [«Примеры использования»](../../examples/).
