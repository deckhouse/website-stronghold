---
title: "GitOps Secrets engine"
linkTitle: "Overview"
weight: 10
---

The GitOps secrets engine monitors a Git repository and applies declarative Stronghold configuration when new commits have enough verified signatures.

Key properties:

- Configuration is described in YAML and applied through the Stronghold API.
- Apply state is stored in Stronghold.
- The engine connects to Stronghold with an address and token from its own configuration.
- A renewable periodic token is required. Stronghold renews it automatically 24 hours before expiration.
- Status and errors are available at `gitops/status`.
- The engine can manage the same Stronghold instance it runs on, or another instance.
- You can enable several mounts to manage different parts of the configuration from different repositories, within the permissions of each token.

The GitOps secrets engine is **built-in**. You do not need to register or load a plugin binary.

## Setup

1. Enable the GitOps secrets engine:

   ```bash
   d8 stronghold secrets enable gitops
   ```

   By default, the engine mounts at `gitops/`. To use another path, pass `-path`.

1. Configure the Git repository to monitor:

   ```bash
   d8 stronghold write gitops/configure/git_repository \
       git_repo_url="https://gitlab.example.com/org/stronghold-gitops.git" \
       required_number_of_verified_signatures_on_commit=1 \
       git_poll_period=1m
   ```

   | Parameter | Required | Default | Description |
   |-----------|----------|---------|-------------|
   | `git_repo_url` | yes (on create) | — | Git repository URL |
   | `git_branch_name` | no | `main` | Branch to track |
   | `git_poll_period` | no | `5m` | Interval between repository polls |
   | `required_number_of_verified_signatures_on_commit` | no | `0` | Minimum number of verified signatures on a commit before apply |
   | `git_ca_certificate` | no | empty | CA certificate for Git TLS verification |
   | `max_clone_size_bytes` | no | `10485760` (10 MiB) | Maximum in-memory clone size in bytes; `0` disables the limit |

1. If the repository is private, configure credentials:

   ```bash
   d8 stronghold write gitops/configure/git_credential \
       username=token \
       password=glpat-XXXXXXXX
   ```

1. Create PGP keys for commit signing:

   ```bash
   gpg --quick-generate-key "key1 <key1@example.com>" rsa4096
   gpg --quick-generate-key "key2 <key2@example.com>" rsa4096
   ```

1. Export the public keys and upload them to Stronghold:

   ```bash
   gpg --armor --output key1.pgp --export key1
   gpg --armor --output key2.pgp --export key2

   d8 stronghold write gitops/configure/trusted_pgp_public_key/key1 public_key=@key1.pgp
   d8 stronghold write gitops/configure/trusted_pgp_public_key/key2 public_key=@key2.pgp
   ```

1. Configure API access for the engine. Prefer a wrapped renewable periodic token. The engine unwraps the token and stores it; the token cannot be read back later:

   ```bash
   TOKEN=$(d8 stronghold token create -orphan -period=7d -policy=gitops-apply \
       -display-name="gitops-plugin" -wrap-ttl=1m -field=wrapping_token)

   d8 stronghold write gitops/configure/vault \
       vault_addr=https://stronghold.example.com:8200 \
       wrapping_token=$TOKEN
   ```

   | Parameter | Description |
   |-----------|-------------|
   | `vault_addr` | Stronghold API address |
   | `wrapping_token` | Wrapped token; preferred way to pass credentials |
   | `vault_token` | Token in plain form (use only if wrapping is unavailable) |
   | `vault_namespace` | Namespace for API calls |
   | `vault_cacert_bytes` | PEM-encoded CA certificate for TLS verification |
   | `rotate` | If `true` and a token is provided, create an orphan token with the same parameters and revoke the old one |

   Grant the token only the policies needed to manage the resources described in the Git repository.

## Sign and push configuration

1. Install [git-signatures](https://github.com/werf/3p-git-signatures).

1. Clone the configuration repository (or create a new one) and set the signing key:

   ```bash
   git clone https://gitlab.example.com/org/stronghold-gitops.git
   cd stronghold-gitops

   gpg --list-key
   git config user.signingKey <KEY_ID>
   ```

1. Add configuration files in the [declarative YAML format](./configuration-format/), commit, and sign:

   ```bash
   git add .
   git commit -m "Add AppRole auth method"
   git signatures add
   git signatures show
   ```

   Example `git signatures show` output:

   ```text
    Public Key ID    | Status     | Trust     | Date                         | Signer Name
   =====================================================================================================
    0C3AAAA10E30D5F3 | VALIDSIG   | ULTIMATE  | Mon 22 Dec 2025 20:19:33 MSK | key1 <key1@example.com>
   ```

1. Push the commit and signatures:

   ```bash
   git push origin main
   git signatures push
   ```

After the next poll, if the commit has enough verified signatures from trusted keys, the engine applies the configuration.

## Status

Check the current status:

```bash
d8 stronghold read gitops/status
```

The response includes:

- `status` — current process status or error;
- `last_run` — time of the last periodic run (`never` if the engine has not run yet);
- `last_finished_commit` — hash of the last successfully applied commit;
- `last_finished_commit_date` — date of that commit.

## Disable

```bash
d8 stronghold secrets disable gitops
```

Disabling the engine deletes its storage data for that mount, including the stored API token and apply state.
