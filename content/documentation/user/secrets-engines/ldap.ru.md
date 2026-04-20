---
title: "LDAP"
linkTitle: "LDAP"
description: "Работа с механизмом секретов LDAP в Deckhouse Stronghold"
weight: 80
---

Механизм секретов `LDAP` в Deckhouse Stronghold позволяет управлять LDAP-учётными данными и создавать их динамически. Он поддерживает интеграцию с LDAP v3-совместимыми системами, включая OpenLDAP, ALD Pro, Active Directory и IBM Resource Access Control Facility (RACF).

Механизм секретов `LDAP` решает три основные задачи:

- управление статическими учётными данными;
- управление динамическими учётными данными;
- ротация паролей для набора учётных записей.

## Когда использовать

Используйте механизм секретов `LDAP`, если нужно:

- централизованно управлять LDAP-учётными данными;
- автоматически ротировать пароли сервисных учётных записей;
- выдавать временные LDAP-учётные данные;
- работать со статическими и динамическими ролями в одном механизме;
- интегрироваться с OpenLDAP, Active Directory или RACF.

## Как это работает

Механизм секретов `LDAP` поддерживает несколько сценариев работы:

- **статические роли** — Stronghold хранит привязку к существующей LDAP-записи и может ротировать пароль для этой учётной записи;
- **динамические роли** — Stronghold создаёт новую LDAP-учётную запись на ограниченное время, а затем удаляет её;
- **ротация для набора учётных записей** — Stronghold управляет паролями группы сервисных учётных записей и периодически меняет их.

Перед началом работы Stronghold нужно настроить для подключения к LDAP-серверу.

## Базовая настройка

### Шаг 1. Включите механизм секретов LDAP

```shell-session
d8 stronghold secrets enable ldap
```

По умолчанию механизм секретов монтируется по пути `ldap`. Если нужен другой путь, используйте аргумент `-path`.

### Шаг 2. Настройте подключение к LDAP

Укажите учётные данные, которые Stronghold будет использовать для подключения к LDAP и генерации паролей:

```shell-session
d8 stronghold write ldap/config \
    binddn=$USERNAME \
    bindpass=$PASSWORD \
    url=ldaps://138.91.247.105
```

Рекомендуется создать отдельную учётную запись специально для Stronghold.

### Шаг 3. Выполните ротацию пароля root-учётной записи

Чтобы пароль хранился только в Stronghold, выполните:

```shell-session
d8 stronghold write -f ldap/rotate-root
```

Сгенерированный пароль после ротации получить из Stronghold нельзя.

## Поддерживаемые схемы LDAP

Механизм секретов `LDAP` поддерживает три схемы:

- `openldap` — используется по умолчанию;
- `racf`;
- `ad`.

### OpenLDAP

По умолчанию механизм секретов `LDAP` предполагает, что пароль учётной записи хранится в атрибуте `userPassword`.

Такой атрибут используют разные object class, например:

- `organization`;
- `organizationalUnit`;
- `organizationalRole`;
- `inetOrgPerson`;
- `person`;
- `posixAccount`.

### RACF

Для IBM Resource Access Control Facility нужно явно указать схему `racf`.

Для поддержки `RACF` генерируемые пароли должны содержать не более 8 символов. Длину пароля можно настроить через политику паролей:

```bash
d8 stronghold write ldap/config \
 binddn=$USERNAME \
 bindpass=$PASSWORD \
 url=ldaps://138.91.247.105 \
 schema=racf \
 password_policy=racf_password_policy
```

### Active Directory

Для работы с Active Directory укажите схему `ad`:

```bash
d8 stronghold write ldap/config \
 binddn=$USERNAME \
 bindpass=$PASSWORD \
 url=ldaps://138.91.247.105 \
 schema=ad
```

## Статические роли

Статическая роль связывает имя роли в Stronghold с существующей записью в LDAP. Stronghold может использовать эту роль для управления ротацией пароля.

