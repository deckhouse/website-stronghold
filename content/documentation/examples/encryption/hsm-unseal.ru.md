---
title: "Автоматическое распечатывание и seal wrap через HSM"
linkTitle: "Auto-unseal через HSM"
description: "Защита root-ключа Stronghold EE аппаратным модулем через PKCS #11: блок seal pkcs11, лабораторный стенд на SoftHSM2, инициализация с ключами восстановления, двойное шифрование и рекомендации для production-HSM."
weight: 30
params:
  edition: ee
  relatedLinks:
    - title: "Поддержка HSM"
      url: ../../../admin/kms-hsm/hsm/
    - title: "Двойное шифрование (seal wrapping)"
      url: ../../../admin/kms-hsm/sealwrap/
    - title: "Запечатывание и распечатывание"
      url: ../../../concepts/seal/
    - title: "Шаблоны конфигурации standalone-установки"
      url: ../../../install/standalone/topologies/
---

Stronghold EE шифрует root-ключ ключом, который хранится в HSM и не покидает его. Доступ к HSM выполняется через стандартный интерфейс PKCS #11. После перезапуска Stronghold распечатывается автоматически, а наиболее чувствительные данные (связка ключей, ключ восстановления, ключи PKI, SSH и Transit) дополнительно шифруются через тот же seal — механизм seal wrap.

{{< alert level="warning" >}}
HSM поддерживается только в Stronghold EE и только при standalone-установке.
{{< /alert >}}

