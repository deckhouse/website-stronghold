---
title: "Ротация паролей служебных учётных записей"
linkTitle: "Ротация статических учётных данных"
description: "Автоматическая смена паролей существующих служебных учётных записей в базах данных и LDAP через статические роли и библиотеки учётных записей."
weight: 60
params:
  relatedLinks:
    - title: "Механизм секретов баз данных"
      url: ../../../user/secrets-engines/databases/overview/
    - title: "Механизм секретов LDAP"
      url: ../../../user/secrets-engines/ldap/
    - title: "API механизмов секретов"
      url: ../../../reference/api/secrets/
    - title: "Stronghold Agent: шаблоны"
      url: ../../../user/agent/key-features/
---

Не всегда приложение может работать с динамическими учётными данными: некоторые системы требуют заранее созданной учётной записи с постоянным именем. В этом случае Stronghold берёт управление паролем существующей учётной записи на себя и меняет его по расписанию. Потребители всегда получают текущий пароль из Stronghold.

## Цель

Настроить автоматическую ротацию пароля служебной учётной записи PostgreSQL и учётной записи в LDAP или Active Directory, а также выдачу учётных записей из общего пула (библиотеки) с возвратом.

## Предварительные требования

- Токен Stronghold с правами на настройку механизмов секретов `database` и `ldap`.
- Настроенное подключение к PostgreSQL в механизме `database` (например, `database/config/myapp-postgres` из руководства [«Динамические учётные данные PostgreSQL»](../postgresql/)).
- Существующая учётная запись `billing_svc` в PostgreSQL.
- Сервер LDAP или Active Directory и учётная запись для Stronghold с правом менять пароли служебных учётных записей.

{{< alert level="warning" >}}
Не назначайте статической роли учётную запись, которую Stronghold использует для подключения в `config/`. После ротации пароль в `config/` станет недействительным, и все роли этого подключения перестанут работать. Для смены пароля служебной учётной записи Stronghold используйте `rotate-root`.
{{< /alert >}}

## Часть 1. Статическая роль в базе данных

1. Разрешите подключению использовать новую роль, добавив её в `allowed_roles`:

   ```bash
   d8 stronghold write database/config/myapp-postgres allowed_roles="myapp,billing-svc"
   ```

   При изменении (update) переданные параметры объединяются с уже сохранёнными, поэтому остальные параметры подключения повторно передавать не нужно.

1. Создайте статическую роль. Stronghold сразу меняет пароль учётной записи и дальше меняет его каждые 24 часа:

   ```bash
   d8 stronghold write database/static-roles/billing-svc \
     db_name="myapp-postgres" \
     username="billing_svc" \
     rotation_period=24h
   ```

1. Создайте политику для потребителей:

   ```bash
   d8 stronghold policy write billing-svc-creds - <<'POLICY'
   path "database/static-creds/billing-svc" {
     capabilities = ["read"]
   }
   POLICY
   ```

1. Получите текущие учётные данные:

   ```bash
   d8 stronghold read database/static-creds/billing-svc
   ```

   Поле `ttl` показывает время до следующей ротации, `last_vault_rotation` — время последней смены пароля.

1. При необходимости смените пароль немедленно, например при подозрении на компрометацию:

   ```bash
   d8 stronghold write -f database/rotate-role/billing-svc
   ```

Приложение должно запрашивать пароль заново после каждой ротации. Для приложений, которые читают пароль из файла, используйте [шаблоны Stronghold Agent](../../../user/agent/key-features/): Agent перерисует файл после ротации и выполнит команду перезагрузки.

## Часть 2. Статическая роль в LDAP

1. Включите механизм секретов и настройте подключение. Для Active Directory укажите `schema=ad`:

   ```bash
   d8 stronghold secrets enable ldap
   d8 stronghold write ldap/config \
     binddn="CN=stronghold,OU=Service,DC=example,DC=com" \
     bindpass="<password>" \
     url=ldaps://dc01.example.com \
     schema=ad
   ```

1. Смените пароль учётной записи Stronghold, чтобы он был известен только Stronghold:

   ```bash
   d8 stronghold write -f ldap/rotate-root
   ```

1. Создайте статическую роль для служебной учётной записи:

   ```bash
   d8 stronghold write ldap/static-role/svc-backup \
     dn="CN=svc-backup,OU=Service,DC=example,DC=com" \
     username="svc-backup" \
     rotation_period=24h
   ```

1. Получите текущий пароль:

   ```bash
   d8 stronghold read ldap/static-cred/svc-backup
   ```

1. Смените пароль вручную при необходимости:

   ```bash
   d8 stronghold write -f ldap/rotate-role/svc-backup
   ```

Механизм LDAP не хеширует пароль перед записью. Убедитесь, что на сервере LDAP настроена политика паролей, которая хранит пароли в хешированном виде. Подробнее — в разделе [«Механизм секретов LDAP»](../../../user/secrets-engines/ldap/).

## Часть 3. Библиотека учётных записей

Библиотека — это пул служебных учётных записей, которые выдаются во временное пользование. При возврате учётной записи Stronghold меняет её пароль, поэтому пароль, выданный предыдущему пользователю, перестаёт действовать.

1. Создайте библиотеку:

   ```bash
   d8 stronghold write ldap/library/ops-team \
     service_account_names="ops-svc1@example.com,ops-svc2@example.com" \
     ttl=4h \
     max_ttl=8h
   ```

1. Возьмите учётную запись:

   ```bash
   d8 stronghold write -f ldap/library/ops-team/check-out
   ```

   Ответ содержит `service_account_name`, `password` и `lease_id`.

1. Верните учётную запись после работы:

   ```bash
   d8 stronghold write ldap/library/ops-team/check-in \
     service_account_names="ops-svc1@example.com"
   ```

   Если учётную запись не вернуть, Stronghold вернёт её автоматически по истечении `ttl` аренды.

## Проверка

1. Прочитайте `database/static-creds/billing-svc`, выполните `rotate-role` и прочитайте учётные данные ещё раз — пароль должен измениться.
1. Подключитесь к PostgreSQL с новым паролем и убедитесь, что старый пароль отклоняется.
1. Проверьте статус библиотеки:

   ```bash
   d8 stronghold read ldap/library/ops-team/status
   ```

   Выданные учётные записи отмечены `available:false`.

## Очистка

Удаление статической роли не меняет пароль. Перед удалением смените пароль вручную, чтобы последнее значение, выданное Stronghold, перестало действовать:

```bash
d8 stronghold write -f database/rotate-role/billing-svc
d8 stronghold delete database/static-roles/billing-svc

d8 stronghold write -f ldap/rotate-role/svc-backup
d8 stronghold delete ldap/static-role/svc-backup

d8 stronghold delete ldap/library/ops-team
d8 stronghold policy delete billing-svc-creds
```

После удаления передайте управление паролями учётных записей их владельцам или отключите учётные записи.
