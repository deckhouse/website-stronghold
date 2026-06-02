---
title: "Механизм секретов Kubernetes"
description: "Сведения о разделе \"Механизм секретов Kubernetes\" в Deckhouse Stronghold."
weight: 70
---

Kubernetes Secrets Engine для Stronghold генерирует токены для ServiceAccount Kubernetes.
При необходимости он также создаёт объекты ServiceAccount, Role и RoleBinding.
Созданные токены имеют настраиваемое время жизни (TTL), а все созданные объекты автоматически удаляются после истечения срока [аренды](../../../concepts/lease/) Stronghold.

Для каждой аренды Stronghold создаёт токен для конкретной ServiceAccount и возвращает его вызывающей стороне.
Дополнительную информацию см. в документации [Kubernetes service accounts](https://kubernetes.io/docs/concepts/security/service-accounts/) и [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

{{< alert level="warning" >}}
Не используйте токены, созданные механизмом секретов Kubernetes, для аутентификации через [Kubernetes Auth Method](../../auth/kubernetes/).
Это приведёт к созданию множества уникальных идентификаторов в Stronghold, которыми будет сложно управлять.
{{< /alert >}}

## Настройка

Перед использованием механизма секретов Kubernetes предварительно настройте его.
Обычно эти действия выполняет администратор Stronghold или система автоматического управления конфигурацией.

По умолчанию Stronghold подключается к Kubernetes с помощью собственной ServiceAccount.
При использовании [Helm chart](https://github.com/hashicorp/vault-helm) эта ServiceAccount создаётся автоматически и получает имя Helm-релиза.
По умолчанию используется имя `stronghold`, но его можно изменить через параметр `server.serviceAccount.name`.

Убедитесь, что ServiceAccount, которую использует Stronghold, имеет права:
- на управление токенами ServiceAccount;
- на управление ServiceAccount, Role и RoleBinding, если используется автоматическое создание этих объектов.

Этими правами можно управлять с помощью ролей Kubernetes.
Роль привязывается к ServiceAccount Stronghold через ClusterRoleBinding или RoleBinding.

Ниже приведён пример ClusterRole только для создания токенов ServiceAccount:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8s-minimal-secrets-abilities
rules:
  - apiGroups: [""]
    resources: ["serviceaccounts/token"]
    verbs: ["create"]
```

Ниже приведён пример ClusterRole с расширенными правами.
Она позволяет управлять токенами, ServiceAccount, RoleBinding и Role:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8s-full-secrets-abilities
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
```

Создайте эту роль в Kubernetes.
Например, с помощью команды `d8 k apply -f`.

Если вы хотите ограничивать выбор неймспейсов по лейблам, предоставьте Stronghold разрешение на чтение неймспейсов:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8s-full-secrets-abilities-with-labels
rules:
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get"]
  - apiGroups: [""]
    resources: ["serviceaccounts", "serviceaccounts/token"]
    verbs: ["create", "update", "delete"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["rolebindings", "clusterrolebindings"]
    verbs: ["create", "update", "delete"]
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources: ["roles", "clusterroles"]
    verbs: ["bind", "escalate", "create", "update", "delete"]
```

{{< alert level="warning" >}}
Подбор корректных разрешений для Stronghold может потребовать нескольких попыток.
Kubernetes строго предотвращает повышение привилегий.
Дополнительную информацию см. в документации [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#privilege-escalation-prevention-and-bootstrapping).
{{< /alert >}}

{{< alert level="warning" >}}
Защитите ServiceAccount Stronghold, особенно если она имеет широкие права.
По сути, такая учётная запись получает административный доступ к кластеру.
{{< /alert >}}

Создайте привязку роли, чтобы связать её с ServiceAccount Stronghold и выдать разрешение на управление токенами:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: stronghold-token-creator-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: k8s-minimal-secrets-abilities
subjects:
  - kind: ServiceAccount
    name: stronghold
    namespace: stronghold
```

Дополнительную информацию о ролях Kubernetes, ServiceAccount, привязках и токенах см. в разделе [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

Если Stronghold не будет автоматически управлять ролями или ServiceAccount, см. раздел [Автоматическое управление ролями и учётными записями сервиса](#roles-and-sa).
В этом случае заранее настройте ServiceAccount, для которой Stronghold будет выпускать токены.

{{< alert level="warning" >}}
Рекомендуется использовать разные ServiceAccount для Stronghold и для выпуска токенов.
Не используйте одну и ту же ServiceAccount в обоих случаях.
{{< /alert >}}

В примерах ниже используется неймспейс `test`.
Создайте его, если он ещё не существует:

```shell
d8 k create namespace test
```

Ниже приведён пример простой настройки ServiceAccount, Role и RoleBinding в неймспейсе `test` с базовыми разрешениями:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-service-account-with-generated-token
  namespace: test
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: test-role-list-pods
  namespace: test
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: test-role-abilities
  namespace: test
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: test-role-list-pods
subjects:
  - kind: ServiceAccount
    name: test-service-account-with-generated-token
    namespace: test
```

Создайте эти объекты с помощью команды `d8 k apply -f`.

Включите механизм секретов Kubernetes:

```shell
stronghold secrets enable kubernetes
```

По умолчанию движок секретов монтируется по пути `kubernetes/`.
Это поведение можно изменить с помощью аргумента `-path` при включении.

Настройте точку монтирования.
Допускается пустая конфигурация:

```shell
stronghold write -f kubernetes/config
```

Теперь настройте роль Stronghold в механизме секретов Kubernetes.
Это не роль Kubernetes.
Она будет использоваться для генерации токенов Kubernetes для созданной ServiceAccount:

```shell
stronghold write kubernetes/roles/my-role \
  allowed_kubernetes_namespaces="*" \
  service_account_name="test-service-account-with-generated-token" \
  token_default_ttl="10m"
```

## Создание учётных данных

После того как пользователь пройдёт аутентификацию в Stronghold и получит необходимые права, запись в эндпоинт `creds` для роли Stronghold сгенерирует и вернёт новый токен ServiceAccount:

```shell
stronghold write kubernetes/creds/my-role \
  kubernetes_namespace=test
```

Пример вывода:

```console
Key                        Value
---                        -----
lease_id                   kubernetes/creds/my-role/31d771a6-...
lease_duration             10m0s
lease_renewable            false
service_account_name       test-service-account-with-generated-token
service_account_namespace  test
service_account_token      eyJHbGci0iJSUzI1NiIsImtpZCI6ImlrUEE...
```

Используйте полученный токен ServiceAccount для авторизованных запросов к API Kubernetes.
Права доступа при этом определяются привязками ролей к ServiceAccount:

```shell
curl -sk "$(d8 k config view --minify -o 'jsonpath={.clusters[].cluster.server}')/api/v1/namespaces/test/pods" \
  --header "Authorization: Bearer eyJHbGci0iJSUzI1Ni..."
```

Пример ответа:

```console
{
  "kind": "PodList",
  "apiVersion": "v1",
  "metadata": {
    "resourceVersion": "1624"
  },
  "items": []
}
```

После истечения срока [аренды](../../../concepts/lease/) проверьте, что токен был отозван и больше не может использоваться для запросов к API Kubernetes:

```shell
curl -sk "$(d8 k config view --minify -o 'jsonpath={.clusters[].cluster.server}')/api/v1/namespaces/test/pods" \
  --header "Authorization: Bearer eyJHbGci0iJSUzI1Ni..."
```

Пример ответа:

```console
{
  "kind": "Status",
  "apiVersion": "v1",
  "metadata": {},
  "status": "Failure",
  "message": "Unauthorized",
  "reason": "Unauthorized",
  "code": 401
}
```

## Время жизни токена

Токены ServiceAccount Kubernetes имеют время жизни.
Когда срок действия токена истекает, токен автоматически отзывается.

При создании или настройке роли Stronghold можно указать стандартное время жизни `token_default_ttl` и максимальное время жизни `token_max_ttl`:

```shell
stronghold write kubernetes/roles/my-role \
  allowed_kubernetes_namespaces="*" \
  service_account_name="new-service-account-with-generated-token" \
  token_default_ttl="10m" \
  token_max_ttl="2h"
```

Также можно задать время жизни `ttl` при генерации токена через эндпоинт `creds`.
Если значение `ttl` не указано, используется значение по умолчанию.
Оно не может превышать `token_max_ttl`, если этот параметр задан:

```shell
stronghold write kubernetes/creds/my-role \
  kubernetes_namespace=test \
  ttl=20m
```

Пример вывода:

```console
Key                        Value
---                        -----
lease_id                   kubernetes/creds/my-role/31d771a6-...
lease_duration             20m0s
lease_renewable            false
service_account_name       new-service-account-with-generated-token
service_account_namespace  test
service_account_token      eyJHbGci0iJSUzI1NiIsImtpZCI6ImlrUEE...
```

Проверить время жизни JWT-токена можно, декодировав его и преобразовав поля `iat` и `exp` из формата timestamp в читаемый вид:

```shell
echo 'eyJhbGc...' | cut -d'.' -f2 | base64 -d | jq -r '.iat,.exp|todate'
```

Пример вывода:

```console
2022-05-20T17:14:50Z
2022-05-20T17:34:50Z
```

## API audience

Токены Kubernetes имеют формат JWT и содержат набор утверждений.
Одно из них — `aud`.
Это строка или массив строк, которые определяют получателей токена.
Дополнительную информацию см. в спецификации [JWT audience claim](https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.3).

При создании или настройке роли Stronghold можно задать аудитории по умолчанию через параметр `token_default_audiences`.
Если параметр не указан, кластер Kubernetes использует собственные значения аудитории для токенов ServiceAccount:

```shell
stronghold write kubernetes/roles/my-role \
  allowed_kubernetes_namespaces="*" \
  service_account_name="new-service-account-with-generated-token" \
  token_default_audiences="custom-audience"
```

Также можно указать аудитории через параметр `audiences` при генерации токена через эндпоинт `creds`.
Если значение не задано, используется `token_default_audiences`:

```shell
stronghold write kubernetes/creds/my-role \
  kubernetes_namespace=test \
  audiences="another-custom-audience"
```

Пример вывода:

```console
Key                        Value
---                        -----
lease_id                   kubernetes/creds/my-role/SriWQf0bPZ...
lease_duration             768h
lease_renewable            false
service_account_name       new-service-account-with-generated-token
service_account_namespace  test
service_account_token      eyJHbGci0iJSUzI1NiIsImtpZCI6ImlrUEE...
```

Проверить аудиторию токена можно, расшифровав JWT:

```shell
echo 'eyJhbGc...' | cut -d'.' -f2 | base64 -d
```

Пример вывода:

```console
{"aud":["another-custom-audience"]...
```

## Автоматическое управление ролями и учётными записями сервиса {#roles-and-sa}

При настройке роли Stronghold можно передать параметры, которые включают автоматическое создание ServiceAccount и RoleBinding.
При необходимости Stronghold может автоматически создавать и саму роль Kubernetes.

Если нужно использовать существующую роль Kubernetes, но при этом автоматически создавать ServiceAccount и RoleBinding, задайте параметр `kubernetes_role_name`:

```shell
stronghold write kubernetes/roles/auto-managed-sa-role \
  allowed_kubernetes_namespaces="test" \
  kubernetes_role_name="test-role-list-pods"
```

{{< alert level="warning" >}}
ServiceAccount Stronghold также потребуется доступ к ресурсам, доступ к которым она выдаёт.
Для примера выше можно выполнить команду `d8 k -n test create rolebinding --role test-role-list-pods --serviceaccount=stronghold:stronghold stronghold stronghold-test-role-abilities`.
Так Kubernetes предотвращает эскалацию привилегий.
Дополнительную информацию см. в документации [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#privilege-escalation-prevention-and-bootstrapping).
{{< /alert >}}

После этого можно получить учётные данные с помощью автоматически созданной ServiceAccount:

```shell
stronghold write kubernetes/creds/auto-managed-sa-role \
  kubernetes_namespace=test
```

Пример вывода:

```console
Key                          Value
---                          -----
lease_id                     kubernetes/creds/auto-managed-sa-role/cujRLYjKZUMQk6dkHBGGWm67
lease_duration               768h
lease_renewable              false
service_account_name         v-token-auto-man-1653001548-5z6hrgsxnmzncxejztml4arz
service_account_namespace    test
service_account_token        eyJHbGci0iJSUzI1Ni...
```

Stronghold также может автоматически создать роль вместе с ServiceAccount и RoleBinding.
Для этого укажите параметр `generated_role_rules`, в который передаётся набор правил JSON или YAML для создаваемой роли:

```shell
stronghold write kubernetes/roles/auto-managed-sa-and-role \
  allowed_kubernetes_namespaces="test" \
  generated_role_rules='{"rules":[{"apiGroups":[""],"resources":["pods"],"verbs":["list"]}]}'
```

После этого получите учётные данные тем же способом:

```shell
stronghold write kubernetes/creds/auto-managed-sa-and-role \
  kubernetes_namespace=test
```

Пример вывода:

```console
Key                          Value
---                          -----
lease_id                     kubernetes/creds/auto-managed-sa-and-role/pehLtegoTP8vCkcaQozUqOHf
lease_duration               768h
lease_renewable              false
service_account_name         v-token-auto-man-1653002096-4imxf3ytjh5hbyro9s1oqdo3
service_account_namespace    test
service_account_token        eyJHbGci0iJSUzI1Ni...
```
