---
title: "Основные возможности"
linkTitle: "Возможности"
description: "Основные возможности Stronghold Agent"
weight: 30
---

Stronghold Agent помогает приложениям получать секреты из Deckhouse Stronghold без прямой интеграции с API. Он поддерживает несколько режимов работы: рендеринг шаблонов в файлы, передачу секретов через переменные окружения, автоматическую аутентификацию, кеширование токенов, обновление секретов и работу через локальный API Proxy.

На этой странице собраны основные возможности Stronghold Agent и типовые сценарии их использования.

## Templating

Templating позволяет создавать файлы конфигурации, заполненные секретами из Deckhouse Stronghold. Для этого Stronghold Agent использует язык шаблонов Consul Template.

Есть два режима работы с шаблонами:

1. `template` — рендеринг в файл. Agent создаёт или обновляет файл на диске, например `application.properties`, `nginx.conf` или `*.pem`, и при необходимости выполняет команду для reload сервиса.
2. `env_template` вместе с `exec` — рендеринг в переменные окружения и запуск процесса. Agent формирует значения переменных окружения и запускает приложение как дочерний процесс. Если секреты меняются, процесс можно перезапустить.

### Как это работает

Обычно Stronghold Agent делает следующее:

1. Читает файл-шаблон с плейсхолдерами.
2. Запрашивает секреты из Deckhouse Stronghold.
3. Подставляет реальные значения в шаблон.
4. Сохраняет итоговый файл с нужными правами доступа.
5. При необходимости выполняет команду для перезагрузки приложения.

### Когда использовать

Templating подходит для таких сценариев:

- legacy-приложения, которые читают конфигурацию из файлов;
- приложения без поддержки API Deckhouse Stronghold;
- доставка секретов в `.properties`, `.conf`, `.ini`, `.yaml`;
- работа с динамическими учётными данными для баз данных;
- доставка PKI-сертификатов.

### Синтаксис шаблонов

Базовая структура:

```go
{{ with secret "path/to/secret" }}
  {{ .Data.field_name }}
{{ end }}
```

Для KV v2:

```go
{{ with secret "secret/data/myapp" }}
username = {{ .Data.data.username }}
password = {{ .Data.data.password }}
{{ end }}
```

Для динамических секретов, например для базы данных или PKI:

```go
{{ with secret "database/creds/myapp" }}
DB_USER={{ .Data.username }}
DB_PASS={{ .Data.password }}
{{ end }}
```

### Основные функции шаблонов

| Функция | Описание | Пример |
| --- | --- | --- |
| `secret` | Получение секрета | `{{ with secret "secret/data/myapp" }}{{ .Data.data.password }}{{ end }}` |
| `base64Encode` | Кодирование в Base64 | `{{ "password" \| base64Encode }}` |
| `base64Decode` | Декодирование из Base64 | `{{ .Data.cert \| base64Decode }}` |
| `toJSON` | Преобразование в JSON | `{{ .Data \| toJSON }}` |
| `toYAML` | Преобразование в YAML | `{{ .Data \| toYAML }}` |
| `toLower` / `toUpper` | Изменение регистра | `{{ .Data.name \| toUpper }}` |
| `trim` | Удаление пробелов | `{{ .Data.value \| trim }}` |
| `range` | Итерация по массиву | `{{ range .Items }}{{ .Name }}{{ end }}` |
| `env` | Получение переменной окружения | `{{ env "HOME" }}` |
| `timestamp` | Получение текущего времени | `{{ timestamp "2006-01-02 15:04:05" }}` |

### Пошаговый пример: рендеринг файла конфигурации

Сценарий: legacy Java-приложение читает database credentials из `application.properties`.

#### Шаг 1. Сохраните секреты в Deckhouse Stronghold

```bash
stronghold kv put secret/myapp/config \
  db_host=postgres.prod.example.com \
  db_port=5432 \
  db_name=production \
  db_user=app_user \
  db_password=SecureP@ssw0rd
```

#### Шаг 2. Создайте файл-шаблон

Создайте `/etc/myapp/templates/application.properties.ctmpl`:

```text
# Database Configuration.
{{ with secret "secret/data/myapp/config" }}
spring.datasource.url=jdbc:postgresql://{{ .Data.data.db_host }}:{{ .Data.data.db_port }}/{{ .Data.data.db_name }}
spring.datasource.username={{ .Data.data.db_user }}
spring.datasource.password={{ .Data.data.db_password }}
{{ end }}
# Подключение пула.
spring.datasource.hikari.maximum-pool-size=10
spring.datasource.hikari.minimum-idle=5
```

#### Шаг 3. Настройте Stronghold Agent

Создайте `/etc/stronghold-agent/agent.hcl`:

```hcl
stronghold {
  address = "https://stronghold.example.com:8200"
}

auto_auth {
  method {
    type = "approle"
    config = {
      role_id_file_path = "/etc/stronghold-agent/role-id"
      secret_id_file_path = "/etc/stronghold-agent/secret-id"
      remove_secret_id_file_after_reading = false
    }
  }

  sink {
    type = "file"
    config = {
      path = "/var/run/stronghold-agent/token"
    }
  }
}

template {
  source      = "/etc/myapp/templates/application.properties.ctmpl"
  destination = "/etc/myapp/application.properties"
  perms       = "0600"
  user        = "myapp"
  group       = "myapp"
  command     = "systemctl reload myapp"
  command_timeout = "30s"

  wait {
    min = "2s"
    max = "10s"
  }

  error_on_missing_key = true
}
```

#### Шаг 4. Проверьте конфигурацию

```bash
stronghold -config=/etc/stronghold-agent/agent.hcl -exit-after-auth -log-level=debug
```

Во время выполнения Agent:

1. читает и разбирает `agent.hcl`;
2. подключается к серверу Deckhouse Stronghold;
3. проходит аутентификацию через AppRole;
4. получает токен и сохраняет его в sink;
5. запрашивает секреты;
6. рендерит шаблон и создаёт `/etc/myapp/application.properties`;
7. завершает работу с кодом `0`.

#### Проверка результата

```bash
ls -la /var/run/stronghold-agent/token
ls -la /etc/myapp/application.properties
sudo cat /etc/myapp/application.properties
```

После успешной проверки Agent можно запустить как `systemd`-сервис:

```bash
systemctl start stronghold-agent
systemctl status stronghold-agent
journalctl -u stronghold-agent -f
```

### Продвинутые сценарии templating

#### Динамические учётные данные для баз данных

```hcl
{{ with secret "database/creds/myapp-role" }}
# Auto-generated credentials (TTL: 1h)
# Rotation: automatic
DB_USER={{ .Data.username }}
DB_PASS={{ .Data.password }}
DB_LEASE_ID={{ .LeaseID }}
DB_LEASE_DURATION={{ .LeaseDuration }}
{{ end }}
```

В этом сценарии Agent:

- запрашивает временные учётные данные;
- обновляет файл при ротации;
- выполняет команду перезагрузки приложения.

#### PKI-сертификаты

```hcl
{{ with secret "pki/issue/web-server" "common_name=app.example.com" "ttl=720h" }}
{{ .Data.certificate }}
{{ .Data.ca_chain }}
{{ end }}
```

```hcl
{{ with secret "pki/issue/web-server" "common_name=app.example.com" "ttl=720h" }}
{{ .Data.private_key }}
{{ end }}
```

Пример конфигурации:

```hcl
template {
  source      = "/etc/nginx/ssl/cert.pem.ctmpl"
  destination = "/etc/nginx/ssl/cert.pem"
  perms       = "0644"
}
template {
  source      = "/etc/nginx/ssl/key.pem.ctmpl"
  destination = "/etc/nginx/ssl/key.pem"
  perms       = "0600"
  command     = "systemctl reload nginx"
}
```

#### Условная логика и циклы

```go
{{ with secret "secret/data/myapp/config" }}
{{ if eq .Data.data.environment "production" }}
LOG_LEVEL=ERROR
DEBUG_MODE=false
{{ else }}
LOG_LEVEL=DEBUG
DEBUG_MODE=true
{{ end }}
API_KEY={{ .Data.data.api_key }}
{{ end }}
```

