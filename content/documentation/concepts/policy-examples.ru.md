---
title: "Примеры политик"
description: "Библиотека готовых политик Stronghold: чтение KV, доступ приложения к своим секретам через шаблоны, CI/CD, администратор пространства имён, резервное копирование, аудитор, PKI, Transit и администратор безопасности."
weight: 42
---

На этой странице собраны готовые политики для типовых сценариев. Синтаксис политик, список возможностей и правила шаблонов описаны в разделе [«Политики»](../policy/).

Перед использованием замените пути точек монтирования (`secret/`, `pki_int/`, `transit/` и т. д.), имена ролей и ключей на свои. Чтобы узнать, какие возможности нужны для конкретной команды, добавьте к ней флаг `-output-policy`.

Сохраните политику в файл и загрузите её в Stronghold:

```bash
d8 stronghold policy write kv-read-only kv-read-only.hcl
```

## Чтение секретов KV только для чтения

Доступ на чтение к секретам `secret/app/*` в механизме KV версии 2. Для KV v2 данные находятся по пути `secret/data/`, а метаданные и список ключей — по пути `secret/metadata/`.

```hcl
# Чтение секретов.
path "secret/data/app/*" {
  capabilities = ["read"]
}

# Просмотр списка ключей и метаданных.
path "secret/metadata/app/*" {
  capabilities = ["read", "list"]
}
```

Для KV версии 1 используйте путь без `data/` и `metadata/`:

```hcl
path "kv/app/*" {
  capabilities = ["read", "list"]
}
```

## Доступ приложения к собственным секретам

Одна политика для всех приложений: каждое приложение получает доступ только к секретам по пути со своим именем сущности. Подходит, если имя сущности совпадает с именем приложения, например при входе через AppRole или Kubernetes.

```hcl
# Полный доступ к собственным секретам.
path "secret/data/apps/{{identity.entity.name}}/*" {
  capabilities = ["create", "read", "update", "patch", "delete"]
}

path "secret/metadata/apps/{{identity.entity.name}}/*" {
  capabilities = ["read", "list", "delete"]
}

# Только чтение общих секретов.
path "secret/data/shared/*" {
  capabilities = ["read"]
}
```

Вместо имени сущности можно использовать метаданные псевдонима, например пространство имён Kubernetes сервисного аккаунта:

```hcl
path "secret/data/{{identity.entity.aliases.auth_kubernetes_xxxx.metadata.service_account_namespace}}/*" {
  capabilities = ["read"]
}
```

Замените `auth_kubernetes_xxxx` на accessor метода аутентификации Kubernetes (`d8 stronghold auth list`).

## Политика для CI/CD

Политика для задания развёртывания: чтение секретов развёртывания, получение динамических учётных данных базы данных, выпуск токенов по роли хранилища токенов и управление собственным токеном.

```hcl
# Секреты для развёртывания.
path "secret/data/ci/deploy/*" {
  capabilities = ["read"]
}

# Динамические учётные данные базы данных.
path "database/creds/deploy" {
  capabilities = ["read"]
}

# Выпуск токенов для развёрнутого приложения по роли хранилища токенов.
path "auth/token/create/app-runtime" {
  capabilities = ["update"]
}

# Управление собственным токеном.
path "auth/token/lookup-self" {
  capabilities = ["read"]
}

path "auth/token/renew-self" {
  capabilities = ["update"]
}

path "auth/token/revoke-self" {
  capabilities = ["update"]
}
```

## Администратор пространства имён

Политика для администратора пространства имён в Stronghold EE. Создайте её внутри пространства имён: все пути в ней считаются относительно этого пространства имён.

```bash
d8 stronghold policy write -namespace=team-a namespace-admin namespace-admin.hcl
```

```hcl
# Управление дочерними пространствами имён.
path "sys/namespaces/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Управление политиками.
path "sys/policies/acl" {
  capabilities = ["list"]
}

path "sys/policies/acl/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Управление методами аутентификации.
path "sys/auth" {
  capabilities = ["read"]
}

path "sys/auth/*" {
  capabilities = ["create", "read", "update", "delete", "sudo"]
}

path "auth/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Управление механизмами секретов.
path "sys/mounts" {
  capabilities = ["read"]
}

path "sys/mounts/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Управление идентификационными данными.
path "identity/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Работа с секретами в пространстве имён.
path "secret/*" {
  capabilities = ["create", "read", "update", "patch", "delete", "list"]
}

# Проверка собственных возможностей.
path "sys/capabilities-self" {
  capabilities = ["update"]
}
```

Подробнее о пространствах имён — в разделе [«Пространства имён»](../../admin/namespaces/overview/).

## Оператор резервного копирования

Политика для создания снимков Raft вручную и управления автоматическим резервным копированием без доступа к секретам. Эндпоинты снимков описаны в разделе [«Резервное копирование в Stronghold»](../../admin/backups/overview/).

```hcl
# Создание снимка (d8 stronghold operator raft snapshot save).
path "sys/storage/raft/snapshot" {
  capabilities = ["read"]
}

# Настройка автоматических снимков.
path "sys/storage/raft/snapshot-auto/config" {
  capabilities = ["list"]
}

path "sys/storage/raft/snapshot-auto/config/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}

# Статус автоматических снимков.
path "sys/storage/raft/snapshot-auto/status/*" {
  capabilities = ["read"]
}

# Просмотр состава Raft-кластера.
path "sys/storage/raft/configuration" {
  capabilities = ["read"]
}
```

