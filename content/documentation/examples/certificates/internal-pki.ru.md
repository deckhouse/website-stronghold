---
title: "Внутренний PKI на Stronghold"
linkTitle: "Внутренний PKI"
description: "Построение двухуровневого PKI: корневой CA вне Stronghold или в отдельном mount, промежуточный CA в Stronghold, роли, выпуск сертификатов, CRL, OCSP, cert-manager и ACME."
weight: 10
params:
  relatedLinks:
    - title: "Механизм секретов PKI"
      url: ../../../user/secrets-engines/pki/
    - title: "Интеграция с cert-manager"
      url: ../cert-manager/
    - title: "API механизмов секретов"
      url: ../../../reference/api/secrets/
---

Двухуровневая схема защищает корневой CA: он подписывает только промежуточные CA и большую часть времени недоступен. Сертификаты для сервисов выпускает промежуточный CA в Stronghold.

## Цель

Развернуть корневой CA, промежуточный CA в Stronghold, роль для выпуска серверных сертификатов, публикацию CRL и OCSP и подключить потребителей: cert-manager и ACME-клиентов.

## Предварительные требования

- Токен Stronghold с правами на включение механизмов секретов и запись в `pki*/`.
- Утилиты `jq` и `openssl` на рабочей станции.
- Внешний адрес Stronghold, доступный клиентам, — в примерах `https://stronghold.example.com`.
- Домен для сертификатов — в примерах `example.com`.

## Шаг 1. Подготовьте корневой CA

Выберите один из вариантов.

### Вариант A. Офлайн-корневой CA (рекомендуется)

Корневой CA хранится вне Stronghold (например, на изолированной машине или в HSM) и используется только для подписи промежуточных CA. В этом варианте в Stronghold нет mount `pki_root`, а шаг 3 выполняется с помощью средств офлайн-CA (`openssl ca`, `openssl x509 -req` или ПО вашего HSM). Сохраните сертификат корневого CA в файл `root-ca.pem`.

### Вариант B. Корневой CA в отдельном mount

Если офлайн-CA не используется, держите корневой CA в отдельном mount, доступ к которому есть только у администраторов PKI.

1. Включите механизм секретов и задайте максимальный срок жизни:

   ```bash
   d8 stronghold secrets enable -path=pki_root pki
   d8 stronghold secrets tune -max-lease-ttl=87600h pki_root
   ```

1. Сгенерируйте корневой CA. Закрытый ключ (`internal`) не покидает Stronghold:

   ```bash
   d8 stronghold write -field=certificate pki_root/root/generate/internal \
     common_name="Example Root CA" \
     issuer_name="root-2026" \
     ttl=87600h > root-ca.pem
   ```

1. Опубликуйте адреса выпускающего сертификата и CRL корневого CA:

   ```bash
   d8 stronghold write pki_root/config/urls \
     issuing_certificates="https://stronghold.example.com/v1/pki_root/ca" \
     crl_distribution_points="https://stronghold.example.com/v1/pki_root/crl"
   ```

Для CA по ГОСТ укажите параметр `key_type`, например `key_type=gost3410-256-paramset-a`. Список допустимых значений приведён в [справочнике API](../../../reference/api/secrets/).

## Шаг 2. Создайте промежуточный CA

1. Включите отдельный mount для промежуточного CA:

   ```bash
   d8 stronghold secrets enable -path=pki_int pki
   d8 stronghold secrets tune -max-lease-ttl=43800h pki_int
   ```

1. Сгенерируйте ключ и запрос на подпись (CSR):

   ```bash
   d8 stronghold write -format=json pki_int/intermediate/generate/internal \
     common_name="Example Intermediate CA" \
     | jq -r '.data.csr' > pki_int.csr
   ```

## Шаг 3. Подпишите промежуточный CA корневым

Для варианта B выполните:

```bash
d8 stronghold write -format=json pki_root/root/sign-intermediate \
  csr=@pki_int.csr \
  format=pem_bundle \
  ttl=43800h \
  | jq -r '.data.certificate' > pki_int.pem
```

Для варианта A подпишите `pki_int.csr` на офлайн-CA и сформируйте файл `pki_int.pem`, содержащий сертификат промежуточного CA и сертификат корневого CA.

## Шаг 4. Импортируйте подписанный сертификат

```bash
d8 stronghold write pki_int/intermediate/set-signed certificate=@pki_int.pem
```

Задайте адреса CRL, OCSP и выпускающего сертификата, которые будут записаны в выпускаемые сертификаты:

