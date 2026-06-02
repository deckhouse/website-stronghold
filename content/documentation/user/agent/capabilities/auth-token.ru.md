---
title: "Аутентификация по токену"
linkTitle: "Token"
description: "Аутентификация Stronghold Agent с помощью токена"
weight: 60
---

Аутентификация по токену — самый простой способ аутентификации Stronghold Agent.
В этом режиме Agent читает готовый токен из файла
и использует его для доступа к Stronghold.

Этот способ подходит для простых и временных сценариев,
но для production-окружения обычно рекомендуется AppRole или JWT/OIDC.

## Когда использовать аутентификацию по токену

Используйте этот метод в следующих случаях:

- Нужен простой способ запуска Stronghold Agent.
- Используется тестовое или временное окружение.
- Нет возможности использовать AppRole.
- Требуется быстро проверить шаблоны,
  Auto-Auth или другие возможности Agent.

Если вы настраиваете production-окружение,
по возможности используйте AppRole или JWT/OIDC.

## Как это работает

При аутентификации по токену Stronghold Agent выполняет следующие действия:

1. Читает токен из файла.
1. Использует токен для обращения к Stronghold.
1. Передаёт токен внутренним подсистемам Agent.
1. При необходимости записывает токен во внешний sink.
1. Пытается обновлять токен,
   если это разрешено его параметрами.

Обычно этот метод используют вместе с Auto-Auth
и методом `token_file`.

## Ограничения

Учитывайте следующие ограничения этого метода:

- Токен остаётся долгоживущим секретом.
- При компрометации токена требуется ручная ротация.
- Этот метод не разделяет публичный идентификатор и секретное значение.
- Для production-окружения этот вариант подходит хуже,
  чем AppRole.

## Создание токена

Сначала создайте политику,
которая разрешит Agent читать нужные секреты:

```shell
stronghold policy write myapp-policy - <<EOF
path "secret/data/myapp/*" {
  capabilities = ["read"]
}
path "database/creds/myapp" {
  capabilities = ["read"]
}
EOF
```

Затем создайте токен:

```shell
stronghold token create \
  -policy=myapp-policy \
  -ttl=720h \
  -renewable=true \
  -display-name="myapp-agent" \
  -format=json
```

## Доставка токена на сервер

Создайте директорию для конфигурации Stronghold Agent:

```shell
ssh root@app-server.example.com << 'ENDSSH'
mkdir -p /etc/stronghold-agent
chown root:stronghold-agent /etc/stronghold-agent
chmod 750 /etc/stronghold-agent
ENDSSH
```

Сохраните токен в файл:

```shell
echo -n "$AGENT_TOKEN" | ssh root@app-server.example.com 'cat > /etc/stronghold-agent/token'
```

Ограничьте права доступа к файлу:

```shell
ssh root@app-server.example.com << 'ENDSSH'
chown stronghold-agent:stronghold-agent /etc/stronghold-agent/token
chmod 0640 /etc/stronghold-agent/token
ENDSSH
```

## Конфигурация Stronghold Agent

Создайте файл `/etc/stronghold-agent/agent.hcl` со следующей конфигурацией:

```hcl
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
```

В этом примере Agent:

- читает токен из файла `/etc/stronghold-agent/token`;
- использует Auto-Auth с методом `token_file`;
- записывает рабочий токен в файл `/var/run/stronghold-agent/token`;
- рендерит файл конфигурации приложения.

## Проверка

Запустите Stronghold Agent и проверьте журнал:

```shell
systemctl start stronghold-agent
journalctl -u stronghold-agent -n 50
```

Если аутентификация выполнена успешно,
Agent сможет читать секреты и выполнять связанные задачи,
например рендеринг шаблонов.

## Важные параметры токена

При создании токена учитывайте следующие параметры:

| Параметр | Описание |
| --- | --- |
| `ttl` | Начальное время жизни токена |
| `renewable` | Разрешает обновление токена |
| `period` | Задаёт периодическое обновление |
| `explicit-max-ttl` | Ограничивает максимальное время жизни |

Если токен нельзя обновлять,
после истечения TTL Agent не сможет продолжить работу без нового токена.

## Практические рекомендации

При использовании аутентификации по токену соблюдайте следующие рекомендации:

- Используйте `renewable=true`,
  если это допускает ваш сценарий.
- Устанавливайте разумный TTL,
  чтобы уменьшить риск при компрометации.
- Ограничивайте максимальное время жизни токена.
- Храните токен с минимально необходимыми правами доступа.
- Регулярно отзывайте неиспользуемые токены.
- Не храните токен в Git или в другой системе контроля версий в открытом виде.

{{< alert level="warning" >}}
Аутентификация по токену подходит в первую очередь для тестовых,
временных и упрощённых сценариев.
Для production-окружения по возможности используйте AppRole или JWT/OIDC.
{{< /alert >}}
