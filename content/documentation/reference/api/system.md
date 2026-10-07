---
title: "system"
weight: 50
---

## system

### GET /sys/audit

**Operation ID:** `auditing-list-enabled-devices`

List the enabled audit devices.

**Required sudo:** yes

#### Responses

**200**: OK

### POST /sys/audit-hash/{path}

**Operation ID:** `auditing-calculate-hash`

The hash of the given string via the given audit backend

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | The name of the backend. Cannot be delimited. Example: "mysql" |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `input` | string | no |  |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `hash` | string | no |  |

### GET /sys/audit-monitor/{path}

**Operation ID:** `auditing-monitor-device`

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | The name of the backend. Cannot be delimited. Example: "mysql" |

#### Responses

**200**: OK

### POST /sys/audit/{path}

**Operation ID:** `auditing-enable-device`

Enable a new audit device at the supplied path.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | The name of the backend. Cannot be delimited. Example: "mysql" |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `description` | string | no | User-friendly description for this audit backend. |
| `local` | boolean (default: False) | no | Mark the mount as a local mount, which is not replicated and is unaffected by replication. |
| `options` | object | no | Configuration options for the audit backend. |
| `type` | string | no | The type of the backend. Example: "mysql" |

#### Responses

**204**: OK

### DELETE /sys/audit/{path}

**Operation ID:** `auditing-disable-device`

Disable the audit device at the given path.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | The name of the backend. Cannot be delimited. Example: "mysql" |

#### Responses

**204**: OK

### GET /sys/auth

**Operation ID:** `auth-list-enabled-methods`

List the currently enabled credential backends.

#### Responses

**200**: OK

### GET /sys/auth/{path}

**Operation ID:** `auth-read-configuration`

Read the configuration of the auth engine at the given path.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | The path to mount to. Cannot be delimited. Example: "user" |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `accessor` | string | no |  |
| `config` | object | no |  |
| `deprecation_status` | string | no |  |
| `description` | string | no |  |
| `external_entropy_access` | boolean | no |  |
| `local` | boolean | no |  |
| `options` | object | no |  |
| `plugin_version` | string | no |  |
| `running_plugin_version` | string | no |  |
| `running_sha256` | string | no |  |
| `seal_wrap` | boolean | no |  |
| `type` | string | no |  |
| `uuid` | string | no |  |

### POST /sys/auth/{path}

**Operation ID:** `auth-enable-method`

Enables a new auth method.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | The path to mount to. Cannot be delimited. Example: "user" |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `config` | object | no | Configuration for this mount, such as plugin_name. |
| `description` | string | no | User-friendly description for this credential backend. |
| `external_entropy_access` | boolean (default: False) | no | Whether to give the mount access to Stronghold's external entropy. |
| `local` | boolean (default: False) | no | Mark the mount as a local mount, which is not replicated and is unaffected by replication. |
| `options` | object | no | The options to pass into the backend. Should be a json object with string keys and values. |
| `plugin_name` | string | no | Name of the auth plugin to use based from the name in the plugin catalog. |
| `plugin_version` | string | no | The semantic version of the plugin to use. |
| `seal_wrap` | boolean (default: False) | no | Whether to turn on seal wrapping for the mount. |
| `type` | string | no | The type of the backend. Example: "userpass" |

#### Responses

**204**: OK

### DELETE /sys/auth/{path}

**Operation ID:** `auth-disable-method`

Disable the auth method at the given auth path

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | The path to mount to. Cannot be delimited. Example: "user" |

#### Responses

**204**: OK

### GET /sys/auth/{path}/tune

**Operation ID:** `auth-read-tuning-information`

Reads the given auth path's configuration.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | Tune the configuration parameters for an auth path. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `allowed_managed_keys` | array | no |  |
| `allowed_response_headers` | array | no |  |
| `audit_non_hmac_request_keys` | array | no |  |
| `audit_non_hmac_response_keys` | array | no |  |
| `default_lease_ttl` | integer | no |  |
| `description` | string | no |  |
| `external_entropy_access` | boolean | no |  |
| `force_no_cache` | boolean | no |  |
| `listing_visibility` | string | no |  |
| `max_lease_ttl` | integer | no |  |
| `options` | object | no |  |
| `passthrough_request_headers` | array | no |  |
| `plugin_version` | string | no |  |
| `token_type` | string | no |  |
| `user_lockout_counter_reset_duration` | integer | no |  |
| `user_lockout_disable` | boolean | no |  |
| `user_lockout_duration` | integer | no |  |
| `user_lockout_threshold` | integer | no |  |

### POST /sys/auth/{path}/tune

**Operation ID:** `auth-tune-configuration-parameters`

Tune configuration parameters for a given auth path.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | Tune the configuration parameters for an auth path. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `allowed_response_headers` | array | no | A list of headers to whitelist and allow a plugin to set on responses. |
| `audit_non_hmac_request_keys` | array | no | The list of keys in the request data object that will not be HMAC'ed by audit devices. |
| `audit_non_hmac_response_keys` | array | no | The list of keys in the response data object that will not be HMAC'ed by audit devices. |
| `default_lease_ttl` | string | no | The default lease TTL for this mount. |
| `description` | string | no | User-friendly description for this credential backend. |
| `listing_visibility` | string | no | Determines the visibility of the mount in the UI-specific listing endpoint. Accepted value are 'unauth' and 'hidden', with the empty default ('') behaving like 'hidden'. |
| `max_lease_ttl` | string | no | The max lease TTL for this mount. |
| `options` | object | no | The options to pass into the backend. Should be a json object with string keys and values. |
| `passthrough_request_headers` | array | no | A list of headers to whitelist and pass from the request to the plugin. |
| `plugin_version` | string | no | The semantic version of the plugin to use. |
| `token_type` | string | no | The type of token to issue (service or batch). |
| `user_lockout_config` | object | no | The user lockout configuration to pass into the backend. Should be a json object with string keys and values. |

#### Responses

**204**: OK

### POST /sys/capabilities

**Operation ID:** `query-token-capabilities`

Fetches the capabilities of the given token on the given path.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `path` | array | no | ⚠️ Deprecated. Use 'paths' instead. |
| `paths` | array | no | Paths on which capabilities are being queried. |
| `token` | string | no | Token for which capabilities are being queried. |

#### Responses

**200**: OK

### POST /sys/capabilities-accessor

**Operation ID:** `query-token-accessor-capabilities`

Fetches the capabilities of the token associated with the given token, on the given path.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `accessor` | string | no | Accessor of the token for which capabilities are being queried. |
| `path` | array | no | ⚠️ Deprecated. Use 'paths' instead. |
| `paths` | array | no | Paths on which capabilities are being queried. |

#### Responses

**200**: OK

### POST /sys/capabilities-self

**Operation ID:** `query-token-self-capabilities`

Fetches the capabilities of the given token on the given path.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `path` | array | no | ⚠️ Deprecated. Use 'paths' instead. |
| `paths` | array | no | Paths on which capabilities are being queried. |
| `token` | string | no | Token for which capabilities are being queried. |

#### Responses

**200**: OK

### GET /sys/config/auditing/request-headers

**Operation ID:** `auditing-list-request-headers`

List the request headers that are configured to be audited.

**Required sudo:** yes

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `headers` | object | no |  |

### GET /sys/config/auditing/request-headers/{header}

**Operation ID:** `auditing-read-request-header-information`

List the information for the given request header.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `header` | string | path | yes |  |

#### Responses

**200**: OK

### POST /sys/config/auditing/request-headers/{header}

**Operation ID:** `auditing-enable-request-header`

Enable auditing of a header.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `header` | string | path | yes |  |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `hmac` | boolean | no |  |

#### Responses

**204**: OK

### DELETE /sys/config/auditing/request-headers/{header}

**Operation ID:** `auditing-disable-request-header`

Disable auditing of the given request header.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `header` | string | path | yes |  |

#### Responses

**204**: OK

### GET /sys/config/control-group

**Operation ID:** `system-read-config-control-group`

Configure control group global settings.

#### Responses

**200**: OK

### POST /sys/config/control-group

**Operation ID:** `system-write-config-control-group`

Configure control group global settings.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `max_ttl` | integer | no | The max TTL for a control group token. |

#### Responses

**200**: OK

### DELETE /sys/config/control-group

**Operation ID:** `system-delete-config-control-group`

Configure control group global settings.

#### Responses

**204**: empty body

### GET /sys/config/cors

**Operation ID:** `cors-read-configuration`

Return the current CORS settings.

**Required sudo:** yes

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `allowed_headers` | array | no |  |
| `allowed_origins` | array | no |  |
| `enabled` | boolean | no |  |

### POST /sys/config/cors

**Operation ID:** `cors-configure`

Configure the CORS settings.

**Required sudo:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `allowed_headers` | array | no | A comma-separated string or array of strings indicating headers that are allowed on cross-origin requests. |
| `allowed_origins` | array | no | A comma-separated string or array of strings indicating origins that may make cross-origin requests. |
| `enable` | boolean | no | Enables or disables CORS headers on requests. |

#### Responses

**204**: OK

### DELETE /sys/config/cors

**Operation ID:** `cors-delete-configuration`

Remove any CORS settings.

**Required sudo:** yes

#### Responses

**204**: OK

### GET /sys/config/group-policy-application

**Operation ID:** `system-read-config-group-policy-application`

Gets the current status of the policy application mode for Stronghold.

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `group_policy_application_mode` | string | no | The current-set group_policy_application_mode. Will be either within_namespace_hierarchy or any. |

### POST /sys/config/group-policy-application

**Operation ID:** `system-write-config-group-policy-application`

Updates the policy application mode for Stronghold.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `group_policy_application_mode` | string | no | Configures how policies in groups should be applied, accepting 'within_namespace_hierarchy' (default) and 'any',which will allow policies to grant permissions in groups outside of those sharing a namespace hierarchy. |

#### Responses

**200**: OK

### POST /sys/config/reload/{subsystem}

**Operation ID:** `reload-subsystem`

Reload the given subsystem

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `subsystem` | string | path | yes |  |

#### Responses

**204**: OK

### GET /sys/config/state/sanitized

**Operation ID:** `read-sanitized-configuration-state`

Return a sanitized version of the Stronghold server configuration.

#### Responses

**200**: OK

### GET /sys/config/ui/headers

**Operation ID:** `ui-headers-list`

Return a list of configured UI headers.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `list` | string (true) | query | yes | Must be set to `true` |

#### Responses

**200**:

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `keys` | array | no | Lists of configured UI headers. Omitted if list is empty |

### GET /sys/config/ui/headers/{header}

**Operation ID:** `ui-headers-read-configuration`

Return the given UI header's configuration

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `header` | string | path | yes | The name of the header. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `value` | string | no | returns the first header value when `multivalue` request parameter is false |
| `values` | array | no | returns all header values when `multivalue` request parameter is true |

### POST /sys/config/ui/headers/{header}

**Operation ID:** `ui-headers-configure`

Configure the values to be returned for the UI header.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `header` | string | path | yes | The name of the header. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `multivalue` | boolean | no | Returns multiple values if true |
| `values` | array | no | The values to set the header. |

#### Responses

**200**: OK

### DELETE /sys/config/ui/headers/{header}

**Operation ID:** `ui-headers-delete-configuration`

Remove a UI header.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `header` | string | path | yes | The name of the header. |

#### Responses

**204**: OK

### POST /sys/control-group/authorize

**Operation ID:** `system-write-control-group-authorize`

Authorize a control group request

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `accessor` | string | no | The accessor of the request to authorize. |

#### Responses

**200**: OK

### POST /sys/control-group/request

**Operation ID:** `system-write-control-group-request`

Check the status of a control group request

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `accessor` | string | no | The accessor of the request. |

#### Responses

**200**: OK

### POST /sys/decode-token

**Operation ID:** `decode`

Decodes the encoded token with the otp.

**Available without authentication:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `encoded_token` | string | no | Specifies the encoded token (result from generate-root). |
| `otp` | string | no | Specifies the otp code for decode. |

#### Responses

**200**: OK

### GET /sys/experiments

**Operation ID:** `list-experimental-features`

Returns the available and enabled experiments

#### Responses

**200**: OK

### GET /sys/generate-root

**Operation ID:** `root-token-generation-read-progress2`

Read the configuration and progress of the current root generation attempt.

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `complete` | boolean | no |  |
| `encoded_root_token` | string | no |  |
| `encoded_token` | string | no |  |
| `nonce` | string | no |  |
| `otp` | string | no |  |
| `otp_length` | integer | no |  |
| `pgp_fingerprint` | string | no |  |
| `progress` | integer | no |  |
| `required` | integer | no |  |
| `started` | boolean | no |  |

### POST /sys/generate-root

**Operation ID:** `root-token-generation-initialize-2`

Initializes a new root generation attempt.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `otp` | string | no | Specifies a one-time password used to encrypt the generated token. |
| `pgp_key` | string | no | Specifies a base64-encoded PGP public key. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `complete` | boolean | no |  |
| `encoded_root_token` | string | no |  |
| `encoded_token` | string | no |  |
| `nonce` | string | no |  |
| `otp` | string | no |  |
| `otp_length` | integer | no |  |
| `pgp_fingerprint` | string | no |  |
| `progress` | integer | no |  |
| `required` | integer | no |  |
| `started` | boolean | no |  |

### DELETE /sys/generate-root

**Operation ID:** `root-token-generation-cancel-2`

Cancels any in-progress root generation attempt.

#### Responses

**204**: OK

### GET /sys/generate-root/attempt

**Operation ID:** `root-token-generation-read-progress`

Read the configuration and progress of the current root generation attempt.

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `complete` | boolean | no |  |
| `encoded_root_token` | string | no |  |
| `encoded_token` | string | no |  |
| `nonce` | string | no |  |
| `otp` | string | no |  |
| `otp_length` | integer | no |  |
| `pgp_fingerprint` | string | no |  |
| `progress` | integer | no |  |
| `required` | integer | no |  |
| `started` | boolean | no |  |

