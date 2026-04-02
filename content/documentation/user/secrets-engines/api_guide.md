---
title: "Deckhouse Stronghold administrator's API guide"
hidden: true
weight: 201
---

## Authentication methods

Each authentication method has its own set of API paths and methods described in this section. Authentication methods can be enabled at a custom path, but for simplicity this documentation assumes the default paths are used. If you enable a method at a different path, adjust your API requests accordingly.

### AppRole

This section assumes the method is enabled at the `/auth/approle` path.

#### List roles

This path returns a list of existing AppRoles for the method.

| Method | Path |
|-------|------|
| LIST  | /auth/approle/role |

Request example:

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  --request LIST \
    ${STRONGHOLD_ADDR}/v1/auth/approle/role
```

API response example:

```json
{
  "auth": null,
  "warnings": null,
  "wrap_info": null,
  "data": {
    "keys": ["dev", "prod", "test"]
  },
  "lease_duration": 0,
  "renewable": false,
  "lease_id": ""
}
```

#### Create or update AppRole

Creates a new AppRole or updates an existing one. This path supports both create and update operations. One or more constraints can be applied to a role. At least one of them must be enabled when creating or updating a role.

| Method | Path |
|-------|------|
| POST  | /auth/approle/role/:role_name |

Parameters:

- `role_name` `(string: <required>)` - AppRole name. Must be shorter than 4096 bytes. Valid characters include `A-Z`, `a-z`, `0-9`, spaces, hyphens, underscores, and dots.
- `bind_secret_id` `(bool: true)` - Requires `secret_id` to be provided when logging in with this AppRole.
- `secret_id_bound_cidrs` `(array: [])` - A comma-separated string or a list of CIDR blocks. The configured value specifies the IP address blocks allowed to perform the login operation.
- `secret_id_num_uses` `(int: 0)` - The number of times any specific SecretID can be used to obtain a token from this AppRole before the SecretID expires by default. Setting the value to zero allows unlimited use. However, this setting can be overridden by the request field `num_uses` when creating a SecretID.
- `secret_id_ttl` `(string: "")` - A duration expressed as an integer number of seconds (`3600`) or as a duration string (`60m`), after which any SecretID expires by default. Setting the value to zero allows the SecretID to never expire. However, this setting can be overridden by the request field `ttl` when creating a SecretID.
- `local_secret_ids` `(bool: false)` - If set, secret IDs created with this role are local to the cluster. This can only be set when the role is created and cannot be changed later.
- `token_ttl` `(int: 0 or string: "")` - Incremental lifetime for generated tokens. The current value of this parameter is used during renewal.
- `token_max_ttl` `(int: 0 or string: "")` - Maximum lifetime for generated tokens. The current value of this parameter is used during renewal.
- `token_policies` `(array: [] or comma-separated string: "")` - List of token policies that will be added to generated tokens. Depending on the authentication method, this list can be supplemented with user, group, or other values.
- `token_bound_cidrs` `(array: [] or comma-separated string: "")` - List of CIDR blocks. The configured value specifies the IP address blocks that can authenticate successfully and binds the resulting token to those blocks.
- `token_explicit_max_ttl` `(int: 0 or string: "")` - If set, an explicit maximum token lifetime is added. This is a hard limit even if `token_ttl` and `token_max_ttl` would otherwise allow renewal.
- `token_no_default_policy` `(bool: false)` - If set, the default policy is not applied to generated tokens. Otherwise, it is added to the policies specified in `token_policies`.
- `token_num_uses` `(int: 0)` - Maximum number of times a generated token can be used during its lifetime. `0` means unlimited use. If the token must be able to create child tokens, set this value to `0`.
- `token_period` `(int: 0 or string: "")` - Maximum allowed period when a periodic token is requested from this role.
- `token_type` `(string: "")` - Type of token to create. The value can be `service`, `batch`, or `default` to use the configured default value, which is `service` unless changed. Token store roles also support `default-service` and `default-batch`, which specify the returned type if the client does not request another type during creation. For machine-oriented authentication flows, use `batch` tokens.

Example data:

```json
{
  "token_type": "batch",
  "token_ttl": "10m",
  "token_max_ttl": "15m",
  "token_policies": ["default"],
  "period": 0,
  "bind_secret_id": true
}
```

Request example:

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  --request POST \
  --data @payload.json \
    ${STRONGHOLD_ADDR}/v1/auth/approle/role/application
```

