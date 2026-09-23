---
title: "Восстановление после потери кворума"
weight: 40
---

Кворум — это минимальное количество узлов в кластере, необходимое для голосования с целью достижения консенсуса.

В модуле `stronghold` по умолчанию включен режим высокой доступности (high availability, HA), который опирается на алгоритм консенсуса Raft. Наличие кворума для Raft является важным фактором эксплуатации среды Stronghold. Когда нет возможности восстановить достаточное количество рабочих узлов Stronghold, кластер окончательно теряет кворум и вместе с этим возможность для достижения консенсуса и избрания лидера. В конечном итоге, без наличия лидера, Stronghold больше не может выполнять операции чтения и записи для клиента.

Stronghold для Deckhouse Platform (DP) поставляется в виде модуля, и каждый его узел запускается в отдельном контейнере отдельных подов. Каждый под поштучно привязывается на master-узлы кластера DP. Как следствие, количество узлов кластера Stronghold динамически обновляется при подключении новых и удалении существующих master-узлов кластера DP. Таким образом, Stronghold будет рассчитывать кворум по формуле `(n+1)/2`, где `n` — по умолчанию равно количеству master-узлов в кластере DP. В случае кластера из 3 узлов полученное число значит, что для функционирования кластера Stronghold потребуется как минимум 2 рабочих пода, `(3+1)/2 = 2`. В частности, для выполнения операций чтения и записи потребуется 2 **постоянно** активных пода.

{{< alert level="info" >}}
Существует исключение из этого правила, если при присоединении к кластеру используется опция `-non-voter`. Эта функция доступна только при установке Stronghold в виде отдельной инсталляции.
{{< /alert >}}

## Признаки потери кворума

Если в кластере у 2 из 3 подов Stronghold зафиксирован статус `Ready: False`, кластер теряет кворум и перестает функционировать.

Несмотря на один полностью работоспособный узел, кластер не может обрабатывать запросы на чтение или запись.

Далее приведены примеры ошибок при потере кворума.

{{< tabs name="stronghold_cmd_6115" >}}
{{% tab name="Stronghold в DP" %}}

- Попытка получить список узлов Raft-кластера:

  ```shell
  d8 stronghold operator raft list-peers
  ```

  Пример вывода:

  ```text
  * local node not active but active cluster node not found
  ```

- Попытка получить данные из хранилища:

  ```shell
  d8 stronghold kv get kv/apikey
  ```

  Пример вывода:

  ```text
  * local node not active but active cluster node not found
  ```

{{% /tab %}}
{{% tab name="Stronghold в Linux" %}}

- Попытка получить список узлов Raft-кластера:

  ```shell
  stronghold operator raft list-peers
  ```

  Пример вывода:

  ```text
  * local node not active but active cluster node not found
  ```

- Попытка получить данные из хранилища:

  ```shell
  stronghold kv get kv/apikey
  ```

  Пример вывода:

  ```text
  * local node not active but active cluster node not found
  ```

{{% /tab %}}
{{< /tabs >}}

В логах нерабочего узла могут появляться следующие сообщения:

```text
{"@level":"info","@message":"attempting to join possible raft leader node","@module":"core","@timestamp":"2025-10-20T10:54:02.578963Z","leader_addr":"https://stronghold-0.stronghold-internal:8300"}
{"@level":"error","@message":"failed to get raft challenge","@module":"core","@timestamp":"2025-10-20T10:54:32.597558Z","error":"error during raft bootstrap init call: Put \"https://10.0.12.69:8300/v1/sys/storage/raft/bootstrap/challenge\": dial tcp 10.10.12.69:8300: i/o timeout","leader_addr":"https://stronghold-0.stronghold-internal:8300"}
```

Для восстановления Stronghold после потери 2 из 3 узлов необходимо временно преобразовать кластер в одноузловой. Для выполнения процедуры необходим как минимум один полностью работоспособный сервер.

{{< alert level="info" >}}
Иногда Stronghold может потерять кворум из-за некорректного добавления или удаления master-узла в DP. В таком случае перед восстановлением с помощью `peers.json` на неработоспособных узлах необходимо остановить запуск подов со Stronghold. Например, через временный cordon соответствующих узлов.

В кластере из 5 серверов или в случае отсутствия голосующих узлов перед восстановлением с помощью `peers.json` нужно остановить другие работоспособные серверы.
{{< /alert >}}

## Шаг 1. Перейдите в каталог хранилища

На master-узле DP с исправным экземпляром Stronghold перейдите в каталог хранилища Raft `/var/lib/deckhouse/stronghold/`. Убедитесь, что в каталоге находится непустой файл `node-id`.

## Шаг 2. Создайте файл peers.json

В каталоге хранилища `/var/lib/deckhouse/stronghold/` находится подкаталог `raft`.

```text
stronghold
├── raft
│   ├── raft.db
│   └── snapshots
├── vault.db
└── node-id
```

Чтобы оставшийся сервер Stronghold мог достичь кворума и избрать себя лидером, создайте файл `peers.json` в подкаталоге `raft`. В файле укажите идентификатор экземпляра Stronghold из `node-id`, его адрес и порт, а также возможность участвовать в голосовании.

Пример команды для создания файла:

```bash
cat > /var/lib/deckhouse/stronghold/raft/peers.json << EOF
[
  {
    "id": "`cat /var/lib/deckhouse/stronghold/node-id`",
    "address": "stronghold-0.stronghold-internal:8301",
    "non_voter": false
  }
]
EOF
```

Описание параметров:

- `id` (обязательный) — идентификатор сервера Stronghold;
- `address` (обязательный) — адрес и порт сервера. В качестве порта укажите порт кластера сервера;
- `non_voter` — указывает, участвует ли сервер в голосовании. Для участия укажите `false`.

