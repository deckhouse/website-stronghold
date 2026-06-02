---
title: "Редакции"
weight: 10
---

Deckhouse Stronghold поставляется в редакциях Community Edition (CE), Enterprise Edition (EE) и Certified Security Edition (CSE), сертифицированной ФСТЭК России для сред с повышенными требованиями к информационной безопасности.

Deckhouse Stronghold CE доступен для использования в любой редакции Deckhouse Kubernetes Platform (DKP).

Deckhouse Stronghold EE и Deckhouse Stronghold CSE лицензируются отдельно. Deckhouse Stronghold EE доступен для использования в любой **коммерческой редакции** DKP. Deckhouse Stronghold CSE доступен для использования только в редакции DKP CSE.

Краткое сравнение ключевых возможностей и особенностей редакций Deckhouse Stronghold:

| Возможности | CE | EE | CSE |
| --- | --- | --- | --- |
| **Механизмы секретов** | | | |
| KV | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| PKI | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Database | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| SSH | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Transit | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| **Аутентификация** | | | |
| Ограничение возможности аутентификации на основе IP-адреса, клиента и т.д. | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Поддержка внешних систем аутентификации (Identity Plugins) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| UserPass | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| AppRole | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Kubernetes | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| LDAP | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Dex | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| OIDC | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| JWT | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| MFA (внешняя интеграция) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| WebAuth (FIDO2/Passkeys) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| SAML | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| **Управление секретами** | | | |
| Управление жизненным циклом секретов (хранение, создание, доставка, отзыв и ротация) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Поддержка динамических секретов | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Создание собственного хранилища секретов, недоступного другим root-токенам | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Безопасная доставка секретов в приложения в виде переменных окружения | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Безопасная доставка секретов в приложения в виде файлов | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Безопасная доставка бинарных файлов как секретов (keytab, GPG-ключей, файлов лицензий ПО и т. д.) в контейнеры | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| **Серверная часть хранилища** | | | |
| Raft integrated storage | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Возможность поддержки другого внешнего хранилища (по согласованию с вендором) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| **Механизмы автоматического распечатывания** | | | |
| Встроенный auto unseal через raft | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Auto unseal через внешний аппаратный модуль безопасности (TPM, HSM) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Auto unseal через внешний сервис управления ключами (KMS) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Auto unseal через Transit Secrets Engine | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| **Отказоустойчивость** | | | |
| HA-конфигурация из коробки (кластер из 3 мастер-узлов) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Поддержка фильтров репликации при передаче секретов между узлами отказоустойчивой архитектуры (Replication Filters) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Резервное копирование по заданному расписанию без использования внешних сервисов-планировщиков и скриптов | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Межкластерная репликация данных (KV1/KV2) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Репликация для катастрофоустойчивости (Disaster Recovery репликация) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| **Производительность** | | | |
| Репликация для производительности (performance репликация) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| **Защита данных** | | | |
| Шифрование данных «на лету» без сохранения (Encryption-as-a-Service) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Прозрачное обновление ключа шифрования | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Проверка подписи при расшифровке данных | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Создание собственного хранилища секретов, недоступного другим root-токенам | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Разделение ключа для распечатывания хранилища на несколько ключей | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Использование криптографических ключей, расположенных во внешней доверенной системе (Managed Keys) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| Шифрование root-ключа с помощью внешнего аппаратного модуля безопасности (HSM) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Двойное шифрование данных с помощью внешнего аппаратного модуля безопасности (HSM) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Шифрование на базе российских криптографических алгоритмов (ГОСТ) для Seal Wrap, TLS, PKI, Transit secret engine | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| **Управление доступом** | | | |
| Пространства имен (namespaces) | {{< icon-edition type="clock" title="В процессе реализации" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Настраиваемые политики контроля доступа | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Идентификация по выданным токенам доступа | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Управление доступом через локальные группы доступа | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Управление жизненным циклом локальной учетной записи | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| **Администрирование** | | | |
| Веб-интерфейс | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Возможность работы через консольную утилиту (CLI) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Возможность централизованного управления кластером через общий API (Cluster Management) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Централизованный сбор событий | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Управление ролями приложений (AppRole, OIDC/JWT Role) через веб-интерфейс (UI) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| **Мониторинг** | | | |
| Централизованный сбор событий | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Встроенный аудит событий системы | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Распределение событий аудита по разным аудит логам на основании заданных фильтров | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| Просмотр аудит событий системы через веб-интерфейс (UI) и программный интерфейс (API) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| **Поддержка внешних систем и компонентов** | | | |
| Инструменты автоматизации IaC (Ansible, Terraform) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Системы управления базами данных (PostgreSQL, MySQL, MongoDB, PostgresPro Enterprise и т. д.) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Российские операционные системы (Astra Linux, Ред ОС, ALT ОС, РОСА Сервер) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| Российские Identity-провайдеры (Identity Blitz) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Российские службы управления каталогами (ALD Pro) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Российские системы резервного копирования данных (RuBackup) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Российские системы двухфакторной аутентификации (MULTIFACTOR, MFA SAS) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Российские аппаратные модули безопасности (КриптоПро HSM, Рутокен ЭЦП 3.0) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Системы управления ключами (Yandex KMS) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| **Исполняемые среды** | | | |
| Запуск в виде модуля в платформе контейнеризации в Deckhouse Kubernetes Platform Community Edition | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="not_supported" >}} |
| Запуск в виде модуля в платформе контейнеризации в коммерческих редакциях Deckhouse Kubernetes Platform | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Запуск в виде исполняемого файла вне платформы контейнеризации Deckhouse Kubernetes Platform на ОС Linux | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| **Сертификация** | | | |
| Сертификат соответствия требованиям Приказа ФСТЭК России от 4 июля 2022 г. № 118 (функциональный модуль в составе платформы контейнеризации) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| Сертификат соответствия требованиям ТУ и Приказа ФСТЭК России от 2 июня 2020 г. № 76 | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} |
