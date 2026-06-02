---
title: "Kubernetes"
linkTitle: "Kubernetes"
weight: 80
description: "Аутентификация в Deckhouse Stronghold с помощью токена ServiceAccount Kubernetes."
---

Метод аутентификации `kubernetes auth` позволяет аутентифицироваться в Deckhouse Stronghold с помощью токена `ServiceAccount` Kubernetes. Этот метод упрощает использование Deckhouse Stronghold в подах Kubernetes.

Токен `ServiceAccount` Kubernetes также можно использовать для входа через `JWT auth`. Подробнее см. в разделе [Использование JWT auth вместо Kubernetes auth](#использование-jwt-auth-вместо-kubernetes-auth).

Этот метод подходит для сценариев, в которых приложение работает внутри Kubernetes и должно получать токены Deckhouse Stronghold на основе identity Kubernetes. Концептуально `Kubernetes auth` использует `TokenReview API`, а `JWT auth` проверяет токен криптографически по открытому ключу.

## Аутентификация

### Через CLI

По умолчанию используется путь `/kubernetes_local`. Если метод аутентификации включён по другому пути, укажите его через параметр `-path`.

```shell
d8 stronghold write auth/kubernetes/login role=demo jwt=...
```

### Через API

По умолчанию используется эндпоинт `auth/kubernetes_local/login`. Если метод аутентификации включён по другому пути, используйте его вместо `kubernetes_local`.

```shell
curl \
  --request POST \
  --data '{"jwt": "<your service account jwt>", "role": "demo"}' \
  https://stronghold.example.com/v1/auth/kubernetes/login
```

Ответ содержит токен в поле `auth.client_token`:

```json
{
  "auth": {
    "client_token": "38fe9691-e623-7238-f618-c94d4e7bc674",
    "accessor": "78e87a38-84ed-2692-538f-ca8b9f400ab3",
    "policies": ["default"],
    "metadata": {
      "role": "demo",
      "service_account_name": "myapp",
      "service_account_namespace": "default",
      "service_account_secret_name": "myapp-token-pd21c",
      "service_account_uid": "aa9aa8ff-98d0-11e7-9bb7-0800276d99bf"
    },
    "lease_duration": 2764800,
    "renewable": true
  }
}
```

## Конфигурация

Метод аутентификации нужно настроить заранее, прежде чем пользователи или машины смогут пройти аутентификацию.

Обычно эти действия выполняет оператор или инструмент управления конфигурацией.

В Deckhouse Stronghold по умолчанию включён метод Kubernetes по пути `kubernetes_local`. Он позволяет аутентифицировать приложения, запущенные в том же кластере, что и Deckhouse Stronghold.

При необходимости можно добавить в Deckhouse Stronghold другой кластер Kubernetes.

Чтобы настроить метод аутентификации, выполните следующие шаги:

1. Включите метод аутентификации Kubernetes.

   ```shell
   d8 stronghold auth enable kubernetes
   ```

1. Используйте эндпоинт `/config`, чтобы настроить Deckhouse Stronghold на взаимодействие с новым кластером Kubernetes. Для получения адреса хоста Kubernetes и TCP-порта используйте команду `d8 k cluster-info`.

   ```shell
   d8 stronghold write auth/kubernetes/config \
     token_reviewer_jwt="<your reviewer service account JWT>" \
     kubernetes_host=https://192.168.99.100:<your TCP port or blank for 443> \
     kubernetes_ca_cert=@ca.crt
   ```

   {{< alert level="warning" >}}
   Шаблон, используемый Deckhouse Stronghold для аутентификации подов, зависит от передачи JWT-токена по сети. С учётом модели безопасности Deckhouse Stronghold это допустимо, поскольку Deckhouse Stronghold является частью доверенной вычислительной системы.

   В общем случае приложения Kubernetes не должны передавать этот JWT другим приложениям, поскольку он позволяет выполнять вызовы API Kubernetes от имени подов. Это может привести к непреднамеренному предоставлению доступа третьим лицам.
   {{< /alert >}}

1. Создайте именованную роль.

   ```shell
   d8 stronghold write auth/kubernetes/role/demo \
     bound_service_account_names=myapp \
     bound_service_account_namespaces=default \
     policies=default \
     ttl=1h
   ```

   Эта роль разрешает вход для `ServiceAccount` `myapp` в неймспейсе `default` и назначает политику `default`.

## Kubernetes 1.21

Начиная с Kubernetes `1.21`, функция `BoundServiceAccountTokenVolume` включена по умолчанию. С этой версии JWT-токен, который по умолчанию монтируется в контейнеры, имеет срок действия и привязан к сроку жизни пода и `ServiceAccount`. Значение параметра JWT `iss` также зависит от конфигурации кластера.

Изменения в сроке жизни токена важны при настройке параметра `token_reviewer_jwt`. Если используется короткоживущий токен, Kubernetes отзовёт его, как только под или `ServiceAccount` будут удалены, либо после истечения срока действия. После этого Deckhouse Stronghold больше не сможет использовать `TokenReview API`.

По этой причине `Kubernetes auth` по умолчанию не проверяет эмитента `iss`. API Kubernetes выполняет ту же проверку при обработке токенов, поэтому повторная проверка эмитента на стороне Deckhouse Stronghold не требуется.

### Как работать с короткоживущими токенами Kubernetes

Существует несколько способов настроить аутентификацию для подов Kubernetes, если смонтированные по умолчанию токены подов являются короткоживущими. У каждого варианта есть свои преимущества и ограничения.

| Вариант | Все токены короткоживущие | Можно досрочно отозвать токены | Особенности |
| --- | --- | --- | --- |
| Использовать локальный токен в качестве `reviewer JWT` | Да | Да | Требуется развернуть Deckhouse Stronghold в кластере Kubernetes |
| Использовать JWT клиента в качестве `reviewer JWT` | Да | Да | Есть накладные расходы |
| Использовать долгоживущий токен в качестве `reviewer JWT` | Нет | Да | — |
| Использовать `JWT auth` вместо `Kubernetes auth` | Да | Нет | — |

{{< alert level="info" >}}
По умолчанию Kubernetes продлевает срок жизни токенов `ServiceAccount` до одного года, чтобы упростить переход на короткоживущие токены.

Если вы хотите отключить это поведение, задайте параметр [`--service-account-extend-token-expiration=false`](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/#options) для `kube-apiserver` или определите собственную конфигурацию тома `serviceAccountToken`.
{{< /alert >}}

#### Использование токена Deckhouse Stronghold в качестве reviewer JWT

Если Deckhouse Stronghold запущен в поде Kubernetes, он использует свой `ServiceAccount` для проверки токенов приложений, запущенных в том же кластере Kubernetes.

#### Использование JWT клиента в качестве reviewer JWT

При настройке `Kubernetes auth` можно не указывать `token_reviewer_jwt`. В этом случае Deckhouse Stronghold будет использовать JWT клиента как собственный токен при обращении к API Kubernetes `TokenReview`.

Также установите `disable_local_ca_jwt=true`.

При таком подходе Deckhouse Stronghold не хранит `reviewer JWT`. Это позволяет использовать короткоживущие токены во всех сценариях. При этом увеличиваются накладные расходы: нужно поддерживать `ClusterRoleBinding` для набора `ServiceAccount`, которым разрешено проходить аутентификацию в Deckhouse Stronghold.

Каждому клиенту Deckhouse Stronghold потребуется роль `system:auth-delegator`.

Пример:

```shell
d8 k create clusterrolebinding myapp-client-auth-delegator \
  --clusterrole=system:auth-delegator \
  --group=group1 \
  --serviceaccount=default:svcaccount1
```

#### Использование долгоживущих токенов

Можно вручную создать долгоживущий токен и использовать его как `token_reviewer_jwt`.

В этом примере для сервиса `myapp` потребуется роль `system:auth-delegator`.

```shell
d8 k apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: myapp-k8s-auth-secret
  annotations:
    kubernetes.io/service-account.name: myapp
type: kubernetes.io/service-account-token
EOF
```

Использование этого подхода упрощает настройку, но не позволяет воспользоваться преимуществами повышенной безопасности короткоживущих токенов.

#### Использование JWT auth вместо Kubernetes auth

`Kubernetes auth` использует `TokenReview API` Kubernetes. При этом JWT-токены, которые генерирует Kubernetes, также можно проверять через `JWT auth`, используя Kubernetes как OIDC-провайдер.

Этот вариант позволяет использовать короткоживущие токены для всех клиентов и устраняет необходимость использовать `reviewer JWT`.

{{< alert level="info" >}}
Клиентские токены нельзя отозвать до истечения их TTL. Поэтому рекомендуется использовать короткий срок жизни токена.
{{< /alert >}}

## Требования к TokenReview API

Метод `Kubernetes auth` обращается к `TokenReview API` Kubernetes, чтобы проверить, что предоставленный JWT всё ещё действителен.

`ServiceAccount`, используемые в этой схеме аутентификации, должны иметь доступ к `TokenReview API`.

Следующий пример `ClusterRoleBinding` предоставляет такие разрешения:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: role-tokenreview-binding
  namespace: default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
  - kind: ServiceAccount
    name: myapp-auth
    namespace: default
```
