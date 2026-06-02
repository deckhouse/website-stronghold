---
title: "JWT/OIDC"
linkTitle: "JWT/OIDC"
description: "Аутентификация Stronghold Agent с помощью JWT/OIDC"
weight: 70
---

JWT/OIDC-аутентификация позволяет Stronghold Agent использовать внешний провайдер идентификации
для получения токена Stronghold.
Этот метод подходит для сценариев,
в которых уже используется централизованная система управления идентификацией.

На этой странице описано,
как работает JWT/OIDC-аутентификация,
в каких случаях её использовать
и как настроить Stronghold Agent для работы с JWT.

## Когда использовать JWT/OIDC

Используйте JWT/OIDC в следующих случаях:

- Нужно интегрировать Stronghold с корпоративной системой единого входа.
- Приложение получает JWT от внешнего провайдера идентификации.
- Требуется использовать федеративную аутентификацию.
- Аутентификация выполняется в CI/CD через OIDC,
  например в GitHub Actions или GitLab CI.

Для сценариев на виртуальных машинах и bare metal
без внешнего провайдера идентификации обычно проще использовать AppRole.

## Как работает JWT/OIDC-аутентификация

Обычно процесс выглядит следующим образом:

1. Приложение или внешний процесс получает JWT от провайдера идентификации.
1. Stronghold Agent читает JWT из файла.
1. Agent передаёт JWT в метод аутентификации Stronghold.
1. Stronghold проверяет подпись токена и его claims.
1. После успешной проверки Stronghold выдаёт собственный токен.
1. Agent использует этот токен для работы с секретами.

JWT и токен Stronghold — это разные токены.
JWT нужен для входа,
а токен Stronghold используется для дальнейшей работы Agent.

## Преимущества JWT/OIDC

JWT/OIDC-аутентификация даёт следующие преимущества:

- Позволяет использовать существующий провайдер идентификации.
- Снижает потребность в отдельных учётных данных для каждого приложения.
- Позволяет использовать claims для ограничения доступа.
- Упрощает интеграцию с корпоративным SSO и CI/CD.

## Ограничения и особенности

Учитывайте следующие особенности этого метода:

- Agent не выпускает JWT самостоятельно.
- JWT обычно имеет короткий срок действия.
- После истечения JWT нужно получить новый токен от провайдера идентификации.
- Stronghold Agent автоматически обновляет токен Stronghold,
  но не всегда может автоматически обновлять исходный JWT.

Если требуется длительная непрерывная работа,
заранее предусмотрите механизм обновления JWT во внешнем процессе.

## Настройка метода JWT в Stronghold

Ниже приведён пример настройки JWT-метода с использованием Keycloak.

### Шаг 1. Включите метод аутентификации

```shell
stronghold auth enable jwt
```

### Шаг 2. Настройте JWT/OIDC

```shell
stronghold write auth/jwt/config \
  oidc_discovery_url="https://keycloak.example.com/realms/myrealm" \
  oidc_client_id="stronghold" \
  oidc_client_secret="client-secret-from-keycloak" \
  default_role="default"
```

### Шаг 3. Создайте политику

```shell
stronghold policy write myapp-jwt-policy - <<EOF
path "secret/data/myapp/*" {
  capabilities = ["read"]
}
path "database/creds/myapp" {
  capabilities = ["read"]
}
EOF
```

### Шаг 4. Создайте роль

```shell
stronghold write auth/jwt/role/myapp-role \
  role_type="jwt" \
  bound_audiences="stronghold" \
  user_claim="sub" \
  bound_subject="service-account-myapp" \
  token_ttl=1h \
  token_max_ttl=4h \
  token_policies="myapp-jwt-policy"
```

При необходимости можно дополнительно ограничить роль по claims:

```shell
stronghold write auth/jwt/role/myapp-role \
  role_type="jwt" \
  bound_audiences="stronghold" \
  user_claim="sub" \
  bound_claims='{"environment":"production","app":"myapp"}' \
  claim_mappings='{"department":"dept"}' \
  token_policies="myapp-jwt-policy"
```

## Получение JWT от провайдера идентификации

Ниже приведён пример получения JWT из Keycloak:

```shell
curl -X POST "https://keycloak.example.com/realms/myrealm/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=myapp-service" \
  -d "client_secret=service-secret" \
  -d "grant_type=client_credentials" \
  | jq -r '.access_token' > /tmp/jwt-token.txt
```

После этого передайте JWT на сервер,
где работает Stronghold Agent.

## Доставка JWT на сервер

Скопируйте JWT-файл на целевой сервер:

```shell
scp /tmp/jwt-token.txt root@app-server.example.com:/etc/stronghold-agent/jwt-token
```

Ограничьте права доступа к файлу:

```shell
ssh root@app-server.example.com << 'ENDSSH'
chown stronghold-agent:stronghold-agent /etc/stronghold-agent/jwt-token
chmod 0640 /etc/stronghold-agent/jwt-token
ENDSSH
```

## Конфигурация Stronghold Agent

Создайте файл `/etc/stronghold-agent/agent.hcl`:

```hcl
stronghold {
  address = "https://stronghold.example.com:8200"
}

auto_auth {
  method "jwt" {
    mount_path = "auth/jwt"
    config = {
      path = "/etc/stronghold-agent/jwt-token"
      role = "myapp-role"
    }
  }

  sink "file" {
    config = {
      path = "/var/run/stronghold-agent/token"
      mode = 0640
    }
  }
}

template {
  source = "/etc/stronghold-agent/templates/database.conf.ctmpl"
  destination = "/etc/myapp/database.conf"
  perms = "0600"
}
```

В этом примере Agent:

- читает JWT из файла `/etc/stronghold-agent/jwt-token`;
- проходит аутентификацию через метод `auth/jwt`;
- получает токен Stronghold;
- записывает токен в файловый sink;
- использует его для рендеринга шаблона.

## Проверка

Запустите Stronghold Agent и проверьте журнал:

```shell
systemctl start stronghold-agent
journalctl -u stronghold-agent -n 50 | grep -i "authentication successful"
cat /var/run/stronghold-agent/token
```

## Проверка содержимого JWT

При необходимости проверьте claims токена:

```shell
cat /etc/stronghold-agent/jwt-token | cut -d. -f2 | base64 -d | jq
```

{{< alert level="info" >}}
JWT и токен Stronghold выполняют разные задачи.
JWT используется для аутентификации,
а токен Stronghold — для дальнейшей работы Agent.
{{< /alert >}}

## Практические рекомендации

При использовании JWT/OIDC соблюдайте следующие рекомендации:

- Используйте короткий TTL для JWT.
- Настраивайте `bound_audiences`.
- Используйте `bound_subject` или `bound_claims`
  для дополнительного ограничения доступа.
- По возможности используйте OIDC discovery.
- Логируйте успешные и неуспешные попытки аутентификации.
- Не храните JWT в Git или в другой системе контроля версий в открытом виде.
