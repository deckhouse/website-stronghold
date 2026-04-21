---
title: "Запуск и управление"
linkTitle: "Запуск и управление"
description: "Запуск, проверка конфигурации и управление Stronghold Agent"
weight: 50
---

Эта страница помогает проверить конфигурацию Stronghold Agent, выполнить пробный запуск и перевести Agent в режим постоянной работы.

Перед запуском в production-окружении рекомендуется сначала проверить конфигурационный файл и убедиться, что Agent может подключиться к серверу Deckhouse Stronghold, пройти аутентификацию и создать нужные файлы.

## Проверка конфигурации

Перед запуском в production-окружении обязательно проверьте корректность конфигурации.

Для этого выполните пробный запуск с автоматическим завершением:

```bash
stronghold -config=/etc/stronghold-agent/agent.hcl -exit-after-auth -log-level=debug
```

Эта команда:

1. проверяет синтаксис HCL-конфигурации;
2. подключается к серверу Deckhouse Stronghold;
3. выполняет полную аутентификацию;
4. создаёт файлы и шаблоны;
5. автоматически завершает работу.

Пример успешного результата:

```text
[INFO]  agent: loaded config: path=/etc/stronghold-agent/agent.hcl
[INFO]  agent.auto_auth.approle: authentication successful
[INFO]  agent.sink.file: writing token to: /var/run/stronghold-agent/token
[INFO]  agent: exit after auth set, exiting
```

Такой запуск удобно использовать как первую проверку перед запуском Agent в фоне или как `systemd`-сервис.

## Запуск в режиме разработки

Для отладки Stronghold Agent можно запускать в foreground-режиме.

Базовый запуск:

```bash
stronghold -config=/etc/stronghold-agent/agent.hcl
```

Запуск с повышенным уровнем логирования:

```bash
stronghold -config=/etc/stronghold-agent/agent.hcl -log-level=debug
```

Запуск с выходом после первой успешной аутентификации:

```bash
stronghold -config=/etc/stronghold-agent/agent.hcl -exit-after-auth
```

Этот режим полезен, если нужно:

- проверить, проходит ли аутентификация;
- посмотреть, как рендерятся шаблоны;
- убедиться, что Agent может записывать токены и файлы;
- быстро найти ошибки в конфигурации.

## Запуск Stronghold Agent как `systemd`-сервиса

Для постоянной работы Stronghold Agent обычно запускают как `systemd`-сервис.

Создайте unit-файл `/etc/systemd/system/stronghold-agent.service`:

```ini
[Unit]
Description=Stronghold Agent
Documentation=https://docs.stronghold.example.com/agent
Requires=network-online.target
After=network-online.target
ConditionFileNotEmpty=/etc/stronghold-agent/agent.hcl

[Service]
Type=notify
User=stronghold-agent
Group=stronghold-agent
ExecStart=/usr/local/bin/stronghold -config=/etc/stronghold-agent/agent.hcl
ExecReload=/bin/kill -HUP $MAINPID
KillMode=process
KillSignal=SIGTERM
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true

ReadWritePaths=/var/run/stronghold-agent /var/log/stronghold-agent /etc/myapp
CapabilityBoundingSet=CAP_IPC_LOCK

[Install]
WantedBy=multi-user.target
```

В этом примере:

- `ConditionFileNotEmpty` проверяет, что конфигурационный файл существует и не пустой;
- `ExecStart` задаёт команду запуска Agent;
- `ExecReload` отправляет `HUP` основному процессу;
- `Restart=on-failure` автоматически перезапускает Agent после сбоя;
- `ProtectSystem=strict` и другие параметры усиливают изоляцию процесса;
- `ReadWritePaths` задаёт директории, в которые Agent может писать.

> Важно  
> В `ReadWritePaths` нужно перечислить все директории, куда Stronghold Agent будет писать. Это могут быть:
> - `template.destination`;
> - sink-файл;
> - Unix socket;
> - директория с логами;
> - директории приложений, если Agent рендерит туда конфигурацию.

Пример в unit-файле базовый. При необходимости добавьте свои пути, например `/etc/myapp` или `/var/lib/myapp`.

## Управление сервисом

После создания unit-файла выполните:

```bash
sudo systemctl daemon-reload
```

Запустите Agent:

```bash
sudo systemctl start stronghold-agent
```

Включите автозапуск:

```bash
sudo systemctl enable stronghold-agent
```

Проверьте статус:

```bash
sudo systemctl status stronghold-agent
```

Посмотрите логи:

```bash
sudo journalctl -u stronghold-agent -f
```

Перезагрузите конфигурацию:

```bash
sudo systemctl reload stronghold-agent
```

Остановите сервис:

```bash
sudo systemctl stop stronghold-agent
```

## Практические рекомендации

Чтобы запуск Stronghold Agent прошёл без проблем:

- перед запуском как сервиса всегда выполняйте пробную проверку через `-exit-after-auth`;
- сначала убедитесь, что Agent может пройти аутентификацию и записать токен;
- заранее проверьте, что пользователь `stronghold-agent` имеет доступ ко всем нужным директориям;
- если используется `ProtectSystem=strict`, не забудьте перечислить все writable-директории в `ReadWritePaths`;
- для диагностики проблем сначала запускайте Agent в foreground-режиме с `-log-level=debug`, а затем переносите его в `systemd`.

## Что дальше

- Если нужно настроить конфигурационный файл Agent, откройте [Настройка](../settings/).
- Если вы хотите разобраться в режимах работы Agent, откройте [Возможности](./capabilities/).
