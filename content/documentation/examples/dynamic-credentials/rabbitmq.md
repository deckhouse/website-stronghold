---
title: "Dynamic RabbitMQ users for a service"
linkTitle: "Dynamic RabbitMQ users"
description: "Issuing short-lived RabbitMQ users limited to their own virtual host and queues to a service with the rabbitmq secrets engine: connection, lease, role, policy, and verification with rabbitmqctl."
weight: 40
params:
  relatedLinks:
    - title: "RabbitMQ secrets engine"
      url: ../../../user/secrets-engines/rabbitmq/
    - title: "Leases, renewal, and revocation"
      url: ../../../concepts/lease/
    - title: "Delivering secrets to Kubernetes pods"
      url: ../../delivery/kubernetes-workloads/
    - title: "Secrets engines API"
      url: ../../../reference/api/secrets/
---

The service receives a dedicated RabbitMQ user from Stronghold with permissions only on its own virtual host and on resources with a given name prefix. Stronghold creates the user through the RabbitMQ management HTTP API and deletes it on revocation or when the lease expires.

## Goal

Configure issuing dynamic RabbitMQ credentials to the `orders` service: on the `orders` virtual host, the user gets `configure`, `write`, and `read` permissions only on queues and exchanges whose names start with `orders.`, and lives for 1 hour with renewal up to 24 hours.

## Prerequisites

- Stronghold and a token with permissions to configure secrets engines and policies.
- RabbitMQ with the `rabbitmq_management` plugin enabled; the management HTTP API is reachable from Stronghold.
- An existing `orders` virtual host, created for example with `rabbitmqctl add_vhost orders`.
- Access to `rabbitmqctl` on a RabbitMQ node for verification.

## Step 1. Prepare an account for Stronghold

Create a dedicated RabbitMQ user with the `administrator` tag. Stronghold uses it to create and delete users:

```bash
rabbitmqctl add_user stronghold-admin '<password>'
rabbitmqctl set_user_tags stronghold-admin administrator
```

Do not use this account for anything else.

## Step 2. Configure the secrets engine

1. Enable the RabbitMQ secrets engine:

   ```bash
   d8 stronghold secrets enable rabbitmq
   ```

1. Configure the connection to the management HTTP API:

   ```bash
   d8 stronghold write rabbitmq/config/connection \
     connection_uri="https://rabbitmq.example.com:15672" \
     username="stronghold-admin" \
     password="<password>"
   ```

1. Set the lease parameters in seconds. They apply to all roles of this secrets engine:

   ```bash
   d8 stronghold write rabbitmq/config/lease \
     ttl=3600 \
     max_ttl=86400
   ```

   To use different lease durations for different services, mount separate instances of the secrets engine with the `-path` argument.

## Step 3. Create a role

```bash
d8 stronghold write rabbitmq/roles/orders \
  vhosts='{"orders":{"configure":"^orders\\..*","write":"^orders\\..*","read":"^orders\\..*"}}'
```

- `vhosts`: a JSON map of virtual hosts and permissions. The `configure`, `write`, and `read` values are RabbitMQ regular expressions for queue and exchange names.
- `tags` is not set, so the user has no access to the management web UI and HTTP API.

If the service publishes to a `topic` exchange, restrict the allowed routing keys with `vhost_topics`:

```bash
d8 stronghold write rabbitmq/roles/orders \
  vhosts='{"orders":{"configure":"^orders\\..*","write":"^orders\\..*","read":"^orders\\..*"}}' \
  vhost_topics='{"orders":{"orders.events":{"write":"^order\\.(created|paid)$","read":".*"}}}'
```

## Step 4. Create a policy

```bash
d8 stronghold policy write orders-rabbitmq - <<'POLICY'
path "rabbitmq/creds/orders" {
  capabilities = ["read"]
}
POLICY
```

Assign the policy to the auth method role the service logs in with, for example a Kubernetes auth role as in [Dynamic PostgreSQL credentials](../postgresql/#step-4-create-a-kubernetes-auth-role). Renewing and revoking your own leases is allowed by the built-in `default` policy.

## Step 5. Get credentials

1. Request credentials:

   ```bash
   d8 stronghold read rabbitmq/creds/orders
   ```

   Example output:

   ```text
   Key                Value
   ---                -----
   lease_id           rabbitmq/creds/orders/I39Hu8XXOombof4wiK5bKMn9
   lease_duration     1h
   lease_renewable    true
   password           3yNDBikgQvrkx2VA2zhq5IdSM7IWk1RyMYJr
   username           token-39669250-3894-8032-c420-3d58483ebfc4
   ```

1. Renew the lease before `lease_duration` expires:

   ```bash
   d8 stronghold lease renew rabbitmq/creds/orders/I39Hu8XXOombof4wiK5bKMn9
   ```

1. When the credentials are no longer needed, revoke the lease. Stronghold deletes the user in RabbitMQ:

   ```bash
   d8 stronghold lease revoke rabbitmq/creds/orders/I39Hu8XXOombof4wiK5bKMn9
   ```

## Step 6. Deliver credentials to the service

Stronghold Agent renews the lease and re-renders the file when it receives new credentials. Example template that builds an AMQP connection URI:

```text
{{ with secret "rabbitmq/creds/orders" }}
AMQP_URL=amqps://{{ .Data.username }}:{{ .Data.password }}@rabbitmq.example.com:5671/orders
{{ end }}
```

The Agent sidecar configuration is in [Delivering secrets to Kubernetes pods](../../delivery/kubernetes-workloads/#stronghold-agent). When `max_ttl` is reached, the service receives a new user and the old one is deleted, so the service must reconnect to the broker with the new credentials.

## Verification

1. Request credentials and store the username and password:

   ```bash
   creds="$(d8 stronghold read -format=json rabbitmq/creds/orders)"
   MQ_USER="$(echo "$creds" | jq -r '.data.username')"
   MQ_PASSWORD="$(echo "$creds" | jq -r '.data.password')"
   LEASE_ID="$(echo "$creds" | jq -r '.lease_id')"
   ```

1. On a RabbitMQ node, check that the user exists and can authenticate:

   ```bash
   rabbitmqctl list_users
   rabbitmqctl authenticate_user "$MQ_USER" "$MQ_PASSWORD"
   ```

1. Check the user's virtual host permissions:

   ```bash
   rabbitmqctl list_user_permissions "$MQ_USER"
   ```

   The output must contain only the `orders` virtual host with the `^orders\..*` regular expressions.

1. Revoke the lease and make sure the user is deleted:

   ```bash
   d8 stronghold lease revoke "$LEASE_ID"
   rabbitmqctl list_users
   ```

If RabbitMQ runs in a DP cluster, run `rabbitmqctl` in a broker pod, for example `d8 k -n rabbitmq exec rabbitmq-0 -- rabbitmqctl list_users`.

## Cleanup

1. Revoke all credentials issued for the role:

   ```bash
   d8 stronghold lease revoke -prefix rabbitmq/creds/orders/
   ```

1. Delete the role and the policy, and disable the secrets engine if needed:

   ```bash
   d8 stronghold delete rabbitmq/roles/orders
   d8 stronghold policy delete orders-rabbitmq
   d8 stronghold secrets disable rabbitmq
   ```

1. Delete the Stronghold account in RabbitMQ: `rabbitmqctl delete_user stronghold-admin`.
