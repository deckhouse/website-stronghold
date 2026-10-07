---
title: "Токен"
description: "Встроенный метод аутентификации Token: вход по токену, создание служебных, пакетных, токенов-сирот и периодических токенов, роли хранилища токенов и работа через accessor."
weight: 60
---

## Метод аутентификации по токену (Token auth)

Метод аутентификации по токену является встроенным и автоматически доступен по адресу `/auth/token`. Он позволяет пользователям проходить аутентификацию с помощью токена, а также создавать новые токены, отзывать секреты по токену и т.д.

Когда любой другой метод аутентификации возвращает идентификатор, ядро Deckhouse Stronghold вызывает метод token для создания нового уникального токена для этого идентификатора.

Хранилище токенов также может быть использовано в обход любого другого метода аутентификации: вы можете создавать токены напрямую, а также выполнять различные другие операции с токенами, такие как обновление и отзыв.

## Аутентификация

### Через CLI

В этом примере пользователь выполняет вход в систему, используя токен:

{{< tabs name="stronghold_cmd_98272" >}}
{{% tab name="Stronghold в DKP" %}}

```shell
d8 stronghold login token=<token>
```

{{% /tab %}}
{{% tab name="Stronghold в Linux" %}}

```shell
stronghold login token=<token>
```

{{% /tab %}}
{{< /tabs >}}

В следующем примере пользователь выполняет вход в систему с использованием метода аутентификации `userpass`. Пользователь вводит свои учетные данные в формате `username=значение` и `password=значение`.

{{< tabs name="stronghold_cmd_6358" >}}
{{% tab name="Stronghold в DKP" %}}

```shell
d8 stronghold login -method=userpass \
   username=mitchellh \
   password=foo
```

{{% /tab %}}
{{% tab name="Stronghold в Linux" %}}

```shell
stronghold login -method=userpass \
   username=mitchellh \
   password=foo
```

{{% /tab %}}
{{< /tabs >}}

### Через API

Токен задается непосредственно в виде заголовка для HTTP API. Заголовок должен иметь вид X-Vault-Token: &lt;token> или Authorization: Bearer &lt;token>.

```shell
curl \
   --request POST \
   --data '{"password": "foo"}' \
   http://127.0.0.1:8200/v1/auth/userpass/login/mitchellh
```

В ответе будет содержаться токен по адресу `auth.client_token`, как представлено ниже в примере:

```json
{
   "lease_id": "",
   "renewable": false,
   "lease_duration": 0,
   "data": null,
   "auth": {
      "client_token": "c4f280f6-fdb2-18eb-89d3-589e2e834cdb",
      "policies": ["admins"],
      "metadata": {
         "username": "mitchellh"
      },
      "lease_duration": 0,
      "renewable": false
   }
}
```

## Создание токенов

Токены создаются командой `d8 stronghold token create` (эндпоинт `auth/token/create`). Новый токен по умолчанию становится дочерним по отношению к токену, от имени которого выполняется запрос, и может получить только подмножество его политик. Подробнее о свойствах токенов — в разделе [«Токен»](../../../concepts/tokens/).

### Служебный токен

Служебный (`service`) токен используется по умолчанию. Он хранится в Stronghold, поддерживает продление, отзыв, создание дочерних токенов и имеет accessor:

```bash
d8 stronghold token create \
  -policy="app-read" \
  -ttl=1h \
  -display-name="app-1"
```

### Пакетный токен

Пакетный (`batch`) токен не хранится в Stronghold, не продлевается, не отзывается вручную и не имеет accessor. Он подходит для большого числа короткоживущих операций:

```bash
d8 stronghold token create \
  -type=batch \
  -policy="app-read" \
  -ttl=15m
```

### Осиротевший токен

Токен-сирота (orphan token) не имеет родителя и не отзывается при отзыве токена, от имени которого был создан. Для его создания нужен доступ к `auth/token/create-orphan` или возможность `sudo` на `auth/token/create`:

```bash
d8 stronghold token create -orphan -policy="app-read" -ttl=24h
```

### Периодический токен

Периодический токен не имеет максимального срока жизни: при каждом продлении его TTL сбрасывается до значения периода. Токен истекает, если его не продлевать дольше периода:

```bash
d8 stronghold token create -period=1h -policy="app-read"
```

### Явный максимальный TTL

Параметр `-explicit-max-ttl` задаёт жёсткое ограничение времени жизни токена, которое нельзя превысить продлением, в том числе для периодических токенов:

```bash
d8 stronghold token create \
  -policy="app-read" \
  -ttl=1h \
  -explicit-max-ttl=8h
```

### Ограничение числа использований

Параметр `-use-limit` задаёт, сколько раз можно использовать токен. После исчерпания лимита токен отзывается:

```bash
d8 stronghold token create -policy="app-read" -use-limit=3 -ttl=10m
```

### Другие параметры

