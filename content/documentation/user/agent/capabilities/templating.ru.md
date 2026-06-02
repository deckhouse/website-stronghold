---
title: "Шаблоны и рендеринг в файлы"
linkTitle: "Шаблоны"
description: "Использование Stronghold Agent для рендеринга секретов в файлы"
weight: 20
---

Stronghold Agent может рендерить секреты из Stronghold в файлы конфигурации.
Этот режим подходит для приложений, которые читают настройки из файлов, а не из переменных окружения.

На этой странице описано, как использовать режим `template`,
какие возможности он поддерживает
и в каких случаях его стоит выбирать.

## Когда использовать режим template

Используйте режим `template` в следующих случаях:

- Приложение читает конфигурацию из файлов.
- Нужно передавать секреты в файлы `.conf`, `.ini`, `.yaml` или `.properties`.
- Нужно сохранять сертификаты, ключи или другие чувствительные данные в файлах.
- Приложение не поддерживает прямую работу с API Stronghold.
- Используются динамические credentials, которые нужно обновлять без ручного вмешательства.

Если приложение читает секреты из переменных окружения,
используйте режим `env_template` вместе с `exec`.
Подробности см. на странице [«Переменные окружения и Process Supervisor»](../process-supervisor/).

## Как работает рендеринг шаблонов

Обычно Stronghold Agent выполняет следующие действия:

1. Читает файл шаблона.
1. Запрашивает секреты из Stronghold.
1. Подставляет значения в шаблон.
1. Записывает результат в целевой файл.
1. При необходимости запускает команду для перезагрузки сервиса.

Если значение секрета меняется,
Agent повторно рендерит файл.
При необходимости после этого он может снова запустить команду,
указанную в параметре `command`.

## Синтаксис шаблонов

Stronghold Agent использует шаблоны в стиле Consul Template.
Шаблон обычно запрашивает секрет по пути и извлекает из него нужные поля.

Базовая структура шаблона:

```go
{{ with secret "path/to/secret" }}
{{ .Data.field_name }}
{{ end }}
```

Пример для KV v2:

```go
{{ with secret "secret/data/myapp" }}
username={{ .Data.data.username }}
password={{ .Data.data.password }}
{{ end }}
```

Пример для динамического секрета базы данных:

```go
{{ with secret "database/creds/myapp" }}
DB_USER={{ .Data.username }}
DB_PASS={{ .Data.password }}
{{ end }}
```

## Основные функции шаблонов

В шаблонах можно использовать следующие функции:

| Функция | Назначение | Пример |
| --- | --- | --- |
| `secret` | Получает секрет | `{{ with secret "secret/data/myapp" }}{{ .Data.data.password }}{{ end }}` |
| `base64Encode` | Кодирует строку в Base64 | `{{ "password" \| base64Encode }}` |
| `base64Decode` | Декодирует строку из Base64 | `{{ .Data.cert \| base64Decode }}` |
| `toJSON` | Преобразует значение в JSON | `{{ .Data \| toJSON }}` |
| `toYAML` | Преобразует значение в YAML | `{{ .Data \| toYAML }}` |
| `toLower` | Преобразует строку в нижний регистр | `{{ .Data.name \| toLower }}` |
| `toUpper` | Преобразует строку в верхний регистр | `{{ .Data.name \| toUpper }}` |
| `trim` | Удаляет пробелы по краям строки | `{{ .Data.value \| trim }}` |
| `range` | Перебирает элементы списка | `{{ range .Items }}{{ .Name }}{{ end }}` |
| `env` | Читает переменную окружения | `{{ env "HOME" }}` |
| `timestamp` | Возвращает текущее время | `{{ timestamp "2006-01-02 15:04:05" }}` |

## Пример рендеринга файла конфигурации

В этом примере Stronghold Agent создаёт файл `application.properties`
для Java-приложения.

### Шаг 1. Сохраните секреты в Stronghold

```shell
stronghold kv put secret/myapp/config \
  db_host=postgres.prod.example.com \
  db_port=5432 \
  db_name=production \
  db_user=app_user \
  db_password=SecureP@ssw0rd
```

### Шаг 2. Создайте файл шаблона

Создайте файл `/etc/myapp/templates/application.properties.ctmpl`:

```text
{{ with secret "secret/data/myapp/config" }}
spring.datasource.url=jdbc:postgresql://{{ .Data.data.db_host }}:{{ .Data.data.db_port }}/{{ .Data.data.db_name }}
spring.datasource.username={{ .Data.data.db_user }}
spring.datasource.password={{ .Data.data.db_password }}
{{ end }}
spring.datasource.hikari.maximum-pool-size=10
spring.datasource.hikari.minimum-idle=5
```

