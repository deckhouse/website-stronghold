---
title: "Автоматические сертификаты через ACME"
linkTitle: "ACME"
description: "Автоматический выпуск и продление TLS-сертификатов для внутренних серверов через ACME-сервер механизма PKI Stronghold: config/cluster, config/acme, EAB, certbot, lego и Caddy."
weight: 30
params:
  relatedLinks:
    - title: "Внутренний PKI на Stronghold"
      url: ../internal-pki/
    - title: "Механизм секретов PKI"
      url: ../../../user/secrets-engines/pki/
    - title: "API механизмов секретов"
      url: ../../../reference/api/secrets/
---

Механизм секретов PKI Stronghold реализует протокол ACME (RFC 8555). Внутренние серверы получают и продлевают сертификаты стандартными ACME-клиентами — certbot, lego, Caddy — без токена Stronghold и без собственных скриптов выпуска.

![Схема выпуска сертификата по ACME](../../../images/ex-acme.png)

## Цель

Включить ACME на mount промежуточного CA, ограничить выпуск одной ролью, выдать ключи привязки внешнего аккаунта (EAB) и настроить на сервере автоматический выпуск и продление сертификата.

## Предварительные требования

- Mount `pki_int` с промежуточным CA и роль `example-com`, настроенные по примеру [«Внутренний PKI на Stronghold»](../internal-pki/) (шаги 1–5).
- Токен Stronghold с правами на запись в `pki_int/config/*`, `pki_int/roles/*`, `pki_int/eab/*` и на изменение параметров mount (`sys/mounts/pki_int/tune`).
- Внешний адрес Stronghold, доступный ACME-клиентам, — в примерах `https://stronghold.example.com`.
- Сервер с именем `api.example.com`, на котором будет работать ACME-клиент. Stronghold должен иметь сетевой доступ к этому серверу для проверки challenge (для `http-01` — порт `80`/TCP). Поддерживаются типы challenge `http-01`, `dns-01` и `tls-alpn-01`.
- Сертификат, которым защищён API Stronghold, должен быть доверенным на сервере-клиенте.

## Шаг 1. Задайте внешний адрес mount

ACME-сервер формирует ссылки каталога на основе параметра `path` из `config/cluster`. Укажите адрес, по которому клиенты обращаются к mount:

```bash
d8 stronghold write pki_int/config/cluster \
  path="https://stronghold.example.com/v1/pki_int" \
  aia_path="https://stronghold.example.com/v1/pki_int"
```

В кластере из нескольких узлов указывайте адрес балансировщика или любого узла этого кластера. Адрес не должен указывать на другой кластер репликации.

## Шаг 2. Разрешите заголовки ACME

Протокол ACME передаёт служебные данные в HTTP-заголовках. Разрешите их для mount:

```bash
d8 stronghold secrets tune \
  -passthrough-request-headers=If-Modified-Since \
  -allowed-response-headers=Last-Modified \
  -allowed-response-headers=Location \
  -allowed-response-headers=Replay-Nonce \
  -allowed-response-headers=Link \
  pki_int
```

## Шаг 3. Включите ACME

Включите ACME, разрешите только роль `example-com`, запретите неограниченный каталог по умолчанию и потребуйте EAB:

```bash
d8 stronghold write pki_int/config/acme \
  enabled=true \
  allowed_roles="example-com" \
  default_directory_policy="role:example-com" \
  eab_policy="always-required" \
  max_ttl=720h
```

Параметры:

- `allowed_roles` — роли, через каталоги которых разрешён выпуск. Значение по умолчанию `*` разрешает все роли, включая `sign-verbatim`.
- `default_directory_policy` — политика для каталогов без роли (`pki_int/acme/directory`). По умолчанию используется `sign-verbatim`, то есть выпуск без ограничений роли. Значение `role:<имя>` применяет ограничения указанной роли; роль должна входить в `allowed_roles`.
- `eab_policy` — требование EAB при регистрации ACME-аккаунта. Допустимые значения: `not-required`, `new-account-required`, `always-required`. Задавайте параметр явно (в примере ниже — `always-required`).
- `max_ttl` — максимальный срок действия сертификатов, выпускаемых через ACME. По умолчанию `2160h` (90 дней).
- `challenge_permitted_ip_ranges`, `challenge_excluded_ip_ranges` — CIDR, к которым Stronghold может или не может обращаться при проверке challenge.
- `dns_resolver` — DNS-резолвер в формате `<host>:<port>` для разрешения имён при проверке challenge. Задайте его, если внутренние имена не разрешаются системным резолвером узлов Stronghold.

Проверьте конфигурацию:

```bash
d8 stronghold read pki_int/config/acme
```

Каталог ACME роли доступен без аутентификации по адресу:

```text
https://stronghold.example.com/v1/pki_int/roles/example-com/acme/directory
```

## Шаг 4. Выпустите ключ EAB

EAB связывает ACME-аккаунт с оператором Stronghold: зарегистрировать аккаунт может только тот, кто получил ключ. Ключ одноразовый и используется при первой регистрации аккаунта.

1. Создайте ключ EAB для каталога роли:

   ```bash
   d8 stronghold write -f -format=json pki_int/roles/example-com/acme/new-eab > eab.json
   jq -r '.data.id, .data.key' eab.json
   ```

   Ответ содержит `id` (идентификатор ключа, `kid`) и `key` (HMAC-ключ). Передайте их администратору сервера по защищённому каналу.

