---
title: "KV v2"
description: "Сведения о разделе \"KV v2\" в Deckhouse Stronghold."
weight: 30
---

Механизм секретов `kv` версии 2 предназначен для хранения произвольных секретов в хранилище Stronghold.
В отличие от `kv` версии 1, он поддерживает версионность данных и частичное обновление секретов.

Имена ключей должны быть строками.
Если записывать нестроковые значения напрямую через CLI, Stronghold преобразует их в строки.
Чтобы сохранить нестроковые значения, передавайте пары «ключ–значение» из JSON-файла или используйте HTTP API.

Механизм секретов `kv` учитывает различие между операциями `create` и `update` в ACL-политиках.
Также поддерживается операция `patch` для частичного обновления секрета.
Операция `update`, напротив, полностью перезаписывает значение.

## Как включить

Большинство механизмов секретов нужно предварительно настроить.
Обычно это делает оператор или система управления конфигурацией, например Terraform.

Чтобы включить хранилище `kv` версии 2, выполните одну из команд:

```shell
d8 stronghold secrets enable -version=2 kv
```

```shell
d8 stronghold secrets enable kv-v2
```

## Обновление с версии 1 до версии 2

Существующее хранилище `kv` версии 1 можно обновить до `kv` версии 2 с помощью CLI или API.
Во время миграции хранилище будет недоступно.
Обновление может занять продолжительное время, поэтому запланируйте его заранее.

После обновления до версии 2 прежние пути доступа к данным перестанут работать.
Обновите ACL-политики, чтобы восстановить доступ.
Также обновите пути в приложениях и пользовательских сценариях, которые работают с данными `kv`.

Чтобы включить версионность для существующего хранилища, выполните команду:

```console
$ d8 stronghold kv enable-versioning secret/
Success! Tuned the secrets engine at: secret/
```

## Правила ACL

Хранилище `kv` версии 2 использует API с префиксами, которые отличаются от API версии 1.
Перед обновлением с `kv` версии 1 измените ACL-политики.
Разные пути API версии 2 можно защищать разными ACL-правилами.

Пути для чтения и записи используют префикс `data/`.
Например, такую политику для `kv` версии 1:

```text
path "secret/dev/team-1/*" {
  capabilities = ["create", "update", "read"]
}
```

замените на политику для `kv` версии 2:

```text
path "secret/data/dev/team-1/*" {
  capabilities = ["create", "update", "read"]
}
```

Для `kv` версии 2 доступны разные уровни удаления данных.

Чтобы разрешить удаление последней версии ключа, используйте такую политику:

```text
path "secret/data/dev/team-1/*" {
  capabilities = ["delete"]
}
```

Чтобы разрешить удаление произвольной версии ключа, используйте такую политику:

```text
path "secret/delete/dev/team-1/*" {
  capabilities = ["update"]
}
```

Чтобы разрешить восстановление удалённых версий, используйте такую политику:

```text
path "secret/undelete/dev/team-1/*" {
  capabilities = ["update"]
}
```

Чтобы разрешить окончательное уничтожение значений без возможности восстановления, используйте такую политику:

```text
path "secret/destroy/dev/team-1/*" {
  capabilities = ["update"]
}
```

Чтобы разрешить получение списка ключей, используйте такую политику:

```text
path "secret/metadata/dev/team-1/*" {
  capabilities = ["list"]
}
```

Чтобы разрешить просмотр метаданных ключей, используйте такую политику:

```text
path "secret/metadata/dev/team-1/*" {
  capabilities = ["read"]
}
```

Чтобы разрешить полное удаление всех версий и метаданных ключа, используйте такую политику:

```text
path "secret/metadata/dev/team-1/*" {
  capabilities = ["delete"]
}
```

Поля `allowed_parameters`, `denied_parameters` и `required_parameters` не поддерживаются в политиках для хранилища `kv` версии 2.

## Использование

После включения механизма секретов и получения токена Stronghold с нужными правами можно работать с секретами.

Для `kv` версии 2 по-прежнему можно использовать синтаксис в стиле `kv` версии 1, например путь `secret/foo`.
Однако предпочтительнее использовать флаг `-mount=secret`, чтобы не путать логический путь секрета с фактическим API-путём.
Фактический путь в этом случае — `secret/data/foo`.

### Запись и чтение произвольных данных

Запишите секрет:

```console
$ d8 stronghold kv put -mount=secret my-secret foo=a bar=b
Key              Value
---              -----
created_time     2024-06-19T17:20:22.985303Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          1
```

Прочитайте секрет:

```console
$ d8 stronghold kv get -mount=secret my-secret
====== Metadata ======
Key              Value
---              -----
created_time     2024-06-19T17:20:22.985303Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          1
====== Data ======
Key         Value
---         -----
foo         a
bar         b
```

Запишите новую версию секрета:

```console
$ d8 stronghold kv put -mount=secret -cas=1 my-secret foo=aa bar=bb
Key              Value
---              -----
created_time     2024-06-19T17:22:23.369372Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          2
```

Флаг `-cas` включает проверку Check-and-Set.
Если флаг не указан, запись выполняется без проверки.
Если флаг указан, его значение должно совпадать с текущей версией секрета.
Значение `0` разрешает запись только в том случае, если ключ ещё не существует.

Учтите, что удаление версии не удаляет информацию о версиях из хранилища.
Поэтому при записи в секрет, у которого есть удалённые версии, значение `cas` должно соответствовать текущей версии секрета.

По умолчанию чтение возвращает последнюю версию:

```console
$ d8 stronghold kv get -mount=secret my-secret
====== Metadata ======
Key              Value
---              -----
created_time     2024-06-19T17:22:23.369372Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          2
====== Data ======
Key         Value
---         -----
foo         aa
bar         bb
```

Для частичного обновления секрета используйте команду `d8 stronghold kv patch`:

```console
$ d8 stronghold kv patch -mount=secret -cas=2 my-secret bar=bbb
Key              Value
---              -----
created_time     2024-06-19T17:23:49.199802Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          3
```

Команда сначала пытается выполнить HTTP-запрос `PATCH`.
Для этого токен должен иметь ACL-возможность `patch`.
Если такой возможности нет, команда выполняет чтение, локальное обновление и последующую запись.
В этом случае нужны ACL-возможности `read` и `update`.

Флаг `-cas` можно использовать и здесь.
Для прямого `PATCH` он применяется сразу.
Для сценария с чтением и повторной записью команда использует значение `version`, полученное при чтении, чтобы выполнить проверку `cas` при записи.

Команда `d8 stronghold kv patch` также поддерживает флаг `-method`.
Он определяет способ обновления: `patch` или `rw`.

Обновите секрет через `patch`:

```console
$ d8 stronghold kv patch -mount=secret -method=patch -cas=2 my-secret bar=bbb
Key              Value
---              -----
created_time     2024-06-19T17:23:49.199802Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          3
```

Обновите секрет через `rw`, то есть через чтение и повторную запись:

```console
$ d8 stronghold kv patch -mount=secret -method=rw my-secret bar=bbb
Key              Value
---              -----
created_time     2024-06-19T17:23:49.199802Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          3
```

Прочитайте обновлённый секрет:

```console
$ d8 stronghold kv get -mount=secret my-secret
====== Metadata ======
Key              Value
---              -----
created_time     2024-06-19T17:23:49.199802Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          3
====== Data ======
Key         Value
---         -----
foo         aa
bar         bbb
```

Чтобы прочитать предыдущую версию, используйте флаг `-version`:

```console
$ d8 stronghold kv get -mount=secret -version=1 my-secret
====== Metadata ======
Key              Value
---              -----
created_time     2024-06-19T17:20:22.985303Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          1
====== Data ======
Key         Value
---         -----
foo         a
bar         b
```

Также можно использовать password policy для генерации значений.

Создайте policy:

```console
$ d8 stronghold write sys/policies/password/example policy=-<<EOF
  length=20
  rule "charset" {
    charset = "abcdefghij0123456789"
    min-chars = 1
  }
  rule "charset" {
    charset = "!@#$%^&*STUVWXYZ"
    min-chars = 1
  }
EOF
```

Создайте секрет с использованием policy `example`:

```console
$ d8 stronghold kv put -mount=secret my-generated-secret \
    password=$(d8 stronghold read -field password sys/policies/password/example/generate)
```

```text
========= Secret Path =========
secret/data/my-generated-secret
======= Metadata =======
Key                Value
---                -----
created_time       2024-06-10T14:32:32.37354939Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            1
```

Прочитайте созданный секрет:

```console
$ d8 stronghold kv get -mount=secret my-generated-secret
========= Secret Path =========
secret/data/my-generated-secret
======= Metadata =======
Key                Value
---                -----
created_time       2024-06-10T14:32:32.37354939Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            1
====== Data ======
Key         Value
---         -----
password    !hh&be1e4j16dVc0ggae
```

### Удаление и уничтожение секретов

Команда `d8 stronghold kv delete` выполняет мягкое удаление.
Она помечает версию как удалённую и заполняет значение `deletion_time` в метаданных секрета.
При мягком удалении данные версии не удаляются из хранилища.
Такую версию можно восстановить с помощью команды `d8 stronghold kv undelete`.

Версия секрета удаляется окончательно в двух случаях:
- если число версий превышает значение `max-versions`;
- если используется команда `d8 stronghold kv destroy`.

Команда `destroy` удаляет данные версии без возможности восстановления.
При этом метаданные версии помечаются как уничтоженные.
Если версия удаляется из-за превышения числа версий, её метаданные тоже удаляются.

Последнюю версию ключа можно удалить с помощью команды `delete`.
Команда также поддерживает флаг `-versions` для удаления предыдущих версий:

