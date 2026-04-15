---
title: "LDAP"
linkTitle: "LDAP"
weight: 70
description: "Аутентификация в Deckhouse Stronghold через LDAP."
---

## LDAP

Метод `ldap` auth позволяет выполнять аутентификацию с использованием существующего LDAP-сервера и учётных данных пользователя — имени пользователя и пароля. Это позволяет интегрировать **Deckhouse Stronghold** в среды, использующие LDAP. Stronghold поддерживает интеграцию с различными LDAP-службами каталогов, включая [ALD Pro](https://www.aldpro.ru/).

Сопоставление групп и пользователей LDAP с политиками Stronghold выполняется через пути `users/` и `groups/`.

## Когда использовать LDAP

Метод `LDAP` обычно выбирают, если:

- в организации уже используется LDAP-каталог или Active Directory;
- нужно аутентифицировать пользователей по корпоративным учётным данным;
- требуется связать LDAP-группы с политиками Stronghold;
- необходимо интегрировать Stronghold в существующую directory-инфраструктуру.

## Аутентификация

### Через CLI

```bash
d8 stronghold login -method=ldap username=mitchellh
```

После этого Stronghold запросит пароль и при успешной аутентификации выдаст токен с соответствующими политиками.

### Через API

```bash
curl \
    --request POST \
    --data '{"password": "foo"}' \
    http://127.0.0.1:8200/v1/auth/ldap/login/mitchellh
```

Пример ответа:

```json
{
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": null,
  "auth": {
    "client_token": "c4f280f6-fdb2-18eb-89d3-589e2e834cdb",
    "policies": [
      "admins"
    ],
    "metadata": {
      "username": "mitchellh"
    },
    "lease_duration": 0,
    "renewable": false
  }
}
```

## Включение метода

Перед тем как пользователи смогут проходить аутентификацию, метод необходимо включить:

```bash
d8 stronghold auth enable ldap
```

## Базовая конфигурация

После включения нужно настроить подключение к LDAP-серверу, параметры поиска пользователя и способ определения членства в группах.

### Основные параметры подключения

Наиболее важные параметры:

- `url` — адрес LDAP-сервера;
- `starttls` — использовать ли `StartTLS`;
- `insecure_tls` — отключать ли проверку сертификата;
- `certificate` — CA-сертификат для проверки LDAP-сервера;
- `client_tls_cert` и `client_tls_key` — клиентский сертификат и ключ, если они требуются.

### Поиск пользователя

Для поиска и аутентификации пользователя могут использоваться:

- `binddn` и `bindpass` — для аутентифицированного поиска;
- `userdn` — базовый DN для поиска;
- `userattr` — атрибут LDAP-пользователя, соответствующий имени пользователя;
- `userfilter` — фильтр поиска.

Также поддерживается сценарий с `UPN` через параметр `upndomain`, особенно для Active Directory.

### Определение членства в группах

Для определения групп используются:

- `groupfilter`;
- `groupdn`;
- `groupattr`.

Это позволяет Stronghold определить, к каким LDAP-группам относится пользователь, и затем сопоставить эти группы с политиками Stronghold.

## Пример конфигурации

Ниже приведён пример настройки с `StartTLS` и поиском пользователей в `Active Directory`:

```bash
d8 stronghold write auth/ldap/config \
    url="ldap://ldap.example.com" \
    userdn="ou=Users,dc=example,dc=com" \
    groupdn="ou=Groups,dc=example,dc=com" \
    groupfilter="(&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))" \
    groupattr="cn" \
    upndomain="example.com" \
    certificate=@ldap_ca_cert.pem \
    insecure_tls=false \
    starttls=true
```

## Сопоставление групп LDAP и политик Stronghold

После настройки подключения можно связать LDAP-группы с политиками Stronghold.

Пример:

```bash
d8 stronghold write auth/ldap/groups/scientists policies=foo,bar
```

Это сопоставляет группу LDAP `scientists` с политиками Stronghold `foo` и `bar`.

Можно также добавить определённого LDAP-пользователя в дополнительную группу Stronghold и назначить ему отдельные политики:

```bash
d8 stronghold write auth/ldap/groups/engineers policies=foobar
d8 stronghold write auth/ldap/users/tesla groups=engineers policies=zoobar
```

## Проверка результата

После настройки можно выполнить вход:

```bash
d8 stronghold login -method=ldap username=tesla
```

Если всё настроено правильно, пользователь получит токен с политиками, соответствующими его LDAP-группам и дополнительным сопоставлениям.

## Важное замечание о сопоставлении политик

Сопоставление пользователь → политика происходит в момент создания токена. Если состав групп пользователя в LDAP изменился, это не повлияет на уже выданные токены.

Чтобы изменения вступили в силу, необходимо:

- отозвать старые токены;
- выполнить повторную аутентификацию.

## Блокировка пользователя

Для `LDAP` поддерживается механизм `user_lockout`. Если пользователь несколько раз подряд введёт неверные учётные данные, Stronghold на некоторое время прекратит попытки проверки и сразу начнёт возвращать отказ в доступе.

Значения по умолчанию:

- `lockout_threshold` — 5 попыток;
- `lockout_duration` — 15 минут;
- `lockout_counter_reset` — 15 минут.

> Предупреждение  
> Этот функционал поддерживается только методами `userpass`, `ldap` и `approle auth`.

## Практические рекомендации

- Для production используйте `ldaps` или `StartTLS`.
- Не включайте `insecure_tls` без крайней необходимости.
- Перед использованием сложных `userfilter` и `groupfilter` проверяйте, что поиск возвращает уникальный и ожидаемый результат.
- Если используется Active Directory, учитывайте особенности экранирования `DN`.
- После настройки обязательно проверьте, какие политики реально получает пользователь при входе.
