---
title: "Интеграция с ALD Pro"
linkTitle: "ALD Pro"
description: "Вход в Stronghold под учётными записями домена ALD Pro: метод LDAP с LDAPS и сопоставлением групп с политиками, единый вход по Kerberos, ротация паролей служебных учётных записей через механизм секретов LDAP и типовые ошибки."
weight: 10
params:
  relatedLinks:
    - title: "Метод LDAP"
      url: ../../../user/auth/ldap/
    - title: "Механизм секретов LDAP"
      url: ../../../user/secrets-engines/ldap/
    - title: "Идентификационные данные (Identity)"
      url: ../../../concepts/identity/
    - title: "API методов аутентификации"
      url: ../../../reference/api/auth/
    - title: "Ротация паролей служебных учётных записей"
      url: ../../dynamic-credentials/static-credentials-rotation/
    - title: "Интеграция с Active Directory"
      url: ../active-directory/
---

ALD Pro — российская служба каталогов на основе FreeIPA: данные пользователей и групп хранятся в 389 Directory Server, аутентификация выполняется через MIT Kerberos. Stronghold подключается к ALD Pro как к обычному LDAP-серверу, а при необходимости — как к центру распределения ключей Kerberos.

## Цель

- Пользователи домена ALD Pro входят в Stronghold под своими доменными учётными записями через метод `ldap`.
- Права в Stronghold назначаются по членству в группах ALD Pro.
- Пользователи с действующим билетом Kerberos входят без ввода пароля (метод `kerberos`, необязательно; только Stronghold EE).
- Пароли служебных учётных записей домена ротирует Stronghold через механизм секретов LDAP.

![Схема входа через LDAP в ALD Pro](../../../images/ex-ald-pro.png)

## Предварительные требования

- Развёрнутый домен ALD Pro, например `ald.example.com` с контроллерами `dc01.ald.example.com` и `dc02.ald.example.com`.
- Сетевой доступ от узлов Stronghold к контроллерам домена по порту 636 (LDAPS), а для Kerberos — по порту 88. <!-- TODO(verify): набор портов ALD Pro, открытых для внешних клиентов -->
- Сертификат корневого УЦ домена. На любом узле, введённом в домен, он находится в файле `/etc/ipa/ca.crt`.
- Токен Stronghold с правами на настройку методов аутентификации, механизмов секретов, политик и Identity.
- Утилиты `ldapsearch` и `jq` на рабочей станции администратора.

В примерах используется базовый DN `dc=ald,dc=example,dc=com`. Замените его на DN своего домена.

## Шаг 1. Подготовьте служебную учётную запись для поиска

Stronghold ищет пользователей и группы от имени отдельной учётной записи с правами только на чтение. Не используйте для этого учётные записи администраторов домена.

1. Создайте системную учётную запись (sysaccount) в ALD Pro. В FreeIPA это делается LDIF-файлом от имени `cn=Directory Manager`:

   ```text
   dn: uid=stronghold-bind,cn=sysaccounts,cn=etc,dc=ald,dc=example,dc=com
   changetype: add
   objectclass: account
   objectclass: simplesecurityobject
   uid: stronghold-bind
   userPassword: <пароль>
   passwordExpirationTime: 20380119031407Z
   nsIdleTimeout: 0
   ```

   ```bash
   ldapmodify -x -H ldaps://dc01.ald.example.com \
     -D "cn=Directory Manager" -W -f stronghold-bind.ldif
   ```

   <!-- TODO(verify): поддерживаемый в ALD Pro способ создания системной учётной записи (веб-интерфейс ALD Pro или LDIF в cn=sysaccounts,cn=etc) -->

   Если системные учётные записи в вашей инсталляции не используются, создайте обычного пользователя домена с бессрочным паролем и укажите его DN `uid=stronghold-bind,cn=users,cn=accounts,dc=ald,dc=example,dc=com`.

1. Скопируйте сертификат УЦ домена на рабочую станцию, с которой настраиваете Stronghold:

   ```bash
   scp admin@dc01.ald.example.com:/etc/ipa/ca.crt ./ald-ca.crt
   ```

1. Проверьте, что учётная запись видит пользователей и группы по LDAPS:

   ```bash
   LDAPTLS_CACERT=./ald-ca.crt ldapsearch -x -H ldaps://dc01.ald.example.com \
     -D "uid=stronghold-bind,cn=sysaccounts,cn=etc,dc=ald,dc=example,dc=com" -W \
     -b "cn=users,cn=accounts,dc=ald,dc=example,dc=com" "(uid=ivanov)" uid memberOf
   ```

   В выводе должны быть атрибут `uid` пользователя и список групп в `memberOf`.

## Шаг 2. Настройте метод LDAP

