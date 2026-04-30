---
title: "MULTIFACTOR Ldap Adapter"
linkTitle: "MULTIFACTOR Ldap Adapter"
weight: 10
description: "Интеграция Deckhouse Stronghold с MULTIFACTOR Ldap Adapter для двухфакторной LDAP-аутентификации."
---

**MULTIFACTOR Ldap Adapter** — это LDAP proxy-сервер, разработанный и поддерживаемый компанией МУЛЬТИФАКТОР. Он используется для двухфакторной защиты пользователей в приложениях, использующих LDAP-аутентификацию.

Система обеспечивает многофакторную аутентификацию и контроль доступа для различных удалённых подключений, включая `RDP`, `VPN`, `VDI`, `SSH` и другие сценарии.

## Когда использовать

Этот сценарий подходит, если:

- в Stronghold уже используется `LDAP`-аутентификация;
- нужно добавить второй фактор без отказа от существующего LDAP-каталога;
- требуется централизованная двухфакторная проверка пользователей через инфраструктуру MULTIFACTOR.

## Как это работает

Stronghold может выполнять двухфакторную аутентификацию пользователей из LDAP или Active Directory по следующей схеме:

1. Пользователь подключается к Stronghold и вводит логин и пароль.
2. Stronghold по LDAP подключается к компоненту **MULTIFACTOR Ldap Adapter**.
3. Адаптер проверяет логин и пароль пользователя в Active Directory или другом LDAP-каталоге и запрашивает второй фактор аутентификации.
4. Пользователь подтверждает запрос доступа выбранным способом аутентификации.

Таким образом, для Stronghold адаптер выглядит как LDAP-сервер, но фактически добавляет второй фактор поверх обычной LDAP-проверки.

## Настройка MULTIFACTOR

### Шаг 1. Создайте LDAP-приложение в MULTIFACTOR

Зайдите в [систему управления MULTIFACTOR](https://admin.multifactor.ru/account/login). В разделе **Ресурсы** создайте новое LDAP-приложение.

После создания будут доступны два параметра:

- `NAS Identifier`;
- `Shared Secret`.

Они потребуются на следующих шагах.

### Шаг 2. Установите MULTIFACTOR Ldap Adapter

Загрузите и установите [MULTIFACTOR Ldap Adapter](https://multifactor.ru/docs/ldap-adapter/ldap-adapter/).

## Запуск LDAP Adapter в Kubernetes

Для запуска можно использовать образ `multifactor-ldap-adapter:3.0.7` и следующий манифест:

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ldap-adapter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ldap-adapter
  template:
    metadata:
      labels:
        app: ldap-adapter
    spec:
      containers:
      - image: registry.deckhouse.ru/stronghold/multifactor/multifactor-ldap-adapter:3.0.7
        name: ldap-adapter
        volumeMounts:
        - mountPath: /opt/multifactor/ldap/multifactor-ldap-adapter.dll.config
          name: config
          subPath: multifactor-ldap-adapter.dll.config
      volumes:
      - configMap:
          defaultMode: 420
          name: ldap-adapter
        name: config
---
apiVersion: v1
kind: Service
metadata:
  name: ldap-adapter
spec:
  ports:
  - port: 389
    protocol: TCP
    targetPort: 389
  selector:
    app: ldap-adapter
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: ldap-adapter
data:
  multifactor-ldap-adapter.dll.config: |
    <?xml version="1.0" encoding="utf-8"?>
    <configuration>
      <configSections>
        <section name="UserNameTransformRules" type="MultiFactor.Ldap.Adapter.Configuration.UserNameTransformRulesSection, multifactor-ldap-adapter" />
      </configSections>
      <appSettings>
        <add key="adapter-ldap-endpoint" value="0.0.0.0:389"/>
        <add key="ldap-server" value="ldap://ldap.example.com"/>
        <add key="ldap-service-accounts" value="CN=admin,DC=example,DC=com"/>
        <add key="ldap-base-dn" value="ou=Users,dc=example,dc=com"/>
        <add key="multifactor-api-url" value="https://api.multifactor.ru" />
        <add key="multifactor-nas-identifier" value="YOUR-NAS-IDENTIFIER" />
        <add key="multifactor-shared-secret" value="YOUR-NAS-SECRET" />
        <add key="logging-level" value="Debug"/>
      </appSettings>
    </configuration>
```

В конфигурации укажите:

- адрес вашего LDAP-сервера;
- `multifactor-nas-identifier`;
- `multifactor-shared-secret`.

Доступные образы:

- на базе Ubuntu 24.04: `registry.deckhouse.ru/stronghold/multifactor/multifactor-ldap-adapter:3.0.7`;
- на базе Alpine 3.22: `registry.deckhouse.ru/stronghold/multifactor/multifactor-ldap-adapter:3.0.7-alpine`.

## Настройка Stronghold

Для настройки Stronghold создайте и настройте метод аутентификации `ldap`, где в качестве сервера укажите адрес `ldap-adapter`. Если адаптер был запущен по примеру выше, используйте адрес:

```text
ldap://ldap-adapter.default.svc
```

Пример настройки:

```bash
d8 stronghold auth enable ldap
d8 stronghold write auth/ldap/config url="ldap://ldap-adapter.default.svc" \
   binddn="cn=admin,dc=example,dc=com" bindpass="Password-1" \
   userdn="ou=Users,dc=example,dc=com" groupdn="ou=Groups,dc=example,dc=com" \
   username_as_alias=true
```

После этого Stronghold будет обращаться не к LDAP напрямую, а к MULTIFACTOR Ldap Adapter, который выполнит первичную LDAP-проверку и запросит второй фактор.

## Тестирование с локальным OpenLDAP

Для тестирования можно запустить сервис `OpenLDAP` в Kubernetes.

Пример манифеста:

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openldap
spec:
  replicas: 1
  selector:
    matchLabels:
      app: openldap
  template:
    metadata:
      labels:
        app: openldap
    spec:
      containers:
      - env:
        - name: LDAP_ADMIN_DN
          value: cn=admin,dc=example,dc=com
        - name: LDAP_ROOT
          value: dc=example,dc=com
        - name: LDAP_ADMIN_USERNAME
          value: admin
        - name: LDAP_ADMIN_PASSWORD
          value: Password-1
        image: bitnami/openldap:2.6.10
        name: openldap
---
apiVersion: v1
kind: Service
metadata:
  name: openldap
spec:
  ports:
  - name: p389
    port: 389
    protocol: TCP
    targetPort: 1389
  selector:
    app: openldap
```

### Создание тестового пользователя

1. Войдите в контейнер OpenLDAP:

   ```bash
   d8 k exec svc/openldap -it -- bash
   ```

2. Создайте пользователя:

   ```bash
   cd /tmp
   cat << EOF > create_entries.ldif
   dn: uid=alice,ou=users,dc=example,dc=com
   objectClass: inetOrgPerson
   objectClass: person
   objectClass: top
   cn: Alice
   sn: User
   userPassword: D3mo-Passw0rd
   EOF

   ldapadd -H ldap://openldap -cxD "cn=admin,dc=example,dc=com" \
           -w "Password-1" -f "create_entries.ldif"
   ```

После этого можно выполнить вход под пользователем `alice` с паролем `D3mo-Passw0rd`. В панели управления MULTIFACTOR будет создан пользователь `alice`, для которого можно назначить второй фактор.

## Что дальше

После настройки и проверки входа пользователь будет проходить обычную LDAP-аутентификацию в Stronghold, а подтверждение второго фактора будет происходить на стороне MULTIFACTOR.
