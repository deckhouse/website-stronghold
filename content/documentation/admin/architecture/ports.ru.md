---
title: "Сетевые порты"
description: "Какие сетевые соединения использует Stronghold: API, связь между узлами, межкластерная репликация, seal, методы аутентификации, механизмы секретов, аудит и резервное копирование."
weight: 30
---

В таблицах ниже перечислены соединения, которые нужно разрешить на межсетевых экранах. Порты Stronghold (`8200`, `8201`) — значения по умолчанию для standalone-установки, их можно изменить в параметрах `listener`, `api_addr` и `cluster_addr` [конфигурационного файла](../../../install/standalone/configuration/). Порты внешних систем зависят от их настроек; в таблицах указаны стандартные значения.

## Порты Stronghold

| Источник | Получатель | Порт/протокол | Назначение |
| --- | --- | --- | --- |
| Клиенты, CLI, веб-интерфейс, Stronghold Agent | Узлы Stronghold (listener) | `8200`/TCP, HTTPS | API и веб-интерфейс. Адрес задаётся в `listener "tcp"` и анонсируется через `api_addr`. |
| Узел Stronghold | Другие узлы кластера | `8201`/TCP, TLS | Кластерный порт (`cluster_addr`): консенсус и репликация Raft, пересылка запросов со standby на активный узел. TLS используется всегда. |
| Новый узел Stronghold | API узла-лидера | `8200`/TCP, HTTPS | Присоединение к кластеру Raft (`retry_join`, `leader_api_addr`). |
| Узел Stronghold | API других узлов | `8200`/TCP, HTTPS | Автоматическое распечатывание `seal "inner-cluster"`: распечатанный узел передаёт доли ключа узлам из блоков `node`. |
| Secondary-кластер (Stronghold EE) | Кластерный порт primary | `8201`/TCP, TLS | Нативная межкластерная репликация (PR) и репликация для аварийного восстановления (DR): поток WAL с primary на secondary и пересылка записей на primary. |
| Secondary-кластер (Stronghold EE) | API primary | `8200`/TCP, HTTPS | Разворачивание activation-токена при включении secondary (`primary_api_addr`). |
| Кластер-потребитель KV-репликации | API кластера-источника | `8200`/TCP, HTTPS | Репликация KV1/KV2 на уровне API. Доступ к кластерному порту не нужен. |

<!-- TODO(verify): для нативной репликации в Stronghold EE — порт и направление соединений (в upstream Vault secondary подключается к кластерному порту primary, по умолчанию 8201; нужен ли доступ primary к secondary). -->

Мониторинг и телеметрия настраиваются в секции `telemetry` конфигурационного файла: например, при `statsite_address = "127.0.0.1:8125"` Stronghold отправляет метрики на указанный адрес. Метрики также доступны через API на порту `8200`.

Метрики Prometheus отдаются эндпоинтом `/v1/sys/metrics?format=prometheus` на API-порту, а `statsite_address` использует TCP.

## Порты в DP

| Источник | Получатель | Порт/протокол | Назначение |
| --- | --- | --- | --- |
| Пользователи, CLI, приложения вне кластера | Ingress-контроллер DP | `443`/TCP, HTTPS | Веб-интерфейс и API по адресу `stronghold.<домен>` из `publicDomainTemplate`. |
| Браузер пользователя | Dex (модуль `user-authn`) | `443`/TCP, HTTPS | Вход в веб-интерфейс через OIDC. |
| Ingress-контроллер или Application Load Balancer (только инлеты `Ingress` и `GatewayAPI`) | Поды Stronghold | `8500`/TCP, HTTPS | HTTPS с обязательным клиентским сертификатом (`https-mtls`). По умолчанию Ingress и ALB отправляют трафик на порт `https-mtls`. |
| Клиенты, приложения (через Service `stronghold`, инлеты кроме Ingress) | Поды Stronghold | `8200`/TCP, HTTPS | Внешний API через Service. Сертификат берётся из секрета `ingress-tls`. |
| Под Stronghold | Другие поды Stronghold | `8201`/TCP, TLS | Кластерный порт. Публикуется наружу только при `publishCluster.enabled` (Stronghold EE). |
| Поды приложений, поды Stronghold, unsealer модуля | Поды Stronghold | `8300`/TCP, HTTPS | Внутренний API: `retry_join`, `seal "inner-cluster"`, пробы и распечатывание. |
| Под Stronghold | Другие поды Stronghold | `8301`/TCP, TLS | Внутренний кластерный адрес: Raft и пересылка запросов между узлами. |
| Prometheus | Поды Stronghold (`kube-rbac-proxy`) | `9889`/TCP, HTTPS | Метрики. |

