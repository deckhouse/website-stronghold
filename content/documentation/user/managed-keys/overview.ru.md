---
title: "Управляемые ключи (Managed Keys) в Stronghold"
linkTitle: "Введение"
weight: 10
description: "Обзор Managed Keys в Stronghold: поддерживаемые backend-ы и механизмы секретов, которые умеют их использовать."
---

`Managed Keys` позволяют Stronghold использовать криптографические ключи, расположенные во внешней доверенной системе, например в HSM или внешнем KMS, без хранения приватного ключевого материала внутри самого Stronghold.

Такой подход полезен в сценариях, где:

- приватные ключи должны храниться вне Stronghold;
- операции подписи или шифрования должны выполняться внешней системой;
- есть требования по безопасности или соответствию, запрещающие экспорт ключевого материала.

В Stronghold managed key представляет собой именованную запись, управляемую через API `sys/managed-keys/<type>/<name>`.

## Как это работает

Stronghold хранит конфигурацию доступа к внешнему ключу, но не сам приватный ключ. Когда движку секретов требуется операция подписи, проверки подписи, шифрования или расшифрования, он обращается к соответствующему managed key, а уже тот делегирует операцию внешнему backend-у.

Это означает, что:

- ключевой материал остается во внешней системе;
- Stronghold использует managed key как абстракцию поверх внешнего backend-а;
- несколько managed keys могут быть настроены для одного и того же backend-а, если нужны разные ключи или разные политики доступа.

## Поддерживаемые backend-ы

В Stronghold поддерживаются следующие типы managed keys:

- `pkcs11`
- `yandexkms`

### `pkcs11`

Тип `pkcs11` используется для работы с HSM и PKCS#11-совместимыми библиотеками. Конфигурация managed key этого типа ссылается на ранее объявленную `kms_library "pkcs11"` в серверной конфигурации Stronghold.

Типовой сценарий:

- настроить `kms_library "pkcs11"` в конфигурации сервера;
- зарегистрировать managed key через `sys/managed-keys/pkcs11/<name>`;
- при необходимости проверить доступность ключа через `test/sign`;
- разрешить использование ключа нужному mount-path.

Пример регистрации:

```shell
stronghold write sys/managed-keys/pkcs11/my-hsm-key \
  library=softhsm \
  token_label=managed-keys \
  pin=1234 \
  key_label=my-signing-key \
  usages=sign,verify
```

### `yandexkms`

Тип `yandexkms` используется для интеграции с Yandex Cloud KMS.

Для конфигурации такого managed key используются параметры:

- `kms_key_id`
- `oauth_token` или `service_account_key_json`
- `endpoint` при необходимости

Для аутентификации Stronghold в Yandex Cloud KMS можно использовать один из следующих вариантов:

- `oauth_token`;
- `service_account_key_json`;
- сервисный аккаунт виртуальной машины в Yandex Cloud.

Если параметры `oauth_token` и `service_account_key_json` не заданы, Stronghold пытается использовать сервисный аккаунт виртуальной машины через стандартный механизм получения instance credentials. Это поведение аналогично типовым сценариям в Vault, где для облачных KMS могут использоваться учетные данные самой виртуальной машины.

Пример регистрации:

```shell
stronghold write sys/managed-keys/yandexkms/my-yc-key \
  kms_key_id=<kms_key_id> \
  oauth_token=<oauth_token> \
  usages=sign,verify
```

Пример использования сервисного аккаунта виртуальной машины:

```shell
stronghold write sys/managed-keys/yandexkms/my-yc-key \
  kms_key_id=<kms_key_id> \
  usages=sign,verify
```

{{< alert level="info" >}}
Параметры `oauth_token` и `service_account_key_json` взаимоисключающие. Если указан один из них, именно он будет использоваться для аутентификации в Yandex Cloud.
{{< /alert >}}

{{< alert level="info" >}}
Для `yandexkms` набор допустимых операций зависит от самого ключа и его конфигурации. Если managed key будет использоваться для `pki` или `ssh`, ему требуется возможность подписи. Для `transit` набор операций зависит от выбранного сценария использования.
{{< /alert >}}

## Какие механизмы секретов поддерживают Managed Keys

В Stronghold managed keys могут использовать:

- `SSH`
- `PKI`
- `Transit`

### SSH

SSH secrets engine может использовать managed key как ключ центра сертификации (CA) и применять его для подписи SSH-сертификатов.

Это полезно, если SSH CA должен оставаться во внешнем HSM или внешнем KMS.

### PKI

PKI secrets engine может использовать managed key для генерации и обслуживания корневых и промежуточных CA, а также для подписи сертификатов.

Это один из основных сценариев использования managed keys в Stronghold. Для PKI дополнительно важно разрешить managed key для конкретного PKI mount-а, если ключ не объявлен с `any_mount=true`.

Пример:

```shell
stronghold secrets tune -allowed-managed-keys=my-hsm-key pki/
```

### Transit

Transit secrets engine может использовать managed key как внешний криптографический ключ.

В зависимости от возможностей backend-а и типа ключа это позволяет выполнять:

- подпись;
- проверку подписи;
- шифрование;
- расшифрование.

Для Transit managed key указывается при создании или ротации transit key.

## Пространства имен и область видимости

Managed key привязан к конкретному namespace. Механизм секретов, который использует этот ключ, должен находиться в том же namespace, что и сам managed key.

Если ключ не объявлен с `any_mount=true`, его использование нужно явно разрешить для конкретного mount-path.

## Базовые операции

Список managed keys заданного типа:

```shell
stronghold list sys/managed-keys/pkcs11
```

Чтение конфигурации ключа:

```shell
stronghold read sys/managed-keys/pkcs11/my-hsm-key
```

Проверка доступности ключа через тестовую подпись:

```shell
stronghold write sys/managed-keys/pkcs11/my-hsm-key/test/sign
```

Удаление managed key:

```shell
stronghold delete sys/managed-keys/pkcs11/my-hsm-key
```

## Практические рекомендации

- Используйте `pkcs11`, если ключи находятся в локальном или сетевом HSM с PKCS#11-библиотекой.
- Используйте `yandexkms`, если ключи управляются в Yandex Cloud KMS.
- Ограничивайте `usages` только необходимыми операциями.
- Если ключ не должен быть доступен всем mount-ам, не включайте `any_mount` и явно настраивайте `allowed-managed-keys`.
- Перед подключением ключа к `pki`, `ssh` или `transit` проверяйте его через `test/sign` или другой подходящий тестовый сценарий.

## Связанные разделы

- [Механизм секретов PKI](../secrets-engines/pki/)
- [Механизм секретов Transit](../secrets-engines/transit/)
- [Подписанные SSH-сертификаты](../secrets-engines/signed-ssh-certificates/)