{{< alert level="warning" >}}
Снимок содержит все данные Stronghold в зашифрованном виде. Храните файлы снимков в защищённом месте. Восстановление из снимка (`POST sys/storage/raft/snapshot`) не включено в эту политику: выдавайте право `update` на этот путь только администраторам.
{{< /alert >}}

Путь `sys/storage/raft/snapshot-auto/config/*` защищён правами root, поэтому для настройки автоматических снимков нужна возможность `sudo`. Для создания снимка (`sys/storage/raft/snapshot`) и чтения статуса `sudo` не требуется.

## Аудитор

Политика для просмотра конфигурации безопасности без доступа к секретам: аудит-устройств, политик, точек монтирования, методов аутентификации и идентификационных данных. Путь `sys/audit` защищён правами root, поэтому для него нужна возможность `sudo`.

```hcl
# Список аудит-устройств.
path "sys/audit" {
  capabilities = ["read", "sudo"]
}

# Просмотр политик.
path "sys/policies/acl" {
  capabilities = ["list"]
}

path "sys/policies/acl/*" {
  capabilities = ["read"]
}

# Просмотр точек монтирования и методов аутентификации.
path "sys/mounts" {
  capabilities = ["read"]
}

path "sys/auth" {
  capabilities = ["read"]
}

# Просмотр сущностей и групп.
path "identity/entity/id" {
  capabilities = ["list"]
}

path "identity/entity/id/*" {
  capabilities = ["read"]
}

path "identity/group/id" {
  capabilities = ["list"]
}

path "identity/group/id/*" {
  capabilities = ["read"]
}
```

Доступ к секретам этой политикой не предоставляется: в Stronghold всё, что явно не разрешено, запрещено.

## Выпуск сертификатов PKI

Политика для сервиса, который выпускает сертификаты по роли `web-server` промежуточного УЦ.

```hcl
# Выпуск сертификата с генерацией ключа в Stronghold.
path "pki_int/issue/web-server" {
  capabilities = ["update"]
}

# Подпись CSR, сгенерированного на стороне клиента.
path "pki_int/sign/web-server" {
  capabilities = ["update"]
}

# Чтение сертификата УЦ и цепочки.
path "pki_int/cert/ca" {
  capabilities = ["read"]
}

path "pki_int/ca_chain" {
  capabilities = ["read"]
}
```

Подробнее — в разделе [«Механизм секретов PKI»](../../user/secrets-engines/pki/).

## Только шифрование в Transit

Приложение может шифровать данные ключом `app-key`, но не может их расшифровывать.

```hcl
path "transit/encrypt/app-key" {
  capabilities = ["update"]
}

# Явный запрет расшифровки.
path "transit/decrypt/app-key" {
  capabilities = ["deny"]
}
```

Для сервиса, который только расшифровывает данные, поменяйте возможности местами. Подробнее — в разделе [«Механизм секретов Transit»](../../user/secrets-engines/transit/).

## Управление собственным токеном

Минимальная политика, похожая на встроенную политику `default`: токен может просматривать, продлевать и отзывать себя, проверять свои возможности и пользоваться cubbyhole и обертыванием ответов.

```hcl
path "auth/token/lookup-self" {
  capabilities = ["read"]
}

path "auth/token/renew-self" {
  capabilities = ["update"]
}

path "auth/token/revoke-self" {
  capabilities = ["update"]
}

path "sys/capabilities-self" {
  capabilities = ["update"]
}

path "sys/leases/renew" {
  capabilities = ["update"]
}

path "sys/leases/lookup" {
  capabilities = ["update"]
}

path "cubbyhole/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "sys/wrapping/wrap" {
  capabilities = ["update"]
}

path "sys/wrapping/lookup" {
  capabilities = ["update"]
}

path "sys/wrapping/unwrap" {
  capabilities = ["update"]
}
```

Актуальное содержимое встроенной политики `default` можно посмотреть командой `d8 stronghold read sys/policy/default`.

## Администратор безопасности

Администратор безопасности управляет политиками, методами аутентификации, аудитом и идентификационными данными, но не имеет доступа к содержимому секретов. Возможность `deny` имеет наивысший приоритет, поэтому запрет сработает, даже если другая политика токена разрешает доступ.

```hcl
# Управление политиками.
path "sys/policies/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Управление методами аутентификации.
path "sys/auth" {
  capabilities = ["read"]
}

path "sys/auth/*" {
  capabilities = ["create", "read", "update", "delete", "sudo"]
}

path "auth/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Управление аудит-устройствами.
path "sys/audit" {
  capabilities = ["read", "sudo"]
}

path "sys/audit/*" {
  capabilities = ["create", "read", "update", "delete", "sudo"]
}

# Управление идентификационными данными.
path "identity/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Просмотр точек монтирования.
path "sys/mounts" {
  capabilities = ["read"]
}

# Запрет доступа к содержимому секретов.
path "secret/data/*" {
  capabilities = ["deny"]
}

path "kv/*" {
  capabilities = ["deny"]
}

path "database/creds/*" {
  capabilities = ["deny"]
}
```

Перечислите в блоках `deny` все точки монтирования с секретами, которые используются в вашей инсталляции.

{{< alert level="warning" >}}
Администратор, который может изменять политики и методы аутентификации, технически способен выдать себе доступ к секретам. Контролируйте такие изменения через [журнал аудита](../../admin/audit/overview/).
{{< /alert >}}
