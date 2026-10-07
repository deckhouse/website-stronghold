---
title: "Клиенты приложений"
description: "Работа со Stronghold из кода приложения: переменные окружения, Go, Python, Java (Spring Cloud Vault), Node.js, чтение KV, динамические учётные данные и продление токенов и аренд."
weight: 70
---

Приложение может обращаться к API Stronghold напрямую через клиентскую библиотеку. Stronghold совместим с API HashiCorp Vault, поэтому используются клиентские библиотеки Vault: `github.com/hashicorp/vault/api` для Go, `hvac` для Python, Spring Cloud Vault для Java, `node-vault` для Node.js. <!-- TODO(verify): протестированная совместимость перечисленных библиотек со Stronghold -->

Если приложение нельзя изменить или в нём сложно реализовать продление токенов и аренд, используйте [Stronghold Agent](../../../user/agent/overview/): он выполнит аутентификацию, продление и доставку секретов в файлы или переменные окружения.

## Переменные окружения

Клиентские библиотеки Vault читают стандартные переменные окружения:

| Переменная | Назначение |
| --- | --- |
| `VAULT_ADDR` | Адрес Stronghold, например `https://stronghold.example.com` |
| `VAULT_TOKEN` | Токен клиента. Не задавайте долгоживущий токен в production — используйте методы аутентификации |
| `VAULT_CACERT` | Путь к сертификату CA для проверки TLS-сертификата Stronghold |
| `VAULT_NAMESPACE` | [Пространство имён](../../../admin/namespaces/overview/) |

Утилита `d8 stronghold` использует переменные `STRONGHOLD_ADDR`, `STRONGHOLD_TOKEN` и `STRONGHOLD_CACERT`. В HTTP-запросах токен передаётся в заголовке `X-Vault-Token`.

## Жизненный цикл токенов и аренд

Каждый [токен](../../../concepts/tokens/) и каждый динамический секрет имеют срок действия ([аренду](../../../concepts/lease/)). Приложение должно:

1. Пройти аутентификацию при старте (для подов — [методом Kubernetes](../../../user/auth/kubernetes/), для виртуальных машин — например, [AppRole](../../../user/auth/approle/)).
1. Продлевать токен до истечения `ttl`, обычно после двух третей срока действия.
1. Повторно пройти аутентификацию, когда достигнут `max_ttl` и продление невозможно, а также при ответе `403`.
1. Продлевать аренды динамических секретов или запрашивать новые учётные данные до истечения аренды.
1. Отзывать аренды при штатном завершении, если учётные данные больше не нужны.

