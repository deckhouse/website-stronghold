---
title: "Документация Deckhouse Stronghold"
linkTitle: "Введение"
description: Документация продукта Deckhouse Stronghold
weight: 10
outputs:
  - HTML
  - search
  - print
params:
  no_list: true
cascade:
  params:
    simple_list: true
---

{{< downloads >}}

Добро пожаловать на главную страницу документации Deckhouse Stronghold.

Deckhouse Stronghold обеспечивает безопасное хранение и управление жизненным циклом конфиденциальных данных (секретов).
Хранилище секретной информации реализовано в формате key-value и совместимо с Hashicorp Vault API.

{{< alert level="info" >}}
Бесплатный практический курс [«Установка и обзор Deckhouse Stronghold»](https://education.flant.ru/course/ustanovka-i-obzor-deckhouse-stronghold/) в [Академии Deckhouse](https://deckhouse.ru/course-catalog/) поможет быстро познакомиться с продуктом.
{{< /alert >}}

В документации представлены следующие разделы:

- [Быстрый старт](/products/stronghold/gs/) - пошаговая инструкция по установке типовой конфигурации Deckhouse Stronghold.
- [О платформе](./about/editions/) — сведения о редакциях, каналах обновлений и технических требованиях.
- [Установка](./install/dkp/install/steps/prepare/) — порядок установки Deckhouse Stronghold.
- [Настройка](./install/dkp/platform-management/node-management/node-group/) — настройка доступа, резервное копирование ключей, установка сертификатов, настройка аутентификации.
- [Работа с политиками](./concepts/policy/) — управление доступом к секретам и операции над ними.
- [Работа с токенами доступа](./user/auth/token/) — методы аутентификации пользователей и управление токенами доступа.
- [Работа с секретами](./user/secrets-engines/kv/overview/) — механизмы секретов и способы их доставки в приложения.
- [Справка](/products/kubernetes-platform/documentation/v1/cli/d8/) — справочная информация о ресурсах, модулях и их конфигурациях.

Если возникнут вопросы, обращайтесь в наш [Telegram-канал](https://t.me/deckhouse_ru).
Мы обязательно поможем и проконсультируем.

Если вы используете Enterprise-редакцию, по вопросам поддержки пишите на почту&nbsp;<a href="mailto:support@deckhouse.ru">support@deckhouse.ru</a>.
