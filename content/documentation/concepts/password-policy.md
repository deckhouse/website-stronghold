---
title: "Password policy"
description: "Configuring password generation policies in Stronghold."
weight: 10
---

A password policy is a set of rules that defines how Stronghold generates a password.
These policies are used only by some secrets engines.
They let you configure password generation requirements for a specific engine.
Support for password policies depends on the secrets engine, so check its documentation before configuring a policy.

Password policies are unrelated to access policies.
These entities have similar names but different purposes.

Stronghold uses a password policy configuration model compatible with Vault `1.5+`.

{% alert level="danger" %}
Password policies are an advanced Stronghold feature.
They are used when generating credentials for external systems, such as databases or LDAP.
Use them with caution.
{% endalert %}

## How It Works

A password policy consists of two parts:

- the `length` parameter, which sets the password length;
- a set of rules the password must satisfy.

Stronghold first generates a candidate password from the combined character set built from all `charset` rules.
Duplicate characters from different sets are removed.
Stronghold then checks the candidate password against all rules.

A rule is an assertion about what the password must look like.
For example, a `charset` rule can require at least one lowercase letter in the password.
If the password does not contain any lowercase letters, it is rejected.

You can specify several rules in one policy.
This lets you define more complex requirements, such as at least one lowercase letter, one uppercase letter, and one digit.

## Candidate Password Generation

The candidate password generation method affects security.
A password must be generated in a way that prevents prediction and avoids generation properties that could be exploited in an attack.

Candidate password generation requires the following data:

1. A cryptographically secure random number generator.
1. A `charset` from which characters are selected.
1. The password length.

At a high level, the process looks like this:

1. Stronghold gets `N` random values, where `N` is the password length.
1. Each value is interpreted as an index in the `charset` array.
1. Characters are selected by these indexes.
1. The selected characters are combined into a candidate password.
1. The resulting password is checked against all policy rules.

For example, you need to generate an `8`-character password from the `abcdefghij` character set.

Stronghold gets the following random values:

```text
[3, 2, 0, 8, 7, 3, 5, 1]
```

These values correspond to indexes in `charset`:

```text
[3, 2, 0, 8, 7, 3, 5, 1] => [d, c, a, i, h, d, f, b]
```

The resulting candidate password is `dcaihdfb`.

### Preventing Bias

Restricting a random value to the size of the `charset` array with the modulo operation can introduce bias.
This causes some characters to be selected more often than others.

For example, if the generator returns values in the `0-255` range and the `charset` length does not divide `256` evenly, the first characters in the set can appear more often.

Consider a simplified example.
Suppose `charset` is `abcdefghij`, which contains `10` characters.
Suppose the generator returns values in the `0-25` range.

In this case:

- values `0-9` correspond to all `10` characters;
- values `10-19` again correspond to all `10` characters;
- values `20-25` correspond only to the first `6` characters.

As a result, the `abcdef` characters are selected more often than `ghij`.

To avoid this behavior, Stronghold calculates the maximum allowed value that can be safely used to index `charset`.
If a random number exceeds this value, it is discarded and generation continues.
This keeps the final password length unchanged and preserves a correct character distribution.

## Performance Characteristics

Password generation speed depends on the policy configuration.
In general, stricter constraints require more time for generation.

A general estimate can be represented as:

```text
(time to generate one candidate password) * (number of generated candidate passwords)
```

The number of attempts depends on how likely it is that the next candidate password will fail validation against all rules.

In practice, this means the following:

- the more `min-chars` constraints there are, the higher the probability of regeneration;
- the smaller the required character subset is, the lower the probability of quickly getting a suitable password;
- as password length increases, the probability of satisfying some requirements increases, but the cost of generation also grows.

A noticeable performance drop occurs when one rule requires at least one character from a very small set, such as `!@#$`, while the overall `charset` is much larger.

To estimate performance for a specific policy, test it through the password generation API in your Stronghold environment.

## Password Policy Syntax

Password policies are defined in HCL or JSON.
They describe the password length and the set of rules the password must satisfy.

A minimal policy looks like this:

```hcl
length = 20

rule "charset" {
  charset = "abcdefghijklmnopqrstuvwxyz"
}
```

This policy generates `20`-character passwords using only lowercase Latin letters.

You can specify several rules of the same type in one policy.
For example, the following policy generates a `20`-character password that contains at least one lowercase letter, one uppercase letter, one digit, and one special character from the `!@#$%^&*` set:

```hcl
length = 20

rule "charset" {
  charset = "abcdefghijklmnopqrstuvwxyz"
  min-chars = 1
}

rule "charset" {
  charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
  min-chars = 1
}

rule "charset" {
  charset = "0123456789"
  min-chars = 1
}

rule "charset" {
  charset = "!@#$%^&*"
  min-chars = 1
}
```