### Настройка статической роли

Создайте статическую роль:

```shell-session
d8 stronghold write ldap/static-role/lf-edge\
    dn='uid=lf-edge,ou=users,dc=lf-edge,dc=com' \
    username='stronghold'\
    rotation_period="24h"
```

В этом примере:

- `dn` — DN существующей LDAP-записи;
- `username` — имя учётной записи;
- `rotation_period` — период автоматической ротации пароля.

### Получение учётных данных статической роли

Запросите учётные данные для роли:

```shell-session
d8 stronghold read ldap/static-cred/lf-edge
```

## Ротация паролей для статических ролей

Для статических ролей механизм секретов `LDAP` поддерживает два способа ротации паролей:

- автоматическую ротацию по времени;
- ручную ротацию.

### Автоматическая ротация

Пароли автоматически меняются в соответствии со значением `rotation_period`, которое указано в статической роли. Минимальное значение — 5 секунд.

При запросе учётных данных Stronghold возвращает `ttl` — время до следующей ротации.

Сейчас авторотация поддерживается только для статических ролей.

Учётную запись `binddn`, которую Stronghold использует для подключения к LDAP, нужно ротировать через вызов `rotate-root`, чтобы пароль знал только Stronghold.

### Ручная ротация

Пароль статической роли можно ротировать вручную через вызов `rotate-role`.

После ручной ротации период ротации начинается заново.

### Удаление статических ролей

При удалении статической роли пароль не меняется автоматически.

Перед удалением роли или отзывом доступа к ней пароль нужно ротировать вручную.

## Динамические роли

Динамическая роль нужна, если Stronghold должен создавать LDAP-учётные данные на ограниченное время.

### Настройка динамической роли

Создайте динамическую роль через путь `/role/:role_name`:

```bash
d8 stronghold write ldap/role/dynamic-role \
  creation_ldif=@/path/to/creation.ldif \
  deletion_ldif=@/path/to/deletion.ldif \
  rollback_ldif=@/path/to/rollback.ldif \
  default_ttl=1h \
  max_ttl=24h
```

Параметр `rollback_ldif` необязателен, но его лучше задавать. Если создание учётной записи завершится ошибкой, Stronghold выполнит операции из `rollback_ldif` и попытается удалить уже созданные объекты.

### Получение динамических учётных данных

Чтобы сгенерировать учётные данные, выполните:

```bash
d8 stronghold read ldap/creds/dynamic-role
```

Пример результата:

```console
Key                    Value
---                    -----
lease_id               ldap/creds/dynamic-role/HFgd6uKaDomVMvJpYbn9q4q5
lease_duration         1h
lease_renewable        true
distinguished_names    [cn=v_token_dynamic-role_FfH2i1c4dO_1611952635,ou=users,dc=learn,dc=example]
password               xWMjkIFMerYttEbzfnBVZvhRQGmhpAA0yeTya8fdmDB3LXDzGrjNEPV2bCPE9CW6
username               v_token_testrole_FfH2i1c4dO_1611952635
```

Поле `distinguished_names` содержит массив DN, которые Stronghold создал на основе `creation_ldif`.

Если в `creation_ldif` описано несколько записей, Stronghold включит в `distinguished_names` DN для каждой записи. Порядок сохраняется, дедупликации нет.

## LDIF-записи

Механизм секретов `LDAP` управляет пользовательскими учётными записями через LDIF-записи.

LDIF можно передавать как строку или как Base64-кодированную версию строки LDIF. Stronghold разберёт строку и проверит её на корректность.

При подготовке LDIF учитывайте следующее:

- в конце строк не должно быть пробелов;
- каждый блок `modify` должен начинаться после пустой строки;
- несколько модификаций для одного `dn` можно описать в одном блоке `modify`;
- каждая модификация должна завершаться одним символом `-`.

## Особенности Active Directory

