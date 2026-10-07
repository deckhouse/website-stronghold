---
title: "Временный доступ к Kubernetes через токены ServiceAccount"
linkTitle: "Временные токены Kubernetes"
description: "Выдача CI-заданиям и инженерам короткоживущих токенов ServiceAccount через механизм секретов kubernetes: роли с разрешёнными пространствами имён, генерируемые правила RBAC, получение токена и проверка через d8 k auth can-i."
weight: 50
params:
  relatedLinks:
    - title: "Механизм секретов Kubernetes"
      url: ../../../user/secrets-engines/kubernetes/
    - title: "Интеграция с CI/CD"
      url: ../../delivery/ci-cd/
    - title: "Аренда, продление и отзыв"
      url: ../../../concepts/lease/
    - title: "API механизмов секретов"
      url: ../../../reference/api/secrets/
---

Вместо постоянного kubeconfig с широкими правами CI-задание или инженер запрашивает у Stronghold токен ServiceAccount на время работы. Stronghold создаёт в кластере учётную запись сервиса, роль и привязку роли с нужными правами, выпускает токен с ограниченным сроком жизни и удаляет созданные объекты по истечении аренды.

![Схема выдачи токенов Kubernetes](../../../images/ex-kubernetes-tokens.png)

## Цель

Настроить две роли механизма секретов Kubernetes:

- `ci-deploy` — для CI-заданий: права на управление Deployment, Service и ConfigMap в пространстве имён `myapp`, токен живёт 15 минут;
- `engineer-view` — для инженеров: права встроенной ClusterRole `view` в пространствах имён `myapp` и `myapp-stage`, токен живёт 1 час, не дольше 8 часов.

## Предварительные требования

- Stronghold, развёрнутый в DP, и токен с правами на настройку механизмов секретов и политик.
- Доступ к кластеру через `d8 k` с правами на создание ClusterRole и ClusterRoleBinding.
- Пространства имён `myapp` и `myapp-stage`.
- Настроенный метод аутентификации для CI (например, JWT, как в разделе [«Интеграция с CI/CD»](../../delivery/ci-cd/)) и для инженеров (например, OIDC).

## Шаг 1. Выдайте Stronghold права в кластере

Stronghold создаёт объекты RBAC от имени своей учётной записи сервиса. Из-за защиты Kubernetes от повышения привилегий этой учётной записи нужны права `bind` и `escalate` на роли.

1. Определите имя учётной записи сервиса Stronghold:

   ```bash
   d8 k -n d8-stronghold get pod stronghold-0 -o jsonpath='{.spec.serviceAccountName}'
   ```

   В DP это учётная запись сервиса `stronghold` в пространстве имён `d8-stronghold`. По умолчанию модуль выдаёт ей только `system:auth-delegator` и доступ к собственному пространству имён (Secrets и Pods), поэтому права ниже нужно добавить.

1. Создайте ClusterRole и привяжите её к учётной записи сервиса Stronghold. Замените `stronghold` в `subjects` на имя, полученное на предыдущем шаге:

   ```yaml
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRole
   metadata:
     name: stronghold-k8s-secrets-engine
   rules:
     - apiGroups: [""]
       resources: ["serviceaccounts", "serviceaccounts/token"]
       verbs: ["create", "update", "delete"]
     - apiGroups: ["rbac.authorization.k8s.io"]
       resources: ["rolebindings", "clusterrolebindings"]
       verbs: ["create", "update", "delete"]
     - apiGroups: ["rbac.authorization.k8s.io"]
       resources: ["roles", "clusterroles"]
       verbs: ["bind", "escalate", "create", "update", "delete"]
   ---
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRoleBinding
   metadata:
     name: stronghold-k8s-secrets-engine
   roleRef:
     apiGroup: rbac.authorization.k8s.io
     kind: ClusterRole
     name: stronghold-k8s-secrets-engine
   subjects:
     - kind: ServiceAccount
       name: stronghold
       namespace: d8-stronghold
   ```

   ```bash
   d8 k apply -f stronghold-k8s-secrets-engine.yaml
   ```

{{< alert level="warning" >}}
Учётная запись сервиса Stronghold с такими правами фактически является администратором кластера. Ограничьте доступ к настройке механизма секретов Kubernetes политиками Stronghold.
{{< /alert >}}

