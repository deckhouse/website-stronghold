---
title: "Kubernetes как OIDC-провайдер"
linkTitle: "Kubernetes"
weight: 40
description: "Использование Kubernetes как OIDC-провайдера для JWT/OIDC-аутентификации в Deckhouse Stronghold."
---

Kubernetes может выступать как OIDC-провайдер для Deckhouse Stronghold. В этом сценарии Stronghold проверяет JWT токены `ServiceAccount` с помощью метода аутентификации `JWT/OIDC auth`.

Этот вариант подходит, если приложения уже используют токены `ServiceAccount` и их нужно аутентифицировать в Stronghold без метода `Kubernetes auth`. Проверка выполняется либо через OIDC Discovery URL, либо через публичные ключи подписи JWT.

{{< alert level="warning" >}}
Метод `JWT`-аутентификации не использует API Kubernetes `TokenReview`. Вместо этого Stronghold проверяет JWT с помощью криптографии с открытым ключом. Из-за этого изменения состояния токена или связанного `ServiceAccount` могут не учитываться до окончания срока действия токена. Чтобы снизить риск, используйте короткий TTL токенов `ServiceAccount` или настройте отдельный метод `Kubernetes auth`, который использует API Kubernetes `TokenReview`.
{{< /alert >}}

## Когда использовать

Используйте этот сценарий в следующих случаях:

- Нужно аутентифицировать в Stronghold приложения, использующие токены `ServiceAccount`.
- Kubernetes-кластер может выступать как OIDC-провайдер.
- Требуется использовать `JWT auth`, а не `Kubernetes auth`.

## Варианты настройки

Для Kubernetes как OIDC-провайдера доступны два основных варианта:

- Через OIDC Discovery URL.
- Через публичные ключи для проверки JWT.

## Настройка через OIDC Discovery URL

Это наиболее простой вариант, если в кластере включена опция `ServiceAccountIssuerDiscovery` и доступны OIDC discovery metadata.

### Требования

Перед началом настройки убедитесь, что:

- включена опция `ServiceAccountIssuerDiscovery`;
- параметр `--service-account-issuer` у `kube-apiserver` содержит адрес, доступный из Stronghold;
- используются короткоживущие токены `ServiceAccount`;
- если OIDC Discovery URL использует пользовательский сертификат, настройте в Stronghold доверие к соответствующему CA-сертификату.

Опция `ServiceAccountIssuerDiscovery` доступна начиная с Kubernetes 1.18 и включена по умолчанию начиная с Kubernetes 1.20.

Короткоживущие токены `ServiceAccount`, которые монтируются в поды, используются по умолчанию начиная с Kubernetes 1.21.

### Шаг 1. Откройте доступ к OIDC Discovery metadata

Убедитесь, что OIDC Discovery URL доступен без аутентификации. Подробнее см. в документации Kubernetes по ServiceAccount issuer discovery:

```shell
d8 k create clusterrolebinding oidc-reviewer \
  --clusterrole=system:service-account-issuer-discovery \
  --group=system:unauthenticated
```

### Шаг 2. Получите значение issuer

Получите значение `issuer` из OIDC-конфигурации кластера:

```shell
ISSUER="$(d8 k get --raw /.well-known/openid-configuration | jq -r '.issuer')"
```

### Шаг 3. Включите и настройте JWT auth

Включите метод аутентификации `jwt` и укажите полученное значение `issuer` в параметре `oidc_discovery_url`:

```shell
d8 stronghold auth enable jwt
d8 stronghold write auth/jwt/config oidc_discovery_url="${ISSUER}"
```

Stronghold использует этот URL для получения OIDC discovery metadata.

### Шаг 4. Создайте роль

