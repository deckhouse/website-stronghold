---
title: "Kubernetes как провайдер OIDC"
linkTitle: "Kubernetes"
weight: 40
---

Kubernetes может выступать в качестве OIDC-провайдера, чтобы Stronghold мог подтверждать токены учётных записей сервиса с помощью метода JWT/OIDC.

{{< alert >}}
Механизм JWT-аутентификации **не** использует API Kubernetes `TokenReview` для проверки токенов, а вместо этого использует криптографию с открытым ключом для проверки содержимого JWT. Это означает, что токены, которые были отозваны Kubernetes, будут считаться действительными до истечения срока их действия.

Чтобы снизить этот риск, используйте короткие TTL для токенов учётных записей сервиса или используйте метод аутентификации [Kubernetes auth](../../kubernetes/), в котором применяется API `TokenReview`.
{{< /alert >}}

## Использование адреса автонастройки

При использовании автонастройки нужно указать только OIDC discovery URL. Если OIDC URL использует пользовательский сертификат, также понадобится доверенный CA. Это самый простой режим настройки, если кластер Kubernetes соответствует требованиям.

Требования к кластеру Kubernetes:

* Включённая опция [`ServiceAccountIssuerDiscovery`][k8s-sa-issuer-discovery].
  * Доступна с версии 1.18, включена по умолчанию с версии 1.20.
* Значение URL в параметре `--service-account-issuer` компонента `kube-apiserver` должно содержать адрес, доступный из Stronghold. Для большинства managed-сервисов Kubernetes этот адрес публичный.
* Должны использоваться короткоживущие токены для учётных записей сервиса Kubernetes.
  * По умолчанию такое поведение включено для токенов, подключаемых в поды, начиная с Kubernetes 1.21.

Чтобы включить автоматическую настройку, выполните следующие шаги:

1. Убедитесь, что URL-адрес обнаружения OIDC не требует аутентификации, как описано в [документации Kubernetes][k8s-sa-issuer-discovery]:

   ```bash
   d8 k create clusterrolebinding oidc-reviewer  \
   --clusterrole=system:service-account-issuer-discovery \
   --group=system:unauthenticated
   ```

1. Определите значение issuer URL для вашего кластера.

   ```bash
   ISSUER="$(d8 k get --raw /.well-known/openid-configuration | jq -r '.issuer')"
   ```

1. Включите и настройте аутентификацию JWT в Stronghold.

   ```bash
   d8 stronghold auth enable jwt
   d8 stronghold write auth/jwt/config oidc_discovery_url="${ISSUER}"
   ```

