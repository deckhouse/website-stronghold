---
title: "RabbitMQ secrets engine"
linkTitle: "RabbitMQ"
description: "Dynamic RabbitMQ credentials: connecting to the management HTTP API, roles with virtual host permissions and tags, requesting credentials, and lease configuration."
weight: 85
---

The RabbitMQ secrets engine generates RabbitMQ user credentials dynamically based on configured roles. For every request, Stronghold creates a separate RabbitMQ user with the virtual host (vhost) permissions and tags defined in the role, and deletes the user when the [lease](../../../concepts/lease/) expires.

Stronghold manages users through the RabbitMQ management HTTP API, so the `rabbitmq_management` plugin must be enabled in RabbitMQ.

The RabbitMQ secrets engine is available in all Stronghold editions.

## Capabilities

| Capability | Supported |
|------------|-----------|
| Dynamic credentials | Yes |
| Virtual host permissions (`vhosts`) | Yes |
| Topic exchange permissions (`vhost_topics`) | Yes |
| User tags (`tags`) | Yes |
| Username customization (`username_template`) | Yes |
| Password policy (`password_policy`) | Yes |

## Setup

Setup is usually performed by a Stronghold administrator or a configuration management tool.

1. Enable the RabbitMQ secrets engine:

   ```bash
   d8 stronghold secrets enable rabbitmq
   ```

   By default, the engine is mounted at `rabbitmq/`. To mount it at a different path, use the `-path` argument.

1. Configure the connection to the RabbitMQ management HTTP API. Specify a RabbitMQ user with administrator rights (the `administrator` tag): Stronghold uses it to create and delete users.

   ```bash
   d8 stronghold write rabbitmq/config/connection \
     connection_uri="https://rabbitmq.example.com:15672" \
     username="stronghold-admin" \
     password="<password>"
   ```

   Parameters of the `config/connection` endpoint:

   | Parameter | Description |
   |-----------|-------------|
   | `connection_uri` | RabbitMQ management HTTP API URI. |
   | `username` | RabbitMQ administrator username. |
   | `password` | RabbitMQ administrator password. |
   | `verify_connection` | Whether to verify `connection_uri` by actually connecting to the management API. Defaults to `true`. |
   | `username_template` | Template used to generate dynamic usernames. |
   | `password_policy` | Name of the [password policy](../../../concepts/password-policy/) used to generate passwords for dynamic users. |

   {{< alert level="warning" >}}
   Create a dedicated RabbitMQ user for Stronghold and do not use it for anything else.
   {{< /alert >}}

1. Configure the lease settings for issued credentials:

   ```bash
   d8 stronghold write rabbitmq/config/lease \
     ttl=1800 \
     max_ttl=3600
   ```

   - `ttl`: how long the credentials are valid without renewal, in seconds;
   - `max_ttl`: the maximum lifetime after which the credentials cannot be renewed, in seconds.

   The value `0` (default) means that system TTL values are used.

   To read the current lease settings:

   ```bash
   d8 stronghold read rabbitmq/config/lease
   ```

1. Create a role that describes the permissions of the generated users:

   ```bash
   d8 stronghold write rabbitmq/roles/my-role \
     vhosts='{"/":{"configure":".*","write":".*","read":".*"}}' \
     tags="management"
   ```

   Role parameters:

   | Parameter | Description |
   |-----------|-------------|
   | `vhosts` | JSON map of virtual hosts to `configure`, `write`, and `read` permissions (RabbitMQ regular expressions). |
   | `vhost_topics` | Nested JSON map: virtual host → exchange → `write` and `read` topic permissions. |
   | `tags` | Comma-separated list of RabbitMQ user tags, for example `management` or `monitoring`. |

   A role with topic permissions on the `amq.topic` exchange:

   ```bash
   d8 stronghold write rabbitmq/roles/topic-role \
     vhosts='{"/":{"configure":"","write":".*","read":".*"}}' \
     vhost_topics='{"/":{"amq.topic":{"write":"^orders\\..*","read":".*"}}}'
   ```

## Usage

Once the secrets engine is configured, a user or application with a Stronghold token and an appropriate policy can request credentials.

1. Request credentials for a role:

   ```bash
   d8 stronghold read rabbitmq/creds/my-role
   ```

   Example output:

   ```text
   Key                Value
   ---                -----
   lease_id           rabbitmq/creds/my-role/I39Hu8XXOombof4wiK5bKMn9
   lease_duration     30m
   lease_renewable    true
   password           3yNDBikgQvrkx2VA2zhq5IdSM7IWk1RyMYJr
   username           root-39669250-3894-8032-c420-3d58483ebfc4
   ```

   Stronghold creates a RabbitMQ user with the role permissions and returns its username and password together with the lease ID.

1. If needed, renew the lease before `lease_duration` expires:

   ```bash
   d8 stronghold lease renew rabbitmq/creds/my-role/I39Hu8XXOombof4wiK5bKMn9
   ```

1. When the credentials are no longer needed, revoke the lease. Stronghold deletes the RabbitMQ user:

   ```bash
   d8 stronghold lease revoke rabbitmq/creds/my-role/I39Hu8XXOombof4wiK5bKMn9
   ```

## Managing roles

- List roles:

  ```bash
  d8 stronghold list rabbitmq/roles
  ```

- Read a role:

  ```bash
  d8 stronghold read rabbitmq/roles/my-role
  ```

- Delete a role:

  ```bash
  d8 stronghold delete rabbitmq/roles/my-role
  ```

## Policy example

A policy that only allows requesting credentials for the `my-role` role:

```hcl
path "rabbitmq/creds/my-role" {
  capabilities = ["read"]
}
```

For the full list of endpoints and parameters, see the [secrets engines API reference](../../../reference/api/secrets/).

## Usage examples

Ready-made examples that use this feature:

- [Dynamic RabbitMQ users for a service](../../../examples/dynamic-credentials/rabbitmq/)

See all examples in [Usage examples](../../../examples/).
