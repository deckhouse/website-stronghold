---
title: "Мультиарендность на основе пространств имён"
linkTitle: "Пространства имён для команд"
description: "Выделение каждой команде отдельного пространства имён с делегированным администрированием, собственными методами аутентификации и шаблонными политиками."
weight: 60
params:
  relatedLinks:
    - title: "Пространства имён"
      url: ../../../admin/namespaces/overview/
    - title: "Политики"
      url: ../../../concepts/policy/
    - title: "Идентификационные данные (Identity)"
      url: ../../../concepts/identity/
    - title: "Метод Kubernetes"
      url: ../../../user/auth/kubernetes/
---

{{< alert level="info" >}}
Пространства имён доступны только в Stronghold EE.
{{< /alert >}}

Пространство имён работает как отдельный виртуальный Stronghold со своими механизмами секретов, методами аутентификации, политиками и токенами. Команде выделяется пространство имён, а её администраторы управляют им самостоятельно, не получая доступа к корневому пространству и к пространствам других команд.

## Цель

Создать пространство имён для команды `team-a`, выдать её администраторам делегированные права, подключить отдельные методы аутентификации для людей и приложений и настроить шаблонные политики, которые не нужно менять при добавлении новых приложений.

## Предварительные требования

- Stronghold EE.
- Токен администратора корневого пространства имён.
- Параметры OIDC-клиента в IdP для команды (адрес, `client_id`, `client_secret`), если люди входят через OIDC.
- Кластер Kubernetes команды, если приложения аутентифицируются методом Kubernetes.

## Шаг 1. Создайте пространство имён

```bash
d8 stronghold namespace create -custom-metadata=owner="team-a@example.com" team-a
```

Все последующие команды выполняются в этом пространстве имён. Чтобы не указывать `-namespace=team-a` в каждой команде, задайте переменную окружения:

```bash
export STRONGHOLD_NAMESPACE=team-a
```

## Шаг 2. Создайте политику делегированного администратора

Пути в политике указываются относительно пространства имён. Политика разрешает управлять механизмами секретов, методами аутентификации, политиками и Identity внутри `team-a`, но не даёт прав за его пределами:

```bash
d8 stronghold policy write -namespace=team-a team-admin - <<'POLICY'
# Управление механизмами секретов и методами аутентификации.
path "sys/mounts/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
path "sys/mounts" {
  capabilities = ["read"]
}
path "sys/auth/*" {
  capabilities = ["create", "read", "update", "delete", "sudo"]
}
path "sys/auth" {
  capabilities = ["read"]
}

# Управление политиками.
path "sys/policies/acl/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Настройка методов аутентификации и Identity.
path "auth/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
path "identity/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Работа с секретами команды.
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Вложенные пространства имён.
path "sys/namespaces/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
POLICY
```

Уберите из политики блоки, которые команда не должна контролировать, например `sys/namespaces/*`.

## Шаг 3. Подключите метод аутентификации для людей

1. Включите OIDC в пространстве имён команды:

   ```bash
   d8 stronghold auth enable -namespace=team-a oidc
   ```

1. Настройте метод и роль по инструкции [«Метод OIDC»](../../../user/auth/oidc/overview/), передав в каждой команде `-namespace=team-a`. В роли укажите claim с группами пользователя (`groups_claim`).

1. Привяжите политику `team-admin` к группе администраторов команды в IdP:

   ```bash
   OIDC_ACCESSOR=$(d8 stronghold read -namespace=team-a -field=accessor sys/auth/oidc)

   GROUP_ID=$(d8 stronghold write -namespace=team-a -field=id identity/group \
     name=team-a-admins type=external policies=team-admin)

   d8 stronghold write -namespace=team-a identity/group-alias \
     name=team-a-admins \
     mount_accessor="$OIDC_ACCESSOR" \
     canonical_id="$GROUP_ID"
   ```

С этого момента администраторы команды входят командой `d8 stronghold login -namespace=team-a -method=oidc` и дальше работают без участия администратора корневого пространства.

## Шаг 4. Подключите метод аутентификации для приложений

1. Включите метод Kubernetes для кластера команды:

   ```bash
   d8 stronghold auth enable -namespace=team-a kubernetes
   d8 stronghold write -namespace=team-a auth/kubernetes/config \
     kubernetes_host="https://<api_server_address>:6443" \
     kubernetes_ca_cert=@ca.crt
   ```

   Параметры конфигурации описаны в разделе [«Метод Kubernetes»](../../../user/auth/kubernetes/).

1. Включите KV-хранилище команды:

   ```bash
   d8 stronghold secrets enable -namespace=team-a -path=secret -version=2 kv
   ```

## Шаг 5. Настройте шаблонную политику для приложений

Вместо отдельной политики на каждое приложение используйте одну шаблонную политику: приложение получает доступ к секретам в каталоге, совпадающем с его пространством имён Kubernetes.

```bash
K8S_ACCESSOR=$(d8 stronghold read -namespace=team-a -field=accessor sys/auth/kubernetes)

d8 stronghold policy write -namespace=team-a app-by-namespace - <<POLICY
path "secret/data/{{identity.entity.aliases.${K8S_ACCESSOR}.metadata.service_account_namespace}}/*" {
  capabilities = ["read"]
}
POLICY

d8 stronghold write -namespace=team-a auth/kubernetes/role/apps \
  bound_service_account_names="*" \
  bound_service_account_namespaces="team-a-*" \
  policies=app-by-namespace \
  ttl=1h
```

В `bound_service_account_namespaces` допустимы и шаблоны с `*` в конце (например, `team-a-*`).

Приложение из пространства имён Kubernetes `team-a-billing` прочитает `secret/data/team-a-billing/*`, но не секреты `team-a-orders`.

Аналогично можно выдать каждому сотруднику личный каталог:

```bash
d8 stronghold policy write -namespace=team-a personal - <<'POLICY'
path "secret/data/users/{{identity.entity.id}}/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "secret/metadata/users/{{identity.entity.id}}/*" {
  capabilities = ["read", "list", "delete"]
}
POLICY
```

## Проверка

1. Войдите как администратор команды и проверьте, что управление внутри пространства имён доступно:

   ```bash
   d8 stronghold login -namespace=team-a -method=oidc
   d8 stronghold secrets list -namespace=team-a
   ```

1. Проверьте, что доступ к корневому пространству закрыт — команда должна вернуть ошибку `permission denied`:

   ```bash
   d8 stronghold secrets list
   ```

1. Сравните возможности токена на путях разных команд:

   ```bash
   d8 stronghold token capabilities -namespace=team-a secret/data/team-a-billing/db
   ```

## Очистка

Удаление пространства имён удаляет все его механизмы секретов, методы аутентификации, политики и токены:

```bash
d8 stronghold namespace delete team-a
```

Перед удалением рабочего пространства имён сделайте [снимок хранилища](../../../admin/backups/save/).