Установите для файла `peers.json` владельца `deckhouse:deckhouse` и права доступа `600`:

```bash
chown deckhouse:deckhouse /var/lib/deckhouse/stronghold/raft/peers.json
chmod 600 /var/lib/deckhouse/stronghold/raft/peers.json
```

## Шаг 3. Перезапустите под Stronghold

Перезапустите под с работоспособным экземпляром Stronghold (`stronghold-0` в примере), чтобы Stronghold мог загрузить созданный файл `peers.json`.

## Шаг 4. Распечатайте Stronghold

Если автоматическая распечатка не настроена, распечатайте Stronghold, а затем проверьте его статус.

{{< tabs name="stronghold_cmd_96020" >}}
{{% tab name="Stronghold в DP" %}}

1. Распечатайте Stronghold и введите unseal-ключ:

   ```bash
   d8 stronghold operator unseal
   Unseal Key (will be hidden):
   ```

1. Проверьте статус Stronghold:

   ```bash
   d8 stronghold status
   ```

   Пример вывода:

   ```console
   Key                      Value
   ---                      -----
   Recovery Seal Type       shamir
   Initialized              true
   Sealed                   false
   Total Recovery Shares    1
   Threshold                1
   Version                  1.16.8+ee
   Storage Type             raft
   Cluster Name             stronghold-cluster-4a1a40af
   Cluster ID               d09df2c7-1d3e-f7d0-a9f7-93fadcc29110
   HA Enabled               true
   HA Cluster               https://stronghold-0.stronghold-internal:8301
   HA Mode                  active
   Active Since             2021-07-20T00:07:32.215236307Z
   Raft Committed Index     155344
   Raft Applied Index       155344
   ```

{{% /tab %}}
{{% tab name="Stronghold в Linux" %}}

1. Распечатайте Stronghold и введите unseal-ключ:

   ```bash
   stronghold operator unseal
   Unseal Key (will be hidden):
   ```

1. Проверьте статус Stronghold:

   ```bash
   stronghold status
   ```

   Пример вывода:

   ```console
   Key                      Value
   ---                      -----
   Recovery Seal Type       shamir
   Initialized              true
   Sealed                   false
   Total Recovery Shares    1
   Threshold                1
   Version                  1.16.8+ee
   Storage Type             raft
   Cluster Name             stronghold-cluster-4a1a40af
   Cluster ID               d09df2c7-1d3e-f7d0-a9f7-93fadcc29110
   HA Enabled               true
   HA Cluster               https://stronghold-0.stronghold-internal:8301
   HA Mode                  active
   Active Since             2021-07-20T00:07:32.215236307Z
   Raft Committed Index     155344
   Raft Applied Index       155344
   ```

{{% /tab %}}
{{< /tabs >}}

## Шаг 5. Проверьте результат восстановления

Процедура восстановления считается успешной, если Stronghold запустился и отобразил следующие сообщения в логах:

```text
...
[INFO]  core.cluster-listener: serving cluster requests: cluster_listen_address=[::]:8201
[INFO]  storage.raft: raft recovery initiated: recovery_file=peers.json
[INFO]  storage.raft: raft recovery found new config: config="{[{Voter stronghold_1 https://10.0.101.22:8201}]}"
[INFO]  storage.raft: raft recovery deleted peers.json
...
```

После восстановления в кластере должен числиться только один сервер. Это позволяет Stronghold достичь кворума и восстановить работоспособность. Чтобы убедиться в количестве серверов, выполните следующую команду.

{{< tabs name="stronghold_cmd_75226" >}}
{{% tab name="Stronghold в DP" %}}

```bash
d8 stronghold operator raft list-peers
```

Пример вывода:

```console
Node                                    Address                                  State       Voter
----                                    -------                                  -----       -----
d3816d62-29eb-4f42-98cb-f25ab05e8fbd    stronghold-0.stronghold-internal:8301    leader      true
```

{{% /tab %}}
{{% tab name="Stronghold в Linux" %}}

```bash
stronghold operator raft list-peers
```

Пример вывода:

```console
Node                                    Address                                  State       Voter
----                                    -------                                  -----       -----
d3816d62-29eb-4f42-98cb-f25ab05e8fbd    stronghold-0.stronghold-internal:8301    leader      true
```

{{% /tab %}}
{{< /tabs >}}

Как видно, в списке узлов кластера указан только один сервер.

## Следующие шаги

Следуя указаниям руководства, вы восстановили кворум, преобразовав кластер из 3 узлов в кластер из одного узла с помощью файла `peers.json`. Этот файл позволил вручную обновить список узлов Raft, оставив единственный работоспособный узел, в результате чего был восстановлен кворум и успешно выбран лидер.

Если вышедшие из строя узлы поддаются восстановлению, верните их в кластер, используя прежние адреса. Это вернет кластер в полностью рабочее состояние. Для этого в файле `raft/peers.json` укажите идентификатор сервера, его адрес и порт, а также информацию о возможности участия в голосовании для каждого сервера, который вы хотите включить в кластер.

Пример конфигурации для трёх серверов Stronghold:

```json
[
  {
    "id": "d3816d62-29eb-4f42-98cb-f25ab05e8fbd",
    "address": "stronghold-0.stronghold-internal:8301",
    "non_voter": false
  },
  {
    "id": "20247ff6-3fd0-4a19-af39-6b173714ccd9",
    "address": "stronghold-1.stronghold-internal:8301",
    "non_voter": false
  },
  {
    "id": "1be581fc-fc9b-45f6-b36a-ecb6e73b108e",
    "address": "stronghold-2.stronghold-internal:8301",
    "non_voter": false
  }
]
```
