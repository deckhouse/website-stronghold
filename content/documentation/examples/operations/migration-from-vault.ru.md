---
title: "Миграция с HashiCorp Vault"
linkTitle: "Миграция с Vault"
description: "Перенос секретов, политик, методов аутентификации, PKI и динамических секретов из HashiCorp Vault в Stronghold с планом переключения клиентов и отката."
weight: 10
params:
  relatedLinks:
    - title: "Репликация KV1/KV2"
      url: ../../../admin/replication/kv-replication/
    - title: "Политики"
      url: ../../../concepts/policy/
    - title: "Механизм секретов KV"
      url: ../../../user/secrets-engines/kv/overview/
    - title: "Механизм секретов PKI"
      url: ../../../user/secrets-engines/pki/
    - title: "Пространства имён"
      url: ../../../admin/namespaces/overview/
---

Stronghold совместим с API HashiCorp Vault, поэтому миграция сводится к переносу конфигурации и данных и к переключению клиентов на новый адрес. Шифрованное хранилище Vault нельзя подключить к Stronghold напрямую: данные переносятся через API.

## Цель

Перенести в Stronghold секреты KV, политики, методы аутентификации, PKI и динамические секреты из работающего HashiCorp Vault, переключить клиентов и сохранить возможность отката.

## Предварительные требования

- Развёрнутый и распечатанный кластер Stronghold.
- Токен HashiCorp Vault с правами `read` и `list` на переносимые пути, а также на `sys/mounts`, `sys/auth` и `sys/policies/acl`.
- Токен Stronghold с административными правами.
- Сетевая связанность между Stronghold и Vault, если используется KV-репликация.
- Утилиты `vault`, `d8` и `jq` на рабочей станции.

В примерах ниже адреса задаются переменными окружения:

```bash
export VAULT_ADDR=https://vault.example.com:8200
export VAULT_TOKEN=<vault_token>
export STRONGHOLD_ADDR=https://stronghold.example.com
```

## Шаг 1. Проведите инвентаризацию

Составьте перечень объектов, которые нужно перенести. Выполните команды на стороне Vault:

```bash
vault secrets list -detailed -format=json > vault-mounts.json
vault auth list -detailed -format=json > vault-auth.json
vault policy list -format=json > vault-policies.json
vault namespace list -format=json > vault-namespaces.json
```

Команда `vault namespace list` доступна только в Vault Enterprise. Если в Vault используются пространства имён, повторите инвентаризацию для каждого из них с параметром `-namespace`.

Для каждого `mount` зафиксируйте тип (`kv` версии 1 или 2, `pki`, `database`, `transit` и другие), путь и параметры (`default_lease_ttl`, `max_lease_ttl`). Для каждого метода аутентификации — тип, путь, роли и привязанные политики.

{{< alert level="info" >}}
Пространства имён поддерживаются только в Stronghold EE. Если в Vault используются пространства имён, а целевая инсталляция — базовый Stronghold, перенесите содержимое каждого пространства имён в отдельные пути монтирования.
{{< /alert >}}

## Шаг 2. Воссоздайте пространства имён и механизмы секретов

1. Если нужны пространства имён (Stronghold EE), создайте их:

   ```bash
   d8 stronghold namespace create team-a
   ```

1. Включите механизмы секретов по тем же путям, что и в Vault. Например, для KV версии 2:

   ```bash
   d8 stronghold secrets enable -path=secret -version=2 kv
   ```

   Сохраняйте прежние пути монтирования: тогда политики и клиентский код не придётся менять.

## Шаг 3. Перенесите секреты KV

Выберите одну из стратегий.

### Стратегия A. Копирование скриптом

Подходит для базового Stronghold и для разового переноса. Скрипт обходит дерево секретов в Vault и записывает последнюю версию каждого секрета в Stronghold. История версий KV2 не переносится.

```bash
#!/usr/bin/env bash
set -euo pipefail

SRC_MOUNT=secret
DST_MOUNT=secret

copy_tree() {
  local path="$1"
  for key in $(vault kv list -mount="$SRC_MOUNT" -format=json "$path" | jq -r '.[]'); do
    if [[ "$key" == */ ]]; then
      copy_tree "${path}${key}"
    else
      vault kv get -mount="$SRC_MOUNT" -format=json "${path}${key}" \
        | jq '.data.data' > /tmp/secret.json
      d8 stronghold kv put -mount="$DST_MOUNT" "${path}${key}" @/tmp/secret.json
      echo "copied ${path}${key}"
    fi
  done
  rm -f /tmp/secret.json
}

copy_tree ""
```

