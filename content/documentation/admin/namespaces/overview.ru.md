---
title: "Пространства имён"
linkTitle: "Введение"
weight: 10
description: "Руководство администратора по изоляции и управлению пространствами имён в Stronghold."
---

## Что такое пространства имён

Пространства имён (`namespaces`) в Stronghold позволяют разделить один сервер на несколько изолированных логических областей. Каждое пространство имён работает как отдельный виртуальный Stronghold со своими собственными:

- политиками (`policies`);
- путями монтирования (`mount paths`);
- секретами;
- методами аутентификации (`auth methods`);
- токенами и сессиями.

Все пространства имён управляются одним экземпляром Stronghold и образуют иерархию с корневым пространством `root`.

## Основные свойства

- пространства имён образуют дерево, в котором у каждого пространства может быть родитель и дочерние пространства;
- данные и конфигурация изолированы между пространствами имён;
- администратор родительского пространства может управлять дочерними пространствами;
- пользователи и сервисы работают только в пределах своего пространства имён и его дочерних пространств, если это разрешено политиками.

Для обращения к нужному пространству имён:

- в CLI используйте параметр `-namespace=<namespace_path>`;
- в REST API передавайте заголовок `X-Vault-Namespace: <namespace_path>`.

## Создание пространства имён

Для создания нового пространства имён нужны права чтения и записи на путь `/namespaces` в текущем пространстве имён.

Имя нового пространства не должно пересекаться с уже существующими путями. После создания к нему можно обращаться как по полному пути в иерархии.

### Через CLI

```shell
d8 stronghold namespace create \
  -namespace=<parent_namespace_name> \
  -custom-metadata=key="value" \
  <new_namespace_name>
```

Пример ответа:

```text
Key                Value
---                -----
custom_metadata    map[key:value]
id                 c88b1992-d9d3-4160-8283-77163a4e7fae
path               <new_namespace_name>/
```

Параметры:

- `-namespace` - абсолютный путь родительского пространства, в котором будет создано новое пространство; если параметр не указан, пространство создаётся в `root`;
- `-custom-metadata` - произвольные метаданные пространства имён в формате `key=value`.

### Через REST API

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_TOKEN" \
  --header "X-Vault-Namespace: <parent_namespace_name>" \
  --request POST \
  --data @payload.json \
  $STRONGHOLD_ADDR/v1/sys/namespaces/<new_namespace_name> | jq -r ".data"
```

Параметры:

- `X-Vault-Namespace` - абсолютный путь родительского пространства, в котором создаётся новое вложенное пространство;
- `payload.json` - JSON с описанием метаданных пространства имён.

## Чтение и список пространств имён

### Чтение через CLI

```shell
d8 stronghold namespace lookup -namespace=<parent_namespace_name> <namespace_name>
```

Пример ответа:

```text
Key                Value
---                -----
custom_metadata    map[key:value]
id                 c88b1992-d9d3-4160-8283-77163a4e7fae
path               <namespace_name>/
```

### Чтение через REST API

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_TOKEN" \
  --header "X-Vault-Namespace: <parent_namespace_name>" \
  --request GET \
  $STRONGHOLD_ADDR/v1/sys/namespaces/<namespace_name>
```

### Список через CLI

```shell
d8 stronghold namespace list -namespace=<parent_namespace_name>
```

Пример ответа:

```text
Keys
----
ns1/
ns2/
```

### Список через REST API

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_TOKEN" \
  --header "X-Vault-Namespace: <parent_namespace_name>" \
  --request LIST \
  $STRONGHOLD_ADDR/v1/sys/namespaces
```

## Удаление пространства имён

### Через CLI

```shell
d8 stronghold namespace delete -namespace=<parent_namespace_name> <namespace_name>
```

Параметр `-namespace` задаёт родительское пространство, в котором находится удаляемое пространство. Если параметр не указан, удаление выполняется из `root`.

### Через REST API

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_TOKEN" \
  --header "X-Vault-Namespace: <parent_namespace_name>" \
  --request DELETE \
  $STRONGHOLD_ADDR/v1/sys/namespaces/<namespace_name>
```