1. Включите метод аутентификации:

   ```bash
   d8 stronghold auth enable ldap
   ```

1. Настройте подключение к ALD Pro. Группы ищутся по атрибуту `member` объектов групп:

   ```bash
   d8 stronghold write auth/ldap/config \
     url="ldaps://dc01.ald.example.com:636,ldaps://dc02.ald.example.com:636" \
     certificate=@ald-ca.crt \
     insecure_tls=false \
     starttls=false \
     binddn="uid=stronghold-bind,cn=sysaccounts,cn=etc,dc=ald,dc=example,dc=com" \
     bindpass='<пароль>' \
     userdn="cn=users,cn=accounts,dc=ald,dc=example,dc=com" \
     userattr="uid" \
     groupdn="cn=groups,cn=accounts,dc=ald,dc=example,dc=com" \
     groupfilter="(&(objectClass=groupOfNames)(member={{.UserDN}}))" \
     groupattr="cn" \
     username_as_alias=true
   ```

   <!-- TODO(verify): objectClass групп пользователей в ALD Pro (groupOfNames / ipausergroup) -->

   Основные параметры:

   | Параметр | Значение для ALD Pro |
   | --- | --- |
   | `url` | Список контроллеров домена через запятую. Stronghold перебирает их по порядку при ошибке подключения. |
   | `certificate` | Сертификат УЦ домена из `/etc/ipa/ca.crt`. |
   | `insecure_tls` | `false` — проверка сертификата сервера обязательна. |
   | `starttls` | `false` для `ldaps://`. Если используется `ldap://` на порту 389, укажите `true`. |
   | `userdn`, `userattr` | Контейнер пользователей FreeIPA и атрибут входа `uid`. |
   | `groupdn`, `groupfilter`, `groupattr` | Контейнер групп и фильтр, который находит группы, где пользователь указан в `member`. Имя группы берётся из `cn`. |
   | `username_as_alias` | Имя псевдонима сущности совпадает с именем пользователя в домене. |

   Альтернативный вариант — читать группы из атрибута `memberOf` объекта пользователя:

   ```bash
   d8 stronghold write auth/ldap/config \
     groupdn="cn=users,cn=accounts,dc=ald,dc=example,dc=com" \
     groupfilter="(&(objectClass=person)(uid={{.Username}}))" \
     groupattr="memberOf"
   ```

   В этом случае в список групп попадут и роли, привилегии, правила HBAC и другие объекты, в которых состоит пользователь. Вариант с `member` возвращает только группы из `cn=groups,cn=accounts`. <!-- TODO(verify): состав memberOf у пользователей ALD Pro -->

   Команда `write` заменяет конфигурацию целиком. При изменении отдельных параметров повторите в команде все остальные.

1. Проверьте вход под доменной учётной записью:

   ```bash
   d8 stronghold login -method=ldap username=ivanov
   ```

   Имя пользователя указывается без домена. Если группам ещё не назначены политики, токен получит только политику `default`.

## Шаг 3. Сопоставьте группы ALD Pro с политиками

Используйте один из двух способов. Не назначайте политики одной и той же группе обоими способами — это усложняет аудит прав.

### Через группы метода LDAP

Имя группы в Stronghold должно совпадать со значением `cn` группы в ALD Pro:

```bash
d8 stronghold write auth/ldap/groups/stronghold-admins policies=admin
d8 stronghold write auth/ldap/groups/developers policies=dev-read
```

Способ подходит, если Stronghold использует только один метод входа.

### Через внешние группы Identity

Внешние группы позволяют назначить политики один раз и использовать их для нескольких методов входа (например, `ldap` и `kerberos`), а также применять к группе правила MFA.

1. Создайте внешнюю группу и сохраните её идентификатор:

   ```bash
   GROUP_ID=$(d8 stronghold write -field=id identity/group \
     name="ald-stronghold-admins" \
     type="external" \
     policies="admin")
   ```

1. Получите accessor метода `ldap/`:

   ```bash
   LDAP_ACCESSOR=$(d8 stronghold auth list -format=json | jq -r '."ldap/".accessor')
   ```

1. Создайте псевдоним группы. Значение `name` должно совпадать с именем группы в ALD Pro:

   ```bash
   d8 stronghold write identity/group-alias \
     name="stronghold-admins" \
     mount_accessor="$LDAP_ACCESSOR" \
     canonical_id="$GROUP_ID"
   ```

Stronghold обновляет членство сущности во внешних группах при каждом входе и продлении токена. Изменения групп в ALD Pro не влияют на уже выданные токены.

## Шаг 4. Настройте единый вход по Kerberos (необязательно)

