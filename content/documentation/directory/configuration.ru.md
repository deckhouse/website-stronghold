---
title: "Параметры конфигурации"
linkTitle: "Параметры конфигурации"
description: "Справочник по основным группам параметров конфигурации Deckhouse Stronghold"
weight: 30
---

Эта страница помогает быстро понять, какие группы параметров конфигурации есть в Deckhouse Stronghold и где искать подробные инструкции.

Это не пошаговое руководство по настройке. Подробные примеры и сценарии уже описаны в тематических разделах документации. Здесь собраны основные области конфигурации и ссылки на страницы, где они раскрыты подробнее.

## Что можно настраивать

В документации Deckhouse Stronghold параметры конфигурации относятся к нескольким уровням:

- конфигурация Stronghold Agent;
- параметры механизмов секретов;
- параметры ролей и TTL;
- параметры identity и OIDC;
- параметры подключений к внешним системам.

Если вам нужно быстро найти, где настраивается конкретный параметр, начните с разделов ниже.

## Stronghold Agent

Параметры Stronghold Agent задаются в HCL-конфигурации. В документации уже описаны основные секции конфигурационного файла:

- `stronghold` — подключение к серверу Deckhouse Stronghold;
- `auto_auth` — автоматическая аутентификация и sink для токена;
- `template` — рендеринг шаблонов в файлы;
- `template_config` — глобальные параметры шаблонов;
- `exec` — запуск дочернего процесса;
- `env_template` — передача секретов через переменные окружения;
- `listener` — listener для API Proxy;
- параметры логирования и отладки.

Подробности смотрите на странице [Настройка](../user/stronghold-agent/configuration/).

## Механизмы секретов

Параметры механизмов секретов зависят от конкретного secrets engine. В документации уже описаны такие группы параметров.

### KV

Для механизма секретов `KV` настраиваются:

- версия хранилища: KV v1 или KV v2;
- политики ACL для путей `data`, `metadata`, `delete`, `undelete`, `destroy`;
- параметры метаданных, например `max_versions`, `delete_version_after`, `custom_metadata`;
- операции `put`, `get`, `patch`, `delete`, `undelete`, `destroy`.

Подробности:
- [KV v1](../user/secrets-engines/kv/kv-v1/)
- [KV v2](../user/secrets-engines/kv/kv-v2/)

### SSH

Для механизма секретов `SSH` настраиваются:

- CA для подписи клиентских и host-ключей;
- роли подписи;
- TTL сертификатов;
- разрешённые пользователи, домены и расширения сертификатов;
- параметры `default_user`, `valid_principals`, `allowed_extensions`, `default_extensions`.

Подробности:
- [SSH](../user/secrets-engines/ssh/)

### Transit

Для `Transit` настраиваются:

- именованные ключи;
- типы ключей;
- ротация ключей;
- `min_decryption_version`;
- параметры конвергентного шифрования;
- импорт собственного ключа (BYOK).

Подробности:
- [Transit](../user/secrets-engines/transit/)

### TOTP

Для `TOTP` настраиваются:

- режим работы: генератор или провайдер;
- именованные ключи;
- параметры `url`, `generate`, `issuer`, `account_name`.

Подробности:
- [TOTP](../user/secrets-engines/totp/)

### PKI

Для `PKI` настраиваются:

- путь монтирования;
- максимальный TTL механизма;
- корневой или промежуточный CA;
- URL выпускающих сертификатов и CRL;
- роли для выпуска сертификатов;
- параметры `allowed_domains`, `allow_subdomains`, `max_ttl`.

Подробности:
- [PKI](../user/secrets-engines/pki/)

### LDAP

Для `LDAP` настраиваются:

- подключение к LDAP через `binddn`, `bindpass`, `url`;
- схема `openldap`, `racf` или `ad`;
- параметры статических ролей;
- `rotation_period`;
- параметры динамических ролей: `creation_ldif`, `deletion_ldif`, `rollback_ldif`, `default_ttl`, `max_ttl`;
- параметры для ротации паролей набора учётных записей.

Подробности:
- [LDAP](../user/secrets-engines/ldap/)

### Kubernetes

Для механизма секретов `Kubernetes` настраиваются:

- роль Stronghold;
- `allowed_kubernetes_namespaces`;
- `service_account_name`;
- `token_default_ttl` и `token_max_ttl`;
- `token_default_audiences` и `audiences`;
- автоматическое управление `ServiceAccount`, `Role` и `RoleBinding`;
- `kubernetes_role_name`;
- `generated_role_rules`.

Подробности:
- [Kubernetes](../user/secrets-engines/kubernetes/)

### Базы данных

Для механизма секретов баз данных настраиваются:

- подключение к базе данных;
- используемый плагин;
- `connection_url`;
- `allowed_roles`;
- `creation_statements`;
- `default_ttl` и `max_ttl`.

Эти параметры описаны отдельно для каждой поддерживаемой СУБД:

- [PostgreSQL](../user/secrets-engines/database/postgresql/)
- [MySQL](../user/secrets-engines/database/mysql/)
- [ClickHouse](../user/secrets-engines/database/clickhouse/)

## Identity и OIDC

Параметры identity и OIDC относятся к клиентским приложениям, токенам и проверке идентичности.

### OIDC identity provider

Для Stronghold как OIDC identity provider настраиваются:

- клиентское приложение;
- `redirect_uris`;
- `assignments`;
- `client_id`;
- `client_secret`;
- `issuer`.

Подробности:
- [OIDC identity provider](../user/secrets-engines/identity/oidc-provider/)

### OIDC identity tokens

Для OIDC identity tokens настраиваются:

- именованные ключи;
- ротация ключей;
- `verification_ttl`;
- алгоритм подписи;
- `client_id`;
- TTL токена;
- шаблоны claims;
- `issuer`.

Подробности:
- [OIDC identity tokens](../user/secrets-engines/identity/oidc-tokens/)

## Роли и TTL

Во многих механизмах секретов есть параметры ролей и срока действия учётных данных.

Чаще всего в документации встречаются:

- `default_ttl`;
- `max_ttl`;
- TTL сертификатов;
- TTL токенов;
- TTL временных учётных данных;
- параметры ролей, которые определяют, какие credentials можно выдавать и на каких условиях.

Если вы ищете параметры, связанные со сроком действия секретов, токенов или сертификатов, сначала откройте страницу конкретного механизма секретов.

## Как искать нужный параметр

Если вы не знаете, где описан параметр, используйте такой порядок:

- если параметр относится к `agent.hcl`, откройте [Настройка](../user/stronghold-agent/configuration/);
- если параметр связан с секретами, откройте страницу соответствующего механизма секретов;
- если параметр связан с `client_id`, `issuer`, ключами подписи или claims, откройте страницы identity и OIDC;
- если параметр связан с TTL, ролями или генерацией временных учётных данных, проверьте страницу нужного secrets engine.

## Что дальше

- Для конфигурации Stronghold Agent откройте [Настройка](../user/stronghold-agent/configuration/).
- Для параметров механизмов секретов откройте нужную страницу в разделе «Механизмы секретов».
- Для identity и OIDC используйте раздел `Identity`.
