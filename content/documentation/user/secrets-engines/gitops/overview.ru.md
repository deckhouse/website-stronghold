---
title: "Механизм секретов GitOps"
linkTitle: "Обзор"
weight: 10
---

Механизм секретов GitOps отслеживает Git-репозиторий и применяет декларативную конфигурацию Stronghold, когда у новых коммитов достаточно проверенных подписей.

Основные свойства:

- конфигурация описывается в YAML и применяется через API Stronghold;
- состояние применения сохраняется в Stronghold;
- для подключения к Stronghold используются адрес и токен из конфигурации механизма;
- требуется обновляемый periodic-токен; Stronghold продлевает его автоматически за 24 часа до окончания срока действия;
- статус и ошибки доступны по пути `gitops/status`;
- механизм может управлять тем же экземпляром Stronghold, на котором работает, или другим;
- можно включить несколько точек монтирования, чтобы из разных репозиториев управлять разными частями конфигурации в рамках прав каждого токена.

## Настройка

1. Включите механизм секретов GitOps:

   ```bash
   d8 stronghold secrets enable gitops
   ```

   По умолчанию механизм монтируется по пути `gitops/`. Чтобы использовать другой путь, укажите `-path`.

1. Настройте Git-репозиторий для мониторинга:

   ```bash
   d8 stronghold write gitops/configure/git_repository \
       git_repo_url="https://gitlab.example.com/org/stronghold-gitops.git" \
       required_number_of_verified_signatures_on_commit=1 \
       git_poll_period=1m
   ```

   | Параметр | Обязательный | По умолчанию | Описание |
   |----------|--------------|--------------|----------|
   | `git_repo_url` | да (при создании) | — | URL Git-репозитория |
   | `git_branch_name` | нет | `main` | Отслеживаемая ветка |
   | `git_poll_period` | нет | `5m` | Интервал опроса репозитория |
   | `required_number_of_verified_signatures_on_commit` | нет | `0` | Минимальное число проверенных подписей коммита перед применением конфигурации |
   | `git_ca_certificate` | нет | пусто | CA-сертификат для проверки TLS Git |
   | `max_clone_size_bytes` | нет | `10485760` (10 МиБ) | Максимальный размер клона в памяти в байтах; `0` отключает ограничение |

1. Если репозиторий приватный, настройте учётные данные:

   ```bash
   d8 stronghold write gitops/configure/git_credential \
       username=token \
       password=glpat-XXXXXXXX
   ```

1. Создайте PGP-ключи для подписи коммитов:

   ```bash
   gpg --quick-generate-key "key1 <key1@example.com>" rsa4096
   gpg --quick-generate-key "key2 <key2@example.com>" rsa4096
   ```

1. Экспортируйте открытые ключи и загрузите их в Stronghold:

   ```bash
   gpg --armor --output key1.pgp --export key1
   gpg --armor --output key2.pgp --export key2

   d8 stronghold write gitops/configure/trusted_pgp_public_key/key1 public_key=@key1.pgp
   d8 stronghold write gitops/configure/trusted_pgp_public_key/key2 public_key=@key2.pgp
   ```

1. Настройте доступ механизма к API. Предпочтительнее передавать обёрнутый обновляемый periodic-токен. Механизм разворачивает токен и сохраняет его; позже токен прочитать нельзя:

   ```bash
   TOKEN=$(d8 stronghold token create -orphan -period=7d -policy=gitops-apply \
       -display-name="gitops-plugin" -wrap-ttl=1m -field=wrapping_token)

   d8 stronghold write gitops/configure/vault \
       vault_addr=https://stronghold.example.com:8200 \
       wrapping_token=$TOKEN
   ```

   | Параметр | Описание |
   |----------|----------|
   | `vault_addr` | Адрес API Stronghold |
   | `wrapping_token` | Обёрнутый токен; предпочтительный способ передачи учётных данных |
   | `vault_token` | Токен в открытом виде (используйте только если wrapping недоступен) |
   | `vault_namespace` | Неймспейс для вызовов API |
   | `vault_cacert_bytes` | CA-сертификат в формате PEM для проверки TLS |
   | `rotate` | При значении `true` и переданном токене создаётся orphan-токен с теми же параметрами, старый токен отзывается |

   Выдайте токену только политики, необходимые для управления ресурсами из Git-репозитория.

## Подпись и публикация конфигурации

1. Установите [git-signatures](https://github.com/werf/3p-git-signatures).

1. Склонируйте репозиторий конфигурации (или создайте новый) и укажите ключ подписи:

   ```bash
   git clone https://gitlab.example.com/org/stronghold-gitops.git
   cd stronghold-gitops

   gpg --list-key
   git config user.signingKey <KEY_ID>
   ```

1. Добавьте файлы конфигурации в [декларативном YAML-формате](./configuration-format/), сделайте коммит и подпишите его:

   ```bash
   git add .
   git commit -m "Add AppRole auth method"
   git signatures add
   git signatures show
   ```

   Пример вывода `git signatures show`:

   ```text
    Public Key ID    | Status     | Trust     | Date                         | Signer Name
   =====================================================================================================
    0C3AAAA10E30D5F3 | VALIDSIG   | ULTIMATE  | Пн 22 дек 2025 20:19:33 MSK | key1 <key1@example.com>
   ```

1. Отправьте коммит и подписи:

   ```bash
   git push origin main
   git signatures push
   ```

После следующего опроса, если у коммита достаточно проверенных подписей доверенными ключами, механизм применит конфигурацию.

## Статус

Проверьте текущий статус:

```bash
d8 stronghold read gitops/status
```

В ответе:

- `status` — текущий статус процесса или ошибка;
- `last_run` — время последнего периодического запуска (`never`, если механизм ещё не запускался);
- `last_finished_commit` — хеш последнего успешно применённого коммита;
- `last_finished_commit_date` — дата этого коммита.

## Отключение

```bash
d8 stronghold secrets disable gitops
```

При отключении механизма удаляются его данные в хранилище для этой точки монтирования, включая сохранённый API-токен и состояние применения.