{{< alert level="warning" >}}
Отдельного руководства по методу `kerberos` в документации Stronghold пока нет — параметры ниже приведены по [справочнику API методов аутентификации](../../../reference/api/auth/). Проверьте сценарий на тестовом стенде перед внедрением.
{{< /alert >}}

{{< alert level="info" >}}
Метод аутентификации `kerberos` (как и `saml`) доступен только в Stronghold EE.
{{< /alert >}}

Метод `kerberos` принимает заголовок SPNEGO от клиента с действующим билетом Kerberos, а группы пользователя получает из LDAP.

1. Создайте в ALD Pro сервисный принципал для HTTP-адреса Stronghold и получите keytab:

   ```bash
   kinit admin
   ipa service-add HTTP/stronghold.ald.example.com
   ipa-getkeytab -s dc01.ald.example.com \
     -p HTTP/stronghold.ald.example.com \
     -k stronghold.keytab
   ```

   <!-- TODO(verify): создание сервисного принципала и keytab средствами ALD Pro; нужна ли запись узла stronghold.ald.example.com в домене -->

1. Включите метод и передайте keytab в кодировке Base64:

   ```bash
   d8 stronghold auth enable kerberos
   base64 -w0 stronghold.keytab > stronghold.keytab.b64
   d8 stronghold write auth/kerberos/config \
     keytab=@stronghold.keytab.b64 \
     service_account="HTTP/stronghold.ald.example.com"
   ```

   <!-- TODO(verify): формат значения service_account для принципала FreeIPA; нужен ли remove_instance_name=true -->

1. Настройте поиск групп в LDAP. Параметры `auth/kerberos/config/ldap` совпадают с параметрами `auth/ldap/config`:

   ```bash
   d8 stronghold write auth/kerberos/config/ldap \
     url="ldaps://dc01.ald.example.com:636,ldaps://dc02.ald.example.com:636" \
     certificate=@ald-ca.crt \
     insecure_tls=false \
     binddn="uid=stronghold-bind,cn=sysaccounts,cn=etc,dc=ald,dc=example,dc=com" \
     bindpass='<пароль>' \
     userdn="cn=users,cn=accounts,dc=ald,dc=example,dc=com" \
     userattr="uid" \
     groupdn="cn=groups,cn=accounts,dc=ald,dc=example,dc=com" \
     groupfilter="(&(objectClass=groupOfNames)(member={{.UserDN}}))" \
     groupattr="cn"
   ```

1. Назначьте политики группам:

   ```bash
   d8 stronghold write auth/kerberos/groups/stronghold-admins policies=admin
   ```

   Если вы используете внешние группы Identity, создайте для каждой группы второй псевдоним с accessor метода `kerberos/` вместо этого шага.

1. Проверьте вход с машины, введённой в домен:

   ```bash
   kinit ivanov
   curl --negotiate -u : https://stronghold.ald.example.com/v1/auth/kerberos/login
   ```

   <!-- TODO(verify): вход по SPNEGO через curl и поддержка метода kerberos в d8 stronghold login и веб-интерфейсе -->

   В ответе в поле `auth.client_token` вернётся токен Stronghold.

## Шаг 5. Ротируйте пароли служебных учётных записей

Механизм секретов LDAP меняет пароли существующих учётных записей домена и выдаёт их приложениям. Для ALD Pro используйте схему `openldap`: пароль хранится в атрибуте `userPassword`.

{{< alert level="warning" >}}
В FreeIPA пароль, заданный другой учётной записью, по умолчанию помечается как истёкший, и пользователь должен сменить его при следующем входе. Учётная запись, от имени которой Stronghold меняет пароли, должна входить в список `passSyncManagersDNs` записи `cn=ipa_pwd_extop,cn=plugins,cn=config`, иначе ротированные пароли будут сразу истекать.
{{< /alert >}}

<!-- TODO(verify): поведение смены пароля и список passSyncManagersDNs в ALD Pro; необходимые права учётной записи stronghold-rotator на атрибут userPassword -->

1. Создайте в ALD Pro учётную запись `stronghold-rotator` с правом менять пароли целевых служебных учётных записей.

1. Включите механизм секретов и настройте подключение:

   ```bash
   d8 stronghold secrets enable ldap
   d8 stronghold write ldap/config \
     url="ldaps://dc01.ald.example.com:636" \
     certificate=@ald-ca.crt \
     insecure_tls=false \
     binddn="uid=stronghold-rotator,cn=users,cn=accounts,dc=ald,dc=example,dc=com" \
     bindpass='<пароль>' \
     userdn="cn=users,cn=accounts,dc=ald,dc=example,dc=com" \
     userattr="uid" \
     schema=openldap
   ```

