---
title: "Документация Deckhouse Stronghold"
linkTitle: "Введение"
description: Документация продукта Deckhouse Stronghold
weight: 10
outputs:
  - HTML
  - markdown
  - search
  - llms
  - corpus
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
Хранилище секретной информации реализовано в формате key-value и совместимо с HashiCorp Vault API.

Чтобы быстро найти нужную информацию:

- воспользуйтесь поиском, если вас интересует конкретная возможность Stronghold, параметр или другой объект;
- используйте боковое меню для навигации по разделам документации.

{{< alert level="info" >}}
Бесплатный практический курс [«Обзор возможностей Deckhouse Stronghold»](https://education.flant.ru/course/obzor-vozmozhnostej-deckhouse-stronghold/) в [Академии Deckhouse](https://deckhouse.ru/course-catalog/) поможет быстро познакомиться с продуктом.
{{< /alert >}}

## С чего начать

Готовые решения типовых задач собраны в разделе [«Примеры использования»](./examples/): вход через ALD Pro и Active Directory, секреты в Kubernetes, CI/CD и GitOps, динамические учётные данные, PKI, шифрование, миграция с Vault.

- **Разработчику** — [первый секрет](./user/get-started/first-secret/), [клиентские библиотеки](./examples/delivery/app-clients/), [секреты в подах Kubernetes](./examples/delivery/kubernetes-workloads/).
- **DevOps-инженеру** — [CI/CD](./examples/delivery/ci-cd/), [GitOps](./examples/delivery/gitops/), [Terraform и Ansible](./examples/delivery/terraform-ansible/), [все примеры](./examples/).
- **Администратору** — [установка](./install/), [архитектура](./admin/architecture/overview/), [эксплуатация](./admin/operations/), [резервное копирование](./admin/backups/overview/).
- **Специалисту по ИБ** — [модель угроз](./admin/architecture/threat-model/), [соответствие требованиям ИБ](./about/compliance/), [аудит](./admin/audit/overview/), [криптоалгоритмы](./admin/cryptography/overview/).
- **Переходите с HashiCorp Vault** — [совместимость](./about/vault-compatibility/) и [руководство по миграции](./examples/operations/migration-from-vault/).

Если вам нужна помощь:

- задайте вопрос [в Telegram-канале «Deckhouse | RU-сообщество»](https://t.me/deckhouse_ru);
- если вы используете Enterprise-редакцию, обратитесь в поддержку по адресу [`support@deckhouse.ru`](mailto:support@deckhouse.ru).
