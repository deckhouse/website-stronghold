---
title: "CI/CD"
description: "Получение секретов Stronghold в GitLab CI, GitHub Actions и Jenkins без статических токенов: JWT/OIDC и AppRole."
weight: 30
---

CI/CD-системы должны получать секреты из Stronghold на время выполнения задания и не хранить долгоживущие токены. Для этого используйте:

- [метод JWT/OIDC](../../../user/auth/jwt/) — если CI-система выпускает подписанный OIDC-токен для каждого задания (GitLab CI, GitHub Actions);
- [метод AppRole](../../../user/auth/approle/) — если такого токена нет (например, Jenkins).

Инструменты, упомянутые на странице (`hashicorp/vault-action`, HashiCorp Vault Plugin для Jenkins, ключевое слово `secrets` в GitLab CI), разработаны для HashiCorp Vault. Stronghold совместим с API Vault, поэтому их можно направить на Stronghold. <!-- TODO(verify): протестированная совместимость hashicorp/vault-action, Jenkins HashiCorp Vault Plugin и GitLab CI secrets:vault со Stronghold -->

## Общие рекомендации

- Не храните токены Stronghold в переменных CI/CD. Получайте токен в начале задания через JWT или AppRole.
- Задавайте короткий срок действия токена (`token_ttl`, `token_explicit_max_ttl`) — не больше длительности задания.
- Создавайте отдельную роль и политику на каждый проект и окружение. Ограничивайте роли по ветке, тегу или окружению через `bound_claims`.
- Выдавайте только права `read` на нужные пути. Изменение политик и точек монтирования выполняйте отдельными процессами (см. [«Terraform и Ansible»](../terraform-ansible/) и [«GitOps»](../gitops/)).
- Не выводите секреты в лог задания. Используйте маскирование переменных средствами CI-системы.
- Для self-hosted раннеров можно использовать [Stronghold Agent](../../../user/agent/use-cases/) на узле раннера.

## GitLab CI

GitLab выпускает для задания OIDC-токен, объявленный в ключевом слове `id_tokens`. Stronghold проверяет подпись токена по ключам GitLab и сопоставляет утверждения (claims) `project_path`, `ref`, `ref_type` с ролью.

1. Включите метод JWT по отдельному пути и укажите адрес GitLab:

   ```bash
   d8 stronghold auth enable -path=gitlab jwt

   d8 stronghold write auth/gitlab/config \
     oidc_discovery_url="https://gitlab.example.com" \
     bound_issuer="https://gitlab.example.com"
   ```

1. Создайте политику для проекта:

   ```bash
   d8 stronghold policy write myproject-ci - <<'POLICY'
   path "secret/data/myproject/*" {
     capabilities = ["read"]
   }
   POLICY
   ```

1. Создайте роль, которая выдаёт токен только заданиям из ветки `main` проекта `mygroup/myproject`:

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

1. В `.gitlab-ci.yml` объявите токен в `id_tokens` и обменяйте его на токен Stronghold. В образе задания должна быть утилита `d8`:

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

Значение `aud` в `id_tokens` должно совпадать с `bound_audiences` роли.

## GitHub Actions

GitHub Actions выпускает OIDC-токен с издателем `https://token.actions.githubusercontent.com`. Для получения токена в workflow нужно разрешение `id-token: write`.

1. Включите метод JWT и укажите издателя GitHub:

   ```bash
   d8 stronghold auth enable -path=github jwt

   d8 stronghold write auth/github/config \
     oidc_discovery_url="https://token.actions.githubusercontent.com" \
     bound_issuer="https://token.actions.githubusercontent.com"
   ```

1. Создайте роль для репозитория `myorg/myrepo` и ветки `main`:

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

1. Используйте в workflow действие `hashicorp/vault-action`:

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

Для задач с `environment` ограничивайте роль утверждением `environment` вместо `ref`.

## Jenkins

У Jenkins нет встроенного OIDC-токена для каждого задания, поэтому используйте AppRole.

1. Включите AppRole и создайте роль с коротким сроком жизни токена и ограниченным по времени и сети `secret_id`:

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

1. Получите `role_id` и `secret_id`:

   ```bash
   d8 stronghold read -field=role_id auth/approle/role/jenkins-myproject/role-id
   d8 stronghold write -f -field=secret_id auth/approle/role/jenkins-myproject/secret-id
   ```

1. Сохраните их в Jenkins как учётные данные типа Vault App Role Credential плагина HashiCorp Vault Plugin и используйте в Jenkinsfile:

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

Регулярно перевыпускайте `secret_id` (например, отдельным заданием) и ограничивайте его по сети через `secret_id_bound_cidrs`. Для доставки `secret_id` можно использовать [обертывание ответа](../../../concepts/response-wrapping/).
