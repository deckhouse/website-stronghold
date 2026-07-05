---
title: "Создание тестового секрета"
linkTitle: "Создание тестового секрета"
description: "Создание тестового секрета в Deckhouse Stronghold"
weight: 30
---

В этом разделе показано, как вы можете создать тестовый секрет в Deckhouse Stronghold, проверить его содержимое и при необходимости изменить значение.

В примере используются:

- [механизм секретов `kv`](../../secrets-engines/kv/overview/);
- тестовый путь `secret/my-first-secret`;
- тестовые ключи `username` и `password`.

Если в вашей инсталляции используется другой путь монтирования, замените `secret` на путь, предоставленный администратором.

{{< alert level="warning" >}}
Используйте в примере только тестовые значения. Не сохраняйте реальные пароли, токены и ключи в тестовых секретах.
{{< /alert >}}

Чтобы создать тестовый секрет, выполните следующие действия:

1. Создайте секрет, записав тестовые значения по выбранному пути:

   ```shell
   d8 stronghold kv put secret/my-first-secret username=demo password=secret123
   ```

   Если операция завершилась успешно, Stronghold подтвердит запись секрета.

1. Проверьте сохранённые значения:

   ```shell
   d8 stronghold kv get secret/my-first-secret
   ```

   Пример вывода:

   ```text
   ====== Data ======
   Key         Value
   ---         -----
   password    secret123
   username    demo
   ```

1. Измените значение секрета, повторно записав его по тому же пути:

   ```shell
   d8 stronghold kv put secret/my-first-secret username=demo password=new-secret
   ```

1. Убедитесь, что значение обновилось:

   ```shell
   d8 stronghold kv get secret/my-first-secret
   ```

{{< alert level="info" >}}
После проверки замените тестовый путь и значения на параметры, которые используются в вашем проекте.
{{< /alert >}}
