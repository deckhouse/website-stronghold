---
title: "Double encryption"
weight: 20
description: "Seal wrap: an extra layer of encryption for sensitive Stronghold data."
---

`Seal wrap` is a mechanism that adds an extra encryption layer on top of the normal storage cryptographic barrier. In Stronghold documentation, it can be treated as **double encryption**: data is protected by Stronghold's built-in barrier and then additionally wrapped through the configured `seal` mechanism.

This approach is especially useful in environments with elevated key protection requirements and in infrastructures that rely on external KMS or HSM systems.

## What seal wrap provides

The additional encryption layer can help with:

- stronger protection of critical security parameters;
- compliance-driven environments and stricter internal security policies;
- deployments where the most sensitive values must be additionally protected by an external `seal` mechanism;
- integrating Stronghold into infrastructures that use KMS or HSM systems.

## How it works

With `seal wrap`, Stronghold stores selected data not only encrypted by the built-in storage barrier, but also wrapped again through a supporting `seal`.

In practice, this means:

- the root key and other critical parameters gain an extra protection layer;
- some values in auth methods and secrets engines can be stored in seal-wrapped form;
- the storage and replication model remains transparent to users, but reads and writes of such values can take longer.

## What is usually seal-wrapped

Double encryption primarily applies to critical security parameters in the Stronghold core, including:

- the root key;
- the keyring;
- the recovery key and related unseal material;
- other internal critical Stronghold parameters.

In addition, supporting backends and plugins may use `seal wrap` for selected sensitive values. The original Vault documentation gives examples such as:

- auth methods including `LDAP`, `RADIUS`, `Okta`, and `AWS`;
- `PKI` for issuer data and keys;
- `SSH` for CA keys;
- `Transit` for keys and key policy data;
- `KMIP` for managed objects and CA keys.

## Enabling and disabling

For supporting `seal` mechanisms, Vault enables `seal wrap` by default. In Stronghold, you can disable the extra wrapping for values other than the root key with:

```hcl
disable_sealwrap = true
```

This parameter is described in the [Standalone configuration](../standalone/configuration/).

Important details:

- disabling `seal wrap` does not disable root key protection by the `seal`;
- changing `disable_sealwrap` takes effect gradually as values are read or rewritten;
- enabling it again is also a lazy process rather than an immediate full rewrap of storage.

## Sealwrap rewrap

When the `seal` configuration changes, Stronghold must wrap all seal-wrapped values again so they match the current encryption configuration. This process is called **seal rewrap**.

`Rewrap` is especially useful:

- after migrating to a new `seal`;
- after changing `seal` configuration parameters;
- after rotating key material in an external KMS or HSM.

Stronghold exposes the following API endpoints for rewrap management:

| Method | Path |
|--------|------|
| `GET`  | `/sys/sealwrap/rewrap` |
| `POST` | `/sys/sealwrap/rewrap` |

- `GET /sys/sealwrap/rewrap` returns the current process status and counters for processed entries;
- `POST /sys/sealwrap/rewrap` starts a rewrap process if one is not already running.

Example request:

```shell
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  --request POST \
  "${VAULT_ADDR}/v1/sys/sealwrap/rewrap"
```

During rewrap:

- the process runs in the background;
- progress is best monitored through `GET /sys/sealwrap/rewrap` and server logs;
- you should avoid making further `seal` configuration changes until it finishes.

Rewrap duration depends on the number of seal-wrapped values and the performance characteristics of the external `seal` backend.

## Operational considerations

- `seal wrap` works only with `seal` mechanisms that support it;
- if you use an external HSM or remote KMS, seal-wrapped reads and writes can be noticeably slower than normal operations;
- this matters most for write-heavy workloads and environments with unstable connectivity to the HSM or KMS;
- if the external `seal` becomes unavailable during runtime, part of Stronghold functionality may be affected.

## Seal wrap and replication

`Seal wrap` exists below replication logic. This means:

- replication carries the data and the requirement that a value should be seal-wrapped;
- the actual additional encryption is performed by the local `seal` on each cluster;
- different clusters can use different local HSM or KMS keys while keeping the same protection model.

## Practical recommendations

- Use `seal wrap` when you need stronger protection for key material and other critical values.
- Evaluate the performance impact before enabling it broadly in production.
- For HSM integrations, verify that connectivity to the device is stable.
- Change `disable_sealwrap` in a controlled way, understanding that rewrapping happens gradually.

## See also

- [HSM support](./hsm/)
- [Standalone configuration](../standalone/configuration/)
