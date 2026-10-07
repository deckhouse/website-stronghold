---
title: "Руководство администратора Deckhouse Stronghold"
linkTitle: "Обзор"
weight: 10
---

Данный раздел предназначен для администраторов Deckhouse Stronghold.

В руководстве администратора представлены следующие разделы:

- Эксплуатация
  - [«Мониторинг»](./operations/monitoring/), [«Журналы сервера»](./operations/logs/), [«Устранение неполадок»](./operations/troubleshooting/);
  - [«Чек-лист перед вводом в эксплуатацию»](./operations/production-checklist/), [«Управление ключами»](./operations/key-management/), [«TLS-сертификаты»](./operations/tls-certificates/);
  - [«Квоты»](./operations/quotas/), [«Сайзинг»](./operations/sizing/), [«FAQ»](./operations/faq/).

- Архитектура
  - [«Обзор»](./architecture/overview/), [«Развёртывание в DP»](./architecture/deployment-dkp/), [«Развёртывание Standalone»](./architecture/deployment-standalone/);
  - [«Сетевые порты»](./architecture/ports/), [«Модель угроз»](./architecture/threat-model/), [«Функциональные характеристики»](./architecture/functional-specifications/).

- Аудит
  - [«Обзор»](./audit/overview/) — что попадает в журналы аудита Stronghold, какие бэкенды поддерживаются и как безопасно настраивать аудит;
  - [«Схема записей журнала аудита»](./audit/log-format/) — структура аудит-записей, ключевые объекты и защита чувствительных данных;
  - [«Фильтрация данных аудита»](./audit/filtering/) — отбор аудит-записей по условиям и настройка  резервного устройства;
  - [«Исключение полей из данных аудита»](./audit/exclusion/) — удаление отдельных полей из аудит-записей перед сохранением.

- Резервное копирование и восстановление
  - [«Обзор»](./backups/overview/) — обзор ручных и автоматических снимков встроенного хранилища Stronghold;
  - [«Создание снимка»](./backups/save/) — ручное сохранение снимка через CLI и API;
  - [«Проверка снимка»](./backups/inspect/) — локальная проверка содержимого и базовой консистентности файла снимка;
  - [«Восстановление из снимка»](./backups/restore/) — восстановление кластера Stronghold из сохранённого снимка;
  - [«Автоматические снимки»](./backups/automated-snapshots/) — настройка расписания, хранилища и статуса автоматических резервных копий.

- KMS и HSM
  - [«Поддержка HSM»](./kms-hsm/hsm/) — работа с шифрованием HSM по стандарту PKCS11 для автоматического распечатывания хранилища и защиты root-ключа;
  - [«Yandex Cloud KMS»](./kms-hsm/yandexcloudkms/) — настройка `seal "yandexcloudkms"` для автоматического распечатывания хранилища и защиты root-ключа;
  - [«Двойное шифрование (seal wrapping)»](./kms-hsm/sealwrap/) — описание механизма `seal wrap`, добавляющего дополнительный уровень шифрования для критически важных данных.

- Репликация
  - [«Обзор»](./replication/overview/) — нативная межкластерная репликация (PR) и репликация для аварийного восстановления (DR) между кластерами Stronghold;
  - [«Архитектура: CE и EE»](./replication/architecture/) — слои узла и как их порядок определяет поведение репликации;
  - [«Межкластерная репликация (PR)»](./replication/performance/) — масштабирование чтений, настройка secondary и фильтры путей;
  - [«Репликация для аварийного восстановления (DR)»](./replication/disaster-recovery/) — горячий резерв, аварийное переключение и церемония promote;
  - [«Performance standby»](./replication/performance-standby/) — чтения на неактивных HA-узлах внутри кластера;
  - [«Репликация KV1/KV2»](./replication/kv-replication/) — настройка pull-репликации KV1/KV2 между кластерами Stronghold, включая CLI и API-сценарии.

- Пространства имён
  - [«Обзор»](./namespaces/overview/) — изоляция конфигурации и секретов между пространствами имён, управление через CLI и API, а также блокировка Namespace API.

- Криптоалгоритмы
  - [«Обзор»](./cryptography/overview/) — обзор TLS, шифрования хранилища, HSM, а также доступных алгоритмов в PKI и Transit.

- Расширения и интеграции
  - [«Обзор»](./plugins/overview/) - обзор встроенных и внешних плагинов Stronghold и различий между Standalone и DP.
  - [«Плагины в Standalone»](./plugins/standalone/) - plugin directory, регистрация, versioning и подключение внешних плагинов на Linux-сервере.
  - [«Плагины в DP»](./plugins/dkp/) - загрузка плагинов через `ModuleConfig`, регистрация и включение в Deckhouse Platform.
