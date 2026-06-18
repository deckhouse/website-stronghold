---
title: "Password policy"
description: "Configuring password generation policies in Stronghold."
weight: 45
---

The password policy (hereinafter referred to as the policy) defines a set of requirements for generated passwords.
Stronghold allows you to configure custom policies and use them instead of the [default policy](#default-password-policy).

Password policies are not related to [access policies](./policy/).
They have similar names but serve different purposes.

Stronghold uses a password policy configuration model compatible with Vault `1.5+`.

## Using password policies in Stronghold

In Stronghold, password policies are used for the following purposes:

- [Generating passwords via the API](#password-generation-via-api) that comply with a specific policy.
- Configuring [automatic password generation](#principles-and-features-of-password-generation) using secret engines. Not all secret engines support password policies. Before configuring a policy for a secret engine, check its [documentation](../user/secrets-engines/overview/).
- Validate a password created by a user using the [`userpass`](../user/auth/userpass/) authentication method.

## Policy structure and syntax

Password policies are defined in HCL or JSON.

The policy describes:

- The password length (specified in the [`length`](#the-length-parameter) parameter).
- One or more rules that the password must satisfy. Each rule is described in the `rule` field. Stronghold allows the use of [`charset`](#the-charset-rule) rules in policies.
  
  {{< alert level="warning" >}}
  A policy must contain at least one `charset` rule. A policy without a `charset` rule will be rejected.
  {{< /alert >}}

Example of a policy:

```hcl
length = 20

rule "charset" {
  charset = "abcdefghijklmnopqrstuvwxyz"
}
```

This policy generates a 20-character password using only lowercase Latin letters.

### The length parameter

The `length` parameter sets the generated password length.

Characteristics:

- type: `int`;
- required;
- value must be at least `4`.

### The charset rule

The `charset` rule defines a character set and, if needed, the minimum number of characters (`min-chars`) from this set that must be present in the password.

If several `charset` rules are specified in a policy, Stronghold combines all character sets and removes duplicates before generation starts (for example, if the sets `abcde` and `cdefg` are specified, the resulting set `abcdefg` will be used for generation).
Each `charset` rule is still checked separately.

{{< alert level="warning" >}}
After combining and deduplicating character sets, the final `charset` used to generate candidate passwords must not contain more than `256` characters.
{{< /alert >}}

#### charset rule parameters

The following parameters are available in the `charset` rule:

| Parameter name | Type | Default value | Description |
| --- | --- | --- | --- |
| `charset` | `string` | - | String representation of the character set. UTF-8 strings are supported. All characters must be printable |
| `min-chars` | `int` | `0` | Minimum number of characters from `charset` that must be present in the password |

If `min-chars` is not specified or is set to `0`, the character set is used for generation but does not add a required constraint to the resulting password.

Example of using the `charset` rule in a policy:

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

## Policy examples

Password 20 characters, only character sets, with no minimum requirement

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

The password must be 20 characters long and include at least one uppercase letter, one lowercase letter, and one number

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

The password must be 20 characters long and include at least one uppercase letter, one lowercase letter, one digit, and one character from the full ASCII set of special characters

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

The password must be 20 characters long and include at least one uppercase letter, one lowercase letter, one number, and one character from the set !@#$

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

```

## Default password policy

Stronghold uses the default password policy for passwords generated without an explicitly specified policy.

This policy requires:

- `20` password characters;
- at least one uppercase letter;
- at least one lowercase letter;
- at least one digit;
- at least one dash character (`-`).

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

## Policy management

### Creating and editing a policy

To create or edit a password policy, use the POST method on `/sys/policies/password/:name`.

Example of creating a policy from an HCL file:

```shell
d8 stronghold write sys/policies/password/my-policy policy=@my-policy.hcl
```

Example of passing a policy directly when creating it:

```shell
d8 stronghold write sys/policies/password/my-policy policy=- <<EOF
length = 20
rule “charset” {
  charset = “abcdefghijklmnopqrstuvwxyz0123456789”
}
EOF
```

To verify that the policy has been created, use the following command:

```bash
d8 stronghold read sys/policies/password/my-policy
```

Example of creating a policy via the API:

```bash
curl \
  --request POST \
  --header “X-Vault-Token: ...” \
  --data my-policy.json \
  https://stronghold.example.com/v1/sys/policies/password/my-policy
```

### Getting a list of policies

To get a list of created policies, use the GET method on `/sys/policies/password`.

Example:

```bash
d8 stronghold read sys/policies/password
```

Example of getting a list of policies via the API:

```bash
curl \
  --header “X-Vault-Token: ...” \
  --request LIST \
  https://stronghold.example.com/v1/sys/policies/password
```

### Retrieving policy information

To get information about a specific policy, use the GET method on `/sys/policies/password/:name`.

Example:

```bash
d8 stronghold read sys/policies/password/my-policy
```

Example of getting information about a specific policy via the API:

```bash
curl \
  --header “X-Vault-Token: ...” \
  https://stronghold.example.com/v1/sys/policies/password/my-policy
```

### Deleting a policy

To delete a password policy, use the DELETE method on `/sys/policies/password/:name`.

Example:

```bash
d8 stronghold delete sys/policies/password/my-policy
```

Example of deleting a policy via the API:

```bash
curl \
  --header “X-Vault-Token: ...” \
  --request DELETE \
  https://stronghold.example.com/v1/sys/policies/password/my-policy
```

## Generating passwords via the API

You can manually generate passwords that comply with a specific policy.
To do this, use the `/sys/policies/password/:name/generate` method, replacing `:name` with the name of the policy the password must comply with.

Example:

```shell
d8 stronghold read sys/policies/password/my-policy/generate
```

Example of generating a password via the API:

```bash
curl \
  --header “X-Vault-Token: ...” \
  https://stronghold.example.com/v1/sys/policies/password/my-policy/generate
```

## Principles and features of password generation

Stronghold first generates a candidate password from the combined character set built from all [`charset`](#the-charset-rule) rules.
Duplicate characters from different sets are removed.
Stronghold then checks the candidate password against all rules.

## Candidate password generation

The candidate password generation method affects security.
The password is generated in such a way that the result cannot be predicted, and the generation process cannot be exploited to launch an attack.

The following are used to generate a candidate password:

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

### Preventing bias

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
For this example, the maximum allowed value will be `19`. In this case, all values from `0` to `19`, inclusive, result in exactly 2 matches for each symbol.
If a random number exceeds this value, it is discarded and generation continues.
This keeps the final password length unchanged and preserves a correct character distribution.

### Performance Characteristics

Password generation speed depends on the policy configuration.
In general, stricter constraints require more time for generation.

A general estimate can be represented as:

```text
password generation speed = (time to generate one candidate password) * (number of generated candidate passwords)
```

The number of attempts depends on how likely it is that the next candidate password will fail validation against all rules.

In practice, this means the following:

- the more `min-chars` constraints there are, the higher the probability of regeneration;
- the smaller the required character subset is, the lower the probability of quickly getting a suitable password;
- as password length increases, the probability of satisfying some requirements increases, but the cost of generation also grows.

A noticeable performance drop occurs when one rule requires at least one character from a very small set, such as `!@#$`, while the overall `charset` is much larger.

To estimate performance for a specific policy, test it through the [password generation API](#generating-passwords-via-the-api) in your Stronghold environment.
