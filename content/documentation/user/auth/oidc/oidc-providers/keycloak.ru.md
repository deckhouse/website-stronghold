---
title: "OIDC провайдер Keycloak"
linkTitle: "Keycloak"
weight: 30
description: "Настройка Keycloak как OIDC-провайдера для Deckhouse Stronghold."
---

Keycloak может использоваться как OIDC-провайдер для аутентификации пользователей в **Deckhouse Stronghold**.

Перед началом убедитесь, что метод `OIDC` уже настроен в Stronghold и что вы знаете корректный `redirect URI`, который должен быть разрешён и в Stronghold, и в Keycloak.

## Порядок настройки

1. Выберите существующий или создайте новый `Realm`.
2. Создайте новый `Client` или выберите существующий.
3. Перейдите в настройки клиента (**Settings**).
4. Установите:
   - **Client Protocol**: `openid-connect`;
   - **Access Type**: `confidential`;
   - **Standard Flow Enabled**: `On`.
5. Настройте допустимые URI перенаправления (`Valid Redirect URIs`).
6. Нажмите **Сохранить**.
7. Перейдите на страницу **Credentials**.
8. Сохраните значения `Client ID` и `Client Secret`.

## Что дальше

После настройки клиента в Keycloak:

- используйте `Client ID` и `Client Secret` в конфигурации OIDC Stronghold;
- проверьте, что `redirect URI` совпадает с `allowed_redirect_uris` в Stronghold;
- выполните тестовый вход через Stronghold UI или CLI.
