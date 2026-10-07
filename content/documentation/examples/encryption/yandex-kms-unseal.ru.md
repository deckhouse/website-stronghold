---
title: "Автоматическое распечатывание через Yandex Cloud KMS"
linkTitle: "Auto-unseal через Yandex Cloud KMS"
description: "Standalone-сервер Stronghold с автоматическим распечатыванием через Yandex Cloud KMS: ключ и сервисный аккаунт в yc CLI, блок seal yandexcloudkms, инициализация с ключами восстановления, миграция с Shamir и проверка после перезапуска."
weight: 20
params:
  edition: ee
  relatedLinks:
    - title: "Yandex Cloud KMS"
      url: ../../../admin/kms-hsm/yandexcloudkms/
    - title: "Запечатывание и распечатывание"
      url: ../../../concepts/seal/
    - title: "Шаблоны конфигурации standalone-установки"
      url: ../../../install/standalone/topologies/
    - title: "Операции с ключами"
      url: ../../../admin/operations/key-management/
---

{{< alert level="warning" >}}
Автоматическое распечатывание через Yandex Cloud KMS доступно только в Stronghold EE.
{{< /alert >}}

При автоматическом распечатывании root-ключ Stronghold шифруется симметричным ключом Yandex Cloud KMS. После перезапуска Stronghold сам расшифровывает root-ключ через KMS, ввод ключей распечатывания не требуется. Вместо ключей распечатывания оператор получает ключи восстановления для административных операций.

{{< alert level="warning" >}}
`seal "yandexcloudkms"` поддерживается только при standalone-установке Stronghold.
{{< /alert >}}

![Схема автоматического распечатывания через Yandex Cloud KMS](../../../images/ex-yandex-kms-unseal.png)

## Цель

Создать ключ KMS и сервисный аккаунт с минимальными правами, настроить `seal "yandexcloudkms"`, инициализировать новый сервер с ключами восстановления или перевести существующий сервер с Shamir на KMS и убедиться, что Stronghold распечатывается автоматически после перезапуска.

## Предварительные требования

