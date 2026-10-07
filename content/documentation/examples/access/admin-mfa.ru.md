---
title: "MFA для администраторов"
linkTitle: "MFA для администраторов"
description: "Обязательный второй фактор TOTP при входе администраторов в Stronghold: метод TOTP, правило login enforcement для группы администраторов, регистрация по QR-коду, вход в CLI и веб-интерфейсе, восстановление при утере устройства и контроль по журналу аудита."
weight: 40
params:
  relatedLinks:
    - title: "TOTP"
      url: ../../../user/auth/mfa/totp/
    - title: "Multifactor"
      url: ../../../user/auth/mfa/multifactor/
    - title: "Идентификационные данные (Identity)"
      url: ../../../concepts/identity/
    - title: "API Identity"
      url: ../../../reference/api/identity/
    - title: "Аудит"
      url: ../../../admin/audit/overview/
    - title: "Аварийный доступ (break-glass)"
      url: ../break-glass-access/
---

Учётная запись администратора Stronghold даёт доступ к политикам, методам аутентификации и секретам, поэтому одного пароля или сессии IdP для неё недостаточно. В этом руководстве для всех администраторов включается обязательная проверка одноразового кода TOTP при входе.

## Механизмы MFA в Stronghold

Stronghold проверяет второй фактор на этапе входа (login MFA): после успешной проверки основного метода аутентификации токен выдаётся только после подтверждения второго фактора. Методы MFA и правила их применения настраиваются в подсистеме Identity по путям `identity/mfa/method/*` и `identity/mfa/login-enforcement/*`.

В документации описаны два метода MFA:

| Метод | Как подтверждается вход | Когда выбирать |
| --- | --- | --- |
| [TOTP](../../../user/auth/mfa/totp/) | Одноразовый код из приложения-аутентификатора | Не нужен внешний сервис, подходит для закрытого контура |
| [Multifactor](../../../user/auth/mfa/multifactor/) | Push-уведомление, Telegram или звонок через сервис Multifactor | В организации уже используется Multifactor |

Правило login enforcement может применяться к сущностям (`identity_entity_ids`), группам Identity (`identity_group_ids`), конкретным подключениям методов аутентификации (`auth_method_accessors`) или ко всем подключениям метода определённого типа (`auth_method_types`).

