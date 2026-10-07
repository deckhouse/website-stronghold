---
title: "Интеграция с Active Directory"
linkTitle: "Active Directory"
description: "Вход в Stronghold под учётными записями Active Directory через метод LDAP с LDAPS, вложенными группами и UPN, а также ротация паролей служебных учётных записей AD механизмом секретов LDAP со схемой ad."
weight: 20
params:
  relatedLinks:
    - title: "Метод LDAP"
      url: ../../../user/auth/ldap/
    - title: "Механизм секретов LDAP"
      url: ../../../user/secrets-engines/ldap/
    - title: "Идентификационные данные (Identity)"
      url: ../../../concepts/identity/
    - title: "Ротация паролей служебных учётных записей"
      url: ../../dynamic-credentials/static-credentials-rotation/
    - title: "Интеграция с ALD Pro"
      url: ../ald-pro/
---

Stronghold подключается к Active Directory (AD) по протоколу LDAP: проверяет пароли пользователей домена, определяет их группы, в том числе вложенные, и назначает политики по группам. Механизм секретов LDAP со схемой `ad` меняет пароли служебных учётных записей домена.

## Цель

- Пользователи AD входят в Stronghold под доменными учётными записями через метод `ldap` по LDAPS.
- Права в Stronghold назначаются по членству в группах AD с учётом вложенности.
- Stronghold ротирует пароли служебных учётных записей AD и выдаёт их приложениям.

## Предварительные требования

- Домен AD, например `corp.example.com` с контроллерами `dc01.corp.example.com` и `dc02.corp.example.com`.
- На контроллерах домена включён LDAPS (порт 636) с сертификатом, выданным корпоративным УЦ.
- Сертификат корневого УЦ в формате PEM (Base-64), например `corp-ca.pem`.
- Сетевой доступ от узлов Stronghold к контроллерам домена по порту 636.
- Токен Stronghold с правами на настройку методов аутентификации, механизмов секретов, политик и Identity.

В примерах пользователи находятся в `OU=Users,DC=corp,DC=example,DC=com`, группы — в `OU=Groups,DC=corp,DC=example,DC=com`, служебные учётные записи — в `OU=Service Accounts,DC=corp,DC=example,DC=com`. Замените DN на свои.

## Шаг 1. Подготовьте учётные записи в AD

1. Создайте учётную запись `svc-stronghold-bind` для поиска пользователей и групп. Достаточно прав обычного пользователя домена на чтение каталога. Задайте бессрочный пароль.

1. Если планируете ротацию паролей, создайте учётную запись `svc-stronghold-rotator` и делегируйте ей право **Reset password** на OU со служебными учётными записями. Не выдавайте ей права администратора домена.

1. Проверьте поиск по LDAPS:

   ```bash
   LDAPTLS_CACERT=./corp-ca.pem ldapsearch -x -H ldaps://dc01.corp.example.com \
     -D "svc-stronghold-bind@corp.example.com" -W \
     -b "OU=Users,DC=corp,DC=example,DC=com" "(sAMAccountName=ivanov)" \
     sAMAccountName userPrincipalName memberOf
   ```

## Шаг 2. Настройте метод LDAP

1. Включите метод аутентификации:

   ```bash
   d8 stronghold auth enable ldap
   ```

