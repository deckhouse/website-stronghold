---
title: "Multifactor"
description: "Настройка многофакторной аутентификации Stronghold через сервис Multifactor"
weight: 10
---

## Настройка Multifactor

Stronghold поддерживает проверку дополнительного фактора при аутентификации через сервис [Multifactor](https://multifactor.ru/docs/intro/).
После успешного входа по основному методу аутентификации Stronghold отправляет запрос в Multifactor и ожидает подтверждения от пользователя.

Подтверждение выполняется одним из способов, настроенных для пользователя в Multifactor:

- Push в мобильное приложение Multifactor;
- Telegram;
- звонок — примите вызов и нажмите `#`.

{{< alert level="info" >}}
Stronghold не принимает одноразовые коды (OTP) для Multifactor MFA.
Запрос обрабатывается как подтверждение доступа: пользователь одобряет или отклоняет его в выбранном канале.
{{< /alert >}}

### Поддерживаемые методы аутентификации

Login MFA, в том числе Multifactor, не ограничивает список типов методов аутентификации на уровне кода.
Проверку можно включить для любого метода, у которого при входе есть сущность (entity), например (но не ограничиваясь):

- `userpass`
- `ldap`
- `oidc`
- `jwt`
- `saml`
- `approle`
- `kubernetes`

Multifactor MFA рассчитан на интерактивное подтверждение.
Для машинных методов вроде `approle` и `kubernetes` автоматический клиент сам подтвердить push, Telegram
или звонок не сможет.

Для таких случаев возможен сценарий с общим подтверждением: в `username_format` задайте статическое имя
учётной записи Multifactor, например администратора.
Тогда все запросы аутентификации будут приходить на эту учётную запись,
а оператор подтверждает вход, ориентируясь на идентификатор машины в запросе.

### Предварительные требования

Перед настройкой выполните следующие шаги:

1. В личном кабинете Multifactor создайте ресурс для серверной аутентификации, например `Linux`.
   Не используйте тип ресурса `WebSite`.

1. Скопируйте из настроек ресурса значения **NAS Identifier** и **Shared Secret**.

1. Убедитесь, что пользователи есть в Multifactor или будет работать их автоматическая регистрация.
   Stronghold не синхронизирует пользователей с Multifactor, но передаёт запрос с возможностью inline-регистрации.
   Если в настройках ресурса Multifactor включена опция **Регистрировать новых пользователей**,
   при первом успешном подключении Multifactor создаёт учётную запись и пишет об этом в журнал доступа.
   Доступ выдаётся в соответствии с политиками Multifactor.
   Если опция выключена, Multifactor отказывает в доступе и не регистрирует новые учётные записи.

   Значение identity в Multifactor должно совпадать со значением, которое формирует параметр `username_format`.

Пользователей также можно завести вручную или через API Multifactor, а также синхронизировать из AD/LDAP средствами Directory Sync в Multifactor.

### Создание метода MFA

Чтобы включить метод MFA Multifactor и получить его идентификатор, выполните:

```shell
d8 stronghold write identity/mfa/method/multifactor method_name=my-mfa nas_identifier="rs_nas_id" shared_secret="secret"

Key          Value
---          -----
method_id    93964fd0-dd7e-e22a-74d0-0880ca5e0398

```

Параметры метода:

| Параметр | Описание |
| --- | --- |
| `method_name` | Уникальное имя метода MFA. |
| `nas_identifier` | NAS Identifier из настроек ресурса Multifactor. |
| `shared_secret` | Shared Secret из настроек ресурса Multifactor. |
| `api_url` | Базовый URL API Multifactor. По умолчанию — `https://api.multifactor.ru`. |
| `username_format` | Шаблон преобразования сущности Stronghold в identity Multifactor. Если не задан, используется имя сущности. |
| `timeout_seconds` | Максимальное время ожидания подтверждения в секундах. Значение по умолчанию — `90`, минимум — `65`. |

{{< alert level="warning" >}}
Параметр `shared_secret` записывается только при создании или обновлении метода.
При чтении конфигурации метода значение секрета не возвращается.
{{< /alert >}}

### Включение MFA

Ниже приведён пример проверки MFA для метода аутентификации Userpass.

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

После успешной проверки основного фактора Stronghold инициирует запрос в Multifactor и ожидает подтверждения.
Подтвердите запрос в мобильном приложении Multifactor, в Telegram или по звонку — в зависимости от настроек пользователя в Multifactor.
После статуса `Granted` Stronghold выдаст токен.

Чтобы отключить проверку MFA, выполните:

```shell
d8 stronghold delete identity/mfa/login-enforcement/userpass-multifactor-enforcement
```