### Шаг 3. Настройте Stronghold Agent

Создайте файл `/etc/stronghold-agent/agent.hcl`:

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
  source = "/etc/myapp/templates/application.properties.ctmpl"
  destination = "/etc/myapp/application.properties"
  perms = "0600"
  user = "myapp"
  group = "myapp"
  command = "systemctl reload myapp"
  command_timeout = "30s"

  wait {
    min = "2s"
    max = "10s"
  }

  error_on_missing_key = true
}
```

### Шаг 4. Проверьте конфигурацию

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl -exit-after-auth -log-level=debug
```

Во время выполнения Agent:

1. Читает файл `agent.hcl`.
1. Подключается к серверу Stronghold.
1. Проходит аутентификацию через AppRole.
1. Получает токен и записывает его в sink.
1. Запрашивает секреты.
1. Рендерит шаблон и создаёт файл `/etc/myapp/application.properties`.
1. Завершает работу с кодом `0`.

### Проверьте результат

Выполните следующие команды:

```shell
ls -la /var/run/stronghold-agent/token
ls -la /etc/myapp/application.properties
sudo cat /etc/myapp/application.properties
```

После успешной проверки можно запустить Agent как сервис systemd:

```shell
systemctl start stronghold-agent
systemctl status stronghold-agent
journalctl -u stronghold-agent -f
```

## Расширенные сценарии

### Динамические credentials для базы данных

Шаблон может получать временные учётные данные базы данных
и записывать их в файл:

```go
{{ with secret "database/creds/myapp-role" }}
DB_USER={{ .Data.username }}
DB_PASS={{ .Data.password }}
DB_LEASE_ID={{ .LeaseID }}
DB_LEASE_DURATION={{ .LeaseDuration }}
{{ end }}
```

В этом случае Agent будет повторно рендерить файл
при обновлении секрета.
При необходимости он также сможет перезапускать или перезагружать сервис.

### Рендеринг сертификатов PKI

Шаблон сертификата:

```go
{{ with secret "pki/issue/web-server" "common_name=app.example.com" "ttl=720h" }}
{{ .Data.certificate }}
{{ .Data.ca_chain }}
{{ end }}
```

Шаблон приватного ключа:

```go
{{ with secret "pki/issue/web-server" "common_name=app.example.com" "ttl=720h" }}
{{ .Data.private_key }}
{{ end }}
```

Пример конфигурации:

```hcl
template {
  source = "/etc/nginx/ssl/cert.pem.ctmpl"
  destination = "/etc/nginx/ssl/cert.pem"
  perms = "0644"
}

template {
  source = "/etc/nginx/ssl/key.pem.ctmpl"
  destination = "/etc/nginx/ssl/key.pem"
  perms = "0600"
  command = "systemctl reload nginx"
}
```

### Условная логика

В шаблонах можно использовать условные выражения:

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

### Циклы

В шаблонах можно перебирать списки:

```go
{{ with secret "secret/data/myapp/allowed-ips" }}
{{ range $index, $ip := .Data.data.ips }}
allow {{ $ip }};
{{ end }}
{{ end }}
```

### Несколько секретов в одном файле

Один шаблон может использовать несколько секретов:

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

## Параметры блока template

Для блока `template` доступны следующие основные параметры:

| Параметр | Описание | Пример |
| --- | --- | --- |
| `source` | Путь к файлу шаблона | `/etc/app/template.ctmpl` |
| `destination` | Путь к итоговому файлу | `/etc/app/config.conf` |
| `perms` | Права доступа к файлу | `"0600"` |
| `user` | Владелец файла | `"myapp"` |
| `group` | Группа файла | `"myapp"` |
| `command` | Команда после рендеринга | `"systemctl reload app"` |
| `command_timeout` | Таймаут выполнения команды | `"30s"` |
| `error_on_missing_key` | Завершать с ошибкой при отсутствии ключа | `true` |
| `wait.min` | Минимальная задержка перед обновлением | `"2s"` |
| `wait.max` | Максимальная задержка перед обновлением | `"10s"` |
| `backup` | Создавать резервную копию файла | `true` |

## Ограничения и особенности

Учитывайте следующие особенности режима `template`:

- Секреты записываются на диск в виде итогового файла.
- Приложение должно иметь доступ к файлу назначения.
- Если Agent работает под systemd с `ProtectSystem=strict`,
  добавьте директорию из параметра `destination` в `ReadWritePaths`.
- Если шаблон использует отсутствующий ключ и включён параметр `error_on_missing_key`,
  рендеринг завершится с ошибкой.