| Флаг CLI | Параметр API | Описание |
|----------|--------------|----------|
| `-policy` | `policies` | Политики токена. Флаг можно указать несколько раз. |
| `-no-default-policy` | `no_default_policy` | Не добавлять политику `default`. |
| `-ttl` | `ttl` | Начальный TTL токена. |
| `-explicit-max-ttl` | `explicit_max_ttl` | Жёсткий максимальный TTL. |
| `-period` | `period` | Период продления периодического токена. |
| `-renewable` | `renewable` | Разрешено ли продление. По умолчанию `true`. |
| `-orphan` | `no_parent` | Создать токен без родителя. |
| `-type` | `type` | Тип токена: `service` или `batch`. |
| `-use-limit` | `num_uses` | Максимальное число использований. |
| `-display-name` | `display_name` | Отображаемое имя токена. |
| `-metadata` | `meta` | Произвольные метаданные `key=value`. |
| `-entity-alias` | `entity_alias` | Псевдоним сущности, с которым связывается токен. |
| `-role` | — | Создать токен по роли хранилища токенов. |

## Роли хранилища токенов

Роль хранилища токенов (`auth/token/roles/<имя>`) задаёт свойства токенов, которые выдаются по ней: допустимые политики, тип токена, TTL, период, признак токена-сироты и другие. С помощью ролей можно разрешить создание токенов с политиками, которых нет у вызывающего токена, не выдавая ему `sudo`.

1. Создайте роль:

   ```bash
   d8 stronghold write auth/token/roles/ci-deploy \
     allowed_policies="deploy" \
     disallowed_policies="admin" \
     orphan=true \
     token_type=service \
     token_period=1h \
     token_explicit_max_ttl=24h \
     renewable=true
   ```

   Основные параметры роли:

   | Параметр | Описание |
   |----------|----------|
   | `allowed_policies`, `allowed_policies_glob` | Политики, которые можно назначить токену. |
   | `disallowed_policies`, `disallowed_policies_glob` | Политики, которые запрещено запрашивать. |
   | `orphan` | Выдавать токены-сироты. |
   | `token_type` | Тип токенов: `service` или `batch`. |
   | `token_period` | Период для периодических токенов. |
   | `token_explicit_max_ttl` | Явный максимальный TTL. |
   | `token_num_uses` | Лимит использований токена. |
   | `token_bound_cidrs` | CIDR-блоки, из которых разрешено использовать токен. |
   | `renewable` | Разрешено ли продление токенов. |
   | `path_suffix` | Суффикс пути токена для последующего отзыва по префиксу. |
   | `allowed_entity_aliases` | Разрешённые псевдонимы сущностей. |

1. Выдайте токен по роли:

   ```bash
   d8 stronghold token create -role=ci-deploy
   ```

   Для этого вызывающему токену нужна возможность `update` на пути `auth/token/create/ci-deploy`.

1. Просмотрите роли:

   ```bash
   d8 stronghold list auth/token/roles
   d8 stronghold read auth/token/roles/ci-deploy
   ```

## Просмотр, продление и отзыв токенов

- Просмотр свойств текущего токена:

  ```bash
  d8 stronghold token lookup
  ```

- Просмотр свойств другого токена:

  ```bash
  d8 stronghold token lookup <токен>
  ```

- Продление текущего токена или другого токена с запросом нового TTL:

  ```bash
  d8 stronghold token renew
  d8 stronghold token renew -increment=1h <токен>
  ```

- Отзыв токена вместе со всеми дочерними токенами и арендами:

  ```bash
  d8 stronghold token revoke <токен>
  ```

- Отзыв только самого токена: его прямые дочерние токены становятся токенами-сиротами:

  ```bash
  d8 stronghold token revoke -mode=orphan <токен>
  ```

- Отзыв текущего токена:

  ```bash
  d8 stronghold token revoke -self
  ```

- Проверка возможностей токена на пути:

  ```bash
  d8 stronghold token capabilities <токен> secret/data/app
  ```

## Работа через accessor

Accessor — это ссылка на токен, которая позволяет просматривать, продлевать и отзывать токен, не зная его значения. Accessor возвращается при создании токена (поле `token_accessor`) и в выводе `token lookup` (поле `accessor`). У пакетных токенов accessor отсутствует.

- Просмотр токена по accessor:

  ```bash
  d8 stronghold token lookup -accessor <accessor>
  ```

- Продление токена по accessor:

  ```bash
  d8 stronghold token renew -accessor <accessor>
  ```

- Отзыв токена по accessor:

  ```bash
  d8 stronghold token revoke -accessor <accessor>
  ```

- Проверка возможностей токена по accessor:

  ```bash
  d8 stronghold token capabilities -accessor <accessor> secret/data/app
  ```

- Список accessor всех токенов (требуется `sudo`):

  ```bash
  d8 stronghold list auth/token/accessors
  ```

{{< alert level="warning" >}}
Список accessor позволяет отозвать любой токен. Выдавайте доступ к `auth/token/accessors` только администраторам.
{{< /alert >}}

## Очистка хранилища токенов

Эндпоинт `auth/token/tidy` удаляет из хранилища токенов устаревшие записи, например accessor уже истёкших токенов:

```bash
d8 stronghold write -force auth/token/tidy
```

## Примеры использования

Готовые примеры с этим механизмом:

- [Аварийный доступ (break-glass)](../../../examples/access/break-glass-access/)

Все примеры собраны в разделе [«Примеры использования»](../../../examples/).