Проверка MFA при обращении к отдельным путям (step-up MFA через `sys/mfa/method/*`) в Stronghold не описана: в справочнике API эти пути отмечены как заглушки Enterprise-функций. <!-- TODO(verify): поддерживается ли step-up MFA (sys/mfa/method/*) в Stronghold EE -->

{{< alert level="info" >}}
Аудит-логирование, которое используется в разделе [«Контроль по журналу аудита»](#контроль-по-журналу-аудита), доступно только в Stronghold EE.
{{< /alert >}}

<!-- TODO(verify): доступность MFA (TOTP и Multifactor) по редакциям — на странице редакций MFA не упомянута -->

## Цель

- Все администраторы Stronghold при входе вводят код TOTP.
- Проверка применяется к группе Identity `stronghold-admins`, поэтому новые администраторы получают её автоматически.
- Задокументирована процедура восстановления, если администратор потерял устройство с аутентификатором.
- Операции с настройками MFA видны в журнале аудита.

![Схема входа администратора с TOTP](../../../images/ex-admin-mfa.png)

## Предварительные требования

- Токен Stronghold с правами на настройку Identity и политик.
- Группа Identity `stronghold-admins`, в которую входят сущности администраторов. Для каталогов и IdP используйте внешнюю группу с алиасом группы каталога, как в руководстве [«Интеграция с Active Directory»](../active-directory/).
- Приложение-аутентификатор с поддержкой TOTP у каждого администратора.
- Подготовленный [аварийный доступ](../break-glass-access/), на который правило MFA не распространяется.
- Утилита `jq` на рабочей станции администратора.
- Для контроля по журналу аудита — Stronghold EE и включённое устройство аудита.

## Шаг 1. Создайте метод TOTP

Создайте метод MFA TOTP и сохраните его идентификатор:

```bash
TOTP_METHOD_ID=$(d8 stronghold write -format=json identity/mfa/method/totp \
  method_name=admin-totp \
  generate=true \
  issuer=Stronghold \
  period=30 \
  algorithm=SHA256 \
  digits=6 \
  max_validation_attempts=5 | jq -r '.data.method_id')
echo $TOTP_METHOD_ID
```

Основные параметры метода:

| Параметр | Описание |
| --- | --- |
| `method_name` | Уникальное имя метода MFA |
| `issuer` | Название, которое приложение-аутентификатор показывает рядом с кодом |
| `period` | Период смены кода в секундах. По умолчанию — `30` |
| `algorithm` | Алгоритм хеширования: `SHA1` (по умолчанию), `SHA256` или `SHA512` |
| `digits` | Количество цифр в коде: `6` или `8` |
| `skew` | Допустимое отклонение в периодах при проверке кода: `0` или `1`. По умолчанию — `1` |
| `max_validation_attempts` | Максимальное количество попыток проверки кода |

Полный список параметров приведён в [справочнике API Identity](../../../reference/api/identity/).

Не все приложения-аутентификаторы поддерживают `SHA256` и 8-значные коды. Если у администраторов разные приложения, используйте `algorithm=SHA1` и `digits=6`. <!-- TODO(verify): рекомендуемые параметры TOTP для совместимости с распространёнными приложениями-аутентификаторами -->

## Шаг 2. Зарегистрируйте аутентификаторы администраторов

Зарегистрируйте TOTP у всех администраторов **до** включения обязательной проверки. Если у сущности нет секрета, проверка второго фактора завершается ошибкой `MFA secret … not present in entity`, и вход не удаётся.

Сущность администратора создаётся при первом входе в Stronghold. Администратор может узнать идентификатор своей сущности так:

```bash
d8 stronghold token lookup -format=json | jq -r '.data.entity_id'
```

Выберите один из способов регистрации.

### Выпуск QR-кода администратором безопасности

1. Получите идентификатор сущности по имени:

   ```bash
   ENTITY_ID=$(d8 stronghold read -field=id identity/entity/name/<entity_name>)
   ```

1. Сгенерируйте TOTP-секрет и QR-код:

   ```bash
   d8 stronghold write -field=barcode \
     identity/mfa/method/totp/admin-generate \
     method_id=$TOTP_METHOD_ID entity_id=$ENTITY_ID \
     | base64 -d > /tmp/qr-code.png
   ```

1. Передайте QR-код администратору по защищённому каналу, дождитесь, пока он отсканирует его в приложении-аутентификаторе, и удалите файл:

   ```bash
   shred -u /tmp/qr-code.png
   ```

Повторный вызов `admin-generate` для сущности, у которой уже есть секрет этого метода, новый секрет не создаёт: Stronghold возвращает предупреждение `Entity already has a secret for MFA method`. Чтобы перерегистрировать TOTP, сначала удалите секрет через `admin-destroy` (см. раздел «Восстановление при утере устройства» ниже).

### Самостоятельная регистрация

1. Добавьте в политику администраторов право на выпуск собственного секрета:

   ```hcl
   path "identity/mfa/method/totp/generate" {
     capabilities = ["update"]
   }
   ```

1. Администратор получает QR-код для своей сущности в веб-интерфейсе Stronghold или через CLI:

   ```bash
   d8 stronghold write -field=barcode identity/mfa/method/totp/generate \
     method_id=$TOTP_METHOD_ID | base64 -d > /tmp/qr-code.png
   ```

Если секрет метода у сущности уже есть, `generate` его не перезаписывает, а возвращает предупреждение, поэтому украденная сессия не позволяет перерегистрировать TOTP; для сброса нужен `admin-destroy`.

## Шаг 3. Включите обязательную проверку

1. Получите идентификатор группы администраторов:

   ```bash
   ADMINS_GROUP_ID=$(d8 stronghold read -field=id identity/group/name/stronghold-admins)
   ```

1. Создайте правило login enforcement:

   ```bash
   d8 stronghold write identity/mfa/login-enforcement/admin-totp \
     mfa_method_ids="$TOTP_METHOD_ID" \
     identity_group_ids="$ADMINS_GROUP_ID"
   ```

Если администраторы входят через отдельное подключение метода аутентификации, например `userpass` по пути `admin/`, правило можно привязать к этому подключению:

```bash
ADMIN_ACCESSOR=$(d8 stronghold auth list -format=json -detailed \
  | jq -r '."admin/".accessor')

d8 stronghold write identity/mfa/login-enforcement/admin-totp \
  mfa_method_ids="$TOTP_METHOD_ID" \
  auth_method_accessors="$ADMIN_ACCESSOR"
```

{{< alert level="warning" >}}
Не включайте в правило аварийную учётную запись и её подключение метода аутентификации (например, `breakglass/`). Иначе при потере аутентификаторов войти в Stronghold не сможет никто.
{{< /alert >}}

## Вход с MFA

### CLI

После проверки основного фактора CLI запрашивает код TOTP:

```bash
d8 stronghold login -method=oidc
```

Пример вывода:

```console
Initiating Interactive MFA Validation...
Enter the passphrase for methodID "22c35aa4-bf37-cf31-4187-c5a676c19aca" of type "totp":
```

Введите текущий код из приложения-аутентификатора. После успешной проверки CLI сохранит токен.

При входе через API ответ метода аутентификации содержит `mfa_request_id`. Передайте его вместе с кодом в `sys/mfa/validate`:

```bash
d8 stronghold write -format=json sys/mfa/validate - <<EOF
{
  "mfa_request_id": "<mfa_request_id>",
  "mfa_payload": {
    "$TOTP_METHOD_ID": ["<totp_code>"]
  }
}
EOF
```

Идентификатор находится в поле `auth.mfa_requirement.mfa_request_id`. Чтобы CLI не запрашивал код интерактивно и только вывел `mfa_request_id`, добавьте флаг `-non-interactive` к `d8 stronghold login`.

### Веб-интерфейс

1. Откройте веб-интерфейс Stronghold и выполните вход выбранным методом.
1. После проверки основного фактора введите код TOTP из приложения-аутентификатора.

## Восстановление при утере устройства

Если администратор потерял устройство с аутентификатором или оно могло быть скомпрометировано:

1. Подтвердите личность администратора по независимому каналу, например лично или через руководителя. Зафиксируйте обращение в журнале инцидентов.
1. Войдите учётной записью администратора безопасности с правами на `admin-destroy` и `admin-generate`. Пример политики:

   ```hcl
   path "identity/entity/name/*" {
     capabilities = ["read"]
   }

   path "identity/mfa/method/totp/admin-destroy" {
     capabilities = ["update"]
   }

   path "identity/mfa/method/totp/admin-generate" {
     capabilities = ["update"]
   }
   ```

1. Удалите TOTP-секрет сущности администратора:

   ```bash
   ENTITY_ID=$(d8 stronghold read -field=id identity/entity/name/<entity_name>)

   d8 stronghold write identity/mfa/method/totp/admin-destroy \
     method_id=$TOTP_METHOD_ID entity_id=$ENTITY_ID
   ```

1. Выпустите новый QR-код, как в шаге 2, и передайте его администратору по защищённому каналу.
1. Если устройство могло попасть к посторонним вместе с паролем, смените пароль или учётные данные основного метода входа.

Если восстановить нужно доступ единственного администратора, а других учётных записей с правами на `admin-destroy` нет, воспользуйтесь [аварийным доступом](../break-glass-access/).

## Контроль по журналу аудита

{{< alert level="info" >}}
Аудит-логирование доступно только в Stronghold EE.
{{< /alert >}}

1. Убедитесь, что устройство аудита включено:

   ```bash
   d8 stronghold audit list
   ```

1. Настройте в SIEM оповещения об изменении настроек MFA. Признаки в журнале аудита:

   - `request.path` равен `identity/mfa/method/totp/admin-destroy` или `identity/mfa/method/totp/admin-generate`;
   - `request.path` начинается с `identity/mfa/login-enforcement/` или `identity/mfa/method/`, а `request.operation` равен `update` или `delete`.

   Пример поиска по файловому журналу:

   ```bash
   jq -c 'select(.type == "request" and (.request.path | startswith("identity/mfa/")))
     | {time, path: .request.path, op: .request.operation, entity: .auth.entity_id}' \
     /var/log/stronghold_audit.log
   ```

1. Отслеживайте проверки второго фактора — запросы к `sys/mfa/validate`. Большое количество неуспешных проверок для одной сущности может означать попытку подбора кода. <!-- TODO(verify): попадают ли в журнал аудита проверки MFA при интерактивном входе через CLI и веб-интерфейс -->

## Проверка

1. Убедитесь, что правило создано:

   ```bash
   d8 stronghold read identity/mfa/login-enforcement/admin-totp
   ```

1. Войдите учётной записью администратора и убедитесь, что Stronghold запрашивает код TOTP.
1. Введите неверный код и убедитесь, что токен не выдан.
1. Войдите аварийной учётной записью и убедитесь, что код TOTP не запрашивается.

## Очистка

1. Удалите правило login enforcement:

   ```bash
   d8 stronghold delete identity/mfa/login-enforcement/admin-totp
   ```

1. Удалите метод TOTP:

   ```bash
   d8 stronghold delete identity/mfa/method/totp/$TOTP_METHOD_ID
   ```