## Шаг 2. Включите механизм секретов

```bash
d8 stronghold secrets enable kubernetes
d8 stronghold write -f kubernetes/config
```

Пустая конфигурация означает, что Stronghold подключается к API текущего кластера с сертификатом CA и JWT своего пода. Для другого кластера задайте параметры `kubernetes_host`, `kubernetes_ca_cert` и `service_account_jwt`.

## Шаг 3. Создайте роль для CI

Роль с параметром `generated_role_rules` создаёт на каждый запрос всю цепочку объектов: ServiceAccount, Role и RoleBinding.

```bash
d8 stronghold write kubernetes/roles/ci-deploy \
  allowed_kubernetes_namespaces="myapp" \
  token_default_ttl="15m" \
  token_max_ttl="1h" \
  generated_role_rules='{"rules":[
    {"apiGroups":["apps"],"resources":["deployments"],"verbs":["get","list","watch","create","update","patch"]},
    {"apiGroups":[""],"resources":["services","configmaps"],"verbs":["get","list","create","update","patch"]}
  ]}'
```

- `allowed_kubernetes_namespaces` — пространства имён, в которых можно запросить токен. Вместо списка можно задать селектор меток `allowed_kubernetes_namespace_selector`; для этого Stronghold нужно право `get` на `namespaces`;
- `generated_role_rules` — правила Role в формате JSON или YAML;
- `token_default_ttl` и `token_max_ttl` — срок жизни токена по умолчанию и максимальный.

## Шаг 4. Создайте роль для инженеров

Роль с параметром `kubernetes_role_name` привязывает создаваемую учётную запись сервиса к существующей роли, здесь — к встроенной ClusterRole `view`:

```bash
d8 stronghold write kubernetes/roles/engineer-view \
  allowed_kubernetes_namespaces="myapp,myapp-stage" \
  kubernetes_role_type="ClusterRole" \
  kubernetes_role_name="view" \
  token_default_ttl="1h" \
  token_max_ttl="8h"
```

При `kubernetes_role_type="ClusterRole"` Stronghold по умолчанию создаёт RoleBinding в запрошенном пространстве имён. Чтобы выдать права на весь кластер, при запросе учётных данных передайте `cluster_role_binding=true` — тогда будет создана ClusterRoleBinding. Разрешайте это только для отдельных ролей и политик.

Если нужно выпускать токены для уже существующей учётной записи сервиса, используйте параметр `service_account_name`. Он несовместим с `kubernetes_role_name` и `generated_role_rules`.

## Шаг 5. Создайте политики

Эндпоинт `creds` вызывается методом `POST`, поэтому в политике нужна возможность `update`:

```bash
d8 stronghold policy write k8s-ci-deploy - <<'POLICY'
path "kubernetes/creds/ci-deploy" {
  capabilities = ["update"]
}
POLICY

d8 stronghold policy write k8s-engineer-view - <<'POLICY'
path "kubernetes/creds/engineer-view" {
  capabilities = ["update"]
}
POLICY
```

Назначьте политику `k8s-ci-deploy` роли метода аутентификации CI, а `k8s-engineer-view` — группе инженеров в методе OIDC или в идентификационных данных (Identity).

## Шаг 6. Получите токен

1. Запросите токен для CI:

   ```bash
   d8 stronghold write kubernetes/creds/ci-deploy kubernetes_namespace=myapp
   ```

   Пример вывода:

   ```text
   Key                          Value
   ---                          -----
   lease_id                     kubernetes/creds/ci-deploy/cujRLYjKZUMQk6dkHBGGWm67
   lease_duration               15m
   lease_renewable              false
   service_account_name         v-token-ci-deplo-1653001548-5z6hrgsxnmzncxejztml4arz
   service_account_namespace    myapp
   service_account_token        eyJHbGci0iJSUzI1Ni...
   ```

   Аренда не продлевается: после её истечения запросите новый токен.

1. Инженер запрашивает токен на нужное время в пределах `token_max_ttl`:

   ```bash
   d8 stronghold write kubernetes/creds/engineer-view \
     kubernetes_namespace=myapp-stage \
     ttl=2h
   ```