```go
{{ with secret "secret/data/myapp/allowed-ips" }}
{{ range $index, $ip := .Data.data.ips }}
allow {{ $ip }};
{{ end }}
{{ end }}
```

#### Несколько секретов в одном файле

```go
{{ with secret "database/creds/app" }}
DB_USER={{ .Data.username }}
DB_PASS={{ .Data.password }}
{{ end }}

{{ with secret "secret/data/myapp/api-keys" }}
STRIPE_KEY={{ .Data.data.stripe_key }}
SENDGRID_KEY={{ .Data.data.sendgrid_key }}
{{ end }}

{{ with secret "secret/data/myapp/redis" }}
REDIS_HOST={{ .Data.data.host }}
REDIS_PASSWORD={{ .Data.data.password }}
{{ end }}
```

### Важные параметры блока `template`

| Параметр | Описание | Пример |
| --- | --- | --- |
| `source` | Путь к файлу-шаблону | `/etc/app/template.ctmpl` |
| `destination` | Путь к итоговому файлу | `/etc/app/config.conf` |
| `perms` | Права доступа | `"0600"`, `"0644"` |
| `user` | Владелец файла | `"myapp"` |
| `group` | Группа файла | `"myapp"` |
| `command` | Команда после рендеринга | `"systemctl reload app"` |
| `command_timeout` | Таймаут команды | `"30s"` |
| `error_on_missing_key` | Ошибка при отсутствии ключа | `true` / `false` |
| `wait.min` | Минимальное время между обновлениями | `"2s"` |
| `wait.max` | Максимальное время между обновлениями | `"10s"` |
| `backup` | Создавать резервную копию | `true` / `false` |

## `template` и `env_template`

`template` рендерит секреты в файл на диске.

Используйте этот режим, если приложение читает конфигурацию из файлов:

- `.conf`, `.ini`, `.yaml`, `.properties`;
- TLS-файлов `*.pem`;
- ключей и сертификатов.

Для `template` доступны файловые параметры: `destination`, `perms`, `user`, `group`, `backup`, `wait`, а также `command` для reload сервиса.

`env_template` вместе с `exec` передаёт секреты в переменные окружения дочернего процесса.

Используйте этот режим, если:

- приложение читает конфигурацию из ENV;
- допустим перезапуск процесса при ротации секретов;
- секреты не должны записываться на диск.

Что важно учитывать:

- каждый `env_template` задаёт значение ровно одной переменной окружения;
- блок всегда пишется как `env_template "VAR_NAME" { ... }`;
- `env_template` не создаёт `.env`-файл;
- поля `destination`, `perms`, `command`, `wait` и похожие в `env_template` не поддерживаются.

Практические нюансы:

- если Agent работает под `systemd` с hardening и `ProtectSystem=strict`, в `ReadWritePaths` нужно добавить директорию из `template.destination`, иначе запись будет запрещена;
- если вы запускаете Docker-контейнер через `env_template`, переменные окружения нужно явно передавать в `docker run` через `--env`.

## Режим Process Supervisor

Process Supervisor позволяет Stronghold Agent запускать приложение как дочерний процесс и передавать секреты напрямую в переменные окружения.

### Как это работает

В этом режиме Agent:

1. запускается как родительский процесс;
2. запрашивает секреты из Deckhouse Stronghold;
3. формирует переменные окружения;
4. запускает приложение как дочерний процесс;
5. отслеживает изменения секретов;
6. при изменении секретов перезапускает приложение с новыми значениями.

### Ограничения

У режима есть несколько ограничений:

- `exec` должен использоваться хотя бы с одним `env_template`;
- `env_template` нельзя комбинировать с `template` и `api_proxy` в одном конфигурационном файле;
- каждый `env_template` формирует только одну переменную окружения.

### Преимущества

Этот режим полезен, потому что:

- секреты не записываются на диск;
- приложение автоматически перезапускается при обновлении секретов;
- секреты изолированы на уровне процесса;
- режим хорошо подходит для 12-factor-приложений;
- он упрощает миграцию legacy-приложений на переменные окружения.

