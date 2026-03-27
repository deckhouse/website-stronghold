---
title: "Audit log record schema"
weight: 15
description: "Reference for the schema of Stronghold audit log records."
---

Stronghold writes audit events as individual JSON objects separated by newlines. Audit log records contain both shared attributes that apply to all API calls and request or response fields that depend on the specific operation.

Audit log records represent:

- API requests received by Stronghold;
- API responses returned by Stronghold.

You can correlate a response record with its matching request by using `request.id`, which appears in both records.

## Top-level record structure

Regardless of record type, the following top-level fields are typically used:

| Attribute | Type | Description |
|---|---|---|
| `auth` | `object` | Authentication object describing the subject that performed the API call. |
| `error` | `string` | Error text if the operation failed. Usually omitted for successful requests. |
| `forwarded_from` | `string` | Host and port of the node that forwarded the request. Present only in matching scenarios. |
| `request` | `object` | Request object describing request details. |
| `response` | `object` | Response object. Present in `response` records. |
| `time` | `string` | Event time in ISO 8601 format. |
| `type` | `string` | Record type: `request` or `response`. |

## Example request record

```json
{
  "type": "request",
  "time": "2025-06-05T16:10:22.292517Z",
  "request": {
    "id": "6e3d8a4a-1c4b-4f7e-8d91-9b0df2c7a111"
  },
  "auth": {
    "client_token": "hmac-sha256:..."
  },
  "error": "..."
}
```

## Example response record

```json
{
  "type": "response",
  "time": "2025-06-05T16:10:22.292639Z",
  "request": {
    "id": "6e3d8a4a-1c4b-4f7e-8d91-9b0df2c7a111"
  },
  "response": {
    "data": {}
  },
  "auth": {
    "client_token": "hmac-sha256:..."
  }
}
```

## The `auth` object

Stronghold includes only relevant authentication attributes in the `auth` object. For example, unauthenticated requests do not include `client_token` or `accessor`, and `metadata` is omitted if the token has no metadata.

In practice, the `auth` object most often contains:

- `accessor`
- `client_token`
- `display_name`
- `token_type`
- `token_issue_time`
- `token_ttl`
- `metadata`
- `entity_id`
- `policies`
- `identity_policies`
- `token_policies`
- `policy_results`

These fields help explain which subject executed the request, under which authentication context, and which policies led to the request being allowed.

## The `request` object

The `request` object describes the technical details of the incoming call. The most important fields are:

- `id` for the unique request identifier;
- `operation` for the type of operation: `create`, `read`, `update`, `delete`, `list`;
- `namespace` for namespace ID and path;
- `path` for the API path that handled the request;
- `request_uri` for the original request URI when it differs from `path`;
- `mount_accessor`, `mount_point`, `mount_type` for mount metadata;
- `mount_running_version`, `mount_running_sha256`, `mount_is_external_plugin` for plugin or builtin engine details;
- `remote_address`, `remote_port` for client network information;
- `headers` for request headers, if configured for logging;
- `client_id`, `client_token`, `client_token_accessor` for client and token identifiers;
- `data` for the request payload;
- `wrap_ttl` when response wrapping is requested.

A simplified example:

```json
{
  "id": "",
  "operation": "",
  "namespace": {
    "id": "",
    "path": ""
  },
  "path": "",
  "mount_point": "",
  "mount_type": "",
  "remote_address": "",
  "remote_port": 1234,
  "headers": {},
  "data": {}
}
```

## The `response` object

The `response` object describes the result of the API request. It can contain:

- `auth` when the operation returns a token;
- `headers` for response headers;
- `redirect` for auth backend redirects;
- `warnings` for API warnings;
- `data` for the response payload;
- `secret` for leased secret metadata;
- `wrap_info` for wrapping token properties;
- `mount_class`, `mount_accessor`, `mount_point`, `mount_type` for mount details;
- `mount_running_plugin_version`, `mount_running_sha256`, `mount_is_external_plugin` for plugin details.

A simplified example:

```json
{
  "data": {},
  "headers": {},
  "mount_point": "",
  "mount_type": "",
  "warnings": [""],
  "secret": {
    "lease_id": ""
  },
  "wrap_info": {
    "token": "",
    "ttl": 60
  }
}
```

## Protection of sensitive data

By default, Stronghold does not store most sensitive string values in plaintext. Instead, they are hashed using HMAC-SHA256 and a salt specific to the audit device.

This means:

- secrets are not written to the log in plaintext;
- values can still be verified later through the audit hash mechanism;
- once an audit device is disabled, its salt is no longer available, so old values can no longer be matched through that device.

Keep in mind that hashing primarily applies to string values received in JSON or returned in JSON. Other types, such as integers and booleans, may still be stored in plaintext.

## Practical notes

- The complete audit trail should be built from the union of all enabled audit devices.
- If `filter` and `exclude` are enabled, a given device stores the already filtered or modified version of the record.
- For most engineering tasks, understanding the top-level structure and the `auth`, `request`, and `response` objects is enough.

## Related pages

- [Audit in Stronghold](./overview/)
- [Audit log filtering](./audit_filtering/)
- [Audit field exclusion](./audit_exclusion/)