1. Смените пароль учётной записи `stronghold-rotator`, чтобы его знал только Stronghold:

   ```bash
   d8 stronghold write -f ldap/rotate-root
   ```

   Получить новый пароль после ротации нельзя. Если Stronghold потеряет доступ к ALD Pro, пароль придётся сбросить средствами домена.

### Статическая роль

Статическая роль привязывает одну учётную запись домена к имени в Stronghold и меняет её пароль по расписанию:

```bash
d8 stronghold write ldap/static-role/svc-gitlab \
  dn="uid=svc-gitlab,cn=users,cn=accounts,dc=ald,dc=example,dc=com" \
  username="svc-gitlab" \
  rotation_period="24h"
```

Приложение читает текущий пароль:

```bash
d8 stronghold read ldap/static-cred/svc-gitlab
```

Поле `ttl` в ответе показывает время до следующей ротации. Внеплановая ротация:

```bash
d8 stronghold write -f ldap/rotate-role/svc-gitlab
```

### Библиотека учётных записей (check-out)

Библиотека выдаёт учётную запись из общего пула на время и меняет её пароль после возврата. Подходит для общих учётных записей, которыми пользуются дежурные инженеры.

```bash
d8 stronghold write ldap/library/ops-shared \
  service_account_names="svc-ops1,svc-ops2" \
  ttl=1h \
  max_ttl=4h
```

Значения `service_account_names` сравниваются с атрибутом из `userattr` (здесь `uid`), поэтому указывайте значение `uid`, а не полный DN.

Выдача, проверка статуса и возврат:

```bash
d8 stronghold write -f ldap/library/ops-shared/check-out
d8 stronghold read ldap/library/ops-shared/status
d8 stronghold write ldap/library/ops-shared/check-in service_account_names="svc-ops1"
```

Пример политики для приложения и дежурных инженеров:

```hcl
path "ldap/static-cred/svc-gitlab" {
  capabilities = ["read"]
}

path "ldap/library/ops-shared/check-out" {
  capabilities = ["update"]
}

path "ldap/library/ops-shared/check-in" {
  capabilities = ["update"]
}

path "ldap/library/ops-shared/status" {
  capabilities = ["read"]
}
```

## Проверка

1. Войдите под пользователем из группы `stronghold-admins` и проверьте политики токена:

   ```bash
   d8 stronghold login -method=ldap username=ivanov
   d8 stronghold token lookup
   ```

   В `policies` (или `identity_policies` при использовании внешних групп) должна быть политика `admin`.

1. Убедитесь, что пользователь не из группы получает только `default`.

1. Прочитайте пароль статической роли и выполните с ним `ldapsearch -D "uid=svc-gitlab,..."` — вход должен пройти успешно.

## Устранение неполадок

| Симптом | Вероятная причина | Что сделать |
| --- | --- | --- |
| `ldap operation failed: failed to bind as user` или `Invalid Credentials` при входе | Неверный пароль пользователя, неверный `userdn`/`userattr` или неверный пароль `binddn` | Проверьте поиск и bind командой `ldapsearch` из шага 1. Проверьте, что пользователь находится в `cn=users,cn=accounts`. |
| `x509: certificate signed by unknown authority` | В `certificate` не указан УЦ домена или указан не тот сертификат | Передайте `/etc/ipa/ca.crt` в `certificate`. Не включайте `insecure_tls`. |
| `x509: certificate is valid for ..., not ...` | Адрес в `url` не совпадает с именем в сертификате контроллера | Указывайте FQDN контроллеров, а не IP-адреса. |
| Вход проходит, но токен получает только `default` | `groupfilter` не находит группы или имена групп не совпадают с `auth/ldap/groups/<name>` и псевдонимами | Выполните `ldapsearch` с фильтром из `groupfilter`, подставив DN пользователя. Сравните `cn` групп с именами в Stronghold. |
| Пользователь не может войти после нескольких неудачных попыток, пароль верный | Сработала блокировка Stronghold (`lockout_threshold`, по умолчанию 5 попыток на 15 минут) или учётная запись заблокирована в ALD Pro | Дождитесь окончания `lockout_duration` или проверьте статус учётной записи в ALD Pro. Настройки блокировки описаны в разделе [«Метод LDAP»](../../../user/auth/ldap/#блокировка-пользователя). |
| Приложение не может войти с паролем из `static-cred` | Пароль помечен в ALD Pro как истёкший | Добавьте учётную запись Stronghold в `passSyncManagersDNs` и выполните `rotate-role`. |

## Очистка

```bash
d8 stronghold auth disable kerberos
d8 stronghold auth disable ldap
d8 stronghold secrets disable ldap
```

При отключении механизма секретов пароли учётных записей не меняются. Смените их средствами ALD Pro, а затем удалите учётные записи `stronghold-bind` и `stronghold-rotator`.
