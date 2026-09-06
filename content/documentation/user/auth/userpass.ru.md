---
title: "Метод аутентификации по имени пользователя и паролю"
linkTitle: "userpass"
weight: 50
---

Метод аутентификации `userpass` позволяет пользователям проходить аутентификацию в Deckhouse Stronghold с помощью имени пользователя и пароля.

## Особенности метода

При использовании метода `userpass` учитывайте следующие особенности:

- Имена пользователей и пароли задаются непосредственно в методе аутентификации по пути `auth/userpass/users/`.
- Метод `userpass` не может считывать имена пользователей и пароли из внешнего источника.
- Введённые имена пользователей приводятся в нижний регистр. Например, `Mary` и `mary` — это равноценные записи.

## Настройка

Чтобы пользователи могли проходить аутентификацию, настройте метод `userpass`.
Обычно эти действия выполняет оператор или система управления конфигурацией.

Для настройки аутентификации с помощью метода `userpass` выполните следующие шаги:

1. Включите метод аутентификации `userpass`:

   {{< tabs name="stronghold_cmd_10" >}}
   {{% tab name="Stronghold в DKP" %}}
   ```shell
   d8 stronghold auth enable userpass
   ```
   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}
   ```shell
   stronghold auth enable userpass
   ```
   {{% /tab %}}
   {{< /tabs >}}

   Метод будет включён по пути `auth/userpass`.

   Чтобы включить метод по другому пути, используйте флаг `-path`:

   {{< tabs name="stronghold_cmd_24481" >}}
   {{% tab name="Stronghold в DKP" %}}
   ```shell
   d8 stronghold auth enable -path=<userpass_mount_path> userpass
   ```
   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}
   ```shell
   stronghold auth enable -path=<userpass_mount_path> userpass
   ```
   {{% /tab %}}
   {{< /tabs >}}

1. Создайте пользователя (если необходимо), которому разрешена аутентификация:

   {{< tabs name="stronghold_cmd_67504" >}}
   {{% tab name="Stronghold в DKP" %}}
   ```shell
   d8 stronghold write auth/<userpass_mount_path>/users/alice \
     password=Pass-123! \
     token_policies=admins
   ```
   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}
   ```shell
   stronghold write auth/<userpass_mount_path>/users/alice \
     password=Pass-123! \
     token_policies=admins
   ```
   {{% /tab %}}
   {{< /tabs >}}

В результате будет создан пользователь `alice` с паролем `Pass-123!` и [политикой](../../concepts/policy/) `admins`.

### Аутентификация пользователя с помощью метода userpass

Пример команды для аутентификации пользователя с помощью метода `userpass`:

{{< tabs name="stronghold_cmd_13653" >}}
{{% tab name="Stronghold в DKP" %}}
```shell
d8 stronghold login -method=userpass username=alice password="Pass-123!"
```
{{% /tab %}}
{{% tab name="Stronghold в Linux" %}}
```shell
stronghold login -method=userpass username=alice password="Pass-123!"
```
{{% /tab %}}
{{< /tabs >}}

### Блокировка пользователя

Если пользователь несколько раз подряд укажет неверные учётные данные, Stronghold на некоторое время прекратит проверять их и вернёт ошибку с отказом в доступе.
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

Функцию блокировки пользователя можно отключить с помощью команды `auth tune`, передав параметру `disable_lockout` значение `true`:

{{< tabs name="stronghold_cmd_24697" >}}
{{% tab name="Stronghold в DKP" %}}
```shell
d8 stronghold auth tune -user-lockout-disable=true userpass
```
{{% /tab %}}
{{% tab name="Stronghold в Linux" %}}
```shell
stronghold auth tune -user-lockout-disable=true userpass
```
{{% /tab %}}
{{< /tabs >}}

{{< alert level="warning" >}}
Функция блокировки пользователя поддерживается только методами аутентификации `userpass`, `ldap` и `approle`.
{{< /alert >}}

## Смена собственного пароля пользователя

Пользователю можно разрешить менять с помощью метода `userpass` только собственный пароль.
Для этого создайте политику, в которой путь к паролю текущего пользователя формируется на основе имени аутентифицированного пользователя.

### Создание политики

Используйте шаблон политики:

```hcl
path "auth/userpass/users/{{identity.entity.aliases.<accessor>.name}}/password" {
  capabilities = ["update"]
}
```

