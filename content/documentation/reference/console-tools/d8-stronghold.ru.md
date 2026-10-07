---
title: "d8 stronghold"
description: "Справочник подкоманд d8 stronghold: работа с сервером, распечатывание, аутентификация, токены, политики, механизмы секретов, аудит, пространства имён и плагины."
weight: 20
---

Команда `d8 stronghold` утилиты [Deckhouse CLI](/products/kubernetes-platform/documentation/v1/cli/d8/) — клиент Stronghold. Подкоманды и флаги совпадают с CLI HashiCorp Vault (см. [«Совместимость с HashiCorp Vault»](../../../about/vault-compatibility/)).

При standalone-установке те же подкоманды доступны через исполняемый файл `stronghold`:

```shell
# Stronghold в DKP:
d8 stronghold <команда> [флаги] [аргументы]
# Stronghold в Linux:
stronghold <команда> [флаги] [аргументы]
```

Справку по любой подкоманде выводит флаг `-help`, например `d8 stronghold kv put -help`.

## Подключение к серверу

### Переменные окружения

| Переменная | Назначение | Пример |
| --- | --- | --- |
| `STRONGHOLD_ADDR` | Адрес сервера Stronghold | `export STRONGHOLD_ADDR=https://stronghold.example.com` |
| `STRONGHOLD_TOKEN` | Токен, с которым выполняются запросы. Заполняется автоматически после `login` | `export STRONGHOLD_TOKEN=<TOKEN>` |
| `STRONGHOLD_CACERT` | Путь к CA-сертификату для проверки TLS-сертификата сервера | `export STRONGHOLD_CACERT=/opt/stronghold/tls/stronghold-ca.pem` |

Клиент также читает переменные, которые задают параметры подключения. Перед запуском любая переменная `STRONGHOLD_*` копируется в одноимённую `VAULT_*`, поэтому все переменные ниже можно задавать с любым из двух префиксов. Если заданы обе, действует `STRONGHOLD_*`.

| Переменная | Назначение |
| --- | --- |
| `STRONGHOLD_NAMESPACE` | Пространство имён, в котором выполняются команды (Stronghold EE) |
| `STRONGHOLD_CAPATH` | Каталог с CA-сертификатами |
| `STRONGHOLD_CLIENT_CERT`, `STRONGHOLD_CLIENT_KEY` | Клиентский сертификат и ключ для взаимного TLS |
| `STRONGHOLD_TLS_SERVER_NAME` | Имя сервера для проверки TLS-сертификата (SNI) |
| `STRONGHOLD_SKIP_VERIFY` | Отключить проверку TLS-сертификата сервера. Использовать не рекомендуется |
| `STRONGHOLD_CLIENT_TIMEOUT` | Тайм-аут запросов клиента |
| `STRONGHOLD_MAX_RETRIES` | Максимальное число повторов запроса при ошибках |
| `STRONGHOLD_WRAP_TTL` | Время жизни для [обертывания ответа](../../../concepts/response-wrapping/) |
| `STRONGHOLD_MFA` | Данные MFA для запроса |
| `STRONGHOLD_FORMAT` | Формат вывода по умолчанию, например `json` |

Адрес сервера Stronghold в DP можно получить из объекта Ingress:

```shell
export STRONGHOLD_ADDR=https://$(d8 k -n d8-stronghold get ing stronghold -o json | jq -r '.spec.rules[0].host')
```

### Общие флаги

| Флаг | Назначение |
| --- | --- |
| `-address=<URL>` | Адрес сервера. Переопределяет `STRONGHOLD_ADDR`, например для обращения к secondary-кластеру |
| `-namespace=<путь>` | [Пространство имён](../../../admin/namespaces/overview/), в котором выполняется команда (Stronghold EE) |
| `-format=<формат>` | Формат вывода, например `json` |
| `-field=<поле>` | Вывести только значение указанного поля ответа (для `read`, `write`, `kv get` и других) |
| `-help` | Справка по команде |