A valid policy must include at least one `charset` rule.
A policy without `charset` is invalid because Stronghold cannot determine the character set to use for generation.

The following policy is rejected:

```hcl
length = 20
```

## Parameters and Available Rules

### The length Parameter

The `length` parameter sets the generated password length.

Main characteristics:

- type: `int`;
- required;
- value must be at least `4`.

`length` is not a rule.
It is a separate part of the configuration that determines the password length before rule validation.

### The charset Rule

The `charset` rule defines a character set and, if needed, the minimum number of characters from this set that must be present in the password.

This rule does two things:

- contributes to the combined character set used for generation;
- defines a constraint on the resulting password content.

If several `charset` rules are specified in a policy, Stronghold combines all character sets and removes duplicates before generation starts.
Each `charset` rule is still checked separately.

{% alert level="warning" %}
After combining and deduplicating character sets, the final `charset` used to generate candidate passwords must not contain more than `256` characters.
{% endalert %}

#### charset Rule Parameters

The following parameters are available in the `charset` rule:

| Parameter name | Type | Default value | Description |
| --- | --- | --- | --- |
| `charset` | `string` | - | String representation of the character set. UTF-8 strings are supported. All characters must be printable |
| `min-chars` | `int` | `0` | Minimum number of characters from `charset` that must be present in the password |

If `min-chars` is not specified or is set to `0`, the character set is used for generation but does not add a required constraint to the resulting password.

#### charset Rule Example

```hcl
length = 20

rule "charset" {
  charset = "abcde"
  min-chars = 1
}

rule "charset" {
  charset = "01234"
  min-chars = 1
}
```

This policy generates a password from the combined `abcde01234` character set.
The password must contain at least one character from `abcde` and at least one character from `01234`.

If character sets in different rules overlap, Stronghold removes duplicates before generation.
For example, if the `abcde` and `cdefg` sets are specified, `abcdefg` is used for generation.
The password must still satisfy both rules.

Another example:

```hcl
length = 8

rule "charset" {
  charset = "abcde"
}

rule "charset" {
  charset = "01234"
  min-chars = 1
}
```

This policy generates an `8`-character password from the `abcde01234` set.
The password must contain at least one character from `01234`, but does not have to contain characters from `abcde`.
Therefore, `04031945` is a valid value for this policy.

## Policy Examples

### Character Sets Only, With No Required Minimum

```hcl
length = 20

rule "charset" {
  charset = "abcdefghijklmnopqrstuvwxyz"
}

rule "charset" {
  charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
}

rule "charset" {
  charset = "0123456789"
}

rule "charset" {
  charset = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
}
```

### One Uppercase Letter, One Lowercase Letter, And One Digit

```hcl
length = 20

rule "charset" {
  charset = "abcdefghijklmnopqrstuvwxyz"
  min-chars = 1
}

rule "charset" {
  charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
  min-chars = 1
}

rule "charset" {
  charset = "0123456789"
  min-chars = 1
}

rule "charset" {
  charset = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
}
```

### One Uppercase Letter, One Lowercase Letter, One Digit, And One Character From The Full ASCII Special Character Set

```hcl
length = 20

rule "charset" {
  charset = "abcdefghijklmnopqrstuvwxyz"
  min-chars = 1
}

rule "charset" {
  charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
  min-chars = 1
}

rule "charset" {
  charset = "0123456789"
  min-chars = 1
}

rule "charset" {
  charset = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
  min-chars = 1
}
```

### One Uppercase Letter, One Lowercase Letter, One Digit, and One Character From !@#$

```hcl
length = 20

rule "charset" {
  charset = "abcdefghijklmnopqrstuvwxyz"
  min-chars = 1
}

rule "charset" {
  charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
  min-chars = 1
}

rule "charset" {
  charset = "0123456789"
  min-chars = 1
}

rule "charset" {
  charset = "!@#$"
  min-chars = 1
}

rule "charset" {
  charset = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
}
```

## Default Password Policy

Stronghold uses the default password policy for passwords generated without an explicitly specified policy.

This policy requires:

- `20` password characters;
- one uppercase letter;
- one lowercase letter;
- one digit;
- one dash character (`-`).

Default policy example:

```hcl
length = 20

rule "charset" {
  charset = "abcdefghijklmnopqrstuvwxyz"
  min-chars = 1
}

rule "charset" {
  charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
  min-chars = 1
}

rule "charset" {
  charset = "0123456789"
  min-chars = 1
}

rule "charset" {
  charset = "-"
  min-chars = 1
}
```
