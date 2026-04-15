---
title: "KMS и HSM"
linkTitle: "KMS и HSM"
description: "Интеграция Deckhouse Stronghold с внешними KMS и HSM для защиты root key и автоматического распечатывания"
weight: 10
---

## KMS и HSM

**Deckhouse Stronghold** поддерживает интеграцию с внешними системами управления ключами (**KMS**) и аппаратными модулями безопасности (**HSM**) для защиты root key и реализации автоматического распечатывания (`auto-unseal`).

Использование KMS и HSM позволяет вынести критически важные операции шифрования за пределы локального хранилища Stronghold и снизить риски, связанные с управлением ключевым материалом внутри самой инсталляции.

## Зачем использовать KMS и HSM

Интеграция с KMS и HSM полезна в следующих сценариях:

- необходимо исключить хранение критичного ключевого материала только внутри Stronghold;
- требуется автоматическое распечатывание без ручного ввода unseal-ключей при каждом перезапуске;
- в инфраструктуре уже используются внешние системы управления ключами;
- есть требования по повышенной защите root key;
- необходимо соответствовать внутренним политикам безопасности или регуляторным требованиям.

## Чем отличаются KMS и HSM

На практике KMS и HSM решают близкие задачи, но относятся к разным классам средств защиты.

### KMS

**KMS** — внешняя система управления ключами. Обычно это программный или облачный сервис, предоставляющий API для операций шифрования и расшифровки.

Для Stronghold KMS может использоваться как внешний `seal`-механизм, который:

- защищает root key;
- участвует в операциях распечатывания;
- позволяет реализовать auto-unseal;
- выносит операции управления ключом во внешний доверенный сервис.

### HSM

**HSM** — аппаратный модуль безопасности, предназначенный для защищённого хранения и использования криптографических ключей.

Для Stronghold HSM обычно используется через `PKCS#11` и позволяет:

- хранить ключи в аппаратном устройстве;
- выполнять криптографические операции без извлечения закрытого ключа;
- защищать root key Stronghold;
- использовать auto-unseal без хранения приватного ключевого материала на сервере Stronghold.

## Как это связано с seal и auto-unseal

Stronghold использует `seal`-механизм для защиты root key.

В базовом сценарии root key защищается встроенной схемой с использованием unseal-ключей. При использовании внешнего KMS или HSM Stronghold может выполнять операции распечатывания автоматически, не требуя ручного ввода unseal-ключей при каждом старте.

Это означает, что:

- root key остаётся защищённым внешним механизмом;
- сервер Stronghold может автоматически переходить в рабочее состояние после перезапуска;
- администратор получает более предсказуемую эксплуатационную модель для production-окружений.

> Важно  
> Использование auto-unseal не отменяет требований к безопасному хранению recovery-материалов и административных ключей. Внешний `seal` упрощает эксплуатацию, но не снимает ответственности за управление доступом и аварийное восстановление.

## Поддерживаемые сценарии

В текущем документированном виде Stronghold поддерживает следующие сценарии:

- `seal "pkcs11"` — для интеграции с HSM и совместимыми устройствами через `PKCS#11`;
- `seal "yandexcloudkms"` — для интеграции с **Yandex Cloud KMS**.

> Предупреждение  
> Сценарии, описанные в этом разделе, относятся только к **Standalone-развёртыванию** Stronghold. Примеры используют локальный конфигурационный файл и настройку `seal` в конфигурации standalone-сервера.

## Когда выбирать HSM, а когда KMS

Выбор зависит от требований инфраструктуры и модели безопасности.

### Когда выбирать HSM

HSM подходит, если:

- требуется аппаратная защита ключевого материала;
- в организации уже используются HSM-устройства;
- есть регуляторные требования к хранению ключей;
- важно, чтобы приватный ключ не покидал аппаратное устройство.

### Когда выбирать KMS

KMS подходит, если:

- используется облачная или сервисная модель управления ключами;
- требуется быстрее интегрировать Stronghold в уже существующую облачную инфраструктуру;
- важна простота эксплуатации без локального аппаратного устройства;
- допустимо доверять внешнему KMS-сервису как корню защиты root key.

## Поддержка HSM через PKCS#11

Stronghold поддерживает шифрование root key с использованием аппаратных модулей защиты, таких как:

- `TPM2`;
- Рутокен ЭЦП 3.0;
- JaCarta;
- другие устройства с поддержкой стандарта `PKCS#11`.

Для целей тестирования и разработки также поддерживается **SoftHSM2**.

> Предупреждение  
> Поддержка HSM в текущем виде относится только к Standalone-установке Stronghold. Примеры в этом разделе используют локальный конфигурационный файл и `seal "pkcs11"` в конфигурации standalone-сервера.

Для использования автоматического распечатывания через `PKCS#11` необходимо предварительно создать ключи в HSM и сконфигурировать Stronghold для работы с ними.

## SoftHSM2

**SoftHSM2** полезен для тестирования, отладки и лабораторных сценариев, когда нужно воспроизвести поведение HSM без использования реального аппаратного устройства.

