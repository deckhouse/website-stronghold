---
title: "Recover from lost quorum"
weight: 40
---

Quorum is the minimum number of nodes in a cluster required to vote and reach consensus.

In the `stronghold` module, high-availability (HA) mode is enabled by default and relies on the Raft consensus algorithm. Maintaining Raft quorum is an important factor when operating a Stronghold environment. When there is no way to restore enough healthy Stronghold nodes, the Stronghold cluster permanently loses quorum and can no longer reach consensus or elect a leader. Without a leader, Stronghold can no longer perform read and write operations for clients.

Stronghold for Deckhouse Platform (DP) is delivered as a module. Each Stronghold node runs in a container inside a separate pod. Each pod is scheduled one-by-one onto a master node of the DP cluster. As a result, the number of Stronghold cluster nodes is updated dynamically when master nodes are added to or removed from the DP cluster. Stronghold calculates quorum with the formula `(n+1)/2`, where `n` by default equals the number of master nodes in the DP cluster. For a Stronghold cluster of 3 nodes, this means at least 2 healthy pods are required for the cluster to function, `(3+1)/2 = 2`. In particular, 2 **permanently** active pods are required to perform read and write operations.

{{< alert level="info" >}}
There is an exception to this rule if you use the `-non-voter` option while joining the cluster. This feature is available only in Stronghold as a standalone installation.
{{< /alert >}}

## Signs of lost quorum

If 2 out of 3 Stronghold pods in a cluster have the `Ready: False` status, the cluster loses quorum and becomes inoperable.

Despite one fully operational node, the cluster cannot process read or write requests.

The following are the error examples when quorum is lost.

{{< tabs name="stronghold_cmd_6115" >}}
{{% tab name="Stronghold in DP" %}}

- Attempt to obtain a list of nodes in a Raft cluster:

  ```shell
  d8 stronghold operator raft list-peers
  ```

  Example output:

  ```text
  * local node not active but active cluster node not found
  ```

- Attempt to get the data from storage:

  ```shell
  d8 stronghold kv get kv/apikey
  ```

  Example output:

  ```text
  * local node not active but active cluster node not found
  ```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

- Attempt to obtain a list of nodes in a Raft cluster:

  ```shell
  stronghold operator raft list-peers
  ```

  Example output:

  ```text
  * local node not active but active cluster node not found
  ```

- Attempt to get the data from storage:

  ```shell
  stronghold kv get kv/apikey
  ```

  Example output:

  ```text
  * local node not active but active cluster node not found
  ```

{{% /tab %}}
{{< /tabs >}}

The non-operational node logs may contain the following entries:

```text
{"@level":"info","@message":"attempting to join possible raft leader node","@module":"core","@timestamp":"2025-10-20T10:54:02.578963Z","leader_addr":"https://stronghold-0.stronghold-internal:8300"}
{"@level":"error","@message":"failed to get raft challenge","@module":"core","@timestamp":"2025-10-20T10:54:32.597558Z","error":"error during raft bootstrap init call: Put \"https://10.0.12.69:8300/v1/sys/storage/raft/bootstrap/challenge\": dial tcp 10.10.12.69:8300: i/o timeout","leader_addr":"https://stronghold-0.stronghold-internal:8300"}
```

To recover Stronghold after the loss of 2 out of 3 nodes, temporarily convert the cluster into a single-node cluster. At least one server must be fully operational to complete this procedure.

{{< alert level="info" >}}
Sometimes Stronghold may lose quorum because of incorrect addition or removal of a master node in DP. In that case, before the recovering procedure via `peers.json`, stop Stronghold pods on the inoperative nodes. For example, by cordoning the corresponding nodes temporarily.

In a 5-server cluster or when voting nodes are absent, before the `peers.json` recovery, stop the other healthy servers.
{{< /alert >}}

## Step 1. Locate the storage directory

On the DP master node with the healthy Stronghold instance, navigate to the Raft storage directory at `/var/lib/deckhouse/stronghold/`. Make sure it contains a non-empty `node-id` file.