Значение `<accessor>` получите с помощью команды:

{{< tabs name="stronghold_cmd_55997" >}}
{{% tab name="Stronghold в DKP" %}}
```shell
d8 stronghold read -field=accessor sys/auth/userpass
```
{{% /tab %}}
{{% tab name="Stronghold в Linux" %}}
```shell
stronghold read -field=accessor sys/auth/userpass
```
{{% /tab %}}
{{< /tabs >}}

Шаблон `{{identity.entity.aliases.<accessor>.name}}` автоматически подставляет имя аутентифицированного пользователя.
Поэтому путь всегда указывает только на пароль текущего пользователя.

Шаблон работает после аутентификации пользователя через метод `userpass`.

### Разрешение смены собственного пароля для пользователя

Чтобы разрешить пользователю менять собственный пароль с помощью метода `userpass`, выполните следующие шаги:

{{< tabs >}}
{{% tab "Если метод userpass уже включен" %}}

1. Получите уникальный идентификатор метода:

   {{< tabs name="stronghold_cmd_3381" >}}
   {{% tab name="Stronghold в DKP" %}}
   ```shell
   ACCESSOR=$(d8 stronghold read -field=accessor sys/auth/userpass)
   ```
   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}
   ```shell
   ACCESSOR=$(stronghold read -field=accessor sys/auth/userpass)
   ```
   {{% /tab %}}
   {{< /tabs >}}

1. Создайте политику, позволяющую пользователю, аутентифицированному через `userpass`, менять свой пароль:

   {{< tabs name="stronghold_cmd_50546" >}}
   {{% tab name="Stronghold в DKP" %}}
   ```shell
   d8 stronghold policy write self-change-password - <<EOF
   path "auth/userpass/users/{{identity.entity.aliases.${ACCESSOR}.name}}/password" {
     capabilities = ["update"]
   }
   EOF
   ```
   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}
   ```shell
   stronghold policy write self-change-password - <<EOF
   path "auth/userpass/users/{{identity.entity.aliases.${ACCESSOR}.name}}/password" {
     capabilities = ["update"]
   }
   EOF
   ```
   {{% /tab %}}
   {{< /tabs >}}

1. Привяжите политику к пользователю, которому нужно разрешить менять свой пароль:

{{< tabs >}}
{{% tab "Если пользователь существует" %}}

{{< tabs name="stronghold_cmd_54664" >}}
{{% tab name="Stronghold в DKP" %}}
```shell
d8 stronghold write auth/userpass/users/alice/policies \
  token_policies="self-change-password"
```
{{% /tab %}}
{{% tab name="Stronghold в Linux" %}}
```shell
stronghold write auth/userpass/users/alice/policies \
  token_policies="self-change-password"
```
{{% /tab %}}
{{< /tabs >}}

Этот пример привяжет политику `self-change-password` к существующему пользователю `alice`

{{% /tab %}}
{{% tab "Если пользователя не существует" %}}

{{< tabs name="stronghold_cmd_90751" >}}
{{% tab name="Stronghold в DKP" %}}
```shell
d8 stronghold write auth/userpass/users/alice \
  password="OldPass-123!" \
  token_policies="self-change-password"
```
{{% /tab %}}
{{% tab name="Stronghold в Linux" %}}
```shell
stronghold write auth/userpass/users/alice \
  password="OldPass-123!" \
  token_policies="self-change-password"
```
{{% /tab %}}
{{< /tabs >}}

Этот пример создаст пользователя `alice`, разрешит ему аутентификацию через `userpass` и привяжет к нему политику `self-change-password`.

{{% /tab %}}
{{< /tabs >}}

{{% /tab %}}
{{% tab "Если метод userpass не включен" %}}

1. Включите метод `userpass`:

   {{< tabs name="stronghold_cmd_10" >}}
   {{% tab name="Stronghold в DKP" %}}
   ```shell
   d8 stronghold auth enable userpass
   ```
   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}
   ```shell
   stronghold auth enable userpass
   ```
   {{% /tab %}}
   {{< /tabs >}}

1. Получите уникальный идентификатор метода:

   {{< tabs name="stronghold_cmd_3381" >}}
   {{% tab name="Stronghold в DKP" %}}
   ```shell
   ACCESSOR=$(d8 stronghold read -field=accessor sys/auth/userpass)
   ```
   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}
   ```shell
   ACCESSOR=$(stronghold read -field=accessor sys/auth/userpass)
   ```
   {{% /tab %}}
   {{< /tabs >}}

