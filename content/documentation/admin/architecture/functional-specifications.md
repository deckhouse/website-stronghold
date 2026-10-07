---
title: "Functional specifications"
description: "The purpose of Deckhouse Stronghold and its main features: centralized secret storage, access control, encryption, audit, and integration."
weight: 50
---

Deckhouse Stronghold is software for access management and secret protection in a company's IT infrastructure.

Stronghold is designed to store, issue, and distribute secrets in a secure and controlled environment. It lets you securely store and access sensitive information that applications and services need, such as passwords, API keys, certificates, and other secrets. Stronghold provides centralized management of these secrets, reliable storage, audit, and monitoring.

To store data, Stronghold uses `AES-256-GCM` (Galois/Counter Mode) with a 96-bit initialization vector, which provides both confidentiality and integrity of the data. See [Cryptography](../../cryptography/overview/).

## Main features

| Feature | Description | Result |
| :--- | :--- | :--- |
| Centralized secret storage | Stronghold keeps all secrets in one place and manages them centrally. | Secret management becomes more efficient and secure, and user access to secrets is simpler. |
| Access control | Access to secrets is configured with roles, policies, and access control lists (ACL). Policies define who can access which secrets and which operations are allowed. | Access to secrets is controlled with roles, policies, and access control lists. |
| Data encryption | All secrets stored by Stronghold are encrypted with a strong algorithm before they are written to storage. This protects the data from unauthorized access. | Data in Stronghold storage is kept only in encrypted form. |
| Logging and monitoring | Stronghold audits all operations with secrets and provides metrics for monitoring. The audit log shows who performed which actions on secrets and when. | Operations with secrets are logged and tracked according to the administrator's settings. |
| Integration with other tools | Stronghold integrates with container platforms, cloud providers, source code management systems, and other systems. | Stronghold can be used as part of an existing infrastructure together with other software products. |

<!-- TODO(verify): the original page said secrets are "encrypted on the client side"; by the Stronghold architecture, data is encrypted by the server's cryptographic barrier before it is written to storage. The wording was corrected; confirm. -->

How these features are implemented is described in [Architecture overview](../overview/).
