---
title: "HSM support"
weight: 10
params:
  edition: ee
---

Stronghold supports root key encryption using Hardware Security Modules (HSM) such as TPM2, JaCarta, and other devices that support the PKCS #11 standard.

For testing and development, you can also use the SoftHSM2 program emulator.

{{< alert level="warning" >}}
Currently, HSM is only supported in standalone Stronghold installations. In the examples below, it's assumed that you use a local configuration file and a `seal "pkcs11"` section in the standalone server configuration.
{{< /alert >}}

To use automatic unsealing via PKCS #11, start by creating keys in the HSM and configuring Stronghold to use them.

## SoftHSM2

To test Stronghold integration with the HSM, you can use SoftHSM2 emulator.
Do the following to create a token, generate a key pair and configure Stronghold to use the key.

1. Install the required packages:

   ```shell
   apt install libsofthsm2 opensc
   ```

1. Create a directory for keeping the SoftHSM2 data and the configuration file:

   ```shell
   mkdir /home/stronghold/softhsm
   cd softhsm
   echo "directories.tokendir = /home/stronghold/softhsm/" > /home/stronghold/softhsm2.conf
   ```

1. Set a path to the SoftHSM2 configuration and the PKCS #11 library:

   ```shell
   export SOFTHSM2_CONF=/home/stronghold/softhsm2.conf
   HSMLIB="/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so"
   ```

1. Initialize a token and set the PIN codes:

   ```shell
   pkcs11-tool --module $HSMLIB --init-token --so-pin 1234 --init-pin --pin 4321 --label my_token --login
   ```

   Example output:

   ```console
   Using slot 0 with a present token (0x0)
   Token successfully initialized
   User PIN successfully initialized
   ```

1. Ensure the token has been created and initialized:

   ```shell
   pkcs11-tool --module $HSMLIB -L
   ```

   Example output:

   ```console
   Available slots:
   Slot 0 (0xe6829d3): SoftHSM slot ID 0xe6829d3
     token label        : my_token
     token manufacturer : SoftHSM project
     token model        : SoftHSM v2
     token flags        : login required, rng, token initialized, PIN initialized, other flags=0x20
     hardware version   : 2.6
     firmware version   : 2.6
     serial num         : 6a5468368e6829d3
     pin min/max        : 4/255
   Slot 1 (0x1): SoftHSM slot ID 0x1
     token state:   uninitialized
   ```

1. Create an RSA key pair in the token:

   ```shell
   pkcs11-tool --module $HSMLIB --login --pin 4321 --keypairgen --key-type rsa:4096 --label "vault-rsa-key"
   ```

   Example output:

   ```console
   Using slot 0 with a present token (0xe6829d3)
   Key pair generated:
   Private Key Object; RSA
     label:      vault-rsa-key
     Usage:      decrypt, sign, signRecover, unwrap
     Access:     sensitive, always sensitive, never extractable, local
   Public Key Object; RSA 4096 bits
     label:      vault-rsa-key
     Usage:      encrypt, verify, verifyRecover, wrap
     Access:     local
   ```

1. Create a Stronghold configuration file (`config.hcl`) and add the PKCS #11 parameters into it:

   ```console
   api_addr="https://0.0.0.0:8200"
   log_level = "warn"
   ui = true

   listener "tcp" {
     address = "0.0.0.0:8200"
     tls_cert_file = "/home/stronghold/cert.pem"
     tls_key_file  = "/home/stronghold/key.pem"
     #tls_require_and_verify_client_cert = true
     #tls_client_ca_file = "ca.crt"
     tls_disable = "false"
   }
   storage "raft" {
     path = "/home/stronghold/data"
   }

   seal "pkcs11" {
     lib = "/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so"
     token_label = "my_token"
     pin = "4321"
     key_label = "vault-rsa-key"
     rsa_oaep_hash = "sha1"
   }
   ```

1. Start Stronghold while specifying the SoftHSM2 configuration:

   ```shell
   export SOFTHSM2_CONF=/home/stronghold/softhsm2.conf
   d8 stronghold server -config config.hcl
   ```

## Migration from Shamir keys to HSM

1. Modify the Stronghold configuration by adding the `seal "pkcs11"` section:

   ```console
   seal "pkcs11" {
     lib = "/usr/lib/librtpkcs11ecp.so"
     token_label = "my_token"
     pin = "12345678"
     key_label = "vault-rsa-key"
   }
   ```

1. Restart Stronghold. The logs should show a message:

   ```console
   2025-04-03T17:08:13.431+0300 [WARN]  core: entering seal migration mode; Stronghold will not automatically unseal even if using an autoseal: from_barrier_type=shamir to_barrier_type=pkcs11
   ```

1. Perform the migration by entering the unseal keys:

   ```shell
   d8 stronghold operator unseal -migrate
   ```

After the migration is complete, Stronghold will automatically unseal using PKCS #11 on restart.

## Migration from HSM to Shamir keys

1. Modify the configuration by adding the parameter `disabled = "true"` to the `seal "pkcs11"` section:

   ```console
   seal "pkcs11" {
     lib = "/usr/lib/librtpkcs11ecp.so"
     token_label = "my_token"
     pin = "12345678"
     key_label = "vault-rsa-key"
     disabled = "true"
   }
   ```

1. Restart Stronghold.

1. Perform the migration by entering the recovery keys:

   ```shell
   d8 stronghold operator unseal -migrate
   ```

After the migration is complete, Stronghold will require manual entry of unseal keys on each restart.
