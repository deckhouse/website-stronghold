---
title: "Редакции"
description: "Сравнение возможностей редакций Stronghold, Stronghold EE и Stronghold CSE и условия их использования в Deckhouse Kubernetes Platform."
weight: 20
---

Deckhouse Stronghold поставляется в трёх редакциях:

- **Stronghold** (базовый Stronghold, Community Edition) — доступен для использования в любой редакции Deckhouse Platform (DP);
- **Stronghold EE** (Enterprise Edition) — лицензируется отдельно и доступен для использования в любой **коммерческой редакции** DP;
- **Stronghold CSE** (Certified Security Edition) — сертифицирован ФСТЭК России для сред с повышенными требованиями к информационной безопасности, лицензируется отдельно и доступен для использования только в редакции DKP CSE.

Порядок перехода между редакциями описан в разделах [«Переключение Stronghold на редакцию EE»](../../install/dkp/platform-management/switching-editions/ce-to-ee/) и [«Переключение Stronghold с EE на CSE»](../../install/dkp/platform-management/switching-editions/ee-to-cse/).

## Сравнение возможностей

| Возможности | Stronghold | Stronghold EE | Stronghold CSE |
| --- | --- | --- | --- |
| Безопасное управление жизненным циклом секретов (хранение, создание, доставка, отзыв и ротация) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Возможность использования инструментов автоматизации IaC (Ansible, Terraform) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Поддержка методов аутентификации | JWT, OIDC, Kubernetes, LDAP, Token, **WebAuthn** | JWT, OIDC, Kubernetes, LDAP, Token, **WebAuthn**, **[SAML](../../user/auth/saml/)** | JWT, OIDC, Kubernetes, LDAP, Token |
| Поддержка механизмов секретов KV, Kubernetes, Database, SSH, PKI | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| [Механизм секретов GitOps](../../user/secrets-engines/gitops/overview/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | Уточняется <!-- TODO(verify): доступность в Stronghold CSE --> |
| [Встроенный плагин `trdl`](../../user/secrets-engines/trdl/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | Уточняется <!-- TODO(verify): доступность в Stronghold CSE --> |
| Поддержка российских ОС ([полный список поддерживаемых ОС](/products/kubernetes-platform/documentation/v1/supported_versions.html)) | РЕД ОС, ALT Linux, Astra Linux Special Edition, **РОСА Сервер** | РЕД ОС, ALT Linux, Astra Linux Special Edition, **РОСА Сервер** | РЕД ОС, ALT Linux, Astra Linux Special Edition |
| Развёртывание в закрытом контуре | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Веб-интерфейс | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Управление ролями и политиками доступа через веб-интерфейс | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| [Пространства имён (namespaces)](../../admin/namespaces/overview/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| [Встроенное автоматическое распечатывание хранилища (auto unseal) без использования внешних сервисов и KMS](../../concepts/seal/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| [Автоматическое распечатывание и защита root-ключа с помощью HSM (PKCS #11)](../../admin/kms-hsm/hsm/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | Уточняется <!-- TODO(verify): доступность в Stronghold CSE --> |
| [Двойное шифрование (seal wrapping)](../../admin/kms-hsm/sealwrap/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | Уточняется <!-- TODO(verify): доступность в Stronghold CSE --> |
| Поддержка конфигураций с высокой доступностью (HA) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| [Performance standby](../../admin/replication/performance-standby/) — обслуживание чтений standby-узлами HA | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | Уточняется <!-- TODO(verify): доступность в Stronghold CSE --> |
| [Межкластерная репликация данных](../../admin/replication/overview/) | {{< icon-edition type="not_supported" >}} | KV1/KV2, Performance, DR | KV1/KV2, Performance, DR |
| [Автоматическое создание резервных копий по заданному расписанию](../../admin/backups/automated-snapshots/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| [Аудит-логирование](../../admin/audit/overview/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| [Управляемые ключи (Managed Keys)](../../user/managed-keys/overview/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| [Поддержка ГОСТ-алгоритмов для PKI/Transit](../../admin/cryptography/overview/) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| Возможность поставки в виде исполняемого файла ([standalone](../../install/standalone/)) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Сертификат соответствия требованиям Приказа ФСТЭК России №76 по 4 уровню доверия | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} |
| Возможность запуска в DP CE | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="not_supported" >}} |

{{< alert level="info" >}}
Работа с HSM и `seal "yandexcloudkms"` в текущей версии поддерживается только при standalone-установке Stronghold.
{{< /alert >}}