Для KV версии 1 используйте поле `.data` вместо `.data.data`. Перед запуском убедитесь, что временный файл создаётся в каталоге, доступном только вам, либо замените его на передачу через стандартный ввод.

### Стратегия B. KV-репликация из внешнего Vault

{{< alert level="info" >}}
KV-репликация доступна только в Stronghold EE.
{{< /alert >}}

[Репликация KV1/KV2](../../../admin/replication/kv-replication/) работает по pull-модели: Stronghold периодически забирает секреты из источника через API. Источником может быть Vault CE, Vault Enterprise или другое хранилище с совместимым API KV/KV2. Такой вариант позволяет держать данные синхронизированными на всё время переключения клиентов.

1. На стороне Vault создайте политику и токен для чтения исходного хранилища:

   ```bash
   vault policy write replicate-secret - <<'POLICY'
   path "secret/*" {
     capabilities = ["read", "list"]
   }
   path "sys/mounts/secret" {
     capabilities = ["read"]
   }
   path "auth/token/lookup-self" {
     capabilities = ["read"]
   }
   path "auth/token/renew-self" {
     capabilities = ["update"]
   }
   POLICY

   vault token create -policy=replicate-secret -orphan=true -period=30d \
     -wrap-ttl=5m -field=wrapping_token
   ```

1. В Stronghold смонтируйте реплицируемое хранилище:

   ```bash
   d8 stronghold secrets enable \
     -path=secret \
     -src-address="$VAULT_ADDR" \
     -src-wrapping-token=<wrapping_token> \
     -src-mount-path=secret \
     -src-ca-cert=@vault-ca.pem \
     -sync-period-min=5 \
     -version=2 \
     kv
   ```

   Версии KV в источнике и в Stronghold должны совпадать.

1. Проверьте параметры репликации:

   ```bash
   d8 stronghold read sys/mounts/secret/tune
   ```

Пока репликация включена, хранилище в Stronghold работает только на чтение. После переключения клиентов отключите репликацию, чтобы разрешить запись:

```bash
d8 stronghold secrets tune -sync-enable=false secret
```

{{< alert level="warning" >}}
Не включайте репликацию повторно после отключения: все локальные изменения будут перезаписаны данными из источника.
{{< /alert >}}

## Шаг 4. Перенесите политики

Синтаксис ACL-политик совместим, поэтому политики переносятся без изменений:

```bash
for p in $(vault policy list | grep -v -E '^(root|default)$'); do
  vault policy read "$p" > "policy-$p.hcl"
  d8 stronghold policy write "$p" "policy-$p.hcl"
done
```

