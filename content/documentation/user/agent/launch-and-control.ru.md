---
title: "Запуск и управление"
linkTitle: "Запуск и управление"
description: "Запуск, проверка конфигурации и управление Stronghold Agent"
weight: 50
---

Эта страница помогает проверить конфигурацию Stronghold Agent, выполнить пробный запуск и перевести Agent в режим постоянной работы [5].

Перед запуском в production-окружении сначала проверьте конфигурационный файл. Убедитесь, что Agent может подключиться к серверу Stronghold, пройти аутентификацию и создать нужные файлы [5].

## Проверка конфигурации

Перед запуском в production-окружении обязательно проверьте корректность конфигурации [5].

Для этого выполните пробный запуск с автоматическим завершением:

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl -exit-after-auth -log-level=debug
```

Эта команда выполняет следующие действия [5]:

1. Проверяет синтаксис HCL-конфигурации.
1. Подключается к серверу Stronghold.
1. Выполняет полную аутентификацию.
1. Создаёт файлы и шаблоны.
1. Автоматически завершает работу.

Пример успешного результата:

```text
[INFO]  agent: loaded config: path=/etc/stronghold-agent/agent.hcl
[INFO]  agent.auto_auth.approle: authentication successful
[INFO]  agent.sink.file: writing token to: /var/run/stronghold-agent/token
[INFO]  agent: exit after auth set, exiting
```

Такой запуск удобно использовать как первую проверку перед запуском Agent в фоне или как `systemd`-сервис.

## Запуск в режиме разработки

Для отладки Stronghold Agent можно запускать в foreground-режиме [5].

Используйте один из следующих вариантов запуска:

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl
```

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl -log-level=debug
```

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl -exit-after-auth
```

Этот режим полезен, если нужно:

- проверить, проходит ли аутентификация;
- посмотреть, как рендерятся шаблоны;
- убедиться, что Agent может записывать токены и файлы;
- быстро найти ошибки в конфигурации.

## Запуск Stronghold Agent как systemd-сервиса

Для постоянной работы Stronghold Agent обычно запускают как `systemd`-сервис [5].

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
- `ReadWritePaths` задаёт директории, в которые Agent может писать [5].

{% alert level="warning" %}
В `ReadWritePaths` перечислите все директории, в которые Stronghold Agent будет писать. Это могут быть директория из `template.destination`, sink-файл, Unix socket, директория с логами и директории приложений, если Agent рендерит туда конфигурацию [5].
{% endalert %}

Пример unit-файла выше — базовый. При необходимости добавьте свои пути, например `/etc/myapp` или `/var/lib/myapp` [5].

## Управление сервисом

После создания unit-файла выполните следующие команды [5]:

```shell
sudo systemctl daemon-reload
```

```shell
sudo systemctl start stronghold-agent
```

```shell
sudo systemctl enable stronghold-agent
```

```shell
sudo systemctl status stronghold-agent
```

```shell
sudo journalctl -u stronghold-agent -f
```

```shell
sudo systemctl reload stronghold-agent
```

```shell
sudo systemctl stop stronghold-agent
```

## Практические рекомендации

Чтобы запуск Stronghold Agent прошёл без проблем, учитывайте следующие рекомендации. Они основаны на исходной странице и на описании связанных режимов работы Agent [4] [5]:

- перед запуском как сервиса всегда выполняйте проверку через `-exit-after-auth`;
- сначала убедитесь, что Agent может пройти аутентификацию и записать токен;
- заранее проверьте, что пользователь `stronghold-agent` имеет доступ ко всем нужным директориям;
- если используется `ProtectSystem=strict`, перечислите все директории для записи в `ReadWritePaths`;
- для диагностики проблем сначала запускайте Agent в foreground-режиме с `-log-level=debug`, а затем переносите его в `systemd`;
- если вы используете блок `template`, убедитесь, что директория назначения доступна для записи;
- если вы используете Unix socket или sink-файл, также добавьте их пути в `ReadWritePaths` [4] [5].