## Step 2. Create the peers.json file

In the storage directory `/var/lib/deckhouse/stronghold/`, there is a subdirectory named `raft`.

```text
stronghold
├── raft
│   ├── raft.db
│   └── snapshots
├── vault.db
└── node-id
```

To enable the single remaining Stronghold server to reach quorum and elect itself as the leader, create a `peers.json` file in the `raft` subdirectory. In the file, specify the ID of the healthy Stronghold node (`node-id`), its address and port, as well as whether it can vote.

Example of a command creating the file:

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

Parameter description:

- `id` (required): Stronghold server ID.
- `address` (required): Address and port of the server. The port is the server's cluster port.
- `non_voter`: Specifies whether the server participates in voting. To allow it to participate, set to `false`.

Set the `deckhouse:deckhouse` owner and the `600` permissions for the `peers.json` file:

```bash
chown deckhouse:deckhouse /var/lib/deckhouse/stronghold/raft/peers.json
chmod 600 /var/lib/deckhouse/stronghold/raft/peers.json
```

## Step 3. Restart the Stronghold pod

Restart the pod with the operational Stronghold instance (named `stronghold-0` in the example) so that Stronghold can load the newly created `peers.json` file.

## Step 4. Unseal Stronghold

If automatic unseal is not configured, unseal Stronghold and then check the status.

{{< tabs name="stronghold_cmd_96020" >}}
{{% tab name="Stronghold in DP" %}}

1. Unseal Stronghold and enter the unseal key:

   ```bash
   d8 stronghold operator unseal
   Unseal Key (will be hidden):
   ```

1. Check the Stronghold status:

   ```bash
   d8 stronghold status
   ```

   Example output:

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
{{% tab name="Stronghold in Linux" %}}

1. Unseal Stronghold and enter the unseal key:

   ```bash
   stronghold operator unseal
   Unseal Key (will be hidden):
   ```

1. Check the Stronghold status:

   ```bash
   stronghold status
   ```

   Example output:

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

## Step 5. Verify the recovery success

The recovery procedure is considered successful if Stronghold starts and displays the following messages in the logs:

```text
...
[INFO]  core.cluster-listener: serving cluster requests: cluster_listen_address=[::]:8201
[INFO]  storage.raft: raft recovery initiated: recovery_file=peers.json
[INFO]  storage.raft: raft recovery found new config: config="{[{Voter stronghold_1 https://10.0.101.22:8201}]}"
[INFO]  storage.raft: raft recovery deleted peers.json
...
```

After the recovery, only one server must be listed in the cluster. This allows Stronghold to reach quorum and restore operation. To verify the number of servers, run the following command.

{{< tabs name="stronghold_cmd_75226" >}}
{{% tab name="Stronghold in DP" %}}

```bash
d8 stronghold operator raft list-peers
```

Example output:

```console
Node                                    Address                                  State       Voter
----                                    -------                                  -----       -----
d3816d62-29eb-4f42-98cb-f25ab05e8fbd    stronghold-0.stronghold-internal:8301    leader      true
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```bash
stronghold operator raft list-peers
```

Example output:

```console
Node                                    Address                                  State       Voter
----                                    -------                                  -----       -----
d3816d62-29eb-4f42-98cb-f25ab05e8fbd    stronghold-0.stronghold-internal:8301    leader      true
```

{{% /tab %}}
{{< /tabs >}}

As shown, the cluster peer list contains only one server.

## Next steps

In this guide, you recovered quorum by converting a 3-node cluster into a single-node cluster using the `peers.json` file. This file let you manually update the Raft peer list to the single remaining healthy node, which allowed that server to reach quorum and successfully elect a leader.

If the failed nodes are recoverable, bring them back to the cluster using the previously used host addresses. This returns the cluster to a fully healthy state. For that, include the server ID, address and port, as well as information whether it can vote for each server you want in the cluster in the `raft/peers.json` file.

Configuration example for three Stronghold servers:

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
