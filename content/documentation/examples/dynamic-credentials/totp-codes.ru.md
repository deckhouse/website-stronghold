---
title: "Одноразовые коды TOTP для приложения"
linkTitle: "Коды TOTP"
description: "Механизм секретов TOTP: создание ключа для пользователя приложения, получение и проверка кода, защита от повторного использования кода."
weight: 70
params:
  relatedLinks:
    - title: "Механизм секретов TOTP"
      url: ../../../user/secrets-engines/totp/
    - title: "MFA для администраторов"
      url: ../../access/admin-mfa/
---

Механизм секретов TOTP позволяет приложению проверять одноразовые коды второго фактора своих пользователей, не храня ключи у себя. Это отдельная возможность: она не связана с MFA при входе в сам Stronghold (см. [«MFA для администраторов»](../../access/admin-mfa/)).

## Цель

Создать ключ TOTP для пользователя приложения, получить код и проверить его, а также убедиться, что один и тот же код нельзя использовать дважды.

## Предварительные требования

- Токен Stronghold с правами на включение механизмов секретов и запись в `totp/*`.

## Шаг 1. Включите механизм и создайте ключ

```bash
d8 stronghold secrets enable totp

d8 stronghold write -format=json totp/keys/alice \
  generate=true \
  issuer="Example" \
  account_name="alice@example.com" \
  period=30 \
  algorithm=SHA1 \
  digits=6 | jq -r '.data.url'
```

Команда возвращает URL вида `otpauth://totp/...`. Приложение показывает его пользователю в виде QR-кода (в ответе также есть поле `barcode` с изображением в base64), и пользователь добавляет ключ в приложение-аутентификатор.

Основные параметры ключа:

| Параметр | Назначение |
| --- | --- |
| `generate` | Если `true`, Stronghold сам создаёт секрет. Иначе нужно передать `key` или `url` готового ключа |
| `issuer`, `account_name` | Название сервиса и учётная запись, которые показывает аутентификатор |
| `period` | Период смены кода в секундах |
| `algorithm` | `SHA1`, `SHA256` или `SHA512` |
| `digits` | Длина кода: `6` или `8` |
| `skew` | Допустимое отклонение по времени в периодах: `0` или `1` |
| `exported` | Если `false`, URL и QR-код после создания ключа не возвращаются |

## Шаг 2. Выдайте приложению минимальные права

Приложению нужно только проверять коды:

```bash
d8 stronghold policy write totp-verifier - <<'POLICY'
path "totp/code/*" {
  capabilities = ["update"]
}
POLICY
```

Право `read` на `totp/code/*` позволяет получать сами коды, поэтому приложению его не давайте.

## Шаг 3. Получите и проверьте код

1. Получите текущий код. Так вы имитируете аутентификатор пользователя:

   ```bash
   CODE=$(d8 stronghold read -field=code totp/code/alice)
   ```

1. Проверьте код, как это делает приложение:

   ```bash
   d8 stronghold write totp/code/alice code="$CODE"
   ```

   В ответе `valid` равен `true`.

1. Проверьте неверный код:

   ```bash
   d8 stronghold write totp/code/alice code=000000
   ```

   В ответе `valid` равен `false`.

## Шаг 4. Убедитесь, что код одноразовый

Повторная проверка уже использованного кода отклоняется:

```bash
# повторная проверка того же кода
d8 stronghold write totp/code/alice code="$CODE"
```

Stronghold отвечает ошибкой `code already used; wait until the next time period`. Следующий код можно проверить после смены периода.

## Проверка

```bash
d8 stronghold list totp/keys
d8 stronghold read totp/keys/alice
```

Список содержит `alice`, а чтение ключа возвращает параметры (`issuer`, `period`, `algorithm`, `digits`) без секрета.

## Очистка

```bash
d8 stronghold delete totp/keys/alice
d8 stronghold policy delete totp-verifier
d8 stronghold secrets disable totp
```