1. Настройте [необходимые роли](#создание-ролей-и-аутентификация).

[k8s-sa-issuer-discovery]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#service-account-issuer-discovery

## Использование публичных ключей для проверки JWT

Этот метод применяют, когда API Kubernetes недоступен из Stronghold или когда один эндпоинт JWT auth должен обслуживать несколько кластеров Kubernetes с использованием цепочки публичных ключей.

Требования к кластеру Kubernetes:

* Включенная опция [`ServiceAccountIssuerDiscovery`][k8s-sa-issuer-discovery].
  * Доступна с версии 1.18, включена по умолчанию с версии 1.20.
  * Это требование не обязательно, если у вас есть доступ к файлу `/etc/kubernetes/pki/sa.pub` на master-узле кластера. В этом случае можно пропустить шаги по получению ключа и преобразованию его в формат PEM, так как ключ уже находится в файле в нужном формате.
* Должны использоваться короткоживущие токены для учётных записей сервиса Kubernetes.
  * По умолчанию такое поведение включено для токенов, подключаемых в поды, начиная с Kubernetes 1.21.

Чтобы настроить JWT auth с использованием публичных ключей Kubernetes, выполните следующие шаги:

1. Получите открытый ключ подписи токенов учётных записей сервиса из JWKS URI вашего кластера.

   ```bash
   # jwks_uri доступен в /.well-known/openid-configuration
   d8 k get --raw "$(d8 k get --raw /.well-known/openid-configuration | jq -r '.jwks_uri' | sed -r 's/.*\.[^/]+(.*)/\1/')"
   ```

1. Преобразуйте ключи из формата JWK в формат PEM. Это можно сделать с помощью консольной утилиты или [онлайн-конвертера JWK в PEM][jwk-to-pem].

1. Настройте эндпоинт JWT auth на использование полученных ключей.

    ```bash
    d8 stronghold write auth/jwt/config \
       jwt_validation_pubkeys="-----BEGIN PUBLIC KEY-----
    MIIBIjANBgkqhkiG9...
    -----END PUBLIC KEY-----","-----BEGIN PUBLIC KEY-----
    MIIBIjANBgkqhkiG9...
    -----END PUBLIC KEY-----"
    ```

1. Настройте [необходимые роли](#создание-ролей-и-аутентификация).

[jwk-to-pem]: https://8gwifi.org/jwkconvertfunctions.jsp

## Создание ролей и аутентификация

После того как эндпоинт JWT auth настроен, можно настроить роль и пройти аутентификацию. Далее предполагается, что вы используете токен учётной записи сервиса, доступный по умолчанию во всех подах. Если нужно контролировать целевую группу (audience) или TTL, обратитесь к разделу [«Указание TTL и целевых групп»](#specifying-ttl-and-audience).

Выберите любое значение из набора стандартных целевых групп по умолчанию. В этих примерах в массиве `aud` есть только одна целевая группа, `https://kubernetes.default.svc.cluster.local`.

Чтобы найти целевую группу по умолчанию, создайте новый токен (требуется [утилита Deckhouse CLI `d8`](/products/kubernetes-platform/documentation/v1/cli/d8/) v1.24.0+):

```shell-session
d8 k create token default | cut -f2 -d. | base64 --decode
{"aud":["https://kubernetes.default.svc.cluster.local"], ... "sub":"system:serviceaccount:default:default"}
```

Также можно прочитать токен из запущенного пода:

```shell-session
d8 k exec my-pod -- cat /var/run/secrets/kubernetes.io/serviceaccount/token | cut -f2 -d. | base64 --decode
{"aud":["https://kubernetes.default.svc.cluster.local"], ... "sub":"system:serviceaccount:default:default"}
```

Создайте роль для JWT auth, которую сможет использовать учётная запись сервиса `default` в неймспейсе `default`.

```bash
d8 stronghold write auth/jwt/role/my-role \
role_type="jwt" \
bound_audiences="<AUDIENCE-FROM-PREVIOUS-STEP>" \
user_claim="sub" \
bound_subject="system:serviceaccount:default:default" \
policies="default" \
ttl="1h"
```

Теперь поды или клиенты, имеющие доступ к JWT учётной записи сервиса, смогут аутентифицироваться с помощью этого токена.

```bash
d8 stronghold write auth/jwt/login \
  role=my-role \
  jwt=@/var/run/secrets/kubernetes.io/serviceaccount/token
```

Пример эквивалентного HTTP-запроса:

```bash
curl \
  --fail \
  --request POST \
  --data '{"jwt":"<JWT-TOKEN-HERE>","role":"my-role"}' \
  "${STRONGHOLD_ADDR}/v1/auth/jwt/login"
```

## Указание TTL и целевых групп {#specifying-ttl-and-audience}

Если нужно указать пользовательский TTL или целевую группу для токенов учётных записей сервиса, в следующем манифесте пода показано монтирование тома, которое переопределяет автоматически монтируемый токен по умолчанию. Это особенно актуально, если вы не можете отключить флаг [--service-account-extend-token-expiration][k8s-extended-tokens] для `kube-apiserver` и хотите использовать короткие TTL.

При использовании полученного токена укажите значение `bound_audiences=stronghold` при создании ролей в JWT auth.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  # Параметр automountServiceAccountToken является лишним в этом примере, поскольку используемый
  # mountPath совпадает с путем по умолчанию. Это перекрытие предотвращает
  # создание токена по умолчанию. Однако вы можете использовать этот параметр, чтобы
  # обеспечить монтирование только одного токена, если вы выберете другой путь монтирования.
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
              expirationSeconds: 600 # Минимальный TTL — 10 минут.
              audience: stronghold   # Должен совпадать со значением `bound_audiences` вашей роли.
          # Остальные параметры добавлены для имитации обычного поведения при создании токена
          # и создают те же объекты, что и при включенном параметре automountServiceAccountToken.
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

[k8s-extended-tokens]: https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/#options
