---
title: "Kubernetes"
linkTitle: "Kubernetes"
description: "Работа с механизмом секретов Kubernetes в Deckhouse Stronghold"
weight: 70
---

## Kubernetes

Механизм секретов `Kubernetes` в Deckhouse Stronghold генерирует токены для учётной записи сервиса Kubernetes. При необходимости он также может создавать сами объекты `ServiceAccount`, `Role` и `RoleBinding`. Созданные токены имеют настраиваемый TTL, а созданные объекты автоматически удаляются после окончания аренды Stronghold.

На каждую аренду Stronghold создаёт отдельный токен для конкретной учётной записи сервиса и возвращает его вызывающей стороне.

Используйте этот механизм, если нужно:

- выдавать приложениям временные токены доступа к API Kubernetes;
- ограничивать срок действия токенов;
- автоматизировать создание `ServiceAccount`, ролей и привязок ролей;
- управлять доступом через роли Stronghold и Kubernetes RBAC.

> Важно  
> Мы не рекомендуем использовать токены, созданные механизмом секретов `Kubernetes`, для аутентификации через [Kubernetes](../../auth/kubernetes/). Это приведёт к созданию множества уникальных идентификаторов в Stronghold, которыми будет сложно управлять.

## Как это работает

Базовый сценарий выглядит так:

1. Администратор настраивает права для учётной записи сервиса, под которой работает Stronghold.
2. Администратор включает механизм секретов `Kubernetes`.
3. Администратор создаёт роль Stronghold, которая определяет, какие токены и в каких неймспейсах можно выдавать.
4. Пользователь или приложение обращается в Stronghold за новыми учётными данными.
5. Stronghold создаёт токен Kubernetes и возвращает его клиенту.

Если включить автоматическое управление, Stronghold сможет не только выпустить токен, но и сам создать `ServiceAccount`, роль Kubernetes и `RoleBinding`.

## Что нужно подготовить

Перед настройкой убедитесь, что Stronghold может работать с API Kubernetes.

По умолчанию Stronghold подключается к Kubernetes через собственную учётную запись сервиса. Если вы используете Helm chart, эта учётная запись обычно создаётся автоматически и по умолчанию называется `stronghold`, но имя можно изменить через значение `server.serviceAccount.name`.

Этой учётной записи нужно выдать права:

- на создание токенов для `ServiceAccount`;
- при использовании автоматического управления — на создание и изменение `ServiceAccount`, ролей и привязок ролей;
- при использовании ограничений по лейблам неймспейсов — на чтение ресурсов `Namespace`.

> Важно  
> Защитите учётную запись сервиса Stronghold. Если вы выдали ей широкие права, по сути она получает административный доступ к кластеру.

## Настройка прав для Stronghold

### Минимальные права для создания токенов

Если Stronghold должен только выпускать токены для уже существующих `ServiceAccount`, достаточно таких прав:

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

### Расширенные права для автоматического управления

Если Stronghold должен создавать `ServiceAccount`, роли и `RoleBinding`, используйте более широкую роль:

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

### Права для работы с неймспейсами по лейблам

Если вы хотите ограничивать доступ по лейблам неймспейсов, добавьте право на чтение `Namespace`:

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

> Важно  
> Подобрать точный набор прав не всегда получается с первого раза. Kubernetes строго ограничивает эскалацию привилегий через RBAC.

### Привяжите роль к учётной записи сервиса Stronghold

После создания `ClusterRole` свяжите её с учётной записью сервиса Stronghold:

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

## Подготовка Kubernetes-объектов для базового сценария

Если Stronghold не будет автоматически создавать `ServiceAccount` и роли, подготовьте их заранее.

> Важно  
> Учётная запись сервиса, для которой Stronghold выпускает токены, не должна совпадать с учётной записью сервиса самого Stronghold.

Для примера используем неймспейс `test`:

```shell-session
$ d8 k create namespace test
namespace/test created
```

Создайте `ServiceAccount`, роль и привязку роли:

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

Примените манифест:

```shell-session
d8 k apply -f <file>.yaml
```

## Включение и базовая настройка механизма секретов

### Шаг 1. Включите механизм секретов Kubernetes

```shell-session
$ stronghold secrets enable kubernetes
Success! Enabled the kubernetes Secrets Engine at: kubernetes/
```