```console
$ d8 stronghold kv delete -mount=secret my-secret
Success! Data deleted (if it existed) at: secret/data/my-secret
```

Удалённые версии можно восстановить:

```console
$ d8 stronghold kv undelete -mount=secret -versions=2 my-secret
Success! Data written to: secret/undelete/my-secret

$ d8 stronghold kv get -mount=secret my-secret
====== Metadata ======
Key              Value
---              -----
created_time     2024-06-19T17:23:21.834403Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          2
====== Data ======
Key         Value
---         -----
my-value    short-lived-s3cr3t
```

Чтобы уничтожить версию окончательно, выполните команду:

```console
$ d8 stronghold kv destroy -mount=secret -versions=2 my-secret
Success! Data written to: secret/destroy/my-secret
```

### Метаданные

Все версии и метаданные ключа можно просмотреть с помощью команды `metadata` или через API.
Если удалить ключ через `metadata delete`, все метаданные и версии этого ключа будут удалены без возможности восстановления.

Просмотрите метаданные и версии ключа:

```console
$ d8 stronghold kv metadata get -mount=secret my-secret
========== Metadata ==========
Key                     Value
---                     -----
cas_required            false
created_time            2024-06-19T17:20:22.985303Z
current_version         2
custom_metadata         <nil>
delete_version_after    0s
max_versions            0
oldest_version          0
updated_time            2024-06-19T17:22:23.369372Z
====== Version 1 ======
Key              Value
---              -----
created_time     2024-06-19T17:20:22.985303Z
deletion_time    n/a
destroyed        false
====== Version 2 ======
Key              Value
---              -----
created_time     2024-06-19T17:22:23.369372Z
deletion_time    n/a
destroyed        true
```

Настройте параметры хранения версий:

```console
$ d8 stronghold kv metadata put -mount=secret -max-versions=2 -delete-version-after=3h25m19s my-secret
Success! Data written to: secret/metadata/my-secret
```

Параметр `delete-version-after` применяется только к новым версиям.
Параметр `max-versions` применяется при следующей операции записи.

```console
$ d8 stronghold kv put -mount=secret my-secret my-value=newer-s3cr3t
Key              Value
---              -----
created_time     2024-06-19T17:31:16.662563Z
custom_metadata  <nil>
deletion_time    2024-06-19T20:56:35.662563Z
destroyed        false
version          4
```

Если число версий превышает `max-versions`, самые старые версии уничтожаются:

```console
$ d8 stronghold kv metadata get -mount=secret my-secret
========== Metadata ==========
Key                     Value
---                     -----
cas_required            false
created_time            2024-06-19T17:20:22.985303Z
current_version         4
custom_metadata         <nil>
delete_version_after    3h25m19s
max_versions            2
oldest_version          3
updated_time            2024-06-19T17:31:16.662563Z
====== Version 3 ======
Key              Value
---              -----
created_time     2024-06-19T17:23:21.834403Z
deletion_time    n/a
destroyed        true
====== Version 4 ======
Key              Value
---              -----
created_time     2024-06-19T17:31:16.662563Z
deletion_time    2024-06-19T20:56:35.662563Z
destroyed        false
```

Метаданные секрета могут включать пользовательские метаданные в виде пар «ключ–значение».
Флаг `-custom-metadata` можно указывать несколько раз.

Команда `d8 stronghold kv metadata put` полностью перезаписывает значение `custom_metadata`:

```console
$ d8 stronghold kv metadata put -mount=secret -custom-metadata=foo=abc -custom-metadata=bar=123 my-secret
Success! Data written to: secret/metadata/my-secret

$ d8 stronghold kv get -mount=secret my-secret
====== Metadata ======
Key              Value
---              -----
created_time     2024-06-19T17:22:23.369372Z
custom_metadata  map[bar:123 foo:abc]
deletion_time    n/a
destroyed        false
version          2
====== Data ======
Key         Value
---         -----
foo         aa
bar         bb
```

Команда `d8 stronghold kv metadata patch` частично обновляет значение `custom_metadata`.
Например, следующая команда обновит поле `foo`, но оставит поле `bar` без изменений:

```console
$ d8 stronghold kv metadata patch -mount=secret -custom-metadata=foo=def my-secret
Success! Data written to: secret/metadata/my-secret
```

```console
$ d8 stronghold kv get -mount=secret my-secret
====== Metadata ======
Key              Value
---              -----
created_time     2024-06-19T17:22:23.369372Z
custom_metadata  map[bar:123 foo:def]
deletion_time    n/a
destroyed        false
version          2
====== Data ======
Key         Value
---         -----
foo         aa
bar         bb
```

Чтобы удалить все метаданные и все версии ключа, выполните команду:

```console
$ d8 stronghold kv metadata delete -mount=secret my-secret
Success! Data deleted (if it existed) at: secret/metadata/my-secret
```