Флаги подключения к серверу:

| Флаг | Назначение |
| --- | --- |
| `-ca-cert=<путь>` | Файл с CA-сертификатом для проверки сертификата сервера. Имеет приоритет над `-ca-path` |
| `-ca-path=<путь>` | Каталог с CA-сертификатами |
| `-tls-skip-verify` | Отключить проверку TLS-сертификата сервера. Использовать не рекомендуется |
| `-output-curl-string` | Не выполнять запрос, а вывести эквивалентную команду `curl` |

## Сервер и состояние

| Команда | Назначение | Подробнее |
| --- | --- | --- |
| `server -config=<файл>` | Запустить сервер Stronghold с указанным конфигурационным файлом (standalone). В DP сервером управляет модуль `stronghold` | [Установка](../../../install/standalone/installation/), [настройка](../../../install/standalone/configuration/) |
| `status` | Показать состояние сервера: инициализация, `Sealed`, тип seal, режим HA, индексы Raft | [Настройка доступа](../../../user/get-started/access/) |
| `version` | Показать версию и редакцию Stronghold | — |
| `monitor -log-level=<уровень>` | Вывести журналы узла в реальном времени | [Журналы](../../../admin/operations/logs/) |

```shell
d8 stronghold status
```

## operator

Команды обслуживания кластера. Требуют привилегированного токена или ключей распечатывания.

| Команда | Назначение | Ключевые флаги | Подробнее |
| --- | --- | --- | --- |
| `operator init` | Инициализировать хранилище и получить доли ключа и root-токен. В DP выполняется модулем автоматически | `-key-shares`, `-key-threshold` | [Установка](../../../install/standalone/installation/) |
| `operator unseal` | Ввести долю ключа распечатывания. Выполните столько раз, сколько задано `-key-threshold` | — | [Seal и unseal](../../../concepts/seal/) |
| `operator seal` | Запечатать узел | — | [Seal и unseal](../../../concepts/seal/) |
| `operator rekey` | Сменить ключи распечатывания или восстановления | `-init`, `-nonce`, `-verify` | [Seal и unseal](../../../concepts/seal/) |
| `operator generate-root` | Сгенерировать новый root-токен по кворуму ключей | `-init`, `-nonce`, `-decode`, `-otp` | [Токены](../../../concepts/tokens/) |
| `operator step-down` | Заставить активный узел HA уступить роль | — | [Переключение с EE на CSE](../../../install/standalone/switching-editions/ee-to-cse/) |
| `operator raft list-peers` | Показать узлы Raft и их роли | — | [Восстановление кворума Raft](../../../install/standalone/raft-lost-quorum-recovery/) |
| `operator raft join` | Присоединить узел к кластеру Raft | `-non-voter` | [Настройка](../../../install/standalone/configuration/) |
| `operator raft autopilot state` | Показать состояние автопилота Raft | — | [Мониторинг](../../../admin/operations/monitoring/) |
| `operator raft snapshot save <файл>` | Сохранить снимок Raft-хранилища | — | [Создание снимка](../../../admin/backups/save/) |
| `operator raft snapshot restore <файл>` | Восстановить хранилище из снимка | `-force` | [Восстановление](../../../admin/backups/restore/) |
| `operator raft snapshot inspect <файл>` | Проверить содержимое и консистентность снимка | `-validate`, `-depth`, `-filter`, `-format` | [Проверка снимка](../../../admin/backups/inspect/) |

Команда `operator seal` запечатывает сервер, а `operator rotate` ротирует ключ шифрования хранилища (см. [«Управление ключами»](../../../admin/operations/key-management/)).

```shell
d8 stronghold operator raft snapshot save stronghold-$(date +%F_%H-%M).snap
```

## Аутентификация

