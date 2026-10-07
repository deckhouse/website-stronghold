---
title: "Динамические учётные данные MySQL и MariaDB для приложения"
linkTitle: "Динамические учётные данные MySQL и MariaDB"
description: "Выдача приложению временных пользователей MySQL или MariaDB с ограниченными правами через механизм секретов database: подключение, роль с GRANT, политика, продление, отзыв и проверка клиентом mysql."
weight: 20
params:
  relatedLinks:
    - title: "Механизм секретов баз данных"
      url: ../../../user/secrets-engines/databases/overview/
    - title: "MySQL"
      url: ../../../user/secrets-engines/databases/mysql-maria/
    - title: "Динамические учётные данные PostgreSQL"
      url: ../postgresql/
    - title: "Доставка секретов в поды Kubernetes"
      url: ../../delivery/kubernetes-workloads/
    - title: "Аренда, продление и отзыв"
      url: ../../../concepts/lease/
    - title: "API механизмов секретов"
      url: ../../../reference/api/secrets/
---

Приложение получает от Stronghold отдельного пользователя MySQL или MariaDB с правами только на свою базу данных. Stronghold создаёт пользователя при запросе учётных данных и удаляет его при отзыве или по истечении аренды. Постоянный пароль приложения не нужен.

![Схема выдачи динамических учётных данных MySQL](../../../images/ex-mysql.png)

## Цель

Настроить выдачу динамических учётных данных MySQL или MariaDB приложению `myapp`: пользователь получает права `SELECT`, `INSERT`, `UPDATE` и `DELETE` на базу данных `myapp`, живёт 1 час с возможностью продления до 24 часов и удаляется после отзыва аренды.

## Предварительные требования

- Stronghold и токен с правами на настройку механизмов секретов и политик.
- Сервер MySQL 5.7 и выше или MariaDB, доступный из Stronghold, и база данных `myapp`.
- Клиент `mysql` и утилита `jq` на рабочей станции для проверки.

Для MySQL 5.7.8 и новее и для MariaDB используйте `mysql-database-plugin` (имена пользователей до 32 символов), для MySQL 5.6 и старее — `mysql-legacy-database-plugin` (до 16 символов). Задайте `root_rotation_statements`, как описано в разделе [«MySQL»](../../../user/secrets-engines/databases/mysql-maria/).

## Шаг 1. Подготовьте служебную учётную запись в MySQL

Создайте учётную запись, от имени которой Stronghold будет создавать и удалять пользователей. Чтобы выдавать права на базу данных `myapp`, учётная запись должна сама иметь эти права с `GRANT OPTION`:

```sql
CREATE USER 'stronghold'@'%' IDENTIFIED BY '<initial_password>';
GRANT CREATE USER ON *.* TO 'stronghold'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON myapp.* TO 'stronghold'@'%' WITH GRANT OPTION;
```

Ограничьте хост `'%'` адресами узлов Stronghold, если это возможно.

## Шаг 2. Настройте подключение

1. Включите механизм секретов, если он ещё не включён:

   ```bash
   d8 stronghold secrets enable database
   ```

1. Настройте подключение к серверу:

   ```bash
   d8 stronghold write database/config/myapp-mysql \
     plugin_name="mysql-database-plugin" \
     connection_url="{{username}}:{{password}}@tcp(mysql.example.com:3306)/" \
     allowed_roles="myapp-mysql" \
     username="stronghold" \
     password="<initial_password>"
   ```

   Для подключения по TLS добавьте параметр `tls_ca=@/path/to/ca.pem`, а для аутентификации по клиентскому сертификату — `tls_certificate_key=@/path/to/client.pem`.

1. Смените пароль служебной учётной записи, чтобы он был известен только Stronghold:

   ```bash
   d8 stronghold write -force database/rotate-root/myapp-mysql
   ```

## Шаг 3. Создайте роль

Роль задаёт SQL-запросы для создания и удаления пользователя и сроки жизни учётных данных:

```bash
d8 stronghold write database/roles/myapp-mysql \
  db_name="myapp-mysql" \
  creation_statements="CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}'; \
    GRANT SELECT, INSERT, UPDATE, DELETE ON myapp.* TO '{{name}}'@'%';" \
  revocation_statements="REVOKE ALL PRIVILEGES, GRANT OPTION FROM '{{name}}'@'%'; DROP USER '{{name}}'@'%';" \
  default_ttl="1h" \
  max_ttl="24h"
```

- `creation_statements` — запросы, которые Stronghold выполняет при выдаче учётных данных. Шаблоны `{{name}}` и `{{password}}` заменяются сгенерированными значениями;
- `revocation_statements` — запросы, которые выполняются при отзыве аренды;
- `default_ttl` — срок аренды при выдаче и при каждом продлении;
- `max_ttl` — максимальный срок жизни учётных данных, после которого пользователь удаляется.

