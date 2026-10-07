---
title: "GitOps"
description: "Как не хранить секреты в Git при работе с Argo CD и Flux: External Secrets Operator, argocd-vault-plugin, SOPS и механизм секретов GitOps."
weight: 40
---

В GitOps-процессе Git — источник истины для манифестов, но значения секретов в репозиторий попадать не должны. В Git хранятся только ссылки на секреты, а значения берутся из Stronghold в кластере.

Инструменты на этой странице (External Secrets Operator, argocd-vault-plugin, SOPS) разработаны для HashiCorp Vault. Stronghold совместим с API Vault, поэтому их можно направить на Stronghold. <!-- TODO(verify): протестированная совместимость ESO, argocd-vault-plugin и SOPS (hc_vault_transit) со Stronghold -->

## Выбор подхода

| Подход | Что хранится в Git | Где оказывается значение секрета | Инструменты |
| --- | --- | --- | --- |
| External Secrets Operator | Манифесты `ExternalSecret` со ссылками на пути Stronghold | Secret Kubernetes (etcd) | Argo CD, Flux |
| argocd-vault-plugin | Манифесты с плейсхолдерами `<path:...#key>` | Отрендеренные манифесты Argo CD и Secret Kubernetes (etcd) | Argo CD |
| SOPS с ключом механизма Transit | Зашифрованные файлы | Secret Kubernetes (etcd) после расшифровки | Flux, Argo CD с плагином |
| CSI или env-injector модуля `secrets-store-integration` | Манифесты подов с аннотациями или томами CSI | Только в поде | Argo CD, Flux |

Рекомендуемый вариант — External Secrets Operator или модуль [`secrets-store-integration`](/products/kubernetes-platform/documentation/v1/modules/secrets-store-integration/): в Git не попадают ни значения, ни зашифрованные данные, а права на чтение задаются ролями Stronghold. Сравнение способов доставки секретов в поды приведено на странице [«Доставка секретов в поды Kubernetes»](../kubernetes-workloads/).

## External Secrets Operator с Argo CD или Flux

Храните в Git манифесты `SecretStore` (или `ClusterSecretStore`) и `ExternalSecret`. Argo CD или Flux применяет их как обычные ресурсы, а оператор создаёт Secret со значениями из Stronghold:

```yaml
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: myapp-db
  namespace: myapp
spec:
  refreshInterval: 15m
  secretStoreRef:
    name: stronghold
    kind: SecretStore
  target:
    name: myapp-db
  dataFrom:
    - extract:
        key: myapp/db
```

Настройка `SecretStore` и роли Kubernetes auth описана в разделе [«Доставка секретов в поды Kubernetes»](../kubernetes-workloads/#external-secrets-operator).

Если Argo CD показывает Secret как рассинхронизированный, исключите сгенерированные оператором Secret из отслеживания Argo CD, так как их нет в Git.

## argocd-vault-plugin

argocd-vault-plugin подставляет значения в манифесты при рендеринге в Argo CD. В Git хранятся плейсхолдеры:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: myapp-db
  namespace: myapp
type: Opaque
stringData:
  password: <path:secret/data/myapp/db#password>
```

Плагин настраивается переменными окружения в repo-server Argo CD, например:

```yaml
env:
  - name: AVP_TYPE
    value: vault
  - name: VAULT_ADDR
    value: https://stronghold.example.com
  - name: AVP_AUTH_TYPE
    value: k8s
  - name: AVP_K8S_ROLE
    value: argocd-repo-server
```

{{< alert level="warning" >}}
Отрендеренные манифесты с открытыми значениями видны Argo CD и могут кешироваться. Выдавайте роли плагина доступ только к нужным путям и ограничьте, кто может просматривать манифесты приложений в Argo CD.
{{< /alert >}}

## SOPS с механизмом Transit

SOPS умеет шифровать файлы ключом [механизма секретов Transit](../../../user/secrets-engines/transit/) через API, совместимый с Vault. Зашифрованные файлы хранятся в Git, а Flux расшифровывает их при применении.

1. Создайте ключ шифрования:

   ```bash
   d8 stronghold secrets enable transit
   d8 stronghold write -f transit/keys/sops
   ```

1. Зашифруйте файл:

   ```bash
   export VAULT_ADDR=https://stronghold.example.com
   export VAULT_TOKEN=<токен с правами на transit/encrypt/sops>
   sops --encrypt --hc-vault-transit $VAULT_ADDR/v1/transit/keys/sops secret.yaml > secret.enc.yaml
   ```

Для расшифровки во Flux контроллеру kustomize-controller нужен токен Stronghold с правами на `transit/decrypt/sops`. Токен хранится в Secret кластера, поэтому используйте отдельную политику и короткий срок жизни с продлением. <!-- TODO(verify): формат Secret для расшифровки hc-vault во Flux и совместимость с ГОСТ-ключами Transit -->

## Механизм секретов GitOps

Для управления конфигурацией самого Stronghold (политики, методы аутентификации, точки монтирования) используйте встроенный [механизм секретов GitOps](../../../user/secrets-engines/gitops/overview/). Он отслеживает Git-репозиторий, проверяет подписи коммитов и применяет декларативную конфигурацию через API Stronghold. Формат описан в разделе [«Формат конфигурации»](../../../user/secrets-engines/gitops/configuration-format/).

Механизм GitOps не доставляет секреты в приложения: используйте его вместе с одним из подходов выше.
