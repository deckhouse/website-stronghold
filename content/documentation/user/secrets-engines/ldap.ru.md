---
title: "Механизм секретов LDAP"
description: "Сведения о разделе \"Механизм секретов LDAP\" в Deckhouse Stronghold."
weight: 80
---

Механизм секретов LDAP предназначен для управления учётными данными LDAP и динамического создания учётных данных.
Он поддерживает интеграцию с реализациями протокола LDAP v3, включая OpenLDAP, [ALD Pro](https://www.aldpro.ru/), Active Directory и IBM Resource Access Control Facility (RACF).

Механизм секретов LDAP выполняет три основные функции:

- [Управление статическими учётными данными](#статические-роли);
- [Управление динамическими учётными данными](#динамические-роли);
- [Ротация паролей для списков учётных записей](#ротация-паролей-для-списков-учётных-записей).

## Настройка

Включите механизм секретов LDAP:

```shell
d8 stronghold secrets enable ldap
```

По умолчанию подключение произойдёт по пути `ldap`.
Для подключения по другому пути используйте аргумент `-path`.

Настройте учётные данные, которые Stronghold использует для подключения к LDAP при генерации паролей:

```shell
d8 stronghold write ldap/config \
  binddn=$USERNAME \
  bindpass=$PASSWORD \
  url=ldaps://138.91.247.105
```

{{< alert level="info" >}}
Рекомендуется создать отдельную учётную запись специально для Stronghold.
{{< /alert >}}

Выполните ротацию пароля, чтобы он хранился только в Stronghold:

```shell
d8 stronghold write -f ldap/rotate-root
```

{{< alert level="warning" >}}
Получить сгенерированный пароль после ротации в Stronghold невозможно.
{{< /alert >}}

### Схемы LDAP

{: #schemas .anchored}

Механизм секретов LDAP поддерживает три схемы:

- `openldap` — используется по умолчанию;
- `racf`;
- `ad`.

#### OpenLDAP

По умолчанию механизм секретов LDAP предполагает, что пароль для учётной записи хранится в поле `userPassword`.

Поле `userPassword` поддерживают, например, следующие классы объектов:

- `organization`;
- `organizationalUnit`;
- `organizationalRole`;
- `inetOrgPerson`;
- `person`;
- `posixAccount`.

#### Resource Access Control Facility

Для управления системой безопасности IBM Resource Access Control Facility (RACF) настройте механизм секретов LDAP на использование схемы `racf`.

Для поддержки RACF генерируемые пароли должны состоять не более чем из восьми символов.
Длину пароля можно настроить с помощью политики паролей:

```shell
d8 stronghold write ldap/config \
  binddn=$USERNAME \
  bindpass=$PASSWORD \
  url=ldaps://138.91.247.105 \
  schema=racf \
  password_policy=racf_password_policy
```

#### Active Directory

Для управления паролями в Active Directory настройте механизм секретов LDAP на использование схемы `ad`:

```shell
d8 stronghold write ldap/config \
  binddn=$USERNAME \
  bindpass=$PASSWORD \
  url=ldaps://138.91.247.105 \
  schema=ad
```

### Статические роли

{: #static-roles .anchored}

#### Настройка

Настройте статическую роль, которая сопоставляет имя в Stronghold с записью в LDAP.
Параметры ротации паролей будут управляться этой ролью.

```shell
d8 stronghold write ldap/static-role/lf-edge \
  dn='uid=lf-edge,ou=users,dc=lf-edge,dc=com' \
  username='stronghold' \
  rotation_period="24h"
```

Запросите учётные данные для роли `lf-edge`:

```shell
d8 stronghold read ldap/static-cred/lf-edge
```

### Ротация паролей

Управление паролями может выполняться двумя способами:

- автоматическая ротация по времени;
- ручная ротация.

### Автоматическая ротация паролей

Пароли автоматически меняются в соответствии со значением `rotation_period`, настроенным в статической роли.
Минимальное значение — `5s`.

При запросе учётных данных для статической роли в ответе будет указано время до следующей ротации в поле `ttl`.

В настоящее время автоматическая ротация поддерживается только для статических ролей.
Учётную запись `binddn`, которую использует Stronghold, нужно ротировать с помощью вызова `rotate-root`, чтобы сгенерировать пароль, известный только Stronghold.

### Ручная ротация

Пароли статической роли можно ротировать вручную с помощью вызова `rotate-role`.
При ручной ротации период ротации начинается заново.

### Удаление статических ролей

При удалении статической роли пароли не меняются.
Перед удалением роли или отзывом доступа к статической роли выполните ротацию пароля вручную.

### Динамические роли

{: #dynamic-roles .anchored}

#### Настройка

Динамическую роль можно настроить с помощью вызова `/role/:role_name`:

```shell
d8 stronghold write ldap/role/dynamic-role \
  creation_ldif=@/path/to/creation.ldif \
  deletion_ldif=@/path/to/deletion.ldif \
  rollback_ldif=@/path/to/rollback.ldif \
  default_ttl=1h \
  max_ttl=24h
```

{{< alert level="warning" >}}
Аргумент `rollback_ldif` необязателен, но рекомендуется.
Операции, указанные в `rollback_ldif`, будут выполнены, если создание завершится неудачей.
Это помогает гарантировать удаление всех объектов в случае ошибки.
{{< /alert >}}

Чтобы сгенерировать учётные данные, выполните команду:

```shell
d8 stronghold read ldap/creds/dynamic-role
```

Пример вывода:

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

Поле `distinguished_names` содержит массив DN, созданных на основе `creation_ldif`.
Если используется несколько записей LDIF, в это поле будут включены DN из каждой записи.
Каждый элемент этого поля соответствует одному LDIF-выражению.
Дедупликация не выполняется, а порядок элементов сохраняется.

### Записи LDIF

Управление учётными записями пользователей выполняется с помощью записей LDIF.
Записи LDIF могут представлять собой Base64-кодированную версию строки LDIF.
Строка будет разобрана и проверена на соответствие синтаксису LDIF.
Подробное описание синтаксиса LDIF доступно [в справочнике LDAP.com](https://ldap.com/ldif-the-ldap-data-interchange-format/).

При создании записей LDIF учитывайте следующее:

- В конце строк не должно быть пробелов.
- Каждый блок `modify` должен предваряться пустой строкой.
- Несколько модификаций для `dn` можно определить в одном блоке `modify`.
  Каждая модификация должна завершаться одним тире `-`.

### Active Directory

Для Active Directory есть несколько дополнительных особенностей.

Чтобы программно создать пользователя в AD, сначала выполните операцию `add` для объекта пользователя.
После этого выполните `modify` для этого пользователя, чтобы задать пароль и включить учётную запись.

Пароли в AD задаются через поле `unicodePwd`.
Перед значением должны стоять два двоеточия `::`.

При программной установке пароля в AD должны соблюдаться следующие требования:

- пароль должен быть заключён в двойные кавычки `""`;
- пароль должен быть в [формате `UTF16LE`](https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/6e803168-f140-4d23-b2d3-c3a8ab5917d2);
- пароль должен быть Base64-кодирован;
- дополнительные сведения доступны [в документации Microsoft](https://docs.microsoft.com/en-us/troubleshoot/windows-server/identity/set-user-password-with-ldifde).

После установки пароля пользователя его можно включить.
Для этого в AD используется поле `userAccountControl`:

- чтобы включить учётную запись, установите `userAccountControl` в `512`;
- чтобы отключить истечение срока действия пароля для этой динамической учётной записи пользователя, установите `userAccountControl` в `65536`;
- флаги `userAccountControl` являются кумулятивными, поэтому для установки обоих флагов задайте значение `66048` (`512 + 65536 = 66048`);
- подробная информация о флагах `userAccountControl` доступна [в документации Microsoft](https://docs.microsoft.com/en-us/troubleshoot/windows-server/identity/useraccountcontrol-manipulate-account-properties#property-flag-descriptions).

Поле `sAMAccountName` часто используется при работе с пользователями AD.
Оно предназначено для совместимости с устаревшими системами Microsoft Windows NT и ограничено 20 символами.
Учитывайте это при определении шаблона `username_template`.
Дополнительные сведения см. [в документации Microsoft](https://docs.microsoft.com/en-us/windows/win32/adschema/a-samaccountname).

Поскольку стандартный `username_template` длиннее 20 символов и соответствует шаблону `v_{{.DisplayName}}_{{.RoleName}}_{{random 10}}_{{unix_time}}`, рекомендуется настроить `username_template` в конфигурации роли так, чтобы генерируемые имена учётных записей были короче 20 символов.

AD не позволяет напрямую изменять атрибут `memberOf` пользователя.
Атрибут `member` группы и атрибут `memberOf` пользователя являются [связанными атрибутами](https://docs.microsoft.com/en-us/windows/win32/ad/linked-attributes).
Это пары вида «прямая ссылка/обратная ссылка», в которых можно изменять только прямую ссылку.
В случае членства в группе AD прямой ссылкой является атрибут `member` группы.
Чтобы добавить созданного динамического пользователя в группу, отправьте запрос `modify` в нужную группу и добавьте туда пользователя.

#### Пример LDIF для Active Directory

Параметры `*_ldif` представляют собой шаблоны, использующие язык [Go template](https://golang.org/pkg/text/template/).
Ниже приведён пример LDIF для создания учётной записи пользователя в Active Directory:

```text
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

## Ротация паролей для списков учётных записей

{: #rotation .anchored}

Stronghold может автоматически менять пароли для группы учётных записей.
Операцию ротации можно выполнить вручную, либо Stronghold выполнит её после истечения TTL с момента предыдущей смены.

Функциональность работает с разными схемами, включая OpenLDAP, Active Directory и RACF.
В следующем примере рассматривается сценарий с Active Directory.

Сначала включите механизм секретов LDAP и укажите параметры подключения к серверу AD.

Пример:

```console
$ d8 stronghold secrets enable ldap
Success! Enabled the ad secrets engine at: ldap/
$ d8 stronghold write ldap/config \
    binddn=$USERNAME \
    bindpass=$PASSWORD \
    url=ldaps://138.91.247.105 \
    userdn='dc=example,dc=com'
```

Затем настройте список учётных записей, для которых требуется выполнять ротацию пароля:

```console
d8 stronghold write ldap/library/accounting-team \
  service_account_names=fizz@example.com,buzz@example.com \
  ttl=10h \
  max_ttl=20h \
  disable_check_in_enforcement=false
```

В этом примере имена служебных учётных записей `fizz@example.com` и `buzz@example.com` уже созданы на удалённом сервере AD.
Параметр `ttl` задаёт время, через которое Stronghold повторно выполнит ротацию пароля учётной записи.
Параметр `max_ttl` задаёт максимальное время действия пароля после ротации.
По умолчанию оба параметра имеют значение `24h`.

По умолчанию служебная учётная запись должна быть зарегистрирована тем же субъектом Stronghold или клиентским токеном, который выполняет ротацию.
Если такое поведение вызывает проблемы, установите `disable_check_in_enforcement=true`.

После создания списка учётных записей можно в любой момент просмотреть их статус.

Пример:

```console
d8 stronghold read ldap/library/accounting-team/status
```

Пример вывода:

```console
Key                 Value
---                 -----
buzz@example.com    map[available:true]
fizz@example.com    map[available:true]
```

Чтобы выполнить ротацию паролей, используйте команду:

```console
d8 stronghold write -f ldap/library/accounting-team/check-out
```

Пример вывода:

```console
Key                     Value
---                     -----
lease_id                ldap/library/accounting-team/check-out/EpuS8cX7uEsDzOwW9kkKOyGW
lease_duration          10h
lease_renewable         true
password                ?@09AZKh03hBORZPJcTDgLfntlHqxLy29tcQjPVThzuwWAx/Twx4a2ZcRQRqrZ1w
service_account_name    fizz@example.com
```

Если стандартное значение `ttl` больше, чем требуется, установите меньшее значение:

```console
d8 stronghold write ldap/library/accounting-team/check-out ttl=30m
```

Пример вывода:

```console
Key                     Value
---                     -----
lease_id                ldap/library/accounting-team/check-out/gMonJ2jB6kYs6d3Vw37WFDCY
lease_duration          30m
lease_renewable         true
password                ?@09AZerLLuJfEMbRqP+3yfQYDSq6laP48TCJRBJaJu/kDKLsq9WxL9szVAvL/E1
service_account_name    buzz@example.com
```

Можно продлить аренду паролей для набора учётных записей:

```console
d8 stronghold lease renew ldap/library/accounting-team/check-out/0C2wmeaDmsToVFc0zDiX9cMq
```

Пример вывода:

```console
Key                Value
---                -----
lease_id           ldap/library/accounting-team/check-out/0C2wmeaDmsToVFc0zDiX9cMq
lease_duration     10h
lease_renewable    true
```

В этом случае текущие пароли для учётных записей будут действовать дольше, поскольку выполнение ротации будет отложено.

## Политика паролей LDAP

Механизм секретов LDAP не хеширует и не шифрует пароли перед изменением значений в LDAP.
Это может привести к тому, что пароли будут храниться в открытом виде.

Чтобы избежать хранения паролей в открытом виде, настройте на сервере LDAP политику паролей LDAP `ppolicy`.
Не путайте её с политикой паролей Stronghold.
Политика `ppolicy` может применять правила обработки паролей, например хеширование по умолчанию.

Ниже приведён пример политики паролей LDAP, которая включает хеширование для `dc=example,dc=com`:

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