### Установка пакетов

```bash
apt install libsofthsm2 opensc
```

### Создание конфигурации SoftHSM2

```bash
mkdir /home/stronghold/softhsm
cd softhsm
echo "directories.tokendir = /home/stronghold/softhsm/" > /home/stronghold/softhsm2.conf
```

### Генерация ключей в HSM

```bash
export SOFTHSM2_CONF=/home/stronghold/softhsm2.conf
HSMLIB="/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so"

pkcs11-tool --module $HSMLIB --init-token --so-pin 1234 --init-pin --pin 4321 --label my_token --login
```

Пример вывода:

```text
Using slot 0 with a present token (0x0)
Token successfully initialized
User PIN successfully initialized
```

Проверка доступных слотов:

```bash
pkcs11-tool --module $HSMLIB -L
```

Генерация ключевой пары:

```bash
pkcs11-tool --module $HSMLIB --login --pin 4321 --keypairgen --key-type rsa:4096 --label "vault-rsa-key"
```

Пример вывода:

```text
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

### Пример конфигурации Stronghold с pkcs11

```hcl
api_addr="https://0.0.0.0:8200"
log_level = "warn"
ui = true

listener "tcp" {
  address = "0.0.0.0:8200"
  tls_cert_file = "/home/stronghold/cert.pem"
  tls_key_file  = "/home/stronghold/key.pem"
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

### Запуск Stronghold

```bash
export SOFTHSM2_CONF=/home/stronghold/softhsm2.conf
stronghold server -config config.hcl
```

## Использование Рутокен ЭЦП 3.0

Для работы с Рутокен ЭЦП 3.0 требуется библиотека `librtpkcs11ecp.so`.

### Установка библиотеки

Скачайте и установите библиотеку `librtpkcs11ecp.so` с сайта производителя.

### Генерация ключей

```bash
HSMLIB="/usr/lib/librtpkcs11ecp.so"

pkcs11-tool --module $HSMLIB --init-token --so-pin 87654321 \
            --init-pin --pin 12345678 --label my_token --login

pkcs11-tool --module $HSMLIB --login --pin 12345678 --keypairgen \
            --key-type rsa:2048 --label "vault-rsa-key"
```

Пример вывода:

```text
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

### Пример seal-конфигурации

```hcl
seal "pkcs11" {
  lib = "/usr/lib/librtpkcs11ecp.so"
  token_label = "my_token"
  pin = "12345678"
  key_label = "vault-rsa-key"
}
```

### Инициализация и проверка

Запустите Stronghold и выполните инициализацию:

```bash
systemctl start stronghold
stronghold operator init
```

Проверьте статус:

```bash
stronghold status
```

Пример вывода:

```text
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

## Миграция с Shamir на HSM

Если Stronghold уже работает с Shamir-ключами, можно перейти на HSM-сценарий.

### Шаг 1. Измените конфигурацию

Добавьте блок `seal`:

```hcl
seal "pkcs11" {
  lib = "/usr/lib/librtpkcs11ecp.so"
  token_label = "my_token"
  pin = "12345678"
  key_label = "vault-rsa-key"
}
```

### Шаг 2. Перезапустите Stronghold

После перезапуска в журналах появится сообщение о переходе в режим миграции `seal`:

```text
2025-04-03T17:08:13.431+0300 [WARN]  core: entering seal migration mode; Stronghold will not automatically unseal even if using an autoseal: from_barrier_type=shamir to_barrier_type=pkcs11
```

### Шаг 3. Выполните миграцию

```bash
stronghold operator unseal -migrate
```

После завершения миграции Stronghold при перезапуске будет автоматически распечатываться с использованием `pkcs11`.

## Миграция с HSM на Shamir

Если требуется отказаться от HSM-сценария и вернуться к Shamir-схеме:

### Шаг 1. Измените конфигурацию

Добавьте параметр `disabled = "true"` в раздел `seal`:

```hcl
seal "pkcs11" {
  lib = "/usr/lib/librtpkcs11ecp.so"
  token_label = "my_token"
  pin = "12345678"
  key_label = "vault-rsa-key"
  disabled = "true"
}
```

### Шаг 2. Перезапустите Stronghold

```bash
systemctl restart stronghold
```

### Шаг 3. Выполните миграцию

```bash
stronghold operator unseal -migrate
```

После завершения миграции при каждом перезапуске Stronghold потребуется вводить ключи распечатывания вручную.

## Yandex Cloud KMS

Stronghold поддерживает `seal "yandexcloudkms"` для автоматического распечатывания и защиты root key с использованием **Yandex Cloud KMS**.

> Предупреждение  
> Поддержка `seal "yandexcloudkms"` в текущем виде относится только к Standalone-установке Stronghold.

В cloud KMS `seal`-сценариях Stronghold поддерживает `yandexcloudkms`. Конфигурации `awskms` и `gcpckms` в Stronghold не поддерживаются.

## Что делает seal "yandexcloudkms"

Конфигурация `seal "yandexcloudkms"` позволяет Stronghold:

- использовать Yandex Cloud KMS для операций шифрования и расшифровки, связанных с root key;
- автоматически распечатываться после перезапуска без ручного ввода unseal-ключей;
- использовать внешний KMS вместо локально управляемого ключевого материала.

Если в конфигурации также используется двойное шифрование, внешний KMS должен быть доступен не только во время распечатывания, но и во время обычной работы Stronghold.

## Пример конфигурации yandexcloudkms

```hcl
seal "yandexcloudkms" {
  kms_key_id = "abj1abc23def456ghi78"
  oauth_token = "y0_AQAAAA..."
}
```

Пример с использованием сервисного аккаунта:

```hcl
seal "yandexcloudkms" {
  kms_key_id = "abj1abc23def456ghi78"
  service_account_key_file = "/etc/stronghold/yc-sa-key.json"
}
```

## Параметры seal "yandexcloudkms"

| Параметр | Обязательный | Описание |
| --- | --- | --- |
| `kms_key_id` | да | Идентификатор симметричного ключа в Yandex Cloud KMS |
| `oauth_token` | нет | OAuth token для аутентификации в Yandex Cloud |
| `service_account_key_file` | нет | Путь к JSON-файлу авторизованного ключа сервисного аккаунта |
| `endpoint` | нет | Пользовательский endpoint API Yandex Cloud |
| `disabled` | нет | Используется при миграции с одного seal-механизма на другой |

Важно учитывать:

- `kms_key_id` обязателен;
- `oauth_token` и `service_account_key_file` взаимоисключающие;
- если `oauth_token` и `service_account_key_file` не заданы, Stronghold пытается использовать сервисный аккаунт виртуальной машины через metadata service;
- если `endpoint` не задан, используется стандартный endpoint Yandex Cloud SDK.

## Порядок выбора учётных данных

Для `yandexcloudkms` используется следующий порядок выбора аутентификации:

1. значения из переменных окружения;
2. значения из конфигурационного файла Stronghold;
3. сервисный аккаунт виртуальной машины в Yandex Cloud.

Это означает, что переменные окружения имеют приоритет над параметрами в `seal`-блоке.

## Переменные окружения

Поддерживаются следующие переменные окружения:

- `YANDEXCLOUD_KMS_KEY_ID`
- `YANDEXCLOUD_OAUTH_TOKEN`
- `YANDEXCLOUD_SERVICE_ACCOUNT_KEY_FILE`
- `YANDEXCLOUD_ENDPOINT`

Их можно использовать вместо соответствующих параметров в конфигурационном файле или вместе с ним, если это соответствует вашей операционной модели.

## Требования к доступу

При инициализации Stronghold проверяет, что указанный ключ существует и что у приложения есть право выполнять операции шифрования.

На практике для работы `seal "yandexcloudkms"` необходимо:

- существование симметричного ключа в Yandex Cloud KMS;
- права на шифрование и расшифровку этим ключом;
- корректная аутентификация через OAuth token, ключ сервисного аккаунта или сервисный аккаунт виртуальной машины.

## Практические рекомендации

### Для HSM

- используйте HSM, если требуется аппаратная защита ключей;
- заранее проверяйте совместимость устройства с `PKCS#11`;
- отдельно тестируйте сценарии миграции `seal`;
- документируйте порядок доступа к устройству, PIN-кодам и recovery-материалам;
- для production-сценариев проверяйте поведение после перезапуска сервиса и узла.

### Для KMS

- для production-окружений предпочтительно использовать сервисный аккаунт виртуальной машины или сервисный аккаунт с минимально необходимыми правами;
-````markdown
- не указывайте `oauth_token` и `service_account_key_file` одновременно;
- при ротации ключей KMS заранее проверяйте процедуру rewrap и доступность старых версий ключевого материала;
- если используется двойное шифрование, учитывайте влияние доступности Yandex Cloud KMS на runtime-операции Stronghold.

## Ограничения и эксплуатационные замечания

При использовании внешних KMS и HSM учитывайте:

- описанные в этом разделе сценарии относятся к **Standalone-развёртыванию**;
- внешний `seal`-механизм упрощает auto-unseal, но не отменяет необходимость безопасного хранения recovery-материалов;
- при недоступности внешнего KMS или HSM могут быть затронуты операции распечатывания, а в отдельных конфигурациях — и часть runtime-операций;
- миграции `seal`-механизма должны выполняться контролируемо и с пониманием последствий для аварийного восстановления;
- любые изменения схемы защиты root key желательно сначала тестировать на отдельном контуре.

## Что дальше

Если требуется дополнительный уровень защиты чувствительных внутренних данных, перейдите в раздел [Двойное шифрование](../Двойное шифрование/).  
Если нужно ознакомиться с поддерживаемыми алгоритмами и криптографическими особенностями, используйте раздел [Криптоалгоритмы](../Криптоалгоритмы/).