API response example:

```shell
{
  "auth": null,
  "warnings": null,
  "wrap_info": null,
  "data": {
    "keys": ["dev", "prod", "test"]
  },
  "lease_duration": 0,
  "renewable": false,
  "lease_id": ""
}
```

#### Read AppRole

Displays the properties of an existing AppRole.

| Method   | Path                            |
| :----- | :------------------------------ |
| `GET`  | `/auth/approle/role/:role_name` |

Parameters:

- `role_name` `(string: <required>)` - AppRole name. Must be shorter than 4096 bytes.

Request example:

```shell
$ curl \
    --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
    ${STRONGHOLD_ADDR}/v1/auth/approle/role/application1
```

API response example:

```json
{
  "auth": null,
  "warnings": null,
  "wrap_info": null,
  "data": {
    "token_ttl": 1200,
    "token_max_ttl": 1800,
    "secret_id_ttl": 600,
    "secret_id_num_uses": 40,
    "token_policies": ["default"],
    "period": 0,
    "bind_secret_id": true,
    "secret_id_bound_cidrs": []
  },
  "lease_duration": 0,
  "renewable": false,
  "lease_id": ""
}
```

#### Delete AppRole

Deletes an existing AppRole from the method.

| Method   | Path                            |
| :------- | :------------------------------ |
| `DELETE` | `/auth/approle/role/:role_name` |

Parameters:

- `role_name` `(string: <required>)` - AppRole name. Must be shorter than 4096 bytes.

Request example:

```shell
$ curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  --request DELETE \
  ${STRONGHOLD_ADDR}/v1/auth/approle/role/application1
```

#### Read AppRole RoleID

Displays the RoleID of an existing AppRole.

| Method   | Path                                    |
| :----- | :-------------------------------------- |
| `GET`  | `/auth/approle/role/:role_name/role-id` |

Parameters:

- `role_name` `(string: <required>)` - AppRole name. Must be shorter than 4096 bytes.

Request example:

```shell
$ curl \
    --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
    ${STRONGHOLD_ADDR}/v1/auth/approle/role/application1/role-id
```

API response example:

```json
{
  "auth": null,
  "warnings": null,
  "wrap_info": null,
  "data": {
    "role_id": "e5a7b66e-5d08-da9c-7075-71984634b882"
  },
  "lease_duration": 0,
  "renewable": false,
  "lease_id": ""
}
```

#### Update AppRole RoleID

Updates the RoleID of an existing AppRole to a specified value.

| Method   | Path                                    |
| :----- | :-------------------------------------- |
| `POST` | `/auth/approle/role/:role_name/role-id` |

Parameters:

- `role_name` `(string: <required>)` - AppRole name. Must be shorter than 4096 bytes.
- `role_id` `(string: <required>)` - RoleID value.

Example data:

```json
{
  "role_id": "custom-role-id"
}
```

Request example:

```shell
$ curl \
    --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
    --request POST \
    --data @payload.json \
    ${STRONGHOLD_ADDR}/v1/auth/approle/role/application1/role-id
```

#### Generate a new SecretID

Generates and returns a new SecretID for an existing AppRole. Similar to tokens, the response also contains the `secret_id_accessor` value, which can be used to read secret properties without disclosing the identifier itself, and to delete the SecretID from the AppRole.

| Method   | Path                                      |
| :----- | :---------------------------------------- |
| `POST` | `/auth/approle/role/:role_name/secret-id` |

Parameters:

- `role_name` `(string: <required>)` - AppRole name. Must be shorter than 4096 bytes.
- `metadata` `(string: "")` - Metadata associated with the SecretID. This must be a JSON string containing key-value metadata pairs. This metadata is attached to tokens issued using this SecretID and is written to the audit log _in plaintext_.
- `cidr_list` `(array: [])` - A comma-separated string or a list of CIDR blocks restricting SecretID use to a specific set of IP addresses. If `secret_id_bound_cidrs` is set on the role, the CIDR blocks specified here must be a subset of the role's CIDR blocks.
- `token_bound_cidrs` `(array: [])` - A comma-separated string or a list of CIDR blocks. If set, it specifies the IP address blocks that can use authentication tokens created with this SecretID. It overrides the value configured on the role, but must be a subset.
- `num_uses` `(int: 0)` - The number of times this SecretID can be used before it expires. A value of zero allows unlimited use. Overrides the role's `secret_id_num_uses` parameter when specified.
  Cannot be greater than the role's `secret_id_num_uses`.
