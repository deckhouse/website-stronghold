---
title: "OIDC-провайдер GitLab"
linkTitle: "GitLab"
weight: 20
description: "Настройка GitLab в качестве OIDC-провайдера для Deckhouse Stronghold."
---

GitLab можно использовать как OIDC-провайдер для аутентификации пользователей в Deckhouse Stronghold.

Перед началом убедитесь, что в Deckhouse Stronghold уже включён и настроен метод `OIDC`. Также проверьте, что `redirect URI` настроен корректно и полностью совпадает в Deckhouse Stronghold и GitLab.

## Порядок настройки

Выполните следующие шаги:

1. Перейдите в GitLab в раздел «Настройки» → «Приложения».
1. Создайте новое приложение.
1. Укажите имя приложения.
1. Заполните `redirect URI`, который используется в конфигурации Deckhouse Stronghold.
1. Убедитесь, что выбрана область действия `openid`.
1. Сохраните приложение.
1. Скопируйте `Client ID` и `Client Secret`.

## Что делать дальше

После создания приложения выполните следующие действия:

- используйте полученные `Client ID` и `Client Secret` при настройке OIDC в Deckhouse Stronghold;
- убедитесь, что `redirect URI` в GitLab и в конфигурации OIDC в Deckhouse Stronghold полностью совпадают;
- выполните тестовый вход через CLI или веб-интерфейс Deckhouse Stronghold.
