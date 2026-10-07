---
title: "Аварийный доступ (break-glass)"
linkTitle: "Аварийный доступ"
description: "Подготовка аварийного доступа к Stronghold на случай недоступности провайдера идентификации: резервная учётная запись в запечатанном конверте, процедура generate-root, аудит использования и ротация после применения."
weight: 50
params:
  relatedLinks:
    - title: "Метод Userpass"
      url: ../../../user/auth/userpass/
    - title: "Токены"
      url: ../../../concepts/tokens/
    - title: "Seal и unseal"
      url: ../../../concepts/seal/
    - title: "Аудит"
      url: ../../../admin/audit/overview/
    - title: "Политики"
      url: ../../../concepts/policy/
---

Обычно люди входят в Stronghold через OIDC. Если провайдер идентификации (в DP — Dex и внешний IdP) недоступен или ошибочно настроен, никто из администраторов не сможет войти. Аварийный доступ — заранее подготовленный, редко используемый и строго контролируемый способ войти в такой ситуации.

## Цель

Подготовить два уровня аварийного доступа:

- **уровень 1** — отдельная учётная запись `userpass` с административной политикой, пароль к которой хранится в запечатанном конверте;
- **уровень 2** — процедура `generate-root` на случай, когда уровень 1 не помогает.

Настроить аудит использования и процедуру ротации после применения.

## Предварительные требования

- Токен Stronghold с административными правами.
- Хранилище для конвертов (сейф) и не менее двух ответственных за доступ к нему.
- Сведения о держателях долей ключа распечатывания или восстановления и пороге их кворума.
- Для аудита — Stronghold EE и настроенное устройство аудита (см. [«Аудит»](../../../admin/audit/overview/)).

## Шаг 1. Создайте аварийную политику

Политика должна позволять восстановить вход, но не обязана давать полный доступ к секретам. Пример политики для восстановления методов аутентификации и политик:

```bash
d8 stronghold policy write break-glass - <<'POLICY'
path "sys/auth" {
  capabilities = ["read"]
}
path "sys/auth/*" {
  capabilities = ["create", "read", "update", "delete", "sudo"]
}
path "auth/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
path "sys/policies/acl/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "identity/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "sys/mounts" {
  capabilities = ["read"]
}
POLICY
```

Расширяйте политику только тем, что действительно нужно при аварии.

## Шаг 2. Создайте аварийную учётную запись

1. Включите отдельный экземпляр метода `userpass`, чтобы его использование легко отличалось в журналах аудита:

   ```bash
   d8 stronghold auth enable -path=breakglass userpass
   ```

1. Сгенерируйте длинный случайный пароль на изолированной рабочей станции:

   ```bash
   openssl rand -base64 32
   ```

1. Создайте пользователя с коротким сроком жизни токена:

   ```bash
   d8 stronghold write auth/breakglass/users/breakglass-1 \
     password="<generated_password>" \
     token_policies=break-glass \
     token_ttl=1h \
     token_max_ttl=4h
   ```

1. Запишите пароль на бумаге, положите в конверт, запечатайте, подпишите и поместите в сейф. Для правила двух лиц разделите пароль на две части и храните их в разных конвертах у разных ответственных.

1. Удалите пароль из истории командной оболочки и буфера обмена.

Блокировка после неудачных попыток входа включена для `userpass` по умолчанию. Злоумышленник может намеренно заблокировать аварийную учётную запись; учитывайте это в процедуре и при необходимости создайте вторую учётную запись `breakglass-2` в отдельном конверте.

{{< alert level="warning" >}}
Не используйте для аварийного доступа долгоживущие токены. Срок жизни токена ограничен системным максимальным TTL, и такой токен может истечь незаметно.
{{< /alert >}}

## Шаг 3. Подготовьте процедуру generate-root

Если аварийная учётная запись недоступна, root-токен можно сгенерировать только с участием кворума держателей долей ключа распечатывания (при auto-unseal — ключа восстановления).

1. Инициатор запускает процедуру. В ответе возвращаются `nonce` и `otp`:

   ```bash
   d8 stronghold operator generate-root -init
   ```

1. Каждый держатель доли вносит её, указав `nonce`:

   ```bash
   d8 stronghold operator generate-root -nonce=<nonce>
   ```

   Команда запрашивает долю ключа интерактивно. Когда кворум набран, выводится `encoded_token`.

1. Инициатор расшифровывает токен:

   ```bash
   d8 stronghold operator generate-root -decode=<encoded_token> -otp=<otp>
   ```

Если процедура запущена по ошибке, отмените её командой `d8 stronghold operator generate-root -cancel`. Опишите процедуру в регламенте, укажите, кто держит доли и как их собрать.

{{< alert level="warning" >}}
В DP в режиме `Automatic` ключ распечатывания и root-токен хранятся в секрете `stronghold-keys` пространства имён `d8-stronghold`. Доступ к этому секрету равнозначен полному доступу к Stronghold: ограничьте его через RBAC и контролируйте по журналам аудита Kubernetes. Подробнее — в разделе [«Настройка»](../../../install/dkp/configuration/).
{{< /alert >}}

<!-- TODO(verify): порядок generate-root и отзыва исходного root-токена для Stronghold в DKP в режиме Automatic -->

## Шаг 4. Настройте контроль использования

{{< alert level="info" >}}
Аудит-логирование доступно только в Stronghold EE.
{{< /alert >}}

1. Убедитесь, что устройство аудита включено:

   ```bash
   d8 stronghold audit list
   ```

1. Настройте в SIEM оповещение о любом входе через аварийный путь. Признаки в журнале аудита:

   - `request.path` равен `auth/breakglass/login/breakglass-1`;
   - `request.path` начинается с `sys/generate-root`;
   - в `auth.policies` присутствует `root`.

   Пример поиска по файловому журналу:

   ```bash
   jq -c 'select(.request.path | startswith("auth/breakglass/login") or startswith("sys/generate-root"))' \
     /var/log/stronghold_audit.log
   ```

## Использование при аварии

1. Два ответственных вскрывают конверт и фиксируют время и причину в журнале инцидента.
1. Войдите:

   ```bash
   d8 stronghold login -method=userpass -path=breakglass username=breakglass-1
   ```

1. Восстановите конфигурацию входа (например, `auth/oidc_deckhouse/config`) и проверьте, что обычный вход работает.
1. Отзовите аварийный токен: `d8 stronghold token revoke -self`.

## Ротация после применения

После каждого использования, даже учебного:

1. Смените пароль и создайте новый конверт:

   ```bash
   d8 stronghold write auth/breakglass/users/breakglass-1/password password="<new_password>"
   ```

1. Отзовите все токены, выданные через аварийный путь:

   ```bash
   d8 stronghold lease revoke -prefix auth/breakglass/
   ```

1. Если использовался `generate-root`, отзовите root-токен сразу после завершения работ:

   ```bash
   d8 stronghold token revoke <root_token>
   ```

1. Проанализируйте журнал аудита за период использования и приложите выгрузку к отчёту об инциденте.

## Проверка

Проводите учебное применение не реже раза в полгода:

1. Войдите аварийной учётной записью на тестовом или рабочем кластере.
1. Проверьте набор прав: `d8 stronghold token capabilities sys/auth/oidc_deckhouse`.
1. Убедитесь, что в SIEM пришло оповещение.
1. Выполните ротацию после применения.

## Очистка

Если аварийный доступ больше не нужен (например, при выводе кластера из эксплуатации):

```bash
d8 stronghold auth disable breakglass
d8 stronghold policy delete break-glass
```

Уничтожьте конверты с паролями по акту.