| Команда | Назначение | Ключевые флаги | Подробнее |
| --- | --- | --- | --- |
| `login` | Войти в Stronghold и сохранить токен | `-method`, `-path`, `-no-print` | [Настройка доступа](../../../user/get-started/access/) |
| `auth enable <тип>` | Включить метод аутентификации | `-path` | [Методы аутентификации](../../../user/auth/overview/) |
| `auth list` | Показать включённые методы аутентификации | `-format` | [Методы аутентификации](../../../user/auth/overview/) |
| `auth tune <путь>` | Изменить параметры метода аутентификации | например, `-user-lockout-disable` | [userpass](../../../user/auth/userpass/) |

```shell
d8 stronghold login -method=oidc -path=oidc_deckhouse -no-print
```

## token

| Команда | Назначение | Ключевые флаги | Подробнее |
| --- | --- | --- | --- |
| `token create` | Создать токен | `-policy`, `-period`, `-orphan`, `-no-default-policy`, `-namespace` | [Токены](../../../concepts/tokens/) |
| `token renew [<токен>]` | Продлить срок действия токена | — | [Токены](../../../concepts/tokens/) |
| `token revoke <токен>` | Отозвать токен и его дочерние токены | `-accessor` | [Токены](../../../concepts/tokens/) |

```shell
d8 stronghold token create -policy=myapp-read -period=24h
```

## policy

| Команда | Назначение | Подробнее |
| --- | --- | --- |
| `policy write <имя> <файл>` | Создать или обновить ACL-политику. Вместо файла можно передать `-` и прочитать политику из stdin | [Политики](../../../concepts/policy/) |

```shell
d8 stronghold policy write myapp-read - <<'POLICY'
path "secret/data/myapp/*" {
  capabilities = ["read"]
}
POLICY
```

## secrets

| Команда | Назначение | Ключевые флаги | Подробнее |
| --- | --- | --- | --- |
| `secrets enable <тип>` | Подключить механизм секретов | `-path`, `-version` (для `kv`) | [Механизмы секретов](../../../user/secrets-engines/overview/) |
| `secrets list` | Показать подключённые механизмы секретов | `-format` | [Механизмы секретов](../../../user/secrets-engines/overview/) |
| `secrets tune <путь>` | Изменить параметры mount | `-max-lease-ttl` | [PKI](../../../user/secrets-engines/pki/) |
| `secrets disable <путь>` | Отключить механизм секретов и удалить его данные | — | [Механизмы секретов](../../../user/secrets-engines/overview/) |

```shell
d8 stronghold secrets enable -path=secret -version=2 kv
```

## kv

| Команда | Назначение | Ключевые флаги | Подробнее |
| --- | --- | --- | --- |
| `kv put <путь> <ключ>=<значение>` | Записать секрет | `-mount` | [KV2](../../../user/secrets-engines/kv/kv-v2/) |
| `kv get <путь>` | Прочитать секрет | `-mount`, `-version`, `-field` | [KV2](../../../user/secrets-engines/kv/kv-v2/) |
| `kv patch <путь> <ключ>=<значение>` | Частично обновить секрет | `-mount` | [KV2](../../../user/secrets-engines/kv/kv-v2/) |
| `kv list <путь>` | Показать ключи по пути | `-mount`, `-recursive` | [KV](../../../user/secrets-engines/kv/overview/) |
| `kv delete <путь>` | Удалить последнюю версию секрета (KV2 — с возможностью восстановления) | `-mount`, `-versions` | [KV2](../../../user/secrets-engines/kv/kv-v2/) |
| `kv undelete <путь>` | Восстановить удалённые версии | `-mount`, `-versions` | [KV2](../../../user/secrets-engines/kv/kv-v2/) |
| `kv destroy <путь>` | Безвозвратно удалить версии | `-mount`, `-versions` | [KV2](../../../user/secrets-engines/kv/kv-v2/) |
| `kv metadata get\|put\|patch\|delete <путь>` | Работа с метаданными секрета KV2 | `-mount`, `-custom-metadata` | [KV2](../../../user/secrets-engines/kv/kv-v2/) |
| `kv enable-versioning <путь>` | Перевести хранилище KV1 на KV2 | — | [KV2](../../../user/secrets-engines/kv/kv-v2/) |