По умолчанию механизм секретов монтируется по пути `kubernetes/`. Если нужен другой путь, используйте аргумент `-path`.

### Шаг 2. Настройте точку монтирования

Допускается пустая конфигурация:

```shell-session
stronghold write -f kubernetes/config
```

### Шаг 3. Создайте роль Stronghold

Роль Stronghold определяет, где и для какой учётной записи сервиса можно выдавать токены.

Пример:

```shell-session
$ stronghold write kubernetes/roles/my-role \
   allowed_kubernetes_namespaces="*" \
   service_account_name="test-service-account-with-generated-token" \
   token_default_ttl="10m"
```

В этом примере:

- токены можно выдавать для любых неймспейсов;
- Stronghold будет использовать `ServiceAccount` `test-service-account-with-generated-token`;
- токен по умолчанию будет жить 10 минут.

## Получение токена Kubernetes

После того как пользователь или приложение прошли аутентификацию в Stronghold и получили нужные права, можно запросить токен Kubernetes.

Пример:

```shell-session
$ stronghold write kubernetes/creds/my-role \
    kubernetes_namespace=test

Key                        Value
–--                        -----
lease_id                   kubernetes/creds/my-role/31d771a6-...
lease_duration             10m0s
lease_renwable             false
service_account_name       test-service-account-with-generated-token
service_account_namespace  test
service_account_token      eyJHbGci0iJSUzI1NiIsImtpZCI6ImlrUEE...
```

Stronghold вернёт:

- `lease_id` — идентификатор аренды;
- `lease_duration` — срок действия токена;
- `service_account_name` — имя учётной записи сервиса;
- `service_account_namespace` — неймспейс;
- `service_account_token` — токен для обращения к API Kubernetes.

### Проверка токена

Используйте полученный токен в запросах к API Kubernetes:

```shell-session
$ curl -sk $(d8 k config view --minify -o 'jsonpath={.clusters[].cluster.server}')/api/v1/namespaces/test/pods \
    --header "Authorization: Bearer eyJHbGci0iJSUzI1Ni..."
{
  "kind": "PodList",
  "apiVersion": "v1",
  "metadata": {
    "resourceVersion": "1624"
  },
  "items": []
}
```

После окончания аренды токен будет отозван и запросы перестанут проходить:

```shell-session
$ curl -sk $(d8 k config view --minify -o 'jsonpath={.clusters[].cluster.server}')/api/v1/namespaces/test/pods \
    --header "Authorization: Bearer eyJHbGci0iJSUzI1Ni..."
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

## Настройка TTL токена

Токены `ServiceAccount` в Kubernetes имеют TTL. Когда срок действия заканчивается, токен автоматически отзывается.

### TTL по умолчанию и максимальный TTL

Укажите `token_default_ttl` и `token_max_ttl` в роли Stronghold:

```shell-session
$ stronghold write kubernetes/roles/my-role \
    allowed_kubernetes_namespaces="*" \
    service_account_name="new-service-account-with-generated-token" \
    token_default_ttl="10m" \
    token_max_ttl="2h"
```

### TTL при запросе токена

При выдаче токена можно задать TTL явно:

```shell-session
$ stronghold write kubernetes/creds/my-role \
    kubernetes_namespace=test \
    ttl=20m

Key                        Value
–--                        -----
lease_id                   kubernetes/creds/my-role/31d771a6-...
lease_duration             20m0s
lease_renwable             false
service_account_name       new-service-account-with-generated-token
service_account_namespace  test
service_account_token      eyJHbGci0iJSUzI1NiIsImtpZCI6ImlrUEE...
```

Если `ttl` не указан, будет использовано значение `token_default_ttl`. При этом TTL не может превышать `token_max_ttl`, если он задан.

### Как проверить срок действия JWT

Можно декодировать JWT и посмотреть поля `iat` и `exp`:

```shell-session
$ echo 'eyJhbGc...' | cut -d'.' -f2 | base64 -d  | jq -r '.iat,.exp|todate'
2022-05-20T17:14:50Z
2022-05-20T17:34:50Z
```

## Настройка аудитории токена

Токены Kubernetes имеют формат JWT и поддерживают claim `aud`. Этот claim определяет получателей, для которых предназначен токен.

### Аудитория по умолчанию

Укажите `token_default_audiences` в роли Stronghold:

```shell-session
$ stronghold write kubernetes/roles/my-role \
    allowed_kubernetes_namespaces="*" \
    service_account_name="new-service-account-with-generated-token" \
    token_default_audiences="custom-audience"