1. Создайте политику, позволяющую пользователю, аутентифицированному через `userpass`, менять свой пароль:

   {{< tabs name="stronghold_cmd_50546" >}}
   {{% tab name="Stronghold в DKP" %}}
   ```shell
   d8 stronghold policy write self-change-password - <<EOF
   path "auth/userpass/users/{{identity.entity.aliases.${ACCESSOR}.name}}/password" {
     capabilities = ["update"]
   }
   EOF
   ```
   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}
   ```shell
   stronghold policy write self-change-password - <<EOF
   path "auth/userpass/users/{{identity.entity.aliases.${ACCESSOR}.name}}/password" {
     capabilities = ["update"]
   }
   EOF
   ```
   {{% /tab %}}
   {{< /tabs >}}

1. Привяжите политику к пользователю, которому нужно разрешить менять свой пароль:

{{< tabs >}}
{{% tab "Если пользователь существует" %}}

{{< tabs name="stronghold_cmd_54664" >}}
{{% tab name="Stronghold в DKP" %}}
```shell
d8 stronghold write auth/userpass/users/alice/policies \
  token_policies="self-change-password"
```
{{% /tab %}}
{{% tab name="Stronghold в Linux" %}}
```shell
stronghold write auth/userpass/users/alice/policies \
  token_policies="self-change-password"
```
{{% /tab %}}
{{< /tabs >}}

Этот пример привяжет политику `self-change-password` к существующему пользователю `alice`

{{% /tab %}}
{{% tab "Если пользователя не существует" %}}

{{< tabs name="stronghold_cmd_90751" >}}
{{% tab name="Stronghold в DKP" %}}
```shell
d8 stronghold write auth/userpass/users/alice \
  password="OldPass-123!" \
  token_policies="self-change-password"
```
{{% /tab %}}
{{% tab name="Stronghold в Linux" %}}
```shell
stronghold write auth/userpass/users/alice \
  password="OldPass-123!" \
  token_policies="self-change-password"
```
{{% /tab %}}
{{< /tabs >}}

Этот пример создаст пользователя `alice`, разрешит ему аутентификацию через `userpass` и привяжет к нему политику `self-change-password`.

{{% /tab %}}
{{< /tabs >}}

{{% /tab %}}
{{< /tabs >}}

### Смена пароля пользователем

После [аутентификации](#аутентификация-пользователя-с-помощью-метода-userpass) пользователь может изменить свой пароль, если для него это [разрешено](#разрешение-смены-собственного-пароля-для-пользователя).

Пример команды для смены пользователем своего пароля:

{{< tabs name="stronghold_cmd_95815" >}}
{{% tab name="Stronghold в DKP" %}}
```shell
d8 stronghold write auth/userpass/users/alice/password password="NewPass-456!"
```
{{% /tab %}}
{{% tab name="Stronghold в Linux" %}}
```shell
stronghold write auth/userpass/users/alice/password password="NewPass-456!"
```
{{% /tab %}}
{{< /tabs >}}

Если пользователь попытается изменить чужой пароль, Stronghold вернёт ошибку `permission denied`.
При указании пользователем неверных учетных данных при смене пароля возможна его [блокировка](#блокировка-пользователя).

### Политика паролей по умолчанию

Если для метода `userpass` не назначена пользовательская [политика паролей](../../concepts/password-policy/), при создании пользователя или смене пароля пользователя Stronghold использует политику по умолчанию.

Политика по умолчанию для метода `userpass` требует:

- `8` символов в пароле;
- минимум одну прописную букву;
- минимум одну строчную букву;
- минимум одну цифру;
- минимум один символ тире (`-`).

Чтобы для метода `userpass` вместо политики паролей по умолчанию использовать пользовательскую политику, выполните команду (вместо `policy_name` укажите название нужной политики):

{{< tabs name="stronghold_cmd_86644" >}}
{{% tab name="Stronghold в DKP" %}}
```shell
d8 stronghold write auth/userpass/password-policy/{policy_name}
```
{{% /tab %}}
{{% tab name="Stronghold в Linux" %}}
```shell
stronghold write auth/userpass/password-policy/{policy_name}
```
{{% /tab %}}
{{< /tabs >}}