### Когда использовать

Process Supervisor подходит, если:

- приложение читает конфигурацию из переменных окружения;
- есть повышенные требования к безопасности;
- приложение контейнеризировано, но работает на VM;
- используются динамические credentials с частой ротацией;
- нужен удобный режим для разработки и тестирования.

### Пошаговый пример

Сценарий: Java Spring Boot приложение читает секреты из переменных окружения.

#### Шаг 1. Подготовьте приложение

```text
server.port=8080
spring.datasource.url=${DB_URL}
spring.datasource.username=${DB_USERNAME}
spring.datasource.password=${DB_PASSWORD}
spring.datasource.driver-class-name=org.postgresql.Driver
spring.jpa.hibernate.ddl-auto=validate
spring.jpa.properties.hibernate.dialect=org.hibernate.dialect.PostgreSQLDialect
api.key=${API_KEY}
```

#### Шаг 2. Подготовьте секреты в Deckhouse Stronghold

```bash
stronghold write database/config/postgresql \
  plugin_name=postgresql-database-plugin \
  allowed_roles="myapp-role" \
  connection_url="postgresql://{{username}}:{{password}}@postgres.prod:5432/myapp?sslmode=require" \
  username="vault_admin" \
  password="admin_password"

stronghold write database/roles/myapp-role \
  db_name=postgresql \
  creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
  default_ttl="1h" \
  max_ttl="24h"

stronghold kv put secret/myapp/config \
  api_key=sk_live_1234567890abcdef
```

#### Шаг 3. Настройте Agent

```hcl
stronghold {
  address = "https://stronghold.example.com:8200"
}

auto_auth {
  method {
    type = "approle"
    config = {
      role_id_file_path = "/etc/stronghold-agent/role-id"
      secret_id_file_path = "/etc/stronghold-agent/secret-id"
      remove_secret_id_file_after_reading = false
    }
  }
}

exec {
  command = ["/usr/bin/java", "-jar", "/opt/myapp/demo-application.jar"]
  restart_on_secret_changes = "always"
  restart_stop_signal = "SIGTERM"
}

env_template "DB_URL" {
  contents = "jdbc:postgresql://postgres.prod:5432/myapp"
}
env_template "DB_USERNAME" {
  contents = "{{ with secret \"database/creds/myapp-role\" }}{{ .Data.username }}{{ end }}"
}
env_template "DB_PASSWORD" {
  contents = "{{ with secret \"database/creds/myapp-role\" }}{{ .Data.password }}{{ end }}"
}
env_template "API_KEY" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.api_key }}{{ end }}"
}
env_template "JAVA_OPTS" {
  contents = "-Xmx2g -Xms512m -XX:+UseG1GC"
}
env_template "SPRING_PROFILES_ACTIVE" {
  contents = "production"
}
```

#### Шаг 4. Запустите Agent

```bash
stronghold -config=/etc/stronghold-agent/agent.hcl
```

После запуска Agent:

- проходит аутентификацию;
- получает database credentials и API-ключ;
- запускает Java-приложение с секретами в ENV;
- при ротации credentials перезапускает приложение.

### Примеры для разных приложений

#### Go-приложение

```hcl
exec {
  command = ["/opt/myapp/myapp-server"]
  restart_on_secret_changes = "always"
  restart_stop_signal = "SIGTERM"
}
env_template "DB_HOST" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.db_host }}{{ end }}"
}
env_template "DB_PORT" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.db_port }}{{ end }}"
}
env_template "DB_NAME" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.db_name }}{{ end }}"
}
env_template "DB_USER" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.db_user }}{{ end }}"
}
env_template "DB_PASSWORD" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.db_password }}{{ end }}"
}
env_template "API_KEY" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.api_key }}{{ end }}"
}
env_template "LOG_LEVEL" {
  contents = "info"
}
```

#### Docker-контейнер на VM

