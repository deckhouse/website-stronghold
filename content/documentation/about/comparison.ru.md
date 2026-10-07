---
title: "Сравнение с другими решениями"
linkTitle: "Сравнение"
description: "Сравнение Stronghold с HashiCorp Vault, OpenBao и Yandex Lockbox по способу развёртывания, совместимости API, механизмам секретов, репликации, HSM и модели лицензирования."
weight: 50
---

Таблица ниже помогает сопоставить Stronghold с другими системами управления секретами по проверяемым характеристикам. Сравнение не оценивает продукты в целом: выбор зависит от требований к инфраструктуре, регуляторных ограничений и имеющихся компетенций.

{{< alert level="info" >}}
Сведения о сторонних продуктах приведены по их публичной документации и могут меняться с выходом новых версий. Перед принятием решения сверьтесь с актуальной документацией производителя.
{{< /alert >}}

<!-- TODO(verify): перед публикацией сверьте все ячейки столбцов HashiCorp Vault, OpenBao и Yandex Lockbox с актуальной документацией производителей и укажите дату проверки. -->

## Сравнительная таблица

| Характеристика | Stronghold | HashiCorp Vault | OpenBao | Yandex Lockbox |
| --- | --- | --- | --- | --- |
| Модель поставки | Модуль DP; исполняемый файл для Linux (standalone) — Stronghold EE | Исполняемый файл, контейнер, Helm-чарт; managed-сервис HCP Vault | Исполняемый файл, контейнер, Helm-чарт | Managed-сервис в Yandex Cloud |
| Развёртывание в собственной инфраструктуре (on-premise), в том числе в закрытом контуре | Да | Да | Да | Нет <!-- TODO(verify) --> |
| Совместимость с HTTP API Vault | Да ([подробнее](../vault-compatibility/)) | — | Да <!-- TODO(verify): степень совместимости в текущих версиях OpenBao --> | Нет, собственный API Yandex Cloud |
| Статические секреты (KV) | Да, KV1 и KV2 с версионированием | Да | Да | Да |
| Динамические секреты (базы данных, Kubernetes, LDAP и другие) | Да | Да | Да | Нет <!-- TODO(verify) --> |
| Выпуск сертификатов (PKI) | Да, включая ГОСТ 34.10 (ГОСТ — кроме Stronghold CSE) | Да | Да | Нет, выпуск сертификатов — в отдельном сервисе Yandex Certificate Manager <!-- TODO(verify) --> |
| Шифрование как сервис (Transit) | Да, включая ГОСТ-алгоритмы (ГОСТ — кроме Stronghold CSE) | Да | Да | Нет, криптографические операции — в отдельном сервисе Yandex KMS <!-- TODO(verify) --> |
| Пространства имён | Stronghold EE | Vault Enterprise | Да <!-- TODO(verify): версия OpenBao, с которой доступны пространства имён --> | Изоляция на уровне каталогов и облаков Yandex Cloud <!-- TODO(verify) --> |
| Высокая доступность внутри кластера | Да, integrated Raft; performance standby — Stronghold EE | Да; performance standby — Vault Enterprise | Да <!-- TODO(verify): поддержка чтения со standby-узлов --> | Обеспечивается провайдером |
| Межкластерная репликация | Stronghold EE: Performance, DR, KV1/KV2 | Vault Enterprise: Performance, DR | Нет <!-- TODO(verify) --> | Обеспечивается провайдером |
| Автоматическое распечатывание и защита root-ключа в HSM (PKCS #11) | Stronghold EE (standalone) | Vault Enterprise | <!-- TODO(verify): наличие seal pkcs11 в OpenBao --> | Неприменимо <!-- TODO(verify): поддержка HSM-ключей Yandex KMS для шифрования секретов Lockbox --> |
| Автоматическое распечатывание через облачный KMS | Yandex Cloud KMS | AWS KMS, Azure Key Vault, GCP Cloud KMS и другие <!-- TODO(verify) --> | AWS KMS, Azure Key Vault, GCP Cloud KMS и другие <!-- TODO(verify) --> | Неприменимо |
| ГОСТ-криптография | Да ([подробнее](../../admin/cryptography/overview/)) | Нет <!-- TODO(verify) --> | Нет <!-- TODO(verify) --> | <!-- TODO(verify) --> |
| Сертификат ФСТЭК России | Stronghold CSE (сертификат №5038) | Нет <!-- TODO(verify) --> | Нет <!-- TODO(verify) --> | <!-- TODO(verify): сертификаты Yandex Cloud, распространяющиеся на Lockbox --> |
| Запись в Едином реестре российских программ | Да, №22339 | Нет <!-- TODO(verify) --> | Нет <!-- TODO(verify) --> | <!-- TODO(verify) --> |
| Модель лицензирования | Базовый Stronghold — бесплатно; Stronghold EE и Stronghold CSE — коммерческая лицензия <!-- TODO(verify): лицензия исходного кода базового Stronghold --> | Business Source License 1.1 (с версии 1.15); Vault Enterprise — коммерческая лицензия <!-- TODO(verify) --> | Mozilla Public License 2.0 <!-- TODO(verify) --> | Оплата по факту использования облачного сервиса <!-- TODO(verify) --> |

## Как пользоваться сравнением

- Если требуется развёртывание в закрытом контуре и совместимость с существующими инструментами Vault, сравните Stronghold, HashiCorp Vault и OpenBao по функциям из вашего списка требований.
- Если инфраструктура уже размещена в Yandex Cloud и нужны только статические секреты, оцените managed-сервис. Для динамических секретов, PKI и Transit в собственной инфраструктуре потребуется решение с API Vault.
- Если действуют требования к сертифицированным средствам защиты информации или к отечественному ПО, см. [«Соответствие требованиям ИБ»](../compliance/).

Отличия Stronghold от upstream Vault подробно описаны на странице [«Совместимость с HashiCorp Vault»](../vault-compatibility/), возможности редакций — на странице [«Редакции»](../editions/).
