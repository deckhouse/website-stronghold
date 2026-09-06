---
title: "Multifactor"
description: "Настройка многофакторной аутентификации Stronghold через сервис Multifactor"
weight: 10
---

Stronghold поддерживает проверку дополнительного фактора аутентификации с помощью сервиса [Multifactor](https://multifactor.ru/docs/intro/).
После успешной аутентификации по основному методу Stronghold отправляет запрос в Multifactor и ожидает подтверждения от пользователя.

В зависимости от настроек пользователя в Multifactor запрос можно подтвердить одним из следующих способов:

- с помощью push-уведомления в мобильном приложении Multifactor;
- в Telegram;
- по телефону — для подтверждения примите вызов и нажмите `#`.

{{< alert level="info" >}}
Stronghold не принимает одноразовые коды (OTP) при использовании Multifactor MFA.

Пользователь подтверждает или отклоняет запрос на доступ в выбранном канале.
{{< /alert >}}

## Поддерживаемые методы аутентификации

MFA с помощью Multifactor можно включить для любого метода аутентификации, при использовании которого создаётся сущность (entity) Stronghold. Например:

- `userpass`
- `ldap`
- `oidc`
- `jwt`
- `saml`
- `webauthn`
- `approle`
- `kubernetes`

Multifactor MFA рассчитан на интерактивное подтверждение.
Для машинных методов вроде `approle` и `kubernetes` автоматический клиент не сможет самостоятельно подтвердить запрос через push-уведомление, Telegram или телефонный звонок.

Для таких методов можно настроить подтверждение запросов через общую учётную запись Multifactor.
Для этого в параметре `username_format` укажите статическое имя учётной записи, например, администратора.
Все запросы будут поступать владельцу этой учётной записи, который сможет подтверждать их с учётом идентификатора машины, указанного в запросе.

## Предварительные требования

Перед настройкой Multifactor MFA выполните следующие шаги:

1. В личном кабинете Multifactor создайте ресурс для серверной аутентификации, например, Linux.
   Не используйте тип ресурса WebSite.

1. Скопируйте из настроек созданного ресурса значения «NAS Identifier» и «Shared Secret».

1. Убедитесь, что для пользователей созданы учётные записи в Multifactor или настроена их автоматическая регистрация.

   Stronghold не синхронизирует пользователей с Multifactor, но позволяет зарегистрировать их при первом обращении.
   Если в настройках ресурса Multifactor включена опция «Регистрировать новых пользователей»,
   при первом успешном подключении Multifactor создаёт учётную запись и пишет информацию об этом в журнал доступа.
   Доступ выдаётся в соответствии с политиками Multifactor.
   Если автоматическая регистрация выключена, Multifactor откажет в доступе и не зарегистрирует новые учётные записи.

   Значение identity в Multifactor должно совпадать со значением, которое формирует параметр `username_format`.

Учётные записи пользователей также можно создать вручную или через API Multifactor, либо синхронизировать из AD/LDAP средствами Directory Sync в Multifactor.

## Создание метода MFA

Чтобы создать метод MFA Multifactor и получить его идентификатор, выполните следующую команду:

```shell
d8 stronghold write identity/mfa/method/multifactor \
  method_name=my-mfa \
  nas_identifier="rs_nas_id" \
  shared_secret="secret"
```

Пример вывода:

```console
Key          Value
---          -----
method_id    93964fd0-dd7e-e22a-74d0-0880ca5e0398
```

Параметры метода:

| Параметр | Описание |
| --- | --- |
| `method_name` | Уникальное имя метода MFA |
| `nas_identifier` | NAS Identifier из настроек ресурса Multifactor |
| `shared_secret` | Shared Secret из настроек ресурса Multifactor |
| `api_url` | Базовый URL-адрес API Multifactor. По умолчанию — `https://api.multifactor.ru` |
| `username_format` | Шаблон преобразования сущности Stronghold в identity Multifactor. Доступные параметры описаны в разделе [«Шаблонные политики»](../../../concepts/policy/#шаблонные-политики). Если шаблон не задан, используется имя сущности |
| `timeout_seconds` | Максимальное время ожидания подтверждения в секундах. Значение по умолчанию — `90`, минимальное значение — `65` |

{{< alert level="warning" >}}
Значение параметра `shared_secret` сохраняется только при создании или обновлении метода, но не возвращается при чтении его конфигурации.
{{< /alert >}}

## Включение MFA

Ниже приведён пример настройки MFA для метода аутентификации Userpass.

1. Получите идентификатор (accessor) метода аутентификации:

   ```shell
   USERPASS_ACCESSOR=$(d8 stronghold auth list -format=json \
       --detailed | jq -r '."userpass/".accessor')
   echo $USERPASS_ACCESSOR
   ```

1. Включите MFA:

   ```shell
   d8 stronghold write /identity/mfa/login-enforcement/userpass-multifactor-enforcement \
       mfa_method_ids="93964fd0-dd7e-e22a-74d0-0880ca5e0398" \
       auth_method_accessors=$USERPASS_ACCESSOR
   ```

1. Выполните вход:

   ```shell
   d8 stronghold login -method=userpass username=mfa-user
   Password (will be hidden):
   ```

   Пример вывода:

   ```console
   Initiating Interactive MFA Validation...
   Asking Stronghold to perform MFA validation with upstream service. You should receive a push notification in your authenticator app shortly
   Success! You are now authenticated. The token information displayed below is
   already stored in the token helper. You do NOT need to run "stronghold login"
   again. Future Stronghold requests will automatically use this token.

   Key                    Value
   ---                    -----
   token                  hv.....
   token_accessor         uH1voyZljOCbttJSICUtom17
   token_duration         768h
   token_renewable        true
   token_policies         ["default"]
   policies               ["default"]
   token_meta_username    mfa-user
   ```

После успешной проверки основного фактора Stronghold отправит запрос в Multifactor и будет ожидать подтверждения.
Подтвердите запрос в мобильном приложении Multifactor, в Telegram или по телефону в зависимости от настроек пользователя.
После получения статуса «Granted» Stronghold выдаст токен.

Чтобы отключить проверку MFA для метода Userpass, выполните следующую команду:

```shell
d8 stronghold delete identity/mfa/login-enforcement/userpass-multifactor-enforcement
```
