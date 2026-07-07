---
title: "Метод Kubernetes"
linkTitle: "Kubernetes"
weight: 80
---

Метод Kubernetes auth можно использовать для аутентификации в Stronghold с помощью токена учётной записи сервиса Kubernetes. Этот метод упрощает использование Stronghold приложениями, работающими в подах Kubernetes.

Токен учётной записи сервиса Kubernetes также можно использовать для [входа в систему через метод JWT auth][k8s-jwt-auth].
Причины, по которым JWT auth может быть использован вместо Kubernetes auth, описаны в разделе [«Как работать с короткоживущими токенами Kubernetes»](#short-lived-tokens).

{{< alert level="info" >}}
При обновлении до Kubernetes v1.21+, убедитесь, что для параметра конфигурации
`disable_iss_validation` установлено значение `true`. Если используется путь монтирования по умолчанию,
проверить значение параметра можно с помощью команды `d8 stronghold read -field disable_iss_validation auth/kubernetes/config`.
Подробнее — в подразделе [«Изменения в поведении JWT-токенов в Kubernetes 1.21+»](#изменения-в-поведении-jwt-токенов-в-kubernetes-121).
{{< /alert >}}

## Аутентификация

### Через CLI

Имя метода аутентификации зависит от способа его создания. В составе Deckhouse Kubernetes Platform (DKP) автоматически создаётся метод `kubernetes_local`, связанный с кластером, в котором запущен Stronghold. Если метод Kubernetes auth создается вручную, по умолчанию используется имя `kubernetes`, если не указан другой путь.

Если метод аутентификации создан под другим именем, укажите его с помощью параметра `-path` в CLI. Например:

```shell
d8 stronghold write -path=your-path auth/kubernetes/login role=demo jwt=...```
```

### Через API

Используйте эндпоинт, соответствующий имени метода аутентификации. Если Stronghold развернут в составе DKP, автоматически созданный метод использует эндпоинт `auth/kubernetes_local/login`. Если метод аутентификации создан под другим именем, используйте соответствующий эндпоинт. В примере ниже используется метод аутентификации с именем `kubernetes`.

```shell-session
curl \
  --request POST \
  --data '{"jwt": "<your service account jwt>", "role": "demo"}' \
  https://stronghold.example.com/v1/auth/kubernetes/login
```

В ответе токен Stronghold возвращается в поле `auth.client_token`:

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

Перед тем как пользователи или приложения смогут проходить аутентификацию, необходимо настроить метод Kubernetes auth. Обычно эту настройку выполняет администратор или средство управления конфигурацией.

Чтобы настроить аутентификацию для другого кластера Kubernetes, включите дополнительный экземпляр метода Kubernetes auth:

1. Включите метод аутентификации Kubernetes:

   ```bash
   d8 stronghold auth enable kubernetes
   ```

1. Используйте эндпоинт `/config`, чтобы настроить Stronghold для работы с новым кластером Kubernetes. Адрес Kubernetes API и TCP-порт можно получить с помощью команды `d8 k cluster-info`.

   ```bash
   d8 stronghold write auth/kubernetes/config \
   token_reviewer_jwt="<your reviewer service account JWT>" \
   kubernetes_host=https://192.168.99.100:<your TCP port or blank for 443> \
   kubernetes_ca_cert=@ca.crt
   ```

   {{< alert level="warning" >}}
   Stronghold использует JWT-токен учётной записи сервиса для проверки подлинности через Kubernetes API. Не передавайте этот токен другим приложениям или сервисам, так как они смогут выполнять запросы к Kubernetes API с правами соответствующей учётной записи сервиса.
   {{< /alert >}}

1. Создайте именованную роль:

   ```shell
   d8 stronghold write auth/kubernetes/role/demo \
      bound_service_account_names=myapp \
      bound_service_account_namespaces=default \
      policies=default \
      ttl=1h
   ```

   Эта роль разрешает учётной записи сервиса `myapp` в неймспейсе `default` проходить аутентификацию и назначает ей политику по умолчанию.

## Изменения в поведении JWT-токенов в Kubernetes 1.21+

Начиная с версии [Kubernetes 1.21](https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.21.md#api-change-2), функция Kubernetes `BoundServiceAccountTokenVolume` включена по умолчанию. В результате JWT-токены, автоматически монтируемые в контейнеры, обладают следующими особенностями:

* имеют ограниченный срок действия;
* привязаны к сроку жизни пода и учётной записи сервиса;
* значение поля `iss` зависит от конфигурации кластера.

Изменения в сроке действия токена важно учитывать при настройке опции `token_reviewer_jwt`. Если используется короткоживущий токен, Kubernetes отзывает его после удаления пода, удаления учётной записи сервиса или истечения срока действия. После этого Stronghold больше не сможет использовать API `TokenReview`.
Подробнее — в разделе [«Как работать с короткоживущими токенами Kubernetes»](#short-lived-tokens).

По этой причине Kubernetes auth по умолчанию не проверяет значение поля `iss`. Проверка выполняется Kubernetes при обработке запросов к API `TokenReview`, поэтому повторно выполнять её на стороне Stronghold не требуется.

### Работа с короткоживущими токенами Kubernetes {#short-lived-tokens}

Существует несколько способов настроить аутентификацию для подов Kubernetes при использовании короткоживущих токенов учётных записей сервисов.
Каждый из них имеет свои преимущества и ограничения, которые рассмотрены в таблице.

| Вариант | Все токены короткоживущие | Возможен досрочный отзыв токенов | Особенности                                           |
|--------------------------------------|----------------------------|----------------------------------|-------------------------------------------------------|
| Использовать локальный токен в качестве JWT для валидации | Да | Да                               | Требуется развернуть Stronghold в кластере Kubernetes |
| Использовать клиентский JWT в качестве JWT для валидации | Да | Да                               | Требуется дополнительная настройка                    |
| Использовать долгоживущий токен в качестве JWT для валидации | Нет | Да                               | Более простая настройка                               |
| Использовать JWT auth | Да | Нет                              | Клиентские токены нельзя отозвать до истечения срока действия                                                      |

{{< alert >}}
По умолчанию Kubernetes продлевает срок действия токенов учётных записей сервисов до одного года, чтобы упростить переход на использование короткоживущих токенов.
Чтобы отключить это поведение, задайте параметр [`--service-account-extend-token-expiration=false`](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/#options) для `kube-apiserver` или определите собственную конфигурацию тома `serviceAccountToken`.
Подробный пример приведён в разделе [«Указание TTL и целевых групп»](../oidc/kubernetes/#specifying-ttl-and-audience).
{{< /alert >}}

#### Использование токена Stronghold в качестве рецензента JWT

Если Stronghold запущен в поде Kubernetes, рекомендуется использовать JWT-токен локальной учётной записи сервиса. Stronghold периодически перечитывает файл токена, поэтому поддерживаются короткоживущие токены.

Чтобы использовать локальный токен и сертификат центра сертификации, не указывайте параметры `token_reviewer_jwt` и `kubernetes_ca_cert` при настройке метода аутентификации. Stronghold автоматически загрузит их из файлов `token` и `ca.crt`, расположенных в каталоге `/var/run/secrets/kubernetes.io/serviceaccount/`.

```bash
d8 stronghold write auth/kubernetes/config \
  kubernetes_host=https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT
```

#### Использование JWT клиента в качестве рецензента JWT

При настройке Kubernetes auth можно не указывать параметр `token_reviewer_jwt`.
В этом случае Stronghold будет использовать JWT-токен клиента при обращении к API `TokenReview`.
Также необходимо установить параметр `disable_local_ca_jwt=true`.

При таком подходе Stronghold не хранит JWT-токены и может использовать короткоживущие токены для всех клиентов.
Однако потребуется дополнительно настроить роли и привязки ролей Kubernetes для учётных записей сервисов, которым разрешена аутентификация в Stronghold.
Каждой такой учётной записи необходимо предоставить кластерную роль `system:auth-delegator`:

```bash
d8 k create clusterrolebinding myapp-client-auth-delegator \
  --clusterrole=system:auth-delegator \
  --group=group1 \
  --serviceaccount=default:svcaccount1 \
  ...
```

#### Использование долгоживущих токенов

Вы можете создать долгоживущий токен, используя инструкции [Kubernetes][k8s-create-secret]
и использовать его в качестве `token_reviewer_jwt`. В примере ниже для учётной записи сервиса `myapp`
необходимо предоставить кластерную роль `system:auth-delegator`:

```bash
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

Этот способ упрощает настройку, но не позволяет воспользоваться преимуществами безопасности, которые обеспечивают короткоживущие токены.

[k8s-create-secret]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#manually-create-a-service-account-api-token

#### Использование JWT auth

Kubernetes auth использует API `TokenReview`. Вместе с тем JWT-токены Kubernetes можно проверять, используя Kubernetes в качестве OIDC-провайдера.
Инструкции по настройке приведены в документации по методу [JWT auth][k8s-jwt-auth].

[k8s-jwt-auth]: ../oidc/kubernetes/

Это решение позволяет использовать короткоживущие токены для всех клиентов и не требует настройки `token_reviewer_jwt`.
Однако клиентские токены нельзя отозвать до истечения срока их действия, поэтому рекомендуется использовать небольшое значение TTL.

## Настройка Kubernetes

Метод Kubernetes auth использует `Kubernetes TokenReview API` для
проверки того, что предоставленный JWT-токен действителен.

Учётная запись сервиса, используемая этим методом аутентификации,
должна иметь разрешение на доступ к `TokenReview API`. Пример конфигурации ресурса ClusterRoleBinding,
представленный ниже, предоставляет необходимые разрешения:

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
