---
title: "AppRole"
linkTitle: "AppRole"
description: "Аутентификация Stronghold Agent с помощью AppRole"
weight: 50
---

AppRole — рекомендуемый метод аутентификации Stronghold Agent для приложений,
которые работают на виртуальных машинах и bare metal.
Этот метод подходит для сценариев,
в которых приложение или сервис должны получать доступ к Stronghold без участия пользователя.

На этой странице описано,
как работает AppRole,
какие сущности он использует
и как настроить Stronghold Agent для аутентификации с помощью этого метода.

## Как работает AppRole

AppRole использует две сущности:

- `Role ID` — идентификатор роли;
- `Secret ID` — секретный идентификатор.

Для успешной аутентификации Stronghold Agent должен получить оба значения.
После этого Agent обращается к методу аутентификации AppRole
и получает токен Stronghold,
который затем использует для работы с секретами.

## Когда использовать AppRole

Используйте AppRole в следующих случаях:

- Приложение работает на виртуальной машине или на bare metal.
- Нужно автоматически проходить аутентификацию без участия пользователя.
- Требуется ограничить доступ набором политик.
- Нужно контролировать срок действия и число использований `Secret ID`.

Для production-окружения AppRole обычно подходит лучше,
чем статический токен,
так как этот метод позволяет гибко настраивать срок действия,
политики и ограничения доступа.

## Преимущества AppRole

AppRole даёт следующие преимущества:

- Разделяет публичный идентификатор и секретное значение.
- Позволяет ограничивать доступ политиками.
- Поддерживает ограничения по CIDR.
- Поддерживает одноразовые и временные `Secret ID`.
- Подходит для автоматизации на виртуальных машинах и bare metal.

## Настройка AppRole в Stronghold

Чтобы включить и настроить AppRole,
выполните следующие команды:

```shell
stronghold auth enable approle

stronghold write auth/approle/role/myapp \
  token_ttl=1h \
  token_max_ttl=4h \
  policies="myapp-policy"
```

После этого получите `Role ID`:

```shell
stronghold read auth/approle/role/myapp/role-id
```

Затем создайте `Secret ID`:

```shell
stronghold write -f auth/approle/role/myapp/secret-id
```

## Конфигурация Stronghold Agent

Ниже приведён пример настройки Stronghold Agent для работы с AppRole:

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

В этом примере Agent:

- читает `Role ID` из файла `/etc/stronghold-agent/role-id`;
- читает `Secret ID` из файла `/etc/stronghold-agent/secret-id`;
- удаляет файл `Secret ID` после чтения.

## Хранение Role ID и Secret ID

Учитывайте следующие особенности хранения идентификаторов:

### Role ID

`Role ID` — это идентификатор роли.
Обычно его хранят в файле `/etc/stronghold-agent/role-id`.

Для `Role ID` характерны следующие особенности:

- его можно доставлять через систему управления конфигурацией;
- его можно хранить в образе виртуальной машины;
- после использования его обычно не удаляют;
- сам по себе он не считается критичным секретом.

### Secret ID

`Secret ID` — это чувствительный секрет.
Обычно его хранят в файле `/etc/stronghold-agent/secret-id`.

Для `Secret ID` характерны следующие особенности:

- его нужно доставлять по защищённому каналу;
- его не следует хранить в Git в открытом виде;
- его можно удалять после чтения;
- при компрометации его нужно перевыпустить.

## Типы Secret ID

Можно использовать следующие типы `Secret ID`:

### Одноразовый Secret ID

Используйте этот вариант,
если хотите разрешить только одно использование секрета:

```shell
stronghold write auth/approle/role/myapp \
  secret_id_num_uses=1 \
  policies="myapp-policy"
```

Или создайте одноразовый `Secret ID` при выпуске:

```shell
stronghold write -f auth/approle/role/myapp/secret-id num_uses=1
```

Этот вариант обычно рекомендуют для production-окружения.

### Многоразовый Secret ID

Используйте этот вариант,
если `Secret ID` должен применяться многократно:

```shell
stronghold write auth/approle/role/myapp \
  secret_id_num_uses=0 \
  policies="myapp-policy"
```

Этот вариант подходит для разработки и тестирования,
но требует ручной ротации при компрометации.

### Secret ID с ограниченным TTL

Используйте этот вариант,
если хотите ограничить срок действия секрета:

```shell
stronghold write auth/approle/role/myapp \
  secret_id_ttl=24h \
  policies="myapp-policy"
```

После истечения TTL потребуется выпустить новый `Secret ID`.

## Полный пример настройки

Ниже приведён пример настройки политики,
роли AppRole и выдачи идентификаторов.

### Шаг 1. Включите AppRole и создайте политику

```shell
stronghold auth enable approle

stronghold policy write myapp-policy - <<EOF
path "secret/data/myapp/*" {
  capabilities = ["read"]
}
path "database/creds/myapp" {
  capabilities = ["read"]
}
EOF
```

### Шаг 2. Создайте роль AppRole

```shell
stronghold write auth/approle/role/myapp \
  token_ttl=1h \
  token_max_ttl=4h \
  policies="myapp-policy" \
  secret_id_num_uses=1 \
  secret_id_ttl=24h
```

### Шаг 3. Получите Role ID и Secret ID

```shell
stronghold read auth/approle/role/myapp/role-id
stronghold write -f auth/approle/role/myapp/secret-id
```

### Шаг 4. Подготовьте директорию на целевом сервере

```shell
ssh root@app-server.example.com << 'ENDSSH'
mkdir -p /etc/stronghold-agent
chown root:stronghold-agent /etc/stronghold-agent
chmod 750 /etc/stronghold-agent
ENDSSH
```

### Шаг 5. Сохраните Role ID

```shell
ssh root@app-server.example.com << 'ENDSSH'
echo -n "abc123-def456-ghi789" > /etc/stronghold-agent/role-id
chown stronghold-agent:stronghold-agent /etc/stronghold-agent/role-id
chmod 0640 /etc/stronghold-agent/role-id
ENDSSH
```

### Шаг 6. Сохраните Secret ID

```shell
ssh root@app-server.example.com << 'ENDSSH'
echo -n "xyz789-abc123-def456" > /etc/stronghold-agent/secret-id
chown stronghold-agent:stronghold-agent /etc/stronghold-agent/secret-id
chmod 0640 /etc/stronghold-agent/secret-id
ENDSSH
```

### Шаг 7. Создайте конфигурацию Agent

```shell
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

### Шаг 8. Ограничьте доступ к конфигурации

```shell
chown root:stronghold-agent /etc/stronghold-agent/agent.hcl
chmod 0640 /etc/stronghold-agent/agent.hcl
```

### Шаг 9. Запустите Agent и проверьте результат

```shell
systemctl start stronghold-agent
journalctl -u stronghold-agent -n 50 | grep -i "authentication successful"
ls -la /var/run/stronghold-agent/token
ls -la /etc/stronghold-agent/secret-id
```

## Практические рекомендации

При использовании AppRole соблюдайте следующие рекомендации:

- Используйте одноразовые `Secret ID` в production-окружении.
- Доставляйте `Role ID` и `Secret ID` по разным каналам.
- Ограничивайте доступ по CIDR, если это поддерживает ваш сценарий.
- Логируйте использование `Secret ID` для аудита.
- Ограничивайте права доступа к файлам `/etc/stronghold-agent/role-id`
  и `/etc/stronghold-agent/secret-id`.

{{< alert level="warning" >}}
Не храните `Secret ID` в Git или другой системе контроля версий в открытом виде.
{{< /alert >}}
