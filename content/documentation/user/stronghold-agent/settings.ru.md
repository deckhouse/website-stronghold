---
title: "Основные настройки"
linkTitle: "Настройка"
description: "Основные настройки Stronghold Agent"
weight: 40
---

Конфигурация Stronghold Agent описывается в формате HCL. Через конфигурационный файл вы настраиваете подключение к серверу Deckhouse Stronghold, автоматическую аутентификацию, рендеринг шаблонов, запуск дочерних процессов, API Proxy и параметры логирования.

Эта страница помогает понять структуру конфигурационного файла и назначение основных секций.

## Структура конфигурационного файла

Ниже показан пример общей структуры конфигурации Stronghold Agent:

```hcl
stronghold {
  address = "https://stronghold.example.com:8200"
  ca_cert = "/etc/stronghold-agent/ca.pem"

  retry {
    num_retries = 5
  }
}

auto_auth {
  method "approle" {
    # ... конфигурация метода.
  }

  sink "file" {
    # ... конфигурация sink.
  }
}

api_proxy {
  use_auto_auth_token = true
}

cache {
  use_auto_auth_token = true
}

listener "tcp" {
  address = "127.0.0.1:8200"
  tls_disable = true
}

template {
  source      = "/path/to/template.ctmpl"
  destination = "/path/to/output"
}

pid_file = "/var/run/stronghold-agent.pid"

log_level = "info"
log_file = "/var/log/stronghold-agent.log"
```

Обычно в конфигурации используются такие секции:

- `stronghold` — подключение к серверу Deckhouse Stronghold;
- `auto_auth` — автоматическая аутентификация и sink для токена;
- `api_proxy` — локальный прокси для API Stronghold;
- `cache` — кеширование;
- `listener` — локальный listener для API Proxy;
- `template` — рендеринг шаблонов;
- `template_config` — глобальные параметры шаблонов;
- `exec` — запуск дочернего процесса;
- `env_template` — передача секретов через переменные окружения;
- параметры `pid_file`, `log_level`, `log_file` и другие — управление процессом и логированием.

## Секция `stronghold`

Для подключения к серверу Deckhouse Stronghold используется секция `stronghold`.

Если нужна интеграция с HashiCorp Vault, вместо `stronghold` в имени секции используется `vault`.

Пример:

```hcl
stronghold {
  address = "https://stronghold.example.com:8200"

  ca_cert = "/etc/stronghold-agent/ca.pem"
  ca_path = "/etc/stronghold-agent/ca-bundle/"
  client_cert = "/etc/stronghold-agent/client.pem"
  client_key = "/etc/stronghold-agent/client-key.pem"
  tls_skip_verify = false
  tls_server_name = "stronghold.example.com"

  retry {
    num_retries = 5
  }
}
```

Что задаётся в этой секции:

- `address` — адрес сервера Stronghold;
- `ca_cert` — путь к CA-сертификату;
- `ca_path` — директория с набором CA-сертификатов;
- `client_cert` и `client_key` — клиентский сертификат и ключ;
- `tls_skip_verify` — отключение проверки TLS;
- `tls_server_name` — имя сервера для SNI;
- `retry` — политика повторных попыток.

> Важно  
> Не отключайте `tls_skip_verify` в production-окружении.

## Секция `auto_auth`

Секция `auto_auth` отвечает за автоматическую аутентификацию Stronghold Agent и за сохранение полученного токена.

Пример:

```hcl
auto_auth {
  method "approle" {
    mount_path = "auth/approle"
    namespace  = "myns"

    config = {
      role_id_file_path = "/etc/stronghold-agent/role-id"
      secret_id_file_path = "/etc/stronghold-agent/secret-id"
    }
  }

  sink "file" {
    config = {
      path = "/var/run/stronghold-agent/token"
      mode = 0640
    }
  }

  sink "file" {
    wrap_ttl = "5m"
    aad_env_var = "VAULT_AAD"
    dh_type = "curve25519"
    dh_path = "/etc/stronghold-agent/dh-pub"

    config = {
      path = "/var/run/stronghold-agent/encrypted-token"
    }
  }
}
```

В этой секции настраиваются:

- `method` — метод аутентификации;
- `mount_path` — путь монтирования auth method;
- `namespace` — namespace, если он используется;
- `config` — параметры конкретного метода;
- `sink` — место, куда Agent сохраняет токен.

Sink может быть обычным или с дополнительным шифрованием.

## Секция `template`

Секция `template` используется для рендеринга шаблонов в файлы.

Пример:

```hcl
template {
  source = "/etc/myapp/config.ctmpl"
  destination = "/etc/myapp/config.conf"
  perms = "0600"
  user = "myapp"
  group = "myapp"
  backup = true
  command = "systemctl reload myapp"
  command_timeout = "30s"

  wait {
    min = "5s"
    max = "10s"
  }

  error_on_missing_key = true
  create_dest_dirs = true
}
```