```bash
d8 stronghold write pki_int/config/urls \
  issuing_certificates="https://stronghold.example.com/v1/pki_int/ca" \
  crl_distribution_points="https://stronghold.example.com/v1/pki_int/crl" \
  ocsp_servers="https://stronghold.example.com/v1/pki_int/ocsp"
```

## Шаг 5. Создайте роль

Роль ограничивает домены и срок жизни сертификатов:

```bash
d8 stronghold write pki_int/roles/example-com \
  allowed_domains="example.com" \
  allow_subdomains=true \
  max_ttl=720h
```

Держите `max_ttl` коротким: чем короче срок жизни, тем реже нужен отзыв и тем меньше CRL.

## Шаг 6. Выпустите сертификат

```bash
d8 stronghold write pki_int/issue/example-com \
  common_name="api.example.com" \
  ttl=72h
```

Ответ содержит `certificate`, `issuing_ca`, `ca_chain`, `private_key` и `serial_number`. Если закрытый ключ создаётся на стороне клиента, отправьте CSR на эндпоинт `pki_int/sign/example-com`.

## Шаг 7. Настройте CRL и OCSP

1. Включите автоматическую пересборку CRL:

   ```bash
   d8 stronghold write pki_int/config/crl \
     auto_rebuild=true \
     expiry=72h
   ```

1. Включите автоматическую очистку просроченных и отозванных сертификатов:

   ```bash
   d8 stronghold write pki_int/config/auto-tidy \
     enabled=true \
     tidy_cert_store=true \
     tidy_revoked_certs=true \
     safety_buffer=72h
   ```

1. Отзыв сертификата выполняется по серийному номеру:

   ```bash
   d8 stronghold write pki_int/revoke serial_number=<serial_number>
   ```

## Шаг 8. Подключите cert-manager

Для выпуска сертификатов для Ingress и подов Kubernetes используйте cert-manager с Issuer типа `vault`, указывающим на `pki_int/sign/example-com`. Настройка роли Stronghold и ресурсов cert-manager описана в разделе [«Интеграция с cert-manager»](../cert-manager/).

## Шаг 9. Включите ACME (необязательно)

Механизм PKI поддерживает протокол ACME, поэтому сертификаты можно получать стандартными ACME-клиентами.

1. Задайте внешний адрес mount — он используется в ссылках ACME:

   ```bash
   d8 stronghold write pki_int/config/cluster \
     path="https://stronghold.example.com/v1/pki_int" \
     aia_path="https://stronghold.example.com/v1/pki_int"
   ```

1. Разрешите заголовки, необходимые протоколу ACME:

   ```bash
   d8 stronghold secrets tune \
     -passthrough-request-headers=If-Modified-Since \
     -allowed-response-headers=Last-Modified \
     -allowed-response-headers=Location \
     -allowed-response-headers=Replay-Nonce \
     -allowed-response-headers=Link \
     pki_int
   ```

1. Включите ACME и ограничьте роли:

   ```bash
   d8 stronghold write pki_int/config/acme \
     enabled=true \
     allowed_roles="example-com" \
     max_ttl=720h
   ```

1. Используйте в ACME-клиенте каталог роли `https://stronghold.example.com/v1/pki_int/roles/example-com/acme/directory`. По умолчанию для регистрации аккаунта требуется привязка к внешнему аккаунту (EAB, параметр `eab_policy`).

## Проверка

1. Проверьте цепочку выпущенного сертификата:

   ```bash
   d8 stronghold write -format=json pki_int/issue/example-com common_name=test.example.com ttl=1h > test.json
   jq -r '.data.certificate' test.json > test.pem
   jq -r '.data.issuing_ca' test.json > int.pem
   openssl verify -CAfile root-ca.pem -untrusted int.pem test.pem
   ```

1. Проверьте, что CRL доступен:

   ```bash
   curl -s https://stronghold.example.com/v1/pki_int/crl/pem | openssl crl -noout -text | head
   ```

1. Проверьте статус сертификата через OCSP:

   ```bash
   openssl ocsp -issuer int.pem -VAfile int.pem -cert test.pem \
     -url https://stronghold.example.com/v1/pki_int/ocsp -text
   ```

## Очистка

1. Отзовите тестовые сертификаты: `d8 stronghold write pki_int/revoke serial_number=<serial_number>`.
1. Удалите локальные файлы `test.json`, `test.pem`, `int.pem`.
1. Если стенд был тестовым, отключите mount: `d8 stronghold secrets disable pki_int` и `d8 stronghold secrets disable pki_root`. Отключение mount удаляет ключи CA без возможности восстановления.