При работе с Active Directory нужно учитывать несколько дополнительных правил.

### Создание пользователя в AD

Чтобы программно создать пользователя в AD, сначала нужно выполнить `add`, а затем `modify`, чтобы задать пароль и включить учётную запись.

### Установка пароля

Для пароля в AD используется атрибут `unicodePwd`. Перед ним должны стоять два двоеточия `::`.

При программной установке пароля должны выполняться такие условия:

- пароль должен быть заключён в двойные кавычки;
- пароль должен быть в формате `UTF16LE`;
- пароль должен быть закодирован в Base64.

### Включение учётной записи

После установки пароля учётную запись можно включить через атрибут `userAccountControl`:

- `512` — включить учётную запись;
- `65536` — отключить истечение срока действия пароля;
- `66048` — одновременно включить учётную запись и отключить истечение срока действия пароля.

### Ограничение для `sAMAccountName`

Поле `sAMAccountName` используется для совместимости с устаревшими системами Windows NT и ограничено 20 символами.

Стандартный `username_template` может оказаться слишком длинным. Если вы создаёте динамических пользователей в Active Directory, лучше заранее настроить `username_template` так, чтобы длина имени была меньше 20 символов.

### Работа с группами

Active Directory не позволяет напрямую менять атрибут `memberOf` пользователя.

Чтобы добавить нового динамического пользователя в группу, нужно отправить `modify` для нужной группы и изменить её атрибут `member`.

### Пример LDIF для Active Directory

Параметры `*_ldif` поддерживают шаблоны на языке Go template.

Пример LDIF для создания пользователя в Active Directory:

```ldif
dn: CN={{.Username}},OU=Stronghold,DC=adtesting,DC=lab
changetype: add
objectClass: top
objectClass: person
objectClass: organizationalPerson
objectClass: user
userPrincipalName: {{.Username}}@adtesting.lab
sAMAccountName: {{.Username}}

dn: CN={{.Username}},OU=Stronghold,DC=adtesting,DC=lab
changetype: modify
replace: unicodePwd
unicodePwd::{{ printf "%q" .Password | utf16le | base64 }}
-
replace: userAccountControl
userAccountControl: 66048
-

dn: CN=test-group,OU=Stronghold,DC=adtesting,DC=lab
changetype: modify
add: member
member: CN={{.Username}},OU=Stronghold,DC=adtesting,DC=lab
-
```

## Ротация паролей для набора учётных записей

Stronghold может автоматически менять пароли для группы учётных записей. Ротацию можно запускать вручную или по TTL.

Этот сценарий поддерживается для разных схем, включая OpenLDAP, Active Directory и RACF.

Ниже показан пример для Active Directory.

### Шаг 1. Включите механизм секретов и настройте подключение

```shell-session
$ d8 stronghold secrets enable ldap
Success! Enabled the ad secrets engine at: ldap/

$ d8 stronghold write ldap/config \
    binddn=$USERNAME \
    bindpass=$PASSWORD \
    url=ldaps://138.91.247.105 \
    userdn='dc=example,dc=com'
```

### Шаг 2. Настройте набор учётных записей

```shell-session
d8 stronghold write ldap/library/accounting-team \
    service_account_names=fizz@example.com,buzz@example.com \
    ttl=10h \
    max_ttl=20h \
    disable_check_in_enforcement=false
```

В этом примере:

- `service_account_names` — список существующих сервисных учётных записей;
- `ttl` — через какое время Stronghold снова выполнит ротацию пароля;
- `max_ttl` — максимальное время действия пароля после ротации;
- `disable_check_in_enforcement=false` — по умолчанию пароль должен регистрировать тот же субъект Stronghold или тот же клиентский токен, который выполнил ротацию.

По умолчанию значения `ttl` и `max_ttl` равны `24h`.

### Шаг 3. Проверьте статус набора учётных записей

```shell-session
d8 stronghold read ldap/library/accounting-team/status
```

