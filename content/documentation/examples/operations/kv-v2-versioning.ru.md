---
title: "Версии секретов в KV v2"
linkTitle: "Версии секретов KV v2"
description: "Жизненный цикл секрета в KV версии 2: новые версии, частичное обновление, чтение и откат версии, ограничение числа версий, check-and-set, удаление, восстановление и уничтожение."
weight: 120
params:
  relatedLinks:
    - title: "Механизм секретов KV версии 2"
      url: ../../../user/secrets-engines/kv/kv-v2/
    - title: "Политики"
      url: ../../../concepts/policy/
---

KV версии 2 хранит несколько версий каждого секрета. Это позволяет откатить ошибочное изменение, защититься от одновременной перезаписи и управлять тем, сколько истории хранить.

## Цель

Пройти полный жизненный цикл секрета: создать несколько версий, прочитать и откатить версию, ограничить хранение истории, включить check-and-set, удалить и восстановить версию и уничтожить её безвозвратно.

## Предварительные требования

- Токен Stronghold с правами на путь `secret/*`.
- Механизм секретов KV версии 2, смонтированный в `secret/` (в режиме разработки он включён по умолчанию).

## Шаг 1. Создайте версии секрета

1. Каждая запись создаёт новую версию:

   ```bash
   d8 stronghold kv put -mount=secret app/config username=app password=v1-pass
   d8 stronghold kv put -mount=secret app/config username=app password=v2-pass
   ```

1. Команда `kv patch` меняет только указанные поля и сохраняет остальные:

   ```bash
   d8 stronghold kv patch -mount=secret app/config password=v3-pass
   ```

   Поле `username` остаётся прежним, `password` получает значение `v3-pass`. Результат — версия 3.

## Шаг 2. Прочитайте нужную версию

```bash
d8 stronghold kv get -mount=secret app/config
d8 stronghold kv get -mount=secret -version=1 app/config
d8 stronghold kv metadata get -mount=secret app/config
```

Первая команда возвращает текущую версию, вторая — версию 1. Третья показывает метаданные: текущую версию, время создания каждой версии и признаки удаления.

## Шаг 3. Откатите версию

```bash
d8 stronghold kv rollback -mount=secret -version=1 app/config
```

Откат не меняет историю: содержимое версии 1 записывается как новая версия 4.

## Шаг 4. Ограничьте историю

```bash
d8 stronghold kv metadata put -mount=secret \
  -max-versions=5 \
  -delete-version-after=720h \
  app/config
```

- `max-versions` — сколько версий хранить. Самая старая версия сверх лимита удаляется автоматически.
- `delete-version-after` — через какое время после создания версия помечается удалённой.

## Шаг 5. Включите check-and-set

Check-and-set защищает от перезаписи, когда два клиента меняют секрет одновременно: запись проходит, только если клиент указал текущую версию.

1. Потребуйте указывать версию при каждой записи. Команда `kv metadata patch` меняет только указанные поля, а `kv metadata put` перезаписывает все метаданные значениями по умолчанию, поэтому здесь нужен `patch`:

   ```bash
   d8 stronghold kv metadata patch -mount=secret -cas-required=true app/config
   ```

1. Запись без версии теперь отклоняется:

   ```bash
   # запись без cas
   d8 stronghold kv put -mount=secret app/config password=without-cas
   ```

1. Запись с текущей версией проходит:

   ```bash
   d8 stronghold kv put -mount=secret -cas=4 app/config username=app password=v5-pass
   ```

   Если за это время секрет изменил другой клиент, версия не совпадёт, и запись будет отклонена.

## Шаг 6. Удалите, восстановите и уничтожьте версию

1. Мягкое удаление помечает последнюю версию удалённой, данные остаются в хранилище:

   ```bash
   d8 stronghold kv delete -mount=secret app/config
   ```

1. Восстановите версию:

   ```bash
   d8 stronghold kv undelete -mount=secret -versions=5 app/config
   ```

1. Уничтожение удаляет данные версии безвозвратно:

   ```bash
   d8 stronghold kv destroy -mount=secret -versions=1,2 app/config
   ```

## Проверка

```bash
d8 stronghold kv metadata get -mount=secret app/config
```

В выводе версии 1 и 2 помечены как уничтоженные (`destroyed: true`), версия 5 восстановлена и является текущей, `max_versions` равен `5`, `cas_required` равен `true`.

## Очистка

Удалите секрет вместе со всеми версиями и метаданными:

```bash
d8 stronghold kv metadata delete -mount=secret app/config
```
