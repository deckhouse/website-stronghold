---
title: "Основные операции"
linkTitle: "Основные операции"
weight: 20
description: "Основные операции с managed key в Deckhouse Stronghold."
---

## Основные операции

В этом разделе описаны базовые операции для работы с `managed key` в Deckhouse Stronghold:

- регистрация managed key;
- просмотр списка и конфигурации ключей;
- проверка доступности ключа;
- удаление managed key;
- разрешение использования ключа в механизмах секретов.

{% alert level="info" %}
Перед выполнением операций убедитесь, что соответствующий backend уже подготовлен. Например, для `pkcs11` должна быть настроена `kms_library "pkcs11"` в конфигурации сервера Stronghold.
{% endalert %}

## Регистрация managed key типа `pkcs11`

Тип `pkcs11` используется для работы с HSM и PKCS#11-совместимыми библиотеками.

Для регистрации ключа используйте следующую команду:

```shell
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
- `key_label` — имя ключа во внешнем backend;
- `usages` — список разрешённых операций.

## Регистрация managed key типа `yandexcloudkms`

Тип `yandexcloudkms` используется для работы с Yandex Cloud KMS.

Для регистрации ключа с `oauth_token` используйте следующую команду:

```shell
stronghold write sys/managed-keys/yandexcloudkms/my-yc-key \
  kms_key_id=<kms_key_id> \
  oauth_token=<oauth_token> \
  usages=sign,verify
```

Для регистрации ключа с использованием ServiceAccount виртуальной машины используйте следующую команду:

```shell
stronghold write sys/managed-keys/yandexcloudkms/my-yc-key \
  kms_key_id=<kms_key_id> \
  usages=sign,verify
```

Если параметры `oauth_token` и `service_account_key_json` не указаны, Stronghold попытается использовать ServiceAccount виртуальной машины.

## Просмотр списка managed key

Чтобы получить список managed key заданного типа, используйте следующую команду:

```shell
stronghold list sys/managed-keys/pkcs11
```

Для `yandexcloudkms` используйте следующий путь:

```shell
stronghold list sys/managed-keys/yandexcloudkms
```

## Чтение конфигурации managed key

Чтобы прочитать конфигурацию конкретного ключа, используйте следующую команду:

```shell
stronghold read sys/managed-keys/pkcs11/my-hsm-key
```

Это полезно, если нужно:

- проверить, что ключ зарегистрирован;
- убедиться в корректности backend;
- сверить параметры и область применения ключа.

## Проверка доступности ключа

Перед подключением ключа к `PKI`, `SSH` или `Transit` рекомендуется проверить, что Stronghold действительно может использовать внешний ключ.

Для тестовой подписи используйте следующую команду:

```shell
stronghold write sys/managed-keys/pkcs11/my-hsm-key/test/sign
```

Эта операция помогает убедиться, что:

- Stronghold может обратиться к backend;
- ключ найден;
- права доступа и параметры аутентификации заданы корректно;
- операция подписи поддерживается.

## Удаление managed key

Если managed key больше не нужен, удалите его следующей командой:

```shell
stronghold delete sys/managed-keys/pkcs11/my-hsm-key
```

Перед удалением рекомендуется убедиться, что ключ больше не используется активными mount.

## Разрешение использования ключа для `PKI`

Для `PKI` secrets engine managed key нужно разрешить для конкретного mount path, если ключ не объявлен с `any_mount=true`.

Используйте следующую команду:

```shell
stronghold secrets tune -allowed-managed-keys=my-hsm-key pki/
```

Эта команда разрешает использовать managed key `my-hsm-key` в `PKI`-mount `pki/`.

## Использование с `Transit`

Для `Transit` managed key указывается при создании или ротации transit key.

Точный сценарий зависит от возможностей выбранного backend и типа ключа. Общий принцип такой: `Transit` использует внешний ключевой backend через managed key и не хранит приватный ключ внутри Stronghold.

## Использование с `SSH`

Для `SSH` secrets engine managed key может использоваться как внешний ключ CA для подписи SSH-сертификатов.

Это позволяет вынести ключ центра сертификации во внешний HSM или KMS и не хранить его внутри Stronghold.

## Практические рекомендации

- После регистрации managed key сразу выполняйте тестовую проверку через `test/sign`.
- Не назначайте `usages` шире, чем требуется вашему сценарию.
- Явно ограничивайте использование ключей по mount path, если не нужен режим `any_mount`.
- Перед удалением managed key убедитесь, что он больше не используется `PKI`, `SSH` или `Transit`.
- Для production-окружений документируйте соответствие между managed key в Stronghold и внешним ключом в HSM или KMS.