1. Просмотрите неиспользованные ключи:

   ```bash
   d8 stronghold list pki_int/eab
   ```

1. При необходимости удалите неиспользованный ключ:

   ```bash
   d8 stronghold delete pki_int/eab/<key_id>
   ```

## Шаг 5. Настройте ACME-клиент

Выполните шаг на сервере `api.example.com`. Задайте переменные:

```bash
ACME_DIR="https://stronghold.example.com/v1/pki_int/roles/example-com/acme/directory"
EAB_KID="<id из eab.json>"
EAB_HMAC="<key из eab.json>"
```

Если сертификат API Stronghold выпущен внутренним CA, укажите клиенту файл с этим CA. В примерах используется `/etc/ssl/stronghold-ca.pem`.

### certbot

1. Получите сертификат. Режим `--standalone` поднимает временный веб-сервер на порту `80` для проверки `http-01`:

   ```bash
   REQUESTS_CA_BUNDLE=/etc/ssl/stronghold-ca.pem \
   certbot certonly --standalone \
     --server "$ACME_DIR" \
     --eab-kid "$EAB_KID" \
     --eab-hmac-key "$EAB_HMAC" \
     --agree-tos -m admin@example.com \
     -d api.example.com
   ```

   Если на сервере уже работает веб-сервер, используйте `--webroot -w <каталог>` или плагин веб-сервера.

1. Сертификат и ключ сохраняются в `/etc/letsencrypt/live/api.example.com/`: `fullchain.pem`, `privkey.pem`. Укажите их в конфигурации сервиса.

### lego

```bash
LEGO_CA_CERTIFICATES=/etc/ssl/stronghold-ca.pem \
lego --server "$ACME_DIR" \
  --eab --kid "$EAB_KID" --hmac "$EAB_HMAC" \
  --email admin@example.com --accept-tos \
  --domains api.example.com \
  --http \
  run
```

Сертификаты сохраняются в каталог `.lego/certificates/`.

### Caddy

Caddy сам получает и продлевает сертификаты для сайтов из `Caddyfile`. Укажите Stronghold как ACME-сервер в глобальных параметрах:

```caddyfile
{
  email admin@example.com
  acme_ca https://stronghold.example.com/v1/pki_int/roles/example-com/acme/directory
  acme_ca_root /etc/ssl/stronghold-ca.pem
  acme_eab {
    key_id <id из eab.json>
    mac_key <key из eab.json>
  }
}

api.example.com {
  reverse_proxy 127.0.0.1:8080
}
```

## Шаг 6. Настройте продление

Срок действия сертификата ограничен `max_ttl` роли и `max_ttl` из `config/acme`. Продление выполняется повторным заказом через уже зарегистрированный аккаунт, новый ключ EAB не требуется.

- **certbot**. Пакеты certbot обычно устанавливают systemd-таймер или задание cron, которое запускает `certbot renew`. Для коротких сертификатов задайте порог продления в `/etc/letsencrypt/renewal/api.example.com.conf`, например `renew_before_expiry = 3 days`, и перезагрузку сервиса через `--deploy-hook`:

  ```bash
  certbot renew --deploy-hook "systemctl reload nginx"
  ```

  Файл CA для `REQUESTS_CA_BUNDLE` должен быть доступен и при автоматическом запуске, например через `Environment=` в unit-файле таймера.

- **lego**. Запускайте продление по расписанию, например ежедневно:

  ```bash
  LEGO_CA_CERTIFICATES=/etc/ssl/stronghold-ca.pem \
  lego --server "$ACME_DIR" --email admin@example.com \
    --domains api.example.com --http \
    renew --days 7 --renew-hook "systemctl reload nginx"
  ```

- **Caddy** продлевает сертификаты автоматически, дополнительная настройка не нужна.

## Проверка

1. Убедитесь, что каталог ACME отвечает:

   ```bash
   curl -s --cacert /etc/ssl/stronghold-ca.pem \
     https://stronghold.example.com/v1/pki_int/roles/example-com/acme/directory | jq
   ```

   В ответе должны быть ссылки `newNonce`, `newAccount` и `newOrder` с адресом из `config/cluster`.

1. Проверьте выпущенный сертификат: издатель, имя и срок действия:

   ```bash
   openssl x509 -in /etc/letsencrypt/live/api.example.com/fullchain.pem \
     -noout -issuer -subject -dates
   ```

1. Проверьте, что сертификат есть в хранилище mount:

   ```bash
   d8 stronghold list pki_int/certs
   ```

1. Проверьте продление без выпуска сертификата:

   ```bash
   REQUESTS_CA_BUNDLE=/etc/ssl/stronghold-ca.pem certbot renew --dry-run
   ```

## Очистка

1. Отзовите тестовый сертификат: `certbot revoke --cert-name api.example.com` (с `REQUESTS_CA_BUNDLE`) или `d8 stronghold write pki_int/revoke serial_number=<serial_number>`.
1. Удалите неиспользованные ключи EAB: `d8 stronghold delete pki_int/eab/<key_id>`.
1. Удалите локальный файл `eab.json`.
1. Если ACME больше не нужен, отключите его: `d8 stronghold write pki_int/config/acme enabled=false`.