Подробнее о размещении модуля — на странице [«Развёртывание в DP»](../deployment-dkp/).

## Seal и внешние хранилища ключей

| Источник | Получатель | Порт/протокол | Назначение |
| --- | --- | --- | --- |
| Узел Stronghold | Yandex Cloud KMS | `443`/TCP, HTTPS | Шифрование и расшифровка root-ключа для `seal "yandexcloudkms"`. При двойном шифровании KMS должен быть доступен постоянно. Эндпоинт можно переопределить параметром `endpoint`. |
| Узел Stronghold | Metadata service виртуальной машины Yandex Cloud | `80`/TCP, HTTP | Получение токена сервисного аккаунта ВМ, если не заданы `oauth_token` и `service_account_key_file`. |
| Узел Stronghold | HSM | Локально или порт производителя | `seal "pkcs11"` работает через локальную библиотеку PKCS#11. Для USB-токенов (Рутокен ЭЦП 3.0) и TPM2 сетевые порты не нужны. Для сетевого HSM откройте порт, указанный в документации производителя. |

<!-- TODO(verify): адрес и порт metadata service (в Yandex Cloud — 169.254.169.254:80). -->

## Методы аутентификации

| Источник | Получатель | Порт/протокол | Назначение |
| --- | --- | --- | --- |
| Узел Stronghold | Сервер LDAP / Active Directory | `389`/TCP (LDAP, StartTLS), `636`/TCP (LDAPS) | Метод аутентификации [LDAP](../../../user/auth/ldap/) и механизм секретов [LDAP](../../../user/secrets-engines/ldap/). |
| Узел Stronghold | Провайдер OIDC / JWT | `443`/TCP, HTTPS | Методы [OIDC](../../../user/auth/oidc/) и [JWT](../../../user/auth/jwt/): получение OIDC Discovery, JWKS, обмен кода на токен. |
| Узел Stronghold | API Kubernetes | `6443`/TCP или `443`/TCP, HTTPS | Метод [Kubernetes](../../../user/auth/kubernetes/): проверка токенов сервисных аккаунтов (TokenReview). Порт задаётся в `kubernetes_host`. |
| Браузер пользователя | Провайдер OIDC, SAML IdP | `443`/TCP, HTTPS | Перенаправление пользователя на страницу входа провайдера. |

## Механизмы секретов

| Источник | Получатель | Порт/протокол | Назначение |
| --- | --- | --- | --- |
| Узел Stronghold | PostgreSQL | `5432`/TCP | [Динамические учётные данные PostgreSQL](../../../user/secrets-engines/databases/postgresql/). |
| Узел Stronghold | MySQL / MariaDB | `3306`/TCP | [Динамические учётные данные MySQL/MariaDB](../../../user/secrets-engines/databases/mysql-maria/). |
| Узел Stronghold | ClickHouse | `9000`/TCP (нативный протокол) | [Динамические учётные данные ClickHouse](../../../user/secrets-engines/databases/clickhouse/). Адрес задаётся в `connection_url`. |
| Узел Stronghold | RabbitMQ management HTTP API | `15672`/TCP, HTTP(S) | Механизм секретов RabbitMQ (`connection_uri`). |
| Узел Stronghold | API Kubernetes | `6443`/TCP или `443`/TCP, HTTPS | [Механизм секретов Kubernetes](../../../user/secrets-engines/kubernetes/): создание сервисных аккаунтов и токенов. |

Плагин ClickHouse подключается по нативному протоколу (TCP, порт `9000`); HTTP-порт `8123` не используется.

Механизмы KV, PKI, Transit, SSH и TOTP не устанавливают исходящих соединений: они работают с данными в хранилище Stronghold.

## Аудит и резервное копирование

| Источник | Получатель | Порт/протокол | Назначение |
| --- | --- | --- | --- |
| Узел Stronghold | Локальная служба syslog | Локальный сокет | Аудит-устройство `syslog`. Пересылку на удалённый сервер настраивайте в самой службе syslog. |
| Узел Stronghold | Приёмник журналов | Порт из параметра `address`, TCP или UDP | Аудит-устройство `socket` (например, `address=127.0.0.1:9090 socket_type=tcp`). Для UNIX-сокета сетевой порт не нужен. |
| Активный узел Stronghold | S3-совместимое хранилище | `443`/TCP, HTTPS, или порт из `aws_s3_endpoint` | [Автоматические снимки](../../backups/automated-snapshots/) (Stronghold EE) с типом хранилища `aws-s3`. |

Аудит-устройство `file` и снимки с типом `local` пишут на локальный диск узла. Правила записи в аудит-устройства и поведение при их недоступности описаны в разделе [«Аудит»](../../audit/overview/#ошибки-аудит-устройств).