### POST /sys/generate-root/attempt

**Operation ID:** `root-token-generation-initialize`

Initializes a new root generation attempt.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `otp` | string | no | Specifies a one-time password used to encrypt the generated token. |
| `pgp_key` | string | no | Specifies a base64-encoded PGP public key. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `complete` | boolean | no |  |
| `encoded_root_token` | string | no |  |
| `encoded_token` | string | no |  |
| `nonce` | string | no |  |
| `otp` | string | no |  |
| `otp_length` | integer | no |  |
| `pgp_fingerprint` | string | no |  |
| `progress` | integer | no |  |
| `required` | integer | no |  |
| `started` | boolean | no |  |

### DELETE /sys/generate-root/attempt

**Operation ID:** `root-token-generation-cancel`

Cancels any in-progress root generation attempt.

#### Responses

**204**: OK

### POST /sys/generate-root/update

**Operation ID:** `root-token-generation-update`

Enter a single unseal key share to progress the root generation attempt.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `key` | string | no | Specifies a single unseal key share. |
| `nonce` | string | no | Specifies the nonce of the attempt. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `complete` | boolean | no |  |
| `encoded_root_token` | string | no |  |
| `encoded_token` | string | no |  |
| `nonce` | string | no |  |
| `otp` | string | no |  |
| `otp_length` | integer | no |  |
| `pgp_fingerprint` | string | no |  |
| `progress` | integer | no |  |
| `required` | integer | no |  |
| `started` | boolean | no |  |

### GET /sys/ha-status

**Operation ID:** `ha-status`

Check the HA status of a Stronghold cluster

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `nodes` | array | no |  |

### GET /sys/health

**Operation ID:** `read-health-status`

Returns the health status of Stronghold.

**Available without authentication:** yes

#### Responses

**200**: initialized, unsealed, and active

**423**: isleaderreadyok set but the local node is not yet a ready HA leader (or cannot reach one)

**425**: raftautopilotok set but the local node has not yet been integrated into the cluster by autopilot

**429**: unsealed and standby

**472**: data recovery mode replication secondary and active

**501**: not initialized

**503**: sealed

### GET /sys/host-info

**Operation ID:** `collect-host-information`

Information about the host instance that this Stronghold server is running on.

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `cpu` | array | no |  |
| `cpu_times` | array | no |  |
| `disk` | array | no |  |
| `host` | object | no |  |
| `memory` | object | no |  |
| `timestamp` | string | no |  |

### GET /sys/in-flight-req

**Operation ID:** `collect-in-flight-request-information`

reports in-flight requests

#### Responses

**200**: OK

### GET /sys/init

**Operation ID:** `read-initialization-status`

Returns the initialization status of Stronghold.

**Available without authentication:** yes

#### Responses

**200**: OK

### POST /sys/init

**Operation ID:** `initialize`

Initialize a new Stronghold.

**Available without authentication:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `pgp_keys` | array | no | Specifies an array of PGP public keys used to encrypt the output unseal keys. Ordering is preserved. The keys must be base64-encoded from their original binary representation. The size of this array must be the same as `secret_shares`. |
| `recovery_pgp_keys` | array | no | Specifies an array of PGP public keys used to encrypt the output recovery keys. Ordering is preserved. The keys must be base64-encoded from their original binary representation. The size of this array must be the same as `recovery_shares`. |
| `recovery_shares` | integer | no | Specifies the number of shares to split the recovery key into. |
| `recovery_threshold` | integer | no | Specifies the number of shares required to reconstruct the recovery key. This must be less than or equal to `recovery_shares`. |
| `root_token_pgp_key` | string | no | Specifies a PGP public key used to encrypt the initial root token. The key must be base64-encoded from its original binary representation. |
| `secret_shares` | integer | no | Specifies the number of shares to split the unseal key into. |
| `secret_threshold` | integer | no | Specifies the number of shares required to reconstruct the unseal key. This must be less than or equal secret_shares. If using Stronghold HSM with auto-unsealing, this value must be the same as `secret_shares`. |
| `stored_shares` | integer | no | Specifies the number of shares that should be encrypted by the HSM and stored for auto-unsealing. Currently must be the same as `secret_shares`. |

#### Responses

**200**: OK

### GET /sys/internal/counters/activity

**Operation ID:** `internal-client-activity-report-counts`

Report the client count metrics, for this namespace and all child namespaces.

#### Responses

**200**: OK

### GET /sys/internal/counters/activity/export

**Operation ID:** `internal-client-activity-export`

Report the client count metrics, for this namespace and all child namespaces.

#### Responses

**200**: OK

### GET /sys/internal/counters/activity/monthly

**Operation ID:** `internal-client-activity-report-counts-this-month`

Report the number of clients for this month, for this namespace and all child namespaces.

#### Responses

**200**: OK

### GET /sys/internal/counters/config

**Operation ID:** `internal-client-activity-read-configuration`

Read the client count tracking configuration.

#### Responses

**200**: OK

### POST /sys/internal/counters/config

**Operation ID:** `internal-client-activity-configure`

Enable or disable collection of client count, set retention period, or set default reporting period.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `default_report_months` | integer (default: 12) | no | Number of months to report if no start date specified. |
| `enabled` | string (default: default) | no | Enable or disable collection of client count: enable, disable, or default. |
| `retention_months` | integer (default: 24) | no | Number of months of client data to retain. Setting to 0 will clear all existing data. |

#### Responses

**200**: OK

### GET /sys/internal/counters/entities

**Operation ID:** `internal-count-entities`

Backwards compatibility is not guaranteed for this API

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `counters` | object | no |  |

### GET /sys/internal/counters/requests

**Operation ID:** `internal-count-requests`

Backwards compatibility is not guaranteed for this API

#### Responses

**200**: OK

### GET /sys/internal/counters/tokens

**Operation ID:** `internal-count-tokens`

Backwards compatibility is not guaranteed for this API

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `counters` | object | no |  |

### GET /sys/internal/inspect/router/{tag}

**Operation ID:** `internal-inspect-router`

Expose the route entry and mount entry tables present in the router

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `tag` | string | path | yes | Name of subtree being observed |

#### Responses

**200**: OK

### GET /sys/internal/specs/openapi

**Operation ID:** `internal-generate-open-api-document`

Generate an OpenAPI 3 document of all mounted paths.

**Available without authentication:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `generic_mount_paths` | boolean | query | no | Use generic mount paths |

#### Responses

**200**: OK

### POST /sys/internal/specs/openapi

**Operation ID:** `internal-generate-open-api-document-with-parameters`

Generate an OpenAPI 3 document of all mounted paths.

**Available without authentication:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `generic_mount_paths` | boolean | query | no | Use generic mount paths |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `context` | string | no | Context string appended to every operationId |

#### Responses

**200**: OK

### GET /sys/internal/ui/feature-flags

**Operation ID:** `internal-ui-list-enabled-feature-flags`

Lists enabled feature flags.

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `feature_flags` | array | no |  |

### GET /sys/internal/ui/mounts

**Operation ID:** `internal-ui-list-enabled-visible-mounts`

Lists all enabled and visible auth and secrets mounts.

**Available without authentication:** yes

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `auth` | object | no | auth mounts |
| `secret` | object | no | secret mounts |

### GET /sys/internal/ui/mounts/{path}

**Operation ID:** `internal-ui-read-mount-information`

Return information about the given mount.

**Available without authentication:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | The path of the mount. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `accessor` | string | no |  |
| `config` | object | no |  |
| `description` | string | no |  |
| `external_entropy_access` | boolean | no |  |
| `local` | boolean | no |  |
| `options` | object | no |  |
| `path` | string | no |  |
| `plugin_version` | string | no |  |
| `running_plugin_version` | string | no |  |
| `running_sha256` | string | no |  |
| `seal_wrap` | boolean | no |  |
| `type` | string | no |  |
| `uuid` | string | no |  |

### GET /sys/internal/ui/namespaces

**Operation ID:** `internal-ui-list-namespaces`

Backwards compatibility is not guaranteed for this API

**Available without authentication:** yes

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `keys` | array | no | field is only returned if there are one or more namespaces |

### GET /sys/internal/ui/resultant-acl

**Operation ID:** `internal-ui-read-resultant-acl`

Backwards compatibility is not guaranteed for this API

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `exact_paths` | object | no |  |
| `glob_paths` | object | no |  |
| `root` | boolean | no |  |

**204**: empty response returned if no client token

### GET /sys/key-status

**Operation ID:** `encryption-key-status`

Provides information about the backend encryption key.

#### Responses

**200**: OK

### GET /sys/leader

**Operation ID:** `leader-status`

Returns the high availability status and current leader instance of Stronghold.

**Available without authentication:** yes

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `active_time` | string | no |  |
| `ha_enabled` | boolean | no |  |
| `is_self` | boolean | no |  |
| `last_wal` | integer | no |  |
| `leader_address` | string | no |  |
| `leader_cluster_address` | string | no |  |
| `performance_standby` | boolean | no |  |
| `performance_standby_last_remote_wal` | integer | no |  |
| `raft_applied_index` | integer | no |  |
| `raft_committed_index` | integer | no |  |

### GET /sys/leases

**Operation ID:** `leases-list`

List leases associated with this Stronghold cluster

**Required sudo:** yes

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `counts` | integer | no | Number of matching leases per mount |
| `lease_count` | integer | no | Number of matching leases |

### GET /sys/leases/count

**Operation ID:** `leases-count`

Count of leases associated with this Stronghold cluster

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `counts` | integer | no | Number of matching leases per mount |
| `lease_count` | integer | no | Number of matching leases |

### POST /sys/leases/lookup

**Operation ID:** `leases-read-lease`

View or list lease metadata.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `lease_id` | string | no | The lease identifier to renew. This is included with a lease. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `expire_time` | string | no | Optional lease expiry time |
| `id` | string | no | Lease id |
| `issue_time` | string | no | Timestamp for the lease's issue time |
| `last_renewal` | string | no | Optional Timestamp of the last time the lease was renewed |
| `renewable` | boolean | no | True if the lease is able to be renewed |
| `ttl` | integer | no | Time to Live set for the lease, returns 0 if unset |

### GET /sys/leases/lookup/

**Operation ID:** `leases-look-up`

View or list lease metadata.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `list` | string (true) | query | yes | Must be set to `true` |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `keys` | array | no | A list of lease ids |

### GET /sys/leases/lookup/{prefix}

**Operation ID:** `leases-look-up-with-prefix`

View or list lease metadata.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `prefix` | string | path | yes | The path to list leases under. Example: "aws/creds/deploy" |
| `list` | string (true) | query | yes | Must be set to `true` |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `keys` | array | no | A list of lease ids |

### POST /sys/leases/renew

**Operation ID:** `leases-renew-lease`

Renews a lease, requesting to extend the lease.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `increment` | integer | no | The desired increment in seconds to the lease |
| `lease_id` | string | no | The lease identifier to renew. This is included with a lease. |
| `url_lease_id` | string | no | The lease identifier to renew. This is included with a lease. |

#### Responses

**204**: OK

### POST /sys/leases/renew/{url_lease_id}

**Operation ID:** `leases-renew-lease-with-id`

Renews a lease, requesting to extend the lease.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `url_lease_id` | string | path | yes | The lease identifier to renew. This is included with a lease. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `increment` | integer | no | The desired increment in seconds to the lease |
| `lease_id` | string | no | The lease identifier to renew. This is included with a lease. |

#### Responses

**204**: OK

### POST /sys/leases/revoke

**Operation ID:** `leases-revoke-lease`

Revokes a lease immediately.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `lease_id` | string | no | The lease identifier to renew. This is included with a lease. |
| `sync` | boolean (default: True) | no | Whether or not to perform the revocation synchronously |
| `url_lease_id` | string | no | The lease identifier to renew. This is included with a lease. |

#### Responses

**204**: OK

### POST /sys/leases/revoke-force/{prefix}

**Operation ID:** `leases-force-revoke-lease-with-prefix`

Revokes all secrets or tokens generated under a given prefix immediately

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `prefix` | string | path | yes | The path to revoke keys under. Example: "prod/aws/ops" |

#### Responses

**204**: OK

### POST /sys/leases/revoke-prefix/{prefix}

**Operation ID:** `leases-revoke-lease-with-prefix`

