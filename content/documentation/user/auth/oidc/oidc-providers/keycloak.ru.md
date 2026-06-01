---
title: "Настройка Keycloak как OIDC-провайдера"
linkTitle: "Keycloak"
weight: 30
description: "Настройка Keycloak как OIDC-провайдера для Deckhouse Stronghold."
---

Keycloak можно использовать как OIDC-провайдер для аутентификации пользователей в Deckhouse Stronghold.

Для настройки подготовьте метод `OIDC` в Stronghold и заранее определите `redirect URI`. Этот адрес должен совпадать в конфигурации Stronghold и в настройках клиента Keycloak.

## Настройка клиента в Keycloak

Выполните следующие шаги:

1. Выберите существующий Realm или создайте новый.
1. Создайте новый Client или выберите существующий.
1. Перейдите в настройки клиента: «Settings».
1. Укажите следующие параметры:

   - `Client Protocol` — `openid-connect`;
   - `Access Type` — `confidential`;
   - `Standard Flow Enabled` — `On`.

1. Настройте допустимые URI перенаправления в параметре `Valid Redirect URIs`.
1. Сохраните изменения.
1. Перейдите на страницу «Credentials».
1. Сохраните значения `Client ID` и `Client Secret`.

## Что сделать после настройки

После настройки клиента в Keycloak выполните следующие действия:

- Используйте `Client ID` и `Client Secret` в конфигурации OIDC в Stronghold.
- Убедитесь, что `redirect URI` совпадает с `allowed_redirect_uris` в Stronghold.
- Выполните тестовый вход через веб-интерфейс Stronghold или CLI.