Примеры ниже используют роль Kubernetes `myapp`, секрет KV версии 2 `secret/myapp/db` и роль механизма баз данных `my-role` (см. [«Доставка секретов в поды Kubernetes»](../kubernetes-workloads/#подготовка-роль-и-политика) и [механизм секретов PostgreSQL](../../../user/secrets-engines/databases/postgresql/)). Политика должна разрешать чтение `database/creds/my-role`.

## Go

Используйте пакеты `github.com/hashicorp/vault/api` и `github.com/hashicorp/vault/api/auth/kubernetes`:

```go
package main

import (
    "context"
    "log"

    vault "github.com/hashicorp/vault/api"
    auth "github.com/hashicorp/vault/api/auth/kubernetes"
)

func main() {
    ctx := context.Background()

    // DefaultConfig читает VAULT_ADDR и VAULT_CACERT.
    client, err := vault.NewClient(vault.DefaultConfig())
    if err != nil {
        log.Fatal(err)
    }

    k8sAuth, err := auth.NewKubernetesAuth("myapp")
    if err != nil {
        log.Fatal(err)
    }
    authInfo, err := client.Auth().Login(ctx, k8sAuth)
    if err != nil {
        log.Fatal(err)
    }

    // Фоновое продление токена.
    watcher, err := client.NewLifetimeWatcher(&vault.LifetimeWatcherInput{Secret: authInfo})
    if err != nil {
        log.Fatal(err)
    }
    go watcher.Start()
    defer watcher.Stop()

    // Чтение KV версии 2.
    kv, err := client.KVv2("secret").Get(ctx, "myapp/db")
    if err != nil {
        log.Fatal(err)
    }
    password := kv.Data["password"].(string)
    _ = password

    // Динамические учётные данные базы данных.
    creds, err := client.Logical().ReadWithContext(ctx, "database/creds/my-role")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("lease %s, ttl %ds", creds.LeaseID, creds.LeaseDuration)

    // Продление аренды учётных данных.
    if _, err := client.Sys().RenewWithContext(ctx, creds.LeaseID, 3600); err != nil {
        log.Print(err)
    }

    // Когда продление токена невозможно, watcher завершается: пройдите аутентификацию заново.
    <-watcher.DoneCh()
}
```

## Python

Используйте библиотеку `hvac`:

```python
import os

import hvac

client = hvac.Client(
    url=os.environ["VAULT_ADDR"],
    verify=os.environ.get("VAULT_CACERT", True),
)

with open("/var/run/secrets/kubernetes.io/serviceaccount/token") as f:
    client.auth.kubernetes.login(role="myapp", jwt=f.read())

# Чтение KV версии 2.
secret = client.secrets.kv.v2.read_secret_version(
    path="myapp/db",
    mount_point="secret",
    raise_on_deleted_version=True,
)
password = secret["data"]["data"]["password"]

# Динамические учётные данные базы данных.
creds = client.secrets.database.generate_credentials(name="my-role")
username = creds["data"]["username"]

# Продление аренды и токена.
client.sys.renew_lease(lease_id=creds["lease_id"], increment=3600)
client.auth.token.renew_self()
```

Запускайте продление токена и аренд периодически (например, в отдельном потоке) и повторяйте вход при ошибке `hvac.exceptions.Forbidden`.

## Java (Spring Cloud Vault)

Spring Cloud Vault загружает секреты в `Environment` приложения и сам продлевает токен и аренды. Пример `application.yml`:

```yaml
spring:
  application:
    name: myapp
  config:
    import: vault://
  cloud:
    vault:
      uri: https://stronghold.example.com
      authentication: KUBERNETES
      kubernetes:
        role: myapp
        kubernetes-path: kubernetes
      kv:
        enabled: true
        backend: secret
        default-context: myapp/db
      database:
        enabled: true
        backend: database
        role: my-role
```

Значения из `secret/myapp/db` становятся свойствами Spring (например, `${password}`), а динамические учётные данные базы данных по умолчанию попадают в `spring.datasource.username` и `spring.datasource.password`.

## Node.js

Используйте библиотеку `node-vault`:

```javascript
const fs = require('fs');
const vault = require('node-vault')({
  apiVersion: 'v1',
  endpoint: process.env.VAULT_ADDR,
});

async function main() {
  const jwt = fs.readFileSync('/var/run/secrets/kubernetes.io/serviceaccount/token', 'utf8');
  const login = await vault.kubernetesLogin({ role: 'myapp', jwt });
  vault.token = login.auth.client_token;

  // Чтение KV версии 2.
  const kv = await vault.read('secret/data/myapp/db');
  const password = kv.data.data.password;

  // Динамические учётные данные базы данных.
  const creds = await vault.read('database/creds/my-role');

  // Продление токена и аренды.
  await vault.tokenRenewSelf();
  await vault.write('sys/leases/renew', { lease_id: creds.lease_id, increment: 3600 });
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
```

## Рекомендации

- Проверяйте TLS-сертификат Stronghold. Не отключайте проверку в production.
- Не записывайте токены и значения секретов в логи.
- Кешируйте статические секреты и перечитывайте их по расписанию, а не при каждом запросе.
- Для одноразовой передачи секрета другому процессу используйте [обертывание ответа](../../../concepts/response-wrapping/).