Что можно настроить:

- `source` — путь к файлу шаблона;
- `destination` — путь к итоговому файлу;
- `perms` — права доступа;
- `user` и `group` — владелец и группа;
- `backup` — создавать ли резервную копию;
- `command` — команда после рендеринга;
- `command_timeout` — таймаут команды;
- `wait` — задержка перед рендерингом;
- `error_on_missing_key` — завершать ли рендеринг ошибкой, если ключ не найден;
- `create_dest_dirs` — создавать ли отсутствующие директории назначения.

## Секция `template_config`

Секция `template_config` задаёт глобальные параметры для всех шаблонов.

Пример:

```hcl
template_config {
  exit_on_retry_failure = false
  static_secret_render_interval = "5m"
}
```

Здесь можно настроить:

- `exit_on_retry_failure` — завершать ли работу после неудачных повторных попыток;
- `static_secret_render_interval` — интервал периодического рендеринга для статических секретов, например из KV.

## Секция `exec`

Секция `exec` используется в режиме Process Supervisor, когда Stronghold Agent запускает дочерний процесс и передаёт ему секреты через переменные окружения.

Пример:

```hcl
exec {
  command = ["/usr/bin/myapp", "--config", "/etc/myapp/config.yaml"]
  restart_on_secret_changes = "always"
  restart_stop_signal = "SIGTERM"
}
```

Здесь задаются:

- `command` — команда запуска приложения;
- `restart_on_secret_changes` — перезапускать ли процесс при изменении секретов;
- `restart_stop_signal` — сигнал для остановки процесса.

### `env_template`

Вместе с `exec` обычно используют `env_template`:

```hcl
env_template "DATABASE_URL" {
  contents = "{{ with secret \"secret/data/myapp\" }}postgresql://{{ .Data.data.username }}:{{ .Data.data.password }}@db:5432{{ end }}"
  error_on_missing_key = true
}

env_template "API_KEY" {
  contents = "{{ with secret \"secret/data/myapp\" }}{{ .Data.data.api_key }}{{ end }}"
  error_on_missing_key = true
}
```

Каждый блок `env_template` формирует значение одной переменной окружения.

## Секция `listener`

Секция `listener` настраивает HTTP(S)-listener для API Proxy.

Пример TCP-listener:

```hcl
listener "tcp" {
  address = "127.0.0.1:8200"
  tls_disable = true

  tls_cert_file = "/etc/stronghold-agent/agent-cert.pem"
  tls_key_file = "/etc/stronghold-agent/agent-key.pem"

  require_request_header = true

  agent_api {
    enable_quit = true
  }
}
```

Пример Unix socket listener:

```hcl
listener "unix" {
  address = "/var/run/stronghold-agent.sock"
  tls_disable = true
  socket_mode = "0660"
  socket_user = "myapp"
  socket_group = "myapp"
}
```

Что можно настроить:

- `address` — адрес TCP-listener или путь к Unix socket;
- `tls_disable` — отключение TLS;
- `tls_cert_file` и `tls_key_file` — TLS-сертификат и ключ;
- `require_request_header` — обязательный специальный заголовок;
- `agent_api.enable_quit` — включение endpoint `/agent/v1/quit`;
- `socket_mode`, `socket_user`, `socket_group` — параметры Unix socket.

## Логирование и отладка

Stronghold Agent поддерживает настройку уровня логирования, формата логов и ротации файлов логов.

Пример:

```hcl
log_level = "info"
log_file = "/var/log/stronghold-agent.log"
log_format = "json"
log_rotate_duration = "24h"
log_rotate_bytes = 104857600
log_rotate_max_files = 10
```

Что можно настроить:

- `log_level` — уровень логирования: `trace`, `debug`, `info`, `warn`, `error`;
- `log_file` — путь к JSON-файлу логов;
- `log_format` — формат логов: `standard` или `json`;
- `log_rotate_duration` — период ротации;
- `log_rotate_bytes` — максимальный размер файла;
- `log_rotate_max_files` — количество сохраняемых файлов.

## Практические рекомендации

Чтобы упростить настройку Stronghold Agent:

- сначала определите, какой режим вам нужен: `template` или `exec` с `env_template`;
- для production-окружения используйте TLS и не отключайте проверку сертификатов;
- храните конфигурационный файл, токены, `role-id` и `secret-id` с минимально необходимыми правами доступа;
- сначала проверьте базовое подключение и `auto_auth`, а затем переходите к шаблонам и Process Supervisor;
- если используете `template`, заранее проверьте права на запись в директорию назначения;
- если используете `listener`, заранее ограничьте доступ к локальному listener или Unix socket.

## Что дальше

- Чтобы понять, какие режимы работы поддерживает Agent, откройте [Возможности](../capabilities/).
- Для запуска и сопровождения Agent используйте [Запуск и управление](./launch_and_control/).
