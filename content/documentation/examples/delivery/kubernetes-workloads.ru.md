---
title: "Доставка секретов в поды Kubernetes"
linkTitle: "Kubernetes-нагрузки"
description: "Сравнение способов доставки секретов Stronghold в поды Kubernetes: CSI, env-injector, Stronghold Agent, External Secrets Operator и прямой вызов API."
weight: 10
---

Приложение в Kubernetes может получать секреты из Stronghold несколькими способами. Все они опираются на [метод аутентификации Kubernetes](../../../user/auth/kubernetes/): под предъявляет токен своей учётной записи сервиса (ServiceAccount), а Stronghold выдаёт токен с политиками, привязанными к роли.

## Сравнение способов

| Способ | Как секрет попадает в под | Попадает ли секрет в etcd | Обновление секрета | Когда использовать |
| --- | --- | --- | --- | --- |
| CSI-драйвер модуля `secrets-store-integration` | Файл в томе, смонтированном в контейнер | Нет | При перемонтировании тома или перезапуске пода<!-- TODO(verify): поддерживает ли модуль периодическое обновление файлов --> | Приложение читает секреты из файлов; кластер под управлением Deckhouse Platform |
| env-injector модуля `secrets-store-integration` | Переменная окружения процесса | Нет | Только при перезапуске пода | Приложение читает конфигурацию только из переменных окружения |
| Stronghold Agent (init-контейнер или sidecar) | Файл, отрендеренный по шаблону, или переменные окружения дочернего процесса | Нет | Sidecar перерисовывает шаблоны при изменении секрета или продлении аренды | Нужны шаблоны, динамические секреты, продление аренд без изменения кода |
| External Secrets Operator (ESO) | Объект Secret Kubernetes, который под подключает как том или переменные | Да | По `refreshInterval`: Secret обновляется, файлы в томе — с задержкой kubelet, переменные — только после перезапуска | Приложение и манифесты уже рассчитаны на обычные Secret; GitOps |
| Прямой вызов API из приложения | Приложение запрашивает секрет само | Нет | Полностью под контролем приложения | Приложение использует SDK и умеет продлевать токены и аренды |

Общие рекомендации:

- Если секрет не должен храниться в etcd, выбирайте CSI, env-injector, Stronghold Agent или прямой вызов API.
- Секреты в переменных окружения видны процессам с доступом к `/proc/<pid>/environ` и могут попасть в дампы и логи. По возможности используйте файлы.
- Для [динамических секретов](../../../user/secrets-engines/databases/overview/) и сертификатов с коротким сроком жизни используйте Stronghold Agent в режиме sidecar или прямой вызов API: только они продлевают аренды (leases) в течение жизни пода.

## Подготовка: роль и политика

Пример ниже подходит для всех способов. Предполагается, что [механизм секретов KV версии 2](../../../user/secrets-engines/kv/kv-v2/) включён по пути `secret`, а метод аутентификации Kubernetes — по пути `kubernetes`.

1. Создайте политику, которая разрешает чтение секретов приложения:

   ```bash
   d8 stronghold policy write myapp-read - <<'POLICY'
   path "secret/data/myapp/*" {
     capabilities = ["read"]
   }
   POLICY
   ```

1. Сохраните секрет:

   ```bash
   d8 stronghold kv put -mount=secret myapp/db username=app password=S3cr3t
   ```

1. Создайте роль, связывающую учётную запись сервиса `myapp` в пространстве имён `myapp` с политикой:

   ```bash
   d8 stronghold write auth/kubernetes/role/myapp \
     bound_service_account_names=myapp \
     bound_service_account_namespaces=myapp \
     policies=myapp-read \
     ttl=1h
   ```

   В DP в режиме `Automatic` метод аутентификации Kubernetes для текущего кластера создаётся модулем по пути `kubernetes_local`: в этом случае используйте `auth/kubernetes_local` вместо `auth/kubernetes` здесь и в примерах ниже (`mount_path`, URL входа). Путь `auth/kubernetes` относится к методу, включённому вручную.

1. Создайте учётную запись сервиса в кластере:

   ```bash
   d8 k create namespace myapp
   d8 k -n myapp create serviceaccount myapp
   ```

Настройка самого метода аутентификации (`auth/kubernetes/config`) описана в разделе [«Метод Kubernetes»](../../../user/auth/kubernetes/). В Deckhouse Platform модуль `stronghold` в режиме `Automatic` [автоматически подключает текущий кластер](../../../install/dkp/configuration/) для работы модуля `secrets-store-integration`.

## CSI-драйвер secrets-store-integration

Модуль Deckhouse Platform [`secrets-store-integration`](/products/kubernetes-platform/documentation/v1/modules/secrets-store-integration/) монтирует секреты в под как файлы через CSI-драйвер. Секрет не сохраняется в объекте Secret и не попадает в etcd.

Пример (имена ресурсов, API-версия и параметры приведены по документации модуля и требуют проверки для вашей версии DP):

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: SecretsStoreImport
metadata:
  name: myapp-db
  namespace: myapp
spec:
  type: CSI
  role: myapp
  files:
    - name: db-password
      source:
        path: secret/data/myapp/db
        key: password
