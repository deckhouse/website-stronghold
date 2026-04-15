---
title: "Установка"
linkTitle: "Установка"
description: "Установка Deckhouse Stronghold в Standalone"
weight: 20
---

## Установка

В этом разделе описан пример установки **Deckhouse Stronghold** в режиме **Standalone** с использованием встроенного backend-хранилища `Raft`.

В качестве основного сценария рассматривается кластер из **трёх узлов**, который обеспечивает конфигурацию с высокой доступностью (HA). Такой вариант рекомендуется для production-окружений.

> Примечание  
> Для тестовых сценариев возможна установка на один сервер. В этом случае часть шагов, связанных с кластером `Raft`, можно упростить.

## Перед началом

Убедитесь, что вы выполнили действия из раздела [Подготовка](../prepare/), а именно:

- выбрали целевую топологию развёртывания;
- подготовили серверы;
- подготовили TLS-сертификаты;
- открыли необходимые порты;
- разместили дистрибутив Stronghold на целевых узлах.

В примерах ниже используются:

- TCP-порт `8200` — для API;
- TCP-порт `8201` — для сервер-серверного взаимодействия в кластере `Raft`;
- рабочая директория `/opt/stronghold`;
- пользователь `stronghold`;
- имена узлов:
  - `raft-node-1.demo.tld`
  - `raft-node-2.demo.tld`
  - `raft-node-3.demo.tld`

## Шаг 1. Подготовьте пользователя и каталоги

На каждом узле создайте системного пользователя и необходимые директории.

```bash
useradd --system --home /opt/stronghold --shell /sbin/nologin stronghold || true

mkdir -p /opt/stronghold/data
mkdir -p /opt/stronghold/tls

chown -R stronghold:stronghold /opt/stronghold
chmod 0700 /opt/stronghold/data
chmod 0700 /opt/stronghold/tls
```

Если бинарный файл Stronghold ещё не размещён, скопируйте его в рабочую директорию или в каталог, используемый в вашей инфраструктуре.

## Шаг 2. Создайте unit-файл systemd

На каждом узле создайте файл `/etc/systemd/system/stronghold.service`:

```ini
[Unit]
Description=Stronghold service
Documentation=https://deckhouse.ru/products/stronghold/
After=network.target

[Service]
Type=simple
ExecStart=/opt/stronghold/stronghold server -config=/opt/stronghold/config.hcl
ExecReload=/bin/kill -HUP $MAINPID
KillMode=process
Restart=on-failure
RestartSec=5
User=stronghold
Group=stronghold
LimitNOFILE=65536
CapabilityBoundingSet=CAP_IPC_LOCK
AmbientCapabilities=CAP_IPC_LOCK
SecureBits=noroot

[Install]
WantedBy=multi-user.target
```

Примените изменения и включите автозапуск сервиса:

```bash
systemctl daemon-reload
systemctl enable stronghold.service
```

> Примечание  
> В примерах предполагается, что сервис запускается от имени пользователя `stronghold`. Если используется другой пользователь, скорректируйте unit-файл и права доступа.

## Шаг 3. Подготовьте TLS-сертификаты

Для настройки TLS разместите сертификаты и ключи в каталоге `/opt/stronghold/tls`.

Минимально потребуется:

- сертификат корневого центра сертификации: `stronghold-ca.pem`;
- сертификаты узлов:
  - `node-1-cert.pem`
  - `node-2-cert.pem`
  - `node-3-cert.pem`;
- закрытые ключи узлов:
  - `node-1-key.pem`
  - `node-2-key.pem`
  - `node-3-key.pem`.

> Предупреждение  
> Приведённый ниже пример с `OpenSSL` и самоподписанными сертификатами подходит только для тестовых сценариев и лабораторных стендов. Для production-окружений используйте сертификаты, выпущенные доверенным центром сертификации.

### Пример генерации самоподписанных сертификатов

На первом узле создайте каталог для сертификатов и перейдите в него:

```bash
mkdir -p /opt/stronghold/tls
cd /opt/stronghold/tls/
```

Сгенерируйте ключ для корневого сертификата:

```bash
openssl genrsa 2048 > stronghold-ca-key.pem
```

Создайте корневой сертификат:

```bash
openssl req -new -x509 -nodes -days 3650 \
  -key stronghold-ca-key.pem \
  -out stronghold-ca.pem
```

Для каждого узла создайте конфигурационный файл с `subjectAltName`. Например, для первого узла:

```bash
cat << EOF > node-1.cnf
[v3_ca]
subjectAltName = @alt_names
[alt_names]
DNS.1 = raft-node-1.demo.tld
IP.1 = 10.20.30.10
IP.2 = 127.0.0.1
EOF
```

Аналогично создайте файлы `node-2.cnf` и `node-3.cnf`, указав корректные DNS-имена и IP-адреса.

Сформируйте запросы на сертификаты и ключи:

```bash
openssl req -newkey rsa:2048 -nodes \
  -keyout node-1-key.pem \
  -out node-1-csr.pem \
  -subj "/CN=raft-node-1.demo.tld"

openssl req -newkey rsa:2048 -nodes \
  -keyout node-2-key.pem \
  -out node-2-csr.pem \
  -subj "/CN=raft-node-2.demo.tld"

openssl req -newkey rsa:2048 -nodes \
  -keyout node-3-key.pem \
  -out node-3-csr.pem \
  -subj "/CN=raft-node-3.demo.tld"
```

Выпустите сертификаты узлов:

