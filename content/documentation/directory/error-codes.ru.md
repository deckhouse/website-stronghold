---
title: "Коды ошибок"
linkTitle: "Коды ошибок"
description: "Типовые ошибки и диагностические сообщения в Deckhouse Stronghold"
weight: 60
---

На этой странице собраны типовые ошибки и диагностические сообщения, которые встречаются в Deckhouse Stronghold и связанных с ним инструментах.

Это не полный каталог всех ошибок продукта. Точный текст ошибки зависит от механизма секретов, используемого API, CLI-команды и внешней системы, с которой работает Stronghold.

## Какие ошибки могут встречаться

В работе с Deckhouse Stronghold обычно встречаются несколько типов ошибок:

- ошибки HTTP API;
- ошибки CLI;
- ошибки конкретного механизма секретов;
- ошибки внешних компонентов, например OpenSSH или Kubernetes API.

Если вы разбираете проблему, сначала определите, на каком уровне она возникла:

- в Stronghold API;
- в CLI;
- в механизме секретов;
- во внешнем сервисе или клиенте.

## Типовые примеры

### `401 Unauthorized`

Такой ответ можно получить при работе с Kubernetes API, если токен Kubernetes уже отозван после окончания аренды Stronghold.

Пример:

```text
{
  "kind": "Status",
  "apiVersion": "v1",
  "metadata": {},
  "status": "Failure",
  "message": "Unauthorized",
  "reason": "Unauthorized",
  "code": 401
}
```

Что это обычно значит:

- срок действия токена истёк;
- токен был отозван;
- токен используется вне допустимого сценария.

### `Certificate invalid: name is not a listed principal`

Эта ошибка встречается при работе с SSH-сертификатами, если имя пользователя не входит в список principal, разрешённых для сертификата.

Пример сообщения:

```text
Certificate invalid: name is not a listed principal
```

Что проверить:

- значение `default_user` в роли;
- параметр `valid_principals` при подписи ключа.

### `no separate private key for certificate`

Эта ошибка описана в разделе SSH как проблема некоторых версий OpenSSH.

Пример:

```text
no separate private key for certificate
```

Что это значит:

- проблема связана не с Stronghold, а с конкретной версией OpenSSH;
- ошибка появилась в OpenSSH 7.2 и была исправлена в OpenSSH 7.5.

### `certificate signature algorithm ssh-rsa: signature algorithm not supported`

Эта ошибка также встречается в SSH-сценариях и относится к совместимости алгоритмов подписи в OpenSSH.

Пример:

```text
userauth_pubkey: certificate signature algorithm ssh-rsa: signature algorithm not supported [preauth]
```

Что это обычно значит:

- используемая версия OpenSSH не поддерживает нужный алгоритм подписи;
- может потребоваться дополнительная настройка `CASignatureAlgorithms`.

## Как искать причину ошибки

Если ошибка возникает при работе с конкретным механизмом секретов, начните с его страницы документации:

- для SSH — [SSH](../user/secrets-engines/ssh/)
- для Kubernetes — [Kubernetes](../user/secrets-engines/kubernetes/)
- для Transit — [Transit](../user/secrets-engines/transit/)
- для LDAP — [LDAP](../user/secrets-engines/ldap/)
- для баз данных — [Базы данных](../user/secrets-engines/database/)

Если ошибка относится к внешнему компоненту, например OpenSSH или Kubernetes API, проверяйте также журналы и настройки этой системы.

## Что делать дальше

Если ошибка неочевидна:

1. зафиксируйте точный текст сообщения;
2. определите, в каком компоненте возникла ошибка;
3. проверьте страницу соответствующего механизма секретов;
4. дополнительно проверьте журналы внешней системы, если Stronghold работает через неё.

## Поддержка

Если в документации нет нужной ошибки или у вас возникли вопросы по работе с Deckhouse Stronghold, используйте доступные каналы поддержки.

> Важно  
> Если вы используете Enterprise-редакцию, обращайтесь в техническую поддержку по адресу [support@deckhouse.ru](mailto:support@deckhouse.ru).

Также вы можете задать вопрос в [Telegram-канале Deckhouse](https://t.me/deckhouse_ru).