---
apiVersion: v1
kind: Pod
metadata:
  name: myapp
  namespace: myapp
spec:
  serviceAccountName: myapp
  containers:
    - name: app
      image: registry.example.com/myapp:1.0.0
      volumeMounts:
        - name: secrets
          mountPath: /mnt/secrets
          readOnly: true
  volumes:
    - name: secrets
      csi:
        driver: secrets-store.csi.deckhouse.io
        volumeAttributes:
          secretsStoreImport: myapp-db
```

Приложение читает пароль из файла `/mnt/secrets/db-password`.

## env-injector secrets-store-integration

env-injector из того же модуля подставляет значения секретов в переменные окружения при старте контейнера. Значения существуют только в памяти процесса и не записываются в спецификацию пода.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: myapp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
      annotations:
        secrets-store.deckhouse.io/role: myapp
    spec:
      serviceAccountName: myapp
      containers:
        - name: app
          image: registry.example.com/myapp:1.0.0
          env:
            - name: DB_USER
              value: secrets-store:secret/data/myapp/db#username
            - name: DB_PASSWORD
              value: secrets-store:secret/data/myapp/db#password
```

Чтобы приложение получило новое значение секрета, перезапустите под, например командой `d8 k -n myapp rollout restart deployment myapp`.

## Stronghold Agent

[Stronghold Agent](../../../user/agent/overview/) аутентифицируется методом Kubernetes, получает секреты и рендерит их в файлы по шаблонам. В поде Agent запускается:

- как init-контейнер с `exit_after_auth = true` — секреты рендерятся один раз при старте пода;
- как sidecar-контейнер — Agent продлевает токен и аренды и перерисовывает шаблоны при изменении секретов.

Файлы передаются приложению через общий том `emptyDir` с `medium: Memory`, поэтому секреты не записываются на диск узла.

Пример конфигурации Agent в ConfigMap (формат секций описан в разделе [«Настройки»](../../../user/agent/settings/)):

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-agent
  namespace: myapp
data:
  agent.hcl: |
    stronghold {
      address = "https://stronghold.example.com"
    }

    auto_auth {
      method "kubernetes" {
        mount_path = "auth/kubernetes"
        config = {
          role = "myapp"
        }
      }
    }

    template {
      destination = "/secrets/db.env"
      contents    = <<-EOT
      {{ with secret "secret/data/myapp/db" }}
      DB_USER={{ .Data.data.username }}
      DB_PASSWORD={{ .Data.data.password }}
      {{ end }}
      EOT
    }
```

В образе Stronghold точкой входа служит `stronghold`, поэтому в `args` передаются `agent` и `-config=<путь к agent.hcl>`; имя образа берите из своего registry.

Пример пода с Agent в режиме sidecar:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
  namespace: myapp
spec:
  serviceAccountName: myapp
  containers:
    - name: stronghold-agent
      image: registry.example.com/stronghold:<version>
      args: ["agent", "-config=/etc/stronghold-agent/agent.hcl"]
      volumeMounts:
        - name: agent-config
          mountPath: /etc/stronghold-agent
        - name: secrets
          mountPath: /secrets
    - name: app
      image: registry.example.com/myapp:1.0.0
      volumeMounts:
        - name: secrets
          mountPath: /secrets
          readOnly: true
  volumes:
    - name: agent-config
      configMap:
        name: myapp-agent
    - name: secrets
      emptyDir:
        medium: Memory
```

Для варианта с init-контейнером перенесите контейнер `stronghold-agent` в `initContainers` и добавьте в `agent.hcl` параметр `exit_after_auth = true`.

## External Secrets Operator

[External Secrets Operator](https://external-secrets.io/) — сторонний оператор с провайдером `vault`. Stronghold совместим с API Vault, поэтому провайдер можно направить на Stronghold. <!-- TODO(verify): протестированная совместимость ESO с Stronghold и поддерживаемая версия API external-secrets.io -->

ESO создаёт обычный объект Secret, поэтому значение секрета хранится в etcd. Включите шифрование объектов Secret в etcd средствами кластера и ограничьте доступ к Secret через RBAC.

```yaml
apiVersion: external-secrets.io/v1
kind: SecretStore
metadata:
  name: stronghold
  namespace: myapp
spec:
  provider:
    vault:
      server: https://stronghold.example.com
      path: secret
      version: v2
      auth:
        kubernetes:
          mountPath: kubernetes
          role: myapp
          serviceAccountRef:
            name: myapp
---
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
  data:
    - secretKey: password
      remoteRef:
        key: myapp/db
        property: password
```

## Прямой вызов API

Приложение может само пройти аутентификацию с токеном учётной записи сервиса и прочитать секрет. Пример на `curl`:

```bash
SA_TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)

STRONGHOLD_TOKEN=$(curl -s --request POST \
  --data "{\"role\": \"myapp\", \"jwt\": \"${SA_TOKEN}\"}" \
  https://stronghold.example.com/v1/auth/kubernetes/login | jq -r '.auth.client_token')

curl -s --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  https://stronghold.example.com/v1/secret/data/myapp/db | jq '.data.data'
```

Примеры для SDK и рекомендации по продлению токенов и аренд приведены в разделе [«Клиенты приложений»](../app-clients/).
