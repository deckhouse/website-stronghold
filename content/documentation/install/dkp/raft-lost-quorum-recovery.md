---
title: "Recover from lost quorum"
weight: 40
---

Quorum is the minimum number of nodes (or "votes") required for the cluster to operate

With Integrated Storage, Raft quorum maintenance is a consideration for configuring and operating your Stronghold environment. A Stronghold cluster permanently loses quorum when there is no way to recover enough servers to reach consensus and elect a leader. Without a quorum of cluster servers, Stronghold can no longer perform read and write operations.

The cluster quorum is dynamically updated when new servers join the cluster. Stronghold calculates quorum with the formula `(n+1)/2`, where `n` is the number of servers in the cluster. For example, for a 3-server cluster, you will need at least 2 servers operational for the cluster to function properly, `(3+1)/2 = 2`. Specifically, you will need 2 servers always active to perform read and write operations.

{{< alert level="info" >}}
There is an exception to this rule if you use the `-non-voter` option while joining the cluster. This feature is available only in Stronghold as a standalone.
{{< /alert >}}

## Scenario overview

When two of the three servers encountered an outage, the cluster loses quorum and becomes inoperable.

Although one of the servers is fully functioning, the cluster won't be able to process read or write requests.

**Example:**
Command outputs for this case:

```text
$ d8 stronghold operator raft list-peers
No raft cluster configuration found

$ d8 stronghold kv get kv/apikey
nil response from pre-flight request
```

Failing Pod logs:

```text
{"@level":"info","@message":"attempting to join possible raft leader node","@module":"core","@timestamp":"2025-10-20T10:54:02.578963Z","leader_addr":"https://stronghold-0.stronghold-internal:8300"}
{"@level":"error","@message":"failed to get raft challenge","@module":"core","@timestamp":"2025-10-20T10:54:32.597558Z","error":"error during raft bootstrap init call: Put \"https://10.0.12.69:8300/v1/sys/storage/raft/bootstrap/challenge\": dial tcp 10.10.12.69:8300: i/o timeout","leader_addr":"https://stronghold-0.stronghold-internal:8300"}
```

In this tutorial, you will recover from the permanent loss of two-of-three Stronghold servers by converting it into a single-server cluster.

The last server must be fully operational to complete this procedure.

{{< alert level="info" >}}
Sometimes Stronghold loses quorum due to autopilot and servers marked as unhealthy but the service is still running. On unhealthy server(s), you must stop services before running the peers.json procedure.

In a 5 server cluster or in the case of non voters, you must stop other healthy before performing the peers.json recovery.
{{< /alert >}}

## Locate the storage directory

On the healthy Stronghold Pod's corresponding DKP master server, locate the Raft storage directory on path `/var/lib/deckhouse/stronghold/`. Also check for `node-id` file existence. If this step is successful, proceed.

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

To enable the single, remaining Stronghold server to reach quorum and elect itself as the leader, create a `raft/peers.json` file that holds the server information. The file format is a JSON array containing the server ID, address:port, and suffrage information of the healthy Stronghold server.

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

- **id** (string: \<required\>) - Specifies the server ID of the server.
- **address** (string: \<required\>) - Specifies the host and port of the server. The port is the server's cluster port.
- **non_voter** (bool: \<false\>) - This controls whether the server is a non-voter.

Make sure file `peers.json` has valid read/write permissions and owner:

```bash
chown deckhouse:deckhouse /var/lib/deckhouse/stronghold/raft/peers.json
chmod 600 /var/lib/deckhouse/stronghold/raft/peers.json
```

## Restart Stronghold Pod

Restart the Stronghold Pod to enable Stronghold to load the new `peers.json` file.

## Unseal Stronghold

If not configured to use auto-unseal, unseal Stronghold and then check the status.

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

The recovery procedure is successful when Stronghold starts up and displays these messages in the system logs.

```text
...snip...
[INFO]  core.cluster-listener: serving cluster requests: cluster_listen_address=[::]:8201
[INFO]  storage.raft: raft recovery initiated: recovery_file=peers.json
[INFO]  storage.raft: raft recovery found new config: config="{[{Voter stronghold_1 https://10.0.101.22:8201}]}"
[INFO]  storage.raft: raft recovery deleted peers.json
...snip...
```

## View the peer list

You now have a cluster with one server that can reach the quorum. Verify that there is just one server in the cluster with `d8 stronghold operator raft list-peers` command.

```bash
$ d8 stronghold operator raft list-peers
Node                                    Address                                  State       Voter
----                                    -------                                  -----       -----
d3816d62-29eb-4f42-98cb-f25ab05e8fbd    stronghold-0.stronghold-internal:8301    leader      true
```

## Next steps

In this tutorial, you recovered the loss of quorum by converting a 3-server cluster into a single-server cluster using the `peers.json`. The `peers.json` file enabled you to manually overwrite the Raft peer list to the one remaining server, which allowed that server to reach quorum and complete a leader election.

If the failed servers are **recoverable**, the best option is to bring them back online and have them reconnect to the cluster using the same host addresses. This will return the cluster to a fully healthy state. In such an event, the `raft/peers.json` should contain the server ID, address:port, and suffrage information of each Stronghold server you wish to be in the cluster.

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