- Standalone-сервер Stronghold, установленный по разделу [«Установка»](../../../install/standalone/installation/). В примерах конфигурация находится в `/opt/stronghold/config.hcl`, сервис называется `stronghold`.
- Сетевой доступ с узлов Stronghold к API Yandex Cloud KMS по порту `443`/TCP.
- Утилита [yc CLI](https://yandex.cloud/ru/docs/cli/), настроенная на каталог Yandex Cloud, и права на создание ключей KMS и сервисных аккаунтов в нём.
- Утилита `jq`.
- Для миграции существующего сервера — пороговое число ключей распечатывания Shamir и резервная копия данных.

## Шаг 1. Создайте ключ KMS

1. Создайте симметричный ключ:

   ```bash
   yc kms symmetric-key create \
     --name stronghold-unseal \
     --default-algorithm aes-256 \
     --rotation-period 8760h
   ```

1. Сохраните идентификатор ключа:

   ```bash
   KMS_KEY_ID=$(yc kms symmetric-key get stronghold-unseal --format json | jq -r '.id')
   echo "$KMS_KEY_ID"
   ```

{{< alert level="danger" >}}
Удаление ключа или его версий делает root-ключ нерасшифровываемым, данные Stronghold будут потеряны. Включите защиту от удаления ключа (`--deletion-protection`) и ограничьте права на управление им.
{{< /alert >}}

## Шаг 2. Создайте сервисный аккаунт

1. Создайте сервисный аккаунт:

   ```bash
   yc iam service-account create --name stronghold-unseal
   ```

1. Выдайте аккаунту роль на шифрование и расшифровку только этим ключом:

   ```bash
   yc kms symmetric-key add-access-binding stronghold-unseal \
     --role kms.keys.encrypterDecrypter \
     --service-account-name stronghold-unseal
   ```

   <!-- TODO(verify): достаточно ли роли kms.keys.encrypterDecrypter для проверки существования ключа при инициализации или нужна также kms.keys.viewer -->

## Шаг 3. Выберите способ аутентификации

Выберите один из вариантов.

### Вариант A. Сервисный аккаунт ВМ (рекомендуется)

Если Stronghold работает на ВМ в Yandex Cloud, привяжите сервисный аккаунт к ВМ. Stronghold получит токен через metadata service, секреты в конфигурации не нужны:

```bash
yc compute instance update <имя-ВМ> --service-account-name stronghold-unseal
```

### Вариант B. Авторизованный ключ сервисного аккаунта

Если Stronghold работает вне Yandex Cloud, создайте авторизованный ключ и разместите его на узле:

```bash
yc iam key create \
  --service-account-name stronghold-unseal \
  --output yc-sa-key.json
sudo install -o stronghold -g stronghold -m 0400 yc-sa-key.json /etc/stronghold/yc-sa-key.json
rm yc-sa-key.json
```

Параметр `oauth_token` тоже поддерживается, но OAuth-токен привязан к пользователю и не подходит для production.

## Шаг 4. Добавьте блок seal в конфигурацию

Добавьте в `/opt/stronghold/config.hcl` на каждом узле блок `seal`.

Для варианта A:

```hcl
seal "yandexcloudkms" {
  kms_key_id = "<KMS_KEY_ID>"
}
```

Для варианта B:

```hcl
seal "yandexcloudkms" {
  kms_key_id               = "<KMS_KEY_ID>"
  service_account_key_file = "/etc/stronghold/yc-sa-key.json"
}
```

Вместо параметров конфигурации можно использовать переменные окружения `YANDEXCLOUD_KMS_KEY_ID`, `YANDEXCLOUD_SERVICE_ACCOUNT_KEY_FILE`, `YANDEXCLOUD_OAUTH_TOKEN` и `YANDEXCLOUD_ENDPOINT`. Переменные окружения имеют приоритет над конфигурацией. Полный список параметров приведён в разделе [«Yandex Cloud KMS»](../../../admin/kms-hsm/yandexcloudkms/).

Для нового сервера перейдите к шагу 5, для сервера, который уже инициализирован с Shamir, — к шагу 6.

## Шаг 5. Инициализируйте новый сервер

1. Запустите Stronghold:

   ```bash
   sudo systemctl start stronghold
   ```

1. Инициализируйте Stronghold. При auto unseal вместо ключей распечатывания создаются ключи восстановления:

   ```bash
   d8 stronghold operator init \
     -recovery-shares=5 \
     -recovery-threshold=3
   ```

   Распределите ключи восстановления между держателями и сохраните root-токен. Ключи восстановления не распечатывают Stronghold, но нужны для генерации root-токена, rekey и миграции seal.

1. Для HA-кластера запустите остальные узлы с тем же блоком `seal`. Они присоединятся к кластеру через `retry_join` и распечатаются автоматически.

## Шаг 6. Переведите существующий сервер с Shamir на KMS

Миграция seal требует кратковременного простоя. Во время миграции должны быть доступны и Shamir-ключи, и KMS.

1. Создайте снимок Raft:

   ```bash
   d8 stronghold operator raft snapshot save pre-seal-migration.snap
   ```

1. Добавьте блок `seal "yandexcloudkms"` в конфигурацию (шаг 4) и перезапустите Stronghold:

   ```bash
   sudo systemctl restart stronghold
   ```

   В журнале появится сообщение о режиме миграции:

   ```bash
   journalctl -u stronghold.service | grep "seal migration"
   ```

   ```text
   core: entering seal migration mode; Stronghold will not automatically unseal even if using an autoseal: from_barrier_type=shamir to_barrier_type=yandexcloudkms
   ```

1. Выполните распечатывание с флагом `-migrate`, введя пороговое число ключей распечатывания:

   ```bash
   d8 stronghold operator unseal -migrate
   ```

   После ввода порога ключи распечатывания Shamir становятся ключами восстановления.

Для HA-кластера выполняйте миграцию по узлам, начиная со standby-узлов, и передайте лидерство с активного узла, как описано в разделе [«Миграция seal»](../../../concepts/seal/#миграция-seal).

## Проверка

1. Проверьте статус:

   ```bash
   d8 stronghold status
   ```

   Ожидаемые значения:

   ```text
   Seal Type                yandexcloudkms
   Recovery Seal Type       shamir
   Initialized              true
   Sealed                   false
   ```

1. Перезапустите сервис и убедитесь, что Stronghold распечатался без ввода ключей:

   ```bash
   sudo systemctl restart stronghold
   sleep 10
   d8 stronghold status -format=json | jq '{type, sealed, recovery_seal}'
   ```

   Поле `sealed` должно быть `false`.

1. Если Stronghold остался запечатанным, проверьте журнал на ошибки аутентификации в Yandex Cloud или доступа к ключу:

   ```bash
   journalctl -u stronghold.service --since "10 minutes ago"
   ```

1. Проверьте, что ключи восстановления работают, запустив и отменив генерацию root-токена:

   ```bash
   d8 stronghold operator generate-root -init
   d8 stronghold operator generate-root -cancel
   ```

## Эксплуатация

- При плановой ротации ключа KMS старые версии сохраняются и остаются пригодными для расшифровки. Не удаляйте версии ключа. Порядок перешифровки данных после ротации описан в разделе [«Двойное шифрование (seal wrapping)»](../../../admin/kms-hsm/sealwrap/).
- Недоступность KMS не влияет на распечатанный сервер без двойного шифрования, но перезапущенный узел не распечатается, пока KMS недоступен.
- Смена и хранение ключей восстановления описаны в разделе [«Операции с ключами»](../../../admin/operations/key-management/).

## Очистка

Для тестового стенда:

1. Остановите Stronghold и удалите его данные.
1. Удалите авторизованный ключ: `yc iam key list --service-account-name stronghold-unseal`, затем `yc iam key delete <id>`.
1. Удалите сервисный аккаунт: `yc iam service-account delete stronghold-unseal`.
1. Удалите ключ KMS: `yc kms symmetric-key delete stronghold-unseal`. Делайте это только после удаления данных Stronghold, зашифрованных этим ключом.
