---
title: "Переменные окружения и Process Supervisor"
linkTitle: "Process Supervisor"
description: "Использование Stronghold Agent для передачи секретов через переменные окружения"
weight: 30
---

Stronghold Agent может передавать секреты приложению через переменные окружения
и управлять жизненным циклом дочернего процесса.
Для этого используются блоки `env_template` и `exec`.

Этот режим подходит для приложений, которые читают конфигурацию из переменных окружения
и не должны получать секреты через файлы на диске.

## Когда использовать этот режим

Используйте режим `env_template` вместе с `exec` в следующих случаях:

- Приложение читает конфигурацию из переменных окружения.
- Секреты не должны записываться на диск в открытом виде.
- Допустим перезапуск приложения при изменении секрета.
- Используются динамические credentials с регулярной ротацией.
- Нужно запускать приложение под управлением Stronghold Agent как дочерний процесс.

Если приложению нужны файлы конфигурации,
используйте режим `template`.
Подробности см. на странице [«Шаблоны и рендеринг в файлы»](../templating/).

## Как работает Process Supervisor

В режиме Process Supervisor Stronghold Agent выполняет следующие действия:

1. Проходит аутентификацию в Stronghold.
1. Запрашивает секреты, указанные в блоках `env_template`.
1. Формирует переменные окружения.
1. Запускает приложение как дочерний процесс.
1. Отслеживает изменение секретов.
1. Перезапускает дочерний процесс при обновлении секрета.

Такой режим позволяет не записывать секреты в файлы
и передавать их приложению только на уровне процесса.

## Блок env_template

Блок `env_template` задаёт значение одной переменной окружения.
Имя переменной указывается в заголовке блока.

Пример:

```hcl
env_template "API_KEY" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.api_key }}{{ end }}"
}
```

Один блок `env_template` формирует только одну переменную окружения.
Если приложению нужно несколько переменных,
создайте отдельный блок для каждой из них.

## Блок exec

Блок `exec` определяет команду,
которую Stronghold Agent запускает как дочерний процесс.

Пример:

```hcl
exec {
  command = ["/usr/bin/java", "-jar", "/opt/myapp/demo-application.jar"]
  restart_on_secret_changes = "always"
  restart_stop_signal = "SIGTERM"
}
```

Обычно блок `exec` используют вместе с одним или несколькими блоками `env_template`.

## Пример для Spring Boot-приложения

В этом примере Stronghold Agent передаёт приложению credentials базы данных
и API-ключ через переменные окружения.

### Шаг 1. Подготовьте приложение

```text
server.port=8080
spring.datasource.url=${DB_URL}
spring.datasource.username=${DB_USERNAME}
spring.datasource.password=${DB_PASSWORD}
api.key=${API_KEY}
```

### Шаг 2. Подготовьте секреты в Stronghold

```shell
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

env_template "SPRING_PROFILES_ACTIVE" {
  contents = "production"
}
```

### Шаг 4. Запустите Agent

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl
```

После запуска Agent:

- проходит аутентификацию;
- получает секреты;
- формирует переменные окружения;
- запускает Java-приложение;
- перезапускает его при изменении секрета.

## Примеры для других приложений

### Go-приложение

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
```

### Docker-контейнер на VM

Если приложение запускается через Docker Engine на виртуальной машине,
передайте переменные окружения в команду `docker run` явно.

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

## Управление жизненным циклом процесса

При изменении секрета Stronghold Agent может перезапустить дочерний процесс.
Обычно последовательность действий выглядит так:

1. Agent получает новое значение секрета.
1. Повторно формирует переменные окружения.
1. Отправляет дочернему процессу сигнал остановки.
1. Запускает процесс заново с обновлёнными значениями.

По умолчанию для остановки обычно используют сигнал `SIGTERM`.

## Ограничения и особенности

Учитывайте следующие особенности этого режима:

- Блок `exec` должен использоваться хотя бы с одним блоком `env_template`.
- Каждый блок `env_template` задаёт только одну переменную окружения.
- Блок `env_template` не создаёт `.env`-файл.
- Параметры `destination`, `perms`, `command` и `wait` для `env_template` не используются.
- `env_template` нельзя комбинировать с `template` и `api_proxy` в одном конфигурационном файле.
- Если приложение запускается через Docker Engine,
  переменные окружения нужно явно передавать через `--env`.

## Параметры блока exec

Для блока `exec` доступны следующие основные параметры:

| Параметр | Описание | Значение по умолчанию |
| --- | --- | --- |
| `command` | Команда запуска приложения | — |
| `restart_on_secret_changes` | Перезапускать процесс при изменении секрета | `always` |
| `restart_stop_signal` | Сигнал для остановки процесса | `SIGTERM` |

## Параметры блока env_template

Для блока `env_template` доступны следующие основные параметры:

| Параметр | Описание | Пример |
| --- | --- | --- |
| `contents` | Встроенный шаблон для переменной окружения | `{{ with secret "secret/data/myapp" }}{{ .Data.data.value }}{{ end }}` |
| `source` | Путь к файлу шаблона | `"/etc/app/env.ctmpl"` |
| `error_on_missing_key` | Завершать с ошибкой при отсутствии ключа | `true` |

{{< alert level="info" >}}
Блок `env_template` всегда содержит имя переменной окружения в заголовке,
например `env_template "MY_VAR" { ... }`.
{{< /alert >}}
