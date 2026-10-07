---
title: "Приложение на виртуальной машине со Stronghold Agent"
linkTitle: "Приложение на ВМ"
description: "Доставка секретов в приложение на виртуальной машине: Stronghold Agent под systemd, аутентификация AppRole с обернутым secret_id, рендеринг шаблонов и перезагрузка приложения."
weight: 20
params:
  relatedLinks:
    - title: "Stronghold Agent"
      url: ../../../user/agent/overview/
    - title: "Настройки Agent"
      url: ../../../user/agent/settings/
    - title: "Запуск и управление Agent"
      url: ../../../user/agent/launch-and-control/
    - title: "Метод AppRole"
      url: ../../../user/auth/approle/
    - title: "Обертывание ответа"
      url: ../../../concepts/response-wrapping/
---

Приложение, которое читает пароли из конфигурационного файла и не умеет обращаться к API Stronghold, можно подключить без изменения кода. Stronghold Agent аутентифицируется в Stronghold, рендерит конфигурационный файл по шаблону и перезагружает приложение при изменении секретов.

## Цель

Настроить на виртуальной машине Stronghold Agent под управлением systemd, который входит по AppRole с одноразовым обернутым `secret_id`, формирует файл `application.properties` и выполняет `systemctl reload myapp` при изменении секретов.

## Предварительные требования

- Виртуальная машина Linux с systemd и приложением `myapp`, запущенным как сервис `myapp.service`.
- Исполняемый файл Stronghold на ВМ (в примерах `/usr/local/bin/stronghold`) и сетевой доступ к Stronghold.
- Токен Stronghold с правами на настройку AppRole, политик и KV.
- KV версии 2, включённое по пути `secret`.

## Шаг 1. Сохраните секрет и создайте политику

```bash
d8 stronghold kv put -mount=secret myapp/config \
  db_user=app_user \
  db_password='S3cure-P@ss'

d8 stronghold policy write myapp-vm - <<'POLICY'
path "secret/data/myapp/config" {
  capabilities = ["read"]
}
POLICY
```

## Шаг 2. Создайте роль AppRole

```bash
d8 stronghold auth enable approle

d8 stronghold write auth/approle/role/myapp-vm \
  token_policies=myapp-vm \
  token_ttl=1h \
  token_max_ttl=24h \
  secret_id_ttl=720h \
  secret_id_bound_cidrs="10.0.10.15/32"
```

Параметр `secret_id_bound_cidrs` разрешает вход только с адреса ВМ. `secret_id_ttl` ограничивает срок, в течение которого Agent может повторно входить с одним `secret_id`.

## Шаг 3. Доставьте role_id и обернутый secret_id на ВМ

1. Получите `role_id` — он не является секретом и может распространяться системой управления конфигурацией:

   ```bash
   d8 stronghold read -field=role_id auth/approle/role/myapp-vm/role-id > role-id
   ```

1. Выпустите `secret_id` в обернутом виде. Вместо самого `secret_id` возвращается одноразовый токен с TTL 10 минут:

   ```bash
   d8 stronghold write -wrap-ttl=10m -field=wrapping_token -f \
     auth/approle/role/myapp-vm/secret-id > secret-id
   ```

1. Скопируйте оба файла на ВМ в каталог `/etc/stronghold-agent/` и ограничьте доступ:

   ```bash
   sudo install -o stronghold-agent -g stronghold-agent -m 0640 role-id /etc/stronghold-agent/role-id
   sudo install -o stronghold-agent -g stronghold-agent -m 0600 secret-id /etc/stronghold-agent/secret-id
   ```

Если оборачивающий токен перехвачен и развёрнут кем-то другим, Agent не сможет его развернуть — это признак инцидента. Порядок проверки описан в разделе [«Обертывание ответа»](../../../concepts/response-wrapping/#проверка-токена-обертывания).

## Шаг 4. Создайте шаблон

Создайте файл `/etc/myapp/templates/application.properties.ctmpl`:

```text
{{ with secret "secret/data/myapp/config" }}
spring.datasource.username={{ .Data.data.db_user }}
spring.datasource.password={{ .Data.data.db_password }}
{{ end }}
```

## Шаг 5. Настройте Agent

Создайте файл `/etc/stronghold-agent/agent.hcl`:

```hcl
stronghold {
  address = "https://stronghold.example.com"
}

auto_auth {
  method "approle" {
    mount_path = "auth/approle"
    config = {
      role_id_file_path                   = "/etc/stronghold-agent/role-id"
      secret_id_file_path                 = "/etc/stronghold-agent/secret-id"
      secret_id_response_wrapping_path    = "auth/approle/role/myapp-vm/secret-id"
      remove_secret_id_file_after_reading = true
    }
  }

  sink "file" {
    config = {
      path = "/var/run/stronghold-agent/token"
      mode = 0600
    }
  }
}

template_config {
  static_secret_render_interval = "5m"
}

template {
  source          = "/etc/myapp/templates/application.properties.ctmpl"
  destination     = "/etc/myapp/application.properties"
  perms           = "0640"
  command         = "systemctl reload myapp"
  command_timeout = "30s"
  error_on_missing_key = true
}
```

- `secret_id_response_wrapping_path` — Agent распознаёт обернутый `secret_id`, проверяет путь его создания и разворачивает его;
- `remove_secret_id_file_after_reading = true` — файл с `secret_id` удаляется после чтения;
- `static_secret_render_interval` — как часто Agent перечитывает статические секреты KV.

Развёрнутый `secret_id` хранится только в памяти Agent. После перезапуска Agent или ВМ, а также по истечении `secret_id_ttl` ему потребуется новый обернутый `secret_id`. Встройте выпуск и доставку `secret_id` в процесс развёртывания, например в CI/CD.

## Шаг 6. Запустите Agent под systemd

1. Создайте unit-файл `/etc/systemd/system/stronghold-agent.service` по примеру из раздела [«Запуск и управление»](../../../user/agent/launch-and-control/). В `ReadWritePaths` укажите каталоги `/var/run/stronghold-agent` и `/etc/myapp`.

1. Разрешите пользователю `stronghold-agent` выполнять только перезагрузку приложения, например через sudoers:

   ```text
   stronghold-agent ALL=(root) NOPASSWD: /usr/bin/systemctl reload myapp
   ```

   В этом случае укажите в шаблоне `command = "sudo /usr/bin/systemctl reload myapp"`.

1. Запустите сервис:

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable --now stronghold-agent
   ```

## Проверка

1. Проверьте журнал Agent — должны быть сообщения об успешной аутентификации и рендеринге:

   ```bash
   sudo journalctl -u stronghold-agent -n 50
   ```

1. Убедитесь, что файл секрета удалён, а конфигурация сформирована:

   ```bash
   ls -la /etc/stronghold-agent/secret-id /etc/myapp/application.properties
   ```

1. Измените секрет и дождитесь перерисовки файла и перезагрузки приложения (не дольше `static_secret_render_interval`):

   ```bash
   d8 stronghold kv patch -mount=secret myapp/config db_password='N3w-P@ss'
   sudo journalctl -u myapp -f
   ```

## Очистка

1. Остановите и отключите Agent: `sudo systemctl disable --now stronghold-agent`.
1. Удалите файлы `/etc/stronghold-agent/role-id`, `/etc/myapp/application.properties` и файл токена в `/var/run/stronghold-agent/`.
1. Удалите роль и политику:

   ```bash
   d8 stronghold delete auth/approle/role/myapp-vm
   d8 stronghold policy delete myapp-vm
   ```
