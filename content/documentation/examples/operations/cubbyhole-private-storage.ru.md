---
title: "Приватное хранилище токена (cubbyhole)"
linkTitle: "Cubbyhole"
description: "Механизм секретов Cubbyhole: у каждого токена своё приватное хранилище, недоступное другим токенам, в том числе root; данные удаляются вместе с токеном."
weight: 140
params:
  relatedLinks:
    - title: "Механизм секретов Cubbyhole"
      url: ../../../user/secrets-engines/cubbyhole/
    - title: "Обертывание ответов"
      url: ../../../concepts/response-wrapping/
---

Cubbyhole — механизм секретов, который включён всегда и монтируется в `cubbyhole/`. У каждого токена в нём собственное изолированное хранилище: другие токены, включая root, его не видят. Когда токен истекает или отзывается, его хранилище удаляется. На этом построено [обертывание ответов](../../../concepts/response-wrapping/).

## Цель

Убедиться, что данные в cubbyhole доступны только создавшему их токену и исчезают вместе с ним.

## Предварительные требования

- Токен Stronghold с правом создавать токены.
- Политика `default` у создаваемых токенов: она разрешает работу с `cubbyhole/*`.

## Шаг 1. Запишите данные своим токеном

```bash
d8 stronghold write cubbyhole/notes text="личная заметка"
d8 stronghold read cubbyhole/notes
```

## Шаг 2. Создайте второй токен и проверьте изоляцию

1. Создайте токен с политикой `default`:

   ```bash
   TOKEN2=$(d8 stronghold token create -policy=default -ttl=10m -field=token)
   ```

1. Второй токен не видит заметку первого, хотя путь тот же:

   ```bash
   # чтение другим токеном
   STRONGHOLD_TOKEN="$TOKEN2" d8 stronghold read cubbyhole/notes
   ```

   Команда возвращает `No value found at cubbyhole/notes`.

1. У второго токена своё хранилище:

   ```bash
   STRONGHOLD_TOKEN="$TOKEN2" d8 stronghold write cubbyhole/notes text="заметка второго токена"
   STRONGHOLD_TOKEN="$TOKEN2" d8 stronghold read cubbyhole/notes
   d8 stronghold read cubbyhole/notes
   ```

   Первая команда чтения возвращает заметку второго токена, вторая — заметку первого.

## Шаг 3. Проверьте удаление данных вместе с токеном

Отзовите второй токен. Его хранилище удаляется вместе с ним, и вернуть данные нельзя:

```bash
d8 stronghold token revoke "$TOKEN2"
```

## Проверка

Заметка первого токена осталась на месте:

```bash
d8 stronghold read -field=text cubbyhole/notes
```

## Применение

Cubbyhole подходит для временного хранения данных, которые не должны быть доступны никому, кроме одного токена: промежуточных результатов скриптов, одноразовых значений при передаче секретов. Для долгого хранения данных и для общего доступа используйте [KV](../kv-v2-versioning/).

## Очистка

```bash
d8 stronghold delete cubbyhole/notes
```
