---
title: "Kubernetes"
linkTitle: "Kubernetes"
weight: 80
description: "Аутентификация в Deckhouse Stronghold с помощью токена ServiceAccount Kubernetes."
---

## Kubernetes

Метод `kubernetes auth` можно использовать для аутентификации в **Deckhouse Stronghold** с помощью токена `ServiceAccount` Kubernetes. Этот метод аутентификации позволяет удобно использовать Stronghold в подах Kubernetes.

Также токен `ServiceAccount` можно использовать для входа через `JWT`-аутентификацию, если Kubernetes выступает как OIDC-провайдер. Эти сценарии различаются: `Kubernetes auth` использует `TokenReview API`, а `JWT auth` — проверку токена по криптографии открытого ключа.

## Когда использовать Kubernetes auth

Метод `Kubernetes` обычно выбирают, если:

- приложение работает внутри Kubernetes;
- нужно выдавать токены Stronghold на основе identity Kubernetes;
- требуется проверка токена через `TokenReview API`;
- нужен более управляемый сценарий для workload'ов внутри кластера.

## Аутентификация

### Через CLI

По умолчанию используется путь `/kubernetes_local`. Если метод включён по другому пути, укажите нужный `-path`.

Пример:

```bash
d8 stronghold write auth/kubernetes/login role=demo jwt=...
```

### Через API

По умолчанию используется endpoint `auth/kubernetes_local/login`. Если метод смонтирован по другому пути, замените стандартный путь на фактический.

```bash
curl \
    --request POST \
    --data '{"jwt": "<your service account jwt>", "role": "demo"}' \
    https://stronghold.example.com/v1/auth/kubernetes/login
```

Пример ответа:

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

## Базовая конфигурация

Перед тем как workload'ы смогут аутентифицироваться, метод нужно предварительно настроить.

### Шаг 1. Включите метод

```bash
d8 stronghold auth enable kubernetes
```

### Шаг 2. Настройте подключение к кластеру Kubernetes

```bash
d8 stronghold write auth/kubernetes/config \
   token_reviewer_jwt="<your reviewer service account JWT>" \
   kubernetes_host=https://192.168.99.100:<your TCP port or blank for 443> \
   kubernetes_ca_cert=@ca.crt
```

### Шаг 3. Создайте роль

```bash
d8 stronghold write auth/kubernetes/role/demo \
   bound_service_account_names=myapp \
   bound_service_account_namespaces=default \
   policies=default \
   ttl=1h
```

Такая роль разрешает вход для `ServiceAccount` `myapp` в namespace `default` и назначает политику `default`.

## Особенности Kubernetes 1.21+

Начиная с Kubernetes `1.21`, смонтированные по умолчанию токены `ServiceAccount` стали короткоживущими и привязанными к сроку жизни пода и самой учётной записи сервиса.

Это важно, потому что:

- `token_reviewer_jwt` тоже может быть короткоживущим;
- после истечения TTL Stronghold больше не сможет использовать его для `TokenReview`;
- поведение зависит от того, какой вариант интеграции вы выбрали.

## Как работать с короткоживущими токенами

В документации описано несколько вариантов:

- использовать локальный токен Stronghold как reviewer JWT;
- использовать JWT клиента Stronghold в качестве reviewer JWT;
- использовать долгоживущий токен как reviewer JWT;
- вместо `Kubernetes auth` использовать `JWT auth`.

### Использование JWT клиента в качестве reviewer JWT

Если при настройке `Kubernetes auth` опустить `token_reviewer_jwt`, Stronghold будет использовать JWT клиента Stronghold в качестве собственного токена при обращении к `TokenReview API`. В таком случае также нужно установить `disable_local_ca_jwt=true`.

Это позволяет использовать короткоживущие токены, но увеличивает эксплуатационные требования: каждому клиенту может понадобиться доступ к `system:auth-delegator`.

### Использование долгоживущих токенов

Можно вручную создать долгоживущий токен `ServiceAccount` и использовать его как `token_reviewer_jwt`. Это упрощает настройку, но лишает преимуществ короткоживущих токенов.

### Использование JWT auth вместо Kubernetes auth

Если для вас критично использовать короткоживущие токены и не зависеть от reviewer JWT, можно использовать `JWT auth` и Kubernetes как OIDC-провайдера.

> Примечание  
> В этом случае токены не могут быть отозваны до истечения TTL, поэтому рекомендуется использовать короткий срок жизни.

## Требования к TokenReview API

Метод `Kubernetes auth` обращается к `Kubernetes TokenReview API`, чтобы убедиться, что предоставленный JWT по-прежнему действителен.

Для этого `ServiceAccount`, используемый в схеме аутентификации, должен иметь право доступа к `TokenReview API`.

Пример `ClusterRoleBinding`:

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

## Практические рекомендации

- Используйте `Kubernetes auth`, если приложение работает внутри Kubernetes и вам нужен сценарий с `TokenReview API`.
- Если важны короткоживущие токены и минимальная зависимость от reviewer JWT, рассмотрите `JWT auth` с Kubernetes как OIDC-провайдером.
- Ограничивайте роли через `bound_service_account_names` и `bound_service_account_namespaces`.
- Не передавайте токены `ServiceAccount` сторонним системам без необходимости.
- Проверяйте, что у reviewer-аккаунта есть право `system:auth-delegator`.
