---
title: "Метод Userpass"
linkTitle: "Userpass"
weight: 50
---

## Метод аутентификации по имени пользователя и паролю

Метод аутентификации `userpass` позволяет пользователям проходить аутентификацию в Deckhouse Stronghold с помощью имени пользователя и пароля.

Имена пользователей и пароли задаются непосредственно в методе аутентификации через путь `users/`.
Метод `userpass` не может считывать имена пользователей и пароли из внешнего источника.

Все введённые имена пользователей сохраняются в нижнем регистре.
Например, `Mary` и `mary` относятся к одной записи.

## Настройка

Перед аутентификацией пользователей настройте метод `userpass`.
Обычно эти действия выполняет оператор или система управления конфигурацией.

Выполните следующие шаги:

1. Включите метод аутентификации `userpass`:

   ```shell
   d8 stronghold auth enable userpass
   ```

   Метод будет включён по пути `auth/userpass`.

   Чтобы включить метод по другому пути, используйте флаг `-path`:

   ```shell
   d8 stronghold auth enable -path=<path> userpass
   ```

1. Создайте пользователя, которому разрешена аутентификация:

   ```shell
   d8 stronghold write auth/<userpass:path>/users/mitchellh \
     password=foo \
     policies=admins
   ```

В результате будет создан пользователь `mitchellh` с паролем `foo` и политикой `admins`.
Это единственная обязательная настройка.

## Смена собственного пароля

Пользователю можно разрешить менять только собственный пароль в методе `userpass`.
Для этого создайте политику, в которой путь к паролю пользователя формируется через алиас сущности.

### Политика

Используйте шаблон политики:

```hcl
path "auth/userpass/users/{{identity.entity.aliases.<accessor>.name}}/password" {
  capabilities = ["update"]
}
```

Значение `<accessor>` получите с помощью команды:

```shell
d8 stronghold read -field=accessor sys/auth/userpass
```

### Настройка администратором

Включите метод `userpass`, создайте политику и назначьте её пользователю:

```shell
d8 stronghold auth enable userpass
ACCESSOR=$(d8 stronghold read -field=accessor sys/auth/userpass)

d8 stronghold policy write self-change-password - <<EOF
path "auth/userpass/users/{{identity.entity.aliases.${ACCESSOR}.name}}/password" {
  capabilities = ["update"]
}
EOF

d8 stronghold write auth/userpass/users/alice \
  password="OldPass-123!" \
  token_policies="self-change-password"
```

### Смена пароля пользователем

После входа пользователь может изменить свой пароль:

```shell
d8 stronghold login -method=userpass username=alice password="OldPass-123!"
d8 stronghold write auth/userpass/users/alice/password password="NewPass-456!"
```

Если пользователь попытается изменить чужой пароль, Stronghold вернёт ошибку `permission denied`.

### Принцип работы

Шаблон `{{identity.entity.aliases.<accessor>.name}}` автоматически подставляет имя аутентифицированного пользователя.
Поэтому путь всегда указывает только на пароль текущего пользователя.

Шаблон работает после входа через метод `userpass`.

### Проверка

Для проверки настройки выполните скрипт `userpass_self_password_verify.sh`, если он доступен в вашем окружении:

```shell
VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=<root-token> \
  ./userpass_self_password_verify.sh
```

## Блокировка пользователя

Если пользователь несколько раз подряд укажет неверные учётные данные, Stronghold на некоторое время прекратит проверять их и сразу вернёт ошибку с отказом в доступе.
Такое поведение называется блокировкой пользователя (`user_lockout`).

Время, на которое пользователь блокируется, называется длительностью блокировки (`lockout_duration`).
После истечения этого времени пользователь сможет снова войти в систему.

Количество неудачных попыток входа, после которых пользователь блокируется, называется порогом блокировки (`lockout_threshold`).
Счётчик порога блокировки сбрасывается через несколько минут без попыток входа или после успешного входа.
Интервал, после которого счётчик сбрасывается при отсутствии попыток входа, называется сбросом счётчика блокировки (`lockout_counter_reset`).

Блокировка пользователя помогает снизить риск атак с подбором пароля.

Функция блокировки пользователя включена по умолчанию.
Значения по умолчанию:

- `lockout_threshold` — 5 попыток;
- `lockout_duration` — 15 минут;
- `lockout_counter_reset` — 15 минут.

Функцию блокировки пользователя можно отключить с помощью команды `auth tune`, передав параметру `disable_lockout` значение `true`.

{{< alert level="warning" >}}
Функция блокировки пользователя поддерживается только методами аутентификации `userpass`, `ldap` и `approle`.
{{< /alert >}}