1. Создайте временный kubeconfig с полученным токеном:

   ```bash
   TOKEN="$(d8 stronghold write -field=service_account_token kubernetes/creds/engineer-view kubernetes_namespace=myapp-stage)"
   SERVER="$(d8 k config view --minify -o jsonpath='{.clusters[].cluster.server}')"
   d8 k config view --minify --raw -o jsonpath='{.clusters[].cluster.certificate-authority-data}' | base64 -d > ca.crt

   export KUBECONFIG="$PWD/tmp-kubeconfig"
   d8 k config set-cluster dkp --server="$SERVER" --certificate-authority=ca.crt --embed-certs=true
   d8 k config set-credentials stronghold-token --token="$TOKEN"
   d8 k config set-context tmp --cluster=dkp --user=stronghold-token --namespace=myapp-stage
   d8 k config use-context tmp
   ```

В CI-задании получите токен Stronghold через метод JWT, как описано в разделе [«Интеграция с CI/CD»](../../delivery/ci-cd/), затем запросите токен Kubernetes и передайте его в `d8 k` параметром `--token`:

```yaml
deploy:
  script:
    - export STRONGHOLD_TOKEN="$(d8 stronghold write -field=token auth/gitlab/login role=myproject-deploy jwt=$STRONGHOLD_ID_TOKEN)"
    - export K8S_TOKEN="$(d8 stronghold write -field=service_account_token kubernetes/creds/ci-deploy kubernetes_namespace=myapp)"
    - d8 k --server="$K8S_SERVER" --certificate-authority="$K8S_CA_FILE" --token="$K8S_TOKEN" -n myapp apply -f deploy/
```

## Проверка

1. Запросите токен CI и проверьте разрешённые действия:

   ```bash
   TOKEN="$(d8 stronghold write -field=service_account_token kubernetes/creds/ci-deploy kubernetes_namespace=myapp)"
   d8 k --token="$TOKEN" auth can-i create deployments -n myapp
   d8 k --token="$TOKEN" auth can-i patch configmaps -n myapp
   ```

   Обе команды должны вернуть `yes`.

1. Проверьте, что лишних прав нет:

   ```bash
   d8 k --token="$TOKEN" auth can-i get secrets -n myapp
   d8 k --token="$TOKEN" auth can-i create deployments -n kube-system
   ```

   Обе команды должны вернуть `no`.

   Если в текущем kubeconfig задан другой способ аутентификации (например, клиентский сертификат), проверку удобнее выполнить через временный kubeconfig из шага 6.

1. Убедитесь, что запрос токена для неразрешённого пространства имён отклоняется:

   ```bash
   d8 stronghold write kubernetes/creds/ci-deploy kubernetes_namespace=default
   ```

1. Просмотрите объекты, созданные Stronghold в пространстве имён:

   ```bash
   d8 k -n myapp get serviceaccounts,roles,rolebindings | grep v-token
   ```

1. Отзовите аренду досрочно и убедитесь, что токен перестал работать:

   ```bash
   d8 stronghold lease revoke -prefix kubernetes/creds/ci-deploy/
   d8 k --token="$TOKEN" auth can-i create deployments -n myapp
   ```

   После удаления учётной записи сервиса запрос с её токеном завершается ошибкой `Unauthorized`.

Для ролей с `service_account_name` Stronghold не создаёт учётную запись сервиса, поэтому отзыв аренды не делает токен недействительным до истечения его срока жизни. Для таких ролей задавайте короткий `token_default_ttl`.

## Очистка

1. Отзовите все выданные токены:

   ```bash
   d8 stronghold lease revoke -prefix kubernetes/creds/
   ```

1. Удалите роли, политики и конфигурацию:

   ```bash
   d8 stronghold delete kubernetes/roles/ci-deploy
   d8 stronghold delete kubernetes/roles/engineer-view
   d8 stronghold policy delete k8s-ci-deploy
   d8 stronghold policy delete k8s-engineer-view
   d8 stronghold secrets disable kubernetes
   ```

1. Удалите права Stronghold в кластере и временные файлы:

   ```bash
   d8 k delete -f stronghold-k8s-secrets-engine.yaml
   rm -f tmp-kubeconfig ca.crt
   ```
