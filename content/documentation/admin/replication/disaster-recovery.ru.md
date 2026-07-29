---
title: "Disaster recovery"
linkTitle: "Disaster recovery"
weight: 30
description: "Настройка репликации Disaster Recovery, повышение DR-secondary при аварийном переключении и церемония выпуска DR operation token."
---

Репликация Disaster Recovery (DR) держит горячий резервный кластер — полную
копию хранилища primary вместе с локальными данными (токенами, арендами).
DR-secondary не обслуживает клиентов (кроме распечатывания и небольшого
служебного набора) и ждёт promote, чтобы принять нагрузку при отказе primary.

## Перед началом

- Убедитесь, что оба кластера работают на Stronghold EE с integrated Raft
  storage и что репликация включена (смотрите [Обзор](../overview/)).
- Убедитесь, что кластерный порт primary доступен с secondary и что у вас есть
  CA-сертификат primary для TLS.
- Подготовьте токен с правами на `sys/replication/*` на primary.
- Обеспечьте доступ к долям ключей для церемонии promote: при ручном
  распечатывании (Shamir) — доли unseal-ключа, при авто-распечатывании — доли
  recovery-ключа. Понадобится их кворум по порогу, заданному при инициализации.

В примерах ниже `${PRIMARY_ADDR}` и `${SECONDARY_ADDR}` — API-адреса кластеров,
а `${VAULT_TOKEN}` — токен с правами на `sys/replication/*`.

## Шаг 1. Включите DR primary

```shell
d8 stronghold write -force sys/replication/dr/primary/enable
```

## Шаг 2. Создайте activation-токен для DR secondary

```shell
d8 stronghold write sys/replication/dr/primary/secondary-token id=dr-1 ttl=24h
```

Параметр `id` обязателен, `ttl` по умолчанию — `24h`. Команда возвращает
wrapping-токен в поле `wrap_info.token` — его и передайте на secondary.

## Шаг 3. Включите DR secondary

```shell
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  --request POST \
  --data @dr-secondary-enable.json \
  "${SECONDARY_ADDR}/v1/sys/replication/dr/secondary/enable"
```

Пример `dr-secondary-enable.json`:

```json
{
  "token": "<wrapping_token_from_step_2>",
  "primary_api_addr": "<primary_api_address>",
  "ca_cert": "<primary_ca_certificate_in_pem>"
}
```

Для окружений с самоподписанными сертификатами параметр `ca_cert` (CA primary в
формате PEM) обязателен. DR-secondary копирует всё хранилище, включая локальные
данные, и не обслуживает клиентские запросы.

## Шаг 4. Проверьте статус

```shell
d8 stronghold read -address="${SECONDARY_ADDR}" sys/replication/dr/status
```

## Повышение DR-secondary

Повышение требует **DR operation token** — его выпускают через церемонию с
долями ключей и одноразовым паролем (OTP).

1. Запустите церемонию. В ответе возвращаются `nonce` и `otp`:

   ```shell
   curl \
     --request PUT \
     --data '{}' \
     "${SECONDARY_ADDR}/v1/sys/replication/dr/secondary/generate-operation-token/attempt"
   ```

1. Внесите долю ключа. Повторите для каждой требуемой доли:

   ```shell
   curl \
     --request PUT \
     --data '{"key":"<unseal_or_recovery_key_share>","nonce":"<nonce>"}' \
     "${SECONDARY_ADDR}/v1/sys/replication/dr/secondary/generate-operation-token/update"
   ```

   Когда `complete` становится `true`, в ответе появляется `encoded_token`.
   Расшифруйте его с помощью `otp`, чтобы получить DR operation token.

1. Повысьте secondary с помощью DR operation token:

   ```shell
   curl \
     --header "X-Vault-Token: ${VAULT_TOKEN}" \
     --request POST \
     --data '{"dr_operation_token":"<dr_operation_token>"}' \
     "${SECONDARY_ADDR}/v1/sys/replication/dr/secondary/promote"
   ```

После promote бывший secondary становится активным DR primary и обслуживает
клиентов.

## Возврат прежнего primary

После аварийного переключения в кластере уже есть активный primary — повышенный
secondary. Прежний primary нельзя просто запустить обратно как primary, иначе в
кластере окажется два primary. Вместо этого подключите его к новому primary как
DR secondary.

1. На новом primary выпустите activation-токен:
   `d8 stronghold write sys/replication/dr/primary/secondary-token id=<id>`.
1. Если прежний primary ещё считает себя primary, понизьте его:
   `d8 stronghold write -force sys/replication/dr/primary/demote`.
1. Подключите его к новому primary:
   `d8 stronghold write sys/replication/dr/secondary/update-primary token=<activation_token>`.

Используйте именно token-метод: у нового primary другая идентичность, и
`primary_cluster_addr` для него не подойдёт.

## Операции управления

| Действие | Эндпоинт |
| --- | --- |
| Статус | `d8 stronghold read sys/replication/dr/status` |
| Отозвать secondary | `d8 stronghold write sys/replication/dr/primary/revoke-secondary id=dr-1` |
| Отключить на primary | `d8 stronghold write -force sys/replication/dr/primary/disable` |
| Понизить primary (demote) | `d8 stronghold write -force sys/replication/dr/primary/demote` |
| Перенаправить secondary | `d8 stronghold write sys/replication/dr/secondary/update-primary token=<activation_token>` |

Примечания:

- `demote` понижает DR primary до отключённого DR secondary, сохраняя cluster ID
  и локальные данные, — узел готов переподключиться без обнуления.
- Плановое переключение: `demote` на текущем primary, `promote` на резерве,
  затем `update-primary`, чтобы вернуть прежний primary как secondary (смотрите
  [Возврат прежнего primary](#возврат-прежнего-primary)).
