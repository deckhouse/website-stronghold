---
title: "Типовые топологии"
description: "Готовые шаблоны конфигурации Stronghold: один узел для разработки и тестирования, HA-кластеры Raft из трёх и пяти узлов с TLS и retry_join, автоматическое распечатывание через HSM, Yandex Cloud KMS и inner-cluster."
weight: 25
---

На этой странице собраны шаблоны конфигурации `/opt/stronghold/config.hcl` для типовых топологий standalone-установки. Параметры описаны в разделе [«Настройка»](../configuration/), порядок развёртывания — в разделе [«Установка»](../installation/).

| Топология | Назначение | Отказоустойчивость |
| --- | --- | --- |
| [Один узел](#один-узел) | Разработка, тестирование, стенды | Нет |
| [Три узла Raft](#ha-кластер-из-трёх-узлов) | Production | Отказ одного узла |
| [Пять узлов Raft](#ha-кластер-из-пяти-узлов) | Production с повышенными требованиями к доступности | Отказ двух узлов |
| [DR-пара](#dr-пара) (Stronghold EE) | Горячий резерв в другой площадке | Отказ кластера целиком |

Во всех примерах используются каталоги `/opt/stronghold/data` (данные Raft) и `/opt/stronghold/tls` (сертификаты), systemd-unit из раздела [«Установка»](../installation/#запуск-через-systemd-unit) и порты `8200` (API) и `8201` (кластерный трафик). Замените IP-адреса, имена узлов и пути на значения вашей инфраструктуры.

## Один узел

Подходит для разработки и тестирования. Самый быстрый способ получить такую установку — команда [`stronghold bootstrap service`](../bootstrap/), которая создаёт каталоги, конфигурацию, systemd-unit и при необходимости самоподписанные сертификаты.

Если вы готовите конфигурацию вручную, используйте шаблон:

```hcl
ui            = true
cluster_addr  = "https://127.0.0.1:8201"
api_addr      = "https://stronghold.demo.tld:8200"
disable_mlock = true

listener "tcp" {
  address       = "0.0.0.0:8200"
  tls_cert_file = "/opt/stronghold/tls/stronghold-cert.pem"
  tls_key_file  = "/opt/stronghold/tls/stronghold-key.pem"
}

storage "raft" {
  path    = "/opt/stronghold/data"
  node_id = "raft-node-1"
}
```

{{< alert level="warning" >}}
Один узел не обеспечивает отказоустойчивость. Не используйте эту топологию в production.
{{< /alert >}}

## HA-кластер из трёх узлов

Минимальная production-конфигурация: кворум Raft — два узла из трёх, кластер продолжает работу при отказе одного узла.

Перед развёртыванием:

- выпустите для каждого узла сертификат, в поле `subjectAltName` которого указаны FQDN и IP-адрес узла (см. [«Подготовка необходимых сертификатов»](../installation/#подготовка-необходимых-сертификатов));
- разместите на каждом узле сертификат и ключ узла, а также сертификат CA `stronghold-ca.pem`;
- откройте TCP-порты `8200` и `8201` между узлами.

Конфигурация первого узла (`raft-node-1`, `10.20.30.10`):

```hcl
ui            = true
cluster_addr  = "https://10.20.30.10:8201"
api_addr      = "https://10.20.30.10:8200"
disable_mlock = true

listener "tcp" {
  address         = "0.0.0.0:8200"
  tls_cert_file   = "/opt/stronghold/tls/node-1-cert.pem"
  tls_key_file    = "/opt/stronghold/tls/node-1-key.pem"
  tls_min_version = "tls12"
}

storage "raft" {
  path    = "/opt/stronghold/data"
  node_id = "raft-node-1"

  retry_join {
    leader_tls_servername   = "raft-node-1.demo.tld"
    leader_api_addr         = "https://10.20.30.10:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }

  retry_join {
    leader_tls_servername   = "raft-node-2.demo.tld"
    leader_api_addr         = "https://10.20.30.11:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }

  retry_join {
    leader_tls_servername   = "raft-node-3.demo.tld"
    leader_api_addr         = "https://10.20.30.12:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }
}
```

Блоки `retry_join` перечисляют все узлы кластера, поэтому одна и та же структура подходит для каждого узла. На остальных узлах измените только значения, относящиеся к самому узлу:

| Параметр | `raft-node-2` | `raft-node-3` |
| --- | --- | --- |
| `cluster_addr` | `https://10.20.30.11:8201` | `https://10.20.30.12:8201` |
| `api_addr` | `https://10.20.30.11:8200` | `https://10.20.30.12:8200` |
| `tls_cert_file`, `leader_client_cert_file` | `/opt/stronghold/tls/node-2-cert.pem` | `/opt/stronghold/tls/node-3-cert.pem` |
| `tls_key_file`, `leader_client_key_file` | `/opt/stronghold/tls/node-2-key.pem` | `/opt/stronghold/tls/node-3-key.pem` |
| `node_id` | `raft-node-2` | `raft-node-3` |

Запустите сервис на всех узлах, инициализируйте кластер на одном узле (`stronghold operator init`) и распечатайте каждый узел (`stronghold operator unseal`). Проверьте состав кластера:

```shell
stronghold operator raft list-peers
```

## HA-кластер из пяти узлов

Кворум Raft — три узла из пяти, кластер продолжает работу при отказе двух узлов. Используйте эту топологию, если узлы размещены в нескольких зонах доступности или требуется обслуживать узлы без снижения отказоустойчивости.

Конфигурация первого узла (`raft-node-1`, `10.20.30.10`):

```hcl
ui            = true
cluster_addr  = "https://10.20.30.10:8201"
api_addr      = "https://10.20.30.10:8200"
disable_mlock = true

listener "tcp" {
  address         = "0.0.0.0:8200"
  tls_cert_file   = "/opt/stronghold/tls/node-1-cert.pem"
  tls_key_file    = "/opt/stronghold/tls/node-1-key.pem"
  tls_min_version = "tls12"
}

storage "raft" {
  path    = "/opt/stronghold/data"
  node_id = "raft-node-1"

  retry_join {
    leader_tls_servername   = "raft-node-1.demo.tld"
    leader_api_addr         = "https://10.20.30.10:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }

  retry_join {
    leader_tls_servername   = "raft-node-2.demo.tld"
    leader_api_addr         = "https://10.20.30.11:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }

  retry_join {
    leader_tls_servername   = "raft-node-3.demo.tld"
    leader_api_addr         = "https://10.20.30.12:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }

  retry_join {
    leader_tls_servername   = "raft-node-4.demo.tld"
    leader_api_addr         = "https://10.20.30.13:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }

  retry_join {
    leader_tls_servername   = "raft-node-5.demo.tld"
    leader_api_addr         = "https://10.20.30.14:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }
}
```

На узлах `raft-node-2` — `raft-node-5` (адреса `10.20.30.11` — `10.20.30.14`) измените `cluster_addr`, `api_addr`, `node_id` и пути к сертификату и ключу узла так же, как для кластера из трёх узлов.

{{< alert level="info" >}}
Используйте нечётное число голосующих узлов. Узлы, которые должны только получать поток репликации и не участвовать в кворуме, подключайте с параметром `retry_join_as_non_voter = true` в секции `storage "raft"`.
{{< /alert >}}

## Автоматическое распечатывание

Добавьте в конфигурацию каждого узла одну из секций `seal`, приведённых ниже. При инициализации кластера с HSM или KMS Stronghold возвращает ключи восстановления вместо ключей распечатывания (см. [«Seal и unseal»](../../../concepts/seal/)). Для инициализации задайте параметры `-recovery-shares` и `-recovery-threshold`.

Для HSM и Yandex Cloud KMS при инициализации задаются recovery-ключи, а не unseal-ключи:

```shell
stronghold operator init -recovery-shares=5 -recovery-threshold=3
```

### HSM (PKCS #11)

Доступно в Stronghold EE. Предварительно создайте ключи в HSM, как описано в разделе [«Поддержка HSM»](../../../admin/kms-hsm/hsm/).

```hcl
seal "pkcs11" {
  lib         = "/usr/lib/librtpkcs11ecp.so"
  token_label = "my_token"
  pin         = "<PIN>"
  key_label   = "vault-rsa-key"
}
```

### Yandex Cloud KMS

Параметры и способы аутентификации описаны в разделе [«Yandex Cloud KMS»](../../../admin/kms-hsm/yandexcloudkms/).

```hcl
seal "yandexcloudkms" {
  kms_key_id               = "<KMS_KEY_ID>"
  service_account_key_file = "/etc/stronghold/yc-sa-key.json"
}
```

### inner-cluster

Доступно в Stronghold EE и Stronghold CSE. Распечатанные узлы кластера распечатывают остальные узлы, внешний KMS не требуется (см. [«Seal и unseal»](../../../concepts/seal/)).

```hcl
seal "inner-cluster" {
  node {
    name        = "raft-node-1"
    address     = "https://10.20.30.10:8200"
    tls_ca_cert = "/opt/stronghold/tls/stronghold-ca.pem"
  }
  node {
    name        = "raft-node-2"
    address     = "https://10.20.30.11:8200"
    tls_ca_cert = "/opt/stronghold/tls/stronghold-ca.pem"
  }
  node {
    name        = "raft-node-3"
    address     = "https://10.20.30.12:8200"
    tls_ca_cert = "/opt/stronghold/tls/stronghold-ca.pem"
  }
}
```

{{< alert level="warning" >}}
Не храните PIN HSM и ключи сервисных аккаунтов в конфигурационном файле, доступном другим пользователям системы. Ограничьте права на `/opt/stronghold/config.hcl` пользователем `stronghold`.
{{< /alert >}}

PIN можно передать через переменную окружения `VAULT_HSM_PIN` или `PKCS11_WRAPPER_PIN` вместо параметра `pin` в конфигурации: так PIN не попадает в файл.

## DR-пара

Доступно в Stronghold EE. DR-пара — два независимых HA-кластера, например по три узла в разных площадках. Каждый кластер развёртывается по шаблону из раздела [«HA-кластер из трёх узлов»](#ha-кластер-из-трёх-узлов) со своими сертификатами и своим seal. Данные, зашифрованные через Seal, не входят в поток репликации, поэтому каждый кластер запечатывает данные своим seal.

Особенности:

- оба кластера используют Stronghold EE с integrated Raft;
- кластерный порт primary должен быть доступен с secondary;
- репликация включается через API `sys/replication/dr/*` после инициализации обоих кластеров.

Порядок включения репликации и переключения описан в разделе [«Репликация для аварийного восстановления (DR)»](../../../admin/replication/disaster-recovery/).
