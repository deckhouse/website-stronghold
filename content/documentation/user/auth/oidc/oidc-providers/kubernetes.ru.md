---
title: "Kubernetes как провайдер OIDC"
linkTitle: "Kubernetes"
weight: 40
description: "Использование Kubernetes как OIDC-провайдера для JWT/OIDC-аутентификации в Deckhouse Stronghold."
---

Kubernetes может выступать в качестве OIDC-провайдера, чтобы Stronghold мог проверять токены `ServiceAccount` с помощью `JWT/OIDC auth` [5].

> Предупреждение  
> Механизм `JWT`-аутентификации не использует API Kubernetes `TokenReview`, а вместо этого использует криптографию с открытым ключом для проверки содержимого JWT. Это означает, что токены, которые были отозваны Kubernetes, будут считаться действительными до истечения срока их действия. Чтобы снизить этот риск, используйте короткие TTL для токенов `ServiceAccount` или используйте [Kubernetes](../../Kubernetes/) как отдельный метод аутентификации, который использует API `TokenReview` [5].

## Когда использовать этот сценарий

Этот сценарий полезен, если:

- Stronghold должен проверять JWT токены `ServiceAccount`;
- Kubernetes-кластер может выступать как OIDC-провайдер;
- вы хотите использовать `JWT auth`, а не `Kubernetes auth` [5].

## Варианты настройки

Для Kubernetes как OIDC-провайдера можно использовать два основных варианта:

- через **OIDC Discovery URL**;
- через **публичные ключи** для проверки JWT [5].

## Использование OIDC Discovery URL

Это наиболее простой вариант, если ваш кластер Kubernetes поддерживает автоматическую публикацию OIDC metadata [5].

### Требования к кластеру

- включена опция `ServiceAccountIssuerDiscovery`;
- значение `--service-account-issuer` у `kube-apiserver` содержит адрес, доступный из Stronghold;
- используются короткоживущие токены `ServiceAccount` [5].

### Шаги настройки

1. Убедитесь, что URL OIDC Discovery не требует аутентификации:

   ```bash
   d8 k create clusterrolebinding oidc-reviewer \
      --clusterrole=system:service-account-issuer-discovery \
      --group=system:unauthenticated
   ```

2. Определите `issuer` URL кластера:

   ```bash
   ISSUER="$(d8 k get --raw /.well-known/openid-configuration | jq -r '.issuer')"
   ```

3. Включите и настройте `JWT auth` в Stronghold:

   ```bash
   d8 stronghold auth enable jwt
   d8 stronghold write auth/jwt/config oidc_discovery_url="${ISSUER}"
   ```

4. Настройте роли, как описано ниже в разделе **Создание ролей и аутентификация** [5].

## Использование публичных ключей для проверки JWT

Этот вариант полезен, если:

- API Kubernetes недоступен из Stronghold;
- вы хотите, чтобы один endpoint `JWT auth` обслуживал несколько кластеров;
- вам удобнее использовать цепочку публичных ключей напрямую [5].

### Требования к кластеру

- включена опция `ServiceAccountIssuerDiscovery`, либо у вас есть доступ к `/etc/kubernetes/pki/sa.pub`;
- используются короткоживущие токены `ServiceAccount` [5].

### Шаги настройки

1. Получите открытый ключ подписи токенов `ServiceAccount` через `jwks_uri`:

   ```bash
   d8 k get --raw "$(d8 k get --raw /.well-known/openid-configuration | jq -r '.jwks_uri' | sed -r 's/.*\.[^/]+(.*)/\1/')"
   ```

2. Преобразуйте ключи из формата `JWK` в `PEM`.

3. Настройте `JWT auth` на использование полученных ключей:

   ```bash
   d8 stronghold write auth/jwt/config \
      jwt_validation_pubkeys="-----BEGIN PUBLIC KEY-----
   MIIBIjANBgkqhkiG9...
   -----END PUBLIC KEY-----","-----BEGIN PUBLIC KEY-----
   MIIBIjANBgkqhkiG9...
   -----END PUBLIC KEY-----"
   ```

4. Настройте роли, как описано ниже [5].

## Создание ролей и аутентификация

После того как endpoint `JWT auth` настроен, можно создать роль и выполнить аутентификацию.

### Как получить audience

Выберите значение из стандартных `aud`, используемых в токене `ServiceAccount`.

Пример получения `aud` через создание токена:

```bash
d8 k create token default | cut -f2 -d. | base64 --decode
```

Пример получения токена из пода:

```bash
d8 k exec my-pod -- cat /var/run/secrets/kubernetes.io/serviceaccount/token | cut -f2 -d. | base64 --decode
```

### Создание роли

Создайте роль, которую сможет использовать `ServiceAccount` `default` в namespace `default`:

```bash
d8 stronghold write auth/jwt/role/my-role \
   role_type="jwt" \
   bound_audiences="<AUDIENCE-FROM-PREVIOUS-STEP>" \
   user_claim="sub" \
   bound_subject="system:serviceaccount:default:default" \
   policies="default" \
   ttl="1h"
```

### Аутентификация

После этого клиент, у которого есть JWT токен `ServiceAccount`, может пройти аутентификацию:

```bash
d8 stronghold write auth/jwt/login \
   role=my-role \
   jwt=@/var/run/secrets/kubernetes.io/serviceaccount/token
```

Эквивалентный API-запрос:

```bash
curl \
   --fail \
   --request POST \
   --data '{"jwt":"<JWT-TOKEN-HERE>","role":"my-role"}' \
   "${STRONGHOLD_ADDR}/v1/auth/jwt/login"
```

## Указание TTL и audience

Если требуется задать собственный TTL или audience для токенов `ServiceAccount`, можно использовать отдельное монтирование `serviceAccountToken` в поде [5].

Пример:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  automountServiceAccountToken: false
  containers:
    - name: nginx
      image: nginx
      volumeMounts:
      - name: custom-token
        mountPath: /var/run/secrets/kubernetes.io/serviceaccount
  volumes:
  - name: custom-token
    projected:
      defaultMode: 420
      sources:
      - serviceAccountToken:
          path: token
          expirationSeconds: 600
          audience: stronghold
      - configMap:
          name: kube-root-ca.crt
          items:
          - key: ca.crt
            path: ca.crt
      - downwardAPI:
          items:
          - fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
            path: namespace
```

При использовании такого токена в роли Stronghold необходимо указать:

```text
bound_audiences=stronghold
```

## Что дальше

После настройки Kubernetes как OIDC-провайдера:

- проверьте, что endpoint `JWT auth` настроен корректно;
- создайте роль с нужными `bound_audiences` и `bound_subject`;
- выполните тестовую аутентификацию из пода;
- убедитесь, что выданный токен Stronghold содержит ожидаемые политики.
