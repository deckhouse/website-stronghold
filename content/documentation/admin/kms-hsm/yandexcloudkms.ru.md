---
title: "Yandex Cloud KMS"
weight: 15
description: "Настройка auto-unseal Stronghold через seal yandexcloudkms."
params:
  relatedLinks:
    - title: "Настройка standalone-сервера Stronghold"
      url: ../../../install/standalone/configuration/
---

Stronghold поддерживает автоматическое распечатывание и защиту root-ключа с помощью [Yandex Cloud KMS](https://yandex.cloud/ru/docs/kms/).
Для интеграции с Yandex Cloud KMS используется блок конфигурации `seal "yandexcloudkms"`.

{{< alert level="warning" >}}
В текущей версии работа с `seal "yandexcloudkms"` поддерживается только при standalone-установке Stronghold.
{{< /alert >}}

Из внешних облачных KMS Stronghold поддерживает только Yandex Cloud KMS.
Конфигурации `awskms` и `gcpckms` не поддерживаются.

## Возможности Yandex Cloud KMS

Конфигурация `seal "yandexcloudkms"` позволяет:

- использовать Yandex Cloud KMS для шифрования и расшифровки данных, связанных с root-ключом;
- автоматически распечатывать Stronghold после перезапуска без ручного ввода unseal-ключей;
- использовать внешний KMS вместо локально управляемого ключевого материала.

Если в конфигурации используется двойное шифрование, Yandex Cloud KMS должен быть доступен не только во время распечатывания, но и в процессе работы Stronghold.

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

## Параметры конфигурации Yandex Cloud KMS

| Параметр | Обязательный | Описание |
| --- | --- | --- |
| `kms_key_id` | Да | Идентификатор симметричного ключа в Yandex Cloud KMS |
| `oauth_token` | Нет | OAuth-токен для аутентификации в Yandex Cloud. Нельзя указывать одновременно с `service_account_key_file` |
| `service_account_key_file` | Нет | Путь к JSON-файлу авторизованного ключа сервисного аккаунта. Нельзя указывать одновременно с `oauth_token` |
| `endpoint` | Нет | Кастомный эндпоинт API Yandex Cloud. Если параметр не задан, используется стандартный эндпоинт Yandex Cloud SDK |
| `disabled` | Нет | Используется при миграции с одного seal-механизма на другой |

{{< alert level="info" >}}
Если параметры `oauth_token` и `service_account_key_file` не заданы, Stronghold пытается выполнить аутентификацию с помощью сервисного аккаунта виртуальной машины через metadata service.
{{< /alert >}}

## Порядок выбора учётных данных

Для Yandex Cloud KMS используется следующий порядок при выборе учётных данных для аутентификации:

1. Значения из переменных окружения.
2. Значения из конфигурационного файла Stronghold.
3. Сервисный аккаунт виртуальной машины в Yandex Cloud.

Таким образом, у переменных окружения есть приоритет над параметрами в блоке `seal "yandexcloudkms"`.

## Переменные окружения

Поддерживаются следующие переменные окружения:

- `YANDEXCLOUD_KMS_KEY_ID`;
- `YANDEXCLOUD_OAUTH_TOKEN`;
- `YANDEXCLOUD_SERVICE_ACCOUNT_KEY_FILE`;
- `YANDEXCLOUD_ENDPOINT`.

Их можно использовать вместо соответствующих параметров в конфигурационном файле или вместе с ним, если это соответствует вашей операционной модели.

## Требования к доступу

При инициализации Stronghold проверяет, что указанный ключ существует и что у приложения есть право выполнять операции шифрования.

Для работы конфигурации `seal "yandexcloudkms"` необходимо:

- существование симметричного ключа в Yandex Cloud KMS;
- наличие прав на шифрование и расшифровку этим ключом;
- корректная аутентификация через OAuth-токен, ключ сервисного аккаунта или сервисный аккаунт виртуальной машины.

## Рекомендации

- Для production-сред используйте сервисный аккаунт виртуальной машины или сервисный аккаунт с минимально необходимыми правами.
- При ротации ключей KMS заранее проверяйте процедуру rewrap и доступность старых версий ключевого материала.
- Если вы используете двойное шифрование, учитывайте влияние доступности Yandex Cloud KMS на runtime-операции Stronghold.
