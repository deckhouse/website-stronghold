---
title: "Поддержка HSM"
weight: 10
params:
  edition: ee
---

Stronghold поддерживает шифрование root-ключа с использованием аппаратных модулей защиты (Hardware Security Module, HSM), таких как TPM2, Rutoken ЭЦП 3.0, JaCarta и другие устройства с поддержкой стандарта PKCS #11.

Для тестирования и разработки можно также использовать программный эмулятор SoftHSM2.

{{< alert level="warning" >}}
В текущей версии работа с HSM поддерживается только при standalone-установке Stronghold. Примеры на этой странице предполагают использование локального конфигурационного файла и блок параметров `seal "pkcs11"` в конфигурации standalone-сервера.
{{< /alert >}}

Для автоматического распечатывания Stronghold с помощью PKCS #11 предварительно создайте ключи в HSM и настройте Stronghold для работы с ними.

## SoftHSM2

Для тестирования интеграции Stronghold с HSM можно использовать программный эмулятор SoftHSM2.
Выполните следующие действия, чтобы создать токен, сгенерировать в нём ключевую пару
и настроить Stronghold для использования созданного ключа.

1. Установите необходимые пакеты:

   ```shell
   apt install libsofthsm2 opensc
   ```

1. Создайте директорию для хранения данных SoftHSM2 и файл конфигурации:

   ```shell
   mkdir /home/stronghold/softhsm
   cd softhsm
   echo "directories.tokendir = /home/stronghold/softhsm/" > /home/stronghold/softhsm2.conf
   ```

1. Задайте путь к конфигурации SoftHSM2 и библиотеке PKCS #11:

   ```shell
   export SOFTHSM2_CONF=/home/stronghold/softhsm2.conf
   HSMLIB="/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so"
   ```

1. Инициализируйте токен и задайте PIN-коды:

   ```shell
   pkcs11-tool --module $HSMLIB --init-token --so-pin 1234 --init-pin --pin 4321 --label my_token --login
   ```

   Пример вывода:

   ```console
   Using slot 0 with a present token (0x0)
   Token successfully initialized
   User PIN successfully initialized
   ```

1. Убедитесь, что токен создан и инициализирован:

   ```shell
   pkcs11-tool --module $HSMLIB -L
   ```

   Пример вывода:

   ```console
   Available slots:
   Slot 0 (0xe6829d3): SoftHSM slot ID 0xe6829d3
     token label        : my_token
     token manufacturer : SoftHSM project
     token model        : SoftHSM v2
     token flags        : login required, rng, token initialized, PIN initialized, other flags=0x20
     hardware version   : 2.6
     firmware version   : 2.6
     serial num         : 6a5468368e6829d3
     pin min/max        : 4/255
   Slot 1 (0x1): SoftHSM slot ID 0x1
     token state:   uninitialized
   ```

1. Создайте в токене ключевую пару RSA:

   ```shell
   pkcs11-tool --module $HSMLIB --login --pin 4321 --keypairgen --key-type rsa:4096 --label "vault-rsa-key"
   ```

   Пример вывода:

   ```console
   Using slot 0 with a present token (0xe6829d3)
   Key pair generated:
   Private Key Object; RSA
     label:      vault-rsa-key
     Usage:      decrypt, sign, signRecover, unwrap
     Access:     sensitive, always sensitive, never extractable, local
   Public Key Object; RSA 4096 bits
     label:      vault-rsa-key
     Usage:      encrypt, verify, verifyRecover, wrap
     Access:     local
   ```

1. Создайте файл конфигурации Stronghold (`config.hcl`) и добавьте в него настройки PKCS #11:

   ```console
   api_addr="https://0.0.0.0:8200"
   log_level = "warn"
   ui = true
   listener "tcp" {
     address = "0.0.0.0:8200"
     tls_cert_file = "/home/stronghold/cert.pem"
     tls_key_file  = "/home/stronghold/key.pem"
     #tls_require_and_verify_client_cert = true
     #tls_client_ca_file = "ca.crt"
     tls_disable = "false"
   }
   storage "raft" {
     path = "/home/stronghold/data"
   }

   seal "pkcs11" {
     lib = "/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so"
     token_label = "my_token"
     pin = "4321"
     key_label = "vault-rsa-key"
     rsa_oaep_hash = "sha1"
   }
   ```