1. Настройте подключение. Пользователь входит по `sAMAccountName`, группы ищутся с учётом вложенности через правило `LDAP_MATCHING_RULE_IN_CHAIN` (OID `1.2.840.113556.1.4.1941`):

   ```bash
   d8 stronghold write auth/ldap/config \
     url="ldaps://dc01.corp.example.com:636,ldaps://dc02.corp.example.com:636" \
     certificate=@corp-ca.pem \
     insecure_tls=false \
     starttls=false \
     binddn="CN=svc-stronghold-bind,OU=Service Accounts,DC=corp,DC=example,DC=com" \
     bindpass='<пароль>' \
     userdn="OU=Users,DC=corp,DC=example,DC=com" \
     userattr="sAMAccountName" \
     groupdn="OU=Groups,DC=corp,DC=example,DC=com" \
     groupfilter="(&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))" \
     groupattr="cn" \
     username_as_alias=true
   ```

   Основные параметры:

   | Параметр | Значение для AD |
   | --- | --- |
   | `url` | Список контроллеров домена через запятую, порт 636. |
   | `certificate` | Сертификат корневого УЦ, выдавшего сертификаты контроллеров. |
   | `insecure_tls`, `starttls` | `false` для `ldaps://`. При подключении по `ldap://` на порту 389 укажите `starttls=true`. |
   | `userattr` | `sAMAccountName` — короткое имя входа, например `ivanov`. |
   | `groupfilter` | Находит все группы, в которые пользователь входит прямо или через вложенные группы. |
   | `groupattr` | `cn` — имя группы в Stronghold. |

   Вместо `groupfilter` можно включить `use_token_groups=true`: Stronghold прочитает вычисляемый атрибут `tokenGroups` пользователя, который содержит все группы безопасности, включая вложенные.

1. Если пользователи должны входить по UPN (`ivanov@corp.example.com`), замените `userattr` на параметр `upndomain`:

   ```bash
   d8 stronghold write auth/ldap/config \
     url="ldaps://dc01.corp.example.com:636,ldaps://dc02.corp.example.com:636" \
     certificate=@corp-ca.pem \
     insecure_tls=false \
     binddn="CN=svc-stronghold-bind,OU=Service Accounts,DC=corp,DC=example,DC=com" \
     bindpass='<пароль>' \
     userdn="OU=Users,DC=corp,DC=example,DC=com" \
     upndomain="corp.example.com" \
     groupdn="OU=Groups,DC=corp,DC=example,DC=com" \
     groupfilter="(&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))" \
     groupattr="cn"
   ```

   Пользователь по-прежнему вводит короткое имя, а Stronghold выполняет bind как `ivanov@corp.example.com` и ищет его по `userPrincipalName`.

   Команда `write` заменяет конфигурацию целиком. При изменении отдельных параметров повторите в команде все остальные.

1. Проверьте вход:

   ```bash
   d8 stronghold login -method=ldap username=ivanov
   ```

## Шаг 3. Сопоставьте группы AD с политиками

Для одного метода входа достаточно групп метода LDAP. Имя группы должно совпадать с `cn` группы в AD:

```bash
d8 stronghold write auth/ldap/groups/SG-Stronghold-Admins policies=admin
d8 stronghold write auth/ldap/groups/SG-Developers policies=dev-read
```

Если группы AD используются и в других методах входа или к ним нужно применить MFA, создайте внешние группы Identity:

```bash
GROUP_ID=$(d8 stronghold write -field=id identity/group \
  name="ad-stronghold-admins" type="external" policies="admin")
LDAP_ACCESSOR=$(d8 stronghold auth list -format=json | jq -r '."ldap/".accessor')
d8 stronghold write identity/group-alias \
  name="SG-Stronghold-Admins" \
  mount_accessor="$LDAP_ACCESSOR" \
  canonical_id="$GROUP_ID"
```

