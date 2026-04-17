---
title: "OIDC провайдер GitLab"
linkTitle: "GitLab"
weight: 20
description: "Настройка GitLab как OIDC-провайдера для Deckhouse Stronghold."
---

GitLab может использоваться как OIDC-провайдер для аутентификации пользователей в **Deckhouse Stronghold**.

Перед началом убедитесь, что метод `OIDC` уже включён и настроен в Stronghold, а также определён корректный `redirect URI`, который должен совпадать в Stronghold и в GitLab.

## Порядок настройки

1. Перейдите в GitLab в раздел **Настройки** → **Приложения**.
2. Создайте новое приложение.
3. Укажите имя приложения.
4. Заполните `redirect URI`, который используется в вашей конфигурации Stronghold.
5. Убедитесь, что выбрана область действия **openid**.
6. Сохраните приложение.
7. Скопируйте `Client ID` и `Client Secret`.

## Что дальше

После создания приложения:

- используйте полученные `Client ID` и `Client Secret` при настройке OIDC в Stronghold;
- убедитесь, что `redirect URI` в GitLab и в роли OIDC Stronghold совпадают полностью;
- выполните тестовый вход через CLI или веб-интерфейс Stronghold.