```

### Аудитория при запросе токена

Можно указать аудиторию явно при выдаче токена:

```shell-session
$ stronghold write kubernetes/creds/my-role \
    kubernetes_namespace=test \
    audiences="another-custom-audience"

Key                        Value
–--                        -----
lease_id                   kubernetes/creds/my-role/SriWQf0bPZ...
lease_duration             768h
lease_renwable             false
service_account_name       new-service-account-with-generated-token
service_account_namespace  test
service_account_token      eyJHbGci0iJSUzI1NiIsImtpZCI6ImlrUEE...
```

Если аудитория не указана, будет использовано значение `token_default_audiences`. Если и оно не задано, Kubernetes применит свои значения по умолчанию.

### Как проверить аудиторию токена

```shell-session
$ echo 'eyJhbGc...' | cut -d'.' -f2 | base64 -d
{"aud":["another-custom-audience"]...
```

## Автоматическое управление `ServiceAccount`, ролями и привязками ролей

Stronghold может автоматически создавать:

- `ServiceAccount`;
- `RoleBinding`;
- при необходимости — саму роль Kubernetes.

### Автоматическое создание `ServiceAccount` и `RoleBinding`

Если роль Kubernetes уже существует, укажите её имя в параметре `kubernetes_role_name`:

```shell-session
$ stronghold write kubernetes/roles/auto-managed-sa-role \
    allowed_kubernetes_namespaces="test" \
    kubernetes_role_name="test-role-list-pods"
```

> Важно  
> Учётной записи сервиса Stronghold также потребуется доступ к тем ресурсам, к которым она выдаёт доступ. Это ограничение связано с защитой Kubernetes от эскалации привилегий.

Для примера из этой страницы такую привязку можно создать так:

```shell-session
d8 k -n test create rolebinding --role test-role-list-pods --serviceaccount=stronghold:stronghold stronghold stronghold-test-role-abilities
```

После этого можно запросить токен:

```shell-session
$ stronghold write kubernetes/creds/auto-managed-sa-role \
    kubernetes_namespace=test
Key                          Value
---                          -----
lease_id                     kubernetes/creds/auto-managed-sa-role/cujRLYjKZUMQk6dkHBGGWm67
lease_duration               768h
lease_renewable              false
service_account_name         v-token-auto-man-1653001548-5z6hrgsxnmzncxejztml4arz
service_account_namespace    test
service_account_token        eyJHbGci0iJSUzI1Ni...
```

### Автоматическое создание роли Kubernetes

Если Stronghold должен создать и роль Kubernetes, передайте `generated_role_rules`:

```shell-session
$ stronghold write kubernetes/roles/auto-managed-sa-and-role \
    allowed_kubernetes_namespaces="test" \
    generated_role_rules='{"rules":[{"apiGroups":[""],"resources":["pods"],"verbs":["list"]}]}'
```

После этого токен можно получить так же:

```shell-session
$ stronghold write kubernetes/creds/auto-managed-sa-and-role \
    kubernetes_namespace=test
Key                          Value
---                          -----
lease_id                     kubernetes/creds/auto-managed-sa-and-role/pehLtegoTP8vCkcaQozUqOHf
lease_duration               768h
lease_renewable              false
service_account_name         v-token-auto-man-1653002096-4imxf3ytjh5hbyro9s1oqdo3
service_account_namespace    test
service_account_token        eyJHbGci0iJSUzI1Ni...
```

## Практические рекомендации

Чтобы механизм секретов `Kubernetes` было проще поддерживать:

- используйте отдельную учётную запись сервиса для самого Stronghold;
- не выпускайте токены для той же `ServiceAccount`, под которой работает Stronghold;
- задавайте короткий TTL для токенов, если они нужны только для краткоживущих задач;
- ограничивайте список неймспейсов в роли Stronghold, если не нужен доступ ко всему кластеру;
- включайте автоматическое управление только там, где это действительно нужно;
- проверяйте права Stronghold в Kubernetes RBAC заранее, особенно при автоматическом создании ролей и `RoleBinding`.

## Что дальше

- Если вам нужны временные учётные данные для баз данных, используйте раздел [Базы данных](../database/).
- Если вам нужно хранить произвольные секреты, используйте [KV](../kv/).