```shell
d8 stronghold kv put -mount=secret myapp/db password=s3cr3t
d8 stronghold kv get -mount=secret -field=password myapp/db
```

## Универсальные команды

Работают с любым путём API: механизмами секретов, методами аутентификации и системным бэкендом `sys/`.

| Команда | Назначение | Ключевые флаги |
| --- | --- | --- |
| `read <путь>` | Прочитать данные | `-field`, `-format`, `-address` |
| `write <путь> [<ключ>=<значение>...]` | Записать данные или выполнить операцию | `-f`/`-force` (запрос без данных), `-field` |
| `list <путь>` | Показать ключи по пути | `-format` |
| `delete <путь>` | Удалить данные по пути | — |
| `path-help <путь>` | Показать справку по путям механизма | — |

```shell
d8 stronghold write -f transit/keys/orders/rotate
d8 stronghold read -address="${SECONDARY_ADDR}" sys/replication/dr/status
d8 stronghold path-help pki
```

## lease

| Команда | Назначение | Ключевые флаги | Подробнее |
| --- | --- | --- | --- |
| `lease renew <lease_id>` | Продлить аренду | `-increment` | [Аренды](../../../concepts/lease/) |
| `lease revoke <lease_id>` | Отозвать аренду | `-prefix` | [Аренды](../../../concepts/lease/) |

```shell
d8 stronghold lease revoke -prefix database/creds/myapp/
```

## audit

Аудит доступен в Stronghold EE.

| Команда | Назначение | Подробнее |
| --- | --- | --- |
| `audit enable <тип> [параметры]` | Включить аудит-устройство `file`, `syslog` или `socket` | [Аудит](../../../admin/audit/overview/) |
| `audit list` | Показать включённые аудит-устройства | [Аудит](../../../admin/audit/overview/) |
| `audit disable <путь>` | Отключить аудит-устройство | [Аудит](../../../admin/audit/overview/) |

```shell
d8 stronghold audit enable file file_path=/var/log/stronghold_audit.log
```

## namespace

Пространства имён доступны в Stronghold EE.

| Команда | Назначение | Ключевые флаги | Подробнее |
| --- | --- | --- | --- |
| `namespace create <имя>` | Создать пространство имён | `-namespace` (родительское) | [Пространства имён](../../../admin/namespaces/overview/) |
| `namespace list` | Показать дочерние пространства имён | `-namespace` | [Пространства имён](../../../admin/namespaces/overview/) |
| `namespace lookup <имя>` | Показать сведения о пространстве имён | `-namespace` | [Пространства имён](../../../admin/namespaces/overview/) |
| `namespace delete <имя>` | Удалить пространство имён | `-namespace` | [Пространства имён](../../../admin/namespaces/overview/) |
| `namespace lock [<путь>]` | Заблокировать Namespace API | — | [Пространства имён](../../../admin/namespaces/overview/) |
| `namespace unlock [<путь>]` | Разблокировать Namespace API | `-unlock-key` | [Пространства имён](../../../admin/namespaces/overview/) |

```shell
d8 stronghold namespace create team-a
```

## plugin

| Команда | Назначение | Ключевые флаги | Подробнее |
| --- | --- | --- | --- |
| `plugin register <тип> <имя>` | Зарегистрировать внешний плагин в каталоге | `-sha256`, `-version` | [Плагины в Standalone](../../../admin/plugins/standalone/), [плагины в DP](../../../admin/plugins/dkp/) |
| `plugin deregister <тип> <имя>` | Удалить плагин из каталога | — | [Плагины](../../../admin/plugins/overview/) |

```shell
d8 stronghold plugin register -sha256="${PLUGIN_SHA}" secret my-custom-plugin
```
