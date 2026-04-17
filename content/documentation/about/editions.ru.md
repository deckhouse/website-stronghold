---
title: "Редакции Deckhouse Stronghold"
linkTitle: "Редакции"
description: "Редакции Deckhouse Stronghold и сравнение их возможностей"
weight: 40
---

**Deckhouse Stronghold** поставляется в трёх редакциях:

- **Deckhouse Stronghold Community Edition (CE)** — базовая редакция для использования в типовых сценариях управления секретами;
- **Deckhouse Stronghold Enterprise Edition (EE)** — коммерческая редакция с расширенными возможностями для production-окружений;
- **Deckhouse Stronghold Certified Security Edition (CSE)** — сертифицированная редакция для организаций с повышенными требованиями к информационной безопасности.

**Deckhouse Stronghold Certified Security Edition** сертифицирована ФСТЭК России. Сертификат соответствия ФСТЭК России № 5038 от 10 февраля 2026 года подтверждает, что продукт соответствует требованиям технических условий и приказа ФСТЭК России № 76 от 2 июня 2020 года по 4-му уровню доверия. Эта редакция предназначена для сценариев, где обязательно использование сертифицированных средств защиты информации.

## Лицензирование и совместимость с DKP

Редакции Deckhouse Stronghold используются в связке с редакциями **Deckhouse Kubernetes Platform** в следующем порядке:

- **Deckhouse Stronghold CE** доступен для использования в любой редакции **Deckhouse Kubernetes Platform**;
- **Deckhouse Stronghold EE** лицензируется отдельно и доступен для использования в любой коммерческой редакции **Deckhouse Kubernetes Platform**;
- **Deckhouse Stronghold CSE** лицензируется отдельно и доступен для использования только в редакции **DKP CSE**.

> Важно  
> Если вы планируете использовать Deckhouse Stronghold вне DKP, обратите внимание на строку **«Возможность поставки в виде исполняемого файла (Standalone)»** в таблице ниже.

## Сравнение редакций

Ниже приведено краткое сравнение ключевых возможностей и особенностей редакций Deckhouse Stronghold.

| Возможности | CE | EE | CSE |
| --- | --- | --- | --- |
| Безопасное управление жизненным циклом секретов (хранение, создание, доставка, отзыв и ротация) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Возможность использования инструментов автоматизации IaC (Ansible, Terraform) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Поддержка методов аутентификации | JWT, OIDC, Kubernetes, LDAP, Token, **WebAuthn** | JWT, OIDC, Kubernetes, LDAP, Token, **WebAuthn**, **SAML** | JWT, OIDC, Kubernetes, LDAP, Token |
| Поддержка механизмов секретов KV, Kubernetes, Database, SSH, PKI | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Веб-интерфейс | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Управление ролями и политиками доступа через веб-интерфейс | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Поддержка пространств имён (namespaces) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Управляемые ключи (Managed Keys) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| Поддержка ГОСТ-алгоритмов для PKI/Transit | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| Развёртывание в закрытом контуре | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Встроенное автоматическое распечатывание хранилища (auto unseal) без использования внешних сервисов и KMS | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Поддержка конфигураций с высокой доступностью (HA) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Межкластерная репликация данных | {{< icon-edition type="not_supported" >}} | KV1/KV2 | KV1/KV2 |
| Автоматическое создание резервных копий по заданному расписанию | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Поддержка аудит-логирования | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Возможность поставки в виде исполняемого файла (Standalone) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Поддержка российских ОС ([подробнее...](/products/kubernetes-platform/documentation/v1/supported_versions.html)) | РЕД ОС, ALT Linux, Astra Linux Special Edition, **РОСА Сервер** | РЕД ОС, ALT Linux, Astra Linux Special Edition, **РОСА Сервер** | РЕД ОС, ALT Linux, Astra Linux Special Edition |
| Сертификат соответствия требованиям приказа ФСТЭК России № 76 по 4-му уровню доверия | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} |
| Возможность запуска в DKP CE | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="not_supported" >}} |
| [Гарантийная техническая поддержка](https://deckhouse.ru/tech-support/) | {{< icon-edition type="techsupport_ce" >}} | {{< icon-edition type="techsupport_commercial" >}} | {{< icon-edition type="techsupport_commercial" >}} |
| [Техподдержка «Стандарт»](https://deckhouse.ru/tech-support/) | {{< icon-edition type="techsupport_ce" >}} | {{< icon-edition type="techsupport_commercial" >}} | {{< icon-edition type="techsupport_commercial" >}} |
| [Техподдержка «Стандарт +»](https://deckhouse.ru/tech-support/) | {{< icon-edition type="techsupport_ce" >}} | {{< icon-edition type="techsupport_commercial" >}} | {{< icon-edition type="techsupport_commercial" >}} |

## Как выбрать редакцию

Выбор редакции зависит от требований вашей инфраструктуры и сценариев использования:

- **CE** подойдёт, если нужен базовый набор возможностей для хранения секретов и интеграции с инфраструктурой в рамках стандартных сценариев.
- **EE** подойдёт, если требуются расширенные функции для enterprise-эксплуатации: пространства имён, расширенное управление доступом, управляемые ключи, аудит, резервное копирование по расписанию и поставка в формате Standalone.
- **CSE** подойдёт, если помимо enterprise-возможностей необходимо использовать сертифицированную редакцию в инфраструктурах с повышенными требованиями к информационной безопасности и нормативному соответствию.

## Что дальше

Чтобы перейти к требованиям к инфраструктуре и условиям эксплуатации продукта, откройте раздел [Системные требования](../system-requirements/).
