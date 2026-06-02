---
title: "Запуск и управление"
linkTitle: "Запуск и управление"
description: "Запуск, проверка конфигурации и управление Stronghold Agent"
weight: 20
---

Эта страница поможет проверить конфигурацию Stronghold Agent, выполнить пробный запуск и перевести агент в режим постоянной работы.

Перед запуском в production-окружении проверьте конфигурационный файл.
Убедитесь, что агент может подключиться к серверу Stronghold, пройти аутентификацию и создать нужные файлы.

## Проверка конфигурации

Перед запуском в production-окружении обязательно проверьте корректность конфигурации.

Для этого выполните пробный запуск с автоматическим завершением:

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl -exit-after-auth -log-level=debug
```

Эта команда выполняет следующие действия:

1. Проверяет синтаксис HCL-конфигурации.
1. Подключается к серверу Stronghold.
1. Выполняет полную аутентификацию.
1. Создаёт файлы и шаблоны.
1. Автоматически завершает работу.

Ниже приведён пример успешного результата:

```text
[INFO]  agent: loaded config: path=/etc/stronghold-agent/agent.hcl
[INFO]  agent.auto_auth.approle: authentication successful
[INFO]  agent.sink.file: writing token to: /var/run/stronghold-agent/token
[INFO]  agent: exit after auth set, exiting
```

Такой запуск удобно использовать как первую проверку перед запуском агента в фоне или как сервиса systemd.

## Запуск в режиме разработки

Для отладки Stronghold Agent можно запускать в интерактивном режиме.

Используйте один из следующих вариантов запуска:

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl
stronghold -config=/etc/stronghold-agent/agent.hcl -log-level=debug
stronghold -config=/etc/stronghold-agent/agent.hcl -exit-after-auth
```

Этот режим полезен, если нужно:

- Проверить, проходит ли аутентификация.
- Посмотреть, как рендерятся шаблоны.
- Убедиться, что агент может записывать токены и файлы.
- Быстро найти ошибки в конфигурации.

## Запуск Stronghold Agent как сервиса systemd

Для постоянной работы Stronghold Agent обычно запускают как сервис systemd.

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
- `ExecStart` задаёт команду запуска агента;
- `ExecReload` отправляет `HUP` основному процессу;
- `Restart=on-failure` автоматически перезапускает агент после сбоя;
- `ProtectSystem=strict` и другие параметры усиливают изоляцию процесса;
- `ReadWritePaths` задаёт директории, в которые агент может писать.

{{< alert level="warning" >}}
В `ReadWritePaths` перечислите все директории, в которые Stronghold Agent будет писать. Это могут быть директория из `template.destination`, sink-файл, Unix socket, директория с логами и директории приложений, если агент рендерит туда конфигурацию.
{{< /alert >}}

Пример unit-файла выше — базовый.
При необходимости добавьте свои пути, например `/etc/myapp` или `/var/lib/myapp`.

## Управление сервисом

После создания unit-файла выполните следующие команды:

```shell
sudo systemctl daemon-reload
sudo systemctl start stronghold-agent
sudo systemctl enable stronghold-agent
sudo systemctl status stronghold-agent
sudo journalctl -u stronghold-agent -f
sudo systemctl reload stronghold-agent
sudo systemctl stop stronghold-agent
```

## Практические рекомендации

Чтобы запуск Stronghold Agent прошёл без проблем, учитывайте следующие рекомендации:

- Перед запуском как сервиса всегда выполняйте проверку через `-exit-after-auth`.
- Сначала убедитесь, что агент может пройти аутентификацию и записать токен.
- Заранее проверьте, что пользователь `stronghold-agent` имеет доступ ко всем нужным директориям.
- Если используется `ProtectSystem=strict`, перечислите все директории для записи в `ReadWritePaths`.
- Для диагностики проблем сначала запускайте агент в интерактивном режиме с `-log-level=debug`, а затем переносите его в systemd.
- Если используется блок `template`, убедитесь, что директория назначения доступна для записи.
- Если используется Unix socket или sink-файл, также добавьте их пути в `ReadWritePaths`.