После настройки метода `JWT auth` создайте роль, как описано в разделе [«Создание роли и аутентификация»](#создание-роли-и-аутентификация).

## Настройка через публичные ключи

Используйте этот вариант в следующих случаях:

- API Kubernetes недоступен из Stronghold;
- один эндпоинт `JWT auth` должен обслуживать несколько кластеров;
- удобнее использовать набор публичных ключей напрямую.

### Требования

Перед началом настройки убедитесь, что:

- включена опция `ServiceAccountIssuerDiscovery` или есть доступ к файлу `/etc/kubernetes/pki/sa.pub`;
- используются короткоживущие токены `ServiceAccount`.

Опция `ServiceAccountIssuerDiscovery` доступна начиная с Kubernetes 1.18 и включена по умолчанию начиная с Kubernetes 1.20.

Если у вас есть доступ к файлу `/etc/kubernetes/pki/sa.pub` на master-узле кластера, шаги по получению ключа и преобразованию его в формат `PEM` можно пропустить, так как ключ уже хранится в нужном формате.

Короткоживущие токены `ServiceAccount`, которые монтируются в поды, используются по умолчанию начиная с Kubernetes 1.21.

1. Получите JWKS с публичными ключами

Если OIDC metadata доступна, получите JWKS через `jwks_uri`:

```shell
d8 k get --raw "$(d8 k get --raw /.well-known/openid-configuration | jq -r '.jwks_uri' | sed -r 's/.*\.[^/]+(.*)/\1/')"
```

1. Преобразуйте ключи в PEM

Преобразуйте ключи из формата `JWK` в формат `PEM`.

1. Настройте JWT auth

Настройте метод `JWT auth` на использование полученных ключей:

```shell
d8 stronghold write auth/jwt/config \
  jwt_validation_pubkeys="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9...
-----END PUBLIC KEY-----","-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9...
-----END PUBLIC KEY-----"
```

1. Создайте роль

После настройки метода `JWT auth` создайте роль, как описано в разделе [«Создание роли и аутентификация»](#создание-роли-и-аутентификация).

## Создание роли и аутентификация

После настройки метода аутентификации `JWT auth` создайте роль и выполните тестовую аутентификацию.

По умолчанию в примерах ниже используется токен `ServiceAccount`, смонтированный в под. Если нужно управлять значением `aud` или TTL, используйте отдельное монтирование `serviceAccountToken`, как показано в разделе [«Настройка TTL и API audience для токена ServiceAccount»](#настройка-ttl-и-api-audience-для-токена-serviceaccount).

### Как получить API audience

Выберите значение из поля `aud`, которое используется в токене `ServiceAccount`. Ниже показаны два способа получить значение `aud`.

Пример получения `aud` через создание токена требует Kubernetes версии 1.24.0 или выше и совместимой команды `d8 k create token`:

```shell
d8 k create token default | cut -f2 -d. | base64 --decode
```

Пример получения токена из пода:

```shell
d8 k exec my-pod -- cat /var/run/secrets/kubernetes.io/serviceaccount/token | cut -f2 -d. | base64 --decode
```

Например, значение `aud` может выглядеть как `https://kubernetes.default.svc.cluster.local`. Всегда проверяйте фактическое значение в своём кластере.

### Создание роли

Следующий пример создаёт роль, которую сможет использовать `ServiceAccount` `default` в неймспейсе `default`:

```shell
d8 stronghold write auth/jwt/role/my-role \
  role_type="jwt" \
  bound_audiences="<AUDIENCE-FROM-PREVIOUS-STEP>" \
  user_claim="sub" \
  bound_subject="system:serviceaccount:default:default" \
  policies="default" \
  ttl="1h"
```

### Аутентификация

После этого приложения в подах и другие клиенты, у которых есть доступ к JWT токену `ServiceAccount`, смогут пройти аутентификацию:

```shell
d8 stronghold write auth/jwt/login \
  role=my-role \
  jwt=@/var/run/secrets/kubernetes.io/serviceaccount/token
```

Эквивалентный API-запрос:

```shell
curl \
  --fail \
  --request POST \
  --data '{"jwt":"<JWT-TOKEN-HERE>","role":"my-role"}' \
  "${STRONGHOLD_ADDR}/v1/auth/jwt/login"
```

## Настройка TTL и API audience для токена ServiceAccount

Если нужно задать собственные TTL или `aud` для токенов `ServiceAccount`, используйте отдельное монтирование `serviceAccountToken` в поде.

Этот вариант особенно полезен, если нельзя отключить параметр `--service-account-extend-token-expiration` у `kube-apiserver` и при этом нужно использовать короткий TTL токена.

Для токена из примера ниже укажите в роли значение:

```text
bound_audiences=stronghold
```

Пример:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  # Параметр automountServiceAccountToken избыточен для этого примера,
  # так как используемый mountPath совпадает с путём по умолчанию.
  # Такое перекрытие предотвращает создание токена по умолчанию.
  # При другом пути монтирования этот параметр помогает гарантировать,
  # что в под будет смонтирован только один токен.
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

## Что дальше

После настройки Kubernetes как OIDC-провайдера выполните следующие действия:

1. Проверьте, что метод `JWT auth` настроен корректно.
1. Создайте роль с нужными значениями `bound_audiences` и `bound_subject`.
1. Выполните тестовую аутентификацию из пода.
1. Убедитесь, что выданный токен Stronghold содержит ожидаемые политики.

## Дополнительная информация

Для получения дополнительной информации используйте следующие материалы:

- [ServiceAccount issuer discovery в документации Kubernetes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#service-account-issuer-discovery)
- [Параметры kube-apiserver в документации Kubernetes](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/#options)
