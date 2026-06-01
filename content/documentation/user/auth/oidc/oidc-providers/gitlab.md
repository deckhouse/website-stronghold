---
title: "GitLab OIDC provider"
linkTitle: "GitLab"
weight: 20
description: "Configure GitLab as an OIDC provider for Deckhouse Stronghold."
---

GitLab can be used as an OIDC provider for user authentication in Deckhouse Stronghold.

Before you begin, make sure that the `OIDC` auth method is already enabled and configured in Deckhouse Stronghold. Also verify that the `redirect URI` is configured correctly and matches exactly in Deckhouse Stronghold and GitLab.

## Configuration steps

Perform the following steps:

1. In GitLab, go to **Settings** → **Applications**.
1. Create a new application.
1. Specify the application name.
1. Fill in the `redirect URI` used in the Deckhouse Stronghold configuration.
1. Make sure the `openid` scope is selected.
1. Save the application.
1. Copy the `Client ID` and `Client Secret`.

## What to do next

After creating the application, do the following:

- Use the obtained `Client ID` and `Client Secret` when configuring OIDC in Deckhouse Stronghold.
- Make sure that the `redirect URI` in GitLab and in the OIDC configuration in Deckhouse Stronghold match exactly.
- Perform a test sign-in through the CLI or the Deckhouse Stronghold web interface.