Revokes all secrets (via a lease ID prefix) or tokens (via the tokens' path property) generated under a given prefix immediately.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `prefix` | string | path | yes | The path to revoke keys under. Example: "prod/aws/ops" |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `sync` | boolean (default: True) | no | Whether or not to perform the revocation synchronously |

#### Responses

**204**: OK

### POST /sys/leases/revoke/{url_lease_id}

**Operation ID:** `leases-revoke-lease-with-id`

Revokes a lease immediately.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `url_lease_id` | string | path | yes | The lease identifier to renew. This is included with a lease. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `lease_id` | string | no | The lease identifier to renew. This is included with a lease. |
| `sync` | boolean (default: True) | no | Whether or not to perform the revocation synchronously |

#### Responses

**204**: OK

### POST /sys/leases/tidy

**Operation ID:** `leases-tidy`

This endpoint performs cleanup tasks that can be run if certain error conditions have occurred.

#### Responses

**204**: OK

### GET /sys/license/status

**Operation ID:** `system-read-license-status`

The path responds to the following HTTP methods.

    GET /
        Returns information on the installed license

    POST
        Sets the license for the server

#### Responses

**200**: OK

### GET /sys/locked-users

**Operation ID:** `locked-users-list`

Report the locked user count metrics, for this namespace and all child namespaces.

#### Responses

**200**: OK

### POST /sys/locked-users/{mount_accessor}/unlock/{alias_identifier}

**Operation ID:** `locked-users-unlock`

Unlocks the user with given mount_accessor and alias_identifier

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `alias_identifier` | string | path | yes | It is the name of the alias (user). For example, if the alias belongs to userpass backend, the name should be a valid username within userpass auth method. If the alias belongs to an approle auth method, the name should be a valid RoleID |
| `mount_accessor` | string | path | yes | MountAccessor is the identifier of the mount entry to which the user belongs |

#### Responses

**200**: OK

### GET /sys/loggers

**Operation ID:** `loggers-read-verbosity-level`

Read the log level for all existing loggers.

#### Responses

**200**: OK

### POST /sys/loggers

**Operation ID:** `loggers-update-verbosity-level`

Modify the log level for all existing loggers.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `level` | string | no | Log verbosity level. Supported values (in order of detail) are "trace", "debug", "info", "warn", and "error". |

#### Responses

**204**: OK

### DELETE /sys/loggers

**Operation ID:** `loggers-revert-verbosity-level`

Revert the all loggers to use log level provided in config.

#### Responses

**204**: OK

### GET /sys/loggers/{name}

**Operation ID:** `loggers-read-verbosity-level-for`

Read the log level for a single logger.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the logger to be modified. |

#### Responses

**200**: OK

### POST /sys/loggers/{name}

**Operation ID:** `loggers-update-verbosity-level-for`

Modify the log level of a single logger.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the logger to be modified. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `level` | string | no | Log verbosity level. Supported values (in order of detail) are "trace", "debug", "info", "warn", and "error". |

#### Responses

**204**: OK

### DELETE /sys/loggers/{name}

**Operation ID:** `loggers-revert-verbosity-level-for`

Revert a single logger to use log level provided in config.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the logger to be modified. |

#### Responses

**204**: OK

### GET /sys/managed-keys/{type}

**Operation ID:** `system-list-managed-keys-type`

List configured managed keys.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `type` | string | path | yes | The type of the managed key. |
| `list` | string (true) | query | yes | Must be set to `true` |

#### Responses

**200**: OK

### GET /sys/managed-keys/{type}/{name}

**Operation ID:** `system-read-managed-keys-type-name`

Create, update and read the managed key configurations.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the managed key. |
| `type` | string | path | yes | The type of the managed key. |

#### Responses

**200**: OK

### POST /sys/managed-keys/{type}/{name}

**Operation ID:** `system-write-managed-keys-type-name`

Create, update and read the managed key configurations.

**Creation supported:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the managed key. |
| `type` | string | path | yes | The type of the managed key. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `access_key` | string | no | Access key for authenticating to AWS |
| `algorithm` | string | no | The signature algorithm to use with this key |
| `allow_generate_key` | boolean | no | If true, allows users of the key to trigger key generation. If false and generation is needed, it will fail. |
| `allow_replace_key` | boolean | no | If true, allows users of the key to provide key material which may replace keys that were previously present. allow_store_key being false overrides this behavior. |
| `allow_store_key` | boolean | no | If true, allows users of the key to provide key material where none was present. |
| `any_mount` | boolean | no | If true, this key may be accessed by any mount without the mount's allowed_manage_keys field being set. |
| `client_id` | string | no | The client ID for credentials to invoke the Azure APIs. |
| `client_secret` | string | no | The client secret for credentials to invoke the Azure APIs. |
| `credentials` | string | no | The path for the GCP credential file |
| `crypto_key` | string | no | The name of the key in GCP Cloud KMS |
| `crypto_key_version` | integer | no | The key version of an existing key in GCP Cloud KMS |
| `curve` | string | no | For ECDSA keys, the desired elliptic curve if the key is to be generated, either P256, P384, or P521. |
| `endpoint` | string | no | The AWS endpoint to use |
| `environment` | string | no | The Azure Cloud environment API endpoints to use. |
| `force_rw_session` | boolean (default: False) | no | If true, forces read/write sessions on the HSM, to work around some buggy HSMs. |
| `key-bits` | integer | no | The desired key length in bits if the RSA key is to be generated, either 2048, 3072, or 4096. |
| `key_bits` | integer | no | For RSA keys, the desired key length in bits if the key is to be generated, either 2048, 3072, or 4096. |
| `key_id` | string | no | The id of a PKCS#11 key to use. As key ids are created by the HSM, it is an error if the key does not yet exist. This value or key_label must be specified. |
| `key_label` | string | no | The label of a PKCS#11 key to use. If the key does not exist and generation is enabled, this is the label that will be given to the generated key. This value or key_id must be specified. |
| `key_name` | string | no | The name of the Azure Key Vault Key. |
| `key_ring` | string | no | The name of the key ring in GCP Cloud KMS |
| `key_type` | string | no | The type of the Azure Key Vault Key. Currently only `RSA-HSM` is supported. |
| `kms_key` | string | no | An identifier for the AWS KMS key. If the key already exists, this can be either the AWS-generated key ID or an alias. If the key is to be generated and the field is non-empty, the provided value will be used to create an alias for the key. |
| `kms_key_id` | string | no | [yandexcloudkms] The ID of the key in Yandex Cloud KMS (symmetric or asymmetric). |
| `library` | string | no | The name of a managed key access library, as defined in the Stronghold server configuration. |
| `max_parallel` | integer | no | The maximum number of concurrent operations that may be submitted to the HSM at a time. |
| `mechanism` | string | no | The mechanism for the given key, specified as a decimal or hexadecimal (prefixed by 0x) string. |
| `oauth_token` | string | no | [yandexcloudkms] OAuth token for Yandex Cloud authentication. Mutually exclusive with service_account_key_json. |
| `pin` | string | no | The access PIN for the slot. |
| `project` | string | no | The name of the GCP project |
| `region` | string | no | The GCP region used for the key ring |
| `resource` | string | no | The Azure Key Vault resource's DNS Suffix to connect to. Needs to be changed to connect to Azure's Managed HSM KeyVault instance type. |
| `secret_key` | string | no | Secret key ID for authenticating to AWS |
| `service_account_key_json` | string | no | [yandexcloudkms] JSON of the authorized key for a Yandex Cloud service account. Stored inside Stronghold, not on the filesystem. Mutually exclusive with oauth_token. |
| `slot` | string | no | The slot number to use, specified as a string (e.g. "0"). |
| `tenant_id` | string | no | The tenant ID for the Azure Active Directory organization. |
| `token_label` | string | no | The slot token label to use. |
| `usages` | string | no | A comma-delimited list of the allowed usages of this key. Valid values are encrypt, decrypt, sign, verify, wrap, unwrap, mac, and random. |
| `vault_name` | string | no | The name of an existing Azure Key Vault instance. |

#### Responses

**200**: OK

### DELETE /sys/managed-keys/{type}/{name}

**Operation ID:** `system-delete-managed-keys-type-name`

Create, update and read the managed key configurations.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the managed key. |
| `type` | string | path | yes | The type of the managed key. |

#### Responses

**204**: empty body

### POST /sys/managed-keys/{type}/{name}/test/sign

**Operation ID:** `system-write-managed-keys-type-name-test-sign`

Test a managed key by signing and verifying some random data.

**Creation supported:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the managed key. |
| `type` | string | path | yes | The type of the managed key. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `hash_algorithm` | string (default: sha2-256) | no | The hashing algorithm to use when signing/verifying the random data. If omitted, GOST PKCS#11 keys use streebog-256 / streebog-512 according to the mechanism; otherwise sha2-256. |
| `use_pss` | boolean (default: False) | no | For RSA backed managed keys attempt to sign with PSS. Defaults to true for Yandex Cloud KMS RSA keys, which sign PSS only. |

#### Responses

**200**: OK

### GET /sys/metrics

**Operation ID:** `metrics`

Export the metrics aggregated for telemetry purpose.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `format` | string | query | no | Format to export metrics into. Currently accepts only "prometheus". |

#### Responses

**200**: OK

### GET /sys/mfa/method

**Operation ID:** `system-list-mfa-method`

Lists all the available MFA methods by their name.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `list` | string (true) | query | yes | Must be set to `true` |

#### Responses

**200**: OK

### GET /sys/mfa/method/duo/{name}

**Operation ID:** `system-read-mfa-method-duo-name`

Defines or updates a Duo MFA method.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Responses

**200**: OK

### POST /sys/mfa/method/duo/{name}

**Operation ID:** `system-write-mfa-method-duo-name`

Defines or updates a Duo MFA method.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `api_hostname` | string | no | API host name for Duo. |
| `integration_key` | string | no | Integration key for Duo. |
| `mount_accessor` | string | no | The mount to tie this method to for use in automatic mappings. The mapping will use the Name field of Aliases associated with this mount as the username in the mapping. |
| `push_info` | string | no | Push information for Duo. |
| `secret_key` | string | no | Secret key for Duo. |
| `use_passcode` | boolean | no | If true, the user is reminded to use the passcode upon MFA validation. This option does not enforce using the passcode. Defaults to false. |
| `username_format` | string | no | A format string for mapping Identity names to MFA method names. Values to subtitute should be placed in {{}}. For example, "{{alias.name}}@example.com". Currently-supported mappings: alias.name: The name returned by the mount configured via the mount_accessor parameter If blank, the Alias's name field will be used as-is. |

#### Responses

**200**: OK

### DELETE /sys/mfa/method/duo/{name}

**Operation ID:** `system-delete-mfa-method-duo-name`

Defines or updates a Duo MFA method.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Responses

**204**: empty body

### GET /sys/mfa/method/multifactor/{name}

**Operation ID:** `system-read-mfa-method-multifactor-name`

Defines or updates a Multifactor MFA method.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Responses

**200**: OK

### POST /sys/mfa/method/multifactor/{name}

**Operation ID:** `system-write-mfa-method-multifactor-name`

Defines or updates a Multifactor MFA method.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `api_url` | string | no | Base URL for the Multifactor API. Defaults to "<https://api.multifactor.ru>". |
| `mount_accessor` | string | no | The mount to tie this method to for use in automatic mappings. The mapping will use the Name field of Aliases associated with this mount as the username in the mapping. |
| `nas_identifier` | string | no | Multifactor NAS Identifier from the resource settings. |
| `shared_secret` | string | no | Multifactor Shared Secret from the resource settings. |
| `timeout_seconds` | integer | no | Maximum number of seconds to wait for Multifactor push confirmation. Defaults to 90. Minimum 65. |
| `username_format` | string | no | A format string for mapping Identity names to Multifactor identities. Values to subtitute should be placed in {{}}. For example, "{{alias.name}}@example.com". Currently-supported mappings: alias.name: The name returned by the mount configured via the mount_accessor parameter If blank, the Alias's name field will be used as-is. |

#### Responses

**200**: OK

### DELETE /sys/mfa/method/multifactor/{name}

**Operation ID:** `system-delete-mfa-method-multifactor-name`

Defines or updates a Multifactor MFA method.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Responses

**204**: empty body

### GET /sys/mfa/method/okta/{name}

**Operation ID:** `system-read-mfa-method-okta-name`

Defines or updates an Okta MFA method.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Responses

**200**: OK

### POST /sys/mfa/method/okta/{name}

**Operation ID:** `system-write-mfa-method-okta-name`

Defines or updates an Okta MFA method.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `api_token` | string | no | Okta API key. |
| `base_url` | string | no | The base domain to use for the Okta API. When not specified in the configuration, "okta.com" is used. |
| `mount_accessor` | string | no | The mount to tie this method to for use in automatic mappings. The mapping will use the Name field of Aliases associated with this mount as the username in the mapping. |
| `org_name` | string | no | Name of the organization to be used in the Okta API. |
| `primary_email` | boolean | no | If true, the username will only match the primary email for the account. Defaults to false. |
| `production` | boolean | no | (DEPRECATED) Use base_url instead. |
| `username_format` | string | no | A format string for mapping Identity names to MFA method names. Values to subtitute should be placed in {{}}. For example, "{{alias.name}}@example.com". Currently-supported mappings: alias.name: The name returned by the mount configured via the mount_accessor parameter If blank, the Alias's name field will be used as-is. |

#### Responses

**200**: OK

### DELETE /sys/mfa/method/okta/{name}

**Operation ID:** `system-delete-mfa-method-okta-name`

Defines or updates an Okta MFA method.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Responses

**204**: empty body

### GET /sys/mfa/method/pingid/{name}

**Operation ID:** `system-read-mfa-method-pingid-name`

Defines or updates a PingID MFA method.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Responses

**200**: OK

### POST /sys/mfa/method/pingid/{name}

**Operation ID:** `system-write-mfa-method-pingid-name`

Defines or updates a PingID MFA method.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `mount_accessor` | string | no | The mount to tie this method to for use in automatic mappings. The mapping will use the Name field of Aliases associated with this mount as the username in the mapping. |
| `settings_file_base64` | string | no | The settings file provided by Ping, Base64-encoded. This must be a settings file suitable for third-party clients, not the PingID SDK or PingFederate. |
| `username_format` | string | no | A format string for mapping Identity names to MFA method names. Values to subtitute should be placed in {{}}. For example, "{{alias.name}}@example.com". Currently-supported mappings: alias.name: The name returned by the mount configured via the mount_accessor parameter If blank, the Alias's name field will be used as-is. |

#### Responses

**200**: OK

### DELETE /sys/mfa/method/pingid/{name}

**Operation ID:** `system-delete-mfa-method-pingid-name`

Defines or updates a PingID MFA method.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Responses

**204**: empty body

### GET /sys/mfa/method/totp/{name}

**Operation ID:** `system-read-mfa-method-totp-name`

Defines or updates a TOTP MFA method.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Responses

**200**: OK

### POST /sys/mfa/method/totp/{name}

**Operation ID:** `system-write-mfa-method-totp-name`

Defines or updates a TOTP MFA method.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `algorithm` | string (default: SHA1) | no | The hashing algorithm used to generate the TOTP token. Options include SHA1, SHA256 and SHA512. |
| `digits` | integer (default: 6) | no | The number of digits in the generated TOTP token. This value can either be 6 or 8. |
| `issuer` | string | no | The name of the key's issuing organization. |
| `key_size` | integer (default: 20) | no | Determines the size in bytes of the generated key. |
| `max_validation_attempts` | integer | no | Max number of allowed validation attempts. |
| `period` | integer (default: 30) | no | The length of time used to generate a counter for the TOTP token calculation. |
| `qr_size` | integer (default: 200) | no | The pixel size of the generated square QR code. |
| `skew` | integer (default: 1) | no | The number of delay periods that are allowed when validating a TOTP token. This value can either be 0 or 1. |

#### Responses

**200**: OK

### DELETE /sys/mfa/method/totp/{name}

**Operation ID:** `system-delete-mfa-method-totp-name`

Defines or updates a TOTP MFA method.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Responses

**204**: empty body

### POST /sys/mfa/method/totp/{name}/admin-destroy

**Operation ID:** `system-write-mfa-method-totp-name-admin-destroy`

Deletes the TOTP secret for the given method name on the given entity.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `entity_id` | string | no | Identifier of the entity from which the MFA method secret needs to be removed. |

#### Responses

**200**: OK

### POST /sys/mfa/method/totp/{name}/admin-generate

**Operation ID:** `system-write-mfa-method-totp-name-admin-generate`

Generates a TOTP secret for the given method name on the given entity.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `entity_id` | string | no | Entity ID on which the generated secret needs to get stored. |

#### Responses

**200**: OK

### GET /sys/mfa/method/totp/{name}/generate

**Operation ID:** `system-read-mfa-method-totp-name-generate`

Generates a TOTP secret for the given method name on the entity of the
		calling token.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the MFA method. |

#### Responses

**200**: OK

### POST /sys/mfa/validate

**Operation ID:** `mfa-validate`

Validates the login for the given MFA methods. Upon successful validation, it returns an auth response containing the client token

**Available without authentication:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `mfa_payload` | object | yes | A map from MFA method ID to a slice of passcodes or an empty slice if the method does not use passcodes |
| `mfa_request_id` | string | yes | ID for this MFA request |

#### Responses

**200**: OK

### GET /sys/monitor

**Operation ID:** `monitor`

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `log_format` | string | query | no | Output format of logs. Supported values are "standard" and "json". The default is "standard". |
| `log_level` | string | query | no | Log level to view system logs at. Currently supported values are "trace", "debug", "info", "warn", "error". |

#### Responses

**200**: OK

### GET /sys/mounts

**Operation ID:** `mounts-list-secrets-engines`

List the currently mounted backends.

#### Responses

**200**: OK

### GET /sys/mounts/{path}

**Operation ID:** `mounts-read-configuration`

Read the configuration of the secret engine at the given path.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | The path to mount to. Example: "aws/east" |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `accessor` | string | no |  |
| `config` | object | no | Configuration for this mount, such as default_lease_ttl and max_lease_ttl. |
| `deprecation_status` | string | no |  |
| `description` | string | no | User-friendly description for this mount. |
| `external_entropy_access` | boolean | no |  |
| `local` | boolean (default: False) | no | Mark the mount as a local mount, which is not replicated and is unaffected by replication. |
| `options` | object | no | The options to pass into the backend. Should be a json object with string keys and values. |
| `plugin_version` | string | no | The semantic version of the plugin to use. |
| `running_plugin_version` | string | no |  |
| `running_sha256` | string | no |  |
| `seal_wrap` | boolean (default: False) | no | Whether to turn on seal wrapping for the mount. |
| `type` | string | no | The type of the backend. Example: "passthrough" |
| `uuid` | string | no |  |

### POST /sys/mounts/{path}

**Operation ID:** `mounts-enable-secrets-engine`

Enable a new secrets engine at the given path.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | The path to mount to. Example: "aws/east" |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `config` | object | no | Configuration for this mount, such as default_lease_ttl and max_lease_ttl. |
| `description` | string | no | User-friendly description for this mount. |
| `external_entropy_access` | boolean (default: False) | no | Whether to give the mount access to Stronghold's external entropy. |
| `local` | boolean (default: False) | no | Mark the mount as a local mount, which is not replicated and is unaffected by replication. |
| `options` | object | no | The options to pass into the backend. Should be a json object with string keys and values. |
| `plugin_name` | string | no | Name of the plugin to mount based from the name registered in the plugin catalog. |
| `plugin_version` | string | no | The semantic version of the plugin to use. |
| `seal_wrap` | boolean (default: False) | no | Whether to turn on seal wrapping for the mount. |
| `type` | string | no | The type of the backend. Example: "passthrough" |

#### Responses

**204**: OK

### DELETE /sys/mounts/{path}

**Operation ID:** `mounts-disable-secrets-engine`

Disable the mount point specified at the given path.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | The path to mount to. Example: "aws/east" |

#### Responses

**200**: OK

### GET /sys/mounts/{path}/tune

**Operation ID:** `mounts-read-tuning-information`

Tune backend configuration parameters for this mount.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | The path to mount to. Example: "aws/east" |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `allowed_managed_keys` | array | no |  |
| `allowed_response_headers` | array | no | A list of headers to whitelist and allow a plugin to set on responses. |
| `audit_non_hmac_request_keys` | array | no |  |
| `audit_non_hmac_response_keys` | array | no |  |
| `default_lease_ttl` | integer | no | The default lease TTL for this mount. |
| `description` | string | no | User-friendly description for this credential backend. |
| `external_entropy_access` | boolean | no |  |
| `force_no_cache` | boolean | no |  |
| `listing_visibility` | string | no |  |
| `max_lease_ttl` | integer | no | The max lease TTL for this mount. |
| `options` | object | no | The options to pass into the backend. Should be a json object with string keys and values. |
| `passthrough_request_headers` | array | no |  |
| `plugin_version` | string | no | The semantic version of the plugin to use. |
| `replication_config` | object | no | KV replication configuration. |
| `token_type` | string | no | The type of token to issue (service or batch). |
| `user_lockout_counter_reset_duration` | integer | no |  |
| `user_lockout_disable` | boolean | no |  |
| `user_lockout_duration` | integer | no |  |
| `user_lockout_threshold` | integer | no |  |

### POST /sys/mounts/{path}/tune

**Operation ID:** `mounts-tune-configuration-parameters`

Tune backend configuration parameters for this mount.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | The path to mount to. Example: "aws/east" |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `allowed_managed_keys` | array | no |  |
| `allowed_response_headers` | array | no | A list of headers to whitelist and allow a plugin to set on responses. |
| `audit_non_hmac_request_keys` | array | no | The list of keys in the request data object that will not be HMAC'ed by audit devices. |
| `audit_non_hmac_response_keys` | array | no | The list of keys in the response data object that will not be HMAC'ed by audit devices. |
| `cmd_enable_repl` | boolean | no | Enable or disable the replication for this mount. |
| `default_lease_ttl` | string | no | The default lease TTL for this mount. |
| `description` | string | no | User-friendly description for this credential backend. |
| `listing_visibility` | string | no | Determines the visibility of the mount in the UI-specific listing endpoint. Accepted value are 'unauth' and 'hidden', with the empty default ('') behaving like 'hidden'. |
| `max_lease_ttl` | string | no | The max lease TTL for this mount. |
| `options` | object | no | The options to pass into the backend. Should be a json object with string keys and values. |
| `passthrough_request_headers` | array | no | A list of headers to whitelist and pass from the request to the plugin. |
| `plugin_version` | string | no | The semantic version of the plugin to use. |
| `src_address` | string | no | Address of the source Stronghold cluster for replication. |
| `src_ca_cert` | string | no | CA certificate for TLS connection to the source cluster. |
| `src_mount_path` | string | no | Mount path of the KV engine on the source cluster. |
| `src_secret_path` | array | no | List of secret paths to replicate from the source cluster. |
| `src_token` | string | no | Token for accessing the API of the remote server. Mutually exclusive with src_wrapping_token. |
| `src_wrapping_token` | string | no | Wrapping token that will be unwrapped on the remote server to obtain the replication token. Mutually exclusive with src_token. |
| `sync_period_min` | integer | no | Replication sync period in minutes. |
| `token_type` | string | no | The type of token to issue (service or batch). |
| `user_lockout_config` | object | no | The user lockout configuration to pass into the backend. Should be a json object with string keys and values. |

#### Responses

**200**: OK

### GET /sys/namespaces

**Operation ID:** `system-list-namespaces`

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `list` | string (true) | query | yes | Must be set to `true` |

#### Responses

**200**: OK

### POST /sys/namespaces/api-lock/lock

**Operation ID:** `system-write-namespaces-api-lock-lock`

Lock the API for a namespace and its descendants

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `path` | string | no | Optional relative path for locking descendants of the current namespace. |

#### Responses

**200**: OK

### POST /sys/namespaces/api-lock/lock/{path}

**Operation ID:** `system-write-namespaces-api-lock-lock-path`

Lock the API for a namespace and its descendants

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | Optional relative path for locking descendants of the current namespace. |

#### Responses

**200**: OK

### POST /sys/namespaces/api-lock/unlock

**Operation ID:** `system-write-namespaces-api-lock-unlock`

Unlock the API for a namespace and its descendants

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `path` | string | no | Optional relative path for unlocking descendants of the current namespace. |
| `unlock_key` | string | no | Key to unlock the namespace. |

#### Responses

**200**: OK

### POST /sys/namespaces/api-lock/unlock/{path}

**Operation ID:** `system-write-namespaces-api-lock-unlock-path`

Unlock the API for a namespace and its descendants

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | Optional relative path for unlocking descendants of the current namespace. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `unlock_key` | string | no | Key to unlock the namespace. |

#### Responses

**200**: OK

### GET /sys/namespaces/{path}

**Operation ID:** `system-read-namespaces-path`

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | Path of the namespace. |

#### Responses

**200**: OK

### POST /sys/namespaces/{path}

**Operation ID:** `system-write-namespaces-path`

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | Path of the namespace. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `custom_metadata` | object | no | User-provided key-value pairs that are used to describe arbitrary information about a namespace. |

#### Responses

**200**: OK

### DELETE /sys/namespaces/{path}

**Operation ID:** `system-delete-namespaces-path`

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `path` | string | path | yes | Path of the namespace. |

#### Responses

**204**: empty body

### GET /sys/plugins/catalog

**Operation ID:** `plugins-catalog-list-plugins`

Lists all the plugins known to Stronghold

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `detailed` | object | no |  |

### GET /sys/plugins/catalog/{name}

**Operation ID:** `plugins-catalog-read-plugin-configuration`

Return the configuration data for the plugin with the given name.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the plugin |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `args` | array | no | The args passed to plugin command. |
| `builtin` | boolean | no |  |
| `command` | string | no | The command used to start the plugin. The executable defined in this command must exist in stronghold's plugin directory. |
| `deprecation_status` | string | no |  |
| `name` | string | no | The name of the plugin |
| `sha256` | string | no | The SHA256 sum of the executable used in the command field. This should be HEX encoded. |
| `version` | string | no | The semantic version of the plugin to use. |

### POST /sys/plugins/catalog/{name}

**Operation ID:** `plugins-catalog-register-plugin`

Register a new plugin, or updates an existing one with the supplied name.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the plugin |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `args` | array | no | The args passed to plugin command. |
| `command` | string | no | The command used to start the plugin. The executable defined in this command must exist in stronghold's plugin directory. |
| `env` | array | no | The environment variables passed to plugin command. Each entry is of the form "key=value". |
| `sha256` | string | no | The SHA256 sum of the executable used in the command field. This should be HEX encoded. |
| `type` | string | no | The type of the plugin, may be auth, secret, or database |
| `version` | string | no | The semantic version of the plugin to use. |

#### Responses

**200**: OK

### DELETE /sys/plugins/catalog/{name}

**Operation ID:** `plugins-catalog-remove-plugin`

Remove the plugin with the given name.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the plugin |

#### Responses

**200**: OK

### GET /sys/plugins/catalog/{type}

**Operation ID:** `plugins-catalog-list-plugins-with-type`

List the plugins in the catalog.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `type` | string | path | yes | The type of the plugin, may be auth, secret, or database |
| `list` | string (true) | query | yes | Must be set to `true` |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `keys` | array | no | List of plugin names in the catalog |

### GET /sys/plugins/catalog/{type}/{name}

**Operation ID:** `plugins-catalog-read-plugin-configuration-with-type`

Return the configuration data for the plugin with the given name.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the plugin |
| `type` | string | path | yes | The type of the plugin, may be auth, secret, or database |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `args` | array | no | The args passed to plugin command. |
| `builtin` | boolean | no |  |
| `command` | string | no | The command used to start the plugin. The executable defined in this command must exist in stronghold's plugin directory. |
| `deprecation_status` | string | no |  |
| `name` | string | no | The name of the plugin |
| `sha256` | string | no | The SHA256 sum of the executable used in the command field. This should be HEX encoded. |
| `version` | string | no | The semantic version of the plugin to use. |

### POST /sys/plugins/catalog/{type}/{name}

**Operation ID:** `plugins-catalog-register-plugin-with-type`

Register a new plugin, or updates an existing one with the supplied name.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the plugin |
| `type` | string | path | yes | The type of the plugin, may be auth, secret, or database |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `args` | array | no | The args passed to plugin command. |
| `command` | string | no | The command used to start the plugin. The executable defined in this command must exist in stronghold's plugin directory. |
| `env` | array | no | The environment variables passed to plugin command. Each entry is of the form "key=value". |
| `sha256` | string | no | The SHA256 sum of the executable used in the command field. This should be HEX encoded. |
| `version` | string | no | The semantic version of the plugin to use. |

#### Responses

**200**: OK

### DELETE /sys/plugins/catalog/{type}/{name}

**Operation ID:** `plugins-catalog-remove-plugin-with-type`

Remove the plugin with the given name.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the plugin |
| `type` | string | path | yes | The type of the plugin, may be auth, secret, or database |

#### Responses

**200**: OK

### POST /sys/plugins/reload/backend

**Operation ID:** `plugins-reload-backends`

Reload mounted plugin backends.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `mounts` | array | no | The mount paths of the plugin backends to reload. |
| `plugin` | string | no | The name of the plugin to reload, as registered in the plugin catalog. |
| `scope` | string | no |  |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `reload_id` | string | no |  |

**202**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `reload_id` | string | no |  |

### GET /sys/plugins/reload/backend/status

**Operation ID:** `system-read-plugins-reload-backend-status`

Get the status of a cluster-scoped reload.

#### Responses

**200**: OK

### GET /sys/policies/acl

**Operation ID:** `policies-list-acl-policies`

List the configured access control policies.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `list` | string (true) | query | yes | Must be set to `true` |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `keys` | array | no |  |
| `policies` | array | no |  |

### GET /sys/policies/acl/{name}

**Operation ID:** `policies-read-acl-policy`

Retrieve information about the named ACL policy.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the policy. Example: "ops" |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `name` | string | no |  |
| `policy` | string | no |  |
| `rules` | string | no |  |

### POST /sys/policies/acl/{name}

**Operation ID:** `policies-write-acl-policy`

Add a new or update an existing ACL policy.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the policy. Example: "ops" |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `policy` | string | no | The rules of the policy. |

#### Responses

**204**: OK

### DELETE /sys/policies/acl/{name}

**Operation ID:** `policies-delete-acl-policy`

Delete the ACL policy with the given name.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the policy. Example: "ops" |

#### Responses

**204**: OK

### GET /sys/policies/egp

**Operation ID:** `system-list-policies-egp`

List the configured access control policies.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `list` | string (true) | query | yes | Must be set to `true` |

#### Responses

**200**: OK

### GET /sys/policies/egp/{name}

**Operation ID:** `system-read-policies-egp-name`

Read, Modify, or Delete an access control policy.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the policy. Example: "ops" |

#### Responses

**200**: OK

### POST /sys/policies/egp/{name}

**Operation ID:** `system-write-policies-egp-name`

Read, Modify, or Delete an access control policy.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the policy. Example: "ops" |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `enforcement_level` | string | no | The enforcement level to apply to the policy. |
| `paths` | array | no | The paths on which the policy should be applied. |
| `policy` | string | no | The rules of the policy. |

#### Responses

**200**: OK

### DELETE /sys/policies/egp/{name}

**Operation ID:** `system-delete-policies-egp-name`

Read, Modify, or Delete an access control policy.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the policy. Example: "ops" |

#### Responses

**204**: empty body

### GET /sys/policies/password

**Operation ID:** `policies-list-password-policies`

List the existing password policies.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `list` | string (true) | query | yes | Must be set to `true` |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `keys` | array | no |  |

### GET /sys/policies/password/{name}

**Operation ID:** `policies-read-password-policy`

Retrieve an existing password policy.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the password policy. |

#### Responses

**204**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `policy` | string | no |  |

### POST /sys/policies/password/{name}

**Operation ID:** `policies-write-password-policy`

Add a new or update an existing password policy.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the password policy. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `policy` | string | no | The password policy |

#### Responses

**204**: OK

### DELETE /sys/policies/password/{name}

**Operation ID:** `policies-delete-password-policy`

Delete a password policy.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the password policy. |

#### Responses

**204**: OK

### GET /sys/policies/password/{name}/generate

**Operation ID:** `policies-generate-password-from-password-policy`

Generate a password from an existing password policy.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the password policy. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `password` | string | no |  |

### GET /sys/policies/rgp

**Operation ID:** `system-list-policies-rgp`

List the configured access control policies.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `list` | string (true) | query | yes | Must be set to `true` |

#### Responses

**200**: OK

### GET /sys/policies/rgp/{name}

**Operation ID:** `system-read-policies-rgp-name`

Read, Modify, or Delete an access control policy.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the policy. Example: "ops" |

#### Responses

**200**: OK

### POST /sys/policies/rgp/{name}

**Operation ID:** `system-write-policies-rgp-name`

Read, Modify, or Delete an access control policy.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the policy. Example: "ops" |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `enforcement_level` | string | no | The enforcement level to apply to the policy. |
| `policy` | string | no | The rules of the policy. |

#### Responses

**200**: OK

### DELETE /sys/policies/rgp/{name}

**Operation ID:** `system-delete-policies-rgp-name`

Read, Modify, or Delete an access control policy.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the policy. Example: "ops" |

#### Responses

**204**: empty body

### GET /sys/policy

**Operation ID:** `policies-list`

List the configured access control policies.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `list` | string | query | no | Return a list if `true` |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `keys` | array | no |  |
| `policies` | array | no |  |

### GET /sys/policy/{name}

**Operation ID:** `policies-read-acl-policy2`

Retrieve the policy body for the named policy.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the policy. Example: "ops" |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `name` | string | no |  |
| `policy` | string | no |  |
| `rules` | string | no |  |

### POST /sys/policy/{name}

**Operation ID:** `policies-write-acl-policy2`

Add a new or update an existing policy.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the policy. Example: "ops" |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `policy` | string | no | The rules of the policy. |
| `rules` | string | no | ⚠️ Deprecated. The rules of the policy. |

#### Responses

**204**: OK

### DELETE /sys/policy/{name}

**Operation ID:** `policies-delete-acl-policy2`

Delete the policy with the given name.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | The name of the policy. Example: "ops" |

#### Responses

**204**: OK

### GET /sys/pprof

**Operation ID:** `pprof-index`

Returns an HTML page listing the available profiles.

#### Responses

**200**: OK

### GET /sys/pprof/allocs

**Operation ID:** `pprof-memory-allocations`

Returns a sampling of all past memory allocations.

#### Responses

**200**: OK

### GET /sys/pprof/block

**Operation ID:** `pprof-blocking`

Returns stack traces that led to blocking on synchronization primitives

#### Responses

**200**: OK

### GET /sys/pprof/cmdline

**Operation ID:** `pprof-command-line`

Returns the running program's command line.

#### Responses

**200**: OK

### GET /sys/pprof/goroutine

**Operation ID:** `pprof-goroutines`

Returns stack traces of all current goroutines.

#### Responses

**200**: OK

### GET /sys/pprof/heap

**Operation ID:** `pprof-memory-allocations-live`

Returns a sampling of memory allocations of live object.

#### Responses

**200**: OK

### GET /sys/pprof/mutex

**Operation ID:** `pprof-mutexes`

Returns stack traces of holders of contended mutexes

#### Responses

**200**: OK

### GET /sys/pprof/profile

**Operation ID:** `pprof-cpu-profile`

Returns a pprof-formatted cpu profile payload.

#### Responses

**200**: OK

### GET /sys/pprof/symbol

**Operation ID:** `pprof-symbols`

Returns the program counters listed in the request.

#### Responses

**200**: OK

### GET /sys/pprof/threadcreate

**Operation ID:** `pprof-thread-creations`

Returns stack traces that led to the creation of new OS threads

#### Responses

**200**: OK

### GET /sys/pprof/trace

**Operation ID:** `pprof-execution-trace`

Returns the execution trace in binary form.

#### Responses

**200**: OK

### GET /sys/quotas/config

**Operation ID:** `rate-limit-quotas-read-configuration`

Create, update and read the quota configuration.

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `enable_rate_limit_audit_logging` | boolean | no |  |
| `enable_rate_limit_response_headers` | boolean | no |  |
| `rate_limit_exempt_paths` | array | no |  |

### POST /sys/quotas/config

**Operation ID:** `rate-limit-quotas-configure`

Create, update and read the quota configuration.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `enable_rate_limit_audit_logging` | boolean | no | If set, starts audit logging of requests that get rejected due to rate limit quota rule violations. |
| `enable_rate_limit_response_headers` | boolean | no | If set, additional rate limit quota HTTP headers will be added to responses. |
| `rate_limit_exempt_paths` | array | no | Specifies the list of exempt paths from all rate limit quotas. If empty no paths will be exempt. |

#### Responses

**204**: OK

### GET /sys/quotas/lease-count

**Operation ID:** `system-list-quotas-lease-count`

Lists the names of all the lease count quotas.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `list` | string (true) | query | yes | Must be set to `true` |

#### Responses

**200**: OK

### GET /sys/quotas/lease-count/{name}

**Operation ID:** `system-read-quotas-lease-count-name`

Get, create or update lease count resource quota for an optional namespace or mount.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the quota rule. |

#### Responses

**200**: OK

### POST /sys/quotas/lease-count/{name}

**Operation ID:** `system-write-quotas-lease-count-name`

Get, create or update lease count resource quota for an optional namespace or mount.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the quota rule. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `max_leases` | integer | no | Maximum number of leases allowed by the quota rule. |
| `path` | string | no | Path including the applicable namespace prefix. |
| `role` | string | no | Login role to apply this quota to. Note that when set, path must be configured to a valid auth method with a concept of roles. |
| `type` | string | no | Type of the quota rule. |

#### Responses

**200**: OK

### DELETE /sys/quotas/lease-count/{name}

**Operation ID:** `system-delete-quotas-lease-count-name`

Get, create or update lease count resource quota for an optional namespace or mount.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the quota rule. |

#### Responses

**204**: empty body

### GET /sys/quotas/rate-limit

**Operation ID:** `rate-limit-quotas-list`

Lists the names of all the rate limit quotas.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `list` | string (true) | query | yes | Must be set to `true` |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `keys` | array | no |  |

### GET /sys/quotas/rate-limit/{name}

**Operation ID:** `rate-limit-quotas-read`

Get, create or update rate limit resource quota for an optional namespace or mount.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the quota rule. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `block_interval` | integer | no |  |
| `interval` | integer | no |  |
| `name` | string | no |  |
| `path` | string | no |  |
| `rate` | number | no |  |
| `role` | string | no |  |
| `type` | string | no |  |

### POST /sys/quotas/rate-limit/{name}

**Operation ID:** `rate-limit-quotas-write`

Get, create or update rate limit resource quota for an optional namespace or mount.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the quota rule. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `block_interval` | integer | no | If set, when a client reaches a rate limit threshold, the client will be prohibited from any further requests until after the 'block_interval' has elapsed. |
| `interval` | integer | no | The duration to enforce rate limiting for (default '1s'). |
| `path` | string | no | Path of the mount or namespace to apply the quota. A blank path configures a global quota. For example namespace1/ adds a quota to a full namespace, namespace1/auth/userpass adds a quota to userpass in namespace1. |
| `rate` | number | no | The maximum number of requests in a given interval to be allowed by the quota rule. The 'rate' must be positive. |
| `role` | string | no | Login role to apply this quota to. Note that when set, path must be configured to a valid auth method with a concept of roles. |
| `type` | string | no | Type of the quota rule. |

#### Responses

**204**: No Content

### DELETE /sys/quotas/rate-limit/{name}

**Operation ID:** `rate-limit-quotas-delete`

Get, create or update rate limit resource quota for an optional namespace or mount.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | Name of the quota rule. |

#### Responses

**204**: OK

### GET /sys/rekey-recovery-key/init

**Operation ID:** `rekey-recovery-key-attempt-read-progress`

Reads the configuration and progress of the current recovery key rekey attempt.

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `backup` | boolean | no |  |
| `n` | integer | no |  |
| `nonce` | string | no |  |
| `pgp_fingerprints` | array | no |  |
| `progress` | integer | no |  |
| `required` | integer | no |  |
| `started` | boolean | no |  |
| `t` | integer | no |  |
| `verification_nonce` | string | no |  |
| `verification_required` | boolean | no |  |

### POST /sys/rekey-recovery-key/init

**Operation ID:** `rekey-recovery-key-attempt-initialize`

Initializes a new recovery key rekey attempt.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `backup` | boolean | no | Specifies if using PGP-encrypted keys, whether Stronghold should also store a plaintext backup of the PGP-encrypted keys. |
| `pgp_keys` | array | no | Specifies an array of PGP public keys used to encrypt the output recovery keys. |
| `require_verification` | boolean | no | Turns on verification functionality |
| `secret_shares` | integer | no | Specifies the number of shares to split the recovery key into. |
| `secret_threshold` | integer | no | Specifies the number of shares required to reconstruct the recovery key. |
| `stored_shares` | integer | no | Specifies the number of shares that should be encrypted by the HSM and stored for auto-unsealing. Currently must be the same as `secret_shares`. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `backup` | boolean | no |  |
| `n` | integer | no |  |
| `nonce` | string | no |  |
| `pgp_fingerprints` | array | no |  |
| `progress` | integer | no |  |
| `required` | integer | no |  |
| `started` | boolean | no |  |
| `t` | integer | no |  |
| `verification_nonce` | string | no |  |
| `verification_required` | boolean | no |  |

### DELETE /sys/rekey-recovery-key/init

**Operation ID:** `rekey-recovery-key-attempt-cancel`

Cancels any in-progress recovery key rekey.

#### Responses

**200**: OK

### POST /sys/rekey-recovery-key/update

**Operation ID:** `rekey-recovery-key-attempt-update`

Enter a single recovery key share to progress the rekey of the Stronghold.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `key` | string | no | Specifies a single recovery key share. |
| `nonce` | string | no | Specifies the nonce of the rekey attempt. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `backup` | boolean | no |  |
| `complete` | boolean | no |  |
| `keys` | array | no |  |
| `keys_base64` | array | no |  |
| `nonce` | string | no |  |
| `pgp_fingerprints` | array | no |  |
| `verification_nonce` | string | no |  |
| `verification_required` | boolean | no |  |

### GET /sys/rekey-recovery-key/verify

**Operation ID:** `rekey-recovery-key-verification-read-progress`

Read the configuration and progress of the current recovery key rekey verification attempt.

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `n` | integer | no |  |
| `nonce` | string | no |  |
| `progress` | integer | no |  |
| `started` | boolean | no |  |
| `t` | integer | no |  |

### POST /sys/rekey-recovery-key/verify

**Operation ID:** `rekey-recovery-key-verification-update`

Enter a single new recovery key share to progress the rekey verification operation.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `key` | string | no | Specifies a single recovery key share from the new set of shares. |
| `nonce` | string | no | Specifies the nonce of the rekey verification operation. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `complete` | boolean | no |  |
| `nonce` | string | no |  |

### DELETE /sys/rekey-recovery-key/verify

**Operation ID:** `rekey-recovery-key-verification-cancel`

Cancel any in-progress recovery key rekey verification operation.

#### Responses

**200**: OK

### GET /sys/rekey/backup

**Operation ID:** `rekey-read-backup-key`

Return the backup copy of PGP-encrypted unseal keys.

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `keys` | object | no |  |
| `keys_base64` | object | no |  |
| `nonce` | string | no |  |

### DELETE /sys/rekey/backup

**Operation ID:** `rekey-delete-backup-key`

Delete the backup copy of PGP-encrypted unseal keys.

#### Responses

**204**: OK

### GET /sys/rekey/init

**Operation ID:** `rekey-attempt-read-progress`

Reads the configuration and progress of the current rekey attempt.

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `backup` | boolean | no |  |
| `n` | integer | no |  |
| `nonce` | string | no |  |
| `pgp_fingerprints` | array | no |  |
| `progress` | integer | no |  |
| `required` | integer | no |  |
| `started` | boolean | no |  |
| `t` | integer | no |  |
| `verification_nonce` | string | no |  |
| `verification_required` | boolean | no |  |

### POST /sys/rekey/init

**Operation ID:** `rekey-attempt-initialize`

Initializes a new rekey attempt.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `backup` | boolean | no | Specifies if using PGP-encrypted keys, whether Stronghold should also store a plaintext backup of the PGP-encrypted keys. |
| `nonce` | string | no | Specifies the nonce of the rekey operation. If the rekey was initialized within the last 10 minutes, you must provide the nonce to cancel the operation. |
| `pgp_keys` | array | no | Specifies an array of PGP public keys used to encrypt the output unseal keys. Ordering is preserved. The keys must be base64-encoded from their original binary representation. The size of this array must be the same as secret_shares. |
| `require_verification` | boolean | no | Turns on verification functionality |
| `secret_shares` | integer | no | Specifies the number of shares to split the unseal key into. |
| `secret_threshold` | integer | no | Specifies the number of shares required to reconstruct the unseal key. This must be less than or equal secret_shares. If using Stronghold HSM with auto-unsealing, this value must be the same as secret_shares. |
| `stored_shares` | integer | no | Specifies the number of shares that should be encrypted by the HSM and stored for auto-unsealing. Currently must be the same as secret_shares. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `backup` | boolean | no |  |
| `n` | integer | no |  |
| `nonce` | string | no |  |
| `pgp_fingerprints` | array | no |  |
| `progress` | integer | no |  |
| `required` | integer | no |  |
| `started` | boolean | no |  |
| `t` | integer | no |  |
| `verification_nonce` | string | no |  |
| `verification_required` | boolean | no |  |

### DELETE /sys/rekey/init

**Operation ID:** `rekey-attempt-cancel`

Cancels any in-progress rekey.

#### Responses

**200**: OK

### GET /sys/rekey/recovery-key-backup

**Operation ID:** `rekey-read-backup-recovery-key`

Allows fetching or deleting the backup of the rotated unseal keys.

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `keys` | object | no |  |
| `keys_base64` | object | no |  |
| `nonce` | string | no |  |

### DELETE /sys/rekey/recovery-key-backup

**Operation ID:** `rekey-delete-backup-recovery-key`

Allows fetching or deleting the backup of the rotated unseal keys.

#### Responses

**204**: OK

### POST /sys/rekey/update

**Operation ID:** `rekey-attempt-update`

Enter a single unseal key share to progress the rekey of the Stronghold.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `key` | string | no | Specifies a single unseal key share. |
| `nonce` | string | no | Specifies the nonce of the rekey attempt. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `backup` | boolean | no |  |
| `complete` | boolean | no |  |
| `keys` | array | no |  |
| `keys_base64` | array | no |  |
| `n` | integer | no |  |
| `nounce` | string | no |  |
| `pgp_fingerprints` | array | no |  |
| `progress` | integer | no |  |
| `required` | integer | no |  |
| `started` | string | no |  |
| `t` | integer | no |  |
| `verification_nonce` | string | no |  |
| `verification_required` | boolean | no |  |

### GET /sys/rekey/verify

**Operation ID:** `rekey-verification-read-progress`

Read the configuration and progress of the current rekey verification attempt.

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `n` | integer | no |  |
| `nounce` | string | no |  |
| `progress` | integer | no |  |
| `started` | string | no |  |
| `t` | integer | no |  |

### POST /sys/rekey/verify

**Operation ID:** `rekey-verification-update`

Enter a single new key share to progress the rekey verification operation.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `key` | string | no | Specifies a single unseal share key from the new set of shares. |
| `nonce` | string | no | Specifies the nonce of the rekey verification operation. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `complete` | boolean | no |  |
| `nounce` | string | no |  |

### DELETE /sys/rekey/verify

**Operation ID:** `rekey-verification-cancel`

Cancel any in-progress rekey verification operation.

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `n` | integer | no |  |
| `nounce` | string | no |  |
| `progress` | integer | no |  |
| `started` | string | no |  |
| `t` | integer | no |  |

### POST /sys/remount

**Operation ID:** `remount`

Initiate a mount migration

**Required sudo:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `from` | string | no | The previous mount point. |
| `to` | string | no | The new mount point. |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `migration_id` | string | no |  |

### GET /sys/remount/status/{migration_id}

**Operation ID:** `remount-status`

Check status of a mount migration

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `migration_id` | string | path | yes | The ID of the migration operation |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `migration_id` | string | no |  |
| `migration_info` | object | no |  |

### POST /sys/renew

**Operation ID:** `leases-renew-lease2`

Renews a lease, requesting to extend the lease.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `increment` | integer | no | The desired increment in seconds to the lease |
| `lease_id` | string | no | The lease identifier to renew. This is included with a lease. |
| `url_lease_id` | string | no | The lease identifier to renew. This is included with a lease. |

#### Responses

**204**: OK

### POST /sys/renew/{url_lease_id}

**Operation ID:** `leases-renew-lease-with-id2`

Renews a lease, requesting to extend the lease.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `url_lease_id` | string | path | yes | The lease identifier to renew. This is included with a lease. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `increment` | integer | no | The desired increment in seconds to the lease |
| `lease_id` | string | no | The lease identifier to renew. This is included with a lease. |

#### Responses

**204**: OK

### POST /sys/replication/dr/primary/demote

**Operation ID:** `system-write-replication-dr-primary-demote`

#### Responses

**200**: OK

### POST /sys/replication/dr/primary/disable

**Operation ID:** `system-write-replication-dr-primary-disable`

#### Responses

**200**: OK

### POST /sys/replication/dr/primary/enable

**Operation ID:** `system-write-replication-dr-primary-enable`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `primary_cluster_addr` | string | no | The address the secondary cluster should connect to. Defaults to the primary's cluster address. |

#### Responses

**200**: OK

### POST /sys/replication/dr/primary/revoke-secondary

**Operation ID:** `system-write-replication-dr-primary-revoke-secondary`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `id` | string | no | The secondary cluster ID to revoke |

#### Responses

**200**: OK

### POST /sys/replication/dr/primary/secondary-token

**Operation ID:** `system-write-replication-dr-primary-secondary-token`

**Required sudo:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `id` | string | no | An opaque identifier that can be used to identify and revoke a secondary cluster's access later. |
| `secondary_public_key` | string | no | A base64-encoded public key generated by the secondary cluster. |
| `ttl` | integer (default: 30m) | no | The TTL to use for the secondary activation token. Defaults to 30 minutes. |

#### Responses

**200**: OK

### POST /sys/replication/dr/secondary/config/reload/{subsystem}

**Operation ID:** `system-write-replication-dr-secondary-config-reload-subsystem`

Reload the given subsystem

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `subsystem` | string | path | yes |  |

#### Responses

**200**: OK

### POST /sys/replication/dr/secondary/disable

**Operation ID:** `system-write-replication-dr-secondary-disable`

**Available without authentication:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `dr_operation_token` | string | no | DR operation token used to authorize this request. |

#### Responses

**200**: OK

### POST /sys/replication/dr/secondary/enable

**Operation ID:** `system-write-replication-dr-secondary-enable`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `ca_file` | string | no | A path to a file containing a PEM-encoded CA certificate to verify the call against the primary's API address |
| `ca_path` | string | no | A path to a directory containing PEM-encoded CA certificates to verify the call against the primary's API address |
| `client_cert_pem` | string | no | The client certificate to use for authentication, in PEM format. Note: client authentication for this operation will always use TLS 1.2 or higher. |
| `client_key_pem` | string | no | The client key to use for authentication, in PEM format. |
| `primary_api_addr` | string | no | The API address of the primary. If not set, the value the primary supplies in the token will be used, which is the primary's redirect address. |
| `token` | string | no | The token given by the primary to activate secondary status for this cluster. |

#### Responses

**200**: OK

### GET /sys/replication/dr/secondary/generate-operation-token/attempt

**Operation ID:** `system-read-replication-dr-secondary-generate-operation-token-attempt`

#### Responses

**200**: OK

### POST /sys/replication/dr/secondary/generate-operation-token/attempt

**Operation ID:** `system-write-replication-dr-secondary-generate-operation-token-attempt`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `otp` | string | no | Specifies a one-time password used to encrypt the generated token. |
| `pgp_key` | string | no | Specifies a base64-encoded PGP public key. |

#### Responses

**200**: OK

### DELETE /sys/replication/dr/secondary/generate-operation-token/attempt

**Operation ID:** `system-delete-replication-dr-secondary-generate-operation-token-attempt`

#### Responses

**204**: empty body

### POST /sys/replication/dr/secondary/generate-operation-token/update

**Operation ID:** `system-write-replication-dr-secondary-generate-operation-token-update`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `key` | string | no | Specifies a single unseal key share. |
| `nonce` | string | no | Specifies the nonce of the attempt. |

#### Responses

**200**: OK

### POST /sys/replication/dr/secondary/generate-public-key

**Operation ID:** `system-write-replication-dr-secondary-generate-public-key`

#### Responses

**200**: OK

### GET /sys/replication/dr/secondary/license/status

**Operation ID:** `system-read-replication-dr-secondary-license-status`

The path responds to the following HTTP methods. GET / Returns information on the installed license POST Sets the license for the server

**Available without authentication:** yes

#### Responses

**200**: OK

### POST /sys/replication/dr/secondary/merkle-check

**Operation ID:** `system-write-replication-dr-secondary-merkle-check`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `dr_operation_token` | string | no | DR operation token used to authorize this request. |

#### Responses

**200**: OK

### POST /sys/replication/dr/secondary/operation-token/delete

**Operation ID:** `system-write-replication-dr-secondary-operation-token-delete`

**Available without authentication:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `dr_operation_token` | string | no | DR operation token used to authorize this request. |

#### Responses

**200**: OK

### POST /sys/replication/dr/secondary/promote

**Operation ID:** `system-write-replication-dr-secondary-promote`

**Available without authentication:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `dr_operation_token` | string | no | DR operation token used to authorize this request. |
| `force` | boolean | no | Set to true if the cluster should be promoted despite replication being in an error state. This could mean some data was not replicated to the secondary |
| `primary_cluster_addr` | string | no | The address the secondary cluster should connect to. Defaults to the primary's cluster address. |

#### Responses

**200**: OK

### POST /sys/replication/dr/secondary/recover

**Operation ID:** `system-write-replication-dr-secondary-recover`

**Available without authentication:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `dr_operation_token` | string | no | DR operation token used to authorize this request. |

#### Responses

**200**: OK

### POST /sys/replication/dr/secondary/reindex

**Operation ID:** `system-write-replication-dr-secondary-reindex`

**Available without authentication:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `diff` | boolean | no | Enables a slower re-indexing which will perform a key level check to diagnose issues. Defaults false. |
| `dr_operation_token` | string | no | DR operation token used to authorize this request. |
| `force` | boolean | no | Forces a complete re-indexing which only scans data available in the storage. Defaults false. |
| `skip_flush` | boolean | no | Skips the tree flushing stage of the reindex process. This setting can be used to reduce the amount of time the tree is locked during a reindex process. If this node is killed before the full tree has been asynchronously flushed the reindex may not have applied fully and a new reindex may need to be done. Shutting down this node cleanly will cause the tree to be flushed prior to shutdown. Defaults false. |

#### Responses

**200**: OK

### POST /sys/replication/dr/secondary/update-primary

**Operation ID:** `system-write-replication-dr-secondary-update-primary`

**Available without authentication:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `ca_file` | string | no | A path to a file containing a PEM-encoded CA certificate to verify the call against the primary's API address |
| `ca_path` | string | no | A path to a directory containing PEM-encoded CA certificates to verify the call against the primary's API address |
| `client_cert_pem` | string | no | The client certificate to use for authentication, in PEM format. Note: client authentication for this operation will always use TLS 1.2 or higher. |
| `client_key_pem` | string | no | The client key to use for authentication, in PEM format. |
| `dr_operation_token` | string | no | DR operation token used to authorize this request. |
| `primary_api_addr` | string | no | The API address of the primary. If not set, the value the primary supplies in the token will be used, which is the primary's redirect address. |
| `token` | string | no | The token given by the primary to activate secondary status for this cluster. |
| `update_primary_addrs` | array | no | A comma separated list of host:port strings to serve as new addresses for the primary cluster |

#### Responses

**200**: OK

### GET /sys/replication/dr/status

**Operation ID:** `system-read-replication-dr-status`

**Available without authentication:** yes

#### Responses

**200**: OK

### POST /sys/replication/merkle-check

**Operation ID:** `system-write-replication-merkle-check`

#### Responses

**200**: OK

### POST /sys/replication/performance/primary/demote

**Operation ID:** `system-write-replication-performance-primary-demote`

#### Responses

**200**: OK

### POST /sys/replication/performance/primary/disable

**Operation ID:** `system-write-replication-performance-primary-disable`

#### Responses

**200**: OK

### GET /sys/replication/performance/primary/dynamic-filter/{id}

**Operation ID:** `system-read-replication-performance-primary-dynamic-filter-id`

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `id` | string | path | yes | The opaque identifier used to identify the secondary. |

#### Responses

**200**: OK

### POST /sys/replication/performance/primary/enable

**Operation ID:** `system-write-replication-performance-primary-enable`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `primary_cluster_addr` | string | no | The address the secondary cluster should connect to. Defaults to the primary's cluster address. |

#### Responses

**200**: OK

### GET /sys/replication/performance/primary/paths-filter/{id}

**Operation ID:** `system-read-replication-performance-primary-paths-filter-id`

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `id` | string | path | yes | The opaque identifier used to identify the secondary. |

#### Responses

**200**: OK

### POST /sys/replication/performance/primary/paths-filter/{id}

**Operation ID:** `system-write-replication-performance-primary-paths-filter-id`

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `id` | string | path | yes | The opaque identifier used to identify the secondary. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `mode` | string (default: allow) | no | The filter mode for the paths filter (allow or deny). Defaults to allow. |
| `paths` | array | no | The paths to filter in replication. Must be a mount or a namespace. |

#### Responses

**200**: OK

### DELETE /sys/replication/performance/primary/paths-filter/{id}

**Operation ID:** `system-delete-replication-performance-primary-paths-filter-id`

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `id` | string | path | yes | The opaque identifier used to identify the secondary. |

#### Responses

**204**: empty body

### POST /sys/replication/performance/primary/revoke-secondary

**Operation ID:** `system-write-replication-performance-primary-revoke-secondary`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `id` | string | no | The secondary cluster ID to revoke |

#### Responses

**200**: OK

### POST /sys/replication/performance/primary/secondary-token

**Operation ID:** `system-write-replication-performance-primary-secondary-token`

**Required sudo:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `id` | string | no | An opaque identifier that can be used to identify and revoke a secondary cluster's access later. |
| `secondary_public_key` | string | no | A base64-encoded public key generated by the secondary cluster. |
| `ttl` | integer (default: 30m) | no | The TTL to use for the secondary activation token. Defaults to 30 minutes. |

#### Responses

**200**: OK

### POST /sys/replication/performance/secondary/disable

**Operation ID:** `system-write-replication-performance-secondary-disable`

#### Responses

**200**: OK

### GET /sys/replication/performance/secondary/dynamic-filter/{id}

**Operation ID:** `system-read-replication-performance-secondary-dynamic-filter-id`

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `id` | string | path | yes | The opaque identifier used to identify the secondary. |

#### Responses

**200**: OK

### POST /sys/replication/performance/secondary/enable

**Operation ID:** `system-write-replication-performance-secondary-enable`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `ca_file` | string | no | A path to a file containing a PEM-encoded CA certificate to verify the call against the primary's API address |
| `ca_path` | string | no | A path to a directory containing PEM-encoded CA certificates to verify the call against the primary's API address |
| `client_cert_pem` | string | no | The client certificate to use for authentication, in PEM format. Note: client authentication for this operation will always use TLS 1.2 or higher. |
| `client_key_pem` | string | no | The client key to use for authentication, in PEM format. |
| `primary_api_addr` | string | no | The API address of the primary. If not set, the value the primary supplies in the token will be used, which is the primary's redirect address. |
| `token` | string | no | The token given by the primary to activate secondary status for this cluster. |

#### Responses

**200**: OK

### POST /sys/replication/performance/secondary/generate-public-key

**Operation ID:** `system-write-replication-performance-secondary-generate-public-key`

#### Responses

**200**: OK

### POST /sys/replication/performance/secondary/promote

**Operation ID:** `system-write-replication-performance-secondary-promote`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `force` | boolean | no | Set to true if the cluster should be promoted despite replication being in an error state. This could mean some data was not replicated to the secondary |
| `primary_cluster_addr` | string | no | The address the secondary cluster should connect to. Defaults to the primary's cluster address. |

#### Responses

**200**: OK

### POST /sys/replication/performance/secondary/update-primary

**Operation ID:** `system-write-replication-performance-secondary-update-primary`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `ca_file` | string | no | A path to a file containing a PEM-encoded CA certificate to verify the call against the primary's API address |
| `ca_path` | string | no | A path to a directory containing PEM-encoded CA certificates to verify the call against the primary's API address |
| `client_cert_pem` | string | no | The client certificate to use for authentication, in PEM format. Note: client authentication for this operation will always use TLS 1.2 or higher. |
| `client_key_pem` | string | no | The client key to use for authentication, in PEM format. |
| `primary_api_addr` | string | no | The API address of the primary. If not set, the value the primary supplies in the token will be used, which is the primary's redirect address. |
| `token` | string | no | The token given by the primary to activate secondary status for this cluster. |
| `update_primary_addrs` | array | no | A comma separated list of host:port strings to serve as new addresses for the primary cluster |

#### Responses

**200**: OK

### GET /sys/replication/performance/status

**Operation ID:** `system-read-replication-performance-status`

**Available without authentication:** yes

#### Responses

**200**: OK

### POST /sys/replication/primary/demote

**Operation ID:** `system-write-replication-primary-demote`

#### Responses

**200**: OK

### POST /sys/replication/primary/disable

**Operation ID:** `system-write-replication-primary-disable`

#### Responses

**200**: OK

### POST /sys/replication/primary/enable

**Operation ID:** `system-write-replication-primary-enable`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `primary_cluster_addr` | string | no | The address the secondary cluster should connect to. Defaults to the primary's cluster address. |

#### Responses

**200**: OK

### POST /sys/replication/primary/revoke-secondary

**Operation ID:** `system-write-replication-primary-revoke-secondary`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `id` | string | no | The secondary cluster ID to revoke |

#### Responses

**200**: OK

### POST /sys/replication/primary/secondary-token

**Operation ID:** `system-write-replication-primary-secondary-token`

**Required sudo:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `id` | string | no | An opaque identifier that can be used to identify and revoke a secondary cluster's access later. |
| `secondary_public_key` | string | no | A base64-encoded public key generated by the secondary cluster. |
| `ttl` | integer (default: 30m) | no | The TTL to use for the secondary activation token. Defaults to 30 minutes. |

#### Responses

**200**: OK

### POST /sys/replication/recover

**Operation ID:** `system-write-replication-recover`

#### Responses

**200**: OK

### POST /sys/replication/reindex

**Operation ID:** `system-write-replication-reindex`

**Required sudo:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `diff` | boolean | no | Enables a slower re-indexing which will perform a key level check to diagnose issues. Defaults false. |
| `force` | boolean | no | Forces a complete re-indexing which only scans data available in the storage. Defaults false. |
| `skip_flush` | boolean | no | Skips the tree flushing stage of the reindex process. This setting can be used to reduce the amount of time the tree is locked during a reindex process. If this node is killed before the full tree has been asynchronously flushed the reindex may not have applied fully and a new reindex may need to be done. Shutting down this node cleanly will cause the tree to be flushed prior to shutdown. Defaults false. |

#### Responses

**200**: OK

### POST /sys/replication/secondary/disable

**Operation ID:** `system-write-replication-secondary-disable`

#### Responses

**200**: OK

### POST /sys/replication/secondary/enable

**Operation ID:** `system-write-replication-secondary-enable`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `ca_file` | string | no | A path to a file containing a PEM-encoded CA certificate to verify the call against the primary's API address |
| `ca_path` | string | no | A path to a directory containing PEM-encoded CA certificates to verify the call against the primary's API address |
| `client_cert_pem` | string | no | The client certificate to use for authentication, in PEM format. Note: client authentication for this operation will always use TLS 1.2 or higher. |
| `client_key_pem` | string | no | The client key to use for authentication, in PEM format. |
| `primary_api_addr` | string | no | The API address of the primary. If not set, the value the primary supplies in the token will be used, which is the primary's redirect address. |
| `token` | string | no | The token given by the primary to activate secondary status for this cluster. |

#### Responses

**200**: OK

### POST /sys/replication/secondary/promote

**Operation ID:** `system-write-replication-secondary-promote`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `force` | boolean | no | Set to true if the cluster should be promoted despite replication being in an error state. This could mean some data was not replicated to the secondary |
| `primary_cluster_addr` | string | no | The address the secondary cluster should connect to. Defaults to the primary's cluster address. |

#### Responses

**200**: OK

### POST /sys/replication/secondary/update-primary

**Operation ID:** `system-write-replication-secondary-update-primary`

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `ca_file` | string | no | A path to a file containing a PEM-encoded CA certificate to verify the call against the primary's API address |
| `ca_path` | string | no | A path to a directory containing PEM-encoded CA certificates to verify the call against the primary's API address |
| `client_cert_pem` | string | no | The client certificate to use for authentication, in PEM format. Note: client authentication for this operation will always use TLS 1.2 or higher. |
| `client_key_pem` | string | no | The client key to use for authentication, in PEM format. |
| `primary_api_addr` | string | no | The API address of the primary. If not set, the value the primary supplies in the token will be used, which is the primary's redirect address. |
| `token` | string | no | The token given by the primary to activate secondary status for this cluster. |

#### Responses

**200**: OK

### GET /sys/replication/status

**Operation ID:** `system-read-replication-status`

**Available without authentication:** yes

#### Responses

**200**: OK

### POST /sys/revoke

**Operation ID:** `leases-revoke-lease2`

Revokes a lease immediately.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `lease_id` | string | no | The lease identifier to renew. This is included with a lease. |
| `sync` | boolean (default: True) | no | Whether or not to perform the revocation synchronously |
| `url_lease_id` | string | no | The lease identifier to renew. This is included with a lease. |

#### Responses

**204**: OK

### POST /sys/revoke-force/{prefix}

**Operation ID:** `leases-force-revoke-lease-with-prefix2`

Revokes all secrets or tokens generated under a given prefix immediately

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `prefix` | string | path | yes | The path to revoke keys under. Example: "prod/aws/ops" |

#### Responses

**204**: OK

### POST /sys/revoke-prefix/{prefix}

**Operation ID:** `leases-revoke-lease-with-prefix2`

Revokes all secrets (via a lease ID prefix) or tokens (via the tokens' path property) generated under a given prefix immediately.

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `prefix` | string | path | yes | The path to revoke keys under. Example: "prod/aws/ops" |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `sync` | boolean (default: True) | no | Whether or not to perform the revocation synchronously |

#### Responses

**204**: OK

### POST /sys/revoke/{url_lease_id}

**Operation ID:** `leases-revoke-lease-with-id2`

Revokes a lease immediately.

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `url_lease_id` | string | path | yes | The lease identifier to renew. This is included with a lease. |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `lease_id` | string | no | The lease identifier to renew. This is included with a lease. |
| `sync` | boolean (default: True) | no | Whether or not to perform the revocation synchronously |

#### Responses

**204**: OK

### POST /sys/rotate

**Operation ID:** `encryption-key-rotate`

Rotates the backend encryption key used to persist data.

**Required sudo:** yes

#### Responses

**204**: OK

### GET /sys/rotate/config

**Operation ID:** `encryption-key-read-rotation-configuration`

Configures settings related to the backend encryption key management.

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `enabled` | boolean | no |  |
| `interval` | integer | no |  |
| `max_operations` | integer | no |  |

### POST /sys/rotate/config

**Operation ID:** `encryption-key-configure-rotation`

Configures settings related to the backend encryption key management.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `enabled` | boolean | no | Whether automatic rotation is enabled. |
| `interval` | integer | no | How long after installation of an active key term that the key will be automatically rotated. |
| `max_operations` | integer | no | The number of encryption operations performed before the barrier key is automatically rotated. |

#### Responses

**204**: OK

### POST /sys/seal

**Operation ID:** `seal`

Seal the Stronghold.

#### Responses

**204**: OK

### GET /sys/seal-status

**Operation ID:** `seal-status`

Check the seal status of a Stronghold.

**Available without authentication:** yes

#### Responses

**200**:

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `build_date` | string | no |  |
| `cluster_id` | string | no |  |
| `cluster_name` | string | no |  |
| `hcp_link_resource_ID` | string | no |  |
| `hcp_link_status` | string | no |  |
| `initialized` | boolean | no |  |
| `migration` | boolean | no |  |
| `n` | integer | no |  |
| `nonce` | string | no |  |
| `progress` | integer | no |  |
| `recovery_seal` | boolean | no |  |
| `sealed` | boolean | no |  |
| `storage_type` | string | no |  |
| `t` | integer | no |  |
| `type` | string | no |  |
| `version` | string | no |  |

### GET /sys/sealwrap/rewrap

**Operation ID:** `system-read-sealwrap-rewrap`

Retrieve the state of any ongoing seal rewrap process

#### Responses

**200**: OK

### POST /sys/sealwrap/rewrap

**Operation ID:** `system-write-sealwrap-rewrap`

Start a seal rewrap process

#### Responses

**200**: OK

### POST /sys/step-down

**Operation ID:** `step-down-leader`

Cause the node to give up active status.

#### Responses

**204**: empty body

### GET /sys/storage/raft/autopilot/configuration

**Operation ID:** `system-read-storage-raft-autopilot-configuration`

Returns autopilot configuration.

#### Responses

**200**: OK

### POST /sys/storage/raft/autopilot/configuration

**Operation ID:** `system-write-storage-raft-autopilot-configuration`

Returns autopilot configuration.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `cleanup_dead_servers` | boolean | no | Controls whether to remove dead servers from the Raft peer list periodically or when a new server joins. |
| `dead_server_last_contact_threshold` | integer | no | Limit on the amount of time a server can go without leader contact before being considered failed. This takes effect only when cleanup_dead_servers is set. |
| `disable_upgrade_migration` | boolean | no | Whether or not to perform automated version upgrades. |
| `dr_operation_token` | string | no | DR operation token used to authorize this request (if a DR secondary node). |
| `last_contact_threshold` | integer | no | Limit on the amount of time a server can go without leader contact before being considered unhealthy. |
| `max_trailing_logs` | integer | no | Amount of entries in the Raft Log that a server can be behind before being considered unhealthy. |
| `min_quorum` | integer | no | Minimum number of servers allowed in a cluster before autopilot can prune dead servers. This should at least be 3. |
| `server_stabilization_time` | integer | no | Minimum amount of time a server must be in a stable, healthy state before it can be added to the cluster. |

#### Responses

**200**: OK

### GET /sys/storage/raft/autopilot/state

**Operation ID:** `system-read-storage-raft-autopilot-state`

Returns the state of the raft cluster under integrated storage as seen by autopilot.

#### Responses

**200**: OK

### POST /sys/storage/raft/bootstrap/answer

**Operation ID:** `system-write-storage-raft-bootstrap-answer`

Accepts an answer from the peer to be joined to the fact cluster.

**Available without authentication:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `answer` | string | no |  |
| `cluster_addr` | string | no |  |
| `non_voter` | boolean | no |  |
| `server_id` | string | no |  |

#### Responses

**200**: OK

### POST /sys/storage/raft/bootstrap/challenge

**Operation ID:** `system-write-storage-raft-bootstrap-challenge`

Creates a challenge for the new peer to be joined to the raft cluster.

**Available without authentication:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `server_id` | string | no |  |

#### Responses

**200**: OK

### GET /sys/storage/raft/configuration

**Operation ID:** `system-read-storage-raft-configuration`

Returns the configuration of the raft cluster.

#### Responses

**200**: OK

### POST /sys/storage/raft/configuration

**Operation ID:** `system-write-storage-raft-configuration`

Returns the configuration of the raft cluster in a DR secondary cluster.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `dr_operation_token` | string | no | DR operation token used to authorize this request (if a DR secondary node). |

#### Responses

**200**: OK

### POST /sys/storage/raft/demote

**Operation ID:** `system-write-storage-raft-demote`

Demotes a voter to a non-voter.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `force` | boolean | no | Demote the server even if the remaining voters would be fewer than min_quorum or would not tolerate a voter failure. |
| `server_id` | string | no |  |

#### Responses

**200**: OK

### POST /sys/storage/raft/promote

**Operation ID:** `system-write-storage-raft-promote`

Promotes a non-voter to a voter.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `server_id` | string | no |  |

#### Responses

**200**: OK

### POST /sys/storage/raft/remove-peer

**Operation ID:** `system-write-storage-raft-remove-peer`

Remove a peer from the raft cluster.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `dr_operation_token` | string | no | DR operation token used to authorize this request (if a DR secondary node). |
| `server_id` | string | no |  |

#### Responses

**200**: OK

### GET /sys/storage/raft/snapshot

**Operation ID:** `system-read-storage-raft-snapshot`

Returns a snapshot of the current state of vault.

#### Responses

**200**: OK

### POST /sys/storage/raft/snapshot

**Operation ID:** `system-write-storage-raft-snapshot`

Installs the provided snapshot, returning the cluster to the state defined in it.

#### Responses

**200**: OK

### GET /sys/storage/raft/snapshot-auto/config/

**Operation ID:** `system-list-storage-raft-snapshot-auto-config`

List automated raft snapshot configurations

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `list` | string (true) | query | yes | Must be set to `true` |

#### Responses

**200**: OK

### GET /sys/storage/raft/snapshot-auto/config/{name}

**Operation ID:** `system-read-storage-raft-snapshot-auto-config-name`

Read automated raft snapshot configuration

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | name of configuration to update |

#### Responses

**200**: OK

### POST /sys/storage/raft/snapshot-auto/config/{name}

**Operation ID:** `system-write-storage-raft-snapshot-auto-config-name`

Update automated raft snapshot configuration

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | name of configuration to update |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `aws_access_key_id` | string | no | AWS access key ID |
| `aws_s3_bucket` | string | no | AWS bucket |
| `aws_s3_ca_certificate` | string | no | PEM-encoded CA certificate used to verify the TLS certificate of the S3 endpoint |
| `aws_s3_disable_tls` | boolean | no | Disable TLS for the AWS endpoint, intended only for testing |
| `aws_s3_enable_kms` | boolean | no | Use KMS to encrypt bucket contents |
| `aws_s3_endpoint` | string | no | AWS endpoint, typically only set when using a non-AWS S3 instance like Minio |
| `aws_s3_kms_key` | string | no | Use named KMS key |
| `aws_s3_region` | string | no | AWS region |
| `aws_s3_server_side_encryption` | boolean | no | Use AES256 to encrypt bucket contents |
| `aws_secret_access_key` | string | no | AWS secret access key |
| `aws_session_token` | string | no | AWS session token |
| `azure_account_key` | string | no | Azure account key |
| `azure_account_name` | string | no | Azure account name |
| `azure_blob_environment` | string (default: AZUREPUBLICCLOUD) | no | Azure blob environment |
| `azure_container_name` | string | no | Azure container name |
| `azure_endpoint` | string | no | Azure blob storage endpoint |
| `file_prefix` | string (default: vault-snapshot) | no | file/object prefix prepended to snapshot ID |
| `google_disable_tls` | boolean | no | Disable TLS, normally only for testing |
| `google_endpoint` | string | no | GCS endpoint |
| `google_gcs_bucket` | string | no | GCS bucket |
| `google_service_account_key` | string | no | Service account key in JSON format |
| `interval` | integer (default: 3600000000000) | no | snapshot schedule |
| `local_max_space` | integer | no | max space on disk to use for snapshots |
| `path_prefix` | string | no | directory (local) or bucket prefix (cloud) for snapshot |
| `retain` | integer (default: 1) | no | how many snapshots to keep |
| `storage_type` | string (local, aws-s3) | no | type of storage to use for the snapshots |

#### Responses

**200**: OK

### DELETE /sys/storage/raft/snapshot-auto/config/{name}

**Operation ID:** `system-delete-storage-raft-snapshot-auto-config-name`

Delete automated raft snapshot configuration

**Required sudo:** yes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | name of configuration to update |

#### Responses

**204**: empty body

### GET /sys/storage/raft/snapshot-auto/status/{name}

**Operation ID:** `system-read-storage-raft-snapshot-auto-status-name`

Read status of automated raft snapshots

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `name` | string | path | yes | name of configuration to update |

#### Responses

**200**: OK

### POST /sys/storage/raft/snapshot-force

**Operation ID:** `system-write-storage-raft-snapshot-force`

Installs the provided snapshot, returning the cluster to the state defined in it. This bypasses checks ensuring the current Autounseal or Shamir keys are consistent with the snapshot data.

#### Responses

**200**: OK

### POST /sys/tools/hash

**Operation ID:** `generate-hash`

Generate a hash sum for input data

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `algorithm` | string (default: sha2-256) | no | Algorithm to use (POST body parameter). Valid values are: \* sha2-224 \* sha2-256 \* sha2-384 \* sha2-512 \* streebog-256 \* streebog-512 Defaults to "sha2-256". |
| `format` | string (default: hex) | no | Encoding format to use. Can be "hex" or "base64". Defaults to "hex". |
| `input` | string | no | The base64-encoded input data |
| `urlalgorithm` | string | no | Algorithm to use (POST URL parameter) |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `sum` | string | no |  |

### POST /sys/tools/hash/{urlalgorithm}

**Operation ID:** `generate-hash-with-algorithm`

Generate a hash sum for input data

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `urlalgorithm` | string | path | yes | Algorithm to use (POST URL parameter) |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `algorithm` | string (default: sha2-256) | no | Algorithm to use (POST body parameter). Valid values are: \* sha2-224 \* sha2-256 \* sha2-384 \* sha2-512 \* streebog-256 \* streebog-512 Defaults to "sha2-256". |
| `format` | string (default: hex) | no | Encoding format to use. Can be "hex" or "base64". Defaults to "hex". |
| `input` | string | no | The base64-encoded input data |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `sum` | string | no |  |

### POST /sys/tools/random

**Operation ID:** `generate-random`

Generate random bytes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `bytes` | integer (default: 32) | no | The number of bytes to generate (POST body parameter). Defaults to 32 (256 bits). |
| `format` | string (default: base64) | no | Encoding format to use. Can be "hex" or "base64". Defaults to "base64". |
| `source` | string (default: platform) | no | Which system to source random data from, ether "platform", "seal", or "all". |
| `urlbytes` | string | no | The number of bytes to generate (POST URL parameter) |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `random_bytes` | string | no |  |

### POST /sys/tools/random/{source}

**Operation ID:** `generate-random-with-source`

Generate random bytes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `source` | string | path | yes | Which system to source random data from, ether "platform", "seal", or "all". |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `bytes` | integer (default: 32) | no | The number of bytes to generate (POST body parameter). Defaults to 32 (256 bits). |
| `format` | string (default: base64) | no | Encoding format to use. Can be "hex" or "base64". Defaults to "base64". |
| `urlbytes` | string | no | The number of bytes to generate (POST URL parameter) |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `random_bytes` | string | no |  |

### POST /sys/tools/random/{source}/{urlbytes}

**Operation ID:** `generate-random-with-source-and-bytes`

Generate random bytes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `source` | string | path | yes | Which system to source random data from, ether "platform", "seal", or "all". |
| `urlbytes` | string | path | yes | The number of bytes to generate (POST URL parameter) |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `bytes` | integer (default: 32) | no | The number of bytes to generate (POST body parameter). Defaults to 32 (256 bits). |
| `format` | string (default: base64) | no | Encoding format to use. Can be "hex" or "base64". Defaults to "base64". |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `random_bytes` | string | no |  |

### POST /sys/tools/random/{urlbytes}

**Operation ID:** `generate-random-with-bytes`

Generate random bytes

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `urlbytes` | string | path | yes | The number of bytes to generate (POST URL parameter) |

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `bytes` | integer (default: 32) | no | The number of bytes to generate (POST body parameter). Defaults to 32 (256 bits). |
| `format` | string (default: base64) | no | Encoding format to use. Can be "hex" or "base64". Defaults to "base64". |
| `source` | string (default: platform) | no | Which system to source random data from, ether "platform", "seal", or "all". |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `random_bytes` | string | no |  |

### POST /sys/unseal

**Operation ID:** `unseal`

Unseal the Stronghold.

**Available without authentication:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `key` | string | no | Specifies a single unseal key share. This is required unless reset is true. |
| `reset` | boolean | no | Specifies if previously-provided unseal keys are discarded and the unseal process is reset. |

#### Responses

**200**:

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `build_date` | string | no |  |
| `cluster_id` | string | no |  |
| `cluster_name` | string | no |  |
| `hcp_link_resource_ID` | string | no |  |
| `hcp_link_status` | string | no |  |
| `initialized` | boolean | no |  |
| `migration` | boolean | no |  |
| `n` | integer | no |  |
| `nonce` | string | no |  |
| `progress` | integer | no |  |
| `recovery_seal` | boolean | no |  |
| `sealed` | boolean | no |  |
| `storage_type` | string | no |  |
| `t` | integer | no |  |
| `type` | string | no |  |
| `version` | string | no |  |

### GET /sys/version-history

**Operation ID:** `version-history`

Returns map of historical version change entries

#### Parameters

| Parameter | Type | Location | Required | Description |
|----------|-----|--------------|--------------|----------|
| `list` | string (true) | query | yes | Must be set to `true` |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `key_info` | object | no |  |
| `keys` | array | no |  |

### GET /sys/wrapping/lookup

**Operation ID:** `read-wrapping-properties2`

Look up wrapping properties for the requester's token.

**Available without authentication:** yes

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `creation_path` | string | no |  |
| `creation_time` | string | no |  |
| `creation_ttl` | integer | no |  |

### POST /sys/wrapping/lookup

**Operation ID:** `read-wrapping-properties`

Look up wrapping properties for the given token.

**Available without authentication:** yes

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `token` | string | no |  |

#### Responses

**200**: OK

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `creation_path` | string | no |  |
| `creation_time` | string | no |  |
| `creation_ttl` | integer | no |  |

### POST /sys/wrapping/rewrap

**Operation ID:** `rewrap`

Rotates a response-wrapped token.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `token` | string | no |  |

#### Responses

**200**: OK

### POST /sys/wrapping/unwrap

**Operation ID:** `unwrap`

Unwraps a response-wrapped token.

#### Request body parameters

| Parameter | Type | Required | Description |
|----------|-----|--------------|----------|
| `token` | string | no |  |

#### Responses

**200**: OK

**204**: No content

### POST /sys/wrapping/wrap

**Operation ID:** `wrap`

Response-wraps an arbitrary JSON object.

#### Responses

**200**: OK