- `ttl` `(string: "")` - A duration in seconds (`3600`) or a duration string (`60m`) after which this SecretID expires. A value of zero allows the SecretID to never expire. Overrides the role's `secret_id_ttl` parameter when specified.
  Cannot be longer than the role's `secret_id_ttl`.

Example data:

```json
{
  "metadata": "{ \"tag1\": \"production\" }",
  "ttl": 600,
  "num_uses": 50
}
```

Request example:

```shell
$ curl \
    --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
    --request POST \
    --data @payload.json \
    ${STRONGHOLD_ADDR}/v1/auth/approle/role/application1/secret-id
```

API response example:

```json
{
  "auth": null,
  "warnings": null,
  "wrap_info": null,
  "data": {
    "secret_id_accessor": "84896a0c-1347-aa80-a3f6-aca8b7558780",
    "secret_id": "841771dc-11c9-bbc7-bcac-6a3945a69cd9",
    "secret_id_ttl": 600,
    "secret_id_num_uses": 50
  },
  "lease_duration": 0,
  "renewable": false,
  "lease_id": ""
}
```

#### List SecretID accessors

Displays the accessors of all issued SecretIDs for an AppRole. This also includes accessors for "custom" SecretIDs.

| Method   | Path                                     |
| :----- | :---------------------------------------- |
| `LIST` | `/auth/approle/role/:role_name/secret-id` |

Parameters:

- `role_name` `(string: <required>)` - AppRole name. Must be shorter than 4096 bytes.

Request example:

```shell
$ curl \
    --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
    --request LIST \
    ${STRONGHOLD_ADDR}/v1/auth/approle/role/application1/secret-id
```

API response example:

```json
{
  "auth": null,
  "warnings": null,
  "wrap_info": null,
  "data": {
    "keys": [
      "ce202d2a-8253-c437-bf9a-aceed4241491",
      "a1c0dee4-b869-e68d-3520-2040c1a0849a",
      "be03b7e2-044c-7244-07e1-47560ca1c787",
      "84896a0c-1347-aa80-a3f6-aca8b7558780",
      "439b1328-6523-15e7-403a-a48038cdc45a"
    ]
  },
  "lease_duration": 0,
  "renewable": false,
  "lease_id": ""
}
```

#### Read AppRole SecretID

Displays the properties of an AppRole SecretID.

| Method   | Path                                            |
| :----- | :----------------------------------------------- |
| `POST` | `/auth/approle/role/:role_name/secret-id/lookup` |

Parameters:

- `role_name` `(string: <required>)` - AppRole name. Must be shorter than 4096 bytes.
- `secret_id` `(string: <required>)` - SecretID bound to the role.

Example data:

```json
{
  "secret_id": "84896a0c-1347-aa80-a3f6-aca8b7558780"
}
```

Request example:

```shell
$ curl \
    --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
    --request POST \
    --data @payload.json \
    ${STRONGHOLD_ADDR}/v1/auth/approle/role/application1/secret-id/lookup
```

API response example:

```json
{
  "request_id": "74752925-f309-6859-3d2d-0fcded95150e",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "cidr_list": [],
    "creation_time": "2023-02-10T18:17:27.089757383Z",
    "expiration_time": "0001-01-01T00:00:00Z",
    "last_updated_time": "2023-02-10T18:17:27.089757383Z",
    "metadata": {
      "tag1": "production"
    },
    "secret_id_accessor": "2be760a4-87bb-2fa9-1637-1b7fa9ba2896",
    "secret_id_num_uses": 0,
    "secret_id_ttl": 0,
    "token_bound_cidrs": []
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

#### Destroy AppRole SecretID

Destroys an AppRole SecretID.

| Method   | Path                                              |
| :----- | :------------------------------------------------ |
| `POST` | `/auth/approle/role/:role_name/secret-id/destroy` |

Parameters:

- `role_name` `(string: <required>)` - AppRole name. Must be shorter than 4096 bytes.
- `secret_id` `(string: <required>)` - SecretID bound to the role.

Example data:

```json
{
  "secret_id": "84896a0c-1347-aa80-a3f6-aca8b7558780"
}
```

Request example:

```shell
$ curl \
    --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
    --request POST \
    --data @payload.json \
    ${STRONGHOLD_ADDR}/v1/auth/approle/role/application1/secret-id/destroy