Пример результата:

```shell-session
Key                 Value
---                 -----
buzz@example.com    map[available:true]
fizz@example.com    map[available:true]
```

### Шаг 4. Выполните ротацию паролей

```shell-session
d8 stronghold write -f ldap/library/accounting-team/check-out
```

Пример результата:

```shell-session
Key                     Value
---                     -----
lease_id                ldap/library/accounting-team/check-out/EpuS8cX7uEsDzOwW9kkKOyGW
lease_duration          10h
lease_renewable         true
password                ?@09AZKh03hBORZPJcTDgLfntlHqxLy29tcQjPVThzuwWAx/Twx4a2ZcRQRqrZ1w
service_account_name    fizz@example.com
```

### Шаг 5. При необходимости задайте меньший TTL

Если стандартный `ttl` слишком большой, можно указать меньшее значение:

```shell-session
d8 stronghold write ldap/library/accounting-team/check-out ttl=30m
```

Пример результата:

```shell-session
Key                     Value
---                     -----
lease_id                ldap/library/accounting-team/check-out/gMonJ2jB6kYs6d3Vw37WFDCY
lease_duration          30m
lease_renewable         true
password                ?@09AZerLLuJfEMbRqP+3yfQYDSq6laP48TCJRBJaJu/kDKLsq9WxL9szVAvL/E1
service_account_name    buzz@example.com
```

### Шаг 6. Продлите аренду при необходимости

```shell-session
d8 stronghold lease renew ldap/library/accounting-team/check-out/0C2wmeaDmsToVFc0zDiX9cMq
```

Пример результата:

```shell-session
Key                Value
---                -----
lease_id           ldap/library/accounting-team/check-out/0C2wmeaDmsToVFc0zDiX9cMq
lease_duration     10h
lease_renewable    true
```

В этом случае текущие пароли будут действовать дольше, потому что ротация будет отложена.

## Политика паролей LDAP

Механизм секретов `LDAP` не хеширует и не шифрует пароли перед изменением их значений в LDAP.

Из-за этого в LDAP пароли могут храниться в plain text, если на стороне LDAP-сервера не настроена соответствующая политика.

Чтобы избежать хранения паролей в открытом виде, настройте на LDAP-сервере LDAP password policy (`ppolicy`). Такая политика может, например, автоматически включать хеширование паролей.

Ниже пример настройки `ppolicy` для `dc=example,dc=com`:

```console
dn: cn=module{0},cn=config
changetype: modify
add: olcModuleLoad
olcModuleLoad: ppolicy

dn: olcOverlay={2}ppolicy,olcDatabase={1}mdb,cn=config
changetype: add
objectClass: olcPPolicyConfig
objectClass: olcOverlayConfig
olcOverlay: {2}ppolicy
olcPPolicyDefault: cn=default,ou=pwpolicies,dc=example,dc=com
olcPPolicyForwardUpdates: FALSE
olcPPolicyHashCleartext: TRUE
olcPPolicyUseLockout: TRUE
```

## Практические рекомендации

Чтобы механизм секретов `LDAP` было проще сопровождать:

- создайте отдельную LDAP-учётную запись для Stronghold;
- сразу настройте ротацию `binddn` через `rotate-root`, чтобы пароль хранился только в Stronghold;
- используйте статические роли для существующих сервисных учётных записей;
- используйте динамические роли там, где нужны временные LDAP-учётные данные;
- для динамических ролей добавляйте `rollback_ldif`, чтобы корректно откатывать неудачные операции;
- в Active Directory заранее проверьте длину `username_template`;
- на стороне LDAP обязательно настройте `ppolicy`, если не хотите хранить пароли в plain text.

## Что дальше

- Если вам нужны временные учётные данные для баз данных, используйте раздел [Базы данных](../database/).
- Если вам нужно хранить произвольные секреты, используйте [KV](../kv/).
