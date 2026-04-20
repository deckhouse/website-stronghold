---
title: "KV v2"
linkTitle: "KV v2"
weight: 30
description: "Работа с механизмом секретов KV v2 в Deckhouse Stronghold."
---

Механизм секретов `KV v2` используется для хранения произвольных секретов в пределах хранилища **Deckhouse Stronghold** и поддерживает версионирование.

По сравнению с `KV v1`, `KV v2` позволяет:

- хранить несколько версий одного секрета;
- получать метаданные по версиям;
- выполнять мягкое удаление;
- восстанавливать удалённые версии;
- окончательно уничтожать версии;
- управлять параметрами хранения через метаданные.

## Что важно знать

- Имена ключей всегда должны быть строками.
- При записи нестроковых значений через CLI они будут преобразованы в строки.
- `KV v2` различает операции `create`, `update` и `patch` в ACL-политиках.
- Для защиты от случайной перезаписи поддерживается механизм `check-and-set` (`CAS`).

## Как включить KV v2

Вы можете включить `KV v2` одним из двух способов.

### Вариант 1

```bash
d8 stronghold secrets enable -version=2 kv
```

### Вариант 2

```bash
d8 stronghold secrets enable kv-v2
```

## Обновление с KV v1 на KV v2

Существующее хранилище `KV v1` можно обновить до `KV v2`.

> Предупреждение  
> Во время миграции хранилище будет недоступно. Это может занять заметное время, поэтому обновление нужно планировать заранее.

Команда обновления:

```bash
d8 stronghold kv enable-versioning secret/
```

После обновления:

- старые пути доступа перестанут работать;
- потребуется обновить ACL-политики;
- приложениям и пользователям нужно будет использовать новые пути `KV v2`.

## ACL для KV v2

`KV v2` использует API с префиксами, отличающимися от `KV v1`.

Например, путь чтения и записи теперь использует префикс `data/`:

### Было в KV v1

```hcl
path "secret/dev/team-1/*" {
  capabilities = ["create", "update", "read"]
}
```

### Стало в KV v2

```hcl
path "secret/data/dev/team-1/*" {
  capabilities = ["create", "update", "read"]
}
```

Для дополнительных операций используются отдельные пути:

- удаление версий:

  ```hcl
  path "secret/delete/dev/team-1/*" {
    capabilities = ["update"]
  }
  ```

- восстановление версий:

  ```hcl
  path "secret/undelete/dev/team-1/*" {
    capabilities = ["update"]
  }
  ```

- уничтожение версий:

  ```hcl
  path "secret/destroy/dev/team-1/*" {
    capabilities = ["update"]
  }
  ```

- список ключей:

  ```hcl
  path "secret/metadata/dev/team-1/*" {
    capabilities = ["list"]
  }
  ```

- чтение метаданных:

  ```hcl
  path "secret/metadata/dev/team-1/*" {
    capabilities = ["read"]
  }
  ```

- полное удаление всех версий и метаданных:

  ```hcl
  path "secret/metadata/dev/team-1/*" {
    capabilities = ["delete"]
  }
  ```

> Примечание  
> Параметры `allowed_parameters`, `denied_parameters` и `required_parameters` не поддерживаются для политик, используемых с `KV v2`.

## Базовые операции

Рекомендуется использовать синтаксис с `-mount`, чтобы не путать логический путь секрета с внутренним API-путём `data/...`.

### Запись секрета

```bash
d8 stronghold kv put -mount=secret my-secret foo=a bar=b
```

Пример вывода:

```text
Key              Value
---              -----
created_time     2024-06-19T17:20:22.985303Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          1
```

### Чтение секрета

```bash
d8 stronghold kv get -mount=secret my-secret
```

Пример вывода:

```text
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

## Новые версии и CAS

При повторной записи Stronghold создаёт новую версию секрета.

Пример:

```bash
d8 stronghold kv put -mount=secret -cas=1 my-secret foo=aa bar=bb
```

Пример вывода:

```text
Key              Value
---              -----
created_time     2024-06-19T17:22:23.369372Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          2
```

Параметр `-cas` означает `check-and-set`:

- если он не указан, запись разрешена;
- если он указан, его значение должно совпадать с текущей версией секрета;
- значение `0` разрешает запись только в случае, если ключ ещё не существует.

## Patch

`KV v2` поддерживает частичное обновление секрета через `patch`.

Пример:

```bash
d8 stronghold kv patch -mount=secret -cas=2 my-secret bar=bbb
```

Stronghold сначала попытается выполнить HTTP `PATCH`, для которого требуется ACL-возможность `patch`. Если токен не имеет этой возможности, CLI может использовать сценарий `read` + `update`.

### Выбор метода patch

Можно явно указать метод:

#### PATCH

```bash
d8 stronghold kv patch -mount=secret -method=patch -cas=2 my-secret bar=bbb
```

#### Read/Write

```bash
d8 stronghold kv patch -mount=secret -method=rw my-secret bar=bbb
```

После этого чтение вернёт только частично изменённые значения.

## Чтение старых версий

Получить старую версию можно с помощью флага `-version`:

```bash
d8 stronghold kv get -mount=secret -version=1 my-secret
```

Это позволяет читать историю версий секрета, пока версия не удалена или не уничтожена.

## Удаление, восстановление и уничтожение

### Мягкое удаление

Команда `delete` выполняет мягкое удаление: версия помечается как удалённая, но её ещё можно восстановить.

```bash
d8 stronghold kv delete -mount=secret my-secret
```

### Восстановление версии

```bash
d8 stronghold kv undelete -mount=secret -versions=2 my-secret
```

### Уничтожение версии

Для безвозвратного удаления используйте:

```bash
d8 stronghold kv destroy -mount=secret -versions=2 my-secret
```

После `destroy` данные версии удаляются окончательно.

## Метаданные

`KV v2` поддерживает отдельные операции для метаданных.

### Просмотр метаданных

```bash
d8 stronghold kv metadata get -mount=secret my-secret
```

Эта команда показывает:

- текущую версию;
- число версий;
- `max_versions`;
- `delete_version_after`;
- `custom_metadata`;
- состояние каждой версии.

### Изменение параметров хранения

```bash
d8 stronghold kv metadata put -mount=secret -max-versions 2 -delete-version-after="3h25m19s" my-secret
```

Параметр `delete-version_after` применяется только к новым версиям, а `max_versions` вступает в силу при следующей записи.

### Пользовательские метаданные

Полная перезапись:

```bash
d8 stronghold kv metadata put -mount=secret -custom-metadata=foo=abc -custom-metadata=bar=123 my-secret
```

Частичное изменение:

```bash
d8 stronghold kv metadata patch -mount=secret -custom-metadata=foo=def my-secret
```

### Полное удаление метаданных и всех версий

```bash
d8 stronghold kv metadata delete -mount=secret my-secret
```

## Практические рекомендации

- Используйте `KV v2`, если вам нужны история изменений и защита от случайной перезаписи.
- Для приложений и пользователей заранее обновляйте ACL при переходе с `KV v1`.
- Используйте `CAS` там, где важно не потерять изменения при конкурентной записи.
- Различайте мягкое удаление (`delete`) и окончательное уничтожение (`destroy`).
- Документируйте `custom_metadata`, если секреты нужно описывать или классифицировать.

## Что дальше

Если вам нужен более простой сценарий без версионирования, используйте [KV v1](./kv-v1/).  
Если нужно общее введение по механизму `KV`, перейдите в раздел [Обзор](../kv/).