```

#### Read AppRole SecretID accessor

Displays the properties of an AppRole SecretID accessor.

| Method   | Path                                                      |
| :----- | :-------------------------------------------------------- |
| `POST` | `/auth/approle/role/:role_name/secret-id-accessor/lookup` |

Parameters:

- `role_name` `(string: <required>)` - AppRole name. Must be shorter than 4096 bytes.
- `secret_id_accessor` `(string: <required>)` - SecretID accessor bound to the role.

Example data:

```json
{
  "secret_id_accessor": "84896a0c-1347-aa80-a3f6-aca8b7558780"
}
```

Request example:

```shell
$ curl \
    --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
    --request POST \
    --data @payload.json \
    ${STRONGHOLD_ADDR}/v1/auth/approle/role/application1/secret-id-accessor/lookup
```

API response example:

```json
{
  "request_id": "72836cd1-139c-fe66-1402-8bb5ca4044b8",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "cidr_list": [],
    "creation_time": "2023-02-10T18:17:27.089757383Z",
    "expiration_time": "0001-01-01T00:00:00Z",
    "last_updated_time": "2023-02-10T18:17:27.089757383Z",
    "metadata": {
      "tag1": "production"
    },
    "secret_id_accessor": "2be760a4-87bb-2fa9-1637-1b7fa9ba2896",
    "secret_id_num_uses": 0,
    "secret_id_ttl": 0,
    "token_bound_cidrs": []
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```

#### Destroy AppRole SecretID by accessor

Destroys an AppRole SecretID by its accessor.

| Method   | Path                                                       |
| :----- | :--------------------------------------------------------- |
| `POST` | `/auth/approle/role/:role_name/secret-id-accessor/destroy` |

Parameters:

- `role_name` `(string: <required>)` - AppRole name. Must be shorter than 4096 bytes.
- `secret_id_accessor` `(string: <required>)` - SecretID accessor bound to the role.

Example data:

```json
{
  "secret_id_accessor": "84896a0c-1347-aa80-a3f6-aca8b7558780"
}
```

Request example:

```shell
$ curl \
    --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
    --request POST \
    --data @payload.json \
    ${STRONGHOLD_ADDR}/v1/auth/approle/role/application1/secret-id-accessor/destroy
```

#### Create a custom AppRole SecretID

Assigns a "custom" SecretID to an existing AppRole. This is used in the "push" operation model.

| Method   | Path                                            |
| :----- | :----------------------------------------------- |
| `POST` | `/auth/approle/role/:role_name/custom-secret-id` |

Parameters:

- `role_name` `(string: <required>)` - AppRole name. Must be shorter than 4096 bytes.
- `secret_id` `(string: <required>)` - SecretID to bind to the role.
- `metadata` `(string: "")` - Metadata associated with the SecretID. This must be a JSON string containing key-value metadata pairs. This metadata is attached to tokens issued using this SecretID and is written to the audit log _in plaintext_.
- `cidr_list` `(array: [])` - A comma-separated string or a list of CIDR blocks restricting SecretID use to a specific set of IP addresses. If `secret_id_bound_cidrs` is set on the role, the CIDR blocks specified here must be a subset of the role's CIDR blocks.
- `token_bound_cidrs` `(array: [])` - A comma-separated string or a list of CIDR blocks. If set, it specifies the IP address blocks that can use authentication tokens created with this SecretID. It overrides the value configured on the role, but must be a subset.
- `num_uses` `(int: 0)` - The number of times this SecretID can be used before it expires. A value of zero allows unlimited use. Overrides the role's `secret_id_num_uses` parameter when specified.
  Cannot be greater than the role's `secret_id_num_uses`.
- `ttl` `(string: "")` - A duration in seconds (`3600`) or a duration string (`60m`) after which this SecretID expires. A value of zero allows the SecretID to never expire. Overrides the role's `secret_id_ttl` parameter when specified.
  Cannot be longer than the role's `secret_id_ttl`.

Example data:

```json
{
  "secret_id": "testsecretid",
  "ttl": 600,
  "num_uses": 50
}
```

Request example:

```shell
$ curl \
    --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
    --request POST \
    --data @payload.json \
    ${STRONGHOLD_ADDR}/v1/auth/approle/role/application1/custom-secret-id
```

API response example:

```json
{
  "auth": null,
  "warnings": null,
  "wrap_info": null,
  "data": {
    "secret_id": "testsecretid",
    "secret_id_accessor": "84896a0c-1347-aa80-a3f6-aca8b7558780",
    "secret_id_ttl": 600,
    "secret_id_num_uses": 50
  },
  "lease_duration": 0,
  "renewable": false,
  "lease_id": ""
}
```

#### Log in with AppRole

Issues a Stronghold token based on the provided credentials. `role_id` is always required. If `bind_secret_id` is enabled for the AppRole, which it is by default, `secret_id` is also required. Any other authentication values bound in the AppRole are also evaluated, for example the client's IP CIDR.

| Method   | Path                  |
| :----- | :-------------------- |
| `POST` | `/auth/approle/login` |

Parameters:

- `role_id` `(string: <required>)` - AppRole RoleID.
- `secret_id` `(string: <required>)` - SecretID belonging to the AppRole.

Example data:

```json
{
  "role_id": "59d6d1ca-47bb-4e7e-a40b-8be3bc5a0ba8",
  "secret_id": "84896a0c-1347-aa80-a3f6-aca8b7558780"
}
```

Request example:

```shell
$ curl \
    --request POST \
    --data @payload.json \
    ${STRONGHOLD_ADDR}/v1/auth/approle/login
```

API response example:

```json
{
  "auth": {
    "renewable": true,
    "lease_duration": 1200,
    "metadata": null,
    "token_policies": ["default"],
    "accessor": "fd6c9a00-d2dc-3b11-0be5-af7ae0e1d374",
    "client_token": "5b1a0318-679c-9c45-e5c6-d1b9a9035d49"
  },
  "warnings": null,
  "wrap_info": null,
  "data": null,
  "lease_duration": 0,
  "renewable": false,
  "lease_id": ""
}
```

#### Read, update, or delete AppRole properties

Updates the corresponding property in an existing AppRole. All of these AppRole parameters can also be updated through direct access to the `/auth/approle/role/:role_name` path. Separate paths are provided for each field so specific paths can be delegated through the Stronghold ACL system.

| Method            | Path                                                  | Code      |
| :---------------- | :---------------------------------------------------- | :-------- |
| `GET/POST/DELETE` | `/auth/approle/role/:role_name/policies`              | `200/204` |
| `GET/POST/DELETE` | `/auth/approle/role/:role_name/secret-id-num-uses`    | `200/204` |
| `GET/POST/DELETE` | `/auth/approle/role/:role_name/secret-id-ttl`         | `200/204` |
| `GET/POST/DELETE` | `/auth/approle/role/:role_name/token-ttl`             | `200/204` |
| `GET/POST/DELETE` | `/auth/approle/role/:role_name/token-max-ttl`         | `200/204` |
| `GET/POST/DELETE` | `/auth/approle/role/:role_name/bind-secret-id`        | `200/204` |
| `GET/POST/DELETE` | `/auth/approle/role/:role_name/secret-id-bound-cidrs` | `200/204` |
| `GET/POST/DELETE` | `/auth/approle/role/:role_name/token-bound-cidrs`     | `200/204` |
| `GET/POST/DELETE` | `/auth/approle/role/:role_name/period`                | `200/204` |

See the `/auth/approle/role/:role_name` path.

#### Tidy tokens

Performs maintenance tasks to clean up invalid entries that may remain in the token store. Normally you do not need to run this operation unless release notes or support instructions explicitly tell you to. It can result in a large amount of storage I/O, so use it with caution.

| Method   | Path                           |
| :----- | :----------------------------- |
| `POST` | `/auth/approle/tidy/secret-id` |

Request example:

```shell
$ curl \
    --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
    --request POST \
    ${STRONGHOLD_ADDR}/v1/auth/approle/tidy/secret-id
```

API response example:

```json
{
  "request_id": "b20b56e3-4699-5b19-cc6b-e74f7b787bbf",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": null,
  "wrap_info": null,
  "warnings": [
    "Tidy operation successfully started. Any information from the operation will be printed to Stronghold's server logs."
  ],
  "auth": null
}
```
