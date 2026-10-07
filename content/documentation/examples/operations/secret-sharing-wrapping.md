---
title: "One-time secret handover with response wrapping"
linkTitle: "Handing a secret to a person"
description: "Securely handing a password or key to another person once with response wrapping: -wrap-ttl, token lookup, and unwrap."
weight: 100
params:
  relatedLinks:
    - title: "Response wrapping"
      url: ../../../concepts/response-wrapping/
    - title: "Policies"
      url: ../../../concepts/policy/
    - title: "Cubbyhole secrets engine"
      url: ../../../user/secrets-engines/cubbyhole/
---

Passwords sent in a messenger or by email remain in the chat history. Response wrapping lets you send a short-lived one-time token instead of the secret: the secret can be retrieved with it only once, and an interception attempt is detectable.

## Goal

Hand a password to a colleague so that it can be read only once, within a limited time, and so that the recipient can make sure the token was not replaced or used before them.

## Prerequisites

- The sender has a Stronghold token with permission to read the secret being handed over or to `sys/wrapping/wrap`.
- The recipient can reach the Stronghold address. The recipient does not need to log in to Stronghold: the wrapping token itself is enough to unwrap.
- A separate communication channel for sending the token (messenger, email).

## Step 1. Wrap the secret

Choose one of the options.

- **The secret is already stored in KV.** Read it with the `-wrap-ttl` flag:

  ```bash
  d8 stronghold kv get -wrap-ttl=1h -mount=secret handover/db-admin
  ```

- **Arbitrary data.** Wrap it with `sys/wrapping/wrap` without storing anything in KV:

  ```bash
  d8 stronghold write -wrap-ttl=1h sys/wrapping/wrap \
    username=db-admin \
    password='T3mp-P@ss'
  ```

The response contains the `wrapping_token`, `wrapping_token_ttl`, `wrapping_token_creation_time`, and `wrapping_token_creation_path` fields. Note the `creation_path`: the recipient will check it.

Choose the smallest sufficient `-wrap-ttl`: the shorter the lifetime, the smaller the interception window.

## Step 2. Send the token

1. Send the `wrapping_token` to the recipient.
1. Separately (for example, by voice), tell them the expected creation path: `secret/data/handover/db-admin` or `sys/wrapping/wrap`.

## Step 3. Check the token (recipient)

Before unwrapping, check the token properties. This operation does not consume the token and does not require authentication:

```bash
d8 stronghold write sys/wrapping/lookup token=<wrapping_token>
```

Make sure that:

- the token exists and has not expired;
- `creation_path` matches the one the sender told you;
- `creation_time` matches the time it was sent.

If the token is not found or the path does not match, do not unwrap it and report to your security team: the token may have been intercepted or replaced.

## Step 4. Unwrap the secret (recipient)

```bash
d8 stronghold unwrap <wrapping_token>
```

The command returns the original response. After that, the token becomes invalid.

## Verification

1. Run `d8 stronghold unwrap <wrapping_token>` again: the command must return an error because the token is single-use.
1. The sender checks that the token was used: `sys/wrapping/lookup` for this token returns an error.
1. With auditing enabled (Stronghold EE), find the `sys/wrapping/unwrap` request in the log and make sure it came from the expected recipient.

## Mandatory wrapping for the handover directory

To prevent reading secrets from the `handover/` directory without wrapping, set `min_wrapping_ttl` and `max_wrapping_ttl` in the policy used to read them:

```bash
d8 stronghold policy write handover-readers - <<'POLICY'
path "secret/data/handover/*" {
  capabilities = ["read"]
  min_wrapping_ttl = "1s"
  max_wrapping_ttl = "24h"
}
POLICY
```

Grant write access to the directory with a separate policy without these parameters.

For details, see [Required response wrapping TTLs](../../../concepts/policy/#required-response-wrapping-ttls).

## Cleanup

1. Delete the source secret if it is no longer needed:

   ```bash
   d8 stronghold kv metadata delete -mount=secret handover/db-admin
   ```

1. If the recipient has not unwrapped the token and the handover is cancelled, revoke the token so that it cannot be used before its TTL expires:

   ```bash
   d8 stronghold token revoke <wrapping_token>
   ```

1. Ask the recipient to change the temporary password after the first login.