```bash
openssl x509 -req -set_serial 01 -days 3650 \
  -in node-1-csr.pem \
  -out node-1-cert.pem \
  -CA stronghold-ca.pem \
  -CAkey stronghold-ca-key.pem \
  -extensions v3_ca \
  -extfile ./node-1.cnf

openssl x509 -req -set_serial 02 -days 3650 \
  -in node-2-csr.pem \
  -out node-2-cert.pem \
  -CA stronghold-ca.pem \
  -CAkey stronghold-ca-key.pem \
  -extensions v3_ca \
  -extfile ./node-2.cnf

openssl x509 -req -set_serial 03 -days 3650 \
  -in node-3-csr.pem \
  -out node-3-cert.pem \
  -CA stronghold-ca.pem \
  -CAkey stronghold-ca-key.pem \
  -extensions v3_ca \
  -extfile ./node-3.cnf
```

Скопируйте на каждый узел:

- сертификат соответствующего узла;
- закрытый ключ соответствующего узла;
- `stronghold-ca.pem`.

Например:

```bash
scp ./node-2-key.pem ./node-2-cert.pem ./stronghold-ca.pem \
  raft-node-2.demo.tld:/opt/stronghold/tls

scp ./node-3-key.pem ./node-3-cert.pem ./stronghold-ca.pem \
  raft-node-3.demo.tld:/opt/stronghold/tls
```

## Шаг 4. Подготовьте конфигурационный файл

На первом узле создайте файл `/opt/stronghold/config.hcl`:

```hcl
ui = true
cluster_addr  = "https://10.20.30.10:8201"
api_addr      = "https://10.20.30.10:8200"
disable_mlock = true

listener "tcp" {
  address       = "0.0.0.0:8200"
  tls_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
  tls_key_file  = "/opt/stronghold/tls/node-1-key.pem"
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

На втором и третьем узлах создайте аналогичный файл, изменив:

- `cluster_addr`;
- `api_addr`;
- `node_id`;
- пути к сертификату и ключу текущего узла.

> Важно  
> Значения `api_addr`, `cluster_addr`, DNS-имён и IP-адресов должны соответствовать вашей инфраструктуре и данным в TLS-сертификатах.

## Шаг 5. Откройте необходимые порты

Если на узлах используется `firewalld`, откройте порты `8200` и `8201`:

```bash
firewall-cmd --add-port=8200/tcp --permanent
firewall-cmd --add-port=8201/tcp --permanent
firewall-cmd --reload
```

Если в вашей инфраструктуре используется другой файрвол, настройте эквивалентные правила.

## Шаг 6. Запустите Stronghold на первом узле

Запустите сервис:

```bash
systemctl start stronghold
```

Проверьте его состояние:

```bash
systemctl status stronghold
```

Если сервис не запустился, проверьте:

- корректность `config.hcl`;
- наличие сертификатов и ключей;
- права доступа к каталогам и файлам;
- доступность используемых портов.

## Шаг 7. Инициализируйте кластер

Инициализацию выполняют **один раз**, только на первом узле.

```bash
stronghold operator init -ca-cert /opt/stronghold/tls/stronghold-ca.pem
```

При необходимости можно использовать параметры:

- `-key-shares` — количество частей ключа;
- `-key-threshold` — минимальное количество частей, достаточное для распечатывания хранилища.

После выполнения команды будут выведены:

- части ключа распечатывания;
- root-токен.

> Предупреждение  
> Обязательно сохраните части ключа и root-токен в надёжном месте. Без достаточного числа частей ключа доступ к данным Stronghold будет невозможен.

## Шаг 8. Распечатайте хранилище

После инициализации распечатайте первый узел. Выполните команду несколько раз, вводя части ключа распечатывания:

```bash
stronghold operator unseal -ca-cert /opt/stronghold/tls/stronghold-ca.pem
```

Если используется значение по умолчанию, потребуется ввести **3 части ключа** из **5**.

## Шаг 9. Запустите остальные узлы

На втором и третьем узлах:

1. проверьте, что размещены корректные сертификаты и конфигурация;
2. запустите сервис:

   ```bash
   systemctl start stronghold
   ```

3. не выполняйте `operator init`;
4. выполните распечатывание:

   ```bash
   stronghold operator unseal -ca-cert /opt/stronghold/tls/stronghold-ca.pem
   ```

После этого узлы присоединятся к кластеру `Raft`.

## Шаг 10. Выполните первичную проверку

Проверьте состояние кластера на одном из узлов:

```bash
stronghold status -ca-cert /opt/stronghold/tls/stronghold-ca.pem
```

Пример ожидаемого вывода:

```text
Key                     Value
---                     -----
Seal Type               shamir
Initialized             true
Sealed                  false
Total Shares            5
Threshold               3
Version                 1.15.2
Build Date              2025-03-07T16:10:46Z
Storage Type            raft
Cluster Name            stronghold-cluster-a3fcc270
Cluster ID              f682968d-5e6c-9ad4-8303-5aecb259ca0b
HA Enabled              true
HA Cluster              https://10.20.30.10:8201
HA Mode                 active
Active Node Address     https://10.20.30.10:8200
Raft Committed Index    40
Raft Applied Index      40
```

Если в выводе отображаются:

- `Initialized: true`;
- `Sealed: false`;
- `Storage Type: raft`;
- `HA Enabled: true`;

значит базовая установка выполнена успешно.

> Примечание  
> Полная проверка работоспособности, включая дополнительные проверки доступа и состояния сервиса, приведена в разделе **«Проверка работоспособности»**.

## Что дальше

После завершения установки перейдите к разделу [Первичная настройка](../initial-configuration/).
