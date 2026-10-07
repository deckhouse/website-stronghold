---
title: "trdl secrets engine"
linkTitle: "trdl"
description: "The built-in trdl plugin in Stronghold EE: secure build, signing, and publishing of releases from a Git repository gated by a quorum of commit signatures."
weight: 115
---

The `trdl` secrets engine is a built-in Stronghold EE plugin for secure software release delivery following the [trdl](https://trdl.dev/) model. The plugin builds release artifacts from a Git tag, signs them, publishes them to S3 storage, and publishes update channels. Builds and publishing run only for commits that have the required number of verified PGP signatures (a quorum).

Key properties:

- the source code and the release configuration (`trdl.yaml`) are stored in a Git repository;
- the update channel configuration (`trdl_channels.yaml`) can be stored in a separate branch;
- trusted PGP public keys and the required number of signatures are set in the plugin configuration;
- release artifacts are signed with a PGP key managed by the plugin;
- macOS build signing and ELF binary signing via Delivery Kit are supported;
- builds run asynchronously as tasks whose status and log are available via the API.

<!-- TODO(verify): clarify the relation to the werf/trdl client, the trdl.yaml and trdl_channels.yaml formats, and how commit signatures are stored; link to the trdl documentation if needed. -->

## Setup

1. Enable the secrets engine:

   ```bash
   d8 stronghold secrets enable -path=trdl-myapp trdl
   ```

1. Configure the plugin: the Git repository, the S3 storage for publishing, and the required number of verified commit signatures:

   ```bash
   d8 stronghold write trdl-myapp/configure \
     git_repo_url="https://git.example.com/org/myapp.git" \
     required_number_of_verified_signatures_on_commit=2 \
     s3_endpoint="https://s3.example.com" \
     s3_region="ru-central1" \
     s3_bucket_name="myapp-tuf" \
     s3_access_key_id="<access key ID>" \
     s3_secret_access_key="<secret access key>"
   ```

   Required parameters:

   | Parameter | Description |
   |-----------|-------------|
   | `git_repo_url` | Git repository URL. |
   | `required_number_of_verified_signatures_on_commit` | Required number of verified commit signatures. |
   | `s3_endpoint`, `s3_region`, `s3_bucket_name` | S3 storage endpoint, region, and bucket name. |
   | `s3_access_key_id`, `s3_secret_access_key` | S3 storage access credentials. |

   Optional parameters:

   | Parameter | Description |
   |-----------|-------------|
   | `git_trdl_path` | Path to the release configuration file in the repository. Defaults to `trdl.yaml`. |
   | `git_trdl_channels_path` | Path to the channel configuration file. Defaults to `trdl_channels.yaml`. |
   | `git_trdl_channels_branch` | Separate Git branch for the channel configuration file. |
   | `initial_last_published_git_commit` | Initial last successfully published commit. |
   | `buildx_driver`, `buildx_driver_opts` | buildx driver for builds (`docker-container` by default, or `kubernetes`) and its options. |
   | `buildkitd_address` | Address of a running buildkitd (`unix://`, `tcp://`, `docker-container://`, or `kube-pod://`). |
   | `buildkitd_driver`, `buildkitd_driver_opts` | Ephemeral buildkitd per build (for example, `kubernetes`) and its options. |

   {{< alert level="warning" >}}
   Build secrets are sent to the buildkitd daemon. A `tcp://` channel is neither encrypted nor authenticated, so you are responsible for securing the channel and isolating the daemon.
   {{< /alert >}}

1. Add trusted PGP public keys whose signatures count toward the quorum:

   ```bash
   d8 stronghold write trdl-myapp/configure/trusted_pgp_public_key \
     name="developer-1" \
     public_key=@developer-1.asc
   ```

   To list the added keys:

   ```bash
   d8 stronghold list trdl-myapp/configure/trusted_pgp_public_key
   ```

1. If the repository is private, configure Git credentials:

   ```bash
   d8 stronghold write trdl-myapp/configure/git_credential \
     username="trdl-bot" \
     password="<access token>"
   ```

1. If needed, add build secrets available while building artifacts:

   ```bash
   d8 stronghold write trdl-myapp/configure/build/secrets \
     id="npm-token" \
     data="<value>"
   ```

1. If needed, configure the task manager:

   ```bash
   d8 stronghold write trdl-myapp/task/configure \
     task_timeout=30m \
     task_history_limit=10
   ```

## Artifact signing

- The public part of the PGP key the plugin uses to sign release artifacts is available at `configure/pgp_signing_key`:

  ```bash
  d8 stronghold read trdl-myapp/configure/pgp_signing_key
  ```

  <!-- TODO(verify): whether the signing key is generated automatically and what exactly GET configure/pgp_signing_key returns. -->

- To sign macOS builds, set the certificate and Notary parameters in `configure/build/mac_signing_identity` (`data`, `password`, `notary_issuer`, `notary_key`, `notary_key_id`).
- To sign ELF binaries via Delivery Kit, set the certificate and private key in `configure/delivery_kit_elf_signing`. The key can be passed in base64 or as a `hashivault://<key>` reference to a Transit secrets engine key; in that case, set the `vault_*` parameters.

## Usage

1. Perform a release for a signed Git tag:

   ```bash
   d8 stronghold write trdl-myapp/release git_tag="v1.2.3"
   ```

   The plugin verifies the signatures, then builds and publishes the artifacts. The operation is asynchronous: the response contains a task ID.

   <!-- TODO(verify): response format (field holding the task UUID). -->

1. Publish update channels from `trdl_channels.yaml`:

   ```bash
   d8 stronghold write -force trdl-myapp/publish
   ```

1. Track tasks:

   ```bash
   d8 stronghold read trdl-myapp/task
   d8 stronghold read trdl-myapp/task/<uuid>
   d8 stronghold read trdl-myapp/task/<uuid>/log
   ```

   To cancel a running task:

   ```bash
   d8 stronghold write -force trdl-myapp/task/<uuid>/cancel
   ```

The last published commit is stored at `configure/last_published_git_commit`; you can read or delete it.

For the full list of endpoints and parameters, see the [secrets engines API reference](../../../reference/api/secrets/).
