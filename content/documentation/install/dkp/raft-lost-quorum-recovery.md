---
title: "Recover from lost quorum"
weight: 40
---

Quorum is the minimum number of nodes in a cluster required to vote and reach consensus.

In the `stronghold` module, HA mode is enabled by default and relies on the Raft consensus algorithm. Maintaining Raft quorum is an important factor when operating a Stronghold environment. When there is no way to restore enough healthy Stronghold nodes, the Stronghold cluster permanently loses quorum and can no longer reach consensus or elect a leader. Without a leader, Stronghold can no longer perform read and write operations for clients.

Stronghold for Deckhouse Kubernetes Platform is delivered as a module. Each Stronghold **node** runs in a container inside a separate **Pod**, and each Pod is scheduled one-to-one onto a control-plane (master) node of the DKP cluster. As a result, the number of Stronghold cluster nodes is updated dynamically when master nodes are added to or removed from the DKP cluster. Stronghold calculates quorum with the formula `(n+1)/2`, where `n` by default equals the number of master nodes in the DKP cluster. For a Stronghold cluster of 3 nodes, this means at least 2 healthy Pods are required for the cluster to function, `(3+1)/2 = 2`. In particular, 2 **permanently** active Pods are required to perform read and write operations.

{{< alert level="info" >}}
There is an exception to this rule if you use the `-non-voter` option while joining the cluster. This feature is available only in Stronghold as a standalone installation.
{{< /alert >}}

## Scenario overview

When the `Ready` status of two out of three Pods is `False`, the cluster loses quorum and becomes inoperable.

Despite one fully operational node, the cluster cannot process read or write requests.

**Examples:**

Console command output when an error is present:

```text
$ d8 stronghold operator raft list-peers
* local node not active but active cluster node not found

$ d8 stronghold kv get kv/apikey
* local node not active but active cluster node not found
```

Logs from an inoperative node:

```text
{"@level":"info","@message":"attempting to join possible raft leader node","@module":"core","@timestamp":"2025-10-20T10:54:02.578963Z","leader_addr":"https://stronghold-0.stronghold-internal:8300"}
{"@level":"error","@message":"failed to get raft challenge","@module":"core","@timestamp":"2025-10-20T10:54:32.597558Z","error":"error during raft bootstrap init call: Put \"https://10.0.12.69:8300/v1/sys/storage/raft/bootstrap/challenge\": dial tcp 10.10.12.69:8300: i/o timeout","leader_addr":"https://stronghold-0.stronghold-internal:8300"}
```

Recovery of Stronghold after the loss of 2 out of 3 nodes is performed by converting the cluster into a single-node cluster.

One server must be fully operational to complete this procedure.

{{< alert level="info" >}}
Sometimes Stronghold loses quorum because of incorrect addition or removal of a master node in DKP. In that case, stop Stronghold Pods on the inoperative nodes before starting the `peers.json` procedure. For example, temporarily cordon the nodes.

In a 5-server cluster or when voting nodes are absent, stop the other healthy servers before performing the `peers.json` recovery.
{{< /alert >}}

## Locate the storage directory

On the DKP master node with the healthy Stronghold node, locate the Raft storage directory at `/var/lib/deckhouse/stronghold/`. In that directory, check that a non-empty `node-id` file exists. If this step is successful, proceed.

## Create the peers.json file

Inside the storage directory (`/var/lib/deckhouse/stronghold/`), there is a folder named `raft`.

```text
stronghold
├── raft
│   ├── raft.db
│   └── snapshots
├── vault.db
└── node-id
```

To enable the single remaining Stronghold server to reach quorum and elect itself as the leader, create a `raft/peers.json` file that holds the server information. The file format is a JSON array containing the ID of the healthy Stronghold node (`node-id`), its address:port, and suffrage information.

**Example:**

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

Parameters:

- **id** (string: \<required\>) — specifies the server ID.
- **address** (string: \<required\>) — specifies the host and port of the server. The port is the server's cluster port.
- **non_voter** (bool: \<false\>) — specifies whether the server participates in voting.

Make sure the `peers.json` file has the correct permissions:

```bash
chown deckhouse:deckhouse /var/lib/deckhouse/stronghold/raft/peers.json
chmod 600 /var/lib/deckhouse/stronghold/raft/peers.json
```

## Restart the Stronghold Pod

Restart the Pod (`stronghold-0` in the example) so that Stronghold can load the new `peers.json` file.

## Unseal Stronghold

If automatic unseal is not configured, unseal Stronghold and then check the status.

**Example:**

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

## Verify success

The recovery procedure is successful if Stronghold starts and displays the following messages in the logs.

```text
...
[INFO]  core.cluster-listener: serving cluster requests: cluster_listen_address=[::]:8201
[INFO]  storage.raft: raft recovery initiated: recovery_file=peers.json
[INFO]  storage.raft: raft recovery found new config: config="{[{Voter stronghold_1 https://10.0.101.22:8201}]}"
[INFO]  storage.raft: raft recovery deleted peers.json
...
```

## View the peer list

The cluster now lists only one server. This allowed Stronghold to reach quorum and restore operation. To verify the number of servers, run `d8 stronghold operator raft list-peers`.

```bash
$ d8 stronghold operator raft list-peers
Node                                    Address                                  State       Voter
----                                    -------                                  -----       -----
d3816d62-29eb-4f42-98cb-f25ab05e8fbd    stronghold-0.stronghold-internal:8301    leader      true
```

As shown, the cluster peer list contains only one server.

## Next steps

In this guide, you recovered quorum by converting a 3-node cluster into a single-node cluster using the `peers.json` file. The `peers.json` file let you manually update the Raft peer list to the single remaining healthy node, which allowed that server to reach quorum and successfully elect a leader.

If the failed nodes are **recoverable**, the best option is to bring them back online and reconnect them to the cluster using the same host addresses. This returns the cluster to a fully healthy state. For that, the `raft/peers.json` file must include the server ID, address:port, and suffrage information for each server you want in the cluster.

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
