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

   {{< tabs name="stronghold_cmd_58984" >}}
   {{% tab name="Stronghold в DKP" %}}

   ```shell
   d8 stronghold kv put secret/my-first-secret username=demo password=secret123
   ```

   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}

   ```shell
   stronghold kv put secret/my-first-secret username=demo password=secret123
   ```

   {{% /tab %}}
   {{< /tabs >}}

   Если операция завершилась успешно, Stronghold подтвердит запись секрета.

1. Проверьте сохранённые значения:

   {{< tabs name="stronghold_cmd_55073" >}}
   {{% tab name="Stronghold в DKP" %}}

   ```shell
   d8 stronghold kv get secret/my-first-secret
   ```

   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}

   ```shell
   stronghold kv get secret/my-first-secret
   ```

   {{% /tab %}}
   {{< /tabs >}}

   Пример вывода:

   ```text
   ====== Data ======
   Key         Value
   ---         -----
   password    secret123
   username    demo
   ```

1. Измените значение секрета, повторно записав его по тому же пути:

   {{< tabs name="stronghold_cmd_85471" >}}
   {{% tab name="Stronghold в DKP" %}}

   ```shell
   d8 stronghold kv put secret/my-first-secret username=demo password=new-secret
   ```

   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}

   ```shell
   stronghold kv put secret/my-first-secret username=demo password=new-secret
   ```

   {{% /tab %}}
   {{< /tabs >}}

1. Убедитесь, что значение обновилось:

   {{< tabs name="stronghold_cmd_55073" >}}
   {{% tab name="Stronghold в DKP" %}}

   ```shell
   d8 stronghold kv get secret/my-first-secret
   ```

   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}

   ```shell
   stronghold kv get secret/my-first-secret
   ```

   {{% /tab %}}
   {{< /tabs >}}

{{< alert level="info" >}}
После проверки замените тестовый путь и значения на параметры, которые используются в вашем проекте.
{{< /alert >}}