## Блокировка API пространства имён

`Namespace API Lock` позволяет временно запретить все API-запросы к выбранному пространству имён и ко всем его дочерним пространствам.

При блокировке Stronghold возвращает одноразовый `unlock_key`. Его нужно сохранить, чтобы затем снять блокировку. Разблокировка также может быть выполнена root-токеном без `unlock_key`.

### Блокировка через CLI

Заблокировать текущее пространство имён и все дочерние:

```shell
d8 stronghold namespace lock
```

Заблокировать конкретное дочернее пространство, например `ns1/ns2/`:

```shell
d8 stronghold namespace lock ns1/ns2
```

Пример ответа:

```text
Key            Value
---            -----
unlock_key     7Hk3xQ9mR2pN5vL8wY4tZa
```

{{< alert level="warning" >}}
Сохраните значение `unlock_key`. Оно потребуется для разблокировки, если операция выполняется не root-токеном.
{{< /alert >}}

### Блокировка через REST API

Заблокировать пространство имён `<namespace_name>` из корневого пространства:

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_TOKEN" \
  --request POST \
  $STRONGHOLD_ADDR/v1/sys/namespaces/api-lock/lock/<namespace_name> | jq -r ".data"
```

Заблокировать текущее пространство имён, определяемое заголовком `X-Vault-Namespace`:

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_TOKEN" \
  --header "X-Vault-Namespace: <namespace_name>" \
  --request POST \
  $STRONGHOLD_ADDR/v1/sys/namespaces/api-lock/lock | jq -r ".data"
```

Пример ответа:

```json
{
  "unlock_key": "7Hk3xQ9mR2pN5vL8wY4tZa"
}
```

### Разблокировка через CLI

Разблокировать текущее пространство имён с помощью ключа:

```shell
d8 stronghold namespace unlock -unlock-key=<key>
```

Разблокировать текущее пространство имён с помощью root-токена:

```shell
d8 stronghold namespace unlock
```

Разблокировать конкретное дочернее пространство:

```shell
d8 stronghold namespace unlock -unlock-key=<key> ns1/ns2
```

### Разблокировка через REST API

С использованием `unlock_key`:

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_TOKEN" \
  --request POST \
  --data '{"unlock_key": "<key>"}' \
  $STRONGHOLD_ADDR/v1/sys/namespaces/api-lock/unlock/<namespace_name>
```

С использованием root-токена:

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_ROOT_TOKEN" \
  --request POST \
  $STRONGHOLD_ADDR/v1/sys/namespaces/api-lock/unlock/<namespace_name>
```

### Параметры API блокировки

| Параметр | Тип | Обязательный | Описание |
|---|---|---|---|
| `<namespace_name>` | string | нет | Путь пространства имён для блокировки или разблокировки. Если не указан, операция применяется к текущему пространству, определяемому заголовком `X-Vault-Namespace`. |
| `unlock_key` | string | нет | Ключ разблокировки, полученный при блокировке. Обязателен при разблокировке, если токен не является root-токеном. |

## Пример блокировки и разблокировки

```shell
# Создаём пространство имён.
d8 stronghold namespace create production

# Блокируем его.
d8 stronghold namespace lock production
# Key            Value
# ---            -----
# unlock_key     7Hk3xQ9mR2pN5vL8wY4tZa

# Любые запросы к production теперь заблокированы.
d8 stronghold secrets list -namespace=production
# Error: API access to this namespace has been locked by an administrator...

# Разблокируем с ключом.
d8 stronghold namespace unlock -unlock-key=7Hk3xQ9mR2pN5vL8wY4tZa production

# Доступ восстановлен.
d8 stronghold secrets list -namespace=production
```