Подробнее о внешних группах — в разделе [«Идентификационные данные (Identity)»](../../../concepts/identity/#внешние-и-внутренние-группы).

## Шаг 4. Ротируйте пароли служебных учётных записей

{{< alert level="info" >}}
AD разрешает менять пароль (атрибут `unicodePwd`) только по защищённому соединению. Используйте `ldaps://` или `starttls=true`.
{{< /alert >}}

1. Включите механизм секретов и настройте его на схему `ad`:

   ```bash
   d8 stronghold secrets enable ldap
   d8 stronghold write ldap/config \
     url="ldaps://dc01.corp.example.com:636" \
     certificate=@corp-ca.pem \
     insecure_tls=false \
     binddn="CN=svc-stronghold-rotator,OU=Service Accounts,DC=corp,DC=example,DC=com" \
     bindpass='<пароль>' \
     userdn="OU=Service Accounts,DC=corp,DC=example,DC=com" \
     schema=ad
   ```

1. Смените пароль `svc-stronghold-rotator`, чтобы его знал только Stronghold:

   ```bash
   d8 stronghold write -f ldap/rotate-root
   ```

1. Создайте статическую роль для служебной учётной записи приложения:

   ```bash
   d8 stronghold write ldap/static-role/svc-jenkins \
     dn="CN=svc-jenkins,OU=Service Accounts,DC=corp,DC=example,DC=com" \
     username="svc-jenkins" \
     rotation_period="24h"
   ```

1. Приложение читает текущий пароль:

   ```bash
   d8 stronghold read ldap/static-cred/svc-jenkins
   ```

Для общих учётных записей, которые выдаются людям на время, используйте библиотеку:

```bash
d8 stronghold write ldap/library/helpdesk \
  service_account_names="svc-helpdesk1@corp.example.com,svc-helpdesk2@corp.example.com" \
  ttl=2h \
  max_ttl=8h
d8 stronghold write -f ldap/library/helpdesk/check-out
d8 stronghold write ldap/library/helpdesk/check-in service_account_names="svc-helpdesk1@corp.example.com"
```

Значения `service_account_names` сравниваются с атрибутом из параметра `userattr`; для схемы `ad` по умолчанию это `userPrincipalName`, поэтому указывайте полное имя вида `svc-helpdesk1@corp.example.com`.

Политики доступа к `static-cred` и `library` аналогичны приведённым в разделе [«Интеграция с ALD Pro»](../ald-pro/#шаг-5-ротируйте-пароли-служебных-учётных-записей).

## Проверка

1. Войдите под пользователем из группы, вложенной в `SG-Stronghold-Admins`, и проверьте, что токен получил политику `admin`:

   ```bash
   d8 stronghold login -method=ldap username=ivanov
   d8 stronghold token lookup
   ```

1. Прочитайте пароль статической роли и проверьте bind:

   ```bash
   LDAPTLS_CACERT=./corp-ca.pem ldapsearch -x -H ldaps://dc01.corp.example.com \
     -D "svc-jenkins@corp.example.com" -w '<пароль из static-cred>' \
     -b "DC=corp,DC=example,DC=com" -s base
   ```

## Устранение неполадок

При ошибке bind AD возвращает код `49` и подкод в поле `data` сообщения об ошибке. Подкод показывает причину.

| Симптом | Вероятная причина | Что сделать |
| --- | --- | --- |
| `Invalid Credentials`, `data 52e` | Неверный пароль пользователя или `bindpass` | Проверьте пароль. После `rotate-root` пароль `binddn` известен только Stronghold. |
| `data 525` | Пользователь не найден | Проверьте `userdn`, `userattr` или `upndomain`. |
| `data 775` | Учётная запись заблокирована в AD | Разблокируйте её в AD. Проверьте также блокировку Stronghold (`lockout_threshold`), описанную в разделе [«Метод LDAP»](../../../user/auth/ldap/#блокировка-пользователя). |
| `data 533`, `data 532`, `data 701` | Учётная запись отключена, пароль или учётная запись истекли | Включите учётную запись или смените пароль в AD. |
| `x509: certificate signed by unknown authority` | Не указан или указан не тот сертификат УЦ | Передайте корневой сертификат корпоративного УЦ в `certificate`. |
| Вход проходит, но нет политик групп | `groupfilter` не находит группы или имена групп не совпадают | Выполните `ldapsearch` с фильтром `(member:1.2.840.113556.1.4.1941:=<DN пользователя>)`. Имена групп в Stronghold должны совпадать с `cn`. |
| `rotate-role` возвращает `Unwilling To Perform` или `Insufficient Access Rights` | Нет защищённого соединения, у `binddn` нет права **Reset password** или пароль не соответствует политике паролей домена | Используйте LDAPS, проверьте делегирование и настройте [политику паролей Stronghold](../../../concepts/password-policy/) через параметр `password_policy`. |

## Очистка

```bash
d8 stronghold auth disable ldap
d8 stronghold secrets disable ldap
```

При отключении механизма секретов пароли не меняются. Смените пароли служебных учётных записей средствами AD.
