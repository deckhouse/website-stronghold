---
title: "Восстановление после потери кворума"
weight: 40
---

Кворум – минимальное количество узлов в кластере для возможности выполнения голосования с целью достижения консенсуса.
В модуле Stronghold по умолчанию включен режим HA, который в свою очередь опирается на алгоритм консенсуса Raft. Наличие кворума для Raft вляется важным фактором эксплуатации среды Stronghold. Когда нет возможности восстановить достаточное количество рабочих узлов Stronghold, кластер Stronghold окончательно теряет кворум и вместе с этим возможность для достижения консенсуса и избрания лидера. В конечном итоге, без наличия лидера, Stronghold больше не может выполнять операции чтения и записи для клиента.

Stronghold для DKP поставляется в виде модуля, и каждый его **узел** запускается в отдельном контейнере отдельных **подов**, и каждый *под* поштучно привязывается на *control-plane* (т.е. *master*) узлы кластера DKP. Как следствие, количество узлов кластера Stronghold динамически обновляется при подключении новых и удалении существующих мастер узлов кластера DKP. Таким образом, Stronghold будет рассчитывать кворум по формуле `(n+1)/2`, где `n` — количество мастер узлов в кластере DKP. Полученное число значит, что для функционирования кластера Stronghold из 3 узлов потребуется как минимум 2 рабочих пода, `(3+1)/2 = 2`. В частности, для выполнения операций чтения и записи потребуется 2 **постоянно** активных пода.

{{< alert level="info" >}}
Существует исключение из этого правила, если при присоединении к кластеру используется опция `-non-voter`. Эта функция доступна только в версии Stronghold формата отдельной инсталляции.
{{< /alert >}}

## Обзор сценария

Когда статус `Ready` у двух из трёх подов равен `False`, кластер теряет кворум и перестает функционировать.

Несмотря на один полностью работоспособный узел, кластер не может обрабатывать запросы на чтение или запись.

**Примеры:**
Вывод команд в консоли, при наличии ошибки:

```text
$ d8 stronghold operator raft list-peers
* local node not active but active cluster node not found

$ d8 stronghold kv get kv/apikey
* local node not active but active cluster node not found
```

В логах нерабочего узла:

```text
{"@level":"info","@message":"attempting to join possible raft leader node","@module":"core","@timestamp":"2025-10-20T10:54:02.578963Z","leader_addr":"https://stronghold-0.stronghold-internal:8300"}
{"@level":"error","@message":"failed to get raft challenge","@module":"core","@timestamp":"2025-10-20T10:54:32.597558Z","error":"error during raft bootstrap init call: Put \"https://10.0.12.69:8300/v1/sys/storage/raft/bootstrap/challenge\": dial tcp 10.10.12.69:8300: i/o timeout","leader_addr":"https://stronghold-0.stronghold-internal:8300"}
```

Процесс восстановления работы Stronghold при потере 2 из 3 узлов будет выполнена путем преобразования кластера в вариант из одного узла.

Для выполнения этой процедуры обязательно, чтобы один сервер был полностью работоспособным.

{{< alert level="info" >}}
Иногда Stronghold теряет кворум из-за некорректного добавления/удаления мастер узла в DKP. В таком случае на неработоспособных узлах необходимо остановить запуск подов со Stronghold, перед запуском процедуры peers.json. Например, временно через cordon узлов.

В кластере из 5 серверов или в случае отсутствия голосующих нужно остановить другие исправные серверы перед выполнением восстановления peers.json.
{{< /alert >}}

## Найдите каталог хранилища

На мастер ноде DKP с исправным узлом Stronghold найдите каталог хранилища Raft по пути `/var/lib/deckhouse/stronghold/`. В этом катологе проверьте наличие непустого файла `node-id`. Если всё получилось, двигаемся дальше.

## Создайте файл peers.json

Внутри каталога хранилища (`/var/lib/deckhouse/stronghold/`) находится папка с именем `raft`.

```text
stronghold
├── raft
│   ├── raft.db
│   └── snapshots
├── vault.db
└── node-id
```

Чтобы единственный оставшийся сервер Stronghold мог достичь кворума и избрать себя лидером, создайте файл `raft/peers.json`, содержащий информацию о сервере. Формат файла — массив JSON, содержащий ID работоспособного узла Stronghold (`node-id`), его *адрес:порт* и информацию о возможности голосовать.

**Пример:**

```bash
$ cat > /var/lib/deckhouse/stronghold/raft/peers.json << EOF
[
  {
    "id": "`cat /var/lib/deckhouse/stronghold/node-id`",
    "address": "stronghold-0.stronghold-internal:8301",
    "non_voter": false
  }
]
EOF
```

Параметры:
- **id** (строка: \<обязательно\>) — указывает идентификатор сервера.
- **address** (строка: \<обязательно\>) — указывает хост и порт сервера. Порт — это порт кластера сервера.
- **non_voter** (bool: \<false\>) — указывает, участвует ли сервер в голосовании.

Убедитесь, что файл `peers.json` содержит корректные права на файл:

```bash
chown deckhouse:deckhouse /var/lib/deckhouse/stronghold/raft/peers.json
chmod 600 /var/lib/deckhouse/stronghold/raft/peers.json
```

## Перезапустите *под* Stronghold

Перезапустите *под* (`stronghold-0` в примере), чтобы Stronghold мог загрузить новый файл `peers.json`.

## Распечатайте Stronghold

Если не настроено использование автоматической распечатки, распечатайте Stronghold, а затем проверьте статус.

**Пример:**

```bash
$ d8 stronghold operator unseal
Unseal Key (will be hidden):

$ d8 stronghold status
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

## Проверка успешности

Процедура восстановления прошла успешно, если Stronghold запустился и отобразил следующие сообщения в логах.

```text
...
[INFO]  core.cluster-listener: serving cluster requests: cluster_listen_address=[::]:8201
[INFO]  storage.raft: raft recovery initiated: recovery_file=peers.json
[INFO]  storage.raft: raft recovery found new config: config="{[{Voter stronghold_1 https://10.0.101.22:8201}]}"
[INFO]  storage.raft: raft recovery deleted peers.json
...
```

## Просмотр списка узлов

Теперь в кластере числится только один сервер. Это позволило Stronghold достичь кворума и восстановить работоспособность. Чтобы убедиться в их количестве, выполните команду `d8 stronghold operator raft list-peers`.

```bash
$ d8 stronghold operator raft list-peers
Node                                    Address                                  State       Voter
----                                    -------                                  -----       -----
d3816d62-29eb-4f42-98cb-f25ab05e8fbd    stronghold-0.stronghold-internal:8301    leader      true
```

Как видно, в списке узлов кластера указан только один сервер.

## Следующие шаги

В этом руководстве мы восстановили кворум, преобразовав кластер из 3 узлов в кластер из одного узла с помощью файла `peers.json`. Файл `peers.json` позволил нам вручную обновить список узлов Raft оставив единственный работоспособный узел, что позволило этому серверу достичь кворума и успешно выборать лидера.

Если вышедшие из строя узлы **поддаются восстановлению**, лучшим вариантом будет вернуть их в сеть и подключить к кластеру с использованием тех же адресов хостов. Это вернет кластер в полностью рабочее состояние. Для этого в файле `raft/peers.json` должны быть указаны данные: идентификатор сервера, *адрес:порт* и информация о возможности голосовать, для каждого сервера, который вы хотите включить в кластер.

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
