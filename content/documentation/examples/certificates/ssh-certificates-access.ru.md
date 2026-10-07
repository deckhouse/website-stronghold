---
title: "Доступ по SSH с подписанными сертификатами"
linkTitle: "SSH-сертификаты"
description: "Настройка доступа к серверам по SSH-сертификатам с коротким сроком жизни, которые пользователь получает в Stronghold после входа через OIDC."
weight: 40
params:
  relatedLinks:
    - title: "Подписанные SSH-сертификаты"
      url: ../../../user/secrets-engines/signed-ssh-certificates/
    - title: "Метод OIDC"
      url: ../../../user/auth/oidc/overview/
    - title: "Шаблонные политики"
      url: ../../../concepts/policy/
    - title: "Идентификационные данные (Identity)"
      url: ../../../concepts/identity/
---

Вместо раздачи открытых ключей в `authorized_keys` на каждом сервере серверы доверяют одному CA в Stronghold. Пользователь входит в Stronghold через OIDC, подписывает свой открытый ключ и получает сертификат на 30 минут. Отзывать доступ на серверах не нужно: сертификат истекает сам.

## Цель

Настроить SSH CA в Stronghold, роль, которая разрешает пользователю входить только под своим именем, доверие к CA на серверах и получение сертификата после входа через OIDC.

## Предварительные требования

- Токен Stronghold с правами на настройку механизмов секретов, политик и Identity.
- Настроенный метод аутентификации OIDC, который передаёт группы пользователя (в DP — метод `oidc_deckhouse`).
- Серверы с OpenSSH, на которых вы можете менять `sshd_config`.
- Совпадение имени пользователя в IdP с именем учётной записи на серверах (или другое правило сопоставления, см. шаг 3).

## Шаг 1. Настройте SSH CA

1. Включите механизм секретов SSH:

   ```bash
   d8 stronghold secrets enable -path=ssh-client-signer ssh
   ```

1. Сгенерируйте ключ CA:

   ```bash
   d8 stronghold write ssh-client-signer/config/ca generate_signing_key=true
   ```

## Шаг 2. Настройте доверие на серверах

1. Сохраните открытый ключ CA на каждом сервере. Эндпоинт `public_key` не требует аутентификации:

   ```bash
   curl -o /etc/ssh/trusted-user-ca-keys.pem \
     https://stronghold.example.com/v1/ssh-client-signer/public_key
   ```

1. Добавьте в `/etc/ssh/sshd_config`:

   ```text
   TrustedUserCAKeys /etc/ssh/trusted-user-ca-keys.pem
   ```

1. Перезапустите `sshd`:

   ```bash
   sudo systemctl restart sshd
   ```

Распространяйте ключ и настройку через систему управления конфигурацией (Ansible, Puppet, Salt).

## Шаг 3. Создайте роль с шаблонным именем пользователя

Роль разрешает подписывать ключ только для principal, совпадающего с именем пользователя в OIDC. Для этого используется шаблон Identity с accessor метода OIDC.

1. Получите accessor метода аутентификации:

   ```bash
   OIDC_ACCESSOR=$(d8 stronghold read -field=accessor sys/auth/oidc_deckhouse)
   ```

1. Создайте роль:

   ```bash
   d8 stronghold write ssh-client-signer/roles/ops - <<EOF
   {
     "key_type": "ca",
     "algorithm_signer": "rsa-sha2-256",
     "allow_user_certificates": true,
     "allowed_users_template": true,
     "allowed_users": "{{identity.entity.aliases.${OIDC_ACCESSOR}.name}}",
     "default_user_template": true,
     "default_user": "{{identity.entity.aliases.${OIDC_ACCESSOR}.name}}",
     "allowed_extensions": "permit-pty,permit-port-forwarding",
     "default_extensions": {
       "permit-pty": ""
     },
     "ttl": "30m",
     "max_ttl": "1h"
   }
   EOF
   ```

   Если имя в IdP отличается от имени учётной записи на сервере (например, содержит домен), сохраните нужное имя в метаданных сущности и используйте шаблон `{{identity.entity.metadata.<key>}}`.

## Шаг 4. Выдайте права на подпись

1. Создайте политику:

   ```bash
   d8 stronghold policy write ssh-ops - <<'POLICY'
   path "ssh-client-signer/sign/ops" {
     capabilities = ["update"]
   }
   POLICY
   ```

1. Привяжите политику к группе IdP через внешнюю группу Identity:

   ```bash
   GROUP_ID=$(d8 stronghold write -field=id identity/group \
     name=ops type=external policies=ssh-ops)

   d8 stronghold write identity/group-alias \
     name=ops \
     mount_accessor="$OIDC_ACCESSOR" \
     canonical_id="$GROUP_ID"
   ```

   Здесь `name=ops` в `group-alias` — имя группы, которое IdP передаёт в claim групп.

## Шаг 5. Получите сертификат и войдите на сервер

Эти действия выполняет пользователь на своей рабочей станции.

1. Войдите в Stronghold:

   ```bash
   d8 stronghold login -method=oidc -path=oidc_deckhouse
   ```

1. Если SSH-ключа нет, создайте его:

   ```bash
   ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519
   ```

1. Подпишите открытый ключ и сохраните сертификат рядом с ключом:

   ```bash
   d8 stronghold write -field=signed_key ssh-client-signer/sign/ops \
     public_key=@$HOME/.ssh/id_ed25519.pub > ~/.ssh/id_ed25519-cert.pub
   ```

1. Подключитесь к серверу. OpenSSH автоматически использует файл `*-cert.pub`:

   ```bash
   ssh <username>@server.example.com
   ```

Повторяйте шаги 1 и 3 после истечения сертификата.

## Проверка

1. Просмотрите содержимое сертификата:

   ```bash
   ssh-keygen -Lf ~/.ssh/id_ed25519-cert.pub
   ```

   В поле `Principals` должно быть ваше имя пользователя, в поле `Valid` — интервал не длиннее 30 минут.

1. Попробуйте подписать ключ для чужого principal — запрос должен быть отклонён:

   ```bash
   d8 stronghold write ssh-client-signer/sign/ops \
     public_key=@$HOME/.ssh/id_ed25519.pub valid_principals=root
   ```

1. Через 30 минут убедитесь, что вход с тем же сертификатом отклоняется сервером.

При ошибках входа смотрите раздел [«Устранение неполадок»](../../../user/secrets-engines/signed-ssh-certificates/#устранение-неполадок).

## Очистка

1. Удалите внешнюю группу и её псевдоним: `d8 stronghold delete identity/group/name/ops`.
1. Удалите политику и роль:

   ```bash
   d8 stronghold policy delete ssh-ops
   d8 stronghold delete ssh-client-signer/roles/ops
   ```

1. Для тестового стенда отключите mount: `d8 stronghold secrets disable ssh-client-signer`. После этого удалите строку `TrustedUserCAKeys` на серверах.
