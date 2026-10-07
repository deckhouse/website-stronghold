---
title: "CI/CD"
description: "Getting Stronghold secrets in GitLab CI, GitHub Actions, and Jenkins without static tokens: JWT/OIDC and AppRole."
weight: 30
---

CI/CD systems should get secrets from Stronghold for the duration of a job and must not store long-lived tokens. Use:

- the [JWT/OIDC auth method](../../../user/auth/jwt/) if the CI system issues a signed OIDC token for each job (GitLab CI, GitHub Actions);
- the [AppRole auth method](../../../user/auth/approle/) if it does not (for example, Jenkins).

The tools mentioned on this page (`hashicorp/vault-action`, the HashiCorp Vault Plugin for Jenkins, the GitLab CI `secrets` keyword) are built for HashiCorp Vault. Stronghold is API-compatible with Vault, so they can be pointed at Stronghold. <!-- TODO(verify): tested compatibility of hashicorp/vault-action, Jenkins HashiCorp Vault Plugin, and GitLab CI secrets:vault with Stronghold -->

## General recommendations

- Do not store Stronghold tokens in CI/CD variables. Get a token at the start of the job via JWT or AppRole.
- Set a short token lifetime (`token_ttl`, `token_explicit_max_ttl`), no longer than the job duration.
- Create a separate role and policy per project and environment. Restrict roles by branch, tag, or environment using `bound_claims`.
- Grant only `read` on the required paths. Change policies and mounts in separate processes (see [Terraform and Ansible](../terraform-ansible/) and [GitOps](../gitops/)).
- Do not print secrets to the job log. Use the CI system's variable masking.
- For self-hosted runners, you can run [Stronghold Agent](../../../user/agent/use-cases/) on the runner host.

## GitLab CI

GitLab issues an OIDC token for a job, declared in the `id_tokens` keyword. Stronghold verifies the token signature using GitLab keys and matches the `project_path`, `ref`, and `ref_type` claims against the role.

1. Enable the JWT method at a dedicated path and set the GitLab address:

   ```bash
   d8 stronghold auth enable -path=gitlab jwt

   d8 stronghold write auth/gitlab/config \
     oidc_discovery_url="https://gitlab.example.com" \
     bound_issuer="https://gitlab.example.com"
   ```

1. Create a policy for the project:

   ```bash
   d8 stronghold policy write myproject-ci - <<'POLICY'
   path "secret/data/myproject/*" {
     capabilities = ["read"]
   }
   POLICY
   ```

1. Create a role that issues a token only to jobs from the `main` branch of the `mygroup/myproject` project:

   ```bash
   d8 stronghold write auth/gitlab/role/myproject-deploy - <<'ROLE'
   {
     "role_type": "jwt",
     "user_claim": "user_login",
     "bound_audiences": ["https://stronghold.example.com"],
     "bound_claims": {
       "project_path": "mygroup/myproject",
       "ref": "main",
       "ref_type": "branch"
     },
     "token_policies": ["myproject-ci"],
     "token_ttl": "10m",
     "token_explicit_max_ttl": "30m"
   }
   ROLE
   ```

1. In `.gitlab-ci.yml`, declare the token in `id_tokens` and exchange it for a Stronghold token. The job image must contain the `d8` utility:

   ```yaml
   deploy:
     stage: deploy
     id_tokens:
       STRONGHOLD_ID_TOKEN:
         aud: https://stronghold.example.com
     variables:
       STRONGHOLD_ADDR: https://stronghold.example.com
     script:
       - export STRONGHOLD_TOKEN="$(d8 stronghold write -field=token auth/gitlab/login role=myproject-deploy jwt=$STRONGHOLD_ID_TOKEN)"
       - export DB_PASSWORD="$(d8 stronghold kv get -mount=secret -field=password myproject/db)"
       - ./deploy.sh
     rules:
       - if: $CI_COMMIT_BRANCH == "main"
   ```

The `aud` value in `id_tokens` must match the role's `bound_audiences`.

## GitHub Actions

GitHub Actions issues an OIDC token with the `https://token.actions.githubusercontent.com` issuer. The workflow needs the `id-token: write` permission to get it.

1. Enable the JWT method and set the GitHub issuer:

   ```bash
   d8 stronghold auth enable -path=github jwt

   d8 stronghold write auth/github/config \
     oidc_discovery_url="https://token.actions.githubusercontent.com" \
     bound_issuer="https://token.actions.githubusercontent.com"
   ```

1. Create a role for the `myorg/myrepo` repository and the `main` branch:

   ```bash
   d8 stronghold write auth/github/role/myrepo-deploy - <<'ROLE'
   {
     "role_type": "jwt",
     "user_claim": "repository",
     "bound_audiences": ["https://stronghold.example.com"],
     "bound_claims": {
       "repository": "myorg/myrepo",
       "ref": "refs/heads/main"
     },
     "token_policies": ["myproject-ci"],
     "token_ttl": "10m"
   }
   ROLE
   ```

1. Use the `hashicorp/vault-action` action in the workflow:

   ```yaml
   name: deploy
   on:
     push:
       branches: [main]
   permissions:
     id-token: write
     contents: read
   jobs:
     deploy:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - uses: hashicorp/vault-action@v3
           with:
             url: https://stronghold.example.com
             method: jwt
             path: github
             role: myrepo-deploy
             jwtGithubAudience: https://stronghold.example.com
             secrets: |
               secret/data/myproject/db password | DB_PASSWORD
         - run: ./deploy.sh
   ```

For jobs that use `environment`, restrict the role by the `environment` claim instead of `ref`.

## Jenkins

Jenkins has no built-in per-job OIDC token, so use AppRole.

1. Enable AppRole and create a role with a short token lifetime and a `secret_id` limited by time and network:

   ```bash
   d8 stronghold auth enable approle

   d8 stronghold write auth/approle/role/jenkins-myproject \
     token_policies=myproject-ci \
     token_ttl=15m \
     token_max_ttl=30m \
     secret_id_ttl=24h \
     secret_id_num_uses=0 \
     secret_id_bound_cidrs=10.0.10.0/24
   ```

1. Get the `role_id` and `secret_id`:

   ```bash
   d8 stronghold read -field=role_id auth/approle/role/jenkins-myproject/role-id
   d8 stronghold write -f -field=secret_id auth/approle/role/jenkins-myproject/secret-id
   ```

1. Store them in Jenkins as a Vault App Role Credential of the HashiCorp Vault Plugin and use them in the Jenkinsfile:

   ```groovy
   def configuration = [
     vaultUrl: 'https://stronghold.example.com',
     vaultCredentialId: 'stronghold-approle',
     engineVersion: 2
   ]
   def secrets = [
     [path: 'secret/myproject/db', engineVersion: 2, secretValues: [
       [envVar: 'DB_PASSWORD', vaultKey: 'password']
     ]]
   ]

   pipeline {
     agent any
     stages {
       stage('Deploy') {
         steps {
           withVault([configuration: configuration, vaultSecrets: secrets]) {
             sh './deploy.sh'
           }
         }
       }
     }
   }
   ```

Rotate `secret_id` regularly (for example, with a separate job) and restrict it by network using `secret_id_bound_cidrs`. To deliver `secret_id`, you can use [response wrapping](../../../concepts/response-wrapping/).