Проверьте политики, которые ссылаются на accessor методов аутентификации (шаблоны `{{identity.entity.aliases.<mount accessor>...}}`). После воссоздания методов аутентификации accessor изменится: замените его в политиках на новое значение. Подробнее — в разделе [«Шаблонные политики»](../../../concepts/policy/#шаблонные-политики).

## Шаг 5. Воссоздайте методы аутентификации

Конфигурацию методов аутентификации и их роли перенесите вручную или через IaC-инструмент: секретные параметры (например, `client_secret` OIDC, `bindpass` LDAP) Vault не возвращает при чтении.

1. Включите методы по прежним путям:

   ```bash
   d8 stronghold auth enable -path=oidc oidc
   d8 stronghold auth enable -path=kubernetes kubernetes
   d8 stronghold auth enable -path=approle approle
   ```

1. Прочитайте несекретные параметры ролей в Vault и запишите их в Stronghold:

   ```bash
   vault read -format=json auth/approle/role/my-app | jq '.data' > role-my-app.json
   d8 stronghold write auth/approle/role/my-app @role-my-app.json
   ```

1. Для AppRole выпустите новые `secret_id`: существующие значения из Vault перенести нельзя. `role_id` можно задать явно, записав его в `auth/approle/role/<role>/role-id`, чтобы не менять конфигурацию клиентов.

1. Для OIDC добавьте в список разрешённых redirect URI у провайдера идентификации адрес Stronghold.

Настройка методов описана в разделах [OIDC](../../../user/auth/oidc/overview/), [Kubernetes](../../../user/auth/kubernetes/), [AppRole](../../../user/auth/approle/), [LDAP](../../../user/auth/ldap/).

## Шаг 6. Перенесите PKI

Выберите вариант в зависимости от того, насколько важна непрерывность цепочки доверия.

- **Перевыпуск.** Создайте в Stronghold новый промежуточный CA, подпишите его существующим корневым CA и выпускайте новые сертификаты из Stronghold. Старые сертификаты продолжают действовать до истечения срока. Это предпочтительный вариант: закрытый ключ CA не покидает Vault. Процедура описана в руководстве [«Внутренний PKI»](../../certificates/internal-pki/).
- **Импорт CA.** Если ключ CA экспортируемый (сгенерирован как `exported` или хранится вне Vault), импортируйте пакет «сертификат + ключ» в Stronghold:

  ```bash
  d8 stronghold secrets enable -path=pki pki
  d8 stronghold write pki/issuers/import/bundle pem_bundle=@ca-bundle.pem
  ```

  Затем воссоздайте роли (`pki/roles/<name>`) и параметры `pki/config/urls`. Если адреса CRL и OCSP в выпущенных сертификатах указывают на Vault, сохраните доступность старых адресов до истечения этих сертификатов.

## Шаг 7. Перенастройте динамические секреты

Динамические учётные данные (`database`, `ldap`, `kubernetes`) не переносятся: они привязаны к арендам исходного кластера.

1. Воссоздайте конфигурацию подключений и роли в Stronghold. Для базы данных используйте отдельную служебную учётную запись, а не ту, что использует Vault, иначе ротация root-пароля в одной системе сломает другую.
1. Выполните `rotate-root` для новой учётной записи, чтобы её пароль знал только Stronghold:

   ```bash
   d8 stronghold write -force database/rotate-root/my-database
   ```

1. После переключения клиентов отзовите аренды в Vault: `vault lease revoke -prefix database/creds/`.

## Шаг 8. Переключите клиентов

Большинство клиентов Vault (CLI, SDK, Vault Agent, External Secrets Operator, Terraform-провайдер) определяют адрес сервера по переменной `VAULT_ADDR` или по параметру конфигурации. Укажите в них адрес Stronghold:

```bash
export VAULT_ADDR=https://stronghold.example.com
```

Для утилиты `d8 stronghold` используйте переменную `STRONGHOLD_ADDR`.

Переключайте клиентов поэтапно: сначала тестовые окружения, затем некритичные сервисы, затем остальные. Для сервисов с DNS-именем Vault можно переключить CNAME на Stronghold, если сертификат Stronghold содержит это имя в SAN.

<!-- TODO(verify): список клиентов экосистемы Vault, совместимость которых со Stronghold подтверждена тестами -->

## Проверка

1. Сравните количество секретов в Vault и Stronghold для каждого KV-хранилища.
1. Прочитайте несколько секретов через клиента приложения и через CLI:

   ```bash
   d8 stronghold kv get -mount=secret myapp/config
   ```

1. Выполните вход каждым перенесённым методом аутентификации и проверьте набор политик в выданном токене: `d8 stronghold token lookup`.
1. Выпустите тестовый сертификат и динамические учётные данные базы данных.
1. Убедитесь по журналам Vault, что к нему больше не обращаются рабочие клиенты.

## План отката

Держите Vault в рабочем состоянии, пока не завершится период наблюдения (обычно одна–две недели).

- При стратегии A не удаляйте данные в Vault. Если после переключения секреты менялись в Stronghold, перед откатом перенесите изменения обратно тем же скриптом, поменяв местами источник и приёмник.
- При стратегии B откат прост, пока репликация включена: Vault остаётся источником, а клиенты возвращаются на прежний адрес.
- Верните `VAULT_ADDR` клиентов или DNS-запись на Vault.
- Динамические учётные данные, выданные Stronghold, отзовите: `d8 stronghold lease revoke -prefix database/creds/`.

## Очистка

После завершения периода наблюдения:

1. Отключите KV-репликацию (`-sync-enable=false`) и отзовите токен репликации в Vault.
1. Удалите временные файлы с политиками и ролями (`policy-*.hcl`, `role-*.json`).
1. Выведите Vault из эксплуатации в соответствии с регламентом организации: сделайте финальный снимок, отзовите токены и аренды, остановите серверы.