```hcl
exec {
  command = [
    "/usr/bin/docker", "run", "--rm",
    "--name", "myapp",
    "-p", "8080:8080",
    "--env", "DOCKER_ENV_API_KEY",
    "--env", "DOCKER_ENV_DATABASE_URL",
    "myapp:latest"
  ]
  restart_on_secret_changes = "always"
  restart_stop_signal = "SIGTERM"
}
env_template "DOCKER_ENV_API_KEY" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.api_key }}{{ end }}"
}
env_template "DOCKER_ENV_DATABASE_URL" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.database_url }}{{ end }}"
}
```

### Важные параметры блока `exec`

| Параметр | Описание | Значение по умолчанию |
| --- | --- | --- |
| `command` | Команда запуска приложения | обязательный |
| `restart_on_secret_changes` | Перезапуск при изменении секретов: `never`, `always` | `always` |
| `restart_stop_signal` | Сигнал для остановки процесса | `SIGTERM` |

### Важные параметры блока `env_template`

| Параметр | Описание | Пример |
| --- | --- | --- |
| `contents` | Встроенный шаблон переменной окружения | `<<-EOT ... EOT` |
| `source` | Путь к файлу-шаблону | `"/etc/app/env.ctmpl"` |
| `error_on_missing_key` | Ошибка при отсутствии ключа | `true` / `false` |

> Примечание  
> Блок `env_template` всегда имеет имя переменной окружения: `env_template "MY_VAR" { ... }`.  
> Поля `destination`, `perms`, `command`, `wait` и похожие параметры в `env_template` не поддерживаются.

### Управление жизненным циклом процесса

Когда секреты меняются, например при ротации database credentials, Agent:

1. получает новые секреты;
2. формирует новые переменные окружения;
3. отправляет `SIGTERM` дочернему процессу;
4. перезапускает процесс с обновлёнными значениями.

## Кеширование и ротация токенов

Stronghold Agent поддерживает кеширование токенов и автоматическое обновление.

### Token caching

Кеширование помогает:

- сохранить токен после аутентификации;
- использовать один и тот же токен для всех запросов;
- снизить нагрузку на сервер Deckhouse Stronghold.

### Token renewal

Если Agent получает токен с ограниченным TTL, он продлевает его заранее. Если продление невозможно, Agent проходит аутентификацию повторно.

Это нужно, чтобы приложение работало непрерывно и не требовало ручного вмешательства.

### Lease renewal

Динамические секреты тоже имеют срок действия. Agent продлевает их заранее, затем обновляет файлы конфигурации и может перезагрузить приложение.

Так секреты остаются актуальными, а приложение не теряет доступ из-за истёкших credentials.

## API Proxy

Stronghold Agent может работать как прокси для API Deckhouse Stronghold.

### Что это даёт

- локальный HTTP(S)-эндпоинт для приложений;
- автоматическое добавление токена аутентификации;
- кеширование ответов;
- снижение сетевой нагрузки.

### Пример конфигурации

```hcl
api_proxy {
  use_auto_auth_token = true
}
listener "tcp" {
  address = "127.0.0.1:8200"
  tls_disable = true
}
```

### Как это выглядит для приложения

```bash
curl http://127.0.0.1:8200/v1/secret/data/myapp
```

В этом сценарии приложение обращается к локальному Agent, а Agent сам добавляет токен и проксирует запрос на сервер Deckhouse Stronghold.

## Auto-Auth

Auto-Auth — одна из ключевых возможностей Stronghold Agent. Она автоматизирует получение и обновление токена аутентификации.

### Как это работает

1. Agent запускается с настроенным методом аутентификации.
2. Сам проходит аутентификацию в Deckhouse Stronghold.
3. Получает токен и использует его для templating, API Proxy и других задач.
4. Если настроен sink, записывает токен в файл.
5. Продлевает токен до истечения TTL.
6. При необходимости проходит аутентификацию заново.

### Sink

Sink — это место, куда Agent записывает полученный токен.

Если sink настроен, токен записывается в файл, например `/var/run/stronghold-agent/token`, и его могут использовать другие процессы.

Если sink не настроен, токен используется только самим Agent.

### Поддерживаемые методы аутентификации

