---
title: "cert-manager"
description: "Выпуск TLS-сертификатов в Kubernetes через cert-manager с Issuer типа vault, механизмом секретов PKI Stronghold и аутентификацией Kubernetes."
weight: 20
---

cert-manager выпускает сертификаты через Issuer или ClusterIssuer типа `vault`, который обращается к [механизму секретов PKI](../../../user/secrets-engines/pki/) по пути `pki/sign/<роль>`. Stronghold совместим с API Vault, поэтому такой Issuer можно направить на Stronghold. <!-- TODO(verify): протестированная совместимость Issuer типа vault в cert-manager со Stronghold, в том числе с ГОСТ-ключами PKI -->

В Deckhouse Platform cert-manager устанавливается модулем [`cert-manager`](/products/kubernetes-platform/documentation/v1/modules/cert-manager/). <!-- TODO(verify): путь к документации модуля cert-manager, пространство имён и имя ServiceAccount контроллера в DKP -->

## Настройка Stronghold

Предполагается, что механизм PKI включён по пути `pki`, создан корневой или промежуточный сертификат и роль `example-dot-ru`, как описано в разделе [«Механизм секретов PKI»](../../../user/secrets-engines/pki/).

1. Создайте политику, разрешающую подписывать запросы на сертификат по роли:

   ```bash
   d8 stronghold policy write cert-manager-pki - <<'POLICY'
   path "pki/sign/example-dot-ru" {
     capabilities = ["create", "update"]
   }
   POLICY
   ```

1. Создайте роль [метода Kubernetes](../../../user/auth/kubernetes/) для учётной записи сервиса `stronghold-issuer` в пространстве имён `myapp`. Параметр `audience` должен совпадать с аудиторией токена, которую запрашивает cert-manager: для Issuer это `vault://<namespace>/<имя Issuer>`, для ClusterIssuer — `vault://<имя ClusterIssuer>`:

   ```bash
   d8 stronghold write auth/kubernetes/role/cert-manager-myapp \
     bound_service_account_names=stronghold-issuer \
     bound_service_account_namespaces=myapp \
     audience=vault://myapp/stronghold \
     policies=cert-manager-pki \
     ttl=20m
   ```

   В DP в режиме `Automatic` метод аутентификации Kubernetes для текущего кластера создаётся модулем по пути `kubernetes_local`: в этом случае используйте `auth/kubernetes_local` вместо `auth/kubernetes` в командах и в `mountPath` Issuer. Путь `auth/kubernetes` относится к методу, включённому вручную.

## Настройка cert-manager

1. Создайте учётную запись сервиса и разрешите cert-manager запрашивать для неё токены:

   ```yaml
   apiVersion: v1
   kind: ServiceAccount
   metadata:
     name: stronghold-issuer
     namespace: myapp
   ---
   apiVersion: rbac.authorization.k8s.io/v1
   kind: Role
   metadata:
     name: stronghold-issuer-token
     namespace: myapp
   rules:
     - apiGroups: [""]
       resources: ["serviceaccounts/token"]
       resourceNames: ["stronghold-issuer"]
       verbs: ["create"]
   ---
   apiVersion: rbac.authorization.k8s.io/v1
   kind: RoleBinding
   metadata:
     name: stronghold-issuer-token
     namespace: myapp
   roleRef:
     apiGroup: rbac.authorization.k8s.io
     kind: Role
     name: stronghold-issuer-token
   subjects:
     - kind: ServiceAccount
       name: cert-manager
       namespace: d8-cert-manager
   ```

1. Создайте Issuer. В `caBundle` укажите сертификат CA, которым подписан TLS-сертификат Stronghold, в кодировке base64:

   ```yaml
   apiVersion: cert-manager.io/v1
   kind: Issuer
   metadata:
     name: stronghold
     namespace: myapp
   spec:
     vault:
       server: https://stronghold.example.com
       path: pki/sign/example-dot-ru
       caBundle: <base64-кодированный CA>
       auth:
         kubernetes:
           role: cert-manager-myapp
           mountPath: /v1/auth/kubernetes
           serviceAccountRef:
             name: stronghold-issuer
   ```

1. Проверьте, что Issuer готов:

   ```bash
   d8 k -n myapp get issuer stronghold
   ```

   В колонке `READY` должно быть значение `True`.

1. Запросите сертификат:

   ```yaml
   apiVersion: cert-manager.io/v1
   kind: Certificate
   metadata:
     name: www
     namespace: myapp
   spec:
     secretName: www-tls
     commonName: www.my-website.ru
     dnsNames:
       - www.my-website.ru
     duration: 72h
     renewBefore: 24h
     issuerRef:
       name: stronghold
       kind: Issuer
   ```

cert-manager сохранит сертификат и закрытый ключ в Secret `www-tls` и будет перевыпускать сертификат до истечения срока. Значение `duration` не должно превышать `max_ttl` роли PKI.

## Другие способы аутентификации

Issuer типа `vault` также поддерживает аутентификацию через AppRole (`auth.appRole` с `roleId` и ссылкой на Secret с `secret_id`) и через токен (`auth.tokenSecretRef`). Предпочтительнее аутентификация Kubernetes: она не требует хранить долгоживущие учётные данные в кластере.

Использование Stronghold PKI как внешнего CA для Istio описано в разделе [«Service mesh»](../service-mesh/).
