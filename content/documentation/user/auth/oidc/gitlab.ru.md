---
title: "OIDC провайдер Gitlab"
linkTitle: "Gitlab"
description: "Настройка входа в Stronghold через GitLab по OIDC: приложение в GitLab, конфигурация метода oidc, роль, сопоставление групп GitLab с группами Identity и проверка входа."
weight: 20
---

На этой странице описана настройка входа в Stronghold через GitLab по протоколу OIDC. Общие сведения о методе приведены в разделе [«Метод OIDC»](../../oidc/).

## Настройка приложения в GitLab

1. Перейдите в раздел **Настройки > Приложения** (**Settings > Applications**) пользователя, группы или экземпляра GitLab.
1. Заполните имя приложения и URI перенаправления (**Redirect URI**). Укажите по одному URI на строку:

   ```text
   https://stronghold.example.com/ui/stronghold/auth/oidc/oidc/callback
   http://localhost:8250/oidc/callback
   ```

   Первый URI используется для входа через веб-интерфейс, второй — для входа через CLI (`d8 stronghold login -method=oidc`). Если метод включён по другому пути, замените `oidc` в первом URI на этот путь.

1. Убедитесь, что выбрана область действия (scope) `openid`. Для передачи групп и имени пользователя также выберите `profile` и `email`.
1. Сохраните приложение и скопируйте идентификатор (**Application ID**) и секрет (**Secret**) клиента.

## Настройка Stronghold

1. Включите метод аутентификации OIDC:

   ```bash
   d8 stronghold auth enable oidc
   ```

   По умолчанию метод включается по пути `auth/oidc/`. Чтобы использовать другой путь, укажите `-path`.

1. Настройте подключение к GitLab. В `oidc_discovery_url` укажите адрес GitLab без `/.well-known/openid-configuration`:

   ```bash
   d8 stronghold write auth/oidc/config \
     oidc_discovery_url="https://gitlab.example.com" \
     oidc_client_id="<Application ID>" \
     oidc_client_secret="<Secret>" \
     default_role="gitlab"
   ```

   Параметр `default_role` задаёт роль, которая используется, если при входе роль не указана.

1. Создайте роль. В `allowed_redirect_uris` перечислите те же URI, что и в приложении GitLab:

   ```bash
   d8 stronghold write auth/oidc/role/gitlab -<<EOF
   {
     "role_type": "oidc",
     "allowed_redirect_uris": [
       "https://stronghold.example.com/ui/stronghold/auth/oidc/oidc/callback",
       "http://localhost:8250/oidc/callback"
     ],
     "oidc_scopes": ["openid", "profile", "email"],
     "user_claim": "sub",
     "groups_claim": "groups",
     "bound_claims": { "groups": ["devops", "security"] },
     "token_policies": ["default"],
     "token_ttl": "1h"
   }
   EOF
   ```

   Основные параметры роли:

   | Параметр | Описание |
   |----------|----------|
   | `allowed_redirect_uris` | Разрешённые URI перенаправления. Должны точно совпадать с URI в GitLab. |
   | `user_claim` | Утверждение (claim), значение которого становится именем псевдонима сущности. Для GitLab можно использовать `sub` или `nickname`. |
   | `groups_claim` | Утверждение со списком групп пользователя. Значения используются как имена псевдонимов групп Identity. |
   | `bound_claims` | Утверждения и значения, которые должны совпасть для успешного входа. В примере вход разрешён только участникам групп `devops` или `security`. |
   | `oidc_scopes` | Запрашиваемые области действия OIDC. |
   | `token_policies`, `token_ttl` | Политики и TTL выдаваемого токена. |

   <!-- TODO(verify): набор утверждений GitLab в ID-токене/userinfo (groups, nickname) для актуальной версии GitLab. -->

## Сопоставление групп GitLab с группами Stronghold

Чтобы назначать политики по членству в группах GitLab, создайте внешние группы Identity и псевдонимы групп. Подробнее о внешних группах — в разделе [«Идентичность»](../../../../concepts/identity/).

1. Создайте внешнюю группу с нужными политиками и сохраните её идентификатор:

   ```bash
   d8 stronghold write -field=id identity/group \
     name="gitlab-devops" \
     type="external" \
     policies="devops"
   ```

1. Получите accessor метода аутентификации `oidc/`:

   ```bash
   d8 stronghold auth list -format=json | jq -r '."oidc/".accessor'
   ```

1. Создайте псевдоним группы. Значение `name` должно совпадать с именем группы в утверждении `groups`:

   ```bash
   d8 stronghold write identity/group-alias \
     name="devops" \
     mount_accessor="<accessor метода oidc>" \
     canonical_id="<идентификатор группы>"
   ```

При каждом входе и продлении токена Stronghold обновляет членство сущности во внешних группах по данным GitLab.

## Проверка входа

1. Выполните вход через CLI:

   ```bash
   d8 stronghold login -method=oidc -path=oidc role=gitlab
   ```

   CLI откроет браузер для входа в GitLab и примет ответ на `http://localhost:8250/oidc/callback`.

1. Проверьте политики и метаданные полученного токена:

   ```bash
   d8 stronghold token lookup
   ```

1. Для проверки входа через веб-интерфейс выберите метод **OIDC**, при необходимости укажите роль `gitlab` и завершите вход в GitLab.

Если вход не удаётся, проверьте совпадение URI перенаправления и воспользуйтесь советами из раздела [«Метод OIDC»](../../oidc/).

## Примеры использования

Готовые примеры с этим механизмом:

- [Единый вход через корпоративный IdP (OIDC)](../../../../examples/access/sso-oidc/)

Все примеры собраны в разделе [«Примеры использования»](../../../../examples/).