- AppRole — рекомендуемый вариант для VM и bare metal;
- Token — для простых сценариев;
- JWT/OIDC — для интеграции с identity provider;
- облачные провайдеры.

## AppRole

AppRole — рекомендуемый метод аутентификации для машин и приложений на VM и bare metal.

### Как устроен AppRole

- `Role ID` — идентификатор роли, аналог имени пользователя;
- `Secret ID` — секретный идентификатор, аналог пароля;
- для аутентификации нужны оба значения.

### Преимущества

- разделение обязанностей;
- гибкая настройка политик;
- поддержка ограничений по CIDR;
- возможность использовать одноразовые `Secret ID`.

### Настройка на стороне Deckhouse Stronghold

```bash
stronghold auth enable approle

stronghold write auth/approle/role/myapp \
  token_ttl=1h \
  token_max_ttl=4h \
  policies="myapp-policy"

stronghold read auth/approle/role/myapp/role-id

stronghold write -f auth/approle/role/myapp/secret-id
```

### Конфигурация Agent

```hcl
auto_auth {
  method "approle" {
    mount_path = "auth/approle"
    config = {
      role_id_file_path = "/etc/stronghold-agent/role-id"
      secret_id_file_path = "/etc/stronghold-agent/secret-id"
      remove_secret_id_file_after_reading = true
    }
  }
}
```

### Хранение и доставка credentials

#### Role ID

- это публичный идентификатор роли;
- его можно доставлять через Ansible, Puppet, образ VM или вручную;
- обычно хранится в `/etc/stronghold-agent/role-id`;
- после использования не удаляется;
- сам по себе не считается критичным секретом.

#### Secret ID

- это чувствительный секрет;
- его лучше доставлять через защищённый канал;
- обычно хранится в `/etc/stronghold-agent/secret-id`;
- при необходимости может быть удалён после чтения;
- его нельзя хранить в Git или configuration management без дополнительной защиты.

### Типы Secret ID

#### Одноразовый

```bash
stronghold write auth/approle/role/myapp \
  secret_id_num_uses=1 \
  policies="myapp-policy"
```

или:

```bash
stronghold write -f auth/approle/role/myapp/secret-id num_uses=1
```

Особенности:

- используется только один раз;
- после использования становится невалидным;
- это самый безопасный вариант для production-окружения.

#### Многоразовый

```bash
stronghold write auth/approle/role/myapp \
  secret_id_num_uses=0 \
  policies="myapp-policy"
```

Особенности:

- может использоваться многократно;
- подходит для разработки и тестирования;
- требует ручной ротации при компрометации.

#### С ограниченным TTL

```bash
stronghold write auth/approle/role/myapp \
  secret_id_ttl=24h \
  policies="myapp-policy"
```

Особенности:

- истекает через заданное время;
- даёт баланс между безопасностью и удобством;
- после истечения требует новый `Secret ID`.

### Полный пример настройки и доставки AppRole

```bash
stronghold auth enable approle

stronghold policy write myapp-policy - <<EOF
path "secret/data/myapp/*" {
  capabilities = ["read"]
}
path "database/creds/myapp" {
  capabilities = ["read"]
}
EOF

stronghold write auth/approle/role/myapp \
  token_ttl=1h \
  token_max_ttl=4h \
  policies="myapp-policy" \
  secret_id_num_uses=1 \
  secret_id_ttl=24h

stronghold read auth/approle/role/myapp/role-id
stronghold write -f auth/approle/role/myapp/secret-id
```

Доставка на целевой сервер:

```bash
ssh root@app-server.example.com << 'ENDSSH'
  mkdir -p /etc/stronghold-agent
  chown root:stronghold-agent /etc/stronghold-agent
  chmod 750 /etc/stronghold-agent
ENDSSH
```

```bash
ssh root@app-server.example.com << 'ENDSSH'
  echo -n "abc123-def456-ghi789" > /etc/stronghold-agent/role-id
  chown stronghold-agent:stronghold-agent /etc/stronghold-agent/role-id
  chmod 0640 /etc/stronghold-agent/role-id
ENDSSH
```

