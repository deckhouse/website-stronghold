---
title: "Основные операции"
linkTitle: "Основные операции"
weight: 20
description: "Основные операции с Managed Keys в Deckhouse Stronghold."
---

## Основные операции

В этом разделе приведены базовые операции для работы с `Managed Keys` в **Deckhouse Stronghold**:

- регистрация managed key;
- просмотр списка и конфигурации ключей;
- проверка доступности ключа;
- удаление managed key;
- разрешение использования ключа в механизмах секретов.

> Примечание  
> Перед выполнением операций убедитесь, что соответствующий backend уже подготовлен. Например, для `pkcs11` должна быть настроена `kms_library "pkcs11"` в конфигурации сервера Stronghold.

## Регистрация managed key типа pkcs11

Тип `pkcs11` используется для работы с HSM и `PKCS#11`-совместимыми библиотеками.

Пример регистрации:

```bash
stronghold write sys/managed-keys/pkcs11/my-hsm-key \
  library=softhsm \
  token_label=managed-keys \
  pin=1234 \
  key_label=my-signing-key \
  usages=sign,verify
```

В этом примере:

- `my-hsm-key` — имя managed key в Stronghold;
- `library` — имя ранее объявленной `kms_library`;
- `token_label` — метка токена во внешнем HSM;
- `pin` — PIN для доступа;
- `key_label` — имя ключа во внешнем backend'е;
- `usages` — список разрешённых операций.

## Регистрация managed key типа yandexcloudkms

Тип `yandexcloudkms` используется для работы с **Yandex Cloud KMS**.

Пример регистрации с `oauth_token`:

```bash
stronghold write sys/managed-keys/yandexcloudkms/my-yc-key \
  kms_key_id=<kms_key_id> \
  oauth_token=<oauth_token> \
  usages=sign,verify
```

Пример регистрации с использованием сервисного аккаунта виртуальной машины:

```bash
stronghold write sys/managed-keys/yandexcloudkms/my-yc-key \
  kms_key_id=<kms_key_id> \
  usages=sign,verify
```

Если параметры `oauth_token` и `service_account_key_json` не указаны, Stronghold пытается использовать сервисный аккаунт виртуальной машины.

## Просмотр списка managed keys

Чтобы получить список managed keys заданного типа, используйте:

```bash
stronghold list sys/managed-keys/pkcs11
```

Аналогично можно использовать путь для `yandexcloudkms`:

```bash
stronghold list sys/managed-keys/yandexcloudkms
```

## Чтение конфигурации managed key

Чтобы прочитать конфигурацию конкретного ключа, используйте:

```bash
stronghold read sys/managed-keys/pkcs11/my-hsm-key
```

Это полезно, если нужно:

- проверить, что ключ зарегистрирован;
- убедиться в корректности backend'а;
- сверить параметры и область применения ключа.

## Проверка доступности ключа

Перед подключением ключа к `PKI`, `SSH` или `Transit` рекомендуется проверить, что Stronghold действительно может использовать внешний ключ.

Пример тестовой подписи:

```bash
stronghold write sys/managed-keys/pkcs11/my-hsm-key/test/sign
```

Эта операция помогает убедиться, что:

- Stronghold может обратиться к backend'у;
- ключ найден;
- права доступа и параметры аутентификации заданы корректно;
- операция подписи поддерживается.

## Удаление managed key

Если managed key больше не нужен, его можно удалить:

```bash
stronghold delete sys/managed-keys/pkcs11/my-hsm-key
```

Перед удалением рекомендуется убедиться, что ключ больше не используется активными `mount`'ами.

## Разрешение использования ключа для PKI

Для `PKI` secrets engine managed key нужно разрешить для конкретного `mount path`, если ключ не объявлен с `any_mount=true`.

Пример:

```bash
stronghold secrets tune -allowed-managed-keys=my-hsm-key pki/
```

Это разрешает использовать managed key `my-hsm-key` в `PKI`-mount'е `pki/`.

## Использование с Transit

Для `Transit` managed key указывается при создании или ротации transit key.

Точный сценарий зависит от возможностей выбранного backend'а и типа ключа, но общий смысл один: `Transit` использует внешний ключевой backend через managed key и не хранит приватный ключ внутри Stronghold.

## Использование с SSH

Для `SSH` secrets engine managed key может использоваться как внешний ключ `CA` для подписи SSH-сертификатов.

Это позволяет вынести ключ центра сертификации во внешний HSM или KMS и не хранить его внутри Stronghold.

## Практические рекомендации

- После регистрации managed key сразу выполняйте тестовую проверку через `test/sign`.
- Не назначайте `usages` шире, чем требуется вашему сценарию.
- Явно ограничивайте использование ключей по `mount path`, если не нужен режим `any_mount`.
- Перед удалением managed key убедитесь, что он больше не используется `PKI`, `SSH` или `Transit`.
- Для production-окружений документируйте соответствие между managed key в Stronghold и внешним ключом в HSM или KMS.

## Что дальше

После регистрации и проверки managed key вы можете перейти к его использованию в конкретных механизмах секретов:

- [PKI](../secrets-engines/pki/)
- [Transit](../secrets-engines/transit/)
- [SSH](../secrets-engines/ssh/)