Если в запросе нужны шаблоны имён баз данных с обратными кавычками (например, ``GRANT SELECT ON `myapp\_%`.* ...``), передайте `creation_statements` в кодировке Base64, как описано в разделе [«MySQL»](../../../user/secrets-engines/databases/mysql-maria/#использование-шаблонов-в-grant-statements).

## Шаг 4. Создайте политику

```bash
d8 stronghold policy write myapp-mysql - <<'POLICY'
path "database/creds/myapp-mysql" {
  capabilities = ["read"]
}
POLICY
```

Назначьте политику роли метода аутентификации, через который входит приложение, например роли метода Kubernetes, как в руководстве [«Динамические учётные данные PostgreSQL»](../postgresql/#шаг-4-создайте-роль-метода-kubernetes). Продление и отзыв собственных аренд разрешены встроенной политикой `default`.

## Шаг 5. Получите учётные данные

1. Запросите учётные данные:

   ```bash
   d8 stronghold read database/creds/myapp-mysql
   ```

   Пример вывода:

   ```text
   Key                Value
   ---                -----
   lease_id           database/creds/myapp-mysql/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
   lease_duration     1h
   lease_renewable    true
   password           yY-57n3X5UQhxnmFRP3f
   username           v_token_myapp-mysql_crBWVqVh2Hc1
   ```

1. Продлите аренду до истечения `lease_duration`. Каждое продление увеличивает срок на `default_ttl`, но не дальше `max_ttl`:

   ```bash
   d8 stronghold lease renew database/creds/myapp-mysql/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
   ```

1. Когда учётные данные больше не нужны, отзовите аренду. Stronghold выполнит `revocation_statements` и удалит пользователя:

   ```bash
   d8 stronghold lease revoke database/creds/myapp-mysql/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
   ```

В рабочей среде продлением и повторным запросом учётных данных управляет Stronghold Agent. Пример шаблона для файла с переменными окружения:

```text
{{ with secret "database/creds/myapp-mysql" }}
DB_USER={{ .Data.username }}
DB_PASSWORD={{ .Data.password }}
{{ end }}
```

Конфигурация Agent в режиме sidecar приведена в разделе [«Доставка секретов в поды Kubernetes»](../../delivery/kubernetes-workloads/#stronghold-agent), для виртуальных машин — в разделе [«Приложение на ВМ»](../../delivery/legacy-app-on-vm/). Требования к приложению при смене учётных данных такие же, как в руководстве [«Динамические учётные данные PostgreSQL»](../postgresql/#шаг-6-настройте-реакцию-приложения-на-смену-учётных-данных).

## Проверка

1. Запросите учётные данные в формате JSON и сохраните их в переменные:

   ```bash
   creds="$(d8 stronghold read -format=json database/creds/myapp-mysql)"
   DB_USER="$(echo "$creds" | jq -r '.data.username')"
   DB_PASSWORD="$(echo "$creds" | jq -r '.data.password')"
   LEASE_ID="$(echo "$creds" | jq -r '.lease_id')"
   ```

1. Подключитесь клиентом `mysql` и проверьте права пользователя:

   ```bash
   mysql -h mysql.example.com -u "$DB_USER" -p"$DB_PASSWORD" myapp -e "SELECT CURRENT_USER(); SHOW GRANTS;"
   ```

   В выводе должны быть только права `SELECT, INSERT, UPDATE, DELETE ON myapp.*`.

1. Убедитесь, что доступа к другим базам данных нет:

   ```bash
   mysql -h mysql.example.com -u "$DB_USER" -p"$DB_PASSWORD" -e "SELECT * FROM mysql.user LIMIT 1;"
   ```

   Команда должна завершиться ошибкой `SELECT command denied`.

1. Отзовите аренду и проверьте, что пользователь удалён:

   ```bash
   d8 stronghold lease revoke "$LEASE_ID"
   mysql -h mysql.example.com -u "$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1;"
   ```

   Подключение должно завершиться ошибкой `Access denied`.

## Очистка

1. Отзовите все выданные учётные данные роли:

   ```bash
   d8 stronghold lease revoke -prefix database/creds/myapp-mysql/
   ```

1. Удалите роль, политику и подключение:

   ```bash
   d8 stronghold policy delete myapp-mysql
   d8 stronghold delete database/roles/myapp-mysql
   d8 stronghold delete database/config/myapp-mysql
   ```

1. При необходимости удалите служебную учётную запись в MySQL: `DROP USER 'stronghold'@'%';`.