1. Запустите Stronghold, указав конфигурацию SoftHSM2:

   ```shell
   export SOFTHSM2_CONF=/home/stronghold/softhsm2.conf
   d8 stronghold server -config config.hcl
   ```

## Использование Рутокен ЭЦП 3.0

Для шифрования root-ключа Stronghold можно использовать Рутокен ЭЦП 3.0 через интерфейс PKCS #11.
Выполните следующие действия, чтобы инициализировать токен, создать в нём ключевую пару
и настроить Stronghold для использования созданного ключа.

1. Скачайте и установите библиотеку `librtpkcs11ecp.so` [с сайта проекта «Рутокен»](https://www.rutoken.ru/support/download/pkcs/).

1. Задайте путь к библиотеке PKCS #11:

   ```shell
   HSMLIB="/usr/lib/librtpkcs11ecp.so"
   ```

1. Инициализируйте токен и задайте PIN-коды:

   ```shell
   pkcs11-tool --module $HSMLIB --init-token --so-pin 87654321 \
                 --init-pin --pin 12345678 --label my_token --login
   ```

1. Создайте в токене ключевую пару RSA, которая будет использоваться для шифрования root-ключа:

   ```shell
   pkcs11-tool --module $HSMLIB --login --pin 12345678 --keypairgen \
                 --key-type rsa:2048 --label "vault-rsa-key"
   ```

   Пример вывода:

   ```console
   Using slot 0 with a present token (0x0)
   Key pair generated:
   Private Key Object; RSA
     label:      vault-rsa-key
     Usage:      decrypt, sign
     Access:     sensitive, always sensitive, never extractable, local
   Public Key Object; RSA 2048 bits
     label:      vault-rsa-key
     Usage:      encrypt, verify
     Access:     local
   ```

1. Добавьте в конфигурацию Stronghold блок `seal "pkcs11"` с параметрами созданного токена и ключа:

   ```console
   seal "pkcs11" {
     lib = "/usr/lib/librtpkcs11ecp.so"
     token_label = "my_token"
     pin = "12345678"
     key_label = "vault-rsa-key"
   }
   ```

1. Запустите Stronghold и выполните `init`:

   ```shell
   systemctl start stronghold
   d8 stronghold operator init
   ```

1. Проверьте статус Stronghold:

   ```shell
   d8 stronghold status
   ```

   Пример вывода:

   ```console
   Key                      Value
   ---                      -----
   Recovery Seal Type       shamir
   Initialized              true
   Sealed                   false
   Total Recovery Shares    5
   Threshold                3
   Version                  1.15.2+hsm
   Build Date               2025-04-03T13:06:02Z
   Storage Type             raft
   Cluster Name             stronghold-cluster-6586e287
   Cluster ID               d7552773-2e8a-33b6-9c32-6749a4c9af13
   HA Enabled               false
   ```

## Миграция с Shamir-ключей на HSM

1. Измените конфигурацию Stronghold, добавив блок `seal "pkcs11"`:

   ```console
   seal "pkcs11" {
     lib = "/usr/lib/librtpkcs11ecp.so"
     token_label = "my_token"
     pin = "12345678"
     key_label = "vault-rsa-key"
   }
   ```

1. Перезапустите Stronghold. В логах появится сообщение:

   ```console
   2025-04-03T17:08:13.431+0300 [WARN]  core: entering seal migration mode; Stronghold will not automatically unseal even if using an autoseal: from_barrier_type=shamir to_barrier_type=pkcs11
   ```

1. Выполните миграцию, введя unseal-ключи:

   ```shell
   d8 stronghold operator unseal -migrate
   ```

После завершения миграции Stronghold при перезапуске будет автоматически распечатываться с использованием PKCS #11.

## Миграция с HSM на Shamir-ключи

1. Измените конфигурацию Stronghold, добавив параметр `disabled = "true"` в блок `seal "pkcs11"`:

   ```console
   seal "pkcs11" {
     lib = "/usr/lib/librtpkcs11ecp.so"
     token_label = "my_token"
     pin = "12345678"
     key_label = "vault-rsa-key"
     disabled = "true"
   }
   ```

1. Перезапустите Stronghold.

1. Выполните миграцию, введя recovery-ключи:

   ```shell
   d8 stronghold operator unseal -migrate
   ```

После завершения миграции при каждом перезапуске Stronghold потребуется вводить unseal-ключи вручную.
