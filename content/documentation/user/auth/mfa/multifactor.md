---
title: "MULTIFACTOR LDAP Adapter"
linkTitle: "MULTIFACTOR LDAP Adapter"
weight: 10
description: "Integration of Deckhouse Stronghold with MULTIFACTOR LDAP Adapter for two-factor LDAP authentication."
---
**MULTIFACTOR LDAP Adapter** is an LDAP proxy server developed and maintained by MULTIFACTOR. It is used to provide two-factor protection for users in applications that use LDAP authentication. The system provides multifactor authentication and access control for remote connections, including `RDP`, `VPN`, `VDI`, `SSH`, and other scenarios.

## When to use

This scenario is suitable if:

- LDAP authentication is already used in Stronghold;
- you need to add a second factor without abandoning the existing LDAP directory;
- centralized two-factor user verification through the MULTIFACTOR infrastructure is required.

## How it works

Stronghold can perform two-factor authentication for users from LDAP or Active Directory as follows:

1. The user connects to Stronghold and enters a username and password.
1. Stronghold connects over LDAP to the **MULTIFACTOR LDAP Adapter** component.
1. The adapter verifies the user's username and password in Active Directory or another LDAP directory and requests the second authentication factor.
1. The user confirms the access request using the selected authentication method.

As a result, the adapter appears to Stronghold as an LDAP server, but actually adds a second factor on top of the standard LDAP verification.

## Configure MULTIFACTOR

### Step 1. Create an LDAP application in MULTIFACTOR

Sign in to the [MULTIFACTOR management system](https://admin.multifactor.ru/account/login). In the "Resources" section, create a new LDAP application.

After the application is created, the following parameters become available:

- `NAS Identifier`;
- `Shared Secret`.

You will need them in the following steps.

### Step 2. Install MULTIFACTOR LDAP Adapter

Download and install [MULTIFACTOR LDAP Adapter](https://multifactor.ru/docs/ldap-adapter/ldap-adapter/).

## Run LDAP Adapter in Kubernetes

To run the adapter, you can use the `multifactor-ldap-adapter:3.0.7` image
and the following manifest:

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
        - name: ldap-adapter
          image: registry.deckhouse.ru/stronghold/multifactor/multifactor-ldap-adapter:3.0.7
          volumeMounts:
            - mountPath: /opt/multifactor/ldap/multifactor-ldap-adapter.dll.config
              name: config
              subPath: multifactor-ldap-adapter.dll.config
      volumes:
        - name: config
          configMap:
            defaultMode: 420
            name: ldap-adapter
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

In the configuration, specify the following:

- the LDAP server address;
- `multifactor-nas-identifier`;
- `multifactor-shared-secret`.

Use the `multifactor-nas-identifier` and `multifactor-shared-secret` values from the MULTIFACTOR management panel.

The following images are available:

- based on Ubuntu 24.04:
  `registry.deckhouse.ru/stronghold/multifactor/multifactor-ldap-adapter:3.0.7`;
- based on Alpine Linux 3.22:
  `registry.deckhouse.ru/stronghold/multifactor/multifactor-ldap-adapter:3.0.7-alpine`.

{% alert level="info" %}
After you switch Stronghold to `ldap-adapter`, second-factor verification is performed on the MULTIFACTOR side. Make sure users can pass MFA verification before using this setup in a production environment.
{% endalert %}

## Configure Stronghold

To configure Stronghold, create and configure the `ldap` authentication method. Specify the `ldap-adapter` address as the server address.

If the adapter is started as shown above, use the following address:

```text
ldap://ldap-adapter.default.svc
```

Configuration example:

```shell
d8 stronghold auth enable ldap
d8 stronghold write auth/ldap/config \
  url="ldap://ldap-adapter.default.svc" \
  binddn="cn=admin,dc=example,dc=com" \
  bindpass="Password-1" \
  userdn="ou=Users,dc=example,dc=com" \
  groupdn="ou=Groups,dc=example,dc=com" \
  username_as_alias=true
```

After that, Stronghold no longer communicates with LDAP directly. Instead, it communicates with MULTIFACTOR LDAP Adapter, which performs the primary LDAP verification and requests the second factor.

## Testing with local OpenLDAP

For testing, you can run the `OpenLDAP` service in Kubernetes.

Example manifest:

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
        - name: openldap
          image: bitnami/openldap:2.6.10
          env:
            - name: LDAP_ADMIN_DN
              value: cn=admin,dc=example,dc=com
            - name: LDAP_ROOT
              value: dc=example,dc=com
            - name: LDAP_ADMIN_USERNAME
              value: admin
            - name: LDAP_ADMIN_PASSWORD
              value: Password-1
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

### Create a test user

Perform the following steps:

1. Sign in to the OpenLDAP container:

   ```shell
   d8 k exec svc/openldap -it -- bash
   ```

1. Create a user:

   ```shell
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

After that, you can sign in as user `alice` with password `D3mo-Passw0rd`.

In the [MULTIFACTOR management panel](https://admin.multifactor.ru/account/login), a user named `alice` is created in the "Users" section. You can assign a second factor to this user.

After the second factor is assigned, its confirmation is required each time the user signs in to Stronghold.

{% alert level="info" %}
After configuration and sign-in verification are complete, the user passes standard LDAP authentication in Stronghold, while second-factor confirmation is performed on the MULTIFACTOR side.
Second-factor confirmation is recorded both in the Stronghold audit logsand on the MULTIFACTOR side.
{% endalert %}