```bash
ssh root@app-server.example.com << 'ENDSSH'
  echo -n "xyz789-abc123-def456" > /etc/stronghold-agent/secret-id
  chown stronghold-agent:stronghold-agent /etc/stronghold-agent/secret-id
  chmod 0640 /etc/stronghold-agent/secret-id
ENDSSH
```

Конфигурация Agent:

```bash
cat > /etc/stronghold-agent/agent.hcl <<EOF
stronghold {
  address = "https://stronghold.example.com:8200"
}
auto_auth {
  method "approle" {
    mount_path = "auth/approle"
    config = {
      role_id_file_path = "/etc/stronghold-agent/role-id"
      secret_id_file_path = "/etc/stronghold-agent/secret-id"
      remove_secret_id_file_after_reading = true
    }
  }

  sink "file" {
    config = {
      path = "/var/run/stronghold-agent/token"
      mode = 0640
    }
  }
}
EOF
```

Права на конфигурацию:

```bash
chown root:stronghold-agent /etc/stronghold-agent/agent.hcl
chmod 0640 /etc/stronghold-agent/agent.hcl
```

Запуск и проверка:

```bash
systemctl start stronghold-agent
journalctl -u stronghold-agent -n 50 | grep -i "authentication successful"
ls -la /var/run/stronghold-agent/token
ls -la /etc/stronghold-agent/secret-id
```

### Практические рекомендации по AppRole

- используйте одноразовые `Secret ID` для production-окружения;
- доставляйте `Role ID` и `Secret ID` разными каналами;
- ограничивайте доступ по CIDR;
- логируйте использование `Secret ID` для аудита.

## Token

Прямое использование токена — самый простой способ аутентификации. Agent читает готовый токен из файла и использует его.

### Когда использовать

Этот вариант подходит для:

- тестовых окружений;
- временных установок;
- сценариев, где нельзя использовать AppRole;
- простых случаев без повышенных требований к безопасности.

### Ограничения

- токен — это долгоживущий credential;
- нет разделения обязанностей, как в AppRole;
- при компрометации нужна ручная ротация;
- для production-окружения этот вариант не рекомендуется.

### Полный пример

Создание токена:

```bash
stronghold policy write myapp-policy - <<EOF
path "secret/data/myapp/*" {
  capabilities = ["read"]
}
path "database/creds/myapp" {
  capabilities = ["read"]
}
EOF

stronghold token create \
  -policy=myapp-policy \
  -ttl=720h \
  -renewable=true \
  -display-name="myapp-agent" \
  -format=json
```

Доставка токена:

```bash
ssh root@app-server.example.com << 'ENDSSH'
  mkdir -p /etc/stronghold-agent
  chown root:stronghold-agent /etc/stronghold-agent
  chmod 750 /etc/stronghold-agent
ENDSSH
```

```bash
echo -n "$AGENT_TOKEN" | ssh root@app-server.example.com 'cat > /etc/stronghold-agent/token'
```

```bash
ssh root@app-server.example.com << 'ENDSSH'
  chown stronghold-agent:stronghold-agent /etc/stronghold-agent/token
  chmod 0640 /etc/stronghold-agent/token
ENDSSH
```

Конфигурация Agent:

```bash
cat > /etc/stronghold-agent/agent.hcl <<EOF
stronghold {
  address = "https://stronghold.example.com:8200"
}
auto_auth {
  method "token_file" {
    config = {
      token_file_path = "/etc/stronghold-agent/token"
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
EOF
```

Проверка:

```bash
systemctl start stronghold-agent
journalctl -u stronghold-agent -n 50
```

### Важные параметры токена

- `ttl` — начальное время жизни токена;
- `renewable` — можно ли его обновлять;
- `period` — период обновления;
- `explicit-max-ttl` — абсолютное максимальное время жизни.

### Практические рекомендации по token-аутентификации

- используйте `renewable=true`;
- задавайте разумный TTL;
- ограничивайте общий срок жизни через `explicit-max-ttl`;
- регулярно отзывайте неиспользуемые токены;
- храните токен с минимальными правами доступа;
- для production по возможности переходите на AppRole.

