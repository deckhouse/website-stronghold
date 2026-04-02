---
title: "Yandex Cloud KMS"
weight: 15
description: "Настройка auto-unseal Stronghold через seal yandexcloudkms."
---

Stronghold поддерживает `seal "yandexcloudkms"` для автоматического распечатывания и защиты root key с использованием Yandex Cloud KMS.

{{< alert level="warning" >}}
Поддержка `seal "yandexcloudkms"` в текущем виде относится только к Standalone-установке Stronghold.
{{< /alert >}}

В cloud KMS seal-сценариях Stronghold поддерживает `yandexcloudkms`. Конфигурации `awskms` и `gcpckms` в Stronghold не поддерживаются.

## Что делает `seal "yandexcloudkms"`

Конфигурация `seal "yandexcloudkms"` позволяет Stronghold:

- использовать Yandex Cloud KMS для операций шифрования и расшифровки, связанных с root key;
- автоматически распечатываться после перезапуска без ручного ввода unseal-ключей;
- использовать внешний KMS вместо локально управляемого ключевого материала.

Если в конфигурации также используется двойное шифрование, внешний KMS должен быть доступен не только во время распечатывания, но и во время обычной работы Stronghold.

## Пример конфигурации

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

## Параметры `seal "yandexcloudkms"`

| Параметр | Обязательный | Описание |
|----------|--------------|----------|
| `kms_key_id` | да | Идентификатор симметричного ключа в Yandex Cloud KMS. |
| `oauth_token` | нет | OAuth token для аутентификации в Yandex Cloud. |
| `service_account_key_file` | нет | Путь к JSON-файлу авторизованного ключа сервисного аккаунта. |
| `endpoint` | нет | Пользовательский endpoint API Yandex Cloud. |
| `disabled` | нет | Используется при миграции с одного seal-механизма на другой. |

Важно:

- `kms_key_id` обязателен;
- `oauth_token` и `service_account_key_file` взаимоисключающие;
- если `oauth_token` и `service_account_key_file` не заданы, Stronghold пытается использовать сервисный аккаунт виртуальной машины через metadata service;
- если `endpoint` не задан, используется стандартный endpoint Yandex Cloud SDK.

## Порядок выбора учётных данных

Для `yandexcloudkms` используется следующий порядок при выборе аутентификации:

1. Значения из переменных окружения.
2. Значения из конфигурационного файла Stronghold.
3. Сервисный аккаунт виртуальной машины в Yandex Cloud.

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

- Для production-сред предпочтительно использовать сервисный аккаунт виртуальной машины или сервисный аккаунт с минимально необходимыми правами.
- Не указывайте `oauth_token` и `service_account_key_file` одновременно.
- При ротации ключей KMS заранее проверяйте процедуру rewrap и доступность старых версий ключевого материала.
- Если вы используете двойное шифрование, учитывайте влияние доступности Yandex Cloud KMS на runtime-операции Stronghold.

## См. также

- [Поддержка HSM](./hsm/)
- [Двойное шифрование](./sealwrap/)
- [Настройка Standalone](../standalone/configuration/)