![Схема подключения Stronghold EE к HSM по PKCS #11](../../../images/ex-hsm-unseal.png)

## Цель

Собрать лабораторный стенд с SoftHSM2, настроить `seal "pkcs11"`, инициализировать Stronghold с ключами восстановления, проверить автоматическое распечатывание и seal wrap, а затем подготовить перенос конфигурации на аппаратный HSM.

## Предварительные требования

- Stronghold EE, установленный на Linux по разделу [«Установка»](../../../install/standalone/installation/). В примерах конфигурация находится в `/opt/stronghold/config.hcl`, сервис называется `stronghold` и запускается от пользователя `stronghold`.
- Для лабораторного стенда — Debian или Ubuntu с пакетами `softhsm2` и `opensc` (утилита `pkcs11-tool`). <!-- TODO(verify): имя пакета SoftHSM2 в поддерживаемых дистрибутивах (в разделе «Поддержка HSM» указан libsofthsm2) -->
- Для production — HSM с поддержкой PKCS #11 (например, Рутокен ЭЦП 3.0 или JaCarta) и PKCS #11-библиотека производителя.
- Для перевода существующего сервера с Shamir — пороговое число ключей распечатывания и резервная копия данных.

## Параметры блока seal "pkcs11"

| Параметр | Описание |
| --- | --- |
| `lib` | Путь к PKCS #11-библиотеке HSM, например `/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so` или `/usr/lib/librtpkcs11ecp.so` |
| `token_label` | Метка токена в HSM |
| `pin` | PIN пользователя токена |
| `key_label` | Метка ключевой пары RSA, которой шифруется root-ключ |
| `rsa_oaep_hash` | Хеш-функция для RSA-OAEP (`sha1`, `sha224`, `sha256`, `sha384`, `sha512`); по умолчанию `sha256`, в примере для SoftHSM2 — `sha1` |
| `slot` | Номер слота HSM (вместо `token_label`) |
| `key_id` | Идентификатор ключевой пары (вместо `key_label`) |
| `mechanism` | Механизм PKCS #11 для шифрования root-ключа |
| `hmac_key_label` | Метка HMAC-ключа |
| `generate_key` | Создать ключ в HSM, если его нет |
| `force_rw_session` | Открывать сессию PKCS #11 в режиме чтения-записи |
| `disabled` | Используется при миграции с HSM на другой seal |

Параметры также можно задать переменными окружения `VAULT_HSM_*` (например, `VAULT_HSM_PIN`, `VAULT_HSM_LIB`) или `PKCS11_WRAPPER_*`; значения из окружения имеют приоритет над конфигурацией.

## Шаг 1. Подготовьте токен SoftHSM2

Выполните шаг на узле Stronghold.

1. Установите пакеты:

   ```bash
   sudo apt install softhsm2 opensc
   ```

1. Создайте каталог токенов и конфигурацию SoftHSM2, доступные пользователю `stronghold`:

   ```bash
   sudo mkdir -p /opt/stronghold/softhsm/tokens
   echo "directories.tokendir = /opt/stronghold/softhsm/tokens" | sudo tee /opt/stronghold/softhsm/softhsm2.conf
   sudo chown -R stronghold:stronghold /opt/stronghold/softhsm
   sudo chmod 0700 /opt/stronghold/softhsm/tokens
   ```

1. Задайте переменные:

   ```bash
   export SOFTHSM2_CONF=/opt/stronghold/softhsm/softhsm2.conf
   HSMLIB=/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so
   ```

1. Инициализируйте токен и задайте PIN-коды. Выполняйте команды от пользователя `stronghold`, чтобы файлы токена принадлежали ему:

   ```bash
   sudo -u stronghold SOFTHSM2_CONF=$SOFTHSM2_CONF \
     pkcs11-tool --module $HSMLIB --init-token --so-pin 1234 \
     --init-pin --pin 4321 --label stronghold_token --login
   ```

1. Создайте ключевую пару RSA:

   ```bash
   sudo -u stronghold SOFTHSM2_CONF=$SOFTHSM2_CONF \
     pkcs11-tool --module $HSMLIB --login --pin 4321 \
     --keypairgen --key-type rsa:4096 --label stronghold-rsa-key
   ```

   В выводе у закрытого ключа должны быть атрибуты `sensitive, always sensitive, never extractable, local`: ключ нельзя извлечь из токена.

1. Проверьте токен и ключи:

   ```bash
   sudo -u stronghold SOFTHSM2_CONF=$SOFTHSM2_CONF \
     pkcs11-tool --module $HSMLIB --login --pin 4321 --list-objects
   ```

## Шаг 2. Настройте Stronghold

1. Добавьте блок `seal` в `/opt/stronghold/config.hcl`:

   ```hcl
   seal "pkcs11" {
     lib           = "/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so"
     token_label   = "stronghold_token"
     pin           = "4321"
     key_label     = "stronghold-rsa-key"
     rsa_oaep_hash = "sha1"
   }
   ```

   Ограничьте права на файл конфигурации, так как он содержит PIN:

   ```bash
   sudo chown stronghold:stronghold /opt/stronghold/config.hcl
   sudo chmod 0600 /opt/stronghold/config.hcl
   ```

   PIN можно передать через переменную окружения `VAULT_HSM_PIN` или `PKCS11_WRAPPER_PIN` вместо параметра `pin` в конфигурации.

1. Передайте сервису путь к конфигурации SoftHSM2 через drop-in для systemd:

   ```bash
   sudo mkdir -p /etc/systemd/system/stronghold.service.d
   sudo tee /etc/systemd/system/stronghold.service.d/softhsm.conf <<'UNIT'
   [Service]
   Environment=SOFTHSM2_CONF=/opt/stronghold/softhsm/softhsm2.conf
   UNIT
   sudo systemctl daemon-reload
   ```

Для нового сервера перейдите к шагу 3, для сервера, который уже инициализирован с Shamir, — к шагу 4.

## Шаг 3. Инициализируйте новый сервер

1. Запустите Stronghold:

   ```bash
   sudo systemctl start stronghold
   ```

1. Инициализируйте Stronghold с ключами восстановления:

   ```bash
   d8 stronghold operator init \
     -recovery-shares=5 \
     -recovery-threshold=3
   ```

   Распределите ключи восстановления между держателями и сохраните root-токен.

## Шаг 4. Переведите существующий сервер с Shamir на HSM

1. Создайте снимок Raft:

   ```bash
   d8 stronghold operator raft snapshot save pre-seal-migration.snap
   ```

1. Добавьте блок `seal "pkcs11"` (шаг 2) и перезапустите Stronghold:

   ```bash
   sudo systemctl restart stronghold
   ```

   В журнале появится сообщение:

   ```text
   core: entering seal migration mode; Stronghold will not automatically unseal even if using an autoseal: from_barrier_type=shamir to_barrier_type=pkcs11
   ```

1. Выполните распечатывание с флагом `-migrate`, введя пороговое число ключей распечатывания:

   ```bash
   d8 stronghold operator unseal -migrate
   ```

1. Дождитесь завершения перешифровки seal-wrapped записей: в журнале должно появиться сообщение `seal re-wrap completed`.

Для HA-кластера следуйте общей процедуре из раздела [«Миграция seal»](../../../concepts/seal/#миграция-seal).

## Шаг 5. Проверьте seal wrap

Для поддерживаемых seal двойное шифрование включено по умолчанию, отдельная настройка не требуется. Проверьте статус перешифровки:

```bash
d8 stronghold read sys/sealwrap/rewrap
```

Ответ содержит статус процесса и счётчики обработанных записей. Чтобы принудительно перешифровать все seal-wrapped записи, например после смены ключа в HSM, запустите rewrap:

```bash
d8 stronghold write -f sys/sealwrap/rewrap
```

Отключить двойное шифрование для всех данных, кроме root-ключа, можно параметром `disable_sealwrap = true` в конфигурации сервера. Подробнее — в разделе [«Двойное шифрование (seal wrapping)»](../../../admin/kms-hsm/sealwrap/).

## Проверка

1. Проверьте статус:

   ```bash
   d8 stronghold status
   ```

   Ожидаемые значения:

   ```text
   Seal Type                pkcs11
   Recovery Seal Type       shamir
   Initialized              true
   Sealed                   false
   ```

1. Перезапустите сервис и убедитесь, что Stronghold распечатался автоматически:

   ```bash
   sudo systemctl restart stronghold
   sleep 10
   d8 stronghold status -format=json | jq '{type, sealed, recovery_seal}'
   ```

1. Проверьте, что без доступа к HSM Stronghold не распечатывается. На лабораторном стенде временно переименуйте каталог токенов:

   ```bash
   sudo systemctl stop stronghold
   sudo mv /opt/stronghold/softhsm/tokens /opt/stronghold/softhsm/tokens.off
   sudo systemctl start stronghold
   d8 stronghold status
   journalctl -u stronghold.service --since "5 minutes ago" | grep -i pkcs11
   sudo systemctl stop stronghold
   sudo mv /opt/stronghold/softhsm/tokens.off /opt/stronghold/softhsm/tokens
   sudo systemctl start stronghold
   ```

   Пока токен недоступен, `Sealed` должно быть `true`, а в журнале — ошибка PKCS #11.

## Переход на аппаратный HSM

- **Библиотека и ключ.** Используйте PKCS #11-библиотеку производителя и создайте ключевую пару в HSM его средствами или через `pkcs11-tool`. Пример для Рутокен ЭЦП 3.0 приведён в разделе [«Поддержка HSM»](../../../admin/kms-hsm/hsm/). Значение `rsa_oaep_hash` подбирайте под поддерживаемые устройством механизмы.
- **Смена seal.** Переход с SoftHSM2 на аппаратный HSM — это миграция между двумя auto seal: добавьте `disabled = "true"` в старый блок `seal "pkcs11"`, добавьте новый блок и выполните `d8 stronghold operator unseal -migrate` с ключами восстановления. Порядок описан в разделе [«Миграция seal»](../../../concepts/seal/#миграция-seal). <!-- TODO(verify): поддерживается ли одновременное указание двух блоков seal "pkcs11" при миграции pkcs11 → pkcs11 -->
- **HA-кластер.** Каждому узлу нужен доступ к ключу с той же меткой. <!-- TODO(verify): требования к HSM в HA-кластере: общий сетевой HSM или клонирование ключа между устройствами -->
- **Резервное копирование ключа.** Потеря ключа в HSM делает данные Stronghold нерасшифровываемыми. Используйте процедуры резервирования ключей, предусмотренные производителем HSM.
- **Доступность.** При seal wrap HSM нужен не только при распечатывании, но и во время работы: операции с seal-wrapped значениями выполняются через HSM и медленнее обычных. Проверяйте стабильность подключения к устройству.
- **PIN.** Храните PIN только в файле конфигурации с правами `0600` для пользователя `stronghold`.

## Очистка

Для лабораторного стенда:

1. Остановите Stronghold: `sudo systemctl stop stronghold`.
1. Удалите данные Stronghold, токен SoftHSM2 (`/opt/stronghold/softhsm`) и drop-in `/etc/systemd/system/stronghold.service.d/softhsm.conf`, затем выполните `sudo systemctl daemon-reload`.
1. Удалите блок `seal "pkcs11"` из конфигурации.