## JWT/OIDC

JWT/OIDC-аутентификация позволяет использовать существующую систему identity management для входа в Deckhouse Stronghold.

### Как это работает

1. Приложение получает JWT от identity provider.
2. JWT содержит claims о пользователе или сервисе.
3. Deckhouse Stronghold проверяет подпись и извлекает claims.
4. На основе claims выдаёт собственный токен Stronghold.

### Когда использовать

Этот вариант подходит для:

- интеграции с корпоративным SSO;
- использования service account из identity provider;
- федеративной аутентификации;
- CI/CD через OIDC, например GitHub Actions или GitLab CI.

### Преимущества

- централизованное управление идентичностью;
- не нужно создавать отдельные credentials для каждого приложения;
- JWT ротируется на стороне identity provider;
- можно использовать MFA и другие возможности identity provider.

### Полный пример с Keycloak

Настройка метода на стороне Deckhouse Stronghold:

```bash
stronghold auth enable jwt

stronghold write auth/jwt/config \
  oidc_discovery_url="https://keycloak.example.com/realms/myrealm" \
  oidc_client_id="stronghold" \
  oidc_client_secret="client-secret-from-keycloak" \
  default_role="default"
```

Создание политики:

```bash
stronghold policy write myapp-jwt-policy - <<EOF
path "secret/data/myapp/*" {
  capabilities = ["read"]
}
path "database/creds/myapp" {
  capabilities = ["read"]
}
EOF
```

Создание роли:

```bash
stronghold write auth/jwt/role/myapp-role \
  role_type="jwt" \
  bound_audiences="stronghold" \
  user_claim="sub" \
  bound_subject="service-account-myapp" \
  token_ttl=1h \
  token_max_ttl=4h \
  token_policies="myapp-jwt-policy"
```

Вариант с дополнительными claims:

```bash
stronghold write auth/jwt/role/myapp-role \
  role_type="jwt" \
  bound_audiences="stronghold" \
  user_claim="sub" \
  bound_claims='{"environment":"production","app":"myapp"}' \
  claim_mappings='{"department":"dept"}' \
  token_policies="myapp-jwt-policy"
```

Получение JWT от identity provider:

```bash
curl -X POST "https://keycloak.example.com/realms/myrealm/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=myapp-service" \
  -d "client_secret=service-secret" \
  -d "grant_type=client_credentials" \
  | jq -r '.access_token' > /tmp/jwt-token.txt
```

Пример для GitHub Actions описан в черновике как получение OIDC-токена через переменные среды Actions.

Доставка JWT на сервер:

```bash
scp /tmp/jwt-token.txt root@app-server.example.com:/etc/stronghold-agent/jwt-token
```

```bash
ssh root@app-server.example.com << 'ENDSSH'
  chown stronghold-agent:stronghold-agent /etc/stronghold-agent/jwt-token
  chmod 0640 /etc/stronghold-agent/jwt-token
ENDSSH
```

Конфигурация Agent:

```bash
cat > /etc/stronghold-agent/agent.hcl <<EOF
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
EOF
```

Проверка:

```bash
systemctl start stronghold-agent
journalctl -u stronghold-agent -n 50 | grep -i "authentication successful"
cat /var/run/stronghold-agent/token
```

### Особенности JWT-метода

- JWT-токен и токен Stronghold — это разные токены;
- JWT обычно живёт 5–60 минут;
- после входа Agent получает токен Stronghold;
- токен Stronghold обновляется автоматически;
- сам JWT Agent автоматически не обновляет.

Если JWT истекает, нужно получить новый токен от identity provider. В черновике для этого предложен вариант с cron.

### Проверка JWT

```bash
cat /etc/stronghold-agent/jwt-token | cut -d. -f2 | base64 -d | jq
```

### Практические рекомендации по JWT/OIDC

- используйте короткий TTL для JWT;
- настраивайте `bound_audiences`;
- используйте `bound_subject` или `bound_claims` для более строгой проверки;
- для production используйте OIDC discovery;
- логируйте аутентификацию для аудита.
