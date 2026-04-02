---
title: "Audit field exclusion"
weight: 20
description: "Configure exclusion of selected fields from Stronghold audit records."
---

Stronghold supports enabling audit devices with the `exclude` option, which removes selected fields from audit records before they are written to the log. This makes it possible to fine-tune the contents of records for each device.

{{< alert level="warning" >}}
Audit field exclusion is an advanced feature. Using exclusions can cause loss of important information in audit logs. Always test the configuration in a non-production environment.
{{< /alert >}}

When an audit device is enabled with exclusions, each audit record is checked against optional conditions before it is written. If a condition matches, the specified fields are removed from the record. A single device can have multiple condition/field combinations.

The behavior of existing devices and new devices without exclusions does not change.

## The `exclude` option

The value of `exclude` must be a valid JSON array of exclusion objects.

### Exclusion object

- `condition` `(string, optional)` is a predicate expression in [bexpr](https://github.com/hashicorp/go-bexpr) syntax. When it matches, Stronghold removes the fields listed in `fields`. If `condition` is omitted or an empty string, the rule is applied unconditionally.
- `fields` `(string[], required)` is an array of fields to remove, expressed in [JSON Pointer](https://tools.ietf.org/html/rfc6901) syntax.

```json
[
  {
    "condition": "",
    "fields": [ "" ]
  }
]
```

### "Source of truth" semantics

Exclusion conditions are always evaluated against the original, unmodified audit record. Removing fields with one rule **does not affect** evaluation of later rules. This guarantees predictable behavior regardless of rule order.

## Exclusion examples

### Exclude response data

Remove the `data` field from the response in any audit record:

```json
[
  {
    "fields": [ "/response/data" ]
  }
]
```

### Exclude request data for transit mounts

Remove the `data` field from the request for records where `mount_type == "transit"`:

```json
[
  {
    "condition": "\"/request/mount_type\" == transit",
    "fields": [ "/request/data" ]
  }
]
```

### Multiple exclusions

Remove `data` from request and response for `transit`, and also remove `entity_id` from `auth` when `client_token` starts with `hmac`:

```json
[
  {
    "condition": "\"/request/mount_type\" == transit",
    "fields": [ "/request/data", "/response/data" ]
  },
  {
    "condition": "\"/auth/client_token\" matches \"hmac.+\"",
    "fields": [ "/auth/entity_id" ]
  }
]
```

## Condition syntax

Conditions in `condition` support two field reference formats:

1. **Quoted JSON Pointer**. Example: `"/request/mount_type" == transit`
2. **Native bexpr format**. Example: `request.mount_type == "transit"`

Supported operators include:

- `==`
- `!=`
- `matches`
- `in`
- `not in`
- `and`
- `or`
- `not`

{{< alert level="info" >}}
Unlike `filter`, `exclude` conditions are evaluated against the full audit record. All record fields are available, not just a limited set of top-level properties.
{{< /alert >}}

## Audit record structure

The fields available in exclusion conditions follow the actual JSON structure of a Stronghold audit record.

In practice this means:

- for JSON Pointer references, use the real path of the field in the JSON structure;
- for bexpr conditions, you can use the equivalent dot notation;
- if a field is an object, you can address both the object itself and its nested fields.

Typical examples:

- `/request/data` <-> `request.data`
- `/request/mount_type` <-> `request.mount_type`
- `/request/namespace/id` <-> `request.namespace.id`
- `/request/namespace/path` <-> `request.namespace.path`
- `/request/request_uri` <-> `request.request_uri`
- `/auth/entity_id` <-> `auth.entity_id`
- `/response/data` <-> `response.data`
- `/response/wrap_info/token` <-> `response.wrap_info.token`

If you are unsure whether a path is correct, first verify how the field is serialized in the relevant audit record type (`request` or `response`) and only then add the exclusion rule.

## Practical example

Enable a `file` audit device and exclude response data for `kv` mounts:

```shell
stronghold audit enable              \
  -path filtered-file                \
  file                               \
  file_path=/logs/audit.log          \
  exclude='[{"condition": "\"/request/mount_type\" == kv", "fields": ["/response/data"]}]'
```

Combine filtering and exclusions:

```shell
stronghold audit enable                   \
  -path transit-only                      \
  file                                    \
  filter='mount_type == "transit"'        \
  file_path=/logs/transit.log             \
  exclude='[{"fields": ["/request/data"]}]'
```

In this example the device:

1. Accepts only records from `transit` mounts.
2. Removes `request.data` from them before writing the record.
