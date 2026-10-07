---
title: "Auto-unseal and seal wrap with an HSM"
linkTitle: "Auto-unseal with an HSM"
description: "Protecting the Stronghold EE root key with a hardware module over PKCS #11: the seal pkcs11 stanza, a SoftHSM2 lab, initialization with recovery keys, double encryption, and recommendations for a production HSM."
weight: 30
params:
  edition: ee
  relatedLinks:
    - title: "HSM support"
      url: ../../../admin/kms-hsm/hsm/
    - title: "Double encryption"
      url: ../../../admin/kms-hsm/sealwrap/
    - title: "Seal and unseal"
      url: ../../../concepts/seal/
    - title: "Standalone configuration templates"
      url: ../../../install/standalone/topologies/
---

Stronghold EE encrypts the root key with a key that is stored in an HSM and never leaves it. The HSM is accessed through the standard PKCS #11 interface. After a restart, Stronghold unseals automatically, and the most sensitive data (the keyring, the recovery key, and PKI, SSH, and Transit keys) is additionally encrypted through the same seal: the seal wrap mechanism.

{{< alert level="warning" >}}
HSM is supported only in Stronghold EE and only in a standalone installation.
{{< /alert >}}

![Stronghold EE to HSM connection over PKCS #11](../../../images/ex-hsm-unseal.en.png)

## Goal

Build a SoftHSM2 lab, configure `seal "pkcs11"`, initialize Stronghold with recovery keys, verify auto-unseal and seal wrap, and then prepare to move the configuration to a hardware HSM.

## Prerequisites

- Stronghold EE installed on Linux as described in [Installation](../../../install/standalone/installation/). In the examples, the configuration is in `/opt/stronghold/config.hcl`, and the `stronghold` service runs as the `stronghold` user.
- For the lab: Debian or Ubuntu with the `softhsm2` and `opensc` packages (the `pkcs11-tool` utility). <!-- TODO(verify): SoftHSM2 package name in supported distributions (HSM support page lists libsofthsm2) -->
- For production: an HSM with PKCS #11 support (for example, Rutoken ECP 3.0 or JaCarta) and the vendor PKCS #11 library.
- To move an existing server from Shamir: the threshold number of unseal keys and a data backup.

## seal "pkcs11" parameters

| Parameter | Description |
| --- | --- |
| `lib` | Path to the HSM PKCS #11 library, for example `/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so` or `/usr/lib/librtpkcs11ecp.so` |
| `token_label` | Token label in the HSM |
| `pin` | Token user PIN |
| `key_label` | Label of the RSA key pair that encrypts the root key |
| `rsa_oaep_hash` | Hash function for RSA-OAEP (`sha1`, `sha224`, `sha256`, `sha384`, `sha512`); `sha256` by default, `sha1` in the SoftHSM2 example |
| `slot` | HSM slot number (instead of `token_label`) |
| `key_id` | Key pair identifier (instead of `key_label`) |
| `mechanism` | PKCS #11 mechanism used to encrypt the root key |
| `hmac_key_label` | HMAC key label |
| `generate_key` | Create the key in the HSM if it does not exist |
| `force_rw_session` | Open the PKCS #11 session in read-write mode |
| `disabled` | Used when migrating from the HSM to another seal |

Parameters can also be set with `VAULT_HSM_*` environment variables (for example, `VAULT_HSM_PIN`, `VAULT_HSM_LIB`) or `PKCS11_WRAPPER_*`; environment values take precedence over the configuration.

## Step 1. Prepare a SoftHSM2 token

Perform this step on the Stronghold node.

1. Install the packages:

   ```bash
   sudo apt install softhsm2 opensc
   ```

1. Create a token directory and a SoftHSM2 configuration accessible to the `stronghold` user:

   ```bash
   sudo mkdir -p /opt/stronghold/softhsm/tokens
   echo "directories.tokendir = /opt/stronghold/softhsm/tokens" | sudo tee /opt/stronghold/softhsm/softhsm2.conf
   sudo chown -R stronghold:stronghold /opt/stronghold/softhsm
   sudo chmod 0700 /opt/stronghold/softhsm/tokens
   ```

1. Set the variables:

   ```bash
   export SOFTHSM2_CONF=/opt/stronghold/softhsm/softhsm2.conf
   HSMLIB=/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so
   ```

1. Initialize the token and set the PINs. Run the commands as the `stronghold` user so that the token files belong to it:

   ```bash
   sudo -u stronghold SOFTHSM2_CONF=$SOFTHSM2_CONF \
     pkcs11-tool --module $HSMLIB --init-token --so-pin 1234 \
     --init-pin --pin 4321 --label stronghold_token --login
   ```

1. Create an RSA key pair:

   ```bash
   sudo -u stronghold SOFTHSM2_CONF=$SOFTHSM2_CONF \
     pkcs11-tool --module $HSMLIB --login --pin 4321 \
     --keypairgen --key-type rsa:4096 --label stronghold-rsa-key
   ```

   In the output, the private key must have the `sensitive, always sensitive, never extractable, local` attributes: the key cannot be extracted from the token.

1. Check the token and keys:

   ```bash
   sudo -u stronghold SOFTHSM2_CONF=$SOFTHSM2_CONF \
     pkcs11-tool --module $HSMLIB --login --pin 4321 --list-objects
   ```

## Step 2. Configure Stronghold

1. Add a `seal` stanza to `/opt/stronghold/config.hcl`:

   ```hcl
   seal "pkcs11" {
     lib           = "/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so"
     token_label   = "stronghold_token"
     pin           = "4321"
     key_label     = "stronghold-rsa-key"
     rsa_oaep_hash = "sha1"
   }
   ```

   Restrict permissions on the configuration file because it contains the PIN:

   ```bash
   sudo chown stronghold:stronghold /opt/stronghold/config.hcl
   sudo chmod 0600 /opt/stronghold/config.hcl
   ```

   The PIN can be passed through the `VAULT_HSM_PIN` or `PKCS11_WRAPPER_PIN` environment variable instead of the `pin` parameter in the configuration.

1. Pass the SoftHSM2 configuration path to the service through a systemd drop-in:

   ```bash
   sudo mkdir -p /etc/systemd/system/stronghold.service.d
   sudo tee /etc/systemd/system/stronghold.service.d/softhsm.conf <<'UNIT'
   [Service]
   Environment=SOFTHSM2_CONF=/opt/stronghold/softhsm/softhsm2.conf
   UNIT
   sudo systemctl daemon-reload
   ```

For a new server, go to step 3. For a server already initialized with Shamir, go to step 4.

## Step 3. Initialize a new server

1. Start Stronghold:

   ```bash
   sudo systemctl start stronghold
   ```

1. Initialize Stronghold with recovery keys:

   ```bash
   d8 stronghold operator init \
     -recovery-shares=5 \
     -recovery-threshold=3
   ```

   Distribute the recovery keys among key holders and store the root token.

## Step 4. Move an existing server from Shamir to the HSM

1. Take a Raft snapshot:

   ```bash
   d8 stronghold operator raft snapshot save pre-seal-migration.snap
   ```

1. Add the `seal "pkcs11"` stanza (step 2) and restart Stronghold:

   ```bash
   sudo systemctl restart stronghold
   ```

   The log shows the message:

   ```text
   core: entering seal migration mode; Stronghold will not automatically unseal even if using an autoseal: from_barrier_type=shamir to_barrier_type=pkcs11
   ```

1. Unseal with the `-migrate` flag, entering the threshold number of unseal keys:

   ```bash
   d8 stronghold operator unseal -migrate
   ```

1. Wait until the seal-wrapped entries are re-wrapped: the log must show the `seal re-wrap completed` message.

For an HA cluster, follow the general procedure in [Seal migration](../../../concepts/seal/#seal-migration).

## Step 5. Check seal wrap

For supported seals, double encryption is enabled by default, and no separate configuration is needed. Check the rewrap status:

```bash
d8 stronghold read sys/sealwrap/rewrap
```

The response contains the process status and counters of processed entries. To force re-wrapping of all seal-wrapped entries, for example after changing the key in the HSM, start a rewrap:

```bash
d8 stronghold write -f sys/sealwrap/rewrap
```

To disable double encryption for all data except the root key, set `disable_sealwrap = true` in the server configuration. For details, see [Double encryption](../../../admin/kms-hsm/sealwrap/).

## Verification

1. Check the status:

   ```bash
   d8 stronghold status
   ```

   Expected values:

   ```text
   Seal Type                pkcs11
   Recovery Seal Type       shamir
   Initialized              true
   Sealed                   false
   ```

1. Restart the service and make sure Stronghold unseals automatically:

   ```bash
   sudo systemctl restart stronghold
   sleep 10
   d8 stronghold status -format=json | jq '{type, sealed, recovery_seal}'
   ```

1. Check that Stronghold does not unseal without access to the HSM. In the lab, temporarily rename the token directory:

   ```bash
   sudo systemctl stop stronghold
   sudo mv /opt/stronghold/softhsm/tokens /opt/stronghold/softhsm/tokens.off
   sudo systemctl start stronghold
   d8 stronghold status
   journalctl -u stronghold.service --since "5 minutes ago" | grep -i pkcs11
   sudo systemctl stop stronghold
   sudo mv /opt/stronghold/softhsm/tokens.off /opt/stronghold/softhsm/tokens
   sudo systemctl start stronghold
   ```

   While the token is unavailable, `Sealed` must be `true`, and the log must contain a PKCS #11 error.

## Moving to a hardware HSM

- **Library and key.** Use the vendor PKCS #11 library and create a key pair in the HSM with the vendor tools or with `pkcs11-tool`. An example for Rutoken ECP 3.0 is in [HSM support](../../../admin/kms-hsm/hsm/). Choose the `rsa_oaep_hash` value according to the mechanisms the device supports.
- **Seal change.** Moving from SoftHSM2 to a hardware HSM is a migration between two auto seals: add `disabled = "true"` to the old `seal "pkcs11"` stanza, add the new stanza, and run `d8 stronghold operator unseal -migrate` with the recovery keys. The procedure is described in [Seal migration](../../../concepts/seal/#seal-migration). <!-- TODO(verify): whether two seal "pkcs11" stanzas can be specified at the same time for a pkcs11 → pkcs11 migration -->
- **HA cluster.** Each node needs access to a key with the same label. <!-- TODO(verify): HSM requirements in an HA cluster: a shared network HSM or key cloning between devices -->
- **Key backup.** Losing the key in the HSM makes Stronghold data impossible to decrypt. Use the key backup procedures provided by the HSM vendor.
- **Availability.** With seal wrap, the HSM is needed not only for unsealing but also at runtime: operations on seal-wrapped values go through the HSM and are slower than usual. Check that the connection to the device is stable.
- **PIN.** Store the PIN only in the configuration file with `0600` permissions for the `stronghold` user.

## Cleanup

For a lab environment:

1. Stop Stronghold: `sudo systemctl stop stronghold`.
1. Delete the Stronghold data, the SoftHSM2 token (`/opt/stronghold/softhsm`), and the `/etc/systemd/system/stronghold.service.d/softhsm.conf` drop-in, then run `sudo systemctl daemon-reload`.
1. Remove the `seal "pkcs11"` stanza from the configuration.
