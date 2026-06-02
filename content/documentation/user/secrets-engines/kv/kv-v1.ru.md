---
title: "KV v1"
description: "Сведения о разделе \"KV v1\" в Deckhouse Stronghold."
weight: 20
---

Механизм секретов `kv` версии 1 предназначен для хранения произвольных секретов в хранилище Stronghold.
При записи по ключу предыдущее значение заменяется новым.

Имена ключей должны быть строками.
Если записывать нестроковые значения напрямую через CLI, Stronghold преобразует их в строки.
Чтобы сохранить нестроковые значения, передавайте пары «ключ–значение» из JSON-файла или используйте HTTP API.

Механизм секретов `kv` учитывает различие между операциями `create` и `update` в ACL-политиках.

{{< alert level="warning" >}}
Пути и имена ключей не обфусцируются и не шифруются.
Шифруются только значения ключей.
Не храните конфиденциальные данные в пути секрета или в имени ключа.
{{< /alert >}}

## Как включить

Чтобы включить хранилище `kv` версии 1, выполните команду:

```shell
d8 stronghold secrets enable -version=1 kv
```

## Использование

Механизм секретов `kv` позволяет записывать ключи с произвольными значениями.
Для работы требуется токен с соответствующими правами.

Выполните следующие действия:

1. Запишите произвольные данные:

   ```console
   $ d8 stronghold kv put kv/my-secret my-value=s3cr3t
   Success! Data written to: kv/my-secret
   ```

1. Прочитайте данные:

   ```console
   $ d8 stronghold kv get kv/my-secret
   Key                 Value
   ---                 -----
   my-value            s3cr3t
   ```

1. Получите список ключей:

   ```console
   $ d8 stronghold kv list kv/
   Keys
   ----
   my-secret
   ```

1. Удалите ключ:

   ```console
   $ d8 stronghold kv delete kv/my-secret
   Success! Data deleted (if it existed) at: kv/my-secret
   ```

Также можно использовать механизм password policy для генерации значений.

1. Создайте password policy:

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

1. Сгенерируйте пароль по policy `example`:

   ```console
   $ d8 stronghold kv put kv/my-generated-secret \
       password=$(d8 stronghold read -field password sys/policies/password/example/generate)
   ```

1. Прочитайте сгенерированное значение секрета:

   ```console
   $ d8 stronghold kv get kv/my-generated-secret
   ====== Data ======
   Key         Value
   ---         -----
   password    ^dajd609Xf8Zhac$dW24
   ```

## Время жизни ключей

В отличие от других механизмов секретов, `kv` не применяет TTL для автоматического истечения срока действия данных.
Значение `lease_duration` здесь носит информационный характер и показывает, как часто рекомендуется проверять обновление значения.

Если для ключа задан параметр `ttl`, механизм секретов `kv` использует его как продолжительность аренды:

```console
$ d8 stronghold kv put kv/my-secret ttl=5s my-value=s3cr3t
Success! Data written to: kv/my-secret
```

Даже если задан `ttl`, механизм секретов никогда не удаляет данные автоматически.
Параметр `ttl` имеет только рекомендательный характер.

При чтении секрета со значением `ttl` и сам ключ `ttl`, и интервал обновления будут отражать это значение:

```console
$ d8 stronghold kv get kv/my-secret
Key                 Value
---                 -----
my-value            s3cr3t
ttl                 5s
```

```console
$ curl -X 'GET' \
    'https://stronghold.example.com/v1/kv/my-secret' \
    -H 'X-Vault-Token: ***'
{
  "request_id": "3879d849-cb78-725a-c2eb-3ba9dfe8a1d3",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 5,
  "data": {
    "my-value": "s3cr3t",
    "ttl": "5s"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```
